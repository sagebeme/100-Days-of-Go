// Package matatubookingapi is the Phase 5 capstone: a matatu booking API with
// passenger accounts, roles, seat booking without double-booking, payments behind an
// interface, and the middleware a public API needs.
package matatubookingapi

import (
	"context"
	"errors"
	"time"
)

// Role is what a user may do.
type Role string

// The roles.
const (
	Passenger Role = "passenger"
	Conductor Role = "conductor"
	Admin     Role = "admin"
)

// Route is a matatu route. Fare is in shillings.
type Route struct {
	ID     int    `json:"id"`
	Number string `json:"number"`
	Name   string `json:"name"` // the operator, e.g. "Citi Hoppa"
	From   string `json:"from"`
	To     string `json:"to"`
	Fare   int    `json:"fare"`
}

// Trip is one departure on a route.
type Trip struct {
	ID      int       `json:"id"`
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

// Booking is one seat on one trip. Fare is copied from the route when it's booked.
type Booking struct {
	ID        int    `json:"id"`
	TripID    int    `json:"tripId"`
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
	Fare      int    `json:"fare"`
	Paid      bool   `json:"paid"`
}

// User is an account. PasswordHash is never sent to a client.
type User struct {
	Phone        string `json:"phone"`
	Name         string `json:"name"`
	Role         Role   `json:"role"`
	PasswordHash string `json:"-"`
}

// Errors a Store returns.
var (
	ErrNotFound    = errors.New("not found")
	ErrConflict    = errors.New("already exists")
	ErrSeatTaken   = errors.New("seat already booked")
	ErrAlreadyPaid = errors.New("booking already paid")
)

// Store is where the API keeps its data. Every method is safe for concurrent use.
type Store interface {
	CreateUser(ctx context.Context, u User) error // ErrConflict if the phone is taken
	User(ctx context.Context, phone string) (User, error)
	SetRole(ctx context.Context, phone string, role Role) (User, error)

	CreateRoute(ctx context.Context, r Route) (Route, error)
	Routes(ctx context.Context) ([]Route, error)
	Route(ctx context.Context, id int) (Route, error)

	CreateTrip(ctx context.Context, t Trip) (Trip, error)
	Trips(ctx context.Context, routeID int) ([]Trip, error) // routeID 0 means every route
	Trip(ctx context.Context, id int) (Trip, error)

	// CreateBooking saves b, or returns ErrSeatTaken if that seat on that trip is
	// already booked. The check and the save happen as one step.
	CreateBooking(ctx context.Context, b Booking) (Booking, error)
	Booking(ctx context.Context, id int) (Booking, error)
	BookingsByPhone(ctx context.Context, phone string) ([]Booking, error)
	TakenSeats(ctx context.Context, tripID int) ([]int, error) // sorted
	MarkPaid(ctx context.Context, id int) (Booking, error)     // ErrAlreadyPaid if it was

	Ping(ctx context.Context) error
}

// Payer charges a passenger, for example with an M-Pesa STK push. ref is the
// booking reference the passenger sees on their phone.
type Payer interface {
	Charge(ctx context.Context, phone string, amount int, ref string) error
}

// FakePayer accepts every charge, or fails every charge when Fail is set. Use it in
// tests and in development; a real one would call Safaricom's Daraja API.
type FakePayer struct {
	Fail bool
}

// ErrPaymentFailed is what FakePayer returns when Fail is set.
var ErrPaymentFailed = errors.New("payment failed")

// Charge pretends to charge the passenger.
func (p *FakePayer) Charge(ctx context.Context, phone string, amount int, ref string) error {
	if p.Fail {
		return ErrPaymentFailed
	}
	return nil
}
