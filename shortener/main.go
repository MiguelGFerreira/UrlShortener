package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"UrlShortener/internal/health"
	"UrlShortener/internal/model"
	"UrlShortener/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	// Load variables from the .env file (if present) into the environment
	godotenv.Load()

	db, err := store.Connect(store.ConfigFromEnv())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Handler to shorten URLs
	shortenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Decode the request body: the long URL plus an optional custom alias
		var body struct {
			LongURL string `json:"long_url"`
			Alias   string `json:"alias"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		// Validate the submitted URL before storing it
		longURL := strings.TrimSpace(body.LongURL)
		if !validLongURL(longURL) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Error: long_url must be a valid http or https URL")
			return
		}

		ctx := r.Context()

		// Choose the alias: the caller's custom one when provided, otherwise a
		// freshly generated unique one.
		var shortURL string
		if alias := strings.TrimSpace(body.Alias); alias != "" {
			if !validAlias(alias) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "Error: alias must be 3-16 chars of letters, digits, '-' or '_'")
				return
			}
			taken, err := store.ShortURLExists(ctx, db, alias)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %v", err)
				return
			}
			if taken {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprintln(w, "Error: alias already taken")
				return
			}
			shortURL = alias
		} else {
			var err error
			shortURL, err = uniqueRandomAlias(ctx, db)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %v", err)
				return
			}
		}

		// Insert the long URL -> short alias mapping into the database
		mapping := model.URLMapping{LongURL: longURL, ShortURL: shortURL}
		if err := store.InsertMapping(ctx, db, mapping); err != nil {
			if errors.Is(err, store.ErrAliasTaken) {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprintln(w, "Error: alias already taken")
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		// Create the response with the short alias
		response := struct {
			ShortURL string `json:"short_url"`
		}{ShortURL: fmt.Sprintf("http://localhost:8081/redirect/%s", shortURL)}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Handler to report usage statistics for a short alias
	statsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		alias := strings.TrimSpace(r.URL.Path[len("/stats/"):])
		if alias == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		m, err := store.StatsByShortURL(r.Context(), db, alias)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		response := struct {
			ShortURL       string     `json:"short_url"`
			LongURL        string     `json:"long_url"`
			Clicks         int64      `json:"clicks"`
			CreatedAt      time.Time  `json:"created_at"`
			LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
		}{
			ShortURL:       alias,
			LongURL:        m.LongURL,
			Clicks:         m.Clicks,
			CreatedAt:      m.CreatedAt,
			LastAccessedAt: m.LastAccessedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Start HTTP server
	http.HandleFunc("/health", health.Handler(db))
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/stats/", statsHandler)
	fmt.Println("Shortening server running on port 8080")
	http.ListenAndServe(":8080", nil)
}

// validLongURL reports whether raw is a well-formed absolute http(s) URL.
func validLongURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

// aliasPattern matches an acceptable custom short alias.
var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,16}$`)

// validAlias reports whether s is an acceptable custom short alias.
func validAlias(s string) bool {
	return aliasPattern.MatchString(s)
}

// uniqueRandomAlias generates random 6-char aliases until it finds one that is
// not already present in the database.
func uniqueRandomAlias(ctx context.Context, db *sql.DB) (string, error) {
	for {
		alias, err := generateRandomString(6)
		if err != nil {
			return "", err
		}
		exists, err := store.ShortURLExists(ctx, db, alias)
		if err != nil {
			return "", err
		}
		if !exists {
			return alias, nil
		}
	}
}

// generateRandomString returns a cryptographically random alphanumeric string.
func generateRandomString(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = chars[int(bytes[i])%len(chars)]
	}
	return string(bytes), nil
}
