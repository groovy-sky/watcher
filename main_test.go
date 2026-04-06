package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVisitURLs_VisitsAndHandlesErrors(t *testing.T) {
	// Serve a minimal HTML page so chromedp can read a title.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><head><title>Test Page</title></head><body>Hello</body></html>`)
	}))
	defer srv.Close()

	// Also include an unreachable URL to verify error-handling does not panic.
	urls := []string{srv.URL, "http://127.0.0.1:1"}

	if err := visitURLs(urls); err != nil {
		t.Fatalf("visitURLs returned unexpected error: %v", err)
	}
}

func TestFetchURLs_ParsesValidFile(t *testing.T) {
	body := `# comment line
https://www.example.com
https://www.golang.org

https://www.github.com
# another comment
https://www.stackoverflow.com
`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	urls, err := fetchURLs(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{
		"https://www.example.com",
		"https://www.golang.org",
		"https://www.github.com",
		"https://www.stackoverflow.com",
	}
	if len(urls) != len(want) {
		t.Fatalf("got %d URLs, want %d: %v", len(urls), len(want), urls)
	}
	for i, u := range urls {
		if u != want[i] {
			t.Errorf("url[%d] = %q, want %q", i, u, want[i])
		}
	}
}

func TestFetchURLs_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := fetchURLs(srv.URL)
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

func TestFetchURLs_SkipsEmptyAndCommentLines(t *testing.T) {
	body := "\n# skip me\n  \nhttps://www.test.com\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	urls, err := fetchURLs(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) != 1 || urls[0] != "https://www.test.com" {
		t.Errorf("unexpected urls: %v", urls)
	}
}

func TestSelectRandom_ReturnsCorrectCount(t *testing.T) {
	input := []string{
		"https://a.com",
		"https://b.com",
		"https://c.com",
		"https://d.com",
		"https://e.com",
	}
	got := selectRandom(input, 3)
	if len(got) != 3 {
		t.Fatalf("want 3 URLs, got %d", len(got))
	}
}

func TestSelectRandom_NoDuplicates(t *testing.T) {
	input := []string{
		"https://a.com",
		"https://b.com",
		"https://c.com",
		"https://d.com",
		"https://e.com",
	}
	for iter := 0; iter < 50; iter++ {
		got := selectRandom(input, 3)
		seen := make(map[string]bool)
		for _, u := range got {
			if seen[u] {
				t.Fatalf("duplicate URL %q in result %v", u, got)
			}
			seen[u] = true
		}
	}
}

func TestSelectRandom_OnlyUsesInputURLs(t *testing.T) {
	input := []string{
		"https://a.com",
		"https://b.com",
		"https://c.com",
	}
	inputSet := make(map[string]bool)
	for _, u := range input {
		inputSet[u] = true
	}
	for iter := 0; iter < 50; iter++ {
		got := selectRandom(input, 3)
		for _, u := range got {
			if !inputSet[u] {
				t.Errorf("got unexpected URL %q not in input", u)
			}
		}
	}
}
