// Package jsonapi reads and writes the matatu-booking API's JSON strictly: the right
// Content-Type, no unknown fields, one value per body, clear errors, and times sent
// in Nairobi time with their offset.
package jsonapi

import (
	"net/http"
	"time"
)

// RequestError is a problem with the request body. Status is the code to answer
// with and Message is safe to show the client.
type RequestError struct {
	Status  int
	Message string
}

func (e *RequestError) Error() string {
	// Your code here.
	return ""
}

// DecodeJSON reads r's body into dst, which must be a pointer. It returns a
// *RequestError when the request is wrong:
//
//	Content-Type isn't application/json (parameters like charset are fine)  415
//	empty body                                     400 "body must not be empty"
//	broken JSON                                    400 "body has malformed JSON at byte N" (or no position at the end)
//	a field of the wrong type                      400 `field "seats" must be a whole number`
//	a time that isn't RFC 3339                     400 `times must look like "2026-11-02T07:30:00+03:00"`
//	a field dst doesn't have                       400 `body has unknown field "seat"`
//	something after the first JSON value           400 "body must hold a single JSON value"
func DecodeJSON(r *http.Request, dst any) error {
	// Your code here. Follow the steps in the README.
	return nil
}

// TripJSON is how a trip looks on the wire. Departs is in Nairobi time, so it
// encodes as "2026-11-02T07:30:00+03:00": a person can read it and a program can
// parse it, because the offset is there.
type TripJSON struct {
	ID      int       `json:"id"`
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

// ToJSON converts a stored trip to its wire form.
func ToJSON(t Trip) TripJSON {
	// Your code here.
	return TripJSON{}
}

// API serves trips as JSON.
type API struct {
	store *MemStore
}

// NewAPI returns an API on s.
func NewAPI(s *MemStore) *API {
	return &API{store: s}
}

// Routes returns the handler:
//
//	POST /trips                201 + Location, the trip as TripJSON
//	GET  /trips                200 a list of TripJSON, [] when empty
//	GET  /trips?date=2026-11-02  only trips leaving on that day in Nairobi; 400 for a bad date
//	GET  /trips/{id}           200 TripJSON, 400, 404
//
// POST bodies go through DecodeJSON. A body that decodes but has routeId or seats
// below 1, or no departs, gets 422.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	// Your code here.
	return mux
}
