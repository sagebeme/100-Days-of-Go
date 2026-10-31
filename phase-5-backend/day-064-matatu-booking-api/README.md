# Day 64: Matatu Booking API

## The brief

This is the Phase 5 capstone. Build the whole matatu booking API, from passenger sign-up to a paid seat, and deploy it. Nobody gets two bookings for one seat, nobody reads someone else's ticket, and the API survives a bot, a browser on another origin and a redeploy in the middle of the evening rush. There's no walkthrough today: you get the requirements, the tests and some hints. Everything you need is from Days 49 to 63.

The API starts with five real Nairobi routes, checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026. Routes change, so confirm with the conductor.

| Route | Operator | Boards at (CBD) | Goes to | Est. fare |
|---|---|---|---|---|
| 46K | Citi Hoppa | Kencom / Ambassadeur | Kawangware | Ksh100 |
| 111 | MCA Sacco | Railways | Ngong | Ksh100 |
| 125 | City Shuttle | Railways | Ongata Rongai | Ksh100 |
| 45 | Nicco Travellers | Ronald Ngala Street | Githurai | Ksh100 |
| 58 | Buruburu 58 Travellers Sacco | Ambassadeur | Buruburu | Ksh150 |

```
POST /auth/register {"phone":"+254 712 345 678","name":"Wanjiku","password":"nimechoka na jam ya thika road"}
  →  201 {"phone":"0712345678","name":"Wanjiku","role":"passenger"}
POST /auth/login {"phone":"0712345678","password":"..."}
  →  200 {"accessToken":"eyJhbGciOiJIUzI1NiIs...","tokenType":"Bearer","expiresIn":900}

GET  /routes                                →  200 [{"id":1,"number":"46K","name":"Citi Hoppa",...,"fare":100}, ...]
POST /trips {"routeId":5,"departs":"2026-03-02T06:30:00Z","seats":14}     (admin)  →  201, Location: /trips/6
GET  /trips/6                               →  200 {"id":6,...,"fare":150,"seatsLeft":14,"takenSeats":[]}

POST /bookings {"tripId":6,"seat":3}        (Wanjiku)  →  201 {"id":7,...,"fare":150,"paid":false}, Location: /bookings/7
POST /bookings {"tripId":6,"seat":3}        (Otieno)   →  409 {"error":"seat 3 is already booked"}
GET  /bookings/7                            (Otieno)   →  404 {"error":"no such booking"}
POST /bookings/7/pay                        (Wanjiku)  →  200 {"id":7,...,"paid":true}
GET  /me/bookings                           (Wanjiku)  →  200 [{"id":7,...}]
PUT  /users/0733000222/role {"role":"conductor"}  (admin)  →  200
```

## What you'll use

All of Phase 5, plus one new idea: a `Payer` interface. A real one would start an M-Pesa STK push through Safaricom's Daraja API; `FakePayer` stands in for it in tests and in development. Handlers that call the outside world through an interface are handlers you can test.

## Why it matters at work

This is the shape of most backend work: a small API whose correctness lives in the rules (one seat, one booking; your ticket, your eyes only), and whose quality lives in everything around them (validation, status codes, auth, headers, logs, health checks, config, shutdown). A reviewer reading a service like this checks each of the things the requirements list, and so does an incident. The tests here run the whole API over real HTTP with an in-memory store, which is how you'll test services at work, with a Postgres version in CI ([Day 62](../day-062-integration-tests/)).

## Steps

The starter gives you the types (`models.go`), the `Store` interface and `MemStore` with empty methods, `SeedRoutes`, the config helpers, the JSON helpers and `decode`, and empty `NewServer`, `SetDraining` and `Handler`. Add files and unexported functions as you like. Meet these requirements in any order.

**Data and config**

1. `MemStore` implements every `Store` method as its doc comment says, safe for concurrent use. `CreateBooking` checks the seat and saves the booking in one step, so two requests for the same seat can never both win.
2. `LoadConfig` reads every variable in its doc comment, applies the defaults, and reports every problem at once without printing secret values. `NormalizePhone` accepts `07…`, `01…`, `254…` and `+254…` and returns the `07…` form.

**Endpoints**

