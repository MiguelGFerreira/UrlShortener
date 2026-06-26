// Package health provides a shared HTTP health-check handler.
package health

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

// Handler returns an http.HandlerFunc that reports 200 when the database is
// reachable and 503 otherwise.
func Handler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "unhealthy")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}
}
