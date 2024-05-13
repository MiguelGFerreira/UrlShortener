module UrlShortener

go 1.18

require github.com/lib/pq v1.10.9 // Importar driver PostgreSQL

replace github.com/lib/pq => github.com/lib/pq v1.10.4 // Substituir driver PostgreSQL se necessário
