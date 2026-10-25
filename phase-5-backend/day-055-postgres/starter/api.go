package postgres

import (
	"net/http"
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
	// Your code here. Bring writeJSON and writeError over from Day 51.
	return mux
}
