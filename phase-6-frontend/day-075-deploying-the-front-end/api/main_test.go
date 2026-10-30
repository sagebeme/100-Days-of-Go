package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "assets"), 0o755)
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html><title>Twende</title>"), 0o644)
	os.WriteFile(filepath.Join(dir, "assets", "index-abc.js"), []byte("console.log(1)"), 0o644)
	os.WriteFile(filepath.Join(dir, "robots.txt"), []byte("User-agent: *"), 0o644)
	env := map[string]string{"API_URL": "https://api.twende.example", "SEAT_MAP": "true", "JWT_SECRET": "never-shown"}
	h := newHandler(dir, func(k string) string { return env[k] })

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/", 200, "<title>Twende</title>"},
		{"/bookings/K7Q2M9XA", 200, "<title>Twende</title>"},
		{"/assets/index-abc.js", 200, "console.log"},
		{"/assets/missing-123.js", 404, "not found"},
		{"/robots.txt", 200, "User-agent"},
		{"/config.json", 200, `{"apiUrl":"https://api.twende.example","seatMap":true}`},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", tt.path, nil))
		if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), tt.wantBody) {
			t.Errorf("GET %s: %d %q, want %d containing %q", tt.path, rec.Code, rec.Body, tt.wantStatus, tt.wantBody)
		}
		if strings.Contains(rec.Body.String(), "never-shown") {
			t.Errorf("GET %s leaked a secret", tt.path)
		}
	}
}
