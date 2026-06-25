package model

import "time"

// URLMapping represents a record in the url_mappings table.
type URLMapping struct {
	ID        int       `json:"id"`
	LongURL   string    `json:"long_url"`
	ShortURL  string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
}
