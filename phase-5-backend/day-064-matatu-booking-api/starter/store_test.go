package matatubookingapi

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestMemStore(t *testing.T) {
	ctx := t.Context()
	s := NewMemStore()
	if err := s.CreateUser(ctx, User{Phone: "0712345678", Name: "Wanjiku", Role: Passenger}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateUser(ctx, User{Phone: "0712345678", Name: "Copy"}); !errors.Is(err, ErrConflict) {
		t.Errorf("CreateUser(same phone) = %v, want ErrConflict", err)
	}
	if u, err := s.SetRole(ctx, "0712345678", Conductor); err != nil || u.Role != Conductor {
		t.Errorf("SetRole = %+v, %v", u, err)
	}
	if _, err := s.User(ctx, "0799999999"); !errors.Is(err, ErrNotFound) {
		t.Errorf("User(missing) = %v, want ErrNotFound", err)
	}

	if err := SeedRoutes(ctx, s); err != nil {
		t.Fatal(err)
	}
	routes, _ := s.Routes(ctx)
	if len(routes) != 5 || routes[0].Number != "46K" || routes[4].Fare != 150 {
		t.Fatalf("SeedRoutes gave %+v, want the five routes in the README, 46K first and 58 at Ksh150", routes)
	}

	late, _ := s.CreateTrip(ctx, Trip{RouteID: routes[1].ID, Departs: time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC), Seats: 14})
	early, _ := s.CreateTrip(ctx, Trip{RouteID: routes[1].ID, Departs: time.Date(2026, 3, 2, 6, 0, 0, 0, time.UTC), Seats: 14})
	s.CreateTrip(ctx, Trip{RouteID: routes[2].ID, Departs: time.Date(2026, 3, 2, 7, 0, 0, 0, time.UTC), Seats: 14})
	trips, _ := s.Trips(ctx, routes[1].ID)
	if len(trips) != 2 || trips[0].ID != early.ID || trips[1].ID != late.ID {
		t.Errorf("Trips(route) = %+v, want the two trips earliest first", trips)
	}
	if all, _ := s.Trips(ctx, 0); len(all) != 3 {
		t.Errorf("Trips(0) returned %d trips, want all 3", len(all))
	}

	b, err := s.CreateBooking(ctx, Booking{TripID: early.ID, Seat: 5, Phone: "0712345678", Fare: 100})
	if err != nil || b.ID == 0 {
		t.Fatalf("CreateBooking = %+v, %v", b, err)
	}
	if _, err := s.CreateBooking(ctx, Booking{TripID: early.ID, Seat: 5, Phone: "0722000111"}); !errors.Is(err, ErrSeatTaken) {
		t.Errorf("CreateBooking(same seat) = %v, want ErrSeatTaken", err)
	}
	if _, err := s.CreateBooking(ctx, Booking{TripID: late.ID, Seat: 5, Phone: "0722000111"}); err != nil {
		t.Errorf("the same seat number on another trip: %v, want nil", err)
	}
	s.CreateBooking(ctx, Booking{TripID: early.ID, Seat: 2, Phone: "0722000111"})
	if seats, _ := s.TakenSeats(ctx, early.ID); !reflect.DeepEqual(seats, []int{2, 5}) {
		t.Errorf("TakenSeats = %v, want [2 5]", seats)
	}
	if mine, _ := s.BookingsByPhone(ctx, "0712345678"); len(mine) != 1 || mine[0].ID != b.ID {
		t.Errorf("BookingsByPhone = %+v", mine)
	}
	if p, err := s.MarkPaid(ctx, b.ID); err != nil || !p.Paid {
		t.Errorf("MarkPaid = %+v, %v", p, err)
	}
	if _, err := s.MarkPaid(ctx, b.ID); !errors.Is(err, ErrAlreadyPaid) {
		t.Errorf("MarkPaid twice = %v, want ErrAlreadyPaid", err)
	}
}

func TestMemStoreNoDoubleBooking(t *testing.T) {
	s := NewMemStore()
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.CreateBooking(t.Context(), Booking{TripID: 1, Seat: 14}); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Errorf("50 goroutines booked seat 14 and %d succeeded, want exactly 1", wins)
	}
}
