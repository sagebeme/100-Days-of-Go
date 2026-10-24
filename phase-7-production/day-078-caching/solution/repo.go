package caching

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// ErrNotFound is returned when a trip doesn't exist.
var ErrNotFound = errors.New("trip not found")

// Trip is one matatu trip and how many seats are still free.
type Trip struct {
	ID        string `json:"id"`
	Route     string `json:"route"`
	Departs   string `json:"departs"`
	SeatsFree int    `json:"seatsFree"`
}

// Repository is where trips really live, usually Postgres. It's slow compared with Redis.
type Repository interface {
	GetTrip(ctx context.Context, id string) (Trip, error)
	UpdateSeats(ctx context.Context, id string, seatsFree int) error
}

// SlowRepo is an in-memory Repository that counts its reads, so tests can tell whether
// a read came from the cache. If Gate is not nil, every GetTrip first sends on Entered
// (when that is not nil) and then waits for Gate to be closed, which lets a test hold a
// slow query open.
type SlowRepo struct {
	mu      sync.Mutex
	trips   map[string]Trip
	Reads   atomic.Int64
	Gate    chan struct{}
	Entered chan struct{}
}

// NewSlowRepo returns a SlowRepo holding trips.
func NewSlowRepo(trips ...Trip) *SlowRepo {
	r := &SlowRepo{trips: map[string]Trip{}}
	for _, t := range trips {
		r.trips[t.ID] = t
	}
	return r
}

// GetTrip returns the trip with this id, or ErrNotFound.
func (r *SlowRepo) GetTrip(ctx context.Context, id string) (Trip, error) {
	r.Reads.Add(1)
	if r.Gate != nil {
		if r.Entered != nil {
			r.Entered <- struct{}{}
		}
		select {
		case <-r.Gate:
		case <-ctx.Done():
			return Trip{}, ctx.Err()
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.trips[id]
	if !ok {
		return Trip{}, ErrNotFound
	}
	return t, nil
}

// UpdateSeats changes how many seats a trip has free, or returns ErrNotFound.
func (r *SlowRepo) UpdateSeats(ctx context.Context, id string, seatsFree int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.trips[id]
	if !ok {
		return ErrNotFound
	}
	t.SeatsFree = seatsFree
	r.trips[id] = t
	return nil
}
