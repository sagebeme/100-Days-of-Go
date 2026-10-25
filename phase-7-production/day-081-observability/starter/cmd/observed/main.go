// Command observed runs a tiny Twende API with logs, metrics and health checks on
// localhost:8080. Try:
//
//	curl -i localhost:8080/trips
//	curl localhost:8080/readyz
//	curl localhost:8080/metrics | grep http_
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	obs "github.com/sagebeme/100-Days-of-Go/phase-7-production/day-081-observability/starter"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	metrics := obs.NewMetrics(prometheus.DefaultRegisterer, time.Now)
	ready := obs.NewReadiness(log, time.Second)
	ready.Add("db", func(ctx context.Context) error { return nil }) // a real service pings Postgres here

	mux := http.NewServeMux()
	mux.Handle("GET /trips", metrics.Instrument("/trips", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"46K-0730","seatsFree":9}]` + "\n"))
	})))
	mux.HandleFunc("GET /healthz", obs.Live)
	mux.Handle("GET /readyz", ready)
	mux.Handle("GET /metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:              "localhost:8080",
		Handler:           obs.RequestID(obs.RandomID, obs.Logging(log, mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
