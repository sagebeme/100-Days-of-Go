package hardenedservice

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Errors a Store returns.
var (
	ErrNoTrip    = errors.New("no such trip")
	ErrSeatTaken = errors.New("seat already taken")
	ErrBadSeat   = errors.New("no such seat")
)

// Trip is one matatu departure.
type Trip struct {
	ID        string `json:"id"`
	Route     string `json:"route"`
	Departs   string `json:"departs"`
	Seats     int    `json:"seats"`
	SeatsFree int    `json:"seatsFree"`
}

// Booking is one seat sold.
type Booking struct {
	ID     string `json:"id"`
	TripID string `json:"tripId"`
	Seat   int    `json:"seat"`
	Phone  string `json:"phone"`
}

// Store is where trips and bookings live. In production it's Postgres; every method
// must give up when ctx is done.
type Store interface {
	ListTrips(ctx context.Context) ([]Trip, error)
	Book(ctx context.Context, tripID string, seat int, phone string) (Booking, error)
	Ping(ctx context.Context) error
}

// MemStore is an in-memory Store for tests and local runs.
type MemStore struct {
	mu       sync.Mutex
	trips    []Trip
	taken    map[string]map[int]bool
	bookings int
}

// NewMemStore returns a MemStore holding trips.
func NewMemStore(trips ...Trip) *MemStore {
	s := &MemStore{taken: map[string]map[int]bool{}}
	for _, t := range trips {
		t.SeatsFree = t.Seats
		s.trips = append(s.trips, t)
		s.taken[t.ID] = map[int]bool{}
	}
	return s
}

// ListTrips returns every trip.
func (s *MemStore) ListTrips(ctx context.Context) ([]Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Trip(nil), s.trips...), nil
}

// Book sells one seat. It returns ErrNoTrip, ErrSeatTaken, or ErrBadSeat (wrapped,
// with the valid range) for a seat outside 1..Seats.
func (s *MemStore) Book(ctx context.Context, tripID string, seat int, phone string) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.trips {
		if t.ID != tripID {
			continue
		}
		if seat < 1 || seat > t.Seats {
			return Booking{}, fmt.Errorf("%w: seat %d is not between 1 and %d", ErrBadSeat, seat, t.Seats)
		}
		if s.taken[tripID][seat] {
			return Booking{}, ErrSeatTaken
		}
		s.taken[tripID][seat] = true
		s.trips[i].SeatsFree--
		s.bookings++
		return Booking{ID: fmt.Sprintf("B%d", s.bookings), TripID: tripID, Seat: seat, Phone: phone}, nil
	}
	return Booking{}, ErrNoTrip
}

// Ping reports whether the store is reachable. A MemStore always is.
func (s *MemStore) Ping(ctx context.Context) error { return nil }
