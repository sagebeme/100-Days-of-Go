package matatubookingapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const testSecret = "test-secret-for-the-matatu-api-0123456789"

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) Now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) Add(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.t = c.t.Add(d) }

// api is one running server for one test.
type api struct {
	t     *testing.T
	url   string
	clock *clock
	payer *FakePayer
	srv   *Server
	logs  *bytes.Buffer
}

func testConfig(c *clock) Config {
	return Config{
		Port:            8080,
		JWTSecret:       []byte(testSecret),
		Issuer:          "matatu-api",
		AccessTTL:       15 * time.Minute,
		AllowedOrigins:  []string{"https://matatu.example"},
		RatePerSecond:   1000,
		RateBurst:       1000,
		BcryptCost:      bcrypt.MinCost,
		AdminPhones:     []string{"0700000001"},
		ShutdownTimeout: time.Second,
		Now:             c.Now,
	}
}

func start(t *testing.T, edit func(*Config)) *api {
	t.Helper()
	c := &clock{t: time.Date(2026, 3, 2, 5, 0, 0, 0, time.UTC)}
	cfg := testConfig(c)
	if edit != nil {
		edit(&cfg)
	}
	store := NewMemStore()
	if err := SeedRoutes(t.Context(), store); err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	payer := &FakePayer{}
	srv := NewServer(cfg, store, payer, slog.New(slog.NewTextHandler(&logs, nil)))
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return &api{t: t, url: ts.URL, clock: c, payer: payer, srv: srv, logs: &logs}
}

type resp struct {
	status int
	header http.Header
	body   []byte
}

func (r resp) decode(t *testing.T, v any) {
	t.Helper()
	if err := json.Unmarshal(r.body, v); err != nil {
		t.Fatalf("response %d %q isn't the JSON expected: %v", r.status, r.body, err)
	}
}

func (a *api) do(method, path, token string, body any, headers ...string) resp {
	a.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, a.url+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		a.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return resp{res.StatusCode, res.Header, b}
}

const goodPassword = "nimechoka na jam ya thika road"

// signup registers and logs in a user, and returns their access token.
func (a *api) signup(phone, name string) string {
	a.t.Helper()
	if r := a.do("POST", "/auth/register", "", map[string]string{"phone": phone, "name": name, "password": goodPassword}); r.status != 201 {
		a.t.Fatalf("register %s: %d %s", phone, r.status, r.body)
	}
	return a.login(phone)
}

func (a *api) login(phone string) string {
	a.t.Helper()
	r := a.do("POST", "/auth/login", "", map[string]string{"phone": phone, "password": goodPassword})
	var tok struct {
		AccessToken string `json:"accessToken"`
		TokenType   string `json:"tokenType"`
		ExpiresIn   int    `json:"expiresIn"`
	}
	r.decode(a.t, &tok)
	if r.status != 200 || tok.AccessToken == "" || tok.TokenType != "Bearer" || tok.ExpiresIn != 900 {
		a.t.Fatalf("login %s: %d %s, want 200 with a Bearer token that expires in 900 seconds", phone, r.status, r.body)
	}
	return tok.AccessToken
}

// tripOn creates a trip on the nth seeded route as admin and returns its ID.
func (a *api) tripOn(admin string, routeIndex int, departs time.Time, seats int) int {
	a.t.Helper()
	var routes []Route
	a.do("GET", "/routes", "", nil).decode(a.t, &routes)
	r := a.do("POST", "/trips", admin, map[string]any{"routeId": routes[routeIndex].ID, "departs": departs, "seats": seats})
	var trip Trip
	r.decode(a.t, &trip)
	if r.status != 201 {
		a.t.Fatalf("POST /trips: %d %s", r.status, r.body)
	}
	return trip.ID
}

func errorOf(t *testing.T, r resp) string {
	t.Helper()
	var e struct {
		Error string `json:"error"`
	}
	r.decode(t, &e)
	return e.Error
}

