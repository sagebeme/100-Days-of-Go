package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// API serves trips and bookings from any Store.
type API struct {
	store   Store
	timeout time.Duration
}

// NewAPI returns an API on store. Every request's database work must finish within
// timeout, or the client gets a 503.
func NewAPI(store Store, timeout time.Duration) *API {
	return &API{store: store, timeout: timeout}
}

// Routes returns the handler:
//
//	GET  /routes/{id}/trips    200 the route's trips, [] when none
//	GET  /trips/{id}           200 · 404
//	POST /trips/{id}/bookings  201 + Location · 400 · 404 · 409 · 422
//
// A bad {id} is 400. A store call that runs past the timeout is 503 with
// Retry-After: 1. Any other store error is a 500 that logs the real error.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /routes/{id}/trips", a.routeTrips)
	mux.HandleFunc("GET /trips/{id}", a.getTrip)
	mux.HandleFunc("POST /trips/{id}/bookings", a.createBooking)
	return mux
}

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writing response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Error: message})
}

// fail answers for a store error.
func fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrSeatTaken):
		writeError(w, http.StatusConflict, "that seat is already booked")
	case errors.Is(err, context.DeadlineExceeded):
		log.Printf("%s %s: database timeout: %v", r.Method, r.URL.Path, err)
		w.Header().Set("Retry-After", "1")
		writeError(w, http.StatusServiceUnavailable, "the database is busy, try again")
	default:
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
	}
}

func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "id must be a positive whole number")
		return 0, false
	}
	return id, true
}

func (a *API) routeTrips(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
	defer cancel()
	trips, err := a.store.TripsOnRoute(ctx, id)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, trips)
}

func (a *API) getTrip(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
	defer cancel()
	t, err := a.store.Trip(ctx, id)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

type bookingRequest struct {
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
}

func (a *API) createBooking(w http.ResponseWriter, r *http.Request) {
	tripID, ok := pathID(w, r)
	if !ok {
		return
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	dec.DisallowUnknownFields()
	var req bookingRequest
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body must be valid JSON with seat, passenger and phone")
		return
	}
	if strings.TrimSpace(req.Passenger) == "" || strings.TrimSpace(req.Phone) == "" {
		writeError(w, http.StatusUnprocessableEntity, "passenger and phone are required")
		return
	}

	// One deadline for the whole request: both queries share it.
	ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
	defer cancel()
	trip, err := a.store.Trip(ctx, tripID)
	if err != nil {
		fail(w, r, err)
		return
	}
	if req.Seat < 1 || req.Seat > trip.Seats {
		writeError(w, http.StatusUnprocessableEntity, "seat must be between 1 and "+strconv.Itoa(trip.Seats))
		return
	}
	b, err := a.store.Book(ctx, Booking{TripID: tripID, Seat: req.Seat, Passenger: strings.TrimSpace(req.Passenger), Phone: strings.TrimSpace(req.Phone)})
	if err != nil {
		fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/bookings/"+strconv.Itoa(b.ID))
	writeJSON(w, http.StatusCreated, b)
}
