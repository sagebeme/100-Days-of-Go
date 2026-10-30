# Day 61: CORS and Security Headers

## The brief

Twende's (made-up) React front end will live at `https://twende.example` and call the API at `https://api.twende.example`. Those are different origins, so the browser blocks the calls until the API says, with CORS headers, that this one front end may make them. While you're in the middleware, add the security headers the tester asked for, refuse cross-site form posts, and stop the bot that tried 4,000 passwords against `/auth/login` last night.

```
OPTIONS /bookings  Origin: https://twende.example  Access-Control-Request-Method: POST
  →  204  Access-Control-Allow-Origin: https://twende.example
          Access-Control-Allow-Credentials: true
          Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE
          Access-Control-Allow-Headers: Authorization, Content-Type, Idempotency-Key
          Access-Control-Max-Age: 600
          Vary: Origin, Access-Control-Request-Method, Access-Control-Request-Headers

OPTIONS /bookings  Origin: https://evil.example  ...         →  403
POST /bookings     Origin: https://evil.example  Sec-Fetch-Site: cross-site  →  403
GET  /trips        (any response)  →  Content-Security-Policy: default-src 'none'; frame-ancestors 'none'
                                      Strict-Transport-Security: max-age=63072000; includeSubDomains
                                      X-Content-Type-Options: nosniff
                                      Referrer-Policy: no-referrer
POST /auth/login   the 3rd try in a second from 41.90.1.1  →  429  Retry-After: 30
```

## What you'll use

**Origins.** An origin is scheme, host and port: `https://twende.example` and `https://api.twende.example` are different, and so are `http://localhost:5173` and `http://localhost:8080`. A browser lets a page read responses only from its own origin, unless the other server opts in with CORS. CORS doesn't protect your API from curl or a bot: they ignore it. It tells *browsers* which other sites may read your responses.

**The CORS headers.**

- `Access-Control-Allow-Origin`: echo back the request's `Origin` if it's on your list. Never reflect any origin you're sent, and never `*` with credentials (browsers refuse that combination anyway).
- `Access-Control-Allow-Credentials: true`: the browser may send cookies, and the page may read the answer.
- `Vary: Origin`: the answer depends on the `Origin` header, so a cache mustn't give one site's answer to another.
- `Access-Control-Expose-Headers`: by default a page can read only a few simple response headers. List `Location` and `Retry-After` so the app can use them.

**Preflight.** Before a "non-simple" request (JSON bodies and an `Authorization` header both count), the browser sends an `OPTIONS` request with `Access-Control-Request-Method` and asks permission. Answer `204` with the allowed methods and headers and a `Max-Age` (seconds the browser may cache the answer; Chrome caps it at 2 hours). The real request comes after.

**Security headers** for a JSON API:

| Header | Value | Why |
|---|---|---|
| `Content-Security-Policy` | `default-src 'none'; frame-ancestors 'none'` | an API serves no scripts, styles or images, so allow none, and refuse to be framed (clickjacking). `frame-ancestors` replaces the old `X-Frame-Options` |
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains` | the browser uses only HTTPS for two years. Browsers ignore it over plain HTTP, so it's safe to send everywhere |
| `X-Content-Type-Options` | `nosniff` | the browser trusts your `Content-Type` instead of guessing |
| `Referrer-Policy` | `no-referrer` | your URLs (which may hold IDs) don't leak to other sites |

**`http.CrossOriginProtection`** (Go 1.25 and later) stops CSRF: a page on `evil.example` that makes the passenger's browser `POST` to your API with their cookies. It reads the browser's `Sec-Fetch-Site` header (or compares `Origin` with `Host`) and refuses cross-site `POST`, `PUT`, `PATCH` and `DELETE`. Requests without those headers, such as curl or a phone app, pass. `AddTrustedOrigin` lets your own front end through.

```go
p := http.NewCrossOriginProtection()
p.AddTrustedOrigin("https://twende.example")
handler := p.Handler(mux)
```

**A token bucket per IP** with `golang.org/x/time/rate` ([Day 47](../../phase-4-concurrency/day-047-rate-limiting/) built one by hand). `rate.NewLimiter(rate.Every(10*time.Second), 3)` allows 3 at once, then one every 10 seconds. `AllowN(now, 1)` takes the time as an argument, which makes it testable with a fake clock. Keep one limiter per IP in a map behind a mutex, and evict idle ones.

**The client IP** is the host part of `r.RemoteAddr`, split off with `net.SplitHostPort`. Don't trust `X-Forwarded-For`: anyone can send it. Behind a proxy you control (Render and Fly.io both put one in front of your app), read the header that proxy sets, and only that one.

## Why it matters at work

CORS bugs come in two kinds, and you'll see both in review. Too tight, and the front-end team is blocked, someone "fixes" it with `Access-Control-Allow-Origin: *` or by echoing back any origin, and now any website can read logged-in users' data. Too loose from day one is the same hole. Allowed origins belong in config (`CORS_ALLOWED_ORIGINS`, see [env-and-config.md](../../reference/env-and-config.md)), so staging can allow `http://localhost:5173` and production can't. Security headers and rate limits are A05 and A07 in [owasp-checklist.md](../../reference/owasp-checklist.md); scanners and pen-testers check for them first, and a missing rate limit on login is how password-guessing bots win. Middleware order matters too: if the rate limiter runs before CORS, the browser can't even read the `429`, and the front end shows a mystery network error. See [http.md](../../reference/http.md) for `OPTIONS`, `403` and `429`.

