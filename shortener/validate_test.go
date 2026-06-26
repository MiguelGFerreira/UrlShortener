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

func TestValidAlias(t *testing.T) {
	valid := []string{"abc", "my-link_1", "ABCdef123", "a1b2c3"}
	for _, a := range valid {
		if !validAlias(a) {
			t.Errorf("validAlias(%q) = false, want true", a)
		}
	}

	invalid := []string{
		"ab",                         // too short
		"this-alias-is-way-too-long", // longer than 16
		"has space",
		"slash/here",
		"dot.dot",
		"",
	}
	for _, a := range invalid {
		if validAlias(a) {
			t.Errorf("validAlias(%q) = true, want false", a)
		}
	}
}
