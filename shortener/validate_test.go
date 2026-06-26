package main

import "testing"

func TestValidLongURL(t *testing.T) {
	valid := []string{
		"http://example.com",
		"https://example.com/path?q=1",
		"https://sub.domain.co/a/b",
	}
	for _, u := range valid {
		if !validLongURL(u) {
			t.Errorf("validLongURL(%q) = false, want true", u)
		}
	}

	invalid := []string{
		"",
		"example.com",         // missing scheme
		"ftp://example.com",   // unsupported scheme
		"javascript:alert(1)", // unsupported scheme
		"http://",             // missing host
		"not a url",
	}
	for _, u := range invalid {
		if validLongURL(u) {
			t.Errorf("validLongURL(%q) = true, want false", u)
		}
	}
}
