package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"UrlShortener/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
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
	mock.ExpectExec(`INSERT INTO url_mappings \(long_url, short_url\) VALUES \(\$1, \$2\)`).
		WithArgs("https://example.com", "abc123").
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

func TestMappingByShortURL(t *testing.T) {
	db, mock := newMockDB(t)
	rows := sqlmock.NewRows([]string{"long_url"}).AddRow("https://example.com")
	mock.ExpectQuery(`SELECT long_url FROM url_mappings WHERE short_url = \$1`).
		WithArgs("abc123").
		WillReturnRows(rows)

	got, err := MappingByShortURL(context.Background(), db, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LongURL != "https://example.com" {
		t.Errorf("LongURL = %q, want %q", got.LongURL, "https://example.com")
	}
	if got.ShortURL != "abc123" {
		t.Errorf("ShortURL = %q, want %q", got.ShortURL, "abc123")
	}
}

func TestMappingByShortURLNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectQuery(`SELECT long_url FROM url_mappings WHERE short_url = \$1`).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	_, err := MappingByShortURL(context.Background(), db, "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}
