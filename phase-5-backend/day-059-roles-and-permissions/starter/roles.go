// Package rolesandpermissions decides who may do what on the Twende booking API:
// passengers, conductors and admins, checked in one place.
package rolesandpermissions

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Role is what a user is allowed to be.
type Role string

// The three roles.
const (
	Passenger Role = "passenger"
	Conductor Role = "conductor"
	Admin     Role = "admin"
)

// Action is something a route lets a user do.
type Action string

// The actions the API knows about.
const (
	ListTrips      Action = "trips:list"
	CreateTrip     Action = "trips:create"
	CreateBooking  Action = "bookings:create"
	ReadBooking    Action = "bookings:read"     // read your own bookings
	ReadAnyBooking Action = "bookings:read-any" // read anyone's booking
	CheckIn        Action = "bookings:checkin"
)

// permissions is the whole policy in one table. Anything not listed is denied.
var permissions = map[Role]map[Action]bool{
	// Your code here: one entry per role, listing the actions it may perform.
}

// Can reports whether role may perform action. Unknown roles and actions are denied.
func Can(role Role, action Action) bool {
	// Your code here
	return false
}

// User is a logged-in user.
type User struct {
	ID    string
	Name  string
	Phone string
	Role  Role
}

// Trip is one departure on a route.
type Trip struct {
	ID      int       `json:"id"`
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

// Booking is one seat on one trip.
type Booking struct {
	ID        int    `json:"id"`
	TripID    int    `json:"tripId"`
	Seat      int    `json:"seat"`
	Passenger string `json:"passenger"`
	Phone     string `json:"phone"`
	CheckedIn bool   `json:"checkedIn"`
}

// CanViewBooking is the object-level check: staff with ReadAnyBooking see every
// booking, and anyone else with ReadBooking sees only bookings made with their phone.
func CanViewBooking(u User, b Booking) bool {
	// Your code here
	return false
}

// Authenticator finds the user making a request, if there is one.
type Authenticator interface {
	Authenticate(r *http.Request) (User, bool)
}

// StaticTokens maps bearer tokens to users. Day 58's JWTs do this job in real code.
type StaticTokens map[string]User

// Authenticate looks up the token in "Authorization: Bearer <token>".
func (s StaticTokens) Authenticate(r *http.Request) (User, bool) {
	tok, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return User{}, false
	}
	u, ok := s[tok]
	return u, ok
}

type userKey struct{}

// WithUser returns a copy of ctx carrying u.
func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}

// UserFrom returns the user stored by WithUser.
func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}

// Authenticate puts the request's user, if any, in the context. It never rejects a
// request: deciding who is allowed is Require's job.
func Authenticate(a Authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := a.Authenticate(r); ok {
			r = r.WithContext(WithUser(r.Context(), u))
		}
		next.ServeHTTP(w, r)
	})
}

// Require lets the request through only if the user may perform action: no user is
// 401 with "WWW-Authenticate: Bearer", a user without the permission is 403.
func Require(action Action, next http.Handler) http.Handler {
	// Your code here
	return next
}

// Store holds trips and bookings.
type Store interface {
	Trips() []Trip
	AddTrip(t Trip) Trip
	Trip(id int) (Trip, bool)
	AddBooking(b Booking) Booking
	Booking(id int) (Booking, bool)
	CheckIn(id int) bool
}

// MemStore is an in-memory Store, safe for concurrent use.
type MemStore struct {
	mu       sync.Mutex
	trips    []Trip
	bookings map[int]Booking
	nextID   int
}

// NewMemStore returns an empty store.
func NewMemStore() *MemStore {
	return &MemStore{bookings: map[int]Booking{}, nextID: 1}
}

// Trips returns every trip.
func (s *MemStore) Trips() []Trip {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Trip{}, s.trips...)
}

// AddTrip saves t with a new ID and returns it.
func (s *MemStore) AddTrip(t Trip) Trip {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = s.nextID
	s.nextID++
	s.trips = append(s.trips, t)
	return t
}

// Trip returns the trip with this ID.
func (s *MemStore) Trip(id int) (Trip, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.trips {
		if t.ID == id {
			return t, true
		}
	}
	return Trip{}, false
}

// AddBooking saves b with a new ID and returns it.
func (s *MemStore) AddBooking(b Booking) Booking {
	s.mu.Lock()
	defer s.mu.Unlock()
	b.ID = s.nextID
	s.nextID++
	s.bookings[b.ID] = b
	return b
}

// Booking returns the booking with this ID.
func (s *MemStore) Booking(id int) (Booking, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[id]
	return b, ok
}

// CheckIn marks a booking as boarded. It reports false if there's no such booking.
func (s *MemStore) CheckIn(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[id]
	if ok {
		b.CheckedIn = true
		s.bookings[id] = b
	}
	return ok
}

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// decode reads a small JSON body into v, rejecting unknown fields.
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "body must be 1 MB or less")
			return false
		}
		writeError(w, http.StatusBadRequest, "body must be valid JSON")
		return false
	}
	return true
}

type server struct {
	store Store
}

func (s *server) listTrips(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Trips())
}

func (s *server) createTrip(w http.ResponseWriter, r *http.Request) {
	var t Trip
	if !decode(w, r, &t) {
		return
	}
	if t.Seats < 1 || t.Seats > 33 {
		writeError(w, http.StatusUnprocessableEntity, "seats must be between 1 and 33")
		return
	}
	t = s.store.AddTrip(t)
	w.Header().Set("Location", "/trips/"+strconv.Itoa(t.ID))
	writeJSON(w, http.StatusCreated, t)
}

// bookingRequest is what a passenger sends. Who the booking is for comes from the
// logged-in user, never from the body.
type bookingRequest struct {
	TripID int `json:"tripId"`
	Seat   int `json:"seat"`
}

func (s *server) createBooking(w http.ResponseWriter, r *http.Request) {
	// Your code here
}

func (s *server) getBooking(w http.ResponseWriter, r *http.Request) {
	// Your code here
}

func (s *server) checkIn(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || !s.store.CheckIn(id) {
		writeError(w, http.StatusNotFound, "no such booking")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// NewServer returns the API. Every route names the Action it needs, and
// Authenticate runs first for all of them.
func NewServer(store Store, auth Authenticator) http.Handler {
	// Your code here
	return http.NewServeMux()
}
