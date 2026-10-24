package caching

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

const ttl = 30 * time.Second

var trip46K = Trip{ID: "46K-0730", Route: "46K Kencom to Kawangware", Departs: "07:30", SeatsFree: 9}

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// setup starts an in-memory Redis and a TripCache in front of a SlowRepo.
func setup(t *testing.T) (*miniredis.Miniredis, *SlowRepo, *TripCache) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1})
	t.Cleanup(func() { rdb.Close() })
	repo := NewSlowRepo(trip46K, Trip{ID: "111-0800", Route: "111 Railways to Ngong", Departs: "08:00", SeatsFree: 14})
	return mr, repo, NewTripCache(rdb, repo, ttl, quiet())
}

// waitFor spins until cond is true. The deadline only stops a broken solution from
// hanging the test run; correctness never depends on timing.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("gave up waiting: %s", what)
		}
		runtime.Gosched()
	}
}

func TestKey(t *testing.T) {
	tests := []struct{ id, want string }{
		{"46K-0730", "trip:46K-0730"},
		{"", "trip:"},
	}
	for _, tt := range tests {
		if got := Key(tt.id); got != tt.want {
			t.Errorf("Key(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestMissThenHit(t *testing.T) {
	mr, repo, c := setup(t)
	ctx := context.Background()
	for i := range 3 {
		got, err := c.Get(ctx, trip46K.ID)
		if err != nil || got != trip46K {
			t.Fatalf("Get #%d = %+v, %v, want %+v, nil", i+1, got, err, trip46K)
		}
	}
	if n := repo.Reads.Load(); n != 1 {
		t.Errorf("repository read %d times for 3 Gets, want 1: the 2nd and 3rd should come from Redis", n)
	}
	raw, err := mr.Get(Key(trip46K.ID))
	if err != nil {
		t.Fatalf("nothing stored under %q: %v", Key(trip46K.ID), err)
	}
	var stored Trip
	if err := json.Unmarshal([]byte(raw), &stored); err != nil || stored != trip46K {
		t.Errorf("Redis holds %q, want the trip as JSON", raw)
	}
	if got := mr.TTL(Key(trip46K.ID)); got != ttl {
		t.Errorf("TTL = %v, want %v: every cached value needs an expiry", got, ttl)
	}
}

func TestTTLExpires(t *testing.T) {
	mr, repo, c := setup(t)
	ctx := context.Background()
	c.Get(ctx, trip46K.ID)
	mr.FastForward(ttl)
	c.Get(ctx, trip46K.ID)
	if n := repo.Reads.Load(); n != 2 {
		t.Errorf("repository read %d times, want 2: after the TTL the entry is gone and Get must reload", n)
	}
}

func TestUpdateInvalidates(t *testing.T) {
	mr, repo, c := setup(t)
	ctx := context.Background()
	c.Get(ctx, trip46K.ID)
	if err := c.UpdateSeats(ctx, trip46K.ID, 3); err != nil {
		t.Fatalf("UpdateSeats = %v, want nil", err)
	}
	if mr.Exists(Key(trip46K.ID)) {
		t.Errorf("%q still cached after UpdateSeats: delete it so nobody sees the old seat count", Key(trip46K.ID))
	}
	got, err := c.Get(ctx, trip46K.ID)
	if err != nil || got.SeatsFree != 3 {
		t.Errorf("Get after UpdateSeats = %+v, %v, want SeatsFree 3", got, err)
	}
	if n := repo.Reads.Load(); n != 2 {
		t.Errorf("repository read %d times, want 2", n)
	}
	if err := c.UpdateSeats(ctx, "nope", 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("UpdateSeats(unknown) = %v, want ErrNotFound", err)
	}
}

func TestNotFoundIsNotCached(t *testing.T) {
	mr, _, c := setup(t)
	_, err := c.Get(context.Background(), "nope")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(unknown) error = %v, want ErrNotFound", err)
	}
	if mr.Exists(Key("nope")) {
		t.Errorf("a missing trip was cached")
	}
}

func TestBadEntryIsReplaced(t *testing.T) {
	mr, repo, c := setup(t)
	mr.Set(Key(trip46K.ID), "{not json")
	got, err := c.Get(context.Background(), trip46K.ID)
	if err != nil || got != trip46K {
		t.Fatalf("Get with a corrupt entry = %+v, %v, want %+v, nil", got, err, trip46K)
	}
	if n := repo.Reads.Load(); n != 1 {
		t.Errorf("repository read %d times, want 1", n)
	}
	if raw, _ := mr.Get(Key(trip46K.ID)); raw == "{not json" {
		t.Errorf("the corrupt entry is still there; store the fresh trip over it")
	}
}

func TestRedisDownFallsBack(t *testing.T) {
	mr, _, c := setup(t)
	mr.Close()
	got, err := c.Get(context.Background(), trip46K.ID)
	if err != nil || got != trip46K {
		t.Errorf("Get with Redis down = %+v, %v, want %+v, nil: a cache outage shouldn't be an outage", got, err, trip46K)
	}
}

func TestGroupSharesOneRun(t *testing.T) {
	var g Group[int]
	var runs atomic.Int64
	gate := make(chan struct{})
	const n = 10
	results := make([]int, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			results[i], _ = g.Do("k", func() (int, error) {
				runs.Add(1)
				<-gate
				return 42, nil
			})
		})
	}
	waitFor(t, "10 callers sharing one run", func() bool { return g.Callers("k") == n })
	close(gate)
	wg.Wait()
	if r := runs.Load(); r != 1 {
		t.Errorf("fn ran %d times for %d concurrent callers, want 1", r, n)
	}
	for i, v := range results {
		if v != 42 {
			t.Errorf("caller %d got %d, want 42", i, v)
		}
	}
	if c := g.Callers("k"); c != 0 {
		t.Errorf("Callers after the run = %d, want 0", c)
	}
}

