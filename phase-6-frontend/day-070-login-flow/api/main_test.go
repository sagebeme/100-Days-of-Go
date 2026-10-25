package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoginRefreshLogout(t *testing.T) {
	s := newServer(time.Minute)
	clock := time.Date(2026, 10, 25, 8, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return clock }
	h := s.routes("http://localhost:5173")

	do := func(method, path, body string, cookie *http.Cookie, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if cookie != nil {
			req.AddCookie(cookie)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	parse := func(rec *httptest.ResponseRecorder) (string, *http.Cookie) {
		var body struct{ AccessToken string }
		json.Unmarshal(rec.Body.Bytes(), &body)
		var c *http.Cookie
		for _, ck := range rec.Result().Cookies() {
			if ck.Name == "refresh" {
				c = ck
			}
		}
		return body.AccessToken, c
	}

	if rec := do("POST", "/api/login", `{"phone":"0712345678","password":"nope"}`, nil, ""); rec.Code != 401 {
		t.Fatalf("wrong password: got %d, want 401", rec.Code)
	}
	rec := do("POST", "/api/login", `{"phone":"0712345678","password":"matatu-2026"}`, nil, "")
	if rec.Code != 200 {
		t.Fatalf("login: got %d %s", rec.Code, rec.Body)
	}
	access, refresh := parse(rec)
	if refresh == nil || !refresh.HttpOnly || !refresh.Secure || refresh.SameSite != http.SameSiteStrictMode {
		t.Fatalf("refresh cookie = %+v, want HttpOnly, Secure, SameSite=Strict", refresh)
	}
	if got := do("GET", "/api/me/bookings", "", nil, access).Code; got != 200 {
		t.Errorf("bookings with a fresh token: got %d, want 200", got)
	}
	clock = clock.Add(2 * time.Minute)
	if got := do("GET", "/api/me/bookings", "", nil, access).Code; got != 401 {
		t.Errorf("bookings with an expired token: got %d, want 401", got)
	}
	rec = do("POST", "/api/refresh", "", refresh, "")
	newAccess, newRefresh := parse(rec)
	if rec.Code != 200 || newAccess == "" {
		t.Fatalf("refresh: got %d %s", rec.Code, rec.Body)
	}
	if got := do("POST", "/api/refresh", "", refresh, "").Code; got != 401 {
		t.Errorf("reusing a rotated refresh token: got %d, want 401", got)
	}
	do("POST", "/api/logout", "", newRefresh, "")
	if got := do("POST", "/api/refresh", "", newRefresh, "").Code; got != 401 {
		t.Errorf("refresh after logout: got %d, want 401", got)
	}
}
