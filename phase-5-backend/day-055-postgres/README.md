# Day 55: Postgres

## The brief

Every restart of the booking API so far has wiped every booking, because the store was a map in memory. Today the data moves into Postgres, where it survives restarts, deploys and crashes, and where two passengers racing for seat 3 on the 07:30 to Ongata Rongai can't both win: the database's `UNIQUE (trip_id, seat)` rule decides.

The handlers don't change shape. They still talk to a `Store` interface. What's new is `PGStore`, which implements that interface with SQL, plus the plumbing a real database needs: a connection pool, a deadline on every query, and parameters for every value.

```
routes:  id 1, name "125", Railways → Ongata Rongai
trips:   id 1, route 1, departs 2026-11-02 07:30 +03, 14 seats

GET  /routes/1/trips                                   →  200  [{"id":1,"routeId":1,"departs":"2026-11-02T04:30:00Z","seats":14}]
POST /trips/1/bookings  {"seat":3,"passenger":"Achieng","phone":"254712345678"}  →  201  Location: /bookings/1
POST /trips/1/bookings  (seat 3 again)                 →  409  {"error":"that seat is already booked"}
GET  /trips/1           (database stuck for 3 s)       →  503  Retry-After: 1
```

Route 125 runs from Railways to Ongata Rongai (City Shuttle), checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026. Routes change, so confirm with the conductor.

## What you'll use

- **`database/sql`** is Go's standard interface to SQL databases. A driver plugs into it. `_ "github.com/jackc/pgx/v5/stdlib"` registers pgx as the driver named `"pgx"`; the `_` import runs the package's `init` without naming it.
- **`*sql.DB` is a pool, not a connection.** Open it once at startup and share it. `SetMaxOpenConns` caps how many connections you hold, which matters because Postgres has its own limit (100 by default) shared by every copy of your app. `SetConnMaxLifetime` and `SetConnMaxIdleTime` recycle old and unused ones. `sql.Open` doesn't connect; `db.PingContext` does.
- **`pgxpool`** is pgx's own pool, for code that uses pgx directly instead of `database/sql`. It skips a layer and gives you more Postgres features, such as `COPY` and batches. `pgxpool.ParseConfig` reads a DSN without connecting, so you can set `MaxConns` before `pgxpool.NewWithConfig`. Today you configure both; the API runs on `database/sql`.
- **Parameters, never string building.** `$1`, `$2` are placeholders, and the values travel separately from the SQL text, so a passenger called `Robert'); DROP TABLE bookings; --` is stored as a name. `fmt.Sprintf("... WHERE id = %s", id)` is SQL injection, OWASP's A03 (see [reference/owasp-checklist.md](../../reference/owasp-checklist.md)).
- **`QueryRowContext(...).Scan(&a, &b)`** for one row, which returns `sql.ErrNoRows` when there's none. **`QueryContext`** for many: loop with `rows.Next()`, `rows.Scan`, always `defer rows.Close()` to give the connection back, and check `rows.Err()` after the loop. `INSERT ... RETURNING id` gives you the new ID in one round trip.
- **Context timeouts** ([Day 44](../../phase-4-concurrency/day-044-context-cancel/)). `context.WithTimeout(r.Context(), 3*time.Second)` gives every query a deadline. When the client hangs up or the deadline passes, pgx stops waiting, returns an error wrapping `context.DeadlineExceeded` (or `context.Canceled`) and drops that connection, and the pool opens a fresh one later.
- **Postgres error codes.** A broken rule comes back as a `*pgconn.PgError` with a five-character `Code`: `23505` is a unique violation, `23503` a foreign-key violation. Translate them into your own errors in one place, so handlers never import the driver.

```go
var t Trip
err := db.QueryRowContext(ctx,
	`SELECT id, route_id, departs, seats FROM trips WHERE id = $1`, id,
).Scan(&t.ID, &t.RouteID, &t.Departs, &t.Seats)
if errors.Is(err, sql.ErrNoRows) {
	// no such trip
}
```

The config follows [reference/env-and-config.md](../../reference/env-and-config.md): `DATABASE_URL` is required and has no default, because it holds a password.

## Why it matters at work

Most backend incidents are database incidents. A query with no timeout holds a connection while the pool runs dry, and then every request waits, which is why each handler gets a deadline and answers `503` instead of hanging. A pool bigger than the database allows fails as soon as you scale to three copies. String-built SQL is still how most data leaks start, and a reviewer will reject any `Sprintf` near a query on sight. Putting storage behind an interface is what lets the handler tests run in milliseconds with a fake, while one integration test proves the SQL against a real Postgres. That's the split most Go teams use.

## Running Postgres

You need Docker. This starts Postgres 16 with the same user, password and database as the repo's `.env.example`:

