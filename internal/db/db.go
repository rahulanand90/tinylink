package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Connect initializes the database connection
func Connect(host string) error {
	if host == "" {
		host = "localhost"
	}

	connStr := fmt.Sprintf("postgres://postgres:postgres@%s/tinylink?sslmode=disable", host)
	log.Printf("Connecting to database: %s", host)

	var err error
	const dbRetries = 5

	// Retry connection a few times (useful when starting with Docker Compose)
	for i := 0; i < dbRetries; i++ {
		DB, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, dbRetries, err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = DB.Ping()
		if err == nil {
			break
		}

		log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, dbRetries, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("could not connect to database after %d attempts: %w", dbRetries, err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// Initialize creates tables if they don't exist
func Initialize() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			long_url TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			clicks INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
