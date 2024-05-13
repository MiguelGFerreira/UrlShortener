package model

import "time"

// URLMapping struct represents a record in the database table storing URL mappings
type URLMapping struct {
	ID        int       `json:"id"`
	LongURL   string    `json:"long_url"`
	ShortURL  string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
}
