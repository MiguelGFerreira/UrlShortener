package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"UrlShortener/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestShortURLExists(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  bool
	}{
		{"exists", 1, true},
		{"missing", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			rows := sqlmock.NewRows([]string{"count"}).AddRow(tt.count)
			mock.ExpectQuery(`SELECT COUNT\(\*\) FROM url_mappings WHERE short_url = \$1`).
				WithArgs("abc123").
				WillReturnRows(rows)

			got, err := ShortURLExists(context.Background(), db, "abc123")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ShortURLExists() = %v, want %v", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestInsertMapping(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec(`INSERT INTO url_mappings \(long_url, short_url, expires_at\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs("https://example.com", "abc123", nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := InsertMapping(context.Background(), db, model.URLMapping{
		LongURL:  "https://example.com",
		ShortURL: "abc123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestInsertMappingAliasTaken(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec(`INSERT INTO url_mappings`).
		WithArgs("https://example.com", "taken", nil).
		WillReturnError(&pq.Error{Code: "23505"})

	err := InsertMapping(context.Background(), db, model.URLMapping{
		LongURL:  "https://example.com",
		ShortURL: "taken",
	})
	if !errors.Is(err, ErrAliasTaken) {
		t.Fatalf("expected ErrAliasTaken, got %v", err)
	}
}

func TestRecordClick(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"long_url"}).AddRow("https://example.com")
	mock.ExpectQuery(`UPDATE url_mappings SET clicks = clicks \+ 1`).
		WithArgs("abc123").
		WillReturnRows(rows)

	got, err := RecordClick(context.Background(), db, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com" {
		t.Errorf("long_url = %q, want %q", got, "https://example.com")
	}
}

func TestRecordClickNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectQuery(`UPDATE url_mappings SET clicks = clicks \+ 1`).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	_, err := RecordClick(context.Background(), db, "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestStatsByShortURL(t *testing.T) {
	db, mock := newMockDB(t)
	created := time.Now()
	rows := sqlmock.NewRows([]string{"long_url", "clicks", "created_at", "last_accessed_at", "expires_at"}).
		AddRow("https://example.com", int64(7), created, nil, nil)
	mock.ExpectQuery(`SELECT long_url, clicks, created_at, last_accessed_at, expires_at FROM url_mappings WHERE short_url = \$1`).
		WithArgs("abc123").
		WillReturnRows(rows)

	m, err := StatsByShortURL(context.Background(), db, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Clicks != 7 {
		t.Errorf("clicks = %d, want 7", m.Clicks)
	}
	if m.LongURL != "https://example.com" {
		t.Errorf("long_url = %q, want %q", m.LongURL, "https://example.com")
	}
	if m.LastAccessedAt != nil {
		t.Errorf("last_accessed_at = %v, want nil", m.LastAccessedAt)
	}
	if m.ExpiresAt != nil {
		t.Errorf("expires_at = %v, want nil", m.ExpiresAt)
	}
}

func TestDeleteByShortURL(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec(`DELETE FROM url_mappings WHERE short_url = \$1`).
		WithArgs("abc123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	deleted, err := DeleteByShortURL(context.Background(), db, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("deleted = false, want true")
	}
}

func TestDeleteByShortURLMissing(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectExec(`DELETE FROM url_mappings WHERE short_url = \$1`).
		WithArgs("missing").
		WillReturnResult(sqlmock.NewResult(0, 0))

	deleted, err := DeleteByShortURL(context.Background(), db, "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Error("deleted = true, want false")
	}
}
