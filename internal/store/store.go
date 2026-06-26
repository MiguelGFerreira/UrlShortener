// Package store holds the PostgreSQL access shared by the shortener and
// redirector services.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"UrlShortener/internal/model"

	"github.com/lib/pq" // PostgreSQL driver
)

// ErrAliasTaken is returned when inserting a mapping whose short alias already
// exists (PostgreSQL unique-violation, SQLSTATE 23505).
var ErrAliasTaken = errors.New("short alias already taken")

// Config holds the PostgreSQL connection settings.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// ConfigFromEnv reads the connection settings from environment variables,
// falling back to sensible local defaults.
func ConfigFromEnv() Config {
	return Config{
		Host:     env("DB_HOST", "localhost"),
		Port:     env("DB_PORT", "5432"),
		User:     env("DB_USER", "postgres"),
		Password: os.Getenv("DB_PASS"),
		Name:     env("DB_NAME", "url_shortener"),
		SSLMode:  env("DB_SSLMODE", "disable"),
	}
}

// env returns the value of the environment variable, or fallback when unset.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Connect opens a connection to PostgreSQL and verifies it is reachable.
func Connect(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)
	db, err := sql.Open("postgres", dsn)
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

// InsertMapping stores a new long URL -> short alias mapping. A nil
// m.ExpiresAt means the link never expires. It returns ErrAliasTaken when the
// short alias already exists.
func InsertMapping(ctx context.Context, db *sql.DB, m model.URLMapping) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO url_mappings (long_url, short_url, expires_at) VALUES ($1, $2, $3)",
		m.LongURL, m.ShortURL, m.ExpiresAt,
	)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrAliasTaken
	}
	return err
}

// RecordClick increments the click counter for the given short alias, stamps
// the access time, and returns the mapped long URL in a single atomic update.
// Expired aliases are treated as absent. It returns sql.ErrNoRows when the
// alias does not exist or has expired.
func RecordClick(ctx context.Context, db *sql.DB, shortURL string) (string, error) {
	var longURL string
	err := db.QueryRowContext(ctx,
		"UPDATE url_mappings SET clicks = clicks + 1, last_accessed_at = now() WHERE short_url = $1 AND (expires_at IS NULL OR expires_at > now()) RETURNING long_url",
		shortURL,
	).Scan(&longURL)
	return longURL, err
}

// DeleteByShortURL removes the mapping for the given short alias, reporting
// whether a row was actually deleted.
func DeleteByShortURL(ctx context.Context, db *sql.DB, shortURL string) (bool, error) {
	res, err := db.ExecContext(ctx, "DELETE FROM url_mappings WHERE short_url = $1", shortURL)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// StatsByShortURL returns the mapping and its usage statistics for the given
// short alias. It returns sql.ErrNoRows when the alias does not exist.
func StatsByShortURL(ctx context.Context, db *sql.DB, shortURL string) (model.URLMapping, error) {
	m := model.URLMapping{ShortURL: shortURL}
	err := db.QueryRowContext(ctx,
		"SELECT long_url, clicks, created_at, last_accessed_at, expires_at FROM url_mappings WHERE short_url = $1",
		shortURL,
	).Scan(&m.LongURL, &m.Clicks, &m.CreatedAt, &m.LastAccessedAt, &m.ExpiresAt)
	return m, err
}
