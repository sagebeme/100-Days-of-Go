package integrationtests

import (
	"errors"
	"os"
	"testing"
	"time"
)

// testStore checks the Store contract. The same tests run on MemStore and on
// PGStore, so the fake can't quietly drift from the real thing.
func testStore(t *testing.T, s Store) {
	t.Helper()
	ctx := t.Context()
	departs := time.Date(2026, 3, 2, 6, 30, 0, 0, time.UTC)
	trip, err := s.AddTrip(ctx, Trip{RouteID: 46, Departs: departs, Seats: 14})
	if err != nil || trip.ID == 0 {
		t.Fatalf("AddTrip = %+v, %v; want an ID", trip, err)
	}
	got, err := s.Trip(ctx, trip.ID)
	if err != nil || got.ID != trip.ID || !got.Departs.Equal(departs) || got.Seats != 14 {
		t.Errorf("Trip(%d) = %+v, %v; want %+v", trip.ID, got, err, trip)
	}
	if _, err := s.Trip(ctx, 424242); !errors.Is(err, ErrNotFound) {
		t.Errorf("Trip(missing) error = %v, want ErrNotFound", err)
	}

	b, err := s.AddBooking(ctx, Booking{TripID: trip.ID, Seat: 4, Passenger: "Akinyi", Phone: "0700111222"})
	if err != nil || b.ID == 0 {
		t.Fatalf("AddBooking = %+v, %v", b, err)
	}
	if _, err := s.AddBooking(ctx, Booking{TripID: trip.ID, Seat: 4, Passenger: "Baraka", Phone: "0700333444"}); !errors.Is(err, ErrSeatTaken) {
		t.Errorf("AddBooking(same seat) error = %v, want ErrSeatTaken", err)
	}
	if _, err := s.AddBooking(ctx, Booking{TripID: 424242, Seat: 1, Passenger: "C", Phone: "0700"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("AddBooking(missing trip) error = %v, want ErrNotFound", err)
	}
	if gb, err := s.Booking(ctx, b.ID); err != nil || gb != b {
		t.Errorf("Booking(%d) = %+v, %v; want %+v", b.ID, gb, err, b)
	}
	if _, err := s.Booking(ctx, 424242); !errors.Is(err, ErrNotFound) {
		t.Errorf("Booking(missing) error = %v, want ErrNotFound", err)
	}
	if n, err := s.TakenSeats(ctx, trip.ID); err != nil || n != 1 {
		t.Errorf("TakenSeats = %d, %v; want 1", n, err)
	}
}

func TestMemStore(t *testing.T) {
	testStore(t, NewMemStore())
}

// TestPostgresStore runs only when DATABASE_URL points at a Postgres you can write
// to. See the README for a Docker one-liner.
func TestPostgresStore(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; skipping the Postgres tests")
	}
	s := OpenTestStore(t)
	if _, ok := s.(*PGStore); !ok {
		t.Fatalf("OpenTestStore with DATABASE_URL set returned %T, want *PGStore", s)
	}
	testStore(t, s)
}

func TestOpenTestStoreDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, ok := OpenTestStore(t).(*MemStore); !ok {
		t.Errorf("OpenTestStore without DATABASE_URL should return a *MemStore")
	}
}
