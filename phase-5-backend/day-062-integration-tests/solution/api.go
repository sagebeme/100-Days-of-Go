// Package integrationtests is a small Twende booking API and the helpers that test it
// end to end: a real HTTP server, fixtures, and a store that is in memory by default
// and Postgres when DATABASE_URL is set.
package integrationtests

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Errors a Store returns.
var (
	ErrNotFound  = errors.New("not found")
	ErrSeatTaken = errors.New("seat already booked")
)

// Trip is one departure on a route.
type Trip struct {
	ID      int       `json:"id"`
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

// Booking is one seat on one trip.
type Booking struct {
	ID        int    `json:"id"`
	TripID    int    `json:"tripId"`
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
}

// Store is where trips and bookings live.
type Store interface {
	AddTrip(ctx context.Context, t Trip) (Trip, error)
	Trip(ctx context.Context, id int) (Trip, error)
	AddBooking(ctx context.Context, b Booking) (Booking, error) // ErrSeatTaken if the seat is booked
	Booking(ctx context.Context, id int) (Booking, error)
	TakenSeats(ctx context.Context, tripID int) (int, error)
}

// MemStore is an in-memory Store, safe for concurrent use.
type MemStore struct {
	mu       sync.Mutex
	trips    map[int]Trip
	bookings map[int]Booking
	nextID   int
}

// NewMemStore returns an empty MemStore.
func NewMemStore() *MemStore {
	return &MemStore{trips: map[int]Trip{}, bookings: map[int]Booking{}, nextID: 1}
}

// AddTrip saves t with a new ID.
func (s *MemStore) AddTrip(_ context.Context, t Trip) (Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = s.nextID
	s.nextID++
	s.trips[t.ID] = t
	return t, nil
}

// Trip returns the trip with this ID, or ErrNotFound.
func (s *MemStore) Trip(_ context.Context, id int) (Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.trips[id]
	if !ok {
		return Trip{}, ErrNotFound
	}
	return t, nil
}

// AddBooking saves b with a new ID, or returns ErrSeatTaken.
func (s *MemStore) AddBooking(_ context.Context, b Booking) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.trips[b.TripID]; !ok {
		return Booking{}, ErrNotFound
	}
	for _, other := range s.bookings {
		if other.TripID == b.TripID && other.Seat == b.Seat {
			return Booking{}, ErrSeatTaken
		}
	}
	b.ID = s.nextID
	s.nextID++
	s.bookings[b.ID] = b
	return b, nil
}

// Booking returns the booking with this ID, or ErrNotFound.
func (s *MemStore) Booking(_ context.Context, id int) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[id]
	if !ok {
		return Booking{}, ErrNotFound
	}
	return b, nil
}

// TakenSeats counts the bookings on a trip.
func (s *MemStore) TakenSeats(_ context.Context, tripID int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, b := range s.bookings {
		if b.TripID == tripID {
			n++
		}
	}
	return n, nil
}

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// Server is the API. Tokens maps a bearer token to the passenger's phone number,
// standing in for Day 58's JWTs.
type Server struct {
	Store  Store
	Tokens map[string]string
}

type phoneKey struct{}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		phone, ok := s.Tokens[tok]
		if tok == "" || !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "log in first")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), phoneKey{}, phone)))
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

type tripView struct {
	Trip
	SeatsLeft int `json:"seatsLeft"`
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getTrip(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "no such trip")
		return
	}
	t, err := s.Store.Trip(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "no such trip")
		return
	}
	if err != nil {
		slog.Error("loading trip", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	taken, err := s.Store.TakenSeats(r.Context(), id)
	if err != nil {
		slog.Error("counting seats", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, tripView{Trip: t, SeatsLeft: t.Seats - taken})
}

type bookingRequest struct {
	TripID    int    `json:"tripId"`
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request) {
	phone, _ := r.Context().Value(phoneKey{}).(string)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req bookingRequest
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body must be valid JSON")
		return
	}
	t, err := s.Store.Trip(r.Context(), req.TripID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "no such trip")
		return
	}
	if err != nil {
		slog.Error("loading trip", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	if req.Seat < 1 || req.Seat > t.Seats {
		writeError(w, http.StatusUnprocessableEntity, "seat must be between 1 and "+strconv.Itoa(t.Seats))
		return
	}
	if strings.TrimSpace(req.Passenger) == "" {
		writeError(w, http.StatusUnprocessableEntity, "passenger name is required")
		return
	}
	b, err := s.Store.AddBooking(r.Context(), Booking{TripID: t.ID, Seat: req.Seat, Passenger: req.Passenger, Phone: phone})
	if errors.Is(err, ErrSeatTaken) {
		writeError(w, http.StatusConflict, "seat "+strconv.Itoa(req.Seat)+" is already booked")
		return
	}
	if err != nil {
		slog.Error("saving booking", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	w.Header().Set("Location", "/bookings/"+strconv.Itoa(b.ID))
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) getBooking(w http.ResponseWriter, r *http.Request) {
	phone, _ := r.Context().Value(phoneKey{}).(string)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "no such booking")
		return
	}
	b, err := s.Store.Booking(r.Context(), id)
	if errors.Is(err, ErrNotFound) || (err == nil && b.Phone != phone) {
		writeError(w, http.StatusNotFound, "no such booking")
		return
	}
	if err != nil {
		slog.Error("loading booking", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// Routes returns the whole API: every route, its middleware, and the security headers.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /trips/{id}", s.getTrip)
	mux.HandleFunc("POST /bookings", s.requireAuth(s.createBooking))
	mux.HandleFunc("GET /bookings/{id}", s.requireAuth(s.getBooking))
	return securityHeaders(mux)
}
