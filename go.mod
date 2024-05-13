module UrlShortener

go 1.18

require github.com/lib/pq v1.10.9 // Import PostgreSQL driver

require github.com/joho/godotenv v1.5.1 // indirect

replace github.com/lib/pq => github.com/lib/pq v1.10.4 // Substitute PostgreSQL driver if needed
