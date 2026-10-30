// Command api is the capstone's booking API: Phase 5's matatu booking API, cut down to run in
// memory with no database. Run it from the repo root:
//
//	go run ./phase-6-frontend/day-076-booking-app/api
//
// Log in as 0712345678 with the password matatu-2026 (a demo account, made up).
//
//	POST /api/login               {"phone","password"} → 200 {"accessToken","user"} + refresh cookie, or 401
//	POST /api/refresh             refresh cookie → 200 {"accessToken","user"} + a new cookie, or 401
//	POST /api/logout              → 204
//	GET  /api/trips               → 200 [{"id","route","goesTo","boardsAt","departs","fareKsh","seatsLeft"}]
//	GET  /api/trips/{id}/seats    → 200 {"tripId","seatCount","taken":[2,5]}, or 404
//	POST /api/bookings            Bearer token, {"tripId","seat","name","phone"}
//	                              → 201 {"reference","tripId","seat","phone","totalKsh"}
//	                              → 401, 404 (no such trip), 409 {"error"} (seat taken), 422 {"errors":{...}}
//
// Environment: ADDR (default :8080), ALLOWED_ORIGIN (default http://localhost:5173),
// ACCESS_TTL (default 5m).
package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const seatCount = 14

// User is who is logged in.
type User struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

// Trip is one departure. Routes were checked on nishukishe.com in October 2026; the times,
// fares on the day and seats are made up. Routes change, so confirm with the conductor.
type Trip struct {
	ID        string `json:"id"`
	Route     string `json:"route"`
	GoesTo    string `json:"goesTo"`
	BoardsAt  string `json:"boardsAt"`
	Departs   string `json:"departs"`
	FareKsh   int    `json:"fareKsh"`
	SeatsLeft int    `json:"seatsLeft"`
}

type session struct {
	phone   string
	expires time.Time
}

type server struct {
	mu        sync.Mutex
	hash      []byte
	user      User
	trips     []Trip
	taken     map[string][]int // trip ID → taken seat numbers
	access    map[string]session
	refresh   map[string]string
	accessTTL time.Duration
	now       func() time.Time
}

func newServer(accessTTL time.Duration) *server {
	hash, err := bcrypt.GenerateFromPassword([]byte("matatu-2026"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return &server{
		hash: hash,
		user: User{Phone: "0712345678", Name: "Amina"},
		trips: []Trip{
			{ID: "T1", Route: "58", GoesTo: "Buruburu", BoardsAt: "Ambassadeur", Departs: "07:45", FareKsh: 150},
			{ID: "T2", Route: "111", GoesTo: "Ngong", BoardsAt: "Railways", Departs: "07:50", FareKsh: 100},
			{ID: "T3", Route: "46K", GoesTo: "Kawangware", BoardsAt: "Kencom / Ambassadeur", Departs: "08:00", FareKsh: 100},
			{ID: "T4", Route: "125", GoesTo: "Ongata Rongai", BoardsAt: "Railways", Departs: "08:10", FareKsh: 100},
			{ID: "T5", Route: "45", GoesTo: "Githurai", BoardsAt: "Ronald Ngala Street", Departs: "08:15", FareKsh: 100},
		},
		taken: map[string][]int{
			"T1": {1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 13},
			"T2": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
			"T3": {2, 5},
			"T4": {},
			"T5": {1, 3, 7, 8},
		},
		access:    map[string]session{},
		refresh:   map[string]string{},
		accessTTL: accessTTL,
		now:       time.Now,
	}
}

func main() {
	addr := getenv("ADDR", ":8080")
	origin := getenv("ALLOWED_ORIGIN", "http://localhost:5173")
	ttl, err := time.ParseDuration(getenv("ACCESS_TTL", "5m"))
	if err != nil {
		log.Fatalf("ACCESS_TTL: %v", err)
	}
	log.Printf("listening on %s, allowing origin %s", addr, origin)
	log.Fatal(http.ListenAndServe(addr, newServer(ttl).routes(origin)))
}

func (s *server) routes(origin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/login", s.login)
	mux.HandleFunc("POST /api/refresh", s.refreshToken)
	mux.HandleFunc("POST /api/logout", s.logout)
	mux.HandleFunc("GET /api/trips", s.listTrips)
	mux.HandleFunc("GET /api/trips/{id}/seats", s.seats)
	mux.HandleFunc("POST /api/bookings", s.book)
	return cors(origin, mux)
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var body struct{ Phone, Password string }
	r.Body = http.MaxBytesReader(w, r.Body, 1<<12)
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "send JSON with phone and password"})
		return
	}
	if body.Phone != s.user.Phone || bcrypt.CompareHashAndPassword(s.hash, []byte(body.Password)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "wrong phone number or password"})
		return
	}
	s.issue(w)
}