## Steps

1. Open `starter/headers.go`. The JSON helpers, the constants, the `IPLimiter` struct and `clientIP` are given.
2. Write `ParseOrigins`. Split on commas, trim, lower-case, skip blanks. Parse each with `url.Parse` and accept only `http` or `https` with a host and no path, query, fragment or user. Collect every bad one and `errors.Join` them. `LoadOrigins` reads `CORS_ALLOWED_ORIGINS` with it and names the variable in its error.
3. Write `CORS`. Add `Vary: Origin` first, on every response. No `Origin` header: call `next`. Not allowed: 403 for a preflight, otherwise call `next` with no CORS headers. Allowed: set `Allow-Origin` (the origin itself) and `Allow-Credentials`. For a preflight also add the two other `Vary` values, the allow headers and `Max-Age`, answer 204 and stop. Otherwise set `Expose-Headers` and call `next`.
4. Write `SecurityHeaders`: set the four headers, then call `next`.
5. Write `CrossOrigin` with `http.NewCrossOriginProtection`, `AddTrustedOrigin` for each origin (return its error), and `SetDenyHandler` with a JSON 403.
6. Write the limiter: `Allow` (create the IP's limiter on first sight, update `lastSeen`, `AllowN(l.Now(), 1)`), `Evict` (delete IPs idle for `idle` or longer), `RetryAfter` (seconds per token, rounded up, at least 1) and `Middleware` (429 with `Retry-After`).
7. Write `Wrap` in the order its doc comment gives.
8. Run the tests from the repo root, with the race detector, because the limiter's map is shared:

```bash
go test -race ./phase-5-backend/day-061-cors-and-security-headers/starter
```

## The kata: Sliding Window Limiter

Replay a request log through a sliding-log rate limiter: allow a request only if the IP had fewer than `k` allowed requests in the last `w` seconds. It's the other classic limiter design, and a common interview question. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 61
```

## When you're stuck

- **The browser console says `No 'Access-Control-Allow-Origin' header is present`** although your handler works in curl: the request went through a path that skipped `CORS`, often a 401, 404 or 429 written by middleware that runs before it. Put `CORS` outside everything that can answer early.
- **`The value of the 'Access-Control-Allow-Origin' header must not be the wildcard '*' when the request's credentials mode is 'include'`**: echo the allowed origin instead of `*`.
- **The preflight reaches your real handler**: return right after `w.WriteHeader(http.StatusNoContent)`.
- **`Vary = ["Origin" "Origin"]`** or a missing Vary: use `h.Add` for Vary (it can have several values) and `h.Set` for the others, and add `Origin` once at the top.
- **`ParseOrigins` accepts `https://twende.example/`**: the trailing slash makes `u.Path` `"/"`. Browsers never send a path in `Origin`, so it can never match; refuse it so the typo is caught at startup.
- **`AddTrustedOrigin` returns an error for a good-looking origin**: it wants exactly `scheme://host[:port]`, no trailing slash.
- **`50 concurrent requests, 7 allowed`**: two goroutines both created a limiter for a new IP. Hold the mutex across the lookup, the create and `AllowN`.
- **Retry-After is `0`**: `rate.Every(30*time.Second)` is 1/30 tokens a second. Seconds per token is `1 / float64(l.Rate)`; round up with `math.Ceil`.
- **Still stuck?** Open `solution/headers.go`, read it once, close it, then write your own version from memory.
