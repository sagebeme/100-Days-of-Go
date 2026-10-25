// Command migrate applies or rolls back the booking API's migrations. The SQL files
// in the migrations folder are embedded in the binary, so a deploy needs only the
// binary and DATABASE_URL. From the repo root:
//
//	export DATABASE_URL="postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable"
//	go run ./phase-5-backend/day-056-migrations/starter/cmd/migrate status
//	go run ./phase-5-backend/day-056-migrations/starter/cmd/migrate up
//	go run ./phase-5-backend/day-056-migrations/starter/cmd/migrate down 1
package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	migrations "github.com/sagebeme/100-Days-of-Go/phase-5-backend/day-056-migrations/starter"
)

//go:embed migrations/*.sql
var files embed.FS

const usage = "usage: migrate up | down [steps] | status"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", usage)
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	all, err := migrations.Parse(files, "migrations")
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("pgx", dsn) // the migrations package registers the driver
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("can't connect: %w", err)
	}
	r := migrations.NewRunner(db, all)

	switch args[0] {
	case "up":
		done, err := r.Up(ctx)
		for _, m := range done {
			fmt.Printf("applied     %04d_%s\n", m.Version, m.Name)
		}
		if err == nil && len(done) == 0 {
			fmt.Println("nothing to apply: the database is up to date")
		}
		return err
	case "down":
		steps := 1
		if len(args) > 1 {
			steps, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("steps must be a whole number, got %q", args[1])
			}
		}
		done, err := r.Down(ctx, steps)
		for _, m := range done {
			fmt.Printf("rolled back %04d_%s\n", m.Version, m.Name)
		}
		return err
	case "status":
		versions, err := r.Applied(ctx)
		if err != nil {
			return err
		}
		have := map[int]bool{}
		for _, v := range versions {
			have[v] = true
		}
		for _, m := range all {
			state := "pending"
			if have[m.Version] {
				state = "applied"
			}
			fmt.Printf("%-8s %04d_%s\n", state, m.Version, m.Name)
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q; %s", args[0], usage)
	}
}
