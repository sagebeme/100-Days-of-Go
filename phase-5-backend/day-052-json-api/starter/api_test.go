package jsonapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type tripBody struct {
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

func TestDecodeJSON(t *testing.T) {
	const good = `{"routeId": 125, "departs": "2026-11-02T07:30:00+03:00", "seats": 14}`
	tests := []struct {
		name        string
		contentType string
		body        string
		status      int    // 0 means no error
		msgHas      string // a piece the message must contain
	}{
		{"good", "application/json", good, 0, ""},
		{"charset is fine", "application/json; charset=utf-8", good, 0, ""},
		{"no Content-Type", "", good, 415, "application/json"},
		{"plain text", "text/plain", good, 415, "application/json"},
		{"a form", "application/x-www-form-urlencoded", "routeId=125", 415, "application/json"},
		{"empty body", "application/json", "", 400, "empty"},
		{"cut off", "application/json", `{"routeId": 125,`, 400, "malformed"},
		{"bad syntax", "application/json", `{"routeId": 125,, "seats": 14}`, 400, "malformed JSON at byte"},
		{"string for a number", "application/json", `{"seats": "ten"}`, 400, `field "seats" must be a whole number`},
		{"fraction for a number", "application/json", `{"seats": 1.5}`, 400, `field "seats" must be a whole number`},
		{"a list, not an object", "application/json", `[1, 2]`, 400, "JSON object"},
		{"time with no offset", "application/json", `{"departs": "2026-11-02T07:30:00"}`, 400, "2026-11-02T07:30:00+03:00"},
		{"time in words", "application/json", `{"departs": "kesho asubuhi"}`, 400, "2026-11-02T07:30:00+03:00"},
		{"unknown field", "application/json", `{"routeId": 125, "seat": 3}`, 400, `unknown field "seat"`},
		{"two values", "application/json", good + good, 400, "single JSON value"},
		{"junk after", "application/json", good + " x", 400, "single JSON value"},
		{"spaces after are fine", "application/json", good + "\n\n  ", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/trips", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			var got tripBody
			err := DecodeJSON(req, &got)
			if tt.status == 0 {
				if err != nil {
					t.Fatalf("DecodeJSON error = %v, want none", err)
				}
				if got.RouteID != 125 || got.Seats != 14 || got.Departs.IsZero() {
					t.Errorf("decoded %+v, want route 125, 14 seats and a time", got)
				}
				return
			}
			var re *RequestError
			if !errors.As(err, &re) {
				t.Fatalf("DecodeJSON error = %v (%T), want a *RequestError", err, err)
			}
			if re.Status != tt.status {
				t.Errorf("Status = %d, want %d (message %q)", re.Status, tt.status, re.Message)
			}
			if !strings.Contains(re.Message, tt.msgHas) {
				t.Errorf("Message = %q, want it to contain %q", re.Message, tt.msgHas)
			}
			if re.Error() != re.Message {
				t.Errorf("Error() = %q, want the Message %q", re.Error(), re.Message)
			}
			if strings.Contains(re.Message, "tripBody") || strings.Contains(re.Message, "jsonapi") {
				t.Errorf("Message %q shows Go type names to the client", re.Message)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	trip := Trip{ID: 7, RouteID: 111, Departs: time.Date(2026, 11, 2, 4, 30, 0, 0, time.UTC), Seats: 33}
	b, err := json.Marshal(ToJSON(trip))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":7,"routeId":111,"departs":"2026-11-02T07:30:00+03:00","seats":33}`
	if string(b) != want {
		t.Errorf("json.Marshal(ToJSON(trip))\n got: %s\nwant: %s", b, want)
	}
}

func post(h http.Handler, contentType, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/trips", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	return rec
}

func newAPI(t *testing.T) (*MemStore, http.Handler) {
	t.Helper()
	s := NewMemStore()
	h := NewAPI(s).Routes()
	if h == nil {
		t.Fatal("Routes() returned nil")
	}
	return s, h
}

func TestCreateTrip(t *testing.T) {
	s, h := newAPI(t)
	// 04:30 UTC is 07:30 in Nairobi.
	rec := post(h, "application/json", `{"routeId": 125, "departs": "2026-11-02T04:30:00Z", "seats": 14}`)
	if rec.Code != 201 {
		t.Fatalf("status = %d, want 201 (body %q)", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if loc := rec.Header().Get("Location"); loc != "/trips/1" {
		t.Errorf("Location = %q, want /trips/1", loc)
	}
	want := `{"id":1,"routeId":125,"departs":"2026-11-02T07:30:00+03:00","seats":14}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body\n got: %s\nwant: %s", got, want)
	}
	stored, ok := s.Trip(1)
	if !ok || stored.Departs.Location() != time.UTC {
		t.Errorf("stored trip = %+v, want it saved with Departs in UTC", stored)
	}

	tests := []struct {
		name        string
		contentType string
		body        string
		status      int
	}{
		{"form post", "application/x-www-form-urlencoded", "routeId=125&seats=14", 415},
		{"unknown field", "application/json", `{"routeId": 125, "departs": "2026-11-02T04:30:00Z", "seats": 14, "fare": 100}`, 400},
		{"string seats", "application/json", `{"routeId": 125, "departs": "2026-11-02T04:30:00Z", "seats": "14"}`, 400},
		{"zero seats", "application/json", `{"routeId": 125, "departs": "2026-11-02T04:30:00Z", "seats": 0}`, 422},
		{"no departs", "application/json", `{"routeId": 125, "seats": 14}`, 422},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(h, tt.contentType, tt.body)
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d (body %q)", rec.Code, tt.status, rec.Body.String())
			}
			var e struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e.Error == "" {
				t.Errorf("body = %q, want {\"error\": \"...\"}", rec.Body.String())
			}
		})
	}
	if n := len(s.Trips()); n != 1 {
		t.Errorf("store has %d trips, want 1: rejected requests must not save anything", n)
	}
}

func TestListTripsByNairobiDate(t *testing.T) {
	s, h := newAPI(t)
	s.CreateTrip(Trip{RouteID: 125, Departs: time.Date(2026, 11, 2, 7, 30, 0, 0, Nairobi), Seats: 14})
	// 00:30 on 3 November in Nairobi, but still 2 November in UTC.
	s.CreateTrip(Trip{RouteID: 111, Departs: time.Date(2026, 11, 3, 0, 30, 0, 0, Nairobi), Seats: 33})
	// 00:15 on 2 November in Nairobi, but 1 November in UTC.
	s.CreateTrip(Trip{RouteID: 58, Departs: time.Date(2026, 11, 2, 0, 15, 0, 0, Nairobi), Seats: 14})

	ids := func(rec *httptest.ResponseRecorder) []int {
		t.Helper()
		var trips []TripJSON
		if err := json.Unmarshal(rec.Body.Bytes(), &trips); err != nil {
			t.Fatalf("body %q isn't a list of trips: %v", rec.Body.String(), err)
		}
		out := []int{}
		for _, tr := range trips {
			out = append(out, tr.ID)
		}
		return out
	}

	tests := []struct {
		path string
		want []int
	}{
		{"/trips", []int{3, 1, 2}},
		{"/trips?date=2026-11-02", []int{3, 1}},
		{"/trips?date=2026-11-03", []int{2}},
		{"/trips?date=2026-11-04", []int{}},
	}
	for _, tt := range tests {
		rec := get(h, tt.path)
		if rec.Code != 200 {
			t.Fatalf("GET %s: status = %d, want 200", tt.path, rec.Code)
		}
		got := ids(rec)
		if len(got) != len(tt.want) {
			t.Errorf("GET %s: trip IDs %v, want %v", tt.path, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("GET %s: trip IDs %v, want %v", tt.path, got, tt.want)
				break
			}
		}
	}
	if body := strings.TrimSpace(get(h, "/trips?date=2026-11-04").Body.String()); body != "[]" {
		t.Errorf("no trips that day: body = %q, want []", body)
	}
	if !strings.Contains(get(h, "/trips").Body.String(), `"departs":"2026-11-03T00:30:00+03:00"`) {
		t.Error("GET /trips doesn't show times in Nairobi time with their offset")
	}

	for _, bad := range []string{"/trips?date=02-11-2026", "/trips?date=2026-13-01", "/trips?date=kesho"} {
		if rec := get(h, bad); rec.Code != 400 {
			t.Errorf("GET %s: status = %d, want 400", bad, rec.Code)
		}
	}
}

func TestGetTrip(t *testing.T) {
	s, h := newAPI(t)
	s.CreateTrip(Trip{RouteID: 125, Departs: time.Date(2026, 11, 2, 7, 30, 0, 0, Nairobi), Seats: 14})
	tests := []struct {
		path   string
		status int
		body   string
	}{
		{"/trips/1", 200, `{"id":1,"routeId":125,"departs":"2026-11-02T07:30:00+03:00","seats":14}`},
		{"/trips/2", 404, ""},
		{"/trips/one", 400, ""},
	}
	for _, tt := range tests {
		rec := get(h, tt.path)
		if rec.Code != tt.status {
			t.Errorf("GET %s: status = %d, want %d", tt.path, rec.Code, tt.status)
		}
		if tt.body != "" && strings.TrimSpace(rec.Body.String()) != tt.body {
			t.Errorf("GET %s\n got: %s\nwant: %s", tt.path, rec.Body.String(), tt.body)
		}
	}
}
