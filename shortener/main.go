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
	"os"
	"regexp"
	"strings"
	"time"

	"UrlShortener/internal/cors"
	"UrlShortener/internal/health"
	"UrlShortener/internal/model"
	"UrlShortener/internal/ratelimit"
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

	// Public base for generated short links (the redirector's public address)
	publicBase := os.Getenv("PUBLIC_BASE_URL")
	if publicBase == "" {
		publicBase = "http://localhost:8081/redirect"
	}
	publicBase = strings.TrimRight(publicBase, "/")

	// Origin allowed to call the API from a browser ("*" = any)
	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	// JSON API: shorten a URL
	shortenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			LongURL   string `json:"long_url"`
			Alias     string `json:"alias"`
			ExpiresIn int    `json:"expires_in"` // seconds from now; 0 = never
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		shortURL, err := createMapping(r.Context(), db, body.LongURL, body.Alias, body.ExpiresIn)
		if err != nil {
			var ce createError
			if errors.As(err, &ce) {
				w.WriteHeader(ce.status)
				fmt.Fprintln(w, "Error: "+ce.msg)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		response := struct {
			ShortURL string `json:"short_url"`
		}{ShortURL: publicBase + "/" + shortURL}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Handler to delete a short alias
	deleteHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		alias := strings.TrimSpace(r.URL.Path[len("/shorten/"):])
		if alias == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		deleted, err := store.DeleteByShortURL(r.Context(), db, alias)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		if !deleted {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
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
			ExpiresAt      *time.Time `json:"expires_at,omitempty"`
		}{
			ShortURL:       alias,
			LongURL:        m.LongURL,
			Clicks:         m.Clicks,
			CreatedAt:      m.CreatedAt,
			LastAccessedAt: m.LastAccessedAt,
			ExpiresAt:      m.ExpiresAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Rate limit the creation endpoint per client IP (5 req/s, burst of 10)
	limiter := ratelimit.New(5, 10)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.Handler(db))
	mux.Handle("/shorten", limiter.Middleware(shortenHandler))
	mux.HandleFunc("/shorten/", deleteHandler)
	mux.HandleFunc("/stats/", statsHandler)

	fmt.Println("Shortening server running on port 8080")
	http.ListenAndServe(":8080", cors.Middleware(allowedOrigin, mux))
}

// createError carries an HTTP status alongside a client-facing message.
type createError struct {
	status int
	msg    string
}

func (e createError) Error() string { return e.msg }

// createMapping validates the inputs, picks an alias (custom or random) and
// stores the mapping, returning the short alias. Validation problems are
// reported as createError values carrying the appropriate HTTP status.
func createMapping(ctx context.Context, db *sql.DB, rawURL, rawAlias string, expiresIn int) (string, error) {
	longURL := strings.TrimSpace(rawURL)
	if !validLongURL(longURL) {
		return "", createError{http.StatusBadRequest, "long_url must be a valid http or https URL"}
	}

	if expiresIn < 0 {
		return "", createError{http.StatusBadRequest, "expires_in must be a non-negative number of seconds"}
	}
	var expiresAt *time.Time
	if expiresIn > 0 {
		t := time.Now().Add(time.Duration(expiresIn) * time.Second)
		expiresAt = &t
	}

	var shortURL string
	if alias := strings.TrimSpace(rawAlias); alias != "" {
		if !validAlias(alias) {
			return "", createError{http.StatusBadRequest, "alias must be 3-16 chars of letters, digits, '-' or '_'"}
		}
		taken, err := store.ShortURLExists(ctx, db, alias)
		if err != nil {
			return "", err
		}
		if taken {
			return "", createError{http.StatusConflict, "alias already taken"}
		}
		shortURL = alias
	} else {
		var err error
		shortURL, err = uniqueRandomAlias(ctx, db)
		if err != nil {
			return "", err
		}
	}

	mapping := model.URLMapping{LongURL: longURL, ShortURL: shortURL, ExpiresAt: expiresAt}
	if err := store.InsertMapping(ctx, db, mapping); err != nil {
		if errors.Is(err, store.ErrAliasTaken) {
			return "", createError{http.StatusConflict, "alias already taken"}
		}
		return "", err
	}
	return shortURL, nil
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
