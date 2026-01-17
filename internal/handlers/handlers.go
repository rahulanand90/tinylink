package handlers

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/tinylink/internal/db"
	"github.com/yourusername/tinylink/internal/models"
)

const (
	baseURL     = "http://tinylink.io/"
	codeLength  = 6
	charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/urls", CreateURL).Methods("POST")
	r.HandleFunc("/{shortCode}", Redirect).Methods("GET")
}

// CreateURL handles the creation of a new shortened URL
func CreateURL(w http.ResponseWriter, r *http.Request) {
	var req models.CreateURLRequest

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate the URL
	if req.LongURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Generate a short code
	shortCode := generateShortCode()

	// Store in database
	_, err = db.DB.Exec(
		"INSERT INTO urls (short_code, long_url) VALUES ($1, $2)",
		shortCode, req.LongURL,
	)
	if err != nil {
		log.Printf("Failed to insert URL: %v", err)
		http.Error(w, "Failed to create shortened URL", http.StatusInternalServerError)
		return
	}

	// Create response
	resp := models.CreateURLResponse{
		ShortCode: shortCode,
		ShortURL:  baseURL + shortCode,
		LongURL:   req.LongURL,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Redirect handles the redirection from a short URL to the original URL
func Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	// Get the original URL from the database
	var longURL string
	err := db.DB.QueryRow(
		"UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1 RETURNING long_url",
		shortCode,
	).Scan(&longURL)

	if err != nil {
		log.Printf("Database error: %v", err)
		http.NotFound(w, r)
		return
	}

	// Redirect to the original URL
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

// generateShortCode generates a random short code for the URL
func generateShortCode() string {
	code := make([]byte, codeLength)

	// Use crypto/rand for secure random generation
	_, err := rand.Read(code)
	if err != nil {
		// Fallback to less secure method if crypto/rand fails
		for i := 0; i < codeLength; i++ {
			code[i] = byte(time.Now().UnixNano() % int64(len(charset)))
			time.Sleep(1 * time.Nanosecond)
		}
	}

	// Map the random bytes to characters in our charset
	for i := 0; i < codeLength; i++ {
		code[i] = charset[int(code[i])%len(charset)]
	}

	return string(code)
}
