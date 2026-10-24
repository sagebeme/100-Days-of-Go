package httpmethods

// This file is already written for you, and it's the same in starter/ and solution/.
// It holds the domain types, an in-memory store and the JSON helpers. Today's work is
// in api.go.

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Trip is one matatu departure on a route.
type Trip struct {
	ID      int       `json:"id"`
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

// Booking is one passenger's seat on a trip.
type Booking struct {
	ID        int    `json:"id"`
	TripID    int    `json:"tripId"`
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
}

// Errors the store returns from Book.
var (
	ErrNoTrip    = errors.New("no such trip")
	ErrSeatTaken = errors.New("seat already booked")
)

// Store keeps trips and bookings in memory. It is safe to use from many goroutines.
type Store struct {
	mu          sync.Mutex
	trips       map[int]Trip
	bookings    map[int]Booking
	nextTrip    int
	nextBooking int
}

// NewStore returns an empty store. The first trip and the first booking get ID 1.
func NewStore() *Store {
	return &Store{trips: map[int]Trip{}, bookings: map[int]Booking{}, nextTrip: 1, nextBooking: 1}
}

// Trips returns every trip, sorted by ID. It never returns nil.
func (s *Store) Trips() []Trip {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Trip, 0, len(s.trips))
	for _, t := range s.trips {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Trip returns the trip with this ID, and false if there isn't one.
func (s *Store) Trip(id int) (Trip, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.trips[id]
	return t, ok
}

// CreateTrip saves t with a new ID, ignoring t.ID, and returns the saved trip.
func (s *Store) CreateTrip(t Trip) Trip {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = s.nextTrip
	s.nextTrip++
	s.trips[t.ID] = t
	return t
}

// UpdateTrip calls change on the trip with this ID and saves the result, all while
// holding the lock, so two updates can't overwrite each other. The ID can't be
// changed. It returns false if there's no such trip.
func (s *Store) UpdateTrip(id int, change func(t *Trip)) (Trip, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.trips[id]
	if !ok {
		return Trip{}, false
	}
	change(&t)
	t.ID = id
	s.trips[id] = t
	return t, true
}

// DeleteTrip removes the trip and its bookings. It returns false if there was no
// such trip.
func (s *Store) DeleteTrip(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.trips[id]; !ok {
		return false
	}
	delete(s.trips, id)
	for bid, b := range s.bookings {
		if b.TripID == id {
			delete(s.bookings, bid)
		}
	}
	return true
}

// Book saves b with a new ID. It returns ErrNoTrip if the trip doesn't exist and
// ErrSeatTaken if someone already has that seat on that trip.
func (s *Store) Book(b Booking) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.trips[b.TripID]; !ok {
		return Booking{}, ErrNoTrip
	}
	for _, other := range s.bookings {
		if other.TripID == b.TripID && other.Seat == b.Seat {
			return Booking{}, ErrSeatTaken
		}
	}
	b.ID = s.nextBooking
	s.nextBooking++
	s.bookings[b.ID] = b
	return b, nil
}

// Booking returns the booking with this ID, and false if there isn't one.
func (s *Store) Booking(id int) (Booking, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[id]
	return b, ok
}

// BookingCount returns how many bookings there are on a trip.
func (s *Store) BookingCount(tripID int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, b := range s.bookings {
		if b.TripID == tripID {
			n++
		}
	}
	return n
}

type errorBody struct {
	Error string `json:"error"`
}

// writeJSON sends v as JSON with the given status. Headers first, then the status,
// then the body: headers set after WriteHeader are ignored.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writing response: %v", err)
	}
}

// writeError sends {"error": message} with the given status.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Error: message})
}
