// Package httpmethods gives the matatu-booking API one route per method: GET to
// read trips, POST to create them, PUT to replace, PATCH to change, DELETE to
// remove, and POST with an Idempotency-Key to book a seat exactly once.
package httpmethods

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
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
	mux.HandleFunc("GET /trips", a.listTrips)
	mux.HandleFunc("POST /trips", a.createTrip)
	mux.HandleFunc("GET /trips/{id}", a.getTrip)
	mux.HandleFunc("PUT /trips/{id}", a.replaceTrip)
	mux.HandleFunc("PATCH /trips/{id}", a.patchTrip)
	mux.HandleFunc("DELETE /trips/{id}", a.deleteTrip)
	mux.HandleFunc("POST /trips/{id}/bookings", a.createBooking)
	mux.HandleFunc("GET /bookings/{id}", a.getBooking)
	return mux
}

// pathID reads the {id} wildcard. On a bad ID it writes a 400 and returns false.
func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "id must be a positive whole number")
		return 0, false
	}
	return id, true
}

// tripInput is the body of POST /trips and PUT /trips/{id}: a whole trip.
type tripInput struct {
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

// tripPatch is the body of PATCH. A nil field wasn't sent and stays as it is.
type tripPatch struct {
	RouteID *int       `json:"routeId"`
	Departs *time.Time `json:"departs"`
	Seats   *int       `json:"seats"`
}

const wholeTrip = "send the whole trip: routeId and seats above 0, and departs"

func (in tripInput) complete() bool {
	return in.RouteID > 0 && in.Seats > 0 && !in.Departs.IsZero()
}

func (a *API) listTrips(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Trips())
}

func (a *API) createTrip(w http.ResponseWriter, r *http.Request) {
	var in tripInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "body must be valid JSON")
		return
	}
	if !in.complete() {
		writeError(w, http.StatusBadRequest, wholeTrip)
		return
	}
	t := a.store.CreateTrip(Trip{RouteID: in.RouteID, Departs: in.Departs, Seats: in.Seats})
	w.Header().Set("Location", "/trips/"+strconv.Itoa(t.ID))
	writeJSON(w, http.StatusCreated, t)
}

func (a *API) getTrip(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	t, ok := a.store.Trip(id)
	if !ok {
		writeError(w, http.StatusNotFound, "trip not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// replaceTrip is PUT: every field is replaced, so the body must hold them all.
// Sending the same PUT twice leaves the same trip: it's idempotent.
func (a *API) replaceTrip(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in tripInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "body must be valid JSON")
		return
	}
	if !in.complete() {
		writeError(w, http.StatusBadRequest, wholeTrip)
		return
	}
	t, ok := a.store.UpdateTrip(id, func(t *Trip) {
		t.RouteID, t.Departs, t.Seats = in.RouteID, in.Departs, in.Seats
	})
	if !ok {
		writeError(w, http.StatusNotFound, "trip not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// patchTrip is PATCH: only the fields in the body change.
func (a *API) patchTrip(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var p tripPatch
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "body must be valid JSON")
		return
	}
	if (p.RouteID != nil && *p.RouteID < 1) || (p.Seats != nil && *p.Seats < 1) {
		writeError(w, http.StatusBadRequest, "routeId and seats must be above 0")
		return
	}
	if p.Departs != nil && p.Departs.IsZero() {
		writeError(w, http.StatusBadRequest, "departs can't be empty")
		return
	}
	t, ok := a.store.UpdateTrip(id, func(t *Trip) {
		if p.RouteID != nil {
			t.RouteID = *p.RouteID
		}
		if p.Departs != nil {
			t.Departs = *p.Departs
		}
		if p.Seats != nil {
			t.Seats = *p.Seats
		}
	})
	if !ok {
		writeError(w, http.StatusNotFound, "trip not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// deleteTrip is DELETE. A second DELETE finds nothing and answers 404, but the
// state is the same as after the first, which is what makes DELETE idempotent.
func (a *API) deleteTrip(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if !a.store.DeleteTrip(id) {
		writeError(w, http.StatusNotFound, "trip not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) getBooking(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	b, ok := a.store.Booking(id)
	if !ok {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// createBooking is POST, which isn't idempotent: two identical requests make two
// bookings. A client that may retry sends an Idempotency-Key header, and a retry
// with the same key and the same body gets the first answer again instead of a
// second booking. The same key with a different body is a client bug: 422.
func (a *API) createBooking(w http.ResponseWriter, r *http.Request) {
	tripID, ok := pathID(w, r)
	if !ok {
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "couldn't read the body")
		return
	}

	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		status, location, v := a.book(tripID, body)
		if location != "" {
			w.Header().Set("Location", location)
		}
		writeJSON(w, status, v)
		return
	}
	if len(key) > 255 {
		writeError(w, http.StatusBadRequest, "Idempotency-Key must be 255 characters or fewer")
		return
	}

	request := strconv.Itoa(tripID) + "\n" + string(body)
	a.mu.Lock()
	defer a.mu.Unlock()

	if saved, ok := a.seen[key]; ok {
		if saved.request != request {
			writeError(w, http.StatusUnprocessableEntity, "this Idempotency-Key was already used for a different request")
			return
		}
		w.Header().Set("Location", saved.location)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Idempotent-Replayed", "true")
		w.WriteHeader(saved.status)
		w.Write(saved.body)
		return
	}

	status, location, v := a.book(tripID, body)
	if location != "" {
		w.Header().Set("Location", location)
	}
	if status >= 200 && status < 300 {
		b, err := json.Marshal(v)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "something went wrong")
			return
		}
		b = append(b, '\n') // the same bytes writeJSON's Encoder would send
		a.seen[key] = savedResponse{request: request, status: status, location: location, body: b}
	}
	writeJSON(w, status, v)
}

type bookingInput struct {
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
}

// book does the booking and returns the status, the Location (or "") and the value
// to send as JSON. It writes nothing, so its answer can be saved and replayed.
func (a *API) book(tripID int, body []byte) (int, string, any) {
	var in bookingInput
	if err := json.Unmarshal(body, &in); err != nil {
		return http.StatusBadRequest, "", errorBody{"body must be valid JSON"}
	}
	trip, ok := a.store.Trip(tripID)
	if !ok {
		return http.StatusNotFound, "", errorBody{"trip not found"}
	}
	if in.Seat < 1 || in.Seat > trip.Seats {
		return http.StatusBadRequest, "", errorBody{"seat must be between 1 and " + strconv.Itoa(trip.Seats)}
	}
	if in.Passenger == "" || in.Phone == "" {
		return http.StatusBadRequest, "", errorBody{"passenger and phone are required"}
	}
	b, err := a.store.Book(Booking{TripID: tripID, Seat: in.Seat, Passenger: in.Passenger, Phone: in.Phone})
	switch {
	case errors.Is(err, ErrSeatTaken):
		return http.StatusConflict, "", errorBody{"that seat is already booked"}
	case errors.Is(err, ErrNoTrip):
		return http.StatusNotFound, "", errorBody{"trip not found"}
	case err != nil:
		return http.StatusInternalServerError, "", errorBody{"something went wrong"}
	}
	return http.StatusCreated, "/bookings/" + strconv.Itoa(b.ID), b
}
