# Day 56: Migrations

## The brief

Yesterday you made the tables by piping `schema.sql` into `psql`. That works once. Now the Sacco wants a fare on every route, and the change has to reach your laptop, a teammate's laptop, CI and production, each of which is at a different point. Somebody will run it twice. Somebody will forget. And if the `ALTER TABLE` fails half way on production, you need the database to be exactly where it was, not half changed.

Today you write the tool that does this properly: a migration runner. Each schema change is a numbered pair of SQL files, one to apply it and one to undo it. A `schema_migrations` table remembers which versions each database has. Each migration runs in a transaction, so it lands completely or not at all. The SQL files are embedded in the Go binary, so a deploy ships one file.

```
cmd/migrate/migrations/
  0001_create_routes.up.sql        0001_create_routes.down.sql
  0002_create_trips.up.sql         0002_create_trips.down.sql
  0003_create_bookings.up.sql      0003_create_bookings.down.sql
  0004_add_fare_to_routes.up.sql   0004_add_fare_to_routes.down.sql

$ go run ./phase-5-backend/day-056-migrations/solution/cmd/migrate up
applied     0001_create_routes
applied     0002_create_trips
applied     0003_create_bookings
applied     0004_add_fare_to_routes
$ go run ./phase-5-backend/day-056-migrations/solution/cmd/migrate up
nothing to apply: the database is up to date
$ go run ./phase-5-backend/day-056-migrations/solution/cmd/migrate down 2
rolled back 0004_add_fare_to_routes
rolled back 0003_create_bookings
$ go run ./phase-5-backend/day-056-migrations/solution/cmd/migrate status
applied  0001_create_routes
applied  0002_create_trips
pending  0003_create_bookings
pending  0004_add_fare_to_routes
```

## What you'll use

- **`embed.FS`** (new today). A `//go:embed migrations/*.sql` line above a `var files embed.FS` copies those files into the binary at build time. The path is relative to the Go file and can't use `..`.
- **`io/fs.FS`** is the interface that `embed.FS`, `os.DirFS` and `testing/fstest.MapFS` all satisfy. `Parse` takes an `fs.FS`, so the tests hand it a `MapFS` built in memory and `main` hands it the embedded files. Same code, no disk.
- **Transactions.** `tx, err := conn.BeginTx(ctx, nil)`, then `tx.ExecContext`, then `tx.Commit()`, with `defer tx.Rollback()` straight after `BeginTx`. Rollback after a commit does nothing, so the defer is always safe. Postgres can roll back `CREATE TABLE` and `ALTER TABLE` too. MySQL can't, which is one reason this course uses Postgres.
- **One connection for the whole run.** `db.Conn(ctx)` takes a single connection from the pool. The runner holds a Postgres advisory lock (`pg_advisory_lock`) on it, so if two copies of the app start at once, the second waits instead of applying the same migration twice. The lock belongs to that connection, which is why everything after it must use the same one.
- **Pure logic, split from the database.** `Parse`, `Pending` and `ToRollBack` take values and return values, so the tests check ordering, naming and drift in microseconds with no database. The `Runner` is a thin layer over them, tested by an integration test when `DATABASE_URL` is set.

```go
//go:embed migrations/*.sql
var files embed.FS

ms, err := migrations.Parse(files, "migrations")

test := fstest.MapFS{
	"m/0001_create_routes.up.sql":   {Data: []byte("CREATE TABLE routes (...);")},
	"m/0001_create_routes.down.sql": {Data: []byte("DROP TABLE routes;")},
}
ms, err = migrations.Parse(test, "m")
```

## Why it matters at work

