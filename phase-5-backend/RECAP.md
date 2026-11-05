# Phase 5 recap: Backend engineering

Days 49 to 64. You built a JSON API the way teams build them at work: the right methods and status codes, input you don't trust, Postgres with migrations, passwords and tokens, the OWASP Top 10, tests through the real router, and a container that CI builds and a platform runs.

## What the phase covered

| Days | Topic |
|---|---|
| 49 to 51 | `net/http`, handlers and `ServeMux`; methods, safety and idempotency; choosing status codes, one error shape |
| 52 to 54 | JSON in and out, content types; validation and body limits; middleware for logging, recovery and request IDs |
| 55 to 56 | Postgres with pgx and connection pools; versioned migrations and rolling back |
| 57 to 59 | hashing passwords; sessions, JWTs and refresh rotation; roles, 401 against 403, checks in one place |
| 60 to 61 | the OWASP Top 10 in Go: injection, access control, SSRF; CORS, preflights and security headers |
| 62 to 63 | integration tests with `httptest` and a real database; a small Docker image, GitHub Actions and secrets in CI |
| 64 | the capstone: the Matatu Booking API, with seats, bookings and payments, deployed |

## Key ideas

1. **Status codes are the API's grammar.** 400 for bad syntax, 422 for valid JSON with wrong values, 401 for "who are you?", 403 for "not you", 404 for missing (or not yours), 409 for a conflict, 5xx only for your own failures.
2. **Never trust the request.** Limit the body, refuse unknown fields, validate every value on the server, and use placeholders in SQL, never string building.
3. **Authorisation is object-level.** "May passengers read bookings?" is the role check; "is this *your* booking?" is the one that's usually missing. Answer other people's things with the same 404 as missing ones.
4. **Passwords are hashed slowly; tokens are short-lived.** bcrypt or argon2 for passwords, a short access token, and a refresh token that rotates.
5. **Middleware order matters.** Recovery, request IDs, logging and CORS go outside everything that can answer early, so a 401 or 429 still has its headers and its log line.
6. **CORS protects the browser's user, not your server.** Echo allowed origins exactly, add `Vary: Origin`, never `*` with credentials, and don't trust `X-Forwarded-For` unless your own proxy sets it.
7. **Test through the router.** A real `httptest.Server` catches the wiring bugs unit tests miss, and a contract test keeps the fake store honest.
8. **Liveness restarts; readiness routes traffic.** `/healthz` checks nothing external, `/readyz` checks dependencies and fails while draining, and `SIGTERM` leads to a graceful `Shutdown`.

## Self-test

1. A client `POST`s `{"seats": "two"}`, then `{"seats": 0}`. Which status for each, and why are they different?
2. Otieno, a passenger, asks for `GET /bookings/7`, which belongs to Wanjiku. What does he get, and why not 403?
3. Why is `fmt.Sprintf("SELECT * FROM users WHERE phone = '%s'", phone)` dangerous, and what do you write instead?
4. A 429 from the rate limiter reaches the browser without `Access-Control-Allow-Origin`. What's wrong with the middleware order?
5. Why does `/healthz` not ping the database?
6. A pull request from a fork runs your CI. Why does it get no secrets?

<details>
<summary>Answers</summary>

1. `400`: the JSON doesn't match the type, so the request can't even be read. `422`: it's well-formed but the value breaks a rule.
2. `404`, exactly like a booking that doesn't exist. A 403 would confirm booking 7 exists and belongs to someone.
3. A phone like `' OR '1'='1` changes the query: SQL injection. Use a placeholder, `WHERE phone = $1`, and pass the value separately.
4. CORS runs inside the limiter, so the limiter answers before CORS adds its headers and the browser hides the response. Put CORS outside it.
5. If the database goes down, a failing liveness check makes the platform restart every API server, which fixes nothing. That's readiness's job.
6. A fork's code is written by anyone, and with secrets it could print them or send them away.

</details>

## Build it without steps: Chama API

Turn the treasurer's tool into a service. A chama has members, a treasurer and an admin; members record contributions and see their own history, the treasurer sees everyone's, and only the admin changes roles.

Use only Phases 1 to 5.

Done when:

- `POST /register`, `POST /login`, `GET /me/contributions`, `POST /contributions`, `GET /contributions` (treasurer and admin) and `PUT /members/{phone}/role` (admin) exist, with the status codes a reviewer expects for every failure, and one JSON error shape.
- Passwords are hashed; logging in gives a short access token and a rotating refresh token in an `HttpOnly` cookie.
- Roles are read from the store on each request, and a member who asks for someone else's contribution gets 404.
- Data lives in Postgres behind a `Store` interface, with versioned migrations that roll forward and back. An in-memory store passes the same contract tests.
- Bodies over 1 MB get 413, unknown fields 400, bad values 422 with the field named.
- Every response has the security headers; CORS allows only the origins in config, with credentials; logins are rate limited per IP.
- Integration tests run through the real router against a real Postgres schema per test.
- A multi-stage Dockerfile builds a small non-root image; CI runs `gofmt`, `go vet` and `go test -race` on every push with read-only permissions; `/healthz` and `/readyz` work, and `SIGTERM` drains cleanly.
