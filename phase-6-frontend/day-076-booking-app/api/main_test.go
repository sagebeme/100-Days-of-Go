package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBookingFlow(t *testing.T) {
	h := newServer(time.Minute).routes("http://localhost:5173")
	do := func(method, path, body, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	rec := do("GET", "/api/trips", "", "")
	var trips []Trip
	json.Unmarshal(rec.Body.Bytes(), &trips)
	if rec.Code != 200 || len(trips) != 5 || trips[0].SeatsLeft != 3 || trips[1].SeatsLeft != 0 {
		t.Fatalf("trips: %d %s", rec.Code, rec.Body)
	}
	if rec := do("GET", "/api/trips/T9/seats", "", ""); rec.Code != 404 {
		t.Errorf("seats for a missing trip: got %d, want 404", rec.Code)
	}

	booking := `{"tripId":"T1","seat":3,"name":"Amina","phone":"0712 345 678"}`
	if rec := do("POST", "/api/bookings", booking, ""); rec.Code != 401 {
		t.Errorf("booking without a token: got %d, want 401", rec.Code)
	}
	rec = do("POST", "/api/login", `{"phone":"0712345678","password":"matatu-2026"}`, "")
	var login struct{ AccessToken string }
	json.Unmarshal(rec.Body.Bytes(), &login)
	if rec.Code != 200 || login.AccessToken == "" {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}

	tests := []struct {
		name, body string
		wantStatus int
		wantBody   string
	}{
		{"good booking", booking, 201, `"totalKsh":150`},
		{"same seat again", booking, 409, "Seat 3 was just taken"},
		{"bad fields", `{"tripId":"T1","seat":0,"name":"","phone":"12"}`, 422, `"seat":"Choose a seat"`},
		{"no such trip", `{"tripId":"T9","seat":1,"name":"Amina","phone":"0712345678"}`, 404, "no such trip"},
	}
	for _, tt := range tests {
		rec := do("POST", "/api/bookings", tt.body, login.AccessToken)
		if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), tt.wantBody) {
			t.Errorf("%s: got %d %s, want %d containing %s", tt.name, rec.Code, rec.Body, tt.wantStatus, tt.wantBody)
		}
	}
	if rec := do("GET", "/api/trips/T1/seats", "", ""); !strings.Contains(rec.Body.String(), `"taken":[1,2,3,4,`) {
		t.Errorf("seat 3 should now be taken: %s", rec.Body)
	}
}
