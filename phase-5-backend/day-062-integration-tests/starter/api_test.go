package integrationtests

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoadFixtures(t *testing.T) {
	f := LoadFixtures(t, "testdata/fixtures.json")
	if len(f.Trips) != 2 || len(f.Users) != 2 {
		t.Fatalf("LoadFixtures: %d trips and %d users, want 2 and 2", len(f.Trips), len(f.Users))
	}
	want := time.Date(2026, 3, 2, 6, 30, 0, 0, time.UTC)
	if f.Trips[0].RouteID != 111 || !f.Trips[0].Departs.Equal(want) || f.Trips[0].Seats != 14 {
		t.Errorf("first trip = %+v, want route 111 at %v with 14 seats", f.Trips[0], want)
	}
	if f.Users[1].Token != "tok-otieno" || f.Users[1].Phone != "0722000111" {
		t.Errorf("second user = %+v", f.Users[1])
	}
}

func TestSeed(t *testing.T) {
	s := NewMemStore()
	f := LoadFixtures(t, "testdata/fixtures.json")
	Seed(t, s, &f)
	for _, trip := range f.Trips {
		if trip.ID == 0 {
			t.Fatalf("Seed left a trip without its stored ID: %+v", trip)
		}
		got, err := s.Trip(t.Context(), trip.ID)
		if err != nil || got.RouteID != trip.RouteID {
			t.Errorf("store has %+v, %v for trip %d, want %+v", got, err, trip.ID, trip)
		}
	}
}

func TestEnvClosesServerOnCleanup(t *testing.T) {
	var url string
	t.Run("inner", func(t *testing.T) {
		env := NewEnv(t, NewMemStore())
		url = env.URL
		if !strings.HasPrefix(url, "http://127.0.0.1:") {
			t.Fatalf("env.URL = %q, want the httptest server's address", url)
		}
		if r := env.Do(t, "GET", "/healthz", "", nil); r.Status != 200 {
			t.Fatalf("GET /healthz inside the test: %d", r.Status)
		}
	})
	c := &http.Client{Timeout: time.Second}
	if resp, err := c.Get(url + "/healthz"); err == nil {
		resp.Body.Close()
		t.Errorf("the server still answers after its test ended: register srv.Close with t.Cleanup")
	}
}

func TestHealth(t *testing.T) {
	env := NewEnv(t, OpenTestStore(t))
	r := env.Do(t, "GET", "/healthz", "", nil)
	var body map[string]string
	r.JSON(t, &body)
	if r.Status != 200 || body["status"] != "ok" {
		t.Errorf("GET /healthz = %d %s, want 200 {\"status\":\"ok\"}", r.Status, r.Body)
	}
	if got := r.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := r.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff on every response", got)
	}
}

// TestBookingJourney follows one passenger through the API, the way the phone app would.
func TestBookingJourney(t *testing.T) {
	env := NewEnv(t, OpenTestStore(t))
	trip := env.Fixtures.Trips[0]
	tripPath := "/trips/" + strconv.Itoa(trip.ID)

	var before struct {
		ID        int `json:"id"`
		SeatsLeft int `json:"seatsLeft"`
	}
	r := env.Do(t, "GET", tripPath, "", nil)
	r.JSON(t, &before)
	if r.Status != 200 || before.ID != trip.ID || before.SeatsLeft != 14 {
		t.Fatalf("GET %s = %d %s, want 200 with 14 seats left", tripPath, r.Status, r.Body)
	}

	book := map[string]any{"tripId": trip.ID, "seat": 3, "passenger": "Wanjiku"}
	r = env.Do(t, "POST", "/bookings", "tok-wanjiku", book)
	if r.Status != http.StatusCreated {
		t.Fatalf("POST /bookings = %d %s, want 201", r.Status, r.Body)
	}
	loc := r.Header.Get("Location")
	var created Booking
	r.JSON(t, &created)
	if loc != "/bookings/"+strconv.Itoa(created.ID) {
		t.Errorf("Location = %q, want /bookings/%d", loc, created.ID)
	}
	if created.Phone != "0712345678" || created.Seat != 3 {
		t.Errorf("created = %+v, want Wanjiku's phone and seat 3", created)
	}

	r = env.Do(t, "GET", loc, "tok-wanjiku", nil)
	var got Booking
	r.JSON(t, &got)
	if r.Status != 200 || got != created {
		t.Errorf("GET %s = %d %+v, want 200 %+v", loc, r.Status, got, created)
	}

	if r := env.Do(t, "GET", loc, "tok-otieno", nil); r.Status != 404 {
		t.Errorf("GET %s as Otieno = %d, want 404: it's not his booking", loc, r.Status)
	}
	if r := env.Do(t, "POST", "/bookings", "tok-otieno", book); r.Status != http.StatusConflict {
		t.Errorf("booking the same seat again = %d %s, want 409", r.Status, r.Body)
	}

	var after struct{ SeatsLeft int }
	env.Do(t, "GET", tripPath, "", nil).JSON(t, &after)
	if after.SeatsLeft != 13 {
		t.Errorf("seatsLeft after one booking = %d, want 13", after.SeatsLeft)
	}
}

