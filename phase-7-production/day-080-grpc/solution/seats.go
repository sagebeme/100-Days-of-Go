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
	"fmt"
	"slices"
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
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.trips[req.GetTripId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no trip %q", req.GetTripId())
	}
	taken := make([]int32, 0, len(t.taken))
	for seat := range t.taken {
		taken = append(taken, seat)
	}
	slices.Sort(taken)
	return &seatpb.Trip{Id: t.id, Route: t.route, Departs: t.departs, Seats: t.seats, Taken: taken}, nil
}

// BookSeat books a seat and tells every watcher of the trip. Booking IDs are "B1", "B2"
// and so on. Errors: codes.InvalidArgument for an empty trip_id or phone or a seat
// outside 1..seats, codes.NotFound for an unknown trip, codes.AlreadyExists for a seat
// someone already has.
func (s *Server) BookSeat(ctx context.Context, req *seatpb.BookSeatRequest) (*seatpb.Booking, error) {
	if req.GetTripId() == "" || req.GetPhone() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id and phone are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.trips[req.GetTripId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no trip %q", req.GetTripId())
	}
	seat := req.GetSeat()
	if seat < 1 || seat > t.seats {
		return nil, status.Errorf(codes.InvalidArgument, "seat %d is not between 1 and %d", seat, t.seats)
	}
	if _, taken := t.taken[seat]; taken {
		return nil, status.Errorf(codes.AlreadyExists, "seat %d is taken", seat)
	}
	t.taken[seat] = req.GetPhone()
	s.bookings++

	update := &seatpb.SeatUpdate{TripId: t.id, Seat: seat, SeatsFree: t.seats - int32(len(t.taken))}
	for ch := range s.watchers[t.id] {
		select {
		case ch <- update:
		default: // a watcher that has fallen this far behind misses the update
		}
	}
	return &seatpb.Booking{Id: fmt.Sprintf("B%d", s.bookings), TripId: t.id, Seat: seat, Phone: req.GetPhone()}, nil
}

// WatchSeats streams seat updates for one trip. It first sends the current seats free
// with Seat 0, then one update per booking, until the client cancels (return nil) or a
// Send fails (return that error). An unknown trip is codes.NotFound.
func (s *Server) WatchSeats(req *seatpb.WatchSeatsRequest, stream grpc.ServerStreamingServer[seatpb.SeatUpdate]) error {
	ch := make(chan *seatpb.SeatUpdate, 16)
	s.mu.Lock()
	t, ok := s.trips[req.GetTripId()]
	if !ok {
		s.mu.Unlock()
		return status.Errorf(codes.NotFound, "no trip %q", req.GetTripId())
	}
	first := &seatpb.SeatUpdate{TripId: t.id, SeatsFree: t.seats - int32(len(t.taken))}
	if s.watchers[t.id] == nil {
		s.watchers[t.id] = map[chan *seatpb.SeatUpdate]bool{}
	}
	s.watchers[t.id][ch] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.watchers[t.id], ch)
		s.mu.Unlock()
	}()

	if err := stream.Send(first); err != nil {
		return err
	}
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case u := <-ch:
			if err := stream.Send(u); err != nil {
				return err
			}
		}
	}
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
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	t, err := c.rpc.GetTrip(ctx, &seatpb.GetTripRequest{TripId: tripID})
	if err != nil {
		return 0, translate(err)
	}
	return int(t.GetSeats()) - len(t.GetTaken()), nil
}

// Book books a seat and returns the booking ID. It returns ErrSeatTaken for a taken
// seat and ErrNoTrip for an unknown trip; other errors come back as they are.
func (c *Client) Book(ctx context.Context, tripID string, seat int, phone string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	b, err := c.rpc.BookSeat(ctx, &seatpb.BookSeatRequest{TripId: tripID, Seat: int32(seat), Phone: phone})
	if err != nil {
		return "", translate(err)
	}
	return b.GetId(), nil
}

// translate turns well-known gRPC status codes into this package's errors.
func translate(err error) error {
	switch status.Code(err) {
	case codes.AlreadyExists:
		return ErrSeatTaken
	case codes.NotFound:
		return ErrNoTrip
	}
	return err
}