3. `GET /healthz` is always 200. `GET /readyz` is 200 when `Store.Ping` works, 503 while draining.
4. `POST /auth/register` takes `phone`, `name` (1 to 50 characters) and `password` (15 to 64 characters, per NIST SP 800-63B as on [Day 57](../day-057-passwords/)), and hashes the password with bcrypt at `cfg.BcryptCost`. It answers 201 with the user (never the hash), 409 for a phone that has an account, 422 for bad values and 400 for broken JSON or unknown fields. Phones in `cfg.AdminPhones` get the admin role; everyone else is a passenger.
5. `POST /auth/login` answers 200 with an HS256 JWT (`sub` is the phone, plus `iss`, `aud`, `iat`, `exp` after `cfg.AccessTTL`), `tokenType` `Bearer` and `expiresIn` in seconds. A wrong password and an unknown phone get the same 401.
6. `GET /routes`, `GET /trips` (optionally `?route=<id>`, 400 if that isn't a number) and `GET /trips/{id}` are public. A trip shows `fare`, `seatsLeft` and `takenSeats`.
7. `POST /routes` and `POST /trips` are admin only. A trip needs an existing route (422), a departure after `cfg.Now()` (422) and 1 to 33 seats (422).
8. `POST /bookings` takes only `tripId` and `seat`. The name and phone come from the logged-in user, the fare from the route. 404 for an unknown trip, 409 for a trip that has left, 422 for a seat outside the trip, 409 for a taken seat, 201 with `Location` otherwise.
9. `GET /bookings/{id}`: the passenger who made it, any conductor and any admin. Everyone else gets the same 404 as for a booking that doesn't exist. `GET /me/bookings` lists your own, `[]` when there are none.
10. `POST /bookings/{id}/pay`: only the passenger who made it. Charge `Fare` through the `Payer` with a reference like `MAT7`. 200 with the paid booking, 409 if it's already paid, 502 with a friendly message if the payer fails (and the booking stays unpaid).
11. `PUT /users/{phone}/role` with `{"role": ...}` is admin only: 200, 404 for an unknown user, 422 for an unknown role.

**Around the endpoints**

12. Roles live in one permission table, checked in one middleware: no user is 401 with `WWW-Authenticate: Bearer`, the wrong role is 403. Conductors read bookings and do nothing else. Read the role from the store on each request, not from the token, so a role change applies at once.
13. A token that's present but bad (wrong signature, `alg: none`, expired) is 401 with `WWW-Authenticate: Bearer error="invalid_token"`, on any route.
14. Every response has the four security headers from [Day 61](../day-061-cors-and-security-headers/). CORS allows only `cfg.AllowedOrigins`, with credentials, `Vary: Origin`, and preflights.
15. A token bucket per client IP (`cfg.RatePerSecond`, `cfg.RateBurst`, time from `cfg.Now`) answers 429 with `Retry-After`. With `cfg.ClientIPHeader` set, the IP comes from that header (the last entry, for `X-Forwarded-For`) instead of `RemoteAddr`. The 429 still carries the CORS and security headers.
16. Bodies are limited to 1 MB (413). Panics become a 500 with a generic message. Each request writes one log line with method, path, status and duration, and never a token or password.
17. `cmd/api` seeds the routes, sets server timeouts, and shuts down gracefully on SIGINT or SIGTERM: `SetDraining`, then `Shutdown` within `cfg.ShutdownTimeout`.
18. A multi-stage `Dockerfile` builds it onto distroless as a non-root user.
19. Run the tests from the repo root, then run it for real:

```bash
go test -race ./phase-5-backend/day-064-matatu-booking-api/starter
JWT_SECRET="$(openssl rand -base64 32)" ADMIN_PHONES=0700000001 \
  go run ./phase-5-backend/day-064-matatu-booking-api/starter/cmd/api
```

20. Deploy it, with the guide below.

## Deploy it

The store is in memory, so every restart or redeploy starts again from the seeded routes. That's fine for a demo. For real use, write a `PGStore` behind the same `Store` interface ([Day 55](../day-055-postgres/)'s queries, [Day 56](../day-056-migrations/)'s migrations, Day 62's test pattern) and set `DATABASE_URL`.

First, give the API a repository of its own: copy the solution (or your starter) into a new repo with its own `go.mod`, put the `Dockerfile` at the root with its build path changed to `./cmd/api`, and add [Day 63](../day-063-docker-and-ci/)'s `ci.yml` as `.github/workflows/api.yml`. Make a secret with `openssl rand -base64 32`, and keep it out of git.

### Render

