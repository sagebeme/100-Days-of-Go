# Day 84: Hardened Service

## The brief

The Twende booking API in `starter/` works on a laptop. It lists trips and sells seats. It also reads its config without checking it, has no timeouts, lets anyone book, accepts a body of any size, logs nothing, has no metrics or health checks, falls over on a panic, and drops requests mid-flight when it's stopped. The SACCO wants it live before the December rush.

This capstone is how a senior engineer would do it: write a short design doc first, get it reviewed, then make exactly the changes it calls for. Everything from Phase 7 comes together here, along with config (Day 21), middleware (Day 54), passwords and keys (Days 57 and 83) and security headers (Day 61).

```
TWENDE_RATE=fast go run ./…/cmd/twende
  →  bad config: TWENDE_API_KEY: is required
                 TWENDE_RATE: must be a number above 0, got "fast"

GET  /trips  (11th request this second from one IP)  →  429  Retry-After: 1
POST /bookings  no key                               →  401  WWW-Authenticate: Bearer
POST /bookings  2 MiB body                           →  413
GET  /trips  database hung                           →  503 after 2 s, not never
GET  /readyz while shutting down                     →  503, and the request in flight still gets its 200
every response                                       →  X-Request-ID, nosniff, DENY, CSP, HSTS, no-store
every request                                        →  one JSON log line; GET /metrics counts it
```

## What you'll use

Nothing new: today is about choosing and combining. Two habits matter more than any one API:

- **Write it down first.** A one-page design doc makes you decide what "done" means before you're deep in code, and gives a reviewer something to disagree with while changes are cheap.
- **One requirement, one test.** Each requirement in the doc maps to a test in `starter/service_test.go`. When you're done, the test names read like the doc's checklist.

**The design doc template.** Copy it into `starter/DESIGN.md` and fill it in before you touch the code. Keep it to a page.

```markdown
# Hardening the Twende booking API

## Context
What the service does, who calls it, and why it needs hardening now (two or three sentences).

## Goals
- The requirements you will meet, each one testable.
## Non-goals
- What you are deliberately not doing yet (for example: user accounts, Postgres, TLS termination).

## Current state
The risks in today's code, each with the STRIDE letter it falls under (Day 82).

## Design
For each goal: the change, where it goes (which middleware, which function), and the
value you chose (timeouts, rate, body size) with a one-line reason.

## Alternatives considered
At least one per hard choice, and why you didn't take it (for example: rate limiting
here or at the load balancer; per-IP or per-API-key limits).

## Rollout and operations
Config to set in production, what to watch on the dashboard (Day 81), what an alert
should fire on, and how to roll back.

## Open questions
What you'd ask a reviewer.
```

## Why it matters at work

This is the checklist many teams run before a service takes real traffic, sometimes called a production readiness review. Each item exists because an outage taught someone it was needed: the server with no `ReadHeaderTimeout` held open by slow clients, the deploy that started with a typo in a rate and limited everyone to zero, the panic that killed the process, the rolling restart that cut off every payment in flight. A reviewer reads the design doc first and the diff second, and the doc is what gets the change approved.

## Steps

Requirements. Meet them in any order. `starter/store.go` (the `Store` interface and `MemStore`) stays as it is; change `config.go` and `server.go` however your design says, keeping the exported names and signatures the tests use.

