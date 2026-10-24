// Package jsonapi reads and writes the matatu-booking API's JSON strictly: the right
// Content-Type, no unknown fields, one value per body, clear errors, and times sent
// in Nairobi time with their offset.
package jsonapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// RequestError is a problem with the request body. Status is the code to answer
// with and Message is safe to show the client.
type RequestError struct {
	Status  int
	Message string
}

func (e *RequestError) Error() string { return e.Message }

func badRequest(format string, args ...any) *RequestError {
	return &RequestError{Status: http.StatusBadRequest, Message: fmt.Sprintf(format, args...)}
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
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &RequestError{Status: http.StatusUnsupportedMediaType, Message: "Content-Type must be application/json"}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return describe(err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return badRequest("body must hold a single JSON value")
	}
	return nil
}

// describe turns a decoding error into a message a client developer can act on,
// without Go type names in it.
func describe(err error) *RequestError {
	var syntax *json.SyntaxError
	var wrongType *json.UnmarshalTypeError
	var badTime *time.ParseError
	switch {
	case errors.Is(err, io.EOF):
		return badRequest("body must not be empty")
	case errors.Is(err, io.ErrUnexpectedEOF):
		return badRequest("body has malformed JSON")
	case errors.As(err, &syntax):
		return badRequest("body has malformed JSON at byte %d", syntax.Offset)
	case errors.As(err, &wrongType):
		if wrongType.Field == "" {
			return badRequest("body must be a JSON object")
		}
		return badRequest("field %q must be %s", wrongType.Field, kindName(wrongType.Type))
	case errors.As(err, &badTime):
		return badRequest(`times must look like "2026-11-02T07:30:00+03:00"`)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return badRequest("body has unknown field %s", strings.TrimPrefix(err.Error(), "json: unknown field "))
	default:
		return badRequest("body isn't valid for this request")
	}
}

func kindName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "a whole number"
	case reflect.Float32, reflect.Float64:
		return "a number"
	case reflect.String:
		return "a string"
	case reflect.Bool:
		return "true or false"
	case reflect.Slice, reflect.Array:
		return "a list"
	default:
		return "an object"
	}
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
	return TripJSON{ID: t.ID, RouteID: t.RouteID, Departs: t.Departs.In(Nairobi), Seats: t.Seats}
}

// createTripRequest is the body of POST /trips.
type createTripRequest struct {
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
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
	mux.HandleFunc("POST /trips", a.createTrip)
	mux.HandleFunc("GET /trips", a.listTrips)
	mux.HandleFunc("GET /trips/{id}", a.getTrip)
	return mux
}

func (a *API) createTrip(w http.ResponseWriter, r *http.Request) {
	var req createTripRequest
	if err := DecodeJSON(r, &req); err != nil {
		var re *RequestError
		if errors.As(err, &re) {
			writeError(w, re.Status, re.Message)
			return
		}
		writeError(w, http.StatusBadRequest, "body isn't valid for this request")
		return
	}
	if req.RouteID < 1 || req.Seats < 1 || req.Departs.IsZero() {
		writeError(w, http.StatusUnprocessableEntity, "routeId and seats must be above 0, and departs is required")
		return
	}
	t := a.store.CreateTrip(Trip{RouteID: req.RouteID, Departs: req.Departs, Seats: req.Seats})
	w.Header().Set("Location", "/trips/"+strconv.Itoa(t.ID))
	writeJSON(w, http.StatusCreated, ToJSON(t))
}

func (a *API) listTrips(w http.ResponseWriter, r *http.Request) {
	all := a.store.Trips()
	out := make([]TripJSON, 0, len(all))

	date := r.URL.Query().Get("date")
	if date == "" {
		for _, t := range all {
			out = append(out, ToJSON(t))
		}
		writeJSON(w, http.StatusOK, out)
		return
	}

	day, err := time.ParseInLocation("2006-01-02", date, Nairobi)
	if err != nil {
		writeError(w, http.StatusBadRequest, `date must look like "2026-11-02"`)
		return
	}
	next := day.AddDate(0, 0, 1)
	for _, t := range all {
		if !t.Departs.Before(day) && t.Departs.Before(next) {
			out = append(out, ToJSON(t))
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) getTrip(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "id must be a positive whole number")
		return
	}
	t, ok := a.store.Trip(id)
	if !ok {
		writeError(w, http.StatusNotFound, "trip not found")
		return
	}
	writeJSON(w, http.StatusOK, ToJSON(t))
}
