// Command twende runs the hardened booking API. Ctrl-C shuts it down gracefully.
//
//	TWENDE_API_KEY=$(openssl rand -hex 32) go run ./phase-7-production/day-084-hardened-service/solution/cmd/twende
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"

	hs "github.com/sagebeme/100-Days-of-Go/phase-7-production/day-084-hardened-service/solution"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := hs.LoadConfig(os.Getenv)
	if err != nil {
		log.Error("bad config", "err", err)
		os.Exit(1)
	}
	log.Info("config loaded", "config", cfg) // safe: the API key logs as [REDACTED]

	store := hs.NewMemStore(
		hs.Trip{ID: "46K-0730", Route: "46K Kencom to Kawangware", Departs: "07:30", Seats: 14},
		hs.Trip{ID: "125-0745", Route: "125 Railways to Ongata Rongai", Departs: "07:45", Seats: 33},
	)
	srv := hs.NewServer(cfg, store, log, prometheus.NewRegistry())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Error("listen", "err", err)
		os.Exit(1)
	}
	log.Info("listening", "addr", ln.Addr().String())
	if err := srv.Run(ctx, ln); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
	log.Info("stopped cleanly")
}
