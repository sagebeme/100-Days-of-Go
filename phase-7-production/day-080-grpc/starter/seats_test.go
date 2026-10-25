package grpc

import (
	"context"
	"errors"
	"net"
	"runtime"
	"slices"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sagebeme/100-Days-of-Go/phase-7-production/day-080-grpc/seatpb"
)

// start runs a real gRPC server on an in-memory listener and returns a connection to
// it. Nothing touches the network.
func start(t *testing.T) (*Server, *grpc.ClientConn) {
	t.Helper()
	srv := NewServer(
		&seatpb.Trip{Id: "46K-0730", Route: "46K Kencom to Kawangware", Departs: "07:30", Seats: 14},
		&seatpb.Trip{Id: "111-0800", Route: "111 Railways to Ngong", Departs: "08:00", Seats: 3},
	)
	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer()
	seatpb.RegisterSeatServiceServer(gs, srv)
	go gs.Serve(lis)
	t.Cleanup(gs.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return srv, conn
}

func ctx(t *testing.T) context.Context {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return c
}

func TestGetTrip(t *testing.T) {
	_, conn := start(t)
	rpc := seatpb.NewSeatServiceClient(conn)
	got, err := rpc.GetTrip(ctx(t), &seatpb.GetTripRequest{TripId: "46K-0730"})
	if err != nil {
		t.Fatalf("GetTrip = %v", err)
	}
	if got.GetId() != "46K-0730" || got.GetRoute() != "46K Kencom to Kawangware" || got.GetDeparts() != "07:30" || got.GetSeats() != 14 || len(got.GetTaken()) != 0 {
		t.Errorf("GetTrip = %v, want the 46K trip with 14 seats and none taken", got)
	}

	tests := []struct {
		id   string
		code codes.Code
	}{
		{"", codes.InvalidArgument},
		{"999-0000", codes.NotFound},
	}
	for _, tt := range tests {
		_, err := rpc.GetTrip(ctx(t), &seatpb.GetTripRequest{TripId: tt.id})
		if got := status.Code(err); got != tt.code {
			t.Errorf("GetTrip(%q) code = %v, want %v", tt.id, got, tt.code)
		}
	}
}

func TestBookSeat(t *testing.T) {
	_, conn := start(t)
	rpc := seatpb.NewSeatServiceClient(conn)
	b, err := rpc.BookSeat(ctx(t), &seatpb.BookSeatRequest{TripId: "46K-0730", Seat: 3, Phone: "0712345678"})
	if err != nil {
		t.Fatalf("BookSeat = %v", err)
	}
	if b.GetId() != "B1" || b.GetTripId() != "46K-0730" || b.GetSeat() != 3 || b.GetPhone() != "0712345678" {
		t.Errorf("BookSeat = %v, want booking B1 for seat 3", b)
	}
	if _, err := rpc.BookSeat(ctx(t), &seatpb.BookSeatRequest{TripId: "46K-0730", Seat: 1, Phone: "0722000111"}); err != nil {
		t.Fatalf("second BookSeat = %v", err)
	}

	tests := []struct {
		name string
		req  *seatpb.BookSeatRequest
		code codes.Code
	}{
		{"taken", &seatpb.BookSeatRequest{TripId: "46K-0730", Seat: 3, Phone: "0733000222"}, codes.AlreadyExists},
		{"seat 0", &seatpb.BookSeatRequest{TripId: "46K-0730", Seat: 0, Phone: "0733000222"}, codes.InvalidArgument},
		{"seat 15", &seatpb.BookSeatRequest{TripId: "46K-0730", Seat: 15, Phone: "0733000222"}, codes.InvalidArgument},
		{"no phone", &seatpb.BookSeatRequest{TripId: "46K-0730", Seat: 5}, codes.InvalidArgument},
		{"no trip id", &seatpb.BookSeatRequest{Seat: 5, Phone: "0733000222"}, codes.InvalidArgument},
		{"unknown trip", &seatpb.BookSeatRequest{TripId: "999-0000", Seat: 5, Phone: "0733000222"}, codes.NotFound},
	}
	for _, tt := range tests {
		_, err := rpc.BookSeat(ctx(t), tt.req)
		if got := status.Code(err); got != tt.code {
			t.Errorf("BookSeat %s: code = %v, want %v (%v)", tt.name, got, tt.code, err)
		}
	}

	trip, err := rpc.GetTrip(ctx(t), &seatpb.GetTripRequest{TripId: "46K-0730"})
	if err != nil || !slices.Equal(trip.GetTaken(), []int32{1, 3}) {
		t.Errorf("Taken = %v, %v, want [1 3], smallest first", trip.GetTaken(), err)
	}
}

func TestWatchSeats(t *testing.T) {
	srv, conn := start(t)
	rpc := seatpb.NewSeatServiceClient(conn)
	rpc.BookSeat(ctx(t), &seatpb.BookSeatRequest{TripId: "111-0800", Seat: 1, Phone: "0712345678"})

	watchCtx, stop := context.WithCancel(ctx(t))
	stream, err := rpc.WatchSeats(watchCtx, &seatpb.WatchSeatsRequest{TripId: "111-0800"})
	if err != nil {
		t.Fatalf("WatchSeats = %v", err)
	}
	// The first message means the server is subscribed, so the bookings below can't be missed.
	first, err := stream.Recv()
	if err != nil || first.GetSeat() != 0 || first.GetSeatsFree() != 2 || first.GetTripId() != "111-0800" {
		t.Fatalf("first update = %v, %v, want seat 0 with 2 free", first, err)
	}

	rpc.BookSeat(ctx(t), &seatpb.BookSeatRequest{TripId: "46K-0730", Seat: 9, Phone: "0700000001"}) // another trip: not sent
	for _, seat := range []int32{2, 3} {
		if _, err := rpc.BookSeat(ctx(t), &seatpb.BookSeatRequest{TripId: "111-0800", Seat: seat, Phone: "0700000002"}); err != nil {
			t.Fatalf("BookSeat(%d) = %v", seat, err)
		}
	}
	for _, want := range []struct{ seat, free int32 }{{2, 1}, {3, 0}} {
		u, err := stream.Recv()
		if err != nil || u.GetSeat() != want.seat || u.GetSeatsFree() != want.free {
			t.Fatalf("update = %v, %v, want seat %d with %d free", u, err, want.seat, want.free)
		}
	}

	stop()
	if _, err := stream.Recv(); status.Code(err) != codes.Canceled {
		t.Errorf("Recv after cancel = %v, want Canceled", err)
	}
	// The server must forget the watcher once the client has gone.
	deadline := time.Now().Add(time.Second)
	for {
		srv.mu.Lock()
		n := len(srv.watchers["111-0800"])
		srv.mu.Unlock()
		if n == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d watchers still registered after the client cancelled", n)
		}
		runtime.Gosched()
	}

	// A streaming call's error arrives on the first Recv.
	missing, err := rpc.WatchSeats(ctx(t), &seatpb.WatchSeatsRequest{TripId: "999-0000"})
	if err == nil {
		_, err = missing.Recv()
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("WatchSeats(unknown) = %v, want NotFound", err)
	}
}

