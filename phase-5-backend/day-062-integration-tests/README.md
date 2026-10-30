# Day 62: Integration Tests

## The brief

Twende's (made-up) booking API has good unit tests. Each handler is tested on its own with `httptest.NewRecorder`, and every one passes. Yet last Friday anyone could book a seat without logging in. The handler was fine; the line in `Routes` that should have wrapped it in `requireAuth` wasn't. No unit test calls `Routes`, so nothing noticed.

Today you write the helpers for tests that run the **whole** API: a real HTTP server on a local port, the full router and middleware, real headers, starting data from a fixtures file, and a store that's in memory by default and a real Postgres when you point `DATABASE_URL` at one. The tests are already written. Once your helpers work, they'll find that Friday bug, and you fix it.

```
env := NewEnv(t, OpenTestStore(t))     // server up, fixtures loaded, closed when t ends

env.Do(t, "GET", "/healthz", "", nil)
  →  Response{Status: 200, Header: {Content-Type: application/json, X-Content-Type-Options: nosniff}, Body: {"status":"ok"}}

env.Do(t, "POST", "/bookings", "tok-wanjiku", map[string]any{"tripId": 1, "seat": 3, "passenger": "Wanjiku"})
  →  201, Location: /bookings/3

env.Do(t, "POST", "/bookings", "", sameBody)       →  401   (fails until you fix Routes)
env.Do(t, "POST", "/bookings", "tok-otieno", sameBody)  →  409   seat taken

DATABASE_URL=postgres://... go test ./...    // the same tests, on a fresh schema in a real database
```

## What you'll use

- **`httptest.NewServer(handler)`** starts a real server on `127.0.0.1` with a random free port. `srv.URL` is its address. Requests go through the network stack, the mux, every middleware and the JSON encoding, exactly as in production. `httptest.NewRecorder` (Days 49 to 61) skips all of that and calls one handler.
- **`t.Cleanup(f)`** runs `f` when the test (or subtest) ends, pass or fail. Cleanups run last-registered first, so register "drop the schema" before "close the connection" and they run in the right order. A helper that starts something registers its own cleanup, so the test can't forget.
- **`t.Helper()`** marks a function as a helper: a failure inside it is reported at the line in the test that called it, not inside the helper.
- **`testing.TB`** is the interface both `*testing.T` and `*testing.B` satisfy. Helpers that take it work in tests and benchmarks.
- **Fixtures**: the starting data in a JSON file in `testdata/`. `go test` runs with the package folder as the working directory, so `"testdata/fixtures.json"` is a plain relative path, and the go tool ignores folders named `testdata` when building.
- **A test database per test.** With `DATABASE_URL` set, `OpenTestStore` makes a schema with a random name (`test_k3j...`), creates the tables in it, and gives the store a connection whose `search_path` is that schema. Tests can run in parallel and never see each other's rows, and the cleanup drops the schema. pgx passes extra URL parameters such as `?search_path=test_k3j` to Postgres as session settings.
- **A contract test.** `testStore(t, s)` runs the same checks on `MemStore` and on `PGStore`. If the fake starts behaving differently from the real store (say, it forgets the "seat taken" rule), its contract test fails.

```go
func NewEnv(t testing.TB, store Store) *Env {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	...
}
```

This is the first time helpers live in a normal `.go` file instead of a `_test.go` file. That's allowed: `net/http/httptest` itself is a package of test helpers. The cost is that `testing` gets compiled into the package. Big projects put such helpers in their own package, often called `apitest` or `testutil`.

## Why it matters at work

Unit tests prove the pieces work; integration tests prove they're connected. Most production bugs in a web API live in the connections: a route registered without its middleware, a header set after `WriteHeader`, a JSON field name that doesn't match the front end, a database constraint the in-memory fake didn't have. Reviewers ask "is there a test that goes through the router?" for any new endpoint, and CI runs the Postgres version on every push ([Day 63](../day-063-docker-and-ci/)). Two habits keep these tests trusted: every test cleans up after itself, so the order tests run in never matters, and the tests that need a database skip clearly when there isn't one instead of failing on a laptop without Docker.

## Steps