func TestHealth(t *testing.T) {
	a := start(t, nil)
	if r := a.do("GET", "/healthz", "", nil); r.status != 200 || !strings.Contains(string(r.body), `"ok"`) {
		t.Errorf("GET /healthz = %d %s", r.status, r.body)
	}
	if r := a.do("GET", "/readyz", "", nil); r.status != 200 {
		t.Errorf("GET /readyz = %d %s, want 200", r.status, r.body)
	}
	a.srv.SetDraining()
	if r := a.do("GET", "/readyz", "", nil); r.status != 503 {
		t.Errorf("GET /readyz while draining = %d, want 503", r.status)
	}
	if r := a.do("GET", "/healthz", "", nil); r.status != 200 {
		t.Errorf("GET /healthz while draining = %d, want 200", r.status)
	}
}

func TestPublicRoutes(t *testing.T) {
	a := start(t, nil)
	r := a.do("GET", "/routes", "", nil)
	var routes []Route
	r.decode(t, &routes)
	if r.status != 200 || len(routes) != 5 {
		t.Fatalf("GET /routes = %d with %d routes, want 200 and the 5 seeded routes", r.status, len(routes))
	}
	if routes[0].Number != "46K" || routes[0].To != "Kawangware" || routes[0].Fare != 100 {
		t.Errorf("first route = %+v, want 46K to Kawangware at Ksh100", routes[0])
	}
	if ct := r.header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	a := start(t, nil)
	r := a.do("POST", "/auth/register", "", map[string]string{"phone": "+254 712 345 678", "name": "Wanjiku", "password": goodPassword})
	if r.status != 201 {
		t.Fatalf("register: %d %s", r.status, r.body)
	}
	var u map[string]any
	r.decode(t, &u)
	if u["phone"] != "0712345678" || u["role"] != "passenger" || u["name"] != "Wanjiku" {
		t.Errorf("registered user = %v, want phone 0712345678, role passenger", u)
	}
	if strings.Contains(string(r.body), "$2a$") || strings.Contains(string(r.body), "assword") {
		t.Errorf("the response includes the password or its hash: %s", r.body)
	}

	tests := []struct {
		name string
		body map[string]string
		want int
	}{
		{"same phone again", map[string]string{"phone": "0712345678", "name": "W", "password": goodPassword}, 409},
		{"bad phone", map[string]string{"phone": "12345", "name": "W", "password": goodPassword}, 422},
		{"short password", map[string]string{"phone": "0722000111", "name": "W", "password": "matatu123"}, 422},
		{"no name", map[string]string{"phone": "0722000111", "name": "  ", "password": goodPassword}, 422},
		{"extra field", map[string]string{"phone": "0722000111", "name": "W", "password": goodPassword, "role": "admin"}, 400},
	}
	for _, tt := range tests {
		if r := a.do("POST", "/auth/register", "", tt.body); r.status != tt.want || errorOf(t, r) == "" {
			t.Errorf("register, %s: %d %s, want %d with an error", tt.name, r.status, r.body, tt.want)
		}
	}

	a.login("0712345678")
	wrong := a.do("POST", "/auth/login", "", map[string]string{"phone": "0712345678", "password": "not the password at all"})
	unknown := a.do("POST", "/auth/login", "", map[string]string{"phone": "0799999999", "password": goodPassword})
	if wrong.status != 401 || unknown.status != 401 || string(wrong.body) != string(unknown.body) {
		t.Errorf("wrong password: %d %s; unknown phone: %d %s; want the same 401 for both", wrong.status, wrong.body, unknown.status, unknown.body)
	}
	if strings.Contains(a.logs.String(), goodPassword) {
		t.Errorf("the log contains a password")
	}
}