func TestGroupRunsAgainAndSharesErrors(t *testing.T) {
	var g Group[string]
	runs := 0
	boom := errors.New("db timeout")
	for range 2 {
		_, err := g.Do("k", func() (string, error) { runs++; return "", boom })
		if !errors.Is(err, boom) {
			t.Errorf("Do error = %v, want %v", err, boom)
		}
	}
	if runs != 2 {
		t.Errorf("fn ran %d times for 2 calls one after the other, want 2: forget a key once its run ends", runs)
	}
	v, err := g.Do("other", func() (string, error) { return "ok", nil })
	if v != "ok" || err != nil {
		t.Errorf("Do(other) = %q, %v, want \"ok\", nil", v, err)
	}
}

func TestStampede(t *testing.T) {
	_, repo, c := setup(t)
	repo.Gate = make(chan struct{})
	repo.Entered = make(chan struct{}, 20)
	const n = 10
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range n {
		wg.Go(func() {
			_, errs[i] = c.Get(context.Background(), trip46K.ID)
		})
	}
	waitFor(t, "10 Gets sharing one database read", func() bool { return c.group.Callers(Key(trip46K.ID)) == n })
	close(repo.Gate)
	wg.Wait()
	if r := repo.Reads.Load(); r != 1 {
		t.Errorf("repository read %d times for %d simultaneous misses, want 1", r, n)
	}
	for i, err := range errs {
		if err != nil {
			t.Errorf("Get %d = %v", i, err)
		}
	}
}

// TestRealRedis runs against a real Redis when REDIS_ADDR is set, for example:
//
//	docker run --rm -p 6379:6379 redis:7
//	REDIS_ADDR=localhost:6379 go test ./phase-7-production/day-078-caching/starter -run Real
func TestRealRedis(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	ctx := context.Background()
	rdb.Del(ctx, Key(trip46K.ID))
	repo := NewSlowRepo(trip46K)
	c := NewTripCache(rdb, repo, ttl, quiet())
	c.Get(ctx, trip46K.ID)
	got, err := c.Get(ctx, trip46K.ID)
	if err != nil || got != trip46K || repo.Reads.Load() != 1 {
		t.Errorf("against real Redis: got %+v, %v after %d reads", got, err, repo.Reads.Load())
	}
	rdb.Del(ctx, Key(trip46K.ID))
}
