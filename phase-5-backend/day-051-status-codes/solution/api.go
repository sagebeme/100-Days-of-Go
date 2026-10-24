// Package statuscodes is the matatu-booking API with the right status code for
// every outcome, and one shape for every error: {"error": "..."}.
package statuscodes

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// InvalidError says a request was readable but broke a rule, such as a seat number
// outside the matatu. Its message is safe to show the client.
type InvalidError struct {
	Message string
}

func (e *InvalidError) Error() string { return e.Message }

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
	var invalid *InvalidError
	switch {
	case err == nil:
		return http.StatusOK, ""
	case errors.As(err, &invalid):
		return http.StatusUnprocessableEntity, invalid.Message
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, ErrSeatTaken):
		return http.StatusConflict, "that seat is already booked"
	default:
		return http.StatusInternalServerError, "something went wrong"
	}
}

type errorBody struct {
	Error string `json:"error"`
}

// writeJSON sends v as JSON with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writing response: %v", err)
	}
}

// writeError sends {"error": message} with the given status.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Error: message})
}

// fail answers with the status StatusFor picks for err. A 500 logs the real error,
// which only you can read, and sends the client the generic message.
func fail(w http.ResponseWriter, r *http.Request, err error) {
	status, msg := StatusFor(err)
	if status == http.StatusInternalServerError {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}
	writeError(w, status, msg)
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
	mux.HandleFunc("GET /trips/{id}", a.getTrip)
	mux.HandleFunc("POST /trips/{id}/bookings", a.createBooking)
	mux.HandleFunc("GET /bookings/{id}", a.getBooking)
	mux.HandleFunc("DELETE /bookings/{id}", a.cancelBooking)
	return mux
}

func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "id must be a positive whole number")
		return 0, false
	}
	return id, true
}

func (a *API) getTrip(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	t, err := a.store.Trip(r.Context(), id)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

type bookingInput struct {
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
}

// clientIP is the address the request came from, without the port.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (a *API) createBooking(w http.ResponseWriter, r *http.Request) {
	if ok, wait := a.limiter.Allow(clientIP(r)); !ok {
		secs := int(math.Ceil(wait.Seconds()))
		if secs < 1 {
			secs = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(secs))
		writeError(w, http.StatusTooManyRequests, "too many bookings, try again later")
		return
	}
	tripID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in bookingInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "body must be valid JSON")
		return
	}
	trip, err := a.store.Trip(r.Context(), tripID)
	if err != nil {
		fail(w, r, err)
		return
	}
	if err := check(in, trip); err != nil {
		fail(w, r, err)
		return
	}
	b, err := a.store.Book(r.Context(), Booking{TripID: tripID, Seat: in.Seat, Passenger: in.Passenger, Phone: in.Phone})
	if err != nil {
		fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/bookings/"+strconv.Itoa(b.ID))
	writeJSON(w, http.StatusCreated, b)
}

// check returns an *InvalidError if the booking breaks a rule. Day 53 does this
// properly, with every problem listed at once.
func check(in bookingInput, trip Trip) error {
	if in.Seat < 1 || in.Seat > trip.Seats {
		return &InvalidError{Message: "seat must be between 1 and " + strconv.Itoa(trip.Seats)}
	}
	if strings.TrimSpace(in.Passenger) == "" {
		return &InvalidError{Message: "passenger is required"}
	}
	if strings.TrimSpace(in.Phone) == "" {
		return &InvalidError{Message: "phone is required"}
	}
	return nil
}

func (a *API) getBooking(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	b, err := a.store.Booking(r.Context(), id)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (a *API) cancelBooking(w http.ResponseWriter, r *http.Request) {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	role, known := a.tokens[token]
	if !found || token == "" || !known {
		// 401: we don't know who you are. Logging in could fix it.
		w.Header().Set("WWW-Authenticate", `Bearer realm="bookings"`)
		writeError(w, http.StatusUnauthorized, "log in first")
		return
	}
	if role != "admin" {
		// 403: we know who you are, and logging in again won't change the answer.
		writeError(w, http.StatusForbidden, "only an admin can cancel bookings")
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := a.store.CancelBooking(r.Context(), id); err != nil {
		fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
