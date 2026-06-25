package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"UrlShortener/internal/model"
	"UrlShortener/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	// Load variables from the .env file (if present) into the environment
	godotenv.Load()

	db, err := store.Connect(os.Getenv("DB_USER"), os.Getenv("DB_PASS"))
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

		// Decode the request body to get the long URL
		var body struct {
			LongURL string `json:"long_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		ctx := r.Context()

		// Generate a unique random short alias (6 characters)
		shortURL, err := generateRandomString(6)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		exists, err := store.ShortURLExists(ctx, db, shortURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		for exists {
			shortURL, err = generateRandomString(6)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %v", err)
				return
			}
			exists, err = store.ShortURLExists(ctx, db, shortURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %v", err)
				return
			}
		}

		// Insert the long URL -> short alias mapping into the database
		mapping := model.URLMapping{LongURL: body.LongURL, ShortURL: shortURL}
		if err := store.InsertMapping(ctx, db, mapping); err != nil {
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

	// Start HTTP server
	http.HandleFunc("/shorten", shortenHandler)
	fmt.Println("Shortening server running on port 8080")
	http.ListenAndServe(":8080", nil)
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
