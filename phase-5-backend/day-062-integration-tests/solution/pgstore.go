package integrationtests

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver for database/sql
)

// Schema creates the tables PGStore needs. The UNIQUE constraint is what stops two
// bookings of one seat, even when they arrive at the same moment.
const Schema = `
CREATE TABLE IF NOT EXISTS trips (
	id       BIGSERIAL PRIMARY KEY,
	route_id INT NOT NULL,
	departs  TIMESTAMPTZ NOT NULL,
	seats    INT NOT NULL CHECK (seats > 0)
);
CREATE TABLE IF NOT EXISTS bookings (
	id        BIGSERIAL PRIMARY KEY,
	trip_id   BIGINT NOT NULL REFERENCES trips (id),
	seat      INT NOT NULL,
	passenger TEXT NOT NULL,
	phone     TEXT NOT NULL,
	UNIQUE (trip_id, seat)
);`

// PGStore is a Store backed by Postgres through database/sql and pgx.
type PGStore struct {
	DB *sql.DB
}

// AddTrip inserts t and returns it with its new ID.
func (s *PGStore) AddTrip(ctx context.Context, t Trip) (Trip, error) {
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO trips (route_id, departs, seats) VALUES ($1, $2, $3) RETURNING id`,
		t.RouteID, t.Departs, t.Seats).Scan(&t.ID)
	return t, err
}

// Trip loads one trip, or returns ErrNotFound.
func (s *PGStore) Trip(ctx context.Context, id int) (Trip, error) {
	var t Trip
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, route_id, departs, seats FROM trips WHERE id = $1`, id).
		Scan(&t.ID, &t.RouteID, &t.Departs, &t.Seats)
	if errors.Is(err, sql.ErrNoRows) {
		return Trip{}, ErrNotFound
	}
	t.Departs = t.Departs.UTC()
	return t, err
}

// AddBooking inserts b. A unique violation on (trip_id, seat) is ErrSeatTaken, and a
// missing trip is ErrNotFound.
func (s *PGStore) AddBooking(ctx context.Context, b Booking) (Booking, error) {
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO bookings (trip_id, seat, passenger, phone) VALUES ($1, $2, $3, $4) RETURNING id`,
		b.TripID, b.Seat, b.Passenger, b.Phone).Scan(&b.ID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return Booking{}, ErrSeatTaken
		case "23503": // foreign_key_violation
			return Booking{}, ErrNotFound
		}
	}
	return b, err
}

// Booking loads one booking, or returns ErrNotFound.
func (s *PGStore) Booking(ctx context.Context, id int) (Booking, error) {
	var b Booking
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, trip_id, seat, passenger, phone FROM bookings WHERE id = $1`, id).
		Scan(&b.ID, &b.TripID, &b.Seat, &b.Passenger, &b.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		return Booking{}, ErrNotFound
	}
	return b, err
}

// TakenSeats counts the bookings on a trip.
func (s *PGStore) TakenSeats(ctx context.Context, tripID int) (int, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM bookings WHERE trip_id = $1`, tripID).Scan(&n)
	return n, err
}
