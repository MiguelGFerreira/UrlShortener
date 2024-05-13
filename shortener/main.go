package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Import PostgreSQL driver
)

func main() {
	DBUSER := godotenv.Load("DB_USER")
	DBPASS := godotenv.Load("DB_PASS")
	// Conect to PostgreSQL
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=url_shortener sslmode=disable", DBUSER, DBPASS))
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
		var longURL struct {
			LongURL string `json:"long_url"`
		}
		err := json.NewDecoder(r.Body).Decode(&longURL)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Erro: %v", err)
			return
		}

		// Generate a random short alias (6 characters)
		shortURL := generateRandomString(6)

		// Check if the short alias already exists in the database
		exists, err := checkShortURLExists(db, shortURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		for exists {
			shortURL = generateRandomString(6)
			exists, err = checkShortURLExists(db, shortURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %v", err)
				return
			}
		}

		// Insert the long URL -> short alias mapping into the database
		err = insertURLMapping(db, longURL.LongURL, shortURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		// Create the response with the short alias
		response := struct {
			ShortURL string `json:"short_url"`
		}{ShortURL: fmt.Sprintf("http://localhost:8080/redirect/%s", shortURL)}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Start HTTP server
	http.HandleFunc("/shorten", shortenHandler)
	fmt.Println("Shortening server running on port 8080")
	http.ListenAndServe(":8080", nil)
}

// Function to generate random string
func generateRandomString(length int) string {
	var chars []rune
	for i := 'a'; i <= 'z'; i++ {
		chars = append(chars, i)
	}
	for i := 'A'; i <= 'Z'; i++ {
		chars = append(chars, i)
	}
	for i := 0; i <= 9; i++ {
		chars = append(chars, rune(i+48))
	}

	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = byte(chars[rand.Intn(len(chars))])
	}
	return string(bytes)
}

// Function to check if the short alias already exists in the database
func checkShortURLExists(db *sql.DB, shortURL string) (bool, error) {
	ctx := context.Background()
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings WHERE short_url = $1", shortURL).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Function to insert long URL -> short alias mapping into the database
func insertURLMapping(db *sql.DB, longURL string, shortURL string) error {
	ctx := context.Background()
	_, err := db.ExecContext(ctx, "INSERT INTO url_mappings (long_url, short_url) VALUES ($1, $2)", longURL, shortURL)
	return err
}