func TestClient(t *testing.T) {
	_, conn := start(t)
	c := NewClient(conn, time.Second)
	free, err := c.SeatsFree(ctx(t), "111-0800")
	if err != nil || free != 3 {
		t.Errorf("SeatsFree = %d, %v, want 3, nil", free, err)
	}
	id, err := c.Book(ctx(t), "111-0800", 2, "0712345678")
	if err != nil || id != "B1" {
		t.Errorf("Book = %q, %v, want \"B1\", nil", id, err)
	}
	if free, _ := c.SeatsFree(ctx(t), "111-0800"); free != 2 {
		t.Errorf("SeatsFree after a booking = %d, want 2", free)
	}
	if _, err := c.Book(ctx(t), "111-0800", 2, "0722000111"); !errors.Is(err, ErrSeatTaken) {
		t.Errorf("Book(taken) = %v, want ErrSeatTaken", err)
	}
	if _, err := c.SeatsFree(ctx(t), "999-0000"); !errors.Is(err, ErrNoTrip) {
		t.Errorf("SeatsFree(unknown) = %v, want ErrNoTrip", err)
	}
	if _, err := c.Book(ctx(t), "111-0800", 99, "0722000111"); status.Code(err) != codes.InvalidArgument {
		t.Errorf("Book(seat 99) = %v, want the InvalidArgument status passed through", err)
	}
}

func TestClientDeadline(t *testing.T) {
	// A server that never answers: the client's timeout must end the call.
	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer()
	seatpb.RegisterSeatServiceServer(gs, hang{})
	go gs.Serve(lis)
	t.Cleanup(gs.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	c := NewClient(conn, 50*time.Millisecond)
	_, err = c.SeatsFree(context.Background(), "46K-0730")
	if status.Code(err) != codes.DeadlineExceeded {
		t.Errorf("SeatsFree against a hung server = %v, want DeadlineExceeded: give every call a timeout", err)
	}
}

// hang is a SeatService whose GetTrip waits until the caller gives up.
type hang struct {
	seatpb.UnimplementedSeatServiceServer
}

func (hang) GetTrip(ctx context.Context, _ *seatpb.GetTripRequest) (*seatpb.Trip, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
