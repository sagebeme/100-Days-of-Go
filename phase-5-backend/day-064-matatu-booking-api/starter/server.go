package matatubookingapi

import (
	"log/slog"
	"net/http"
)

// Server is the booking API. Add the fields you need: a rate limiter, a draining
// flag for readiness, and so on.
type Server struct {
	cfg   Config
	store Store
	payer Payer
	log   *slog.Logger
}

// NewServer returns the API. A nil cfg.Now means time.Now.
func NewServer(cfg Config, store Store, payer Payer, log *slog.Logger) *Server {
	// Your code here: default cfg.Now, set up the rate limiter.
	return &Server{cfg: cfg, store: store, payer: payer, log: log}
}

// SetDraining makes /readyz answer 503, for graceful shutdown.
func (s *Server) SetDraining() {
	// Your code here
}

// Handler returns every route in the README wrapped in the middleware, outermost
// first: access log, panic recovery, security headers, CORS, rate limit,
// authentication.
func (s *Server) Handler() http.Handler {
	// Your code here
	return http.NewServeMux()
}