func TestAuthIsWiredUp(t *testing.T) {
	env := NewEnv(t, OpenTestStore(t))
	book := map[string]any{"tripId": env.Fixtures.Trips[0].ID, "seat": 1, "passenger": "Mgeni"}
	for _, tok := range []string{"", "tok-made-up"} {
		r := env.Do(t, "POST", "/bookings", tok, book)
		if r.Status != http.StatusUnauthorized {
			t.Errorf("POST /bookings with token %q = %d %s, want 401", tok, r.Status, r.Body)
		}
		if r.Header.Get("WWW-Authenticate") != "Bearer" {
			t.Errorf("401 without WWW-Authenticate: Bearer")
		}
	}
	if r := env.Do(t, "GET", "/bookings/1", "", nil); r.Status != http.StatusUnauthorized {
		t.Errorf("GET /bookings/1 with no token = %d, want 401", r.Status)
	}
}

func TestBookingErrors(t *testing.T) {
	env := NewEnv(t, OpenTestStore(t))
	id := env.Fixtures.Trips[0].ID
	tests := []struct {
		name string
		body any
		want int
	}{
		{"unknown trip", map[string]any{"tripId": 999999, "seat": 1, "passenger": "W"}, 404},
		{"seat 0", map[string]any{"tripId": id, "seat": 0, "passenger": "W"}, 422},
		{"seat 15 of 14", map[string]any{"tripId": id, "seat": 15, "passenger": "W"}, 422},
		{"no name", map[string]any{"tripId": id, "seat": 2, "passenger": " "}, 422},
		{"unknown field", map[string]any{"tripId": id, "seat": 2, "passenger": "W", "price": 0}, 400},
		{"not an object", []int{1, 2}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := env.Do(t, "POST", "/bookings", "tok-wanjiku", tt.body)
			var e struct{ Error string }
			r.JSON(t, &e)
			if r.Status != tt.want || e.Error == "" {
				t.Errorf("status %d body %s, want %d with an \"error\"", r.Status, r.Body, tt.want)
			}
		})
	}
	if r := env.Do(t, "DELETE", "/bookings", "tok-wanjiku", nil); r.Status != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /bookings = %d, want 405 from the mux", r.Status)
	}
}

func TestLastSeatRace(t *testing.T) {
	env := NewEnv(t, OpenTestStore(t))
	trip := env.Fixtures.Trips[1] // 2 seats
	var wg sync.WaitGroup
	var mu sync.Mutex
	codes := map[int]int{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// env.Do calls t.Fatal on errors, which only the test's own goroutine may
			// do, so these requests are sent by hand.
			body := `{"tripId":` + strconv.Itoa(trip.ID) + `,"seat":2,"passenger":"Otieno"}`
			req, _ := http.NewRequest("POST", env.URL+"/bookings", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer tok-otieno")
			status := -1 // a network error
			if resp, err := env.Client.Do(req); err == nil {
				status = resp.StatusCode
				resp.Body.Close()
			}
			mu.Lock()
			codes[status]++
			mu.Unlock()
		}()
	}
	wg.Wait()
	if codes[201] != 1 || codes[409] != 9 {
		t.Errorf("10 requests for the same seat: %v, want exactly one 201 and nine 409s", codes)
	}
}