```bash
docker run --name matatu-db -e POSTGRES_USER=gopher -e POSTGRES_PASSWORD=gopher -e POSTGRES_DB=matatu -p 5432:5432 -d postgres:16-alpine
docker exec -i matatu-db psql -U gopher -d matatu < phase-5-backend/day-055-postgres/starter/testdata/schema.sql
```

The schema is in `starter/testdata/schema.sql`: `routes`, `trips` and `bookings`, with `bookings` holding `UNIQUE (trip_id, seat)`. `docker stop matatu-db` stops it and `docker start matatu-db` brings it back with your data. `docker rm -f matatu-db` deletes it.

## Steps

`starter/store.go` is written for you: the types and the `Store` interface.

1. In `starter/pg.go`, write `OpenDB`: `sql.Open("pgx", dsn)`, then the four pool settings from the doc comment.
2. Write `PoolConfig`: `pgxpool.ParseConfig`, then set `MaxConns`, `MinConns`, `MaxConnLifetime` and `MaxConnIdleTime`.
3. Write `TranslateError`. `errors.Is` for `sql.ErrNoRows`, then `errors.As` into a `*pgconn.PgError` and switch on its `Code`.
4. Write the five `PGStore` methods. Each is one SQL statement with `$n` placeholders. The inserts end in `RETURNING id` and `Scan` it. Pass every error through `TranslateError`. Store and return times as UTC with `.UTC()`. In `TripsOnRoute`, start with `trips := []Trip{}` so an empty result is `[]`, `ORDER BY departs, id`, and check `rows.Err()`.
5. In `starter/api.go`, write the three handlers. Each makes `ctx, cancel := context.WithTimeout(r.Context(), a.timeout)`, defers `cancel()`, and passes `ctx` to the store. One `fail` helper maps `ErrNotFound` to `404`, `ErrSeatTaken` to `409`, `context.DeadlineExceeded` to `503` with `Retry-After: 1`, and anything else to a logged `500`. The booking handler checks a blank passenger or phone and a seat outside 1 to `trip.Seats` with `422`.
6. Run the tests from the repo root. The integration test is skipped:

```bash
go test ./phase-5-backend/day-055-postgres/starter
```

7. With Postgres running, run the integration test too. It works in its own schema, `day55_test`, so it doesn't touch the tables the app uses:

```bash
DATABASE_URL="postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable" go test -run Integration -v ./phase-5-backend/day-055-postgres/starter
```

8. Run the API and book a seat. Add a route and a trip with `psql` first:

```bash
docker exec matatu-db psql -U gopher -d matatu -c "INSERT INTO routes (name, from_stage, to_stage) VALUES ('125', 'Railways', 'Ongata Rongai'); INSERT INTO trips (route_id, departs, seats) VALUES (1, '2026-11-02 07:30+03', 14);"
DATABASE_URL="postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable" go run ./phase-5-backend/day-055-postgres/starter/cmd/api
curl -i -X POST localhost:8080/trips/1/bookings -d '{"seat":3,"passenger":"Achieng","phone":"254712345678"}'
```

## The kata: Connection Pool

Is a pool of 10 enough? Simulate it: queries arrive, each takes the connection that frees up soonest, and you report when each finishes and the longest anyone waited. It's a classic scheduling problem, and the efficient answer uses a min-heap. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 55
```

## When you're stuck

- **`sql: unknown driver "pgx" (forgotten import?)`**: add `_ "github.com/jackc/pgx/v5/stdlib"` to the imports of `pg.go`.
- **`failed to connect to ... connection refused`**: Postgres isn't running, or not on that port. `docker ps` shows it; `docker start matatu-db` starts it.
- **`ERROR: relation "trips" does not exist (SQLSTATE 42P01)`**: you haven't loaded the schema. Run the `docker exec -i ... < schema.sql` line.
- **`sql: expected 4 destination arguments in Scan, not 3`**: the number of columns in `SELECT` and of pointers in `Scan` must match, in the same order.
- **`ERROR: syntax error at or near "from"`**: `from` is a reserved word. The columns are `from_stage` and `to_stage`.
- **Booking the same seat twice gives `500`**: `TranslateError` didn't find the `*pgconn.PgError`. Use `errors.As` with `var pgErr *pgconn.PgError`, and compare `pgErr.Code` with `"23505"`.
- **The slow-database test hangs**: the handler passed `r.Context()` to the store instead of the context from `context.WithTimeout`.
- **After a while every request hangs**: a `rows` was never closed, so its connection never went back to the pool. Put `defer rows.Close()` straight after the error check.
- **Still stuck?** Open `solution/pg.go` and `solution/api.go`, read them once, close them, then write your own version from memory.
