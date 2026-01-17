package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// URL represents the shortened URL data structure
type URL struct {
	ID        int64     `json:"id"`
	ShortCode string    `json:"short_code"`
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int64     `json:"clicks"`
}

// CreateURLRequest represents the request to create a shortened URL
type CreateURLRequest struct {
	LongURL string `json:"long_url"`
}

// CreateURLResponse represents the response after creating a shortened URL
type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
	LongURL   string `json:"long_url"`
}

var (
	db *sql.DB
	// baseURL is the base URL for the shortened links
	baseURL = "http://tinylink.io/"
)

func main() {
	// Initialize database connection
	initDB()

	// Create router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/api/urls", createURLHandler).Methods("POST")
	r.HandleFunc("/{shortCode}", redirectHandler).Methods("GET")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initDB() {
	// In a production environment, these would be environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	
	connStr := fmt.Sprintf("postgres://postgres:postgres@%s/tinylink?sslmode=disable", host)
	log.Printf("Connecting to database: %s", host)
	
	var err error
	var dbRetries = 5
	
	// Retry connection a few times (useful when starting with Docker Compose)
	for i := 0; i < dbRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, dbRetries, err)
			time.Sleep(2 * time.Second)
			continue
		}
		
		err = db.Ping()
		if err == nil {
			break
		}
		
		log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, dbRetries, err)
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		log.Fatalf("Could not connect to database after %d attempts: %v", dbRetries, err)
	}

	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			long_url TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			clicks INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	log.Println("Database initialized successfully")
}

// createURLHandler handles the creation of a new shortened URL
func createURLHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest
	
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
	_, err = db.Exec(
		"INSERT INTO urls (short_code, long_url) VALUES ($1, $2)",
		shortCode, req.LongURL,
	)
	if err != nil {
		log.Printf("Failed to insert URL: %v", err)
		http.Error(w, "Failed to create shortened URL", http.StatusInternalServerError)
		return
	}

	// Create response
	resp := CreateURLResponse{
		ShortCode: shortCode,
		ShortURL:  baseURL + shortCode,
		LongURL:   req.LongURL,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// redirectHandler handles the redirection from a short URL to the original URL
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	// Get the original URL from the database
	var longURL string
	err := db.QueryRow(
		"UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1 RETURNING long_url",
		shortCode,
	).Scan(&longURL)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to the original URL
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

// generateShortCode generates a random short code for the URL
func generateShortCode() string {
	// More reliable implementation using crypto/rand
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6
	
	// Import crypto/rand at the top of the file if you haven't already
	code := make([]byte, codeLength)
	
	// Try to use crypto/rand, but fall back to a less secure method if it fails
	_, err := rand.Read(code)
	if err != nil {
		// Fallback to less secure method
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