func (s *server) refreshToken(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh")
	ok := false
	if err == nil {
		s.mu.Lock()
		_, ok = s.refresh[c.Value]
		delete(s.refresh, c.Value) // each refresh token works once
		s.mu.Unlock()
	}
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "please log in"})
		return
	}
	s.issue(w)
}

func (s *server) issue(w http.ResponseWriter) {
	access, refresh := rand.Text(), rand.Text()
	s.mu.Lock()
	s.access[access] = session{s.user.Phone, s.now().Add(s.accessTTL)}
	s.refresh[refresh] = s.user.Phone
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name: "refresh", Value: refresh, Path: "/api", MaxAge: 30 * 24 * 60 * 60,
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, map[string]any{"accessToken": access, "user": s.user})
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("refresh"); err == nil {
		s.mu.Lock()
		delete(s.refresh, c.Value)
		s.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "refresh", Path: "/api", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) listTrips(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Trip, len(s.trips))
	for i, t := range s.trips {
		t.SeatsLeft = seatCount - len(s.taken[t.ID])
		out[i] = t
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *server) seats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	defer s.mu.Unlock()
	taken, ok := s.taken[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such trip"})
		return
	}
	sorted := slices.Sorted(slices.Values(taken))
	writeJSON(w, http.StatusOK, map[string]any{"tripId": id, "seatCount": seatCount, "taken": sorted})
}

var (
	localPhone = regexp.MustCompile(`^0([17][0-9]{8})$`)
	intlPhone  = regexp.MustCompile(`^\+?254([17][0-9]{8})$`)
)

// normalisePhone is the same rule as Day 69 and the front end's validate.ts.
func normalisePhone(s string) (string, bool) {
	s = strings.NewReplacer(" ", "", "-", "").Replace(s)
	for _, re := range []*regexp.Regexp{localPhone, intlPhone} {
		if m := re.FindStringSubmatch(s); m != nil {
			return "254" + m[1], true
		}
	}
	return "", false
}

func (s *server) book(w http.ResponseWriter, r *http.Request) {
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	s.mu.Lock()
	sess, ok := s.access[token]
	s.mu.Unlock()
	if !ok || !s.now().Before(sess.expires) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "please log in"})
		return
	}

	var b struct {
		TripID string `json:"tripId"`
		Seat   int    `json:"seat"`
		Name   string `json:"name"`
		Phone  string `json:"phone"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<12)
	if json.NewDecoder(r.Body).Decode(&b) != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "the body must be a JSON booking"})
		return
	}
	errs := map[string]string{}
	name := strings.TrimSpace(b.Name)
	if name == "" {
		errs["name"] = "Enter your name"
	} else if utf8.RuneCountInString(name) > 50 {
		errs["name"] = "Name is too long (50 characters at most)"
	}
	phone, okPhone := normalisePhone(b.Phone)
	if !okPhone {
		errs["phone"] = "Enter a Kenyan mobile number like 0712 345 678"
	}
	if b.Seat < 1 || b.Seat > seatCount {
		errs["seat"] = "Choose a seat"
	}
	if len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	taken, ok := s.taken[b.TripID]
	var trip Trip
	for _, t := range s.trips {
		if t.ID == b.TripID {
			trip = t
		}
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such trip"})
		return
	}
	if slices.Contains(taken, b.Seat) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("Seat %d was just taken", b.Seat)})
		return
	}
	s.taken[b.TripID] = append(taken, b.Seat)
	writeJSON(w, http.StatusCreated, map[string]any{
		"reference": rand.Text()[:8], "tripId": b.TripID, "seat": b.Seat, "phone": phone, "totalKsh": trip.FareKsh,
	})
}

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
