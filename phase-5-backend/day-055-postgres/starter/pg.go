// Package postgres keeps the matatu-booking API's routes, trips and bookings in
// Postgres, through database/sql and the pgx driver, with a connection pool, a
// timeout on every query and parameters for every value.
package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver with database/sql
)

// Pool settings shared by OpenDB and PoolConfig.
const (
	MaxConns        = 10
	MaxConnLifetime = 30 * time.Minute
	MaxConnIdleTime = 5 * time.Minute
)

// OpenDB returns a *sql.DB for dsn using the pgx driver, with its pool set to at
// most MaxConns open connections, all of which may stay idle, each replaced after
// MaxConnLifetime and closed after MaxConnIdleTime unused. It doesn't connect, and
// the pgx driver only reads dsn when it opens the first connection, so call
// db.PingContext to find out whether dsn and the database are good.
//
//	OpenDB("postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable")
func OpenDB(dsn string) (*sql.DB, error) {
	// Your code here. Follow the steps in the README.
	return nil, nil
}

// PoolConfig parses dsn into a pgxpool configuration with the same limits as
// OpenDB, plus MinConns of 2 kept open and ready. It doesn't connect. Pass it to
// pgxpool.NewWithConfig when you use pgx directly instead of database/sql.
func PoolConfig(dsn string) (*pgxpool.Config, error) {
	// Your code here.
	return nil, nil
}

// TranslateError turns database errors into the Store's errors, so handlers never
// see driver types:
//
//	sql.ErrNoRows                       ErrNotFound
//	*pgconn.PgError with code 23503     ErrNotFound (a trip or route that isn't there)
//	*pgconn.PgError with code 23505     ErrSeatTaken (the UNIQUE (trip_id, seat) rule)
//	nil                                 nil
//	anything else                       err, unchanged
func TranslateError(err error) error {
	// Your code here.
	return nil
}

// PGStore is a Store on Postgres.
type PGStore struct {
	db *sql.DB
}

// NewPGStore returns a Store that uses db.
func NewPGStore(db *sql.DB) *PGStore {
	return &PGStore{db: db}
}

// CreateRoute saves r and returns it with its new ID.
func (s *PGStore) CreateRoute(ctx context.Context, r Route) (Route, error) {
	// Your code here.
	return Route{}, nil
}

// CreateTrip saves t and returns it with its new ID. ErrNotFound means the route
// doesn't exist.
func (s *PGStore) CreateTrip(ctx context.Context, t Trip) (Trip, error) {
	// Your code here.
	return Trip{}, nil
}

// Trip returns the trip with this ID, or ErrNotFound.
func (s *PGStore) Trip(ctx context.Context, id int) (Trip, error) {
	// Your code here.
	return Trip{}, nil
}

// TripsOnRoute returns the route's trips, earliest first. A route with no trips, or
// no such route, gives an empty slice, not nil.
func (s *PGStore) TripsOnRoute(ctx context.Context, routeID int) ([]Trip, error) {
	// Your code here.
	return nil, nil
}

// Book saves b and returns it with its new ID. ErrSeatTaken means someone has the
// seat already, and ErrNotFound means the trip doesn't exist. The database's UNIQUE
// rule decides, so two passengers racing for one seat can't both win.
func (s *PGStore) Book(ctx context.Context, b Booking) (Booking, error) {
	// Your code here.
	return Booking{}, nil
}
