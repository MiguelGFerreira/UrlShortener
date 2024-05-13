package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq" // Importar driver PostgreSQL
)

func main() {
	DBUSER := os.Getenv("DB_USER")
	DBPASS := os.Getenv("DB_PASS")
	// Conectar ao banco de dados PostgreSQL
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=url_shortener sslmode=disable",DBUSER,DBPASS))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Criar handler para redirecionar URLs
	redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		shortURL := r.URL.Path[len("/redirect/"):]

		// Buscar a URL longa correspondente ao alias curto no banco de dados
		longURL, err := getLongURLByShortURL(db, shortURL)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Erro: %v", err)
			return
		}

		// Redirecionar o usuário para a URL longa
		http.Redirect(w, r, longURL, http.StatusMovedPermanently)
	})

	// Iniciar servidor HTTP
	http.HandleFunc("/redirect/", redirectHandler)
	fmt.Println("Servidor de redirecionamento em execução na porta 8081")
	http.ListenAndServe(":8081", nil)
}

// Função para buscar a URL longa pelo alias curto
func getLongURLByShortURL(db *sql.DB, shortURL string) (string, error) {
	ctx := context.Background()
	var longURL string
	err := db.QueryRowContext(ctx, "SELECT long_url FROM url_mappings WHERE short_url = $1", shortURL).Scan(&longURL)
	return longURL, err
}
