// Command api is Day 69's booking API. Run it from the repo root:
//
//	go run ./phase-6-frontend/day-069-forms/api
//
// POST /api/bookings with {"name","phone","route","seats"} answers
// 201 {"reference","phone","route","seats"}, or 422 {"errors": {"field": "message"}}.
//
// Environment: ADDR (default :8080), ALLOWED_ORIGIN (default http://localhost:5173).
package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

// Booking is the request body. The JSON names match the TypeScript BookingInput.
type Booking struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Route string `json:"route"`
	Seats int    `json:"seats"`
}

// The messages are the same, word for word, as in validate.ts.
const (
	msgNameMissing = "Enter your name"
	msgNameLong    = "Name is too long (50 characters at most)"
	msgPhone       = "Enter a Kenyan mobile number like 0712 345 678"
	msgRoute       = "Choose a route"
	msgSeats       = "Book between 1 and 14 seats"
)

var (
	localPhone = regexp.MustCompile(`^0([17][0-9]{8})$`)
	intlPhone  = regexp.MustCompile(`^\+?254([17][0-9]{8})$`)
)

// NormalisePhone turns 0712 345 678, +254712345678 or 254-712-345678 into 254712345678.
// ok is false if it isn't a Kenyan mobile number.
func NormalisePhone(s string) (string, bool) {
	s = strings.NewReplacer(" ", "", "-", "").Replace(s)
	if m := localPhone.FindStringSubmatch(s); m != nil {
		return "254" + m[1], true
	}
	if m := intlPhone.FindStringSubmatch(s); m != nil {
		return "254" + m[1], true
	}
	return "", false
}

// Validate checks the rules the browser also checks. It returns one message per bad field.
func Validate(b Booking, routes map[string]int) map[string]string {
	errs := map[string]string{}
	name := strings.TrimSpace(b.Name)
	switch {
	case name == "":
		errs["name"] = msgNameMissing
	case utf8.RuneCountInString(name) > 50:
		errs["name"] = msgNameLong
	}
	if _, ok := NormalisePhone(b.Phone); !ok {
		errs["phone"] = msgPhone
	}
	if _, ok := routes[b.Route]; !ok {
		errs["route"] = msgRoute
	}
	if b.Seats < 1 || b.Seats > 14 {
		errs["seats"] = msgSeats
	}
	return errs
}

// store holds seats left per route. The numbers are made up for the exercise.
type store struct {
	mu   sync.Mutex
	left map[string]int
}

func newStore() *store {
	return &store{left: map[string]int{"46K": 14, "111": 3, "125": 14, "45": 14, "58": 14}}
}

func main() {
	addr := getenv("ADDR", ":8080")
	origin := getenv("ALLOWED_ORIGIN", "http://localhost:5173")
	log.Printf("listening on %s, allowing origin %s", addr, origin)
	log.Fatal(http.ListenAndServe(addr, newHandler(origin, newStore())))
}

func newHandler(origin string, s *store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/bookings", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<14)
		var b Booking
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "the body must be a JSON booking"})
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		errs := Validate(b, s.left)
		// A rule only the server can check: how many seats are really left.
		if len(errs) == 0 && b.Seats > s.left[b.Route] {
			errs["seats"] = fmt.Sprintf("Only %d seats left on %s", s.left[b.Route], b.Route)
		}
		if len(errs) > 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
			return
		}
		s.left[b.Route] -= b.Seats
		phone, _ := NormalisePhone(b.Phone)
		writeJSON(w, http.StatusCreated, map[string]any{
			"reference": rand.Text()[:8], "phone": phone, "route": b.Route, "seats": b.Seats,
		})
	})
	return cors(origin, mux)
}

// cors allows one origin. A JSON POST is not a "simple" request, so the browser first sends
// an OPTIONS preflight asking whether POST with a Content-Type header is allowed.
func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		if r.Header.Get("Origin") != origin {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "600")
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
