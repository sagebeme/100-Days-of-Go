// Package statuscodes is the matatu-booking API with the right status code for
// every outcome, and one shape for every error: {"error": "..."}.
package statuscodes

import (
	"net/http"
)

// InvalidError says a request was readable but broke a rule, such as a seat number
// outside the matatu. Its message is safe to show the client.
type InvalidError struct {
	Message string
}

func (e *InvalidError) Error() string {
	// Your code here.
	return ""
}

// StatusFor picks the status code and the client-safe message for err:
//
//	nil                        200, ""
//	*InvalidError (wrapped ok) 422, its Message
//	ErrNotFound                404, "not found"
//	ErrSeatTaken               409, "that seat is already booked"
//	anything else              500, "something went wrong"
//
// The message for a 500 never includes err itself: it may hold SQL, hostnames or
// file paths that are none of the client's business.
func StatusFor(err error) (int, string) {
	// Your code here. Follow the steps in the README.
	return 0, ""
}

// API serves trips and bookings.
type API struct {
	store   Store
	limiter Limiter
	tokens  map[string]string // bearer token → role, "admin" or "clerk"
}

// NewAPI returns an API on store. limiter is asked before every booking, keyed by
// the client's IP address. tokens maps each valid bearer token to its role; only
// "admin" may cancel bookings. Day 58 replaces this map with real logins.
func NewAPI(store Store, limiter Limiter, tokens map[string]string) *API {
	return &API{store: store, limiter: limiter, tokens: tokens}
}

// Routes returns the API's handler:
//
//	GET    /trips/{id}           200 · 400 bad id · 404
//	POST   /trips/{id}/bookings  201 + Location · 400 bad JSON · 404 · 409 seat taken
//	                             · 422 bad values · 429 + Retry-After
//	GET    /bookings/{id}        200 · 400 · 404
//	DELETE /bookings/{id}        204 · 401 + WWW-Authenticate · 403 not admin · 404
//
// Any other store error is a 500.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	// Your code here: write writeJSON, writeError and the four handlers, then
	// register them.
	return mux
}
