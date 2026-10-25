package matatubookingapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

// Server is the booking API.
type Server struct {
	cfg      Config
	store    Store
	payer    Payer
	log      *slog.Logger
	limiter  *ipLimiter
	draining atomic.Bool
	dummy    []byte // a bcrypt hash to compare against for unknown phones
}

// NewServer returns the API. A nil cfg.Now means time.Now.
func NewServer(cfg Config, store Store, payer Payer, log *slog.Logger) *Server {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	dummy, _ := bcrypt.GenerateFromPassword([]byte("not-a-real-password-at-all"), cfg.BcryptCost)
	return &Server{
		cfg:   cfg,
		store: store,
		payer: payer,
		log:   log,
		limiter: &ipLimiter{
			rate: rate.Limit(cfg.RatePerSecond), burst: cfg.RateBurst, now: cfg.Now, header: cfg.ClientIPHeader,
			clients: map[string]*rate.Limiter{},
		},
		dummy: dummy,
	}
}

// SetDraining makes /readyz answer 503, for graceful shutdown.
func (s *Server) SetDraining() { s.draining.Store(true) }

// Handler returns every route wrapped in the middleware, outermost first: access
// log, panic recovery, security headers, CORS, rate limit, authentication.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)

	mux.HandleFunc("POST /auth/register", s.register)
	mux.HandleFunc("POST /auth/login", s.login)
	mux.Handle("PUT /users/{phone}/role", require(SetRoles, s.setRole))

	mux.HandleFunc("GET /routes", s.listRoutes)
	mux.Handle("POST /routes", require(CreateRoute, s.createRoute))
	mux.HandleFunc("GET /trips", s.listTrips)
	mux.HandleFunc("GET /trips/{id}", s.getTrip)
	mux.Handle("POST /trips", require(CreateTrip, s.createTrip))

	mux.Handle("POST /bookings", require(CreateBooking, s.createBooking))
	mux.Handle("GET /bookings/{id}", require(ReadBooking, s.getBooking))
	mux.Handle("POST /bookings/{id}/pay", require(PayBooking, s.payBooking))
	mux.Handle("GET /me/bookings", require(ReadBooking, s.myBookings))

	var h http.Handler = s.authenticate(mux)
	h = s.limiter.middleware(h)
	h = cors(s.cfg.AllowedOrigins, h)
	h = securityHeaders(h)
	h = recoverPanics(s.log, h)
	return logRequests(s.log, h)
}

// fail logs an unexpected error and answers a generic 500.
func (s *Server) fail(w http.ResponseWriter, what string, err error) {
	s.log.Error(what, "err", err)
	writeError(w, http.StatusInternalServerError, "something went wrong")
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if s.draining.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "draining"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		s.log.Error("readiness: store", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type registerRequest struct {
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decode(w, r, &req) {
		return
	}
	phone, ok := NormalizePhone(req.Phone)
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "phone must be a Kenyan mobile number like 0712345678")
		return
	}
	name := strings.TrimSpace(req.Name)
	if n := utf8.RuneCountInString(name); n < 1 || n > 50 {
		writeError(w, http.StatusUnprocessableEntity, "name must be 1 to 50 characters")
		return
	}
	if n := utf8.RuneCountInString(req.Password); n < 15 || n > 64 || len(req.Password) > 72 {
		writeError(w, http.StatusUnprocessableEntity, "password must be 15 to 64 characters")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.cfg.BcryptCost)
	if err != nil {
		s.fail(w, "hashing password", err)
		return
	}
	role := Passenger
	if slices.Contains(s.cfg.AdminPhones, phone) {
		role = Admin
	}
	u := User{Phone: phone, Name: name, Role: role, PasswordHash: string(hash)}
	if err := s.store.CreateUser(r.Context(), u); errors.Is(err, ErrConflict) {
		writeError(w, http.StatusConflict, "that phone number already has an account")
		return
	} else if err != nil {
		s.fail(w, "saving user", err)
		return
	}
	s.log.Info("user registered", "phone", phone, "role", role)
	writeJSON(w, http.StatusCreated, u)
}

type loginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int    `json:"expiresIn"` // seconds
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decode(w, r, &req) {
		return
	}
	phone, _ := NormalizePhone(req.Phone)
	u, err := s.store.User(r.Context(), phone)
	if err != nil && !errors.Is(err, ErrNotFound) {
		s.fail(w, "loading user", err)
		return
	}
	hash := []byte(u.PasswordHash)
	if err != nil {
		hash = s.dummy // same work for an unknown phone, so timing doesn't tell
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(req.Password)) != nil || err != nil {
		s.log.Warn("login failed", "phone", phone)
		writeError(w, http.StatusUnauthorized, "wrong phone number or password")
		return
	}
	tok, err := s.newToken(u)
	if err != nil {
		s.fail(w, "signing token", err)
		return
	}
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: tok, TokenType: "Bearer", ExpiresIn: int(s.cfg.AccessTTL / time.Second)})
}

