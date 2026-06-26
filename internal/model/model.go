package model

import "time"

// URLMapping represents a record in the url_mappings table.
type URLMapping struct {
	ID             int        `json:"id"`
	LongURL        string     `json:"long_url"`
	ShortURL       string     `json:"short_url"`
	Clicks         int64      `json:"clicks"`
	CreatedAt      time.Time  `json:"created_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}
