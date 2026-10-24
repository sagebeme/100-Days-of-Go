package httpmethods

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var nairobi = time.FixedZone("EAT", 3*60*60)

// newAPI returns an API whose store already holds trip 1: route 125, 14 seats.
func newAPI(t *testing.T) (*Store, http.Handler) {
	t.Helper()
	s := NewStore()
	s.CreateTrip(Trip{RouteID: 125, Departs: time.Date(2026, 11, 2, 7, 30, 0, 0, nairobi), Seats: 14})
	h := NewAPI(s).Routes()
	if h == nil {
		t.Fatal("Routes() returned nil")
	}
	return s, h
}

type call struct {
	method string
	path   string
	body   string
	key    string // Idempotency-Key, "" for none
}

func do(h http.Handler, c call) *httptest.ResponseRecorder {
	var body io.Reader
	if c.body != "" {
		body = strings.NewReader(c.body)
	}
	req := httptest.NewRequest(c.method, c.path, body)
	if c.key != "" {
		req.Header.Set("Idempotency-Key", c.key)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeTrip(t *testing.T, rec *httptest.ResponseRecorder) Trip {
	t.Helper()
	var got Trip
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body %q isn't a trip: %v", rec.Body.String(), err)
	}
	return got
}

func wantStatus(t *testing.T, c call, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("%s %s: status = %d, want %d (body %q)", c.method, c.path, rec.Code, want, rec.Body.String())
	}
}

func TestGetTrips(t *testing.T) {
	s := NewStore()
	h := NewAPI(s).Routes()
	c := call{method: "GET", path: "/trips"}
	rec := do(h, c)
	wantStatus(t, c, rec, 200)
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("GET /trips with no trips = %q, want [] (an empty list, not null)", got)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	_, h = newAPI(t)
	tests := []struct {
		path   string
		status int
	}{
		{"/trips/1", 200},
		{"/trips/99", 404},
		{"/trips/abc", 400},
		{"/trips/0", 400},
	}
	for _, tt := range tests {
		c := call{method: "GET", path: tt.path}
		wantStatus(t, c, do(h, c), tt.status)
	}
	rec = do(h, call{method: "GET", path: "/trips/1"})
	got := decodeTrip(t, rec)
	if got.ID != 1 || got.RouteID != 125 || got.Seats != 14 {
		t.Errorf("GET /trips/1 = %+v, want trip 1 on route 125 with 14 seats", got)
	}
}

func TestPostTrip(t *testing.T) {
	s, h := newAPI(t)
	c := call{method: "POST", path: "/trips", body: `{"routeId": 111, "departs": "2026-11-02T08:00:00+03:00", "seats": 33}`}
	rec := do(h, c)
	wantStatus(t, c, rec, 201)
	if loc := rec.Header().Get("Location"); loc != "/trips/2" {
		t.Errorf("Location = %q, want /trips/2", loc)
	}
	got := decodeTrip(t, rec)
	if got.ID != 2 || got.RouteID != 111 || got.Seats != 33 {
		t.Errorf("created trip = %+v, want ID 2, route 111, 33 seats", got)
	}
	if len(s.Trips()) != 2 {
		t.Errorf("store has %d trips after one POST, want 2", len(s.Trips()))
	}

	bad := []string{
		`{"routeId": 111`,
		`{"routeId": 111, "seats": 33}`,
		`{"departs": "2026-11-02T08:00:00+03:00", "seats": 33}`,
		`{"routeId": 111, "departs": "2026-11-02T08:00:00+03:00", "seats": 0}`,
	}
	for _, body := range bad {
		c := call{method: "POST", path: "/trips", body: body}
		wantStatus(t, c, do(h, c), 400)
	}
	if len(s.Trips()) != 2 {
		t.Errorf("store has %d trips after bad POSTs, want still 2", len(s.Trips()))
	}
}

