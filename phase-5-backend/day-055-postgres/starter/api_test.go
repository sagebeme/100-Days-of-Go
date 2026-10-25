package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeStore is a Store in memory. It also remembers whether each call's context
// had a deadline, so the tests can check the handlers set one.
type fakeStore struct {
	mu          sync.Mutex
	trips       map[int]Trip
	bookings    []Booking
	sawDeadline bool
}

func newFake() *fakeStore {
	return &fakeStore{trips: map[int]Trip{
		1: {ID: 1, RouteID: 125, Departs: time.Date(2026, 11, 2, 4, 30, 0, 0, time.UTC), Seats: 14},
		2: {ID: 2, RouteID: 125, Departs: time.Date(2026, 11, 2, 5, 0, 0, 0, time.UTC), Seats: 14},
	}}
}

func (f *fakeStore) note(ctx context.Context) {
	_, ok := ctx.Deadline()
	f.sawDeadline = ok
}

func (f *fakeStore) CreateRoute(ctx context.Context, r Route) (Route, error) { return r, nil }
func (f *fakeStore) CreateTrip(ctx context.Context, t Trip) (Trip, error)    { return t, nil }

func (f *fakeStore) Trip(ctx context.Context, id int) (Trip, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.note(ctx)
	t, ok := f.trips[id]
	if !ok {
		return Trip{}, ErrNotFound
	}
	return t, nil
}

func (f *fakeStore) TripsOnRoute(ctx context.Context, routeID int) ([]Trip, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.note(ctx)
	out := []Trip{}
	for id := 1; id <= len(f.trips); id++ {
		if t := f.trips[id]; t.RouteID == routeID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeStore) Book(ctx context.Context, b Booking) (Booking, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.note(ctx)
	for _, o := range f.bookings {
		if o.TripID == b.TripID && o.Seat == b.Seat {
			return Booking{}, ErrSeatTaken
		}
	}
	b.ID = len(f.bookings) + 1
	f.bookings = append(f.bookings, b)
	return b, nil
}

// slowStore never answers before the context gives up, like a database stuck on a lock.
type slowStore struct{ *fakeStore }

func (s *slowStore) Trip(ctx context.Context, id int) (Trip, error) {
	<-ctx.Done()
	return Trip{}, ctx.Err()
}

func (s *slowStore) TripsOnRoute(ctx context.Context, routeID int) ([]Trip, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// brokenStore fails like a database that refuses connections.
type brokenStore struct{ *fakeStore }

func (*brokenStore) Trip(ctx context.Context, id int) (Trip, error) {
	return Trip{}, errors.New("dial tcp 10.0.3.7:5432: connect: connection refused")
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func newAPI(t *testing.T, s Store, timeout time.Duration) http.Handler {
	t.Helper()
	h := NewAPI(s, timeout).Routes()
	if h == nil {
		t.Fatal("Routes() returned nil")
	}
	return h
}

func TestRoutes(t *testing.T) {
	const good = `{"seat": 3, "passenger": "Achieng", "phone": "254712345678"}`
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{"trips on a route", "GET", "/routes/125/trips", "", 200},
		{"route id not a number", "GET", "/routes/abc/trips", "", 400},
		{"a trip", "GET", "/trips/1", "", 200},
		{"no such trip", "GET", "/trips/9", "", 404},
		{"book", "POST", "/trips/1/bookings", good, 201},
		{"book on no trip", "POST", "/trips/9/bookings", good, 404},
		{"broken JSON", "POST", "/trips/1/bookings", `{"seat": 3`, 400},
		{"unknown field", "POST", "/trips/1/bookings", `{"seat": 3, "passenger": "A", "phone": "1", "fare": 100}`, 400},
		{"seat outside the matatu", "POST", "/trips/1/bookings", `{"seat": 15, "passenger": "Achieng", "phone": "254712345678"}`, 422},
		{"no passenger", "POST", "/trips/1/bookings", `{"seat": 3, "passenger": " ", "phone": "254712345678"}`, 422},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFake()
			rec := do(newAPI(t, f, time.Second), tt.method, tt.path, tt.body)
			if rec.Code != tt.status {
				t.Fatalf("%s %s: status = %d, want %d (body %q)", tt.method, tt.path, rec.Code, tt.status, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}

func TestRouteTripsBody(t *testing.T) {
	h := newAPI(t, newFake(), time.Second)
	rec := do(h, "GET", "/routes/125/trips", "")
	var trips []Trip
	if err := json.Unmarshal(rec.Body.Bytes(), &trips); err != nil || len(trips) != 2 || trips[0].ID != 1 || trips[1].ID != 2 {
		t.Errorf("GET /routes/125/trips = %s, want trips 1 and 2", rec.Body.String())
	}
	if got := strings.TrimSpace(do(h, "GET", "/routes/111/trips", "").Body.String()); got != "[]" {
		t.Errorf("a route with no trips = %q, want []", got)
	}
}

func TestBookingFlow(t *testing.T) {
	f := newFake()
	h := newAPI(t, f, time.Second)
	body := `{"seat": 3, "passenger": " Achieng ", "phone": "254712345678"}`
	rec := do(h, "POST", "/trips/1/bookings", body)
	if rec.Code != 201 {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/bookings/1" {
		t.Errorf("Location = %q, want /bookings/1", loc)
	}
	if len(f.bookings) != 1 || f.bookings[0].Passenger != "Achieng" || f.bookings[0].TripID != 1 {
		t.Errorf("store holds %+v, want one booking for Achieng on trip 1", f.bookings)
	}
	if !f.sawDeadline {
		t.Error("the store got a context with no deadline: wrap r.Context() with context.WithTimeout")
	}
	if rec := do(h, "POST", "/trips/1/bookings", body); rec.Code != 409 {
		t.Errorf("same seat again: status = %d, want 409", rec.Code)
	}
}

func TestEveryCallHasADeadline(t *testing.T) {
	for _, path := range []string{"/routes/125/trips", "/trips/1"} {
		f := newFake()
		do(newAPI(t, f, time.Second), "GET", path, "")
		if !f.sawDeadline {
			t.Errorf("GET %s: the store got a context with no deadline", path)
		}
	}
}

func TestSlowDatabase(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	defer log.SetOutput(os.Stderr)

	h := newAPI(t, &slowStore{newFake()}, 20*time.Millisecond)
	for _, path := range []string{"/routes/125/trips", "/trips/1"} {
		done := make(chan *httptest.ResponseRecorder, 1)
		go func() { done <- do(h, "GET", path, "") }()
		select {
		case rec := <-done:
			if rec.Code != 503 {
				t.Errorf("GET %s with a stuck database: status = %d, want 503", path, rec.Code)
			}
			if rec.Header().Get("Retry-After") == "" {
				t.Errorf("GET %s: 503 without Retry-After", path)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("GET %s never answered: the handler must stop waiting when its timeout passes", path)
		}
	}
}

func TestDatabaseErrorsDontLeak(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	defer log.SetOutput(os.Stderr)

	rec := do(newAPI(t, &brokenStore{newFake()}, time.Second), "GET", "/trips/1", "")
	if rec.Code != 500 {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "10.0.3.7") {
		t.Errorf("the 500 leaks the database address: %s", rec.Body.String())
	}
	if !strings.Contains(logs.String(), "connection refused") {
		t.Errorf("the real error wasn't logged; log was %q", logs.String())
	}
}
