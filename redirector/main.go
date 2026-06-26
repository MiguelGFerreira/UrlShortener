package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"UrlShortener/internal/health"
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

	// Handler to redirect URLs
	redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		shortURL := r.URL.Path[len("/redirect/"):]

		// Resolve the alias, counting this visit, and redirect to the long URL
		longURL, err := store.RecordClick(r.Context(), db, shortURL)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		// Redirect user to long URL
		http.Redirect(w, r, longURL, http.StatusMovedPermanently)
	})

	// Start HTTP server
	http.HandleFunc("/health", health.Handler(db))
	http.HandleFunc("/redirect/", redirectHandler)
	fmt.Println("Redirect server running on port 8081")
	http.ListenAndServe(":8081", nil)
}
