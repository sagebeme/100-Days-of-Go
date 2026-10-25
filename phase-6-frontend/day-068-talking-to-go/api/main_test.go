package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	h := newHandler("http://localhost:5173", 0)
	tests := []struct {
		name       string
		url        string
		origin     string
		wantStatus int
		wantCount  int
		wantACAO   string
	}{
		{"all routes, allowed origin", "/api/routes", "http://localhost:5173", 200, 6, "http://localhost:5173"},
		{"search is case-insensitive", "/api/routes?q=NGONG", "http://localhost:5173", 200, 1, "http://localhost:5173"},
		{"no match is an empty list", "/api/routes?q=mombasa", "", 200, 0, ""},
		{"other origin gets no CORS header", "/api/routes", "https://evil.example", 200, 6, ""},
		{"forced failure", "/api/routes?fail=1", "http://localhost:5173", 500, -1, "http://localhost:5173"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantACAO {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantACAO)
			}
			if tt.wantCount < 0 {
				return
			}
			var got []Route
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("body %q: %v", rec.Body, err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("got %d routes, want %d", len(got), tt.wantCount)
			}
		})
	}
}
