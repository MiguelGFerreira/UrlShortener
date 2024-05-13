package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	_"github.com/lib/pq" // Import PostgreSQL driver
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

	// Criar handler para encurtar URLs
	shortenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Decodificar o corpo da requisição para obter a URL longa
		var longURL struct {
			LongURL string `json:"long_url"`
		}
		err := json.NewDecoder(r.Body).Decode(&longURL)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Erro: %v", err)
			return
		}

		// Gerar um alias curto aleatório (6 caracteres)
		shortURL := generateRandomString(6)

		// Verificar se o alias curto já existe no banco de dados
		exists, err := checkShortURLExists(db, shortURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Erro: %v", err)
			return
		}
		for exists {
			shortURL = generateRandomString(6)
			exists, err = checkShortURLExists(db, shortURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Erro: %v", err)
				return
			}
		}

		// Inserir o mapeamento URL longa -> alias curto no banco de dados
		err = insertURLMapping(db, longURL.LongURL, shortURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Erro: %v", err)
			return
		}

		// Criar a resposta com o alias curto
		response := struct {
			ShortURL string `json:"short_url"`
		}{ShortURL: fmt.Sprintf("http://localhost:8080/redirect/%s", shortURL)}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Iniciar servidor HTTP
	http.HandleFunc("/shorten", shortenHandler)
	fmt.Println("Servidor de encurtamento em execução na porta 8080")
	http.ListenAndServe(":8080", nil)
}

// Função para gerar string aleatória (6 caracteres)
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

// Função para verificar se o alias curto já existe no banco de dados
func checkShortURLExists(db *sql.DB, shortURL string) (bool, error) {
	ctx := context.Background()
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings WHERE short_url = $1", shortURL).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Função para inserir mapeamento URL longa -> alias curto no banco de dados
func insertURLMapping(db *sql.DB, longURL string, shortURL string) error {
	ctx := context.Background()
	_, err := db.ExecContext(ctx, "INSERT INTO url_mappings (long_url, short_url) VALUES ($1, $2)", longURL, shortURL)
	return err
}
