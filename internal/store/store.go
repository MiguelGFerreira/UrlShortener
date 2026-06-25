// Package store holds the PostgreSQL access shared by the shortener and
// redirector services.
package store

import (
	"context"
	"database/sql"
	"fmt"

	"UrlShortener/internal/model"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Connect opens a connection to PostgreSQL and verifies it is reachable.
func Connect(user, pass string) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=url_shortener sslmode=disable", user, pass))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// ShortURLExists reports whether the given short alias is already taken.
func ShortURLExists(ctx context.Context, db *sql.DB, shortURL string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings WHERE short_url = $1", shortURL).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// InsertMapping stores a new long URL -> short alias mapping.
func InsertMapping(ctx context.Context, db *sql.DB, m model.URLMapping) error {
	_, err := db.ExecContext(ctx, "INSERT INTO url_mappings (long_url, short_url) VALUES ($1, $2)", m.LongURL, m.ShortURL)
	return err
}

// MappingByShortURL returns the mapping for the given short alias.
// It returns sql.ErrNoRows when the alias does not exist.
func MappingByShortURL(ctx context.Context, db *sql.DB, shortURL string) (model.URLMapping, error) {
	m := model.URLMapping{ShortURL: shortURL}
	err := db.QueryRowContext(ctx, "SELECT long_url FROM url_mappings WHERE short_url = $1", shortURL).Scan(&m.LongURL)
	return m, err
}