1. On [render.com](https://render.com), create a **Web Service** from your GitHub repo and choose the **Docker** runtime. Render builds the `Dockerfile` on every push to your main branch.
2. Under **Environment**, add `JWT_SECRET`, `ADMIN_PHONES` (your number), `CORS_ALLOWED_ORIGINS` (your front end's origin, once you have one in Phase 6) and `CLIENT_IP_HEADER` (see the next step). Render's default `PORT` is 10000 and it also detects the port you bind to; the API reads `PORT`, so leave it to Render.
3. **Check which header holds the client's IP before you trust it.** Render doesn't document this. Its staff say the first `X-Forwarded-For` entry is the client, but Render adds to whatever `X-Forwarded-For` the client sent instead of replacing it, so a client can fake that entry. Render sits behind Cloudflare, which sets `True-Client-IP`. On your first deploy, log the request headers once, call the API from your phone, and compare them with your phone's real IP. Then set `CLIENT_IP_HEADER` to the header that matches and that a client can't fake. If you get this wrong, either every passenger shares one rate limit or anyone can dodge it.
4. Set the **Health Check Path** to `/readyz`. Render waits for it before switching traffic to a new deploy, and sends `SIGTERM` to the old instance, which your graceful shutdown handles.
5. To deploy from CI instead of on every push, turn off auto-deploy, copy the service's **Deploy Hook** URL into a GitHub secret called `RENDER_DEPLOY_HOOK_URL`, and Day 63's `deploy` job calls it after the tests and the image build pass.
6. Free instances sleep when idle, so the first request after a quiet spell is slow, and the in-memory data is gone.

### Fly.io

1. Install `flyctl` and run `fly launch` in the repo. It finds the `Dockerfile` and writes a `fly.toml`. Say no to a database for now.
2. Check `fly.toml` has the right port and a health check on `/readyz`:

```toml
# Fly's default is 5s, shorter than the API's 10s SHUTDOWN_TIMEOUT.
kill_timeout = "15s"

[http_service]
  internal_port = 8080
  force_https = true

  [[http_service.checks]]
    method = "GET"
    path = "/readyz"
    interval = "15s"
    timeout = "2s"
    grace_period = "5s"
```

3. Set the secrets, which Fly passes to the app as env variables: `fly secrets set JWT_SECRET="$(openssl rand -base64 32)" ADMIN_PHONES=0700000001 CLIENT_IP_HEADER=Fly-Client-IP`.
4. `fly deploy`. Fly stops old machines with `SIGINT` by default, which your shutdown also handles, and kills them after `kill_timeout`, which is why it's 15s above.
5. From CI, run `flyctl deploy --remote-only` with a token from `fly tokens create deploy` stored as a GitHub secret named `FLY_API_TOKEN`.

Then try it: register, log in, list routes, book a seat, pay, and look at the response headers with `curl -i`.

## The kata: Double Bookings

Replay a log of bookings and cancellations under the one-seat-one-booking rule: who got the seat, who was refused, which cancellations were real. Two maps that point at each other. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 64
```

## When you're stuck

- **Every test fails with `register 0700000001: 404`**: `Handler` still returns an empty mux. Register the routes, then wrap them.
- **`6 passengers booked the last seat at once: map[201:2 409:4]`**, more than one 201: you check the seat and save the booking in two separate locked sections, or check in the handler. Do both inside `CreateBooking`, under one lock. In Postgres, a `UNIQUE (trip_id, seat)` constraint does the same job.
- **Otieno gets 403 for Wanjiku's booking instead of 404**: `require(ReadBooking)` passes (passengers may read bookings), so the handler must do the object-level check after loading it, and answer exactly like a missing booking.
- **A role change doesn't take effect**: you read the role from the token's claims. Look the user up in the store in the auth middleware.
- **`token has invalid claims: token is expired` straight after login in a test**: you used `time.Now` somewhere. Every time check uses `cfg.Now`, including `jwt.WithTimeFunc`.
- **`creating a trip in the past` passes**: compare with `cfg.Now()`, not `time.Now()`. The tests' clock is 05:00 on 2 March 2026.
- **The 429 has no `Access-Control-Allow-Origin`**: the rate limiter runs outside CORS. The order in `Handler`'s doc comment matters.
- **`a body over 1 MB: 400, want 413`**: check for `*http.MaxBytesError` with `errors.As` before the generic JSON error. The given `decode` does it; use it.
- **`GET /me/bookings` returns `null`**: a nil slice encodes as `null`. Start from `[]Booking{}`.
- **On Render or Fly, everyone hits the rate limit at once**: they all share the proxy's IP. Set `CLIENT_IP_HEADER`. Never read that header when the app isn't behind a proxy that sets it, because then the client chooses its own IP.
- **Stretch goals:** a `PGStore`; refresh tokens with rotation from [Day 58](../day-058-sessions-and-jwts/); an `Idempotency-Key` on `POST /bookings/{id}/pay`, so a retry over a bad connection can't charge twice; evicting idle IPs from the limiter (Day 61's `Evict`); booking cancellation.
- **Still stuck?** Open `solution/`, read the file for the part you're on once, close it, then write your own version from memory.
