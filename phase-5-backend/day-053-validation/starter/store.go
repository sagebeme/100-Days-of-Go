package validation

// This file is already written for you, and it's the same in starter/ and solution/.
// Today's work is in validate.go and api.go.

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
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

// Errors the store returns.
var (
	ErrNotFound  = errors.New("not found")
	ErrSeatTaken = errors.New("seat already booked")
)

// MemStore keeps trips and bookings in memory. It is safe to use from many goroutines.
type MemStore struct {
	mu       sync.Mutex
	trips    map[int]Trip
	bookings []Booking
}

// NewMemStore returns a store holding these trips.
func NewMemStore(trips ...Trip) *MemStore {
	s := &MemStore{trips: map[int]Trip{}}
	for _, t := range trips {
		s.trips[t.ID] = t
	}
	return s
}

// Trip returns the trip, or ErrNotFound.
func (s *MemStore) Trip(id int) (Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.trips[id]
	if !ok {
		return Trip{}, ErrNotFound
	}
	return t, nil
}

// Book saves b with the next ID, or returns ErrNotFound or ErrSeatTaken.
func (s *MemStore) Book(b Booking) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.trips[b.TripID]; !ok {
		return Booking{}, ErrNotFound
	}
	for _, other := range s.bookings {
		if other.TripID == b.TripID && other.Seat == b.Seat {
			return Booking{}, ErrSeatTaken
		}
	}
	b.ID = len(s.bookings) + 1
	s.bookings = append(s.bookings, b)
	return b, nil
}

// Bookings returns a copy of every booking.
func (s *MemStore) Bookings() []Booking {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Booking(nil), s.bookings...)
}

type errorBody struct {
	Error string `json:"error"`
}

// writeJSON sends v as JSON with the given status.
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
