package rolesandpermissions

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCan(t *testing.T) {
	all := []Action{ListTrips, CreateTrip, CreateBooking, ReadBooking, ReadAnyBooking, CheckIn}
	allowed := map[Role][]Action{
		Passenger: {ListTrips, CreateBooking, ReadBooking},
		Conductor: {ListTrips, ReadBooking, ReadAnyBooking, CheckIn},
		Admin:     all,
		"":        nil,
		"driver":  nil,
	}
	for role, yes := range allowed {
		for _, a := range all {
			want := false
			for _, y := range yes {
				if y == a {
					want = true
				}
			}
			if got := Can(role, a); got != want {
				t.Errorf("Can(%q, %q) = %v, want %v", role, a, got, want)
			}
		}
		if Can(role, "trips:delete") {
			t.Errorf("Can(%q, \"trips:delete\") = true: an action nobody listed must be denied", role)
		}
	}
}

var (
	wanjiku  = User{ID: "u1", Name: "Wanjiku", Phone: "0712345678", Role: Passenger}
	otieno   = User{ID: "u2", Name: "Otieno", Phone: "0722000111", Role: Passenger}
	kamau    = User{ID: "u3", Name: "Kamau", Phone: "0733000222", Role: Conductor}
	amina    = User{ID: "u4", Name: "Amina", Phone: "0744000333", Role: Admin}
	stranger = User{ID: "u5", Name: "Ghost", Phone: "0700000000", Role: "driver"}
)

func TestCanViewBooking(t *testing.T) {
	b := Booking{ID: 7, TripID: 1, Seat: 3, Passenger: "Wanjiku", Phone: "0712345678"}
	tests := []struct {
		user User
		want bool
	}{
		{wanjiku, true},
		{otieno, false},
		{kamau, true},
		{amina, true},
		{stranger, false},
		{User{Phone: "0712345678", Role: "driver"}, false}, // right phone, but a role with no ReadBooking
	}
	for _, tt := range tests {
		if got := CanViewBooking(tt.user, b); got != tt.want {
			t.Errorf("CanViewBooking(%s the %s, Wanjiku's booking) = %v, want %v", tt.user.Name, tt.user.Role, got, tt.want)
		}
	}
}

func TestRequire(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	h := Require(CreateTrip, ok)
	tests := []struct {
		name    string
		user    *User
		want    int
		wantWWW string
	}{
		{"no user", nil, 401, "Bearer"},
		{"passenger", &wanjiku, 403, ""},
		{"conductor", &kamau, 403, ""},
		{"admin", &amina, http.StatusTeapot, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/trips", nil)
			if tt.user != nil {
				req = req.WithContext(WithUser(req.Context(), *tt.user))
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
			if got := rec.Header().Get("WWW-Authenticate"); got != tt.wantWWW {
				t.Errorf("WWW-Authenticate = %q, want %q", got, tt.wantWWW)
			}
			if tt.want == 401 || tt.want == 403 {
				var body struct{ Error string }
				if json.Unmarshal(rec.Body.Bytes(), &body) != nil || body.Error == "" {
					t.Errorf("body = %q, want JSON {\"error\": \"...\"}", rec.Body.String())
				}
			}
		})
	}
}

var tokens = StaticTokens{
	"tok-wanjiku": wanjiku,
	"tok-otieno":  otieno,
	"tok-kamau":   kamau,
	"tok-amina":   amina,
}

// setup returns a server with one trip (ID 1) and one booking of Wanjiku's (ID 2).
func setup(t *testing.T) (http.Handler, *MemStore) {
	t.Helper()
	store := NewMemStore()
	store.AddTrip(Trip{RouteID: 111, Departs: time.Date(2026, 3, 2, 6, 30, 0, 0, time.UTC), Seats: 14})
	store.AddBooking(Booking{TripID: 1, Seat: 3, Passenger: "Wanjiku", Phone: "0712345678"})
	return NewServer(store, tokens), store
}

