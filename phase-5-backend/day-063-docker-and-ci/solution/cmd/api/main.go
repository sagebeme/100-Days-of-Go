// Command api runs the Twende health endpoints the way the real API will run in a
// container: config checked first, timeouts on the server, and a graceful shutdown
// that drains traffic before it stops.
//
//	go run ./phase-5-backend/day-063-docker-and-ci/solution/cmd/api
//	api -healthcheck   exit 0 if the server on $PORT answers /healthz, else 1
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	dockerandci "github.com/sagebeme/100-Days-of-Go/phase-5-backend/day-063-docker-and-ci/solution"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "call /healthz on the local server and exit 0 or 1")
	flag.Parse()

	cfg, err := dockerandci.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config:\n%v\n", err)
		os.Exit(1)
	}
	if *healthcheck {
		os.Exit(probe(cfg.Port))
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(cfg, log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// probe is for Docker's HEALTHCHECK: a distroless image has no curl or wget.
func probe(port int) int {
	c := &http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/healthz")
	if err != nil {
		return 1
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func run(cfg dockerandci.Config, log *slog.Logger) error {
	health := &dockerandci.Health{Timeout: cfg.ReadyTimeout, Log: log}
	if cfg.DatabaseURL != "" {
		db, err := sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer db.Close()
		health.Checks = append(health.Checks, dockerandci.NewCheck("db", db.PingContext))
	}

	mux := http.NewServeMux()
	health.Register(mux)
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		log.Info("listening", "port", cfg.Port, "env", cfg.Env)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down")
	health.SetDraining()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("stopped cleanly")
	return nil
}
