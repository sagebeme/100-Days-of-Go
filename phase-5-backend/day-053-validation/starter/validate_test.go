package validation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNormalizePhone(t *testing.T) {
	good := []struct {
		raw  string
		want string
	}{
		{"0712 345 678", "254712345678"},
		{"+254712345678", "254712345678"},
		{"254712345678", "254712345678"},
		{"0712-345-678", "254712345678"},
		{"712345678", "254712345678"},
		{"0110 123 456", "254110123456"},
		{"  0722000111 ", "254722000111"},
	}
	for _, tt := range good {
		got, err := NormalizePhone(tt.raw)
		if err != nil || got != tt.want {
			t.Errorf("NormalizePhone(%q) = %q, %v, want %q, nil", tt.raw, got, err, tt.want)
		}
	}
	bad := []string{"", "   ", "071234567", "07123456789", "0712a45678", "+0712345678", "0712+345678", "0201234567", "255712345678", "0712\n345678"}
	for _, raw := range bad {
		if got, err := NormalizePhone(raw); err == nil || got != "" {
			t.Errorf("NormalizePhone(%q) = %q, %v, want \"\" and an error", raw, got, err)
		}
	}
}

func TestErrorsError(t *testing.T) {
	e := Errors{{Field: "seat", Message: "must be between 1 and 14"}, {Field: "phone", Message: "is required"}}
	want := "seat must be between 1 and 14; phone is required"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

var (
	now  = time.Date(2026, 11, 2, 6, 0, 0, 0, time.UTC)
	trip = Trip{ID: 1, RouteID: 125, Departs: now.Add(90 * time.Minute), Seats: 14}
	gone = Trip{ID: 2, RouteID: 111, Departs: now.Add(-time.Minute), Seats: 14}
)

func TestValidateBooking(t *testing.T) {
	tests := []struct {
		name  string
		req   BookingRequest
		trip  Trip
		want  BookingRequest // the cleaned request, checked when there are no errors
		wantE Errors
	}{
		{
			name: "all good, and cleaned up",
			req:  BookingRequest{Seat: 3, Passenger: "  Achieng' Odhiambo ", Phone: "0712 345 678"},
			trip: trip,
			want: BookingRequest{Seat: 3, Passenger: "Achieng' Odhiambo", Phone: "254712345678"},
		},
		{
			name: "letters from other alphabets",
			req:  BookingRequest{Seat: 14, Passenger: "Wanjirũ Ng'ang'a", Phone: "+254 110 123456"},
			trip: trip,
			want: BookingRequest{Seat: 14, Passenger: "Wanjirũ Ng'ang'a", Phone: "254110123456"},
		},
		{
			name: "full stops and hyphens",
			req:  BookingRequest{Seat: 1, Passenger: "J.K. Mwangi-Kamau", Phone: "0722000111"},
			trip: trip,
			want: BookingRequest{Seat: 1, Passenger: "J.K. Mwangi-Kamau", Phone: "254722000111"},
		},
		{
			name:  "blank passenger",
			req:   BookingRequest{Seat: 3, Passenger: "   ", Phone: "0712345678"},
			trip:  trip,
			wantE: Errors{{"passenger", "is required"}},
		},
		{
			name:  "one letter",
			req:   BookingRequest{Seat: 3, Passenger: "K", Phone: "0712345678"},
			trip:  trip,
			wantE: Errors{{"passenger", "must be 2 to 60 characters"}},
		},
		{
			name:  "61 letters",
			req:   BookingRequest{Seat: 3, Passenger: strings.Repeat("a", 61), Phone: "0712345678"},
			trip:  trip,
			wantE: Errors{{"passenger", "must be 2 to 60 characters"}},
		},
		{
			name: "60 two-byte letters is 60 characters",
			req:  BookingRequest{Seat: 3, Passenger: strings.Repeat("ũ", 60), Phone: "0712345678"},
			trip: trip,
			want: BookingRequest{Seat: 3, Passenger: strings.Repeat("ũ", 60), Phone: "254712345678"},
		},
		{
			name:  "a script tag",
			req:   BookingRequest{Seat: 3, Passenger: "<script>alert(1)</script>", Phone: "0712345678"},
			trip:  trip,
			wantE: Errors{{"passenger", "may only contain letters, spaces, hyphens, apostrophes and full stops"}},
		},
		{
			name:  "a control character",
			req:   BookingRequest{Seat: 3, Passenger: "Kip\x00Rono", Phone: "0712345678"},
			trip:  trip,
			wantE: Errors{{"passenger", "may only contain letters, spaces, hyphens, apostrophes and full stops"}},
		},
		{
			name:  "landline",
			req:   BookingRequest{Seat: 3, Passenger: "Kip", Phone: "020 123 4567"},
			trip:  trip,
			wantE: Errors{{"phone", "must be a Kenyan mobile number like 0712 345 678"}},
		},
		{
			name:  "no phone",
			req:   BookingRequest{Seat: 3, Passenger: "Kip"},
			trip:  trip,
			wantE: Errors{{"phone", "is required"}},
		},
		{
			name:  "seat 15 on a 14-seater",
			req:   BookingRequest{Seat: 15, Passenger: "Kip", Phone: "0712345678"},
			trip:  trip,
			wantE: Errors{{"seat", "must be between 1 and 14"}},
		},
		{
			name:  "seat 0",
			req:   BookingRequest{Seat: 0, Passenger: "Kip", Phone: "0712345678"},
			trip:  trip,
			wantE: Errors{{"seat", "must be between 1 and 14"}},
		},
		{
			name:  "trip leaving right now",
			req:   BookingRequest{Seat: 3, Passenger: "Kip", Phone: "0712345678"},
			trip:  Trip{ID: 3, Departs: now, Seats: 14},
			wantE: Errors{{"trip", "has already left"}},
		},
		{
			name: "everything wrong at once, in order",
			req:  BookingRequest{Seat: -1, Passenger: "", Phone: "hello"},
			trip: gone,
			wantE: Errors{
				{"passenger", "is required"},
				{"phone", "must be a Kenyan mobile number like 0712 345 678"},
				{"seat", "must be between 1 and 14"},
				{"trip", "has already left"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errs := ValidateBooking(tt.req, tt.trip, now)
			if !reflect.DeepEqual(errs, tt.wantE) {
				t.Fatalf("errors\n got: %#v\nwant: %#v", errs, tt.wantE)
			}
			if tt.wantE == nil && got != tt.want {
				t.Errorf("cleaned request\n got: %+v\nwant: %+v", got, tt.want)
			}
		})
	}
}

func newAPI(t *testing.T) (*MemStore, http.Handler) {
	t.Helper()
	s := NewMemStore(trip, gone)
	h := NewAPI(s, func() time.Time { return now }).Routes()
	if h == nil {
		t.Fatal("Routes() returned nil")
	}
	return s, h
}

func post(h http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCreateBooking(t *testing.T) {
	s, h := newAPI(t)
	rec := post(h, "/trips/1/bookings", `{"seat": 3, "passenger": " Achieng ", "phone": "0712 345 678"}`)
	if rec.Code != 201 {
		t.Fatalf("status = %d, want 201 (body %q)", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "/bookings/1" {
		t.Errorf("Location = %q, want /bookings/1", loc)
	}
	var b Booking
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("body %q isn't a booking: %v", rec.Body.String(), err)
	}
	want := Booking{ID: 1, TripID: 1, Seat: 3, Passenger: "Achieng", Phone: "254712345678"}
	if b != want {
		t.Errorf("booking\n got: %+v\nwant: %+v", b, want)
	}
	if got := s.Bookings(); len(got) != 1 || got[0] != want {
		t.Errorf("store holds %+v, want the cleaned booking %+v", got, want)
	}
}

func TestRejectsEarly(t *testing.T) {
	big := `{"seat": 3, "passenger": "` + strings.Repeat("a", MaxBody) + `", "phone": "0712345678"}`
	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"body too big", "/trips/1/bookings", big, 413},
		{"broken JSON", "/trips/1/bookings", `{"seat": 3`, 400},
		{"unknown field", "/trips/1/bookings", `{"seat": 3, "passenger": "Kip", "phone": "0712345678", "fare": 1}`, 400},
		{"bad trip id", "/trips/x/bookings", `{"seat": 3, "passenger": "Kip", "phone": "0712345678"}`, 400},
		{"no such trip", "/trips/9/bookings", `{"seat": 3, "passenger": "Kip", "phone": "0712345678"}`, 404},
		{"rules broken", "/trips/1/bookings", `{"seat": 30, "passenger": "Kip", "phone": "0712345678"}`, 422},
		{"trip has left", "/trips/2/bookings", `{"seat": 3, "passenger": "Kip", "phone": "0712345678"}`, 422},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, h := newAPI(t)
			rec := post(h, tt.path, tt.body)
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d (body %.200q)", rec.Code, tt.status, rec.Body.String())
			}
			var e struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e.Error == "" {
				t.Errorf("body = %.200q, want {\"error\": \"...\"}", rec.Body.String())
			}
			if n := len(s.Bookings()); n != 0 {
				t.Errorf("a rejected request saved %d booking(s)", n)
			}
		})
	}
}

