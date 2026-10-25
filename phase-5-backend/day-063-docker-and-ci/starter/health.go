package dockerandci

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

// Checker is one dependency the API needs before it can serve traffic: the
// database, a cache, a payment provider.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

type checkFunc struct {
	name string
	fn   func(context.Context) error
}

func (c checkFunc) Name() string                    { return c.name }
func (c checkFunc) Check(ctx context.Context) error { return c.fn(ctx) }

// NewCheck turns a function into a Checker. For a database:
//
//	NewCheck("db", db.PingContext)
func NewCheck(name string, fn func(context.Context) error) Checker {
	// Your code here
	return checkFunc{name: name, fn: func(context.Context) error { return nil }}
}

// Health serves the liveness and readiness endpoints.
type Health struct {
	Checks  []Checker
	Timeout time.Duration // for all checks together
	Log     *slog.Logger

	draining atomic.Bool
}

// SetDraining makes /readyz answer 503 from now on, so the load balancer stops
// sending new requests while the server shuts down.
func (h *Health) SetDraining() {
	// Your code here
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing response", "err", err)
	}
}

// Liveness handles GET /healthz: 200 {"status":"ok"} whenever the process can answer
// at all. It checks no dependencies, so a database outage never makes the platform
// restart a healthy process.
func (h *Health) Liveness(w http.ResponseWriter, r *http.Request) {
	// Your code here
}

// ReadyResponse is the body of GET /readyz.
type ReadyResponse struct {
	Status string            `json:"status"`           // "ready", "not ready" or "draining"
	Checks map[string]string `json:"checks,omitempty"` // check name → "ok" or "failed"
}

// Readiness handles GET /readyz. While draining it answers 503 "draining" without
// running checks. Otherwise it runs every check at the same time, under one context
// that ends after Timeout, and answers 200 "ready" if all pass, or 503 "not ready".
// A failed check's error is logged, never sent to the client.
func (h *Health) Readiness(w http.ResponseWriter, r *http.Request) {
	// Your code here
}

// Register adds GET /healthz and GET /readyz to mux.
func (h *Health) Register(mux *http.ServeMux) {
	// Your code here
}
