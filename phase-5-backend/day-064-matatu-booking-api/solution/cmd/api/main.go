// Command api runs the matatu booking API.
//
//	JWT_SECRET=$(openssl rand -base64 32) ADMIN_PHONES=0712345678 \
//	  go run ./phase-5-backend/day-064-matatu-booking-api/solution/cmd/api
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	api "github.com/sagebeme/100-Days-of-Go/phase-5-backend/day-064-matatu-booking-api/solution"
)

func main() {
	cfg, err := api.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config:\n%v\n", err)
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(cfg, log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(cfg api.Config, log *slog.Logger) error {
	store := api.NewMemStore()
	if err := api.SeedRoutes(context.Background(), store); err != nil {
		return err
	}
	srv := api.NewServer(cfg, store, &api.FakePayer{}, log)
	httpSrv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errc := make(chan error, 1)
	go func() {
		log.Info("listening", "port", cfg.Port)
		errc <- httpSrv.ListenAndServe()
	}()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down")
	srv.SetDraining()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("stopped cleanly")
	return nil
}
