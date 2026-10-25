package postgres

// This file is already written for you, and it's the same in starter/ and solution/.
// It holds the domain types and the Store interface. The handlers depend on the
// interface, so the tests can use a fake and production can use Postgres.

import (
	"context"
	"errors"
	"time"
)

// Route is a matatu route, such as 125 from Railways to Ongata Rongai.
type Route struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

// Trip is one departure on a route.
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

// Errors every Store returns, whatever database is behind it.
var (
	ErrNotFound  = errors.New("not found")
	ErrSeatTaken = errors.New("seat already booked")
)

// Store is everything the API needs from storage. Every method takes a context, so
// a slow query stops when the request's deadline passes.
type Store interface {
	CreateRoute(ctx context.Context, r Route) (Route, error)
	CreateTrip(ctx context.Context, t Trip) (Trip, error)
	Trip(ctx context.Context, id int) (Trip, error)
	TripsOnRoute(ctx context.Context, routeID int) ([]Trip, error)
	Book(ctx context.Context, b Booking) (Booking, error)
}