type roleRequest struct {
	Role Role `json:"role"`
}

func (s *Server) setRole(w http.ResponseWriter, r *http.Request) {
	var req roleRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Role != Passenger && req.Role != Conductor && req.Role != Admin {
		writeError(w, http.StatusUnprocessableEntity, "role must be passenger, conductor or admin")
		return
	}
	phone, _ := NormalizePhone(r.PathValue("phone"))
	u, err := s.store.SetRole(r.Context(), phone, req.Role)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "no such user")
		return
	}
	if err != nil {
		s.fail(w, "setting role", err)
		return
	}
	admin, _ := userFrom(r.Context())
	s.log.Info("role changed", "by", admin.Phone, "phone", u.Phone, "role", u.Role)
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) listRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := s.store.Routes(r.Context())
	if err != nil {
		s.fail(w, "listing routes", err)
		return
	}
	if routes == nil {
		routes = []Route{}
	}
	writeJSON(w, http.StatusOK, routes)
}

func (s *Server) createRoute(w http.ResponseWriter, r *http.Request) {
	var rt Route
	if !decode(w, r, &rt) {
		return
	}
	if rt.Number == "" || rt.Name == "" || rt.From == "" || rt.To == "" {
		writeError(w, http.StatusUnprocessableEntity, "number, name, from and to are required")
		return
	}
	if rt.Fare < 1 || rt.Fare > 5000 {
		writeError(w, http.StatusUnprocessableEntity, "fare must be between 1 and 5000 shillings")
		return
	}
	rt, err := s.store.CreateRoute(r.Context(), rt)
	if err != nil {
		s.fail(w, "saving route", err)
		return
	}
	w.Header().Set("Location", "/routes/"+strconv.Itoa(rt.ID))
	writeJSON(w, http.StatusCreated, rt)
}

// tripView is a trip as the API shows it, with what's left to book.
type tripView struct {
	Trip
	Fare       int   `json:"fare"`
	SeatsLeft  int   `json:"seatsLeft"`
	TakenSeats []int `json:"takenSeats"`
}

func (s *Server) view(ctx context.Context, t Trip) (tripView, error) {
	rt, err := s.store.Route(ctx, t.RouteID)
	if err != nil {
		return tripView{}, err
	}
	taken, err := s.store.TakenSeats(ctx, t.ID)
	if err != nil {
		return tripView{}, err
	}
	return tripView{Trip: t, Fare: rt.Fare, SeatsLeft: t.Seats - len(taken), TakenSeats: taken}, nil
}

func (s *Server) listTrips(w http.ResponseWriter, r *http.Request) {
	routeID := 0
	if q := r.URL.Query().Get("route"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "route must be a route ID")
			return
		}
		routeID = n
	}
	trips, err := s.store.Trips(r.Context(), routeID)
	if err != nil {
		s.fail(w, "listing trips", err)
		return
	}
	out := []tripView{}
	for _, t := range trips {
		v, err := s.view(r.Context(), t)
		if err != nil {
			s.fail(w, "loading trip", err)
			return
		}
		out = append(out, v)
	}
	writeJSON(w, http.StatusOK, out)
}

// pathID reads the {id} wildcard; anything that isn't a positive number is a 404.
func pathID(w http.ResponseWriter, r *http.Request, what string) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		writeError(w, http.StatusNotFound, "no such "+what)
		return 0, false
	}
	return id, true
}