func TestPutIsIdempotent(t *testing.T) {
	s, h := newAPI(t)
	c := call{method: "PUT", path: "/trips/1", body: `{"routeId": 58, "departs": "2026-11-02T09:15:00+03:00", "seats": 11}`}
	first := do(h, c)
	wantStatus(t, c, first, 200)
	second := do(h, c)
	wantStatus(t, c, second, 200)
	if first.Body.String() != second.Body.String() {
		t.Errorf("the same PUT twice gave different answers:\n%s\n%s", first.Body, second.Body)
	}
	got, _ := s.Trip(1)
	want := Trip{ID: 1, RouteID: 58, Departs: time.Date(2026, 11, 2, 9, 15, 0, 0, nairobi), Seats: 11}
	if got.ID != want.ID || got.RouteID != want.RouteID || got.Seats != want.Seats || !got.Departs.Equal(want.Departs) {
		t.Errorf("after PUT the trip is %+v, want %+v", got, want)
	}
	if len(s.Trips()) != 1 {
		t.Errorf("PUT changed the number of trips to %d, want 1", len(s.Trips()))
	}

	tests := []struct {
		name   string
		c      call
		status int
	}{
		{"missing seats", call{method: "PUT", path: "/trips/1", body: `{"routeId": 58, "departs": "2026-11-02T09:15:00+03:00"}`}, 400},
		{"no such trip", call{method: "PUT", path: "/trips/99", body: `{"routeId": 58, "departs": "2026-11-02T09:15:00+03:00", "seats": 11}`}, 404},
		{"broken JSON", call{method: "PUT", path: "/trips/1", body: `seats=11`}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantStatus(t, tt.c, do(h, tt.c), tt.status)
		})
	}
}

func TestPatchChangesOnlyWhatWasSent(t *testing.T) {
	s, h := newAPI(t)
	before, _ := s.Trip(1)

	c := call{method: "PATCH", path: "/trips/1", body: `{"seats": 10}`}
	rec := do(h, c)
	wantStatus(t, c, rec, 200)
	got := decodeTrip(t, rec)
	if got.Seats != 10 {
		t.Errorf("PATCH seats: seats = %d, want 10", got.Seats)
	}
	if got.RouteID != before.RouteID || !got.Departs.Equal(before.Departs) {
		t.Errorf("PATCH {seats} also changed other fields: %+v, before %+v", got, before)
	}

	c = call{method: "PATCH", path: "/trips/1", body: `{}`}
	rec = do(h, c)
	wantStatus(t, c, rec, 200)
	if got := decodeTrip(t, rec); got.Seats != 10 || got.RouteID != 125 {
		t.Errorf("PATCH {} changed the trip to %+v", got)
	}

	tests := []struct {
		name   string
		c      call
		status int
	}{
		{"zero seats", call{method: "PATCH", path: "/trips/1", body: `{"seats": 0}`}, 400},
		{"no such trip", call{method: "PATCH", path: "/trips/99", body: `{"seats": 5}`}, 404},
		{"broken JSON", call{method: "PATCH", path: "/trips/1", body: `{"seats":`}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantStatus(t, tt.c, do(h, tt.c), tt.status)
		})
	}
	if got, _ := s.Trip(1); got.Seats != 10 {
		t.Errorf("a rejected PATCH changed seats to %d", got.Seats)
	}
}

func TestDelete(t *testing.T) {
	s, h := newAPI(t)
	c := call{method: "DELETE", path: "/trips/1"}
	rec := do(h, c)
	wantStatus(t, c, rec, 204)
	if rec.Body.Len() != 0 {
		t.Errorf("204 must have no body, got %q", rec.Body.String())
	}
	if _, ok := s.Trip(1); ok {
		t.Error("trip 1 is still in the store after DELETE")
	}
	// The second DELETE finds nothing, but the state is the same: idempotent.
	wantStatus(t, c, do(h, c), 404)
	g := call{method: "GET", path: "/trips/1"}
	wantStatus(t, g, do(h, g), 404)
}

func TestMethodNotAllowed(t *testing.T) {
	_, h := newAPI(t)
	tests := []struct {
		c         call
		wantAllow []string
	}{
		{call{method: "DELETE", path: "/trips"}, []string{"GET", "HEAD", "POST"}},
		{call{method: "PUT", path: "/trips"}, []string{"GET", "HEAD", "POST"}},
		{call{method: "POST", path: "/trips/1"}, []string{"GET", "HEAD", "PUT", "PATCH", "DELETE"}},
		{call{method: "GET", path: "/trips/1/bookings"}, []string{"POST"}},
	}
	for _, tt := range tests {
		rec := do(h, tt.c)
		wantStatus(t, tt.c, rec, 405)
		allow := rec.Header().Get("Allow")
		for _, m := range tt.wantAllow {
			if !strings.Contains(allow, m) {
				t.Errorf("%s %s: Allow = %q, want it to list %s", tt.c.method, tt.c.path, allow, m)
			}
		}
	}
}

func TestHead(t *testing.T) {
	_, h := newAPI(t)
	srv := httptest.NewServer(h) // a real server: it drops the body of a HEAD answer
	defer srv.Close()

	req, _ := http.NewRequest("HEAD", srv.URL+"/trips/1", nil)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("HEAD /trips/1: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Errorf("HEAD /trips/1: status = %d, want 200", resp.StatusCode)
	}
	if len(body) != 0 {
		t.Errorf("HEAD /trips/1 sent a body: %q", body)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("HEAD /trips/1: Content-Type = %q, want application/json, the same as GET", ct)
	}
}

