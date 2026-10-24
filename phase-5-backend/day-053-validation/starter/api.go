package validation

import (
	"net/http"
	"time"
)

// MaxBody is the largest booking body the API reads, in bytes. A real booking is
// under 200 bytes; anything near this limit is a mistake or an attack.
const MaxBody = 4096

// API takes bookings.
type API struct {
	store *MemStore
	now   func() time.Time
}

// NewAPI returns an API on store. now tells the time; tests pass a fixed clock.
func NewAPI(store *MemStore, now func() time.Time) *API {
	return &API{store: store, now: now}
}

// Routes returns the handler for POST /trips/{id}/bookings:
//
//	body over MaxBody bytes      413
//	broken JSON or unknown field 400
//	bad trip ID                  400
//	no such trip                 404
//	any rule broken              422 {"error": "validation failed", "fields": [...]}
//	seat already booked          409
//	all good                     201 + Location, the booking with its normalised phone
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	// Your code here.
	return mux
}
