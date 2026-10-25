package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalisePhone(t *testing.T) {
	tests := []struct {
		in, want string
		ok       bool
	}{
		{"0712345678", "254712345678", true},
		{"0712 345 678", "254712345678", true},
		{"+254712345678", "254712345678", true},
		{"254-110-123456", "254110123456", true},
		{"0812345678", "", false},
		{"071234567", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := NormalisePhone(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Errorf("NormalisePhone(%q) = %q, %v; want %q, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestBookings(t *testing.T) {
	origin := "http://localhost:5173"
	tests := []struct {
		name, body string
		wantStatus int
		wantBody   string
	}{
		{"good booking", `{"name":"Amina","phone":"0712345678","route":"58","seats":2}`, 201, `"phone":"254712345678"`},
		{"every field wrong", `{"name":" ","phone":"12","route":"999","seats":0}`, 422, `"route":"Choose a route"`},
		{"more seats than left", `{"name":"Kip","phone":"0712345678","route":"111","seats":4}`, 422, `Only 3 seats left on 111`},
		{"not JSON", `nope`, 400, `"error"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(origin, newStore())
			req := httptest.NewRequest("POST", "/api/bookings", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("got %d %s, want %d containing %s", rec.Code, rec.Body, tt.wantStatus, tt.wantBody)
			}
		})
	}
}

func TestPreflight(t *testing.T) {
	h := newHandler("http://localhost:5173", newStore())
	req := httptest.NewRequest("OPTIONS", "/api/bookings", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 204 || rec.Header().Get("Access-Control-Allow-Headers") != "Content-Type" {
		t.Errorf("preflight got %d %v", rec.Code, rec.Header())
	}
}