func TestBookingJourney(t *testing.T) {
	a := start(t, nil)
	admin := a.signup("0700000001", "Amina")
	wanjiku := a.signup("0712345678", "Wanjiku")
	otieno := a.signup("0722000111", "Otieno")
	departs := time.Date(2026, 3, 2, 6, 30, 0, 0, time.UTC)
	tripID := a.tripOn(admin, 4, departs, 14) // route 58, Ksh150
	tripPath := "/trips/" + strconv.Itoa(tripID)

	r := a.do("POST", "/bookings", wanjiku, map[string]int{"tripId": tripID, "seat": 3})
	if r.status != 201 {
		t.Fatalf("POST /bookings: %d %s", r.status, r.body)
	}
	var b Booking
	r.decode(t, &b)
	loc := r.header.Get("Location")
	if loc != "/bookings/"+strconv.Itoa(b.ID) || b.Phone != "0712345678" || b.Passenger != "Wanjiku" || b.Fare != 150 || b.Paid {
		t.Errorf("booking = %+v at %q, want Wanjiku's, Ksh150, unpaid, with Location /bookings/<id>", b, loc)
	}

	if r := a.do("POST", "/bookings", otieno, map[string]int{"tripId": tripID, "seat": 3}); r.status != 409 {
		t.Errorf("booking a taken seat: %d %s, want 409", r.status, r.body)
	}

	var view struct {
		SeatsLeft  int   `json:"seatsLeft"`
		TakenSeats []int `json:"takenSeats"`
		Fare       int   `json:"fare"`
	}
	a.do("GET", tripPath, "", nil).decode(t, &view)
	if view.SeatsLeft != 13 || !slices.Equal(view.TakenSeats, []int{3}) || view.Fare != 150 {
		t.Errorf("GET %s = %+v, want 13 seats left, seat 3 taken, fare 150", tripPath, view)
	}

	if r := a.do("GET", loc, wanjiku, nil); r.status != 200 {
		t.Errorf("GET her own booking: %d", r.status)
	}
	theirs := a.do("GET", loc, otieno, nil)
	missing := a.do("GET", "/bookings/99999", otieno, nil)
	if theirs.status != 404 || string(theirs.body) != string(missing.body) {
		t.Errorf("GET someone else's booking: %d %s, want the same 404 as a missing one (%s)", theirs.status, theirs.body, missing.body)
	}

	if r := a.do("POST", loc+"/pay", otieno, nil); r.status != 404 {
		t.Errorf("paying for someone else's booking: %d, want 404", r.status)
	}
	r = a.do("POST", loc+"/pay", wanjiku, nil)
	var paid Booking
	r.decode(t, &paid)
	if r.status != 200 || !paid.Paid {
		t.Errorf("pay: %d %s, want 200 paid", r.status, r.body)
	}
	if r := a.do("POST", loc+"/pay", wanjiku, nil); r.status != 409 {
		t.Errorf("paying twice: %d, want 409", r.status)
	}

	var mine []Booking
	a.do("GET", "/me/bookings", wanjiku, nil).decode(t, &mine)
	if len(mine) != 1 || mine[0].ID != b.ID || !mine[0].Paid {
		t.Errorf("GET /me/bookings = %+v, want her one paid booking", mine)
	}
	var none []Booking
	a.do("GET", "/me/bookings", otieno, nil).decode(t, &none)
	if none == nil || len(none) != 0 {
		t.Errorf("GET /me/bookings for Otieno = %s, want []", a.do("GET", "/me/bookings", otieno, nil).body)
	}
}

func TestPaymentFailure(t *testing.T) {
	a := start(t, nil)
	admin := a.signup("0700000001", "Amina")
	p := a.signup("0712345678", "Wanjiku")
	trip := a.tripOn(admin, 0, time.Date(2026, 3, 2, 7, 0, 0, 0, time.UTC), 14)
	var b Booking
	a.do("POST", "/bookings", p, map[string]int{"tripId": trip, "seat": 1}).decode(t, &b)
	a.payer.Fail = true
	r := a.do("POST", "/bookings/"+strconv.Itoa(b.ID)+"/pay", p, nil)
	if r.status != http.StatusBadGateway || strings.Contains(string(r.body), ErrPaymentFailed.Error()) {
		t.Errorf("pay with M-Pesa down: %d %s, want 502 with a friendly message", r.status, r.body)
	}
	var after Booking
	a.do("GET", "/bookings/"+strconv.Itoa(b.ID), p, nil).decode(t, &after)
	if after.Paid {
		t.Errorf("a failed payment marked the booking paid")
	}
}

