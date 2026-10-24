package owasptop10

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"unicode/utf8"
)

// Profile is a Twende account. Only Name is the passenger's to change.
type Profile struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// ProfileStore keeps profiles by phone number.
type ProfileStore interface {
	Get(phone string) (Profile, bool)
	Put(p Profile)
}

// MemProfiles is an in-memory ProfileStore, safe for concurrent use.
type MemProfiles struct {
	mu sync.Mutex
	m  map[string]Profile
}

// NewMemProfiles returns a store holding ps.
func NewMemProfiles(ps ...Profile) *MemProfiles {
	s := &MemProfiles{m: map[string]Profile{}}
	for _, p := range ps {
		s.m[p.Phone] = p
	}
	return s
}

// Get returns the profile for phone.
func (s *MemProfiles) Get(phone string) (Profile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.m[phone]
	return p, ok
}

// Put saves p.
func (s *MemProfiles) Put(p Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[p.Phone] = p
}

type phoneKey struct{}

// WithPhone returns a copy of ctx carrying the logged-in user's phone, as the auth
// middleware from Days 58 and 59 would.
func WithPhone(ctx context.Context, phone string) context.Context {
	return context.WithValue(ctx, phoneKey{}, phone)
}

// PhoneFrom returns the phone stored by WithPhone.
func PhoneFrom(ctx context.Context) (string, bool) {
	p, ok := ctx.Value(phoneKey{}).(string)
	return p, ok && p != ""
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

// profileUpdate lists the only fields a passenger may send. Decoding into Profile
// itself would let {"role": "admin"} through: mass assignment.
type profileUpdate struct {
	Name string `json:"name"`
}

// UpdateProfile handles PATCH /me. The profile is the logged-in user's, found by the
// phone in the context, never by an ID or phone in the request. The body may hold
// only "name" (1 to 50 characters); any other field is a 400.
func UpdateProfile(store ProfileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phone, ok := PhoneFrom(r.Context())
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "log in first")
			return
		}
		p, ok := store.Get(phone)
		if !ok {
			writeError(w, http.StatusNotFound, "no such profile")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var upd profileUpdate
		if err := dec.Decode(&upd); err != nil {
			writeError(w, http.StatusBadRequest, "body must be JSON with only a name")
			return
		}
		if n := utf8.RuneCountInString(upd.Name); n < 1 || n > 50 {
			writeError(w, http.StatusUnprocessableEntity, "name must be 1 to 50 characters")
			return
		}
		p.Name = upd.Name
		store.Put(p)
		writeJSON(w, http.StatusOK, p)
	}
}
