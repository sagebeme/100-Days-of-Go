package validation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// MaxBody is the largest booking body the API reads, in bytes. A real booking is
// under 200 bytes; anything near this limit is a mistake or an attack.
const MaxBody = 4096

// validationBody is the 422 answer: the usual error plus every field problem.
type validationBody struct {
	Error  string `json:"error"`
	Fields Errors `json:"fields"`
}

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
	mux.HandleFunc("POST /trips/{id}/bookings", a.createBooking)
	return mux
}

func (a *API) createBooking(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req BookingRequest
	if err := dec.Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "body must be "+strconv.Itoa(MaxBody)+" bytes or less")
			return
		}
		writeError(w, http.StatusBadRequest, "body must be valid JSON with seat, passenger and phone")
		return
	}

	tripID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || tripID < 1 {
		writeError(w, http.StatusBadRequest, "id must be a positive whole number")
		return
	}
	trip, err := a.store.Trip(tripID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "trip not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	clean, errs := ValidateBooking(req, trip, a.now())
	if errs != nil {
		writeJSON(w, http.StatusUnprocessableEntity, validationBody{Error: "validation failed", Fields: errs})
		return
	}

	b, err := a.store.Book(Booking{TripID: tripID, Seat: clean.Seat, Passenger: clean.Passenger, Phone: clean.Phone})
	switch {
	case errors.Is(err, ErrSeatTaken):
		writeError(w, http.StatusConflict, "that seat is already booked")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	w.Header().Set("Location", "/bookings/"+strconv.Itoa(b.ID))
	writeJSON(w, http.StatusCreated, b)
}
