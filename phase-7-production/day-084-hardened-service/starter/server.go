// Package hardenedservice is a small Twende booking API. It works, and it isn't ready
// for production. Your design doc decides what to change; the tests check the result.
package hardenedservice

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// Server is the booking API.
type Server struct {
	cfg   Config
	store Store
	log   *slog.Logger
	reg   *prometheus.Registry
}

// NewServer returns a Server. Metrics, when you add them, go on reg and are served
// from it.
func NewServer(cfg Config, store Store, log *slog.Logger, reg *prometheus.Registry) *Server {
	return &Server{cfg: cfg, store: store, log: log, reg: reg}
}

// HTTPServer returns the *http.Server to run Handler with.
func (s *Server) HTTPServer() *http.Server {
	return &http.Server{Addr: s.cfg.Addr, Handler: s.Handler()}
}

// Handler returns every route.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /trips", s.listTrips)
	mux.HandleFunc("POST /bookings", s.createBooking)
	return mux
}

// Run serves on ln until ctx is cancelled, then stops.
func (s *Server) Run(ctx context.Context, ln net.Listener) error {
	srv := s.HTTPServer()
	go srv.Serve(ln)
	<-ctx.Done()
	return srv.Close()
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) listTrips(w http.ResponseWriter, r *http.Request) {
	trips, err := s.store.ListTrips(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, trips)
}

type bookingRequest struct {
	TripID string `json:"tripId"`
	Seat   int    `json:"seat"`
	Phone  string `json:"phone"`
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request) {
	var req bookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	b, err := s.store.Book(r.Context(), req.TripID, req.Seat, req.Phone)
	switch {
	case errors.Is(err, ErrNoTrip):
		writeError(w, http.StatusNotFound, "no such trip")
	case errors.Is(err, ErrSeatTaken):
		writeError(w, http.StatusConflict, "seat already taken")
	case err != nil:
		writeError(w, http.StatusInternalServerError, err.Error())
	default:
		writeJSON(w, http.StatusCreated, b)
	}
}
