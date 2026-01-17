package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/yourusername/tinylink/internal/db"
	"github.com/yourusername/tinylink/internal/handlers"
)

func main() {
	// Initialize database connection
	dbHost := os.Getenv("DB_HOST")
	if err := db.Connect(dbHost); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database tables
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create router
	r := mux.NewRouter()

	// Register routes
	handlers.RegisterRoutes(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
