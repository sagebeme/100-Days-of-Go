package migrations

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// TestRunnerIntegration runs the Runner against a real Postgres. It is skipped
// unless DATABASE_URL is set. It works in its own schema, day56_test, which it drops
// and creates again. See the Day 55 README for a docker one-liner.
func TestRunnerIntegration(t *testing.T) {
	base := os.Getenv("DATABASE_URL")
	if base == "" {
		t.Skip("DATABASE_URL isn't set; see the README to run this against Postgres in Docker")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	for _, q := range []string{`DROP SCHEMA IF EXISTS day56_test CASCADE`, `CREATE SCHEMA day56_test`} {
		if _, err := admin.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	t.Cleanup(func() { admin.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS day56_test CASCADE`) })

	db, err := sql.Open("pgx", withSearchPath(t, base, "day56_test"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	fsys := fstest.MapFS{
		"m/0001_create_routes.up.sql":   file("CREATE TABLE routes (id bigserial PRIMARY KEY, name text NOT NULL);"),
		"m/0001_create_routes.down.sql": file("DROP TABLE routes;"),
		// Two statements in one file.
		"m/0002_create_trips.up.sql":   file("CREATE TABLE trips (id bigserial PRIMARY KEY, route_id bigint NOT NULL REFERENCES routes (id));\nCREATE INDEX trips_route_idx ON trips (route_id);"),
		"m/0002_create_trips.down.sql": file("DROP TABLE trips;"),
		"m/0003_add_fare.up.sql":       file("ALTER TABLE routes ADD COLUMN fare integer NOT NULL DEFAULT 100;"),
		"m/0003_add_fare.down.sql":     file("ALTER TABLE routes DROP COLUMN fare;"),
	}
	ms, err := Parse(fsys, "m")
	if err != nil {
		t.Fatal(err)
	}
	r := NewRunner(db, ms)

	got, err := r.Applied(ctx)
	if err != nil || len(got) != 0 {
		t.Fatalf("Applied on a fresh schema = %v, %v, want [] and the table created", got, err)
	}

	done, err := r.Up(ctx)
	if err != nil {
		t.Fatalf("Up: %v", err)
	}
	if v := versions(done); !reflect.DeepEqual(v, []int{1, 2, 3}) {
		t.Errorf("Up applied %v, want [1 2 3]", v)
	}
	if got, _ := r.Applied(ctx); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("Applied = %v, want [1 2 3]", got)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO routes (name, fare) VALUES ('125', 100)`); err != nil {
		t.Errorf("the migrated routes table doesn't work: %v", err)
	}

	done, err = r.Up(ctx)
	if err != nil || len(done) != 0 {
		t.Errorf("Up again = %v, %v, want nothing applied", versions(done), err)
	}

	done, err = r.Down(ctx, 1)
	if err != nil || !reflect.DeepEqual(versions(done), []int{3}) {
		t.Fatalf("Down(1) = %v, %v, want [3]", versions(done), err)
	}
	if got, _ := r.Applied(ctx); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("Applied after Down(1) = %v, want [1 2]", got)
	}
	if _, err := db.ExecContext(ctx, `SELECT fare FROM routes`); err == nil {
		t.Error("the fare column is still there after rolling back 0003")
	}

	// A broken migration: its first statement works, its second fails. The
	// transaction must undo the first, and the version must not be recorded.
	fsys["m/0004_broken.up.sql"] = file("CREATE TABLE half_made (id int);\nSELECT * FROM no_such_table;")
	fsys["m/0004_broken.down.sql"] = file("DROP TABLE half_made;")
	ms, err = Parse(fsys, "m")
	if err != nil {
		t.Fatal(err)
	}
	r = NewRunner(db, ms)
	done, err = r.Up(ctx)
	if err == nil {
		t.Fatal("Up with a broken migration returned no error")
	}
	if !strings.Contains(err.Error(), "0004_broken") {
		t.Errorf("error %q doesn't name the migration that failed", err)
	}
	if v := versions(done); !reflect.DeepEqual(v, []int{3}) {
		t.Errorf("Up applied %v before failing, want [3]", v)
	}
	if got, _ := r.Applied(ctx); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("Applied after the failure = %v, want [1 2 3]", got)
	}
	var exists bool
	db.QueryRowContext(ctx, `SELECT to_regclass('half_made') IS NOT NULL`).Scan(&exists)
	if exists {
		t.Error("the broken migration left table half_made behind: run each migration in a transaction")
	}

	done, err = r.Down(ctx, 10)
	if err != nil || !reflect.DeepEqual(versions(done), []int{3, 2, 1}) {
		t.Errorf("Down(10) = %v, %v, want [3 2 1]", versions(done), err)
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
