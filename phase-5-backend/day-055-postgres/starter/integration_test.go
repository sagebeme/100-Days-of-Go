package postgres

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// TestPGStoreIntegration runs PGStore against a real Postgres. It is skipped unless
// DATABASE_URL is set. It works in its own schema, day55_test, which it drops and
// creates again, so it never touches your other tables. See the README for a
// docker one-liner.
func TestPGStoreIntegration(t *testing.T) {
	base := os.Getenv("DATABASE_URL")
	if base == "" {
		t.Skip("DATABASE_URL isn't set; see the README to run this against Postgres in Docker")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	for _, q := range []string{`DROP SCHEMA IF EXISTS day55_test CASCADE`, `CREATE SCHEMA day55_test`} {
		if _, err := admin.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	t.Cleanup(func() { admin.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS day55_test CASCADE`) })

	db, err := OpenDB(withSearchPath(t, base, "day55_test"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	schema, err := os.ReadFile("testdata/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range strings.Split(string(schema), ";") {
		if strings.TrimSpace(stripComments(stmt)) == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("schema: %v\n%s", err, stmt)
		}
	}

	s := NewPGStore(db)
	route, err := s.CreateRoute(ctx, Route{Name: "125", From: "Railways", To: "Ongata Rongai"})
	if err != nil || route.ID == 0 {
		t.Fatalf("CreateRoute = %+v, %v", route, err)
	}

	nairobi := time.FixedZone("EAT", 3*60*60)
	late, err := s.CreateTrip(ctx, Trip{RouteID: route.ID, Departs: time.Date(2026, 11, 2, 8, 0, 0, 0, nairobi), Seats: 14})
	if err != nil {
		t.Fatalf("CreateTrip: %v", err)
	}
	early, err := s.CreateTrip(ctx, Trip{RouteID: route.ID, Departs: time.Date(2026, 11, 2, 7, 30, 0, 0, nairobi), Seats: 14})
	if err != nil {
		t.Fatalf("CreateTrip: %v", err)
	}
	if _, err := s.CreateTrip(ctx, Trip{RouteID: 999999, Departs: time.Now(), Seats: 14}); !errors.Is(err, ErrNotFound) {
		t.Errorf("CreateTrip on a missing route: err = %v, want ErrNotFound", err)
	}

	got, err := s.Trip(ctx, early.ID)
	if err != nil || !got.Departs.Equal(early.Departs) || got.Seats != 14 || got.RouteID != route.ID {
		t.Errorf("Trip(%d) = %+v, %v, want %+v", early.ID, got, err, early)
	}
	if _, err := s.Trip(ctx, 999999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Trip(999999): err = %v, want ErrNotFound", err)
	}

	trips, err := s.TripsOnRoute(ctx, route.ID)
	if err != nil || len(trips) != 2 || trips[0].ID != early.ID || trips[1].ID != late.ID {
		t.Errorf("TripsOnRoute = %+v, %v, want the 07:30 then the 08:00", trips, err)
	}
	if none, err := s.TripsOnRoute(ctx, 999999); err != nil || none == nil || len(none) != 0 {
		t.Errorf("TripsOnRoute(no such route) = %#v, %v, want an empty, non-nil slice", none, err)
	}

	b, err := s.Book(ctx, Booking{TripID: early.ID, Seat: 3, Passenger: "Achieng", Phone: "254712345678"})
	if err != nil || b.ID == 0 {
		t.Fatalf("Book = %+v, %v", b, err)
	}
	if _, err := s.Book(ctx, Booking{TripID: early.ID, Seat: 3, Passenger: "Baraka", Phone: "254722000111"}); !errors.Is(err, ErrSeatTaken) {
		t.Errorf("Book the same seat: err = %v, want ErrSeatTaken", err)
	}
	if _, err := s.Book(ctx, Booking{TripID: 999999, Seat: 1, Passenger: "Baraka", Phone: "254722000111"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Book on a missing trip: err = %v, want ErrNotFound", err)
	}

	// A name that would break string-built SQL is just data with parameters.
	odd := `Robert'); DROP TABLE bookings; --`
	if _, err := s.Book(ctx, Booking{TripID: early.ID, Seat: 4, Passenger: odd, Phone: "254700000000"}); err != nil {
		t.Fatalf("Book with an odd name: %v", err)
	}
	var name string
	if err := db.QueryRowContext(ctx, `SELECT passenger FROM bookings WHERE seat = 4`).Scan(&name); err != nil || name != odd {
		t.Errorf("stored passenger = %q, %v, want it saved exactly", name, err)
	}

	short, cancelShort := context.WithDeadline(ctx, time.Now().Add(-time.Second)) // already past
	defer cancelShort()
	if _, err := s.Trip(short, early.ID); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Trip with an expired context: err = %v, want context.DeadlineExceeded", err)
	}
}

// withSearchPath adds search_path to a DSN, in URL or key=value form.
func withSearchPath(t *testing.T, dsn, schema string) string {
	t.Helper()
	if !strings.Contains(dsn, "://") {
		return dsn + " search_path=" + schema
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("DATABASE_URL: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	return u.String()
}

func stripComments(sql string) string {
	var lines []string
	for _, line := range strings.Split(sql, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
