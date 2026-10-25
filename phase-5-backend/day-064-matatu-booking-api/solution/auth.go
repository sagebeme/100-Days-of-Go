package matatubookingapi

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Action is something a route lets a user do.
type Action string

// The actions.
const (
	CreateRoute    Action = "routes:create"
	CreateTrip     Action = "trips:create"
	CreateBooking  Action = "bookings:create"
	ReadBooking    Action = "bookings:read"
	ReadAnyBooking Action = "bookings:read-any"
	PayBooking     Action = "bookings:pay"
	SetRoles       Action = "users:set-role"
)

var permissions = map[Role]map[Action]bool{
	Passenger: {CreateBooking: true, ReadBooking: true, PayBooking: true},
	Conductor: {ReadBooking: true, ReadAnyBooking: true},
	Admin: {CreateRoute: true, CreateTrip: true, CreateBooking: true, ReadBooking: true,
		ReadAnyBooking: true, PayBooking: true, SetRoles: true},
}

// Can reports whether role may perform action. Anything not listed is denied.
func Can(role Role, action Action) bool { return permissions[role][action] }

type claims struct {
	Role Role `json:"role"`
	jwt.RegisteredClaims
}

// newToken signs an HS256 access token for u.
func (s *Server) newToken(u User) (string, error) {
	now := s.cfg.Now()
	c := claims{Role: u.Role, RegisteredClaims: jwt.RegisteredClaims{
		Subject:   u.Phone,
		Issuer:    s.cfg.Issuer,
		Audience:  jwt.ClaimStrings{s.cfg.Issuer},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTTL)),
		ID:        rand.Text(),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.cfg.JWTSecret)
}

// parseToken checks a token's signature, algorithm, issuer, audience and expiry.
func (s *Server) parseToken(raw string) (*claims, error) {
	var c claims
	_, err := jwt.ParseWithClaims(raw, &c,
		func(*jwt.Token) (any, error) { return s.cfg.JWTSecret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(s.cfg.Issuer),
		jwt.WithAudience(s.cfg.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(s.cfg.Now),
	)
	if err != nil {
		return nil, err
	}
	if c.Subject == "" {
		return nil, errors.New("no subject")
	}
	return &c, nil
}

type userKey struct{}

func userFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}

// authenticate puts the user from a valid bearer token in the context. The role is
// read from the store, not the token, so a role change applies at once. A token
// that's present but bad is a 401; no token at all just means "anonymous".
func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			next.ServeHTTP(w, r)
			return
		}
		raw, ok := strings.CutPrefix(h, "Bearer ")
		c, err := s.parseToken(raw)
		if !ok || err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			writeError(w, http.StatusUnauthorized, "token is invalid or expired")
			return
		}
		u, err := s.store.User(r.Context(), c.Subject)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			writeError(w, http.StatusUnauthorized, "token is invalid or expired")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, u)))
	})
}

// require answers 401 without a user and 403 without the permission.
func require(action Action, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := userFrom(r.Context())
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "log in first")
			return
		}
		if !Can(u.Role, action) {
			writeError(w, http.StatusForbidden, fmt.Sprintf("a %s can't do that", u.Role))
			return
		}
		next(w, r)
	})
}