Every team that runs a database has a migration tool, often [goose](https://github.com/pressly/goose), [golang-migrate](https://github.com/golang-migrate/migrate) or [Atlas](https://atlasgo.io/), and they all work the way yours does. Knowing the inside makes the rules stick. Never edit a migration that has already run anywhere else: write a new one, because the old one won't run again. Every up needs a down that has been tried, or the rollback during an incident is the first time anyone runs it. Run migrations before the new code starts, and keep each change safe for the old code that's still running, so add a column with a default first and remove the old one in a later deploy. In review, people read the migration more carefully than the Go code: a migration that locks a big table for minutes takes the whole API down with it.

## Steps

You need the Postgres container from Day 55 only for the last two steps.

1. Open `starter/migrations.go`. Write a helper that splits a file name such as `0002_create_trips.up.sql` into its version, name and direction, or returns an error naming the file. Use `strings.HasSuffix` and `strings.TrimSuffix` for the direction, `strings.Cut` on the first `_`, and check every character of the version and the name yourself.
2. Write `Parse`. `fs.ReadDir(fsys, dir)` lists the folder. Skip directories and anything not ending in `.sql`. For each file, split the name, read it with `fs.ReadFile(fsys, path.Join(dir, name))`, and collect it in a `map[int]*Migration`. Check each rule from the doc comment as you go and at the end. Then copy the map into a slice and sort it by version.
3. Write `Pending`. Put the known versions and the applied versions in maps. An applied version that isn't known is `fmt.Errorf("version %d: %w", v, ErrUnknownVersion)`. Walk `all` in order and keep the ones not applied. If one of them is older than the newest applied version, return an error naming it. Return `[]Migration{}`, not nil, when nothing is pending.
4. Write `ToRollBack`: reject `steps < 1`, sort a copy of `applied` from newest to oldest, and take up to `steps` of them, looking each one up in `all`.
5. Run the unit tests:

```bash
go test ./phase-5-backend/day-056-migrations/starter
```

6. Open `starter/runner.go`. `withLock` and `applied` are written for you. Write `inTx` from its doc comment. Write `Up`: inside `r.withLock`, read the applied versions, get the pending ones, and for each run `m.Up` and `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)` in one `inTx`. Stop at the first error and wrap it with the migration's name. `Down` is the mirror: `ToRollBack`, then `m.Down` and `DELETE FROM schema_migrations WHERE version = $1` in one `inTx`.
7. With Postgres running (see the Day 55 README), run the integration test. It uses its own schema, `day56_test`. It checks that a migration whose second statement fails leaves nothing behind:

```bash
DATABASE_URL="postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable" go test -run Integration -v ./phase-5-backend/day-056-migrations/starter
```

8. Migrate a real database. If you loaded Day 55's `schema.sql` into it, drop those tables first (`docker exec matatu-db psql -U gopher -d matatu -c "DROP TABLE bookings, trips, routes"`), because the migrations create them:

```bash
export DATABASE_URL="postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable"
go run ./phase-5-backend/day-056-migrations/starter/cmd/migrate status
go run ./phase-5-backend/day-056-migrations/starter/cmd/migrate up
go run ./phase-5-backend/day-056-migrations/starter/cmd/migrate down 1
```

## The kata: Migration Order

Numbers put migrations in one line. Some tools let each migration name what it depends on instead, and then the tool has to work out an order. That's a topological sort, one of the graph problems interviews love: Kahn's algorithm, with the smallest name first when there's a choice, and `CYCLE` when the rules loop. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 56
```

## When you're stuck

- **`pattern migrations/*.sql: no matching files found`**: the `//go:embed` path is relative to the Go file's folder. The SQL files must be in `cmd/migrate/migrations/`.
- **`Parse` returns `0010` before `0002`**: you sorted the file names as strings. Sort the migrations by their `Version` number.
- **`open m/0001_create_routes.up.sql: file does not exist`**: `fs.ReadDir` returns names without the folder. Join them with `path.Join(dir, name)`; use `path`, not `filepath`, because an `fs.FS` always uses `/`.
- **`Pending` returns nil and the test wants `[]`**: start with `pending := []Migration{}`.
- **The broken-migration test finds table `half_made`**: the migration's SQL ran outside the transaction. Use `tx.ExecContext`, not `conn.ExecContext`, inside `inTx`.
- **`ERROR: duplicate key value violates unique constraint "schema_migrations_pkey"`**: you inserted the version before running the migration's SQL, or ran `Up` without checking what was applied. Read `applied` first and insert after `m.Up` succeeds.
- **`Up` hangs**: something took a second connection and is waiting on the advisory lock the first one holds. Pass the `conn` from `withLock` everywhere inside it.
- **Still stuck?** Open `solution/migrations.go` and `solution/runner.go`, read them once, close them, then write your own version from memory.