1. Read `starter/api.go` and `starter/pgstore.go`. They're complete, apart from one bug the tests will find. Read `starter/testdata/fixtures.json` too.
2. Open `starter/testenv.go`. Write `LoadFixtures`: `os.ReadFile`, then decode with `DisallowUnknownFields`, so a typo in the fixtures file fails loudly. Call `t.Helper()` first and `t.Fatalf` on any error.
3. Write `Seed`: `AddTrip` each fixture trip and store the saved copy back into `f.Trips[i]`, so tests know the real IDs. That's why it takes a `*Fixtures`.
4. Write `NewEnv`: load `testdata/fixtures.json`, seed the store, build the `tokens` map from the fixture users, start `httptest.NewServer((&Server{...}).Routes())`, `t.Cleanup(srv.Close)`, and return the `Env` with a `Client` that has a 5-second timeout.
5. Write `Do`: marshal a non-nil body to JSON and set `Content-Type`, set `Authorization: Bearer <token>` when there's a token, send it with `e.Client`, read the whole body and close it. Then `Response.JSON`, which unmarshals the body or fails the test.
6. Run the tests:

```bash
go test ./phase-5-backend/day-062-integration-tests/starter
```

   Everything should pass except `TestAuthIsWiredUp`. Read its message, find the bug in `Routes` in `api.go`, and fix it. That one-line fix is the reason this day exists.

7. Write `OpenTestStore`. Without `DATABASE_URL`, return `NewMemStore()`. With it: open an admin `*sql.DB` with `sql.Open("pgx", dsn)`, `CREATE SCHEMA` with a name from `"test_" + strings.ToLower(rand.Text())`, and register a cleanup that drops it with `CASCADE` and closes the admin connection. Then add `search_path` to the URL's query with `url.Parse`, `u.Query()`, `q.Set` and `u.RawQuery = q.Encode()`, open a second `*sql.DB` with it, register its `Close`, run `Schema`, and return a `*PGStore`.
8. Run the tests against a real Postgres. Start one in Docker, run the tests, then stop it:

```bash
docker run --rm -d --name twende-pg -e POSTGRES_USER=gopher -e POSTGRES_PASSWORD=gopher \
  -e POSTGRES_DB=matatu -p 5432:5432 postgres:16-alpine

DATABASE_URL='postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable' \
  go test -race -v ./phase-5-backend/day-062-integration-tests/starter

docker stop twende-pg
```

   `TestPostgresStore` now runs instead of skipping, and every API test uses the database. Afterwards, `docker exec twende-pg psql -U gopher -d matatu -c '\dn'` (before you stop it) lists only `public`: every test schema was dropped.

## The kata: Flaky Tests

Read a log of CI results and find the tests that both passed and failed on the same commit, the flaky ones, ranked by how often. Struct map keys and a two-level sort. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 62
```

## When you're stuck

- **`panic: runtime error: index out of range [0] with length 0`** in the API tests: `NewEnv` isn't loading or seeding the fixtures yet, so `env.Fixtures.Trips` is empty.
- **`reading fixtures: open testdata/fixtures.json: no such file or directory`**: run `go test` on the package (`./phase-5-backend/day-062-integration-tests/starter`), not on a single file. The working directory is then the package folder.
- **`the server still answers after its test ended`**: you closed the server with `defer srv.Close()` inside `NewEnv`, which closes it as soon as `NewEnv` returns, or you never closed it at all. Use `t.Cleanup(srv.Close)`.
- **Every request fails with `connection refused`**: the same `defer` mistake: the server is already closed when the test runs.
- **A failure is reported at `testenv.go:120`** instead of in the test: add `t.Helper()` as the first line of each helper.
- **`sql: unknown driver "pgx"`**: the blank import `_ "github.com/jackc/pgx/v5/stdlib"` registers it. It's in `pgstore.go`; keep it.
- **`ERROR: relation "trips" does not exist`**: the second connection doesn't have `search_path` set, or the `Schema` ran on the admin connection. Run `Schema` on the store's own `*sql.DB`.
- **`ERROR: cannot drop schema ... because other objects depend on it`**: add `CASCADE`.
- **The race test reports `map[-1:10]`**: every request had a network error. Is the server running when the goroutines start?
- **Still stuck?** Open `solution/testenv.go` (and compare `solution/api.go` with yours), read them once, close them, then write your own version from memory.