func TestValidationBody(t *testing.T) {
	_, h := newAPI(t)
	rec := post(h, "/trips/2/bookings", `{"seat": 0, "passenger": "", "phone": "0712"}`)
	if rec.Code != 422 {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	var body struct {
		Error  string       `json:"error"`
		Fields []FieldError `json:"fields"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q: %v", rec.Body.String(), err)
	}
	if body.Error != "validation failed" {
		t.Errorf("error = %q, want %q", body.Error, "validation failed")
	}
	var fields []string
	for _, f := range body.Fields {
		fields = append(fields, f.Field)
	}
	if want := []string{"passenger", "phone", "seat", "trip"}; !reflect.DeepEqual(fields, want) {
		t.Errorf("fields = %v, want %v: report every problem at once", fields, want)
	}
	if !strings.Contains(rec.Body.String(), `"field":"seat","message":"must be between 1 and 14"`) {
		t.Errorf("body %s should hold {\"field\":\"seat\",\"message\":...}", rec.Body.String())
	}
}

func TestSeatTaken(t *testing.T) {
	_, h := newAPI(t)
	body := `{"seat": 3, "passenger": "Kip", "phone": "0712345678"}`
	if rec := post(h, "/trips/1/bookings", body); rec.Code != 201 {
		t.Fatalf("first booking: status = %d, want 201", rec.Code)
	}
	if rec := post(h, "/trips/1/bookings", body); rec.Code != 409 {
		t.Errorf("same seat again: status = %d, want 409", rec.Code)
	}
}
