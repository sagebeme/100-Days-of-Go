package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

const dsn = "postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable"

func TestOpenDB(t *testing.T) {
	db, err := OpenDB(dsn)
	if err != nil {
		t.Fatalf("OpenDB: %v (it shouldn't connect, so no database is needed)", err)
	}
	if db == nil {
		t.Fatal("OpenDB returned a nil *sql.DB")
	}
	defer db.Close()
	if got := db.Stats().MaxOpenConnections; got != MaxConns {
		t.Errorf("MaxOpenConnections = %d, want %d: call db.SetMaxOpenConns", got, MaxConns)
	}
}

func TestPoolConfig(t *testing.T) {
	cfg, err := PoolConfig(dsn)
	if err != nil {
		t.Fatalf("PoolConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("PoolConfig returned nil")
	}
	if cfg.MaxConns != MaxConns || cfg.MinConns != 2 {
		t.Errorf("MaxConns, MinConns = %d, %d, want %d, 2", cfg.MaxConns, cfg.MinConns, MaxConns)
	}
	if cfg.MaxConnLifetime != MaxConnLifetime || cfg.MaxConnIdleTime != MaxConnIdleTime {
		t.Errorf("MaxConnLifetime, MaxConnIdleTime = %v, %v, want %v, %v", cfg.MaxConnLifetime, cfg.MaxConnIdleTime, MaxConnLifetime, MaxConnIdleTime)
	}
	if cfg.ConnConfig.Database != "matatu" || cfg.ConnConfig.User != "gopher" {
		t.Errorf("parsed database %q user %q, want matatu and gopher", cfg.ConnConfig.Database, cfg.ConnConfig.User)
	}
	if _, err := PoolConfig("postgres://gopher@localhost:notaport/matatu"); err == nil {
		t.Error("PoolConfig with a bad port returned no error")
	}
}

func TestTranslateError(t *testing.T) {
	other := errors.New("connection reset by peer")
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"nil", nil, nil},
		{"no rows", sql.ErrNoRows, ErrNotFound},
		{"wrapped no rows", fmt.Errorf("trip 7: %w", sql.ErrNoRows), ErrNotFound},
		{"unique violation", &pgconn.PgError{Code: "23505", ConstraintName: "bookings_trip_id_seat_key"}, ErrSeatTaken},
		{"foreign key violation", &pgconn.PgError{Code: "23503", ConstraintName: "bookings_trip_id_fkey"}, ErrNotFound},
		{"wrapped unique violation", fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505"}), ErrSeatTaken},
		{"another Postgres error", &pgconn.PgError{Code: "42P01", Message: `relation "trips" does not exist`}, nil},
		{"something else", other, other},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateError(tt.err)
			if tt.name == "another Postgres error" {
				var pg *pgconn.PgError
				if !errors.As(got, &pg) || pg.Code != "42P01" {
					t.Errorf("TranslateError(%v) = %v, want the error passed through unchanged", tt.err, got)
				}
				return
			}
			if got != tt.want {
				t.Errorf("TranslateError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
