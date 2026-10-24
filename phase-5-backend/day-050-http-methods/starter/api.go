// Package httpmethods gives the matatu-booking API one route per method: GET to
// read trips, POST to create them, PUT to replace, PATCH to change, DELETE to
// remove, and POST with an Idempotency-Key to book a seat exactly once.
package httpmethods

import (
	"net/http"
	"sync"
)

// API serves trips and bookings from a Store.
type API struct {
	store *Store

	mu   sync.Mutex               // held while a keyed booking runs, so a retry waits for the first try
	seen map[string]savedResponse // Idempotency-Key → the answer we gave
}

// savedResponse is a successful answer to a booking, kept so a retry with the same
// Idempotency-Key gets exactly the same answer.
type savedResponse struct {
	request  string // trip ID and body of the first request
	status   int
	location string
	body     []byte
}

// NewAPI returns an API that keeps its data in s.
func NewAPI(s *Store) *API {
	return &API{store: s, seen: map[string]savedResponse{}}
}

// Routes returns the API's handler:
//
//	GET    /trips                200 every trip, [] when there are none
//	POST   /trips                201 + Location: /trips/{id}
//	GET    /trips/{id}           200 the trip, 404 if there's none
//	PUT    /trips/{id}           200 the trip, replaced whole
//	PATCH  /trips/{id}           200 the trip, with only the sent fields changed
//	DELETE /trips/{id}           204, no body
//	POST   /trips/{id}/bookings  201 + Location: /bookings/{id}, honours Idempotency-Key
//	GET    /bookings/{id}        200 the booking
//
// GET routes answer HEAD too. A known path with the wrong method gets 405 with an
// Allow header. Errors are {"error": "..."}.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	// Your code here: register one handler method per route. Follow the README.
	return mux
}