func TestBookingValidation(t *testing.T) {
	a := start(t, nil)
	admin := a.signup("0700000001", "Amina")
	p := a.signup("0712345678", "Wanjiku")
	trip := a.tripOn(admin, 1, time.Date(2026, 3, 2, 6, 0, 0, 0, time.UTC), 14)
	tests := []struct {
		name string
		body any
		want int
	}{
		{"unknown trip", map[string]int{"tripId": 99999, "seat": 1}, 404},
		{"seat 0", map[string]int{"tripId": trip, "seat": 0}, 422},
		{"seat 15 of 14", map[string]int{"tripId": trip, "seat": 15}, 422},
		{"name in the body", map[string]any{"tripId": trip, "seat": 2, "phone": "0722000111"}, 400},
		{"not JSON", "seat two please", 400},
	}
	for _, tt := range tests {
		if r := a.do("POST", "/bookings", p, tt.body); r.status != tt.want || errorOf(t, r) == "" {
			t.Errorf("%s: %d %s, want %d with an error", tt.name, r.status, r.body, tt.want)
		}
	}
	big := map[string]string{"tripId": strings.Repeat("9", 1<<20)}
	if r := a.do("POST", "/bookings", p, big); r.status != http.StatusRequestEntityTooLarge {
		t.Errorf("a body over 1 MB: %d, want 413", r.status)
	}

	a.clock.Add(90 * time.Minute) // 06:30: the 06:00 trip has left
	if r := a.do("POST", "/bookings", a.login("0712345678"), map[string]int{"tripId": trip, "seat": 5}); r.status != 409 {
		t.Errorf("booking a trip that has left: %d %s, want 409", r.status, r.body)
	}
}

func TestTripValidation(t *testing.T) {
	a := start(t, nil)
	admin := a.signup("0700000001", "Amina")
	var routes []Route
	a.do("GET", "/routes", "", nil).decode(t, &routes)
	future := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		body map[string]any
		want int
	}{
		{"ok", map[string]any{"routeId": routes[0].ID, "departs": future, "seats": 14}, 201},
		{"unknown route", map[string]any{"routeId": 99999, "departs": future, "seats": 14}, 422},
		{"in the past", map[string]any{"routeId": routes[0].ID, "departs": future.Add(-24 * time.Hour), "seats": 14}, 422},
		{"no seats", map[string]any{"routeId": routes[0].ID, "departs": future, "seats": 0}, 422},
		{"a bus, not a matatu", map[string]any{"routeId": routes[0].ID, "departs": future, "seats": 60}, 422},
	}
	for _, tt := range tests {
		if r := a.do("POST", "/trips", admin, tt.body); r.status != tt.want {
			t.Errorf("POST /trips, %s: %d %s, want %d", tt.name, r.status, r.body, tt.want)
		}
	}
	var trips []map[string]any
	a.do("GET", "/trips?route="+strconv.Itoa(routes[0].ID), "", nil).decode(t, &trips)
	if len(trips) != 1 || trips[0]["seatsLeft"] != float64(14) {
		t.Errorf("GET /trips?route=... = %v, want the one new trip with 14 seats left", trips)
	}
	if r := a.do("GET", "/trips?route=abc", "", nil); r.status != 400 {
		t.Errorf("GET /trips?route=abc = %d, want 400", r.status)
	}
	if r := a.do("POST", "/routes", admin, map[string]any{"number": "33", "name": "Test Sacco", "from": "Odeon", "to": "Embakasi", "fare": 80}); r.status != 201 || r.header.Get("Location") == "" {
		t.Errorf("POST /routes as admin: %d %s, want 201 with Location", r.status, r.body)
	}
	if r := a.do("POST", "/routes", admin, map[string]any{"number": "33", "name": "Test Sacco", "from": "Odeon", "to": "Embakasi", "fare": 0}); r.status != 422 {
		t.Errorf("POST /routes with fare 0: %d, want 422", r.status)
	}
}

