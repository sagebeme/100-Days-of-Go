package statuscodes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestStatusFor(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"no error", nil, 200, ""},
		{"not found", ErrNotFound, 404, "not found"},
		{"wrapped not found", fmt.Errorf("trip 7: %w", ErrNotFound), 404, "not found"},
		{"seat taken", ErrSeatTaken, 409, "that seat is already booked"},
		{"wrapped seat taken", fmt.Errorf("booking seat 3: %w", ErrSeatTaken), 409, "that seat is already booked"},
		{"invalid", &InvalidError{Message: "seat must be between 1 and 14"}, 422, "seat must be between 1 and 14"},
		{"wrapped invalid", fmt.Errorf("checking: %w", &InvalidError{Message: "phone is required"}), 422, "phone is required"},
		{"database down", errors.New("dial tcp 10.0.3.7:5432: connection refused"), 500, "something went wrong"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := StatusFor(tt.err)
			if status != tt.wantStatus || msg != tt.wantMsg {
				t.Errorf("StatusFor(%v) = %d, %q, want %d, %q", tt.err, status, msg, tt.wantStatus, tt.wantMsg)
			}
		})
	}
}

func TestInvalidErrorMessage(t *testing.T) {
	err := &InvalidError{Message: "seat must be between 1 and 14"}
	if got := err.Error(); got != "seat must be between 1 and 14" {
		t.Errorf("Error() = %q, want the message", got)
	}
}

// limiter is a fake Limiter: it says no while full is true.
type limiter struct {
	full bool
	wait time.Duration
	keys []string
}

func (l *limiter) Allow(key string) (bool, time.Duration) {
	l.keys = append(l.keys, key)
	if l.full {
		return false, l.wait
	}
	return true, 0
}

// brokenStore fails every call the way a database that's down does.
type brokenStore struct{}

var errDB = errors.New(`pq: password authentication failed for user "matatu_admin" at db.internal:5432`)

func (brokenStore) Trip(context.Context, int) (Trip, error)        { return Trip{}, errDB }
func (brokenStore) Book(context.Context, Booking) (Booking, error) { return Booking{}, errDB }
func (brokenStore) Booking(context.Context, int) (Booking, error)  { return Booking{}, errDB }
func (brokenStore) CancelBooking(context.Context, int) error       { return errDB }

var tokens = map[string]string{"admin-token-wanjiru": "admin", "clerk-token-kamau": "clerk"}

func newAPI(t *testing.T) (*MemStore, *limiter, http.Handler) {
	t.Helper()
	s := NewMemStore(Trip{ID: 1, RouteID: 125, Departs: time.Date(2026, 11, 2, 7, 30, 0, 0, time.UTC), Seats: 14})
	if _, err := s.Book(context.Background(), Booking{TripID: 1, Seat: 3, Passenger: "Achieng", Phone: "0712345678"}); err != nil {
		t.Fatal(err)
	}
	l := &limiter{}
	h := NewAPI(s, l, tokens).Routes()
	if h == nil {
		t.Fatal("Routes() returned nil")
	}
	return s, l, h
}

type call struct {
	method string
	path   string
	body   string
	token  string
}

func do(h http.Handler, c call) *httptest.ResponseRecorder {
	var body io.Reader
	if c.body != "" {
		body = strings.NewReader(c.body)
	}
	req := httptest.NewRequest(c.method, c.path, body)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// errorMessage checks the body is {"error": "..."} sent as JSON, and returns the message.
func errorMessage(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("error Content-Type = %q, want application/json", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("error body %q isn't JSON: %v", rec.Body.String(), err)
	}
	msg, ok := body["error"].(string)
	if !ok || msg == "" || len(body) != 1 {
		t.Errorf("error body = %s, want exactly {\"error\": \"...\"}", rec.Body.String())
	}
	return msg
}

const goodBooking = `{"seat": 5, "passenger": "Baraka Mwangi", "phone": "0722000111"}`

func TestStatuses(t *testing.T) {
	tests := []struct {
		name   string
		c      call
		status int
	}{
		{"read a trip", call{method: "GET", path: "/trips/1"}, 200},
		{"no such trip", call{method: "GET", path: "/trips/9"}, 404},
		{"trip id not a number", call{method: "GET", path: "/trips/abc"}, 400},
		{"book", call{method: "POST", path: "/trips/1/bookings", body: goodBooking}, 201},
		{"broken JSON", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 5,`}, 400},
		{"seat outside the matatu", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 15, "passenger": "Kip", "phone": "0711111111"}`}, 422},
		{"seat zero", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 0, "passenger": "Kip", "phone": "0711111111"}`}, 422},
		{"no passenger", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 6, "passenger": "  ", "phone": "0711111111"}`}, 422},
		{"no phone", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 6, "passenger": "Kip"}`}, 422},
		{"seat taken", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 3, "passenger": "Kip", "phone": "0711111111"}`}, 409},
		{"book on no trip", call{method: "POST", path: "/trips/9/bookings", body: goodBooking}, 404},
		{"read a booking", call{method: "GET", path: "/bookings/1"}, 200},
		{"no such booking", call{method: "GET", path: "/bookings/99"}, 404},
		{"cancel with no token", call{method: "DELETE", path: "/bookings/1"}, 401},
		{"cancel with an unknown token", call{method: "DELETE", path: "/bookings/1", token: "made-up"}, 401},
		{"cancel as a clerk", call{method: "DELETE", path: "/bookings/1", token: "clerk-token-kamau"}, 403},
		{"cancel a missing booking", call{method: "DELETE", path: "/bookings/99", token: "admin-token-wanjiru"}, 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, h := newAPI(t)
			rec := do(h, tt.c)
			if rec.Code != tt.status {
				t.Fatalf("%s %s: status = %d, want %d (body %q)", tt.c.method, tt.c.path, rec.Code, tt.status, rec.Body.String())
			}
			if tt.status >= 400 {
				errorMessage(t, rec)
			}
		})
	}
}

