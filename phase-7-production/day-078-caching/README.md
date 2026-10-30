# Day 78: Caching

## The brief

Every passenger who opens Twende sees the trip list, and every trip card asks Postgres for the same row: route, departure time, seats free. At 7 am that's thousands of identical queries a second, and the database is the box from [Day 77](../day-077-system-design-basics/) you can't simply copy. Today you put Redis in front of it with the **cache-aside** pattern: look in Redis first, go to the database only on a miss, and keep a copy for next time.

```
Get("46K-0730")   Redis miss  →  Postgres  →  SET trip:46K-0730 {...} EX 30   →  {SeatsFree: 9}
Get("46K-0730")   Redis hit                                                   →  {SeatsFree: 9}
UpdateSeats("46K-0730", 3)    →  Postgres UPDATE  →  DEL trip:46K-0730
Get("46K-0730")   Redis miss  →  Postgres                                     →  {SeatsFree: 3}

10 passengers open the same trip at the same moment, cache empty
                  →  1 Postgres query, 10 answers
Redis is down     →  Postgres, a warning in the log, the passenger notices nothing
```

## What you'll use

- **go-redis.** `redis.NewClient(&redis.Options{Addr: "localhost:6379"})` connects. `rdb.Get(ctx, key).Bytes()` returns the value, or the error `redis.Nil` when the key doesn't exist. Any other error means Redis itself has a problem. `rdb.Set(ctx, key, value, ttl).Err()` stores with an expiry, and `rdb.Del(ctx, key).Err()` deletes.
- **miniredis.** Tests run against `miniredis.RunT(t)`, a Redis that lives in memory inside the test. `mr.FastForward(30*time.Second)` moves its clock on, so TTLs expire without sleeping.
- **TTL (time to live).** Every cached value gets an expiry. It's your safety net: if an invalidation is ever missed, the stale value still goes away.
- **Invalidate on write.** Write the database first, then **delete** the key. Don't write the new value into the cache: two writers can finish in the opposite order to how they started and leave the older value cached. A delete can't be out of order; the next read loads whatever is true. One gap remains: a reader that loaded the old row just before your write can put it back just after your delete. The TTL limits how long that lasts.
- **Stampede protection.** When a hot key expires, every request misses at once and they all hit the database together. A `Group` lets the first caller do the load while the rest wait for its answer. It's a small version of `golang.org/x/sync/singleflight`, built with a mutex, a map and a channel that's closed when the load finishes (Days 42 and 45).
- **Generics ([Day 18](../../phase-2-modelling/day-018-typed-inventory/)).** `Group[T any]` works for any result type, so `Group[Trip]` returns a `Trip` with no type assertions.

```go
raw, err := rdb.Get(ctx, key).Bytes()
switch {
case err == nil:          // hit: decode raw
case errors.Is(err, redis.Nil): // miss
default:                  // Redis is unwell: log it and go to the database
}
```

**What not to cache.** The seat count on the trip card can be 30 seconds old; the booking itself can't. `BookSeat` must check and decrement seats in Postgres inside a transaction, never trust the cached number. Don't cache anything per user with secrets in it (tokens, phone numbers) under a key someone else could guess, don't cache error responses as if they were data, and don't cache what's cheap to compute anyway. A cache you can't explain is a bug you can't find.

## Why it matters at work

Caches turn a database that would need upgrading into one that coasts, so they appear in nearly every service. They also cause a famous class of bugs: the price that didn't update, the user who still sees a deleted post, the outage that started when the cache restarted and the database took the whole load at once. In code review, look for three things: every `Set` has a TTL, every write path deletes the keys it affects, and a Redis error falls back to the database instead of failing the request. The stampede fix matters most on the hottest keys, which are exactly the ones a load test at 5 pm on a Friday will find.

## Steps

1. Open `starter/repo.go`. It's finished: `Trip`, the `Repository` interface and `SlowRepo`, an in-memory repository that counts its reads so the tests can tell a hit from a miss.
2. In `starter/cache.go`, write `Key`.
3. Write `Group.Do` in `starter/singleflight.go`. Lock the mutex. If the key has a call in progress, add one to its `callers`, unlock, wait on `<-c.done` and return its result. Otherwise make a new `call` with `callers: 1`, store it, unlock, run `fn`, then lock again, delete the key, unlock and `close(c.done)`. Make the map the first time if it's `nil`.
4. Write `Group.Callers`.
5. Write `TripCache.Get`. Read Redis. On a hit, `json.Unmarshal` and return. On `redis.Nil`, or on any other error or bad JSON after logging a warning with `c.log.Warn`, call `c.group.Do(key, ...)`: load from `c.repo`, `json.Marshal` the trip and `Set` it with `c.ttl`. Return repository errors unchanged and don't cache them. A failed `Set` is only a warning.
6. Write `TripCache.UpdateSeats`: repository first, then `Del`.
7. Run the tests from the repo root:

```bash
go test -race ./phase-7-production/day-078-caching/starter
```

To try a real Redis, start one with `docker run --rm -p 6379:6379 redis:7` and run `REDIS_ADDR=localhost:6379 go test ./phase-7-production/day-078-caching/starter -run Real`. Without `REDIS_ADDR` that test is skipped.

## The kata: LRU Cache

Build the classic interview cache: a fixed capacity, O(1) `get` and `put`, and the least recently used key thrown out when it's full. It's what an in-process cache does when there's no Redis. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 78
```

## When you're stuck

- **`repository read 3 times for 3 Gets, want 1`**: nothing reaches Redis. Check `Set` runs on a miss and that you return early on a hit.
- **`TTL = 0s, want 30s`**: you passed `0` as the expiry, which means "keep for ever". Pass `c.ttl`.
- **`a missing trip was cached`**: return the repository error before marshalling and setting.
- **`Get with Redis down = {...}, dial tcp ...: connection refused`**: you returned the Redis error. Only `redis.Nil` means "miss", but every other error should also fall through to the database.
- **`gave up waiting: 10 callers sharing one run`**: `Callers` never reaches 10. Either each caller starts its own run (look up the key before adding a new call), or you forgot `c.callers++`.
- **`fn ran 1 times ... want 2`**: you never deleted the key after the run, so later calls get the old result for ever.
- **`fatal error: all goroutines are asleep - deadlock!`**: you waited on `<-c.done` while still holding `g.mu`, or never closed `c.done`. Unlock before waiting.
- **The race detector reports `c.val`**: write `c.val` and `c.err` before `close(c.done)`; waiters read them only after the close.
- **Still stuck?** Open `solution/cache.go` and `solution/singleflight.go`, read them once, close them, then write your own version from memory.
