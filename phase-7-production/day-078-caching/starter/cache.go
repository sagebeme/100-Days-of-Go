// Package caching puts a Redis cache-aside layer in front of a slow trip repository for
// Twende, with TTLs, invalidation on write and protection against a cache stampede.
package caching

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// TripCache reads trips through Redis. Reads try Redis first and fall back to the
// repository on a miss; writes go to the repository and then delete the cached copy.
type TripCache struct {
	rdb   *redis.Client
	repo  Repository
	ttl   time.Duration
	group Group[Trip]
	log   *slog.Logger
}

// NewTripCache returns a TripCache that keeps each trip in Redis for ttl.
func NewTripCache(rdb *redis.Client, repo Repository, ttl time.Duration, log *slog.Logger) *TripCache {
	return &TripCache{rdb: rdb, repo: repo, ttl: ttl, log: log}
}

// Key returns the Redis key for a trip: "trip:" followed by its id.
//
//	Key("46K-0730") → "trip:46K-0730"
func Key(id string) string {
	// Your code here
	return ""
}

// Get returns the trip with this id. It reads Redis first. On a miss it loads the trip
// from the repository, at most once at a time per trip however many callers miss
// together, and stores it as JSON with the cache's TTL. If Redis is down or holds
// something that isn't valid JSON, Get logs a warning and uses the repository.
// ErrNotFound from the repository is returned as it is and nothing is cached.
func (c *TripCache) Get(ctx context.Context, id string) (Trip, error) {
	// Your code here
	return Trip{}, nil
}

// UpdateSeats writes the new seat count to the repository first and then deletes the
// cached trip, so the next Get loads the fresh value. A failed delete is returned as an
// error, because the cache would keep serving the old seat count until the TTL ran out.
func (c *TripCache) UpdateSeats(ctx context.Context, id string, seatsFree int) error {
	// Your code here
	return nil
}
