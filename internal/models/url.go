package models

import "time"

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
