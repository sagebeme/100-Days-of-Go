// Command api runs the Day 55 booking API on Postgres. Start the database and load
// the schema as the README shows, then from the repo root:
//
//	DATABASE_URL=postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable \
//	  go run ./phase-5-backend/day-055-postgres/solution/cmd/api
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	postgres "github.com/sagebeme/100-Days-of-Go/phase-5-backend/day-055-postgres/solution"
)

func main() {
	// Config, as on Day 21: read it all once, fail fast, never give a secret a default.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "config: DATABASE_URL is required")
		os.Exit(1)
	}
	port := 8080
	if s := os.Getenv("PORT"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 65535 {
			fmt.Fprintf(os.Stderr, "config: PORT must be a whole number from 1 to 65535, got %q\n", s)
			os.Exit(1)
		}
		port = n
	}

	db, err := postgres.OpenDB(dsn)
	if err != nil {
		log.Fatalf("database: %v", err) // a parse error; it never contains the password
	}
	defer db.Close()
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		log.Fatalf("database: can't connect: %v", err)
	}

	api := postgres.NewAPI(postgres.NewPGStore(db), 3*time.Second)
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           api.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	log.Print("stopped cleanly")
}
