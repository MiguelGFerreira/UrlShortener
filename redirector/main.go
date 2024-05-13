package main

import (
	"context"
	"database/sql"
	"fmt"
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

	// Handler to redirect URLs
	redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		shortURL := r.URL.Path[len("/redirect/"):]

		// Fetch the long URL corresponding to the short alias in the database
		longURL, err := getLongURLByShortURL(db, shortURL)
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
	http.HandleFunc("/redirect/", redirectHandler)
	fmt.Println("Redirect server running on port 8081")
	http.ListenAndServe(":8081", nil)
}

// Function to search long URL by short alias
func getLongURLByShortURL(db *sql.DB, shortURL string) (string, error) {
	ctx := context.Background()
	var longURL string
	err := db.QueryRowContext(ctx, "SELECT long_url FROM url_mappings WHERE short_url = $1", shortURL).Scan(&longURL)
	return longURL, err
}