func TestRoles(t *testing.T) {
	a := start(t, nil)
	admin := a.signup("0700000001", "Amina")
	passenger := a.signup("0712345678", "Wanjiku")
	a.signup("0733000222", "Kamau")
	trip := a.tripOn(admin, 2, time.Date(2026, 3, 2, 7, 0, 0, 0, time.UTC), 14)

	newTrip := map[string]any{"routeId": 1, "departs": time.Date(2026, 3, 3, 7, 0, 0, 0, time.UTC), "seats": 14}
	if r := a.do("POST", "/trips", passenger, newTrip); r.status != 403 {
		t.Errorf("passenger POST /trips: %d, want 403", r.status)
	}
	if r := a.do("POST", "/trips", "", newTrip); r.status != 401 || r.header.Get("WWW-Authenticate") != "Bearer" {
		t.Errorf("anonymous POST /trips: %d %q, want 401 with WWW-Authenticate: Bearer", r.status, r.header.Get("WWW-Authenticate"))
	}
	if r := a.do("PUT", "/users/0733000222/role", passenger, map[string]string{"role": "admin"}); r.status != 403 {
		t.Errorf("passenger making someone admin: %d, want 403", r.status)
	}

	r := a.do("PUT", "/users/0733000222/role", admin, map[string]string{"role": "conductor"})
	if r.status != 200 || !strings.Contains(string(r.body), `"conductor"`) {
		t.Fatalf("admin sets Kamau to conductor: %d %s", r.status, r.body)
	}
	if r := a.do("PUT", "/users/0733000222/role", admin, map[string]string{"role": "driver"}); r.status != 422 {
		t.Errorf("unknown role: %d, want 422", r.status)
	}
	if r := a.do("PUT", "/users/0799999999/role", admin, map[string]string{"role": "conductor"}); r.status != 404 {
		t.Errorf("unknown user: %d, want 404", r.status)
	}

	// Kamau logs in again, though his old token would work too: the role comes from the store.
	conductor := a.login("0733000222")
	var b Booking
	a.do("POST", "/bookings", passenger, map[string]int{"tripId": trip, "seat": 7}).decode(t, &b)
	if r := a.do("GET", "/bookings/"+strconv.Itoa(b.ID), conductor, nil); r.status != 200 {
		t.Errorf("conductor reads a passenger's booking: %d, want 200", r.status)
	}
	if r := a.do("POST", "/bookings/"+strconv.Itoa(b.ID)+"/pay", conductor, nil); r.status == 200 {
		t.Errorf("a conductor paid for a passenger's booking")
	}
	if r := a.do("POST", "/bookings", conductor, map[string]int{"tripId": trip, "seat": 8}); r.status != 403 {
		t.Errorf("conductor books a seat: %d, want 403", r.status)
	}
}

func TestTokens(t *testing.T) {
	a := start(t, nil)
	tok := a.signup("0712345678", "Wanjiku")
	if r := a.do("GET", "/me/bookings", tok, nil); r.status != 200 {
		t.Fatalf("fresh token: %d", r.status)
	}
	bad := []string{"garbage", tok[:len(tok)-3] + "abc", "eyJhbGciOiJub25lIn0.eyJzdWIiOiIwNzEyMzQ1Njc4IiwiZXhwIjo5OTk5OTk5OTk5fQ."}
	for _, b := range bad {
		r := a.do("GET", "/me/bookings", b, nil)
		if r.status != 401 || !strings.Contains(r.header.Get("WWW-Authenticate"), "invalid_token") {
			t.Errorf("token %q: %d %q, want 401 with error=\"invalid_token\"", b, r.status, r.header.Get("WWW-Authenticate"))
		}
	}
	a.clock.Add(16 * time.Minute)
	if r := a.do("GET", "/me/bookings", tok, nil); r.status != 401 {
		t.Errorf("a 16-minute-old token: %d, want 401", r.status)
	}
}

func TestDoubleBookingRace(t *testing.T) {
	a := start(t, nil)
	admin := a.signup("0700000001", "Amina")
	trip := a.tripOn(admin, 3, time.Date(2026, 3, 2, 7, 0, 0, 0, time.UTC), 14)
	tokens := make([]string, 6)
	for i := range tokens {
		tokens[i] = a.signup("07120000"+strconv.Itoa(10+i), "Abiria")
	}
	body := `{"tripId":` + strconv.Itoa(trip) + `,"seat":14}`
	var wg sync.WaitGroup
	var mu sync.Mutex
	codes := map[int]int{}
	for _, tok := range tokens {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", a.url+"/bookings", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+tok)
			status := -1
			if res, err := http.DefaultClient.Do(req); err == nil {
				status = res.StatusCode
				res.Body.Close()
			}
			mu.Lock()
			codes[status]++
			mu.Unlock()
		}()
	}
	wg.Wait()
	if codes[201] != 1 || codes[409] != 5 {
		t.Errorf("6 passengers booked the last seat at once: %v, want one 201 and five 409s", codes)
	}
}

