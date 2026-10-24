package statuscodes

// This file is already written for you, and it's the same in starter/ and solution/.
// It holds the domain types, the Store and Limiter interfaces the API depends on, and
// an in-memory store. Today's work is in api.go.

import (
	"context"
	"errors"
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

// Errors a Store returns. Wrap them with %w if you add detail.
var (
	ErrNotFound  = errors.New("not found")
	ErrSeatTaken = errors.New("seat already booked")
)

// Store is everything the API needs from storage. Day 55 puts Postgres behind it.
type Store interface {
	Trip(ctx context.Context, id int) (Trip, error)
	Book(ctx context.Context, b Booking) (Booking, error)
	Booking(ctx context.Context, id int) (Booking, error)
	CancelBooking(ctx context.Context, id int) error
}

// Limiter decides whether a client may make another request now. When it says no,
// retryAfter is how long the client should wait.
type Limiter interface {
	Allow(key string) (ok bool, retryAfter time.Duration)
}

// MemStore is a Store in memory, for tests and for running the API locally.
type MemStore struct {
	mu       sync.Mutex
	trips    map[int]Trip
	bookings map[int]Booking
	next     int
}

// NewMemStore returns a store holding these trips and no bookings.
func NewMemStore(trips ...Trip) *MemStore {
	s := &MemStore{trips: map[int]Trip{}, bookings: map[int]Booking{}, next: 1}
	for _, t := range trips {
		s.trips[t.ID] = t
	}
	return s
}

// Trip returns the trip, or ErrNotFound.
func (s *MemStore) Trip(ctx context.Context, id int) (Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.trips[id]
	if !ok {
		return Trip{}, ErrNotFound
	}
	return t, nil
}

// Book saves b with a new ID. It returns ErrNotFound if the trip doesn't exist and
// ErrSeatTaken if the seat is booked already.
func (s *MemStore) Book(ctx context.Context, b Booking) (Booking, error) {
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
	b.ID = s.next
	s.next++
	s.bookings[b.ID] = b
	return b, nil
}

// Booking returns the booking, or ErrNotFound.
func (s *MemStore) Booking(ctx context.Context, id int) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[id]
	if !ok {
		return Booking{}, ErrNotFound
	}
	return b, nil
}

// CancelBooking deletes the booking, or returns ErrNotFound.
func (s *MemStore) CancelBooking(ctx context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bookings[id]; !ok {
		return ErrNotFound
	}
	delete(s.bookings, id)
	return nil
}
