package main

import "testing"

func TestExtractEmailAddr(t *testing.T) {
	tests := []struct {
		email     string
		extracted string
	}{
		{"example@example.com", "example@example.com"},
		{"<example@example.com>", "example@example.com"},
		{"Example Email <example@example.com>", "example@example.com"},
	}
	for _, test := range tests {
		if got, want := extractEmailAddr(test.email), test.extracted; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