func TestCreatedHasLocation(t *testing.T) {
	_, _, h := newAPI(t)
	rec := do(h, call{method: "POST", path: "/trips/1/bookings", body: goodBooking})
	if rec.Code != 201 {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/bookings/2" {
		t.Errorf("Location = %q, want /bookings/2", loc)
	}
	var b Booking
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || b.ID != 2 || b.Seat != 5 {
		t.Errorf("body = %q, want booking 2 for seat 5", rec.Body.String())
	}
}

func TestCancelAsAdmin(t *testing.T) {
	s, _, h := newAPI(t)
	rec := do(h, call{method: "DELETE", path: "/bookings/1", token: "admin-token-wanjiru"})
	if rec.Code != 204 {
		t.Fatalf("status = %d, want 204 (body %q)", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Errorf("204 must have no body, got %q", rec.Body.String())
	}
	if _, err := s.Booking(context.Background(), 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("booking 1 is still there after the cancel")
	}
}

func TestUnauthorizedSaysHowToAuthenticate(t *testing.T) {
	_, _, h := newAPI(t)
	rec := do(h, call{method: "DELETE", path: "/bookings/1"})
	if got := rec.Header().Get("WWW-Authenticate"); !strings.HasPrefix(got, "Bearer") {
		t.Errorf("401: WWW-Authenticate = %q, want it to start with Bearer", got)
	}
	rec = do(h, call{method: "DELETE", path: "/bookings/1", token: "clerk-token-kamau"})
	if got := rec.Header().Get("WWW-Authenticate"); got != "" {
		t.Errorf("403: WWW-Authenticate = %q, want none: logging in again won't help", got)
	}
}

func TestTooManyRequests(t *testing.T) {
	tests := []struct {
		wait time.Duration
		want string
	}{
		{30 * time.Second, "30"},
		{1500 * time.Millisecond, "2"}, // round up: retrying early just gets another 429
		{0, "1"},
	}
	for _, tt := range tests {
		s, l, h := newAPI(t)
		l.full, l.wait = true, tt.wait
		rec := do(h, call{method: "POST", path: "/trips/1/bookings", body: goodBooking})
		if rec.Code != 429 {
			t.Fatalf("status = %d, want 429", rec.Code)
		}
		if got := rec.Header().Get("Retry-After"); got != tt.want {
			t.Errorf("wait %v: Retry-After = %q, want %q", tt.wait, got, tt.want)
		}
		errorMessage(t, rec)
		if _, err := s.Booking(context.Background(), 2); err == nil {
			t.Error("a rate-limited request still made a booking")
		}
	}

	_, l, h := newAPI(t)
	do(h, call{method: "POST", path: "/trips/1/bookings", body: goodBooking})
	if len(l.keys) != 1 || l.keys[0] != "192.0.2.1" {
		t.Errorf("limiter keys = %q, want [192.0.2.1]: the client's IP without its port", l.keys)
	}
}

func TestServerErrorsDontLeak(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	defer log.SetOutput(os.Stderr)

	h := NewAPI(brokenStore{}, &limiter{}, tokens).Routes()
	calls := []call{
		{method: "GET", path: "/trips/1"},
		{method: "POST", path: "/trips/1/bookings", body: goodBooking},
		{method: "GET", path: "/bookings/1"},
		{method: "DELETE", path: "/bookings/1", token: "admin-token-wanjiru"},
	}
	for _, c := range calls {
		logs.Reset()
		rec := do(h, c)
		if rec.Code != 500 {
			t.Errorf("%s %s with the database down: status = %d, want 500", c.method, c.path, rec.Code)
			continue
		}
		msg := errorMessage(t, rec)
		for _, secret := range []string{"matatu_admin", "db.internal", "pq:"} {
			if strings.Contains(rec.Body.String(), secret) {
				t.Errorf("%s %s: the 500 body leaks %q: %s", c.method, c.path, secret, rec.Body.String())
			}
		}
		if msg != "something went wrong" {
			t.Errorf("%s %s: 500 message = %q, want %q", c.method, c.path, msg, "something went wrong")
		}
		if !strings.Contains(logs.String(), "password authentication failed") {
			t.Errorf("%s %s: the real error wasn't logged with log.Printf; log was %q", c.method, c.path, logs.String())
		}
	}
}