func do(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRoutesByRole(t *testing.T) {
	newTrip := `{"routeId":125,"departs":"2026-03-02T07:00:00Z","seats":14}`
	newBooking := `{"tripId":1,"seat":5}`
	tests := []struct {
		method, path, body string
		want               map[string]int // token → status; "" is no token
	}{
		{"GET", "/trips", "", map[string]int{"": 401, "tok-wanjiku": 200, "tok-kamau": 200, "tok-amina": 200, "tok-fake": 401}},
		{"POST", "/trips", newTrip, map[string]int{"": 401, "tok-wanjiku": 403, "tok-kamau": 403, "tok-amina": 201}},
		{"POST", "/bookings", newBooking, map[string]int{"": 401, "tok-wanjiku": 201, "tok-kamau": 403, "tok-amina": 201}},
		{"GET", "/bookings/2", "", map[string]int{"": 401, "tok-wanjiku": 200, "tok-otieno": 404, "tok-kamau": 200, "tok-amina": 200}},
		{"GET", "/bookings/999", "", map[string]int{"tok-wanjiku": 404, "tok-amina": 404}},
		{"POST", "/bookings/2/checkin", "", map[string]int{"": 401, "tok-wanjiku": 403, "tok-otieno": 403, "tok-kamau": 204, "tok-amina": 204}},
	}
	for _, tt := range tests {
		for tok, want := range tt.want {
			h, _ := setup(t)
			rec := do(h, tt.method, tt.path, tok, tt.body)
			if rec.Code != want {
				t.Errorf("%s %s with token %q: status %d, want %d (body %s)", tt.method, tt.path, tok, rec.Code, want, strings.TrimSpace(rec.Body.String()))
			}
		}
	}
}

func TestIDORLooksLikeNotFound(t *testing.T) {
	h, _ := setup(t)
	theirs := do(h, "GET", "/bookings/2", "tok-otieno", "")
	missing := do(h, "GET", "/bookings/999", "tok-otieno", "")
	if theirs.Code != missing.Code || theirs.Body.String() != missing.Body.String() {
		t.Errorf("someone else's booking and a missing one must look the same\nsomeone else's: %d %s\nmissing:        %d %s",
			theirs.Code, theirs.Body.String(), missing.Code, missing.Body.String())
	}
	if strings.Contains(theirs.Body.String(), "Wanjiku") || strings.Contains(theirs.Body.String(), "0712345678") {
		t.Errorf("the response leaks the booking: %s", theirs.Body.String())
	}
}

func TestBookingOwnerComesFromLogin(t *testing.T) {
	h, store := setup(t)
	rec := do(h, "POST", "/bookings", "tok-otieno", `{"tripId":1,"seat":9}`)
	if rec.Code != 201 {
		t.Fatalf("POST /bookings as Otieno: status %d, want 201 (body %s)", rec.Code, rec.Body.String())
	}
	var b Booking
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("decoding the booking: %v", err)
	}
	if b.Phone != otieno.Phone || b.Passenger != "Otieno" || b.Seat != 9 || b.TripID != 1 {
		t.Errorf("booking = %+v\nwant Otieno's name and phone, trip 1, seat 9", b)
	}
	if got, ok := store.Booking(b.ID); !ok || got.Phone != otieno.Phone {
		t.Errorf("the store has %+v for ID %d", got, b.ID)
	}
	if loc := rec.Header().Get("Location"); loc == "" {
		t.Errorf("201 without a Location header")
	}

	// Trying to book in someone else's name is refused, not quietly obeyed.
	rec = do(h, "POST", "/bookings", "tok-otieno", `{"tripId":1,"seat":10,"phone":"0712345678"}`)
	if rec.Code != 400 {
		t.Errorf("POST /bookings with a phone in the body: status %d, want 400 (unknown field)", rec.Code)
	}

	// Otieno can read his own booking.
	if r := do(h, "GET", "/bookings/"+strconv.Itoa(b.ID), "tok-otieno", ""); r.Code != 200 {
		t.Errorf("GET his own booking: status %d, want 200", r.Code)
	}
}

func TestBookingValidation(t *testing.T) {
	h, _ := setup(t)
	tests := []struct {
		body string
		want int
	}{
		{`{"tripId":42,"seat":1}`, 404},
		{`{"tripId":1,"seat":0}`, 422},
		{`{"tripId":1,"seat":15}`, 422},
		{`{"tripId":1,`, 400},
	}
	for _, tt := range tests {
		if rec := do(h, "POST", "/bookings", "tok-wanjiku", tt.body); rec.Code != tt.want {
			t.Errorf("POST /bookings %s: status %d, want %d", tt.body, rec.Code, tt.want)
		}
	}
}