func (s *Server) getTrip(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "trip")
	if !ok {
		return
	}
	t, err := s.store.Trip(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "no such trip")
		return
	}
	if err != nil {
		s.fail(w, "loading trip", err)
		return
	}
	v, err := s.view(r.Context(), t)
	if err != nil {
		s.fail(w, "loading trip", err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

type tripRequest struct {
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

func (s *Server) createTrip(w http.ResponseWriter, r *http.Request) {
	var req tripRequest
	if !decode(w, r, &req) {
		return
	}
	if _, err := s.store.Route(r.Context(), req.RouteID); errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusUnprocessableEntity, "routeId must be an existing route")
		return
	} else if err != nil {
		s.fail(w, "loading route", err)
		return
	}
	if !req.Departs.After(s.cfg.Now()) {
		writeError(w, http.StatusUnprocessableEntity, "departs must be in the future")
		return
	}
	if req.Seats < 1 || req.Seats > 33 {
		writeError(w, http.StatusUnprocessableEntity, "seats must be between 1 and 33")
		return
	}
	t, err := s.store.CreateTrip(r.Context(), Trip{RouteID: req.RouteID, Departs: req.Departs.UTC(), Seats: req.Seats})
	if err != nil {
		s.fail(w, "saving trip", err)
		return
	}
	w.Header().Set("Location", "/trips/"+strconv.Itoa(t.ID))
	writeJSON(w, http.StatusCreated, t)
}

type bookingRequest struct {
	TripID int `json:"tripId"`
	Seat   int `json:"seat"`
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request) {
	u, _ := userFrom(r.Context())
	var req bookingRequest
	if !decode(w, r, &req) {
		return
	}
	t, err := s.store.Trip(r.Context(), req.TripID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "no such trip")
		return
	}
	if err != nil {
		s.fail(w, "loading trip", err)
		return
	}
	if !t.Departs.After(s.cfg.Now()) {
		writeError(w, http.StatusConflict, "this trip has already left")
		return
	}
	if req.Seat < 1 || req.Seat > t.Seats {
		writeError(w, http.StatusUnprocessableEntity, "seat must be between 1 and "+strconv.Itoa(t.Seats))
		return
	}
	rt, err := s.store.Route(r.Context(), t.RouteID)
	if err != nil {
		s.fail(w, "loading route", err)
		return
	}
	b, err := s.store.CreateBooking(r.Context(), Booking{
		TripID: t.ID, Seat: req.Seat, Passenger: u.Name, Phone: u.Phone, Fare: rt.Fare,
	})
	if errors.Is(err, ErrSeatTaken) {
		writeError(w, http.StatusConflict, "seat "+strconv.Itoa(req.Seat)+" is already booked")
		return
	}
	if err != nil {
		s.fail(w, "saving booking", err)
		return
	}
	w.Header().Set("Location", "/bookings/"+strconv.Itoa(b.ID))
	writeJSON(w, http.StatusCreated, b)
}

// ownBooking loads a booking the user may see. Someone else's booking gets the
// same 404 as a missing one.
func (s *Server) ownBooking(w http.ResponseWriter, r *http.Request, anyOK bool) (Booking, bool) {
	u, _ := userFrom(r.Context())
	id, ok := pathID(w, r, "booking")
	if !ok {
		return Booking{}, false
	}
	b, err := s.store.Booking(r.Context(), id)
	if err != nil && !errors.Is(err, ErrNotFound) {
		s.fail(w, "loading booking", err)
		return Booking{}, false
	}
	if err != nil || (b.Phone != u.Phone && !(anyOK && Can(u.Role, ReadAnyBooking))) {
		writeError(w, http.StatusNotFound, "no such booking")
		return Booking{}, false
	}
	return b, true
}

func (s *Server) getBooking(w http.ResponseWriter, r *http.Request) {
	if b, ok := s.ownBooking(w, r, true); ok {
		writeJSON(w, http.StatusOK, b)
	}
}

func (s *Server) myBookings(w http.ResponseWriter, r *http.Request) {
	u, _ := userFrom(r.Context())
	bs, err := s.store.BookingsByPhone(r.Context(), u.Phone)
	if err != nil {
		s.fail(w, "listing bookings", err)
		return
	}
	writeJSON(w, http.StatusOK, bs)
}

func (s *Server) payBooking(w http.ResponseWriter, r *http.Request) {
	b, ok := s.ownBooking(w, r, false) // only the passenger pays for their own seat
	if !ok {
		return
	}
	if b.Paid {
		writeError(w, http.StatusConflict, "this booking is already paid")
		return
	}
	ref := "MAT" + strconv.Itoa(b.ID)
	if err := s.payer.Charge(r.Context(), b.Phone, b.Fare, ref); err != nil {
		s.log.Warn("payment failed", "booking", b.ID, "err", err)
		writeError(w, http.StatusBadGateway, "the payment didn't go through, try again")
		return
	}
	paid, err := s.store.MarkPaid(r.Context(), b.ID)
	if errors.Is(err, ErrAlreadyPaid) {
		writeError(w, http.StatusConflict, "this booking is already paid")
		return
	}
	if err != nil {
		s.fail(w, "marking paid", err)
		return
	}
	s.log.Info("booking paid", "booking", b.ID, "amount", b.Fare)
	writeJSON(w, http.StatusOK, paid)
}
