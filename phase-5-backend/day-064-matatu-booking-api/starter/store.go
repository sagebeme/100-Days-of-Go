package matatubookingapi

import (
	"context"
	"sync"
)

type seatKey struct{ trip, seat int }

// MemStore is an in-memory Store. One mutex guards everything, which makes
// "check the seat, then book it" a single step.
type MemStore struct {
	mu       sync.Mutex
	users    map[string]User
	routes   []Route
	trips    []Trip
	bookings map[int]Booking
	seats    map[seatKey]int // (trip, seat) → booking ID
	nextID   int
}

// NewMemStore returns an empty store.
func NewMemStore() *MemStore {
	return &MemStore{
		users:    map[string]User{},
		bookings: map[int]Booking{},
		seats:    map[seatKey]int{},
		nextID:   1,
	}
}

func (s *MemStore) id() int {
	id := s.nextID
	s.nextID++
	return id
}

// CreateUser saves u, or returns ErrConflict if the phone is taken.
func (s *MemStore) CreateUser(_ context.Context, u User) error {
	// Your code here
	return nil
}

// User returns the user with this phone, or ErrNotFound.
func (s *MemStore) User(_ context.Context, phone string) (User, error) {
	// Your code here
	return User{}, nil
}

// SetRole changes a user's role and returns the updated user.
func (s *MemStore) SetRole(_ context.Context, phone string, role Role) (User, error) {
	// Your code here
	return User{}, nil
}

// CreateRoute saves r with a new ID.
func (s *MemStore) CreateRoute(_ context.Context, r Route) (Route, error) {
	// Your code here
	return Route{}, nil
}

// Routes returns every route, in the order they were created.
func (s *MemStore) Routes(_ context.Context) ([]Route, error) {
	// Your code here
	return nil, nil
}

// Route returns one route, or ErrNotFound.
func (s *MemStore) Route(_ context.Context, id int) (Route, error) {
	// Your code here
	return Route{}, nil
}

// CreateTrip saves t with a new ID.
func (s *MemStore) CreateTrip(_ context.Context, t Trip) (Trip, error) {
	// Your code here
	return Trip{}, nil
}

// Trips returns the trips on a route (every route for 0), earliest first.
func (s *MemStore) Trips(_ context.Context, routeID int) ([]Trip, error) {
	// Your code here
	return nil, nil
}

// Trip returns one trip, or ErrNotFound.
func (s *MemStore) Trip(_ context.Context, id int) (Trip, error) {
	// Your code here
	return Trip{}, nil
}

// CreateBooking saves b with a new ID, or returns ErrSeatTaken.
func (s *MemStore) CreateBooking(_ context.Context, b Booking) (Booking, error) {
	// Your code here
	return Booking{}, nil
}

// Booking returns one booking, or ErrNotFound.
func (s *MemStore) Booking(_ context.Context, id int) (Booking, error) {
	// Your code here
	return Booking{}, nil
}

// BookingsByPhone returns a passenger's bookings, oldest first.
func (s *MemStore) BookingsByPhone(_ context.Context, phone string) ([]Booking, error) {
	// Your code here
	return nil, nil
}

// TakenSeats returns the booked seat numbers on a trip, sorted.
func (s *MemStore) TakenSeats(_ context.Context, tripID int) ([]int, error) {
	// Your code here
	return nil, nil
}

// MarkPaid marks a booking paid, or returns ErrAlreadyPaid.
func (s *MemStore) MarkPaid(_ context.Context, id int) (Booking, error) {
	// Your code here
	return Booking{}, nil
}

// Ping always succeeds: memory is always there.
func (s *MemStore) Ping(context.Context) error { return nil }

// SeedRoutes adds the five routes in the README (checked on nishukishe.com in
// October 2026) to s. Routes change, so confirm with the conductor.
func SeedRoutes(ctx context.Context, s Store) error {
	for _, r := range []Route{
		{Number: "46K", Name: "Citi Hoppa", From: "Kencom / Ambassadeur", To: "Kawangware", Fare: 100},
		{Number: "111", Name: "MCA Sacco", From: "Railways", To: "Ngong", Fare: 100},
		{Number: "125", Name: "City Shuttle", From: "Railways", To: "Ongata Rongai", Fare: 100},
		{Number: "45", Name: "Zamzam", From: "Ronald Ngala Street", To: "Githurai", Fare: 100},
		{Number: "58", Name: "Buruburu 58 Travellers Sacco", From: "Ambassadeur", To: "Buruburu", Fare: 150},
	} {
		if _, err := s.CreateRoute(ctx, r); err != nil {
			return err
		}
	}
	return nil
}
