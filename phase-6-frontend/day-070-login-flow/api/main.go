// Command api is Day 70's login API. Run it from the repo root:
//
//	go run ./phase-6-frontend/day-070-login-flow/api
//
// Log in as 0712345678 with the password matatu-2026 (a demo account, made up).
//
//	POST /api/login         {"phone","password"} → 200 {"accessToken","user"} + refresh cookie
//	POST /api/refresh       refresh cookie → 200 {"accessToken","user"} + a new refresh cookie, or 401
//	POST /api/logout        forgets the refresh token and clears the cookie → 204
//	GET  /api/me/bookings   Authorization: Bearer <access token> → 200 [...], or 401
//
// Environment: ADDR (default :8080), ALLOWED_ORIGIN (default http://localhost:5173),
// ACCESS_TTL (default 1m, short so you can watch the page refresh its token).
package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User is who is logged in. The JSON names match the TypeScript User.
type User struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

type account struct {
	user User
	hash []byte
}

type session struct {
	phone   string
	expires time.Time
}

// server keeps everything in memory. A real one would use Postgres or Redis (Days 55 and 58).
type server struct {
	mu        sync.Mutex
	accounts  map[string]account
	access    map[string]session // access token → who, until when
	refresh   map[string]string  // refresh token → phone
	accessTTL time.Duration
	now       func() time.Time
}

func newServer(accessTTL time.Duration) *server {
	hash, err := bcrypt.GenerateFromPassword([]byte("matatu-2026"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return &server{
		accounts:  map[string]account{"0712345678": {User{"0712345678", "Amina"}, hash}},
		access:    map[string]session{},
		refresh:   map[string]string{},
		accessTTL: accessTTL,
		now:       time.Now,
	}
}

func main() {
	addr := getenv("ADDR", ":8080")
	origin := getenv("ALLOWED_ORIGIN", "http://localhost:5173")
	ttl, err := time.ParseDuration(getenv("ACCESS_TTL", "1m"))
	if err != nil {
		log.Fatalf("ACCESS_TTL: %v", err)
	}
	log.Printf("listening on %s, allowing origin %s, access tokens last %s", addr, origin, ttl)
	log.Fatal(http.ListenAndServe(addr, newServer(ttl).routes(origin)))
}

func (s *server) routes(origin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/login", s.login)
	mux.HandleFunc("POST /api/refresh", s.refreshToken)
	mux.HandleFunc("POST /api/logout", s.logout)
	mux.HandleFunc("GET /api/me/bookings", s.bookings)
	return cors(origin, mux)
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var body struct{ Phone, Password string }
	r.Body = http.MaxBytesReader(w, r.Body, 1<<12)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "send JSON with phone and password"})
		return
	}
	s.mu.Lock()
	acc, ok := s.accounts[body.Phone]
	s.mu.Unlock()
	// Same answer for a wrong phone and a wrong password, so nobody can test which phones have accounts.
	if !ok || bcrypt.CompareHashAndPassword(acc.hash, []byte(body.Password)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "wrong phone number or password"})
		return
	}
	s.issue(w, acc.user)
}

func (s *server) refreshToken(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh")
	s.mu.Lock()
	phone, ok := "", false
	if err == nil {
		phone, ok = s.refresh[c.Value]
		delete(s.refresh, c.Value) // rotation: each refresh token works once
	}
	acc := s.accounts[phone]
	s.mu.Unlock()
	if !ok {
		clearCookie(w)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "please log in"})
		return
	}
	s.issue(w, acc.user)
}

// issue makes a new access token (returned in JSON, kept in memory by the page) and a new
// refresh token (in an HttpOnly cookie the page's JavaScript can never read).
func (s *server) issue(w http.ResponseWriter, u User) {
	access, refresh := rand.Text(), rand.Text()
	s.mu.Lock()
	s.access[access] = session{u.Phone, s.now().Add(s.accessTTL)}
	s.refresh[refresh] = u.Phone
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name: "refresh", Value: refresh, Path: "/api",
		MaxAge: 30 * 24 * 60 * 60, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, map[string]any{"accessToken": access, "user": u})
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("refresh"); err == nil {
		s.mu.Lock()
		delete(s.refresh, c.Value)
		s.mu.Unlock()
	}
	clearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) bookings(w http.ResponseWriter, r *http.Request) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	s.mu.Lock()
	sess, found := s.access[token]
	s.mu.Unlock()
	if !ok || !found || !s.now().Before(sess.expires) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "access token missing or expired"})
		return
	}
	writeJSON(w, http.StatusOK, []map[string]any{
		{"reference": "K7Q2M9XA", "route": "58", "seats": 2},
		{"reference": "P4WD8N2C", "route": "111", "seats": 1},
	})
}

func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: "refresh", Value: "", Path: "/api", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	})
}

// cors allows one origin, with credentials so the browser sends and stores the refresh cookie.
// With credentials, the allowed origin must be exact: "*" is refused by the browser.
func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		if r.Header.Get("Origin") != origin {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
