// Package grpc is Twende's seat service over gRPC: a server that implements the
// SeatService contract in ../seatpb/seat.proto, and a small client that wraps the
// generated one with timeouts and friendly errors.
//
// The package is named after its folder. Inside it, the name grpc still refers to the
// imported google.golang.org/grpc package, because an import name and the package's own
// name never clash in Go.
package grpc

import (
	"context"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sagebeme/100-Days-of-Go/phase-7-production/day-080-grpc/seatpb"
)

// trip is the server's own record of a trip: the details and who has which seat.
type trip struct {
	id, route, departs string
	seats              int32
	taken              map[int32]string // seat → phone
}

// Server implements seatpb.SeatServiceServer. Embedding the Unimplemented type makes
// any method you haven't written answer codes.Unimplemented, and keeps the server
// compiling when the contract gains new methods.
type Server struct {
	seatpb.UnimplementedSeatServiceServer

	mu       sync.Mutex
	trips    map[string]*trip
	watchers map[string]map[chan *seatpb.SeatUpdate]bool // trip ID → its watchers
	bookings int
}

// NewServer returns a Server that sells seats on these trips. Only ID, Route, Departs
// and Seats are read from each one.
func NewServer(trips ...*seatpb.Trip) *Server {
	s := &Server{trips: map[string]*trip{}, watchers: map[string]map[chan *seatpb.SeatUpdate]bool{}}
	for _, t := range trips {
		s.trips[t.GetId()] = &trip{id: t.GetId(), route: t.GetRoute(), departs: t.GetDeparts(), seats: t.GetSeats(), taken: map[int32]string{}}
	}
	return s
}

// GetTrip returns the trip with its taken seats, smallest first.
// An empty trip_id is codes.InvalidArgument; an unknown one is codes.NotFound.
func (s *Server) GetTrip(ctx context.Context, req *seatpb.GetTripRequest) (*seatpb.Trip, error) {
	// Your code here
	return nil, status.Error(codes.Unimplemented, "GetTrip is not written yet")
}

// BookSeat books a seat and tells every watcher of the trip. Booking IDs are "B1", "B2"
// and so on. Errors: codes.InvalidArgument for an empty trip_id or phone or a seat
// outside 1..seats, codes.NotFound for an unknown trip, codes.AlreadyExists for a seat
// someone already has.
func (s *Server) BookSeat(ctx context.Context, req *seatpb.BookSeatRequest) (*seatpb.Booking, error) {
	// Your code here
	return nil, status.Error(codes.Unimplemented, "BookSeat is not written yet")
}

// WatchSeats streams seat updates for one trip. It first sends the current seats free
// with Seat 0, then one update per booking, until the client cancels (return nil) or a
// Send fails (return that error). An unknown trip is codes.NotFound.
func (s *Server) WatchSeats(req *seatpb.WatchSeatsRequest, stream grpc.ServerStreamingServer[seatpb.SeatUpdate]) error {
	// Your code here
	return status.Error(codes.Unimplemented, "WatchSeats is not written yet")
}

// ErrSeatTaken and ErrNoTrip are what Client returns instead of raw gRPC statuses.
var (
	ErrSeatTaken = errors.New("seat already taken")
	ErrNoTrip    = errors.New("no such trip")
)

// Client wraps the generated client: every call gets a deadline, and well-known
// status codes come back as plain Go errors.
type Client struct {
	rpc     seatpb.SeatServiceClient
	timeout time.Duration
}

// NewClient returns a Client that gives each unary call timeout to finish.
func NewClient(conn grpc.ClientConnInterface, timeout time.Duration) *Client {
	return &Client{rpc: seatpb.NewSeatServiceClient(conn), timeout: timeout}
}

// SeatsFree returns how many seats are left on a trip, or ErrNoTrip.
func (c *Client) SeatsFree(ctx context.Context, tripID string) (int, error) {
	// Your code here
	return 0, nil
}

// Book books a seat and returns the booking ID. It returns ErrSeatTaken for a taken
// seat and ErrNoTrip for an unknown trip; other errors come back as they are.
func (c *Client) Book(ctx context.Context, tripID string, seat int, phone string) (string, error) {
	// Your code here
	return "", nil
}

// translate turns well-known gRPC status codes into this package's errors.
func translate(err error) error {
	// Your code here
	return err
}
