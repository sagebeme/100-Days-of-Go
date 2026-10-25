// Package postgres keeps the matatu-booking API's routes, trips and bookings in
// Postgres, through database/sql and the pgx driver, with a connection pool, a
// timeout on every query and parameters for every value.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(MaxConns)
	db.SetMaxIdleConns(MaxConns)
	db.SetConnMaxLifetime(MaxConnLifetime)
	db.SetConnMaxIdleTime(MaxConnIdleTime)
	return db, nil
}

// PoolConfig parses dsn into a pgxpool configuration with the same limits as
// OpenDB, plus MinConns of 2 kept open and ready. It doesn't connect. Pass it to
// pgxpool.NewWithConfig when you use pgx directly instead of database/sql.
func PoolConfig(dsn string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = MaxConns
	cfg.MinConns = 2
	cfg.MaxConnLifetime = MaxConnLifetime
	cfg.MaxConnIdleTime = MaxConnIdleTime
	return cfg, nil
}

// Postgres error codes. The full list is in the Postgres manual, appendix A.
const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

// TranslateError turns database errors into the Store's errors, so handlers never
// see driver types:
//
//	sql.ErrNoRows                       ErrNotFound
//	*pgconn.PgError with code 23503     ErrNotFound (a trip or route that isn't there)
//	*pgconn.PgError with code 23505     ErrSeatTaken (the UNIQUE (trip_id, seat) rule)
//	nil                                 nil
//	anything else                       err, unchanged
func TranslateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case foreignKeyViolation:
			return ErrNotFound
		case uniqueViolation:
			return ErrSeatTaken
		}
	}
	return err
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
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO routes (name, from_stage, to_stage) VALUES ($1, $2, $3) RETURNING id`,
		r.Name, r.From, r.To,
	).Scan(&r.ID)
	return r, TranslateError(err)
}

// CreateTrip saves t and returns it with its new ID. ErrNotFound means the route
// doesn't exist.
func (s *PGStore) CreateTrip(ctx context.Context, t Trip) (Trip, error) {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO trips (route_id, departs, seats) VALUES ($1, $2, $3) RETURNING id`,
		t.RouteID, t.Departs, t.Seats,
	).Scan(&t.ID)
	if err != nil {
		return Trip{}, TranslateError(err)
	}
	t.Departs = t.Departs.UTC()
	return t, nil
}

// Trip returns the trip with this ID, or ErrNotFound.
func (s *PGStore) Trip(ctx context.Context, id int) (Trip, error) {
	var t Trip
	err := s.db.QueryRowContext(ctx,
		`SELECT id, route_id, departs, seats FROM trips WHERE id = $1`, id,
	).Scan(&t.ID, &t.RouteID, &t.Departs, &t.Seats)
	if err != nil {
		return Trip{}, TranslateError(err)
	}
	t.Departs = t.Departs.UTC()
	return t, nil
}

// TripsOnRoute returns the route's trips, earliest first. A route with no trips, or
// no such route, gives an empty slice, not nil.
func (s *PGStore) TripsOnRoute(ctx context.Context, routeID int) ([]Trip, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, route_id, departs, seats FROM trips WHERE route_id = $1 ORDER BY departs, id`, routeID,
	)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close() // gives the connection back to the pool

	trips := []Trip{}
	for rows.Next() {
		var t Trip
		if err := rows.Scan(&t.ID, &t.RouteID, &t.Departs, &t.Seats); err != nil {
			return nil, err
		}
		t.Departs = t.Departs.UTC()
		trips = append(trips, t)
	}
	// rows.Next returns false on an error too, such as the timeout firing half way.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return trips, nil
}

// Book saves b and returns it with its new ID. ErrSeatTaken means someone has the
// seat already, and ErrNotFound means the trip doesn't exist. The database's UNIQUE
// rule decides, so two passengers racing for one seat can't both win.
func (s *PGStore) Book(ctx context.Context, b Booking) (Booking, error) {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO bookings (trip_id, seat, passenger, phone) VALUES ($1, $2, $3, $4) RETURNING id`,
		b.TripID, b.Seat, b.Passenger, b.Phone,
	).Scan(&b.ID)
	if err != nil {
		return Booking{}, TranslateError(err)
	}
	return b, nil
}