func TestSecurityHeadersAndCORS(t *testing.T) {
	a := start(t, nil)
	r := a.do("GET", "/routes", "", nil, "Origin", "https://matatu.example")
	for k, v := range map[string]string{
		"Content-Security-Policy":          "default-src 'none'; frame-ancestors 'none'",
		"Strict-Transport-Security":        "max-age=63072000; includeSubDomains",
		"X-Content-Type-Options":           "nosniff",
		"Referrer-Policy":                  "no-referrer",
		"Access-Control-Allow-Origin":      "https://matatu.example",
		"Access-Control-Allow-Credentials": "true",
	} {
		if got := r.header.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if !slices.Contains(r.header.Values("Vary"), "Origin") {
		t.Errorf("Vary = %q, want Origin", r.header.Values("Vary"))
	}
	if r := a.do("GET", "/routes", "", nil, "Origin", "https://evil.example"); r.header.Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("an unknown origin got Access-Control-Allow-Origin")
	}
	pre := a.do("OPTIONS", "/bookings", "", nil, "Origin", "https://matatu.example", "Access-Control-Request-Method", "POST")
	if pre.status != 204 || !strings.Contains(pre.header.Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Errorf("preflight: %d %v, want 204 allowing Authorization", pre.status, pre.header)
	}
	if r := a.do("OPTIONS", "/bookings", "", nil, "Origin", "https://evil.example", "Access-Control-Request-Method", "POST"); r.status != 403 {
		t.Errorf("preflight from another origin: %d, want 403", r.status)
	}
	if r := a.do("DELETE", "/routes", "", nil); r.status != 405 {
		t.Errorf("DELETE /routes: %d, want 405", r.status)
	}
}

func TestRateLimit(t *testing.T) {
	a := start(t, func(c *Config) { c.RatePerSecond = 0.1; c.RateBurst = 3 })
	for i := 1; i <= 3; i++ {
		if r := a.do("GET", "/routes", "", nil); r.status != 200 {
			t.Fatalf("request %d: %d, want 200", i, r.status)
		}
	}
	r := a.do("GET", "/routes", "", nil, "Origin", "https://matatu.example")
	if r.status != 429 || r.header.Get("Retry-After") != "10" {
		t.Errorf("request 4: %d Retry-After %q, want 429 and 10", r.status, r.header.Get("Retry-After"))
	}
	if r.header.Get("Access-Control-Allow-Origin") != "https://matatu.example" || r.header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("the 429 is missing CORS or security headers, so the front end can't read it")
	}
	a.clock.Add(10 * time.Second)
	if r := a.do("GET", "/routes", "", nil); r.status != 200 {
		t.Errorf("after 10 s: %d, want 200", r.status)
	}
}

func TestRateLimitBehindAProxy(t *testing.T) {
	a := start(t, func(c *Config) { c.RatePerSecond = 0.1; c.RateBurst = 1; c.ClientIPHeader = "X-Forwarded-For" })
	if r := a.do("GET", "/routes", "", nil, "X-Forwarded-For", "41.90.1.1"); r.status != 200 {
		t.Fatalf("first passenger: %d", r.status)
	}
	if r := a.do("GET", "/routes", "", nil, "X-Forwarded-For", "41.90.2.2"); r.status != 200 {
		t.Errorf("a second passenger through the same proxy: %d, want 200: each client IP has its own bucket", r.status)
	}
	if r := a.do("GET", "/routes", "", nil, "X-Forwarded-For", "6.6.6.6, 41.90.1.1"); r.status != 429 {
		t.Errorf("the first passenger again, with a made-up address in front: %d, want 429 (use the last entry)", r.status)
	}
}

func TestAccessLog(t *testing.T) {
	a := start(t, nil)
	tok := a.signup("0712345678", "Wanjiku")
	a.do("GET", "/me/bookings", tok, nil)
	logs := a.logs.String()
	if !strings.Contains(logs, "path=/me/bookings") || !strings.Contains(logs, "status=200") {
		t.Errorf("want one access log line per request with the path and status, got:\n%s", logs)
	}
	if strings.Contains(logs, tok) || strings.Contains(logs, goodPassword) {
		t.Errorf("the log contains a token or password")
	}
}