1. Write `starter/DESIGN.md` from the template, and have someone review it (a classmate, a colleague, or yourself after a night's sleep).
2. **Config is validated.** `LoadConfig` rejects a missing or short (under 32 characters) `TWENDE_API_KEY`, a `TWENDE_ENV` other than `dev` or `prod`, a `TWENDE_RATE` that isn't a number above 0, a `TWENDE_BURST` below 1, and a `TWENDE_HANDLER_TIMEOUT` or `TWENDE_SHUTDOWN_GRACE` that isn't a positive duration. It reports every problem at once, naming each variable, never including the key. Defaults are in the comments on `Config`.
3. **Secrets don't leak.** Neither `fmt` (`%v`, `%+v`, `%#v`, `%s`), `json.Marshal` nor `slog` with the text or JSON handler shows the API key, whether you log `cfg` or `cfg.APIKey`.
4. **Timeouts everywhere.** `HTTPServer` sets `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` and `IdleTimeout` from the config and a `MaxHeaderBytes` limit. Every store call has a deadline of `HandlerTimeout`; a store that doesn't answer in time gives `503` with a plain JSON error and the real error in the log.
5. **Rate limiting.** Each client IP (from `RemoteAddr`, never `X-Forwarded-For`) gets `TWENDE_RATE` requests a second with a burst of `TWENDE_BURST`. Over the limit: `429`, a whole-number `Retry-After` and a JSON error. `/healthz`, `/readyz` and `/metrics` are never limited.
6. **Security headers** on every response, errors and 404s included: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`, `Referrer-Policy: no-referrer`, `Strict-Transport-Security: max-age=63072000; includeSubDomains`, `Cache-Control: no-store`.
7. **Bookings check the caller and the input.** `POST /bookings` needs `Authorization: Bearer <TWENDE_API_KEY>`, compared in constant time; otherwise `401` with `WWW-Authenticate: Bearer`. The body is at most `MaxBodyBytes` (`413` above it) and is one JSON object with no unknown fields (`400`). A missing `tripId` or a phone that isn't `07` or `01` plus eight digits is `422`, and so is a seat that doesn't exist. `404` for an unknown trip, `409` for a taken seat, `201` with `Location: /bookings/<id>` on success.
8. **Structured logs.** Exactly one JSON line per request with `msg` `request`, `request_id` (also sent as the `X-Request-ID` header), `method`, `path`, `status` and `duration_ms`. No keys, query strings or phone numbers.
9. **Panics are contained.** A panic in a handler gives `500` without the panic text, is logged at Error level with the request ID, and the server keeps serving.
10. **Metrics.** `GET /metrics` serves the registry passed to `NewServer`, with `http_requests_total{method, route, code}` labelled by route pattern.
11. **Health checks.** `/healthz` answers `200` `ok\n` and checks nothing. `/readyz` answers `200` when `store.Ping` succeeds within `HandlerTimeout`, `503` otherwise, with no internal details in the body.
12. **Graceful shutdown.** When `Run`'s context is cancelled, `/readyz` fails at once, no new connections are accepted, requests in flight finish (up to `ShutdownGrace`), and `Run` returns `nil`.
13. Run the tests from the repo root, then run it:

```bash
go test -race ./phase-7-production/day-084-hardened-service/starter
TWENDE_API_KEY=$(openssl rand -hex 32) go run ./phase-7-production/day-084-hardened-service/starter/cmd/twende
```

## The kata: Log Aggregation

Summarise a weekend of request logs: requests, errors and p95 per route, and the worst minute. It's the report you'd build from requirement 8's logs. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 84
```

## When you're stuck

- **Where does everything go?** Middleware, outermost first: request ID, logging, panic recovery, security headers, then rate limiting on the API routes only. Two `ServeMux`es make the last part easy: one for the API routes, wrapped in the limiter, mounted at `/` on another that also holds the probes and `/metrics`.
- **`error ... doesn't name TWENDE_RATE`** in the "everything at once" case: collect problems in a slice and return `errors.Join(errs...)` at the end instead of returning on the first.
- **`logging the config with the json handler shows the API key`**: the JSON handler encodes structs with `encoding/json`, which ignores `String` and `LogValue` on fields. Give `Secret` a `MarshalJSON`, or give `Config` a `LogValue` that lists safe fields.
- **`GET /trips with a hung store never answered`**: pass `context.WithTimeout(r.Context(), cfg.HandlerTimeout)` to the store, not `r.Context()`.
- **`/healthz from a limited client: status 429`**: the limiter wraps the whole handler. Wrap only the API routes.
- **The 429 or 404 is missing security headers**: set them in a middleware outside the limiter and the mux, before calling `next`.
- **`a 2 MiB body: 400, want 413`**: wrap the body in `http.MaxBytesReader` and check for `*http.MaxBytesError` with `errors.As`.
- **`logged 2 lines for one request`**: only the logging middleware logs requests; handlers log only real problems.
- **`the panic and its request ID weren't logged`**: recovery must run inside the request-ID middleware, and inside logging so the 500 is logged too.
- **`Run returned ... while a request was still in flight`**: `srv.Close` drops connections. Use `srv.Shutdown` with a `ShutdownGrace` timeout, and set your draining flag before calling it.
- **`/readyz still 200 after shutdown began`**: keep an `atomic.Bool` on the server, set it in `Run` when the context is done, and check it first in `/readyz`.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory. Then compare your `DESIGN.md` with what you built.
