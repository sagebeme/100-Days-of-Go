package migrations

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver with database/sql
)

// lockID is the key of the Postgres advisory lock the runner holds, so two copies of
// the app starting at once don't both apply the same migration. Any fixed number
// works, as long as nothing else in the database uses it.
const lockID = 56_0001

// Runner applies and rolls back migrations on a database.
type Runner struct {
	db         *sql.DB
	migrations []Migration
}

// NewRunner returns a Runner for these migrations, sorted as Parse returns them.
func NewRunner(db *sql.DB, migrations []Migration) *Runner {
	return &Runner{db: db, migrations: migrations}
}

const createTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version    bigint PRIMARY KEY,
	name       text NOT NULL,
	applied_at timestamptz NOT NULL DEFAULT now()
)`

// withLock runs fn on one connection that holds the advisory lock. A session lock
// belongs to a connection, so everything must use that same connection.
func (r *Runner) withLock(ctx context.Context, fn func(conn *sql.Conn) error) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, lockID); err != nil {
		return fmt.Errorf("taking the migration lock: %w", err)
	}
	defer conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, lockID)

	if _, err := conn.ExecContext(ctx, createTable); err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}
	return fn(conn)
}

func applied(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}) ([]int, error) {
	rows, err := q.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	versions := []int{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// Applied returns the versions recorded in schema_migrations, oldest first, creating
// the table if it doesn't exist yet.
func (r *Runner) Applied(ctx context.Context) ([]int, error) {
	var versions []int
	err := r.withLock(ctx, func(conn *sql.Conn) error {
		var err error
		versions, err = applied(ctx, conn)
		return err
	})
	return versions, err
}

// Up applies every pending migration in order, each in its own transaction that
// runs the Up SQL and records the version. It stops at the first failure, which
// leaves that migration and everything after it unapplied. It returns the
// migrations it applied.
func (r *Runner) Up(ctx context.Context) ([]Migration, error) {
	// Your code here. Use r.withLock, applied, Pending and inTx.
	return nil, nil
}

// Down rolls back the newest steps applied migrations, newest first, each in its
// own transaction that runs the Down SQL and deletes the version. It returns the
// migrations it rolled back.
func (r *Runner) Down(ctx context.Context, steps int) ([]Migration, error) {
	// Your code here. Use r.withLock, applied, ToRollBack and inTx.
	return nil, nil
}

// inTx runs fn in a transaction on conn: commit if fn succeeds, roll back if not.
// Postgres can roll back CREATE TABLE and ALTER TABLE too, which is what makes a
// failed migration leave no half-made tables behind.
func inTx(ctx context.Context, conn *sql.Conn, fn func(tx *sql.Tx) error) error {
	// Your code here.
	return nil
}
