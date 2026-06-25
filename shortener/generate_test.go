package main

import (
	"strings"
	"testing"
)

const allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func TestGenerateRandomString(t *testing.T) {
	const length = 6

	got, err := generateRandomString(length)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != length {
		t.Errorf("length = %d, want %d", len(got), length)
	}
	for _, c := range got {
		if !strings.ContainsRune(allowedChars, c) {
			t.Errorf("unexpected character %q in %q", c, got)
		}
	}
}

func TestGenerateRandomStringUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		s, err := generateRandomString(8)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[s] {
			t.Fatalf("duplicate string generated: %q", s)
		}
		seen[s] = true
	}
}