const seat3 = `{"seat": 3, "passenger": "Achieng Odhiambo", "phone": "0712345678"}`

func TestPostWithoutKeyIsNotIdempotent(t *testing.T) {
	s, h := newAPI(t)
	c := call{method: "POST", path: "/trips/1/bookings", body: seat3}
	rec := do(h, c)
	wantStatus(t, c, rec, 201)
	if loc := rec.Header().Get("Location"); loc != "/bookings/1" {
		t.Errorf("Location = %q, want /bookings/1", loc)
	}
	// The app retries after a dropped connection, without a key: the seat is gone.
	wantStatus(t, c, do(h, c), 409)
	c4 := call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 4, "passenger": "Achieng Odhiambo", "phone": "0712345678"}`}
	wantStatus(t, c4, do(h, c4), 201)
	if n := s.BookingCount(1); n != 2 {
		t.Errorf("BookingCount(1) = %d, want 2", n)
	}

	g := call{method: "GET", path: "/bookings/1"}
	rec = do(h, g)
	wantStatus(t, g, rec, 200)
	var b Booking
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || b.Seat != 3 || b.TripID != 1 {
		t.Errorf("GET /bookings/1 = %q, want seat 3 on trip 1", rec.Body.String())
	}
}

func TestIdempotencyKey(t *testing.T) {
	s, h := newAPI(t)
	c := call{method: "POST", path: "/trips/1/bookings", body: seat3, key: "c0ffee-2026-11-02-achieng"}
	first := do(h, c)
	wantStatus(t, c, first, 201)
	second := do(h, c)
	wantStatus(t, c, second, 201)

	if first.Body.String() != second.Body.String() {
		t.Errorf("a retry with the same key gave a different body:\nfirst:  %s\nsecond: %s", first.Body, second.Body)
	}
	if a, b := first.Header().Get("Location"), second.Header().Get("Location"); a != b || a == "" {
		t.Errorf("Location: first %q, retry %q, want the same non-empty value", a, b)
	}
	if got := second.Header().Get("Idempotent-Replayed"); got != "true" {
		t.Errorf("retry: Idempotent-Replayed = %q, want true", got)
	}
	if got := first.Header().Get("Idempotent-Replayed"); got != "" {
		t.Errorf("first try: Idempotent-Replayed = %q, want it unset", got)
	}
	if n := s.BookingCount(1); n != 1 {
		t.Errorf("BookingCount(1) = %d after a retry with the same key, want 1", n)
	}

	// The same key for a different booking is a bug in the client.
	other := call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 5, "passenger": "Achieng Odhiambo", "phone": "0712345678"}`, key: c.key}
	wantStatus(t, other, do(h, other), 422)

	// A different key is a different booking: seat 3 is taken.
	again := call{method: "POST", path: "/trips/1/bookings", body: seat3, key: "another-key"}
	wantStatus(t, again, do(h, again), 409)

	// A new key for a free seat works.
	fresh := call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 6, "passenger": "Baraka Mwangi", "phone": "0722000111"}`, key: "baraka-1"}
	wantStatus(t, fresh, do(h, fresh), 201)
	if n := s.BookingCount(1); n != 2 {
		t.Errorf("BookingCount(1) = %d, want 2", n)
	}
}

func TestBookingErrors(t *testing.T) {
	_, h := newAPI(t)
	tests := []struct {
		name   string
		c      call
		status int
	}{
		{"no such trip", call{method: "POST", path: "/trips/99/bookings", body: seat3}, 404},
		{"seat 0", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 0, "passenger": "Kip", "phone": "0712345678"}`}, 400},
		{"seat 15 on a 14-seater", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 15, "passenger": "Kip", "phone": "0712345678"}`}, 400},
		{"no passenger", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 2, "phone": "0712345678"}`}, 400},
		{"broken JSON", call{method: "POST", path: "/trips/1/bookings", body: `{"seat": 2,`}, 400},
		{"bad trip id", call{method: "POST", path: "/trips/x/bookings", body: seat3}, 400},
		{"key too long", call{method: "POST", path: "/trips/1/bookings", body: seat3, key: strings.Repeat("k", 256)}, 400},
		{"no such booking", call{method: "GET", path: "/bookings/42"}, 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(h, tt.c)
			wantStatus(t, tt.c, rec, tt.status)
			var e struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e.Error == "" {
				t.Errorf("body = %q, want {\"error\": \"...\"}", rec.Body.String())
			}
		})
	}
}
