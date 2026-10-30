package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCacheControl(t *testing.T) {
	tests := []struct{ path, want string }{
		{"/assets/index-B3kx9a.js", "public, max-age=31536000, immutable"},
		{"/assets/SeatMap-Cc81d0.js", "public, max-age=31536000, immutable"},
		{"/", "no-cache"},
		{"/index.html", "no-cache"},
		{"/favicon.ico", "public, max-age=3600"},
	}
	for _, tt := range tests {
		if got := CacheControl(tt.path); got != tt.want {
			t.Errorf("CacheControl(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestServesWithHeaders(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "assets"), 0o755)
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html>"), 0o644)
	os.WriteFile(filepath.Join(dir, "assets", "index-abc123.js"), []byte("console.log(1)"), 0o644)
	h := newHandler(dir)

	for _, tt := range []struct{ path, want string }{
		{"/", "no-cache"},
		{"/assets/index-abc123.js", "public, max-age=31536000, immutable"},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", tt.path, nil))
		if rec.Code != 200 || rec.Header().Get("Cache-Control") != tt.want {
			t.Errorf("GET %s: %d %q, want 200 %q", tt.path, rec.Code, rec.Header().Get("Cache-Control"), tt.want)
		}
	}

	// A browser re-checking the page sends If-Modified-Since and gets an empty 304.
	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest("GET", "/", nil))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("If-Modified-Since", first.Header().Get("Last-Modified"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 304 || rec.Body.Len() != 0 {
		t.Errorf("re-check: got %d with %d bytes, want an empty 304", rec.Code, rec.Body.Len())
	}
}
