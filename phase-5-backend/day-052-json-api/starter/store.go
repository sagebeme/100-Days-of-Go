package jsonapi

// This file is already written for you, and it's the same in starter/ and solution/.
// Today's work is in api.go.

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Nairobi is East Africa Time, UTC+3. Kenya has no daylight saving, so a fixed zone
// is exact, and it works on machines without a time zone database.
var Nairobi = time.FixedZone("EAT", 3*60*60)

// Trip is one matatu departure on a route. The store keeps Departs in UTC.
type Trip struct {
	ID      int       `json:"id"`
	RouteID int       `json:"routeId"`
	Departs time.Time `json:"departs"`
	Seats   int       `json:"seats"`
}

// MemStore keeps trips in memory. It is safe to use from many goroutines.
type MemStore struct {
	mu    sync.Mutex
	trips map[int]Trip
	next  int
}

// NewMemStore returns an empty store. The first trip gets ID 1.
func NewMemStore() *MemStore {
	return &MemStore{trips: map[int]Trip{}, next: 1}
}

// CreateTrip saves t with a new ID and its departure time converted to UTC.
func (s *MemStore) CreateTrip(t Trip) Trip {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = s.next
	s.next++
	t.Departs = t.Departs.UTC()
	s.trips[t.ID] = t
	return t
}

// Trip returns the trip with this ID, and false if there isn't one.
func (s *MemStore) Trip(id int) (Trip, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.trips[id]
	return t, ok
}

// Trips returns every trip, sorted by departure time and then ID.
func (s *MemStore) Trips() []Trip {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Trip, 0, len(s.trips))
	for _, t := range s.trips {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Departs.Equal(out[j].Departs) {
			return out[i].Departs.Before(out[j].Departs)
		}
		return out[i].ID < out[j].ID
	})
	return out
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
