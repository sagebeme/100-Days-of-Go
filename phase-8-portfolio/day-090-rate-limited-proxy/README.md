# Day 90: Rate-Limited Proxy

## The brief

Halima is the only backend developer at a small Mombasa fintech. Their payments service runs on a private network, and the mobile app talks to it through a gateway. Last month one buggy app release retried in a tight loop and took the payments service down for an hour. She wants a reverse proxy in front of it: it forwards only to the services she lists, gives every client its own rate limit, strips headers that shouldn't cross the boundary, tells the service who the real client is, and gives up on a slow upstream instead of hanging.

```
GET /mpesa/stk/push?amount=100        →  forwarded to http://10.0.0.5:9000/api/stk/push?amount=100
                                         upstream sees X-Forwarded-For: 41.90.64.7
GET /mpesa/x  (X-Forwarded-For: 6.6.6.6 from a phone)
                                      →  upstream sees X-Forwarded-For: 41.90.64.7   (the fake is dropped)
GET /mpesa/x  (from our LB 10.0.0.2, X-Forwarded-For: 41.90.64.7)
                                      →  upstream sees X-Forwarded-For: 41.90.64.7, 10.0.0.2
GET /admin/users                      →  404 {"error":"no such route"}
GET /mpesa/../admin                   →  400 {"error":"bad path"}
GET /mpesa/x  (11th in a second)      →  429 Retry-After: 1
GET /slow/x   (upstream takes 30 s)   →  504 {"error":"upstream timed out"}
GET /dead/x   (upstream is down)      →  502 {"error":"upstream unavailable"}
```

Acceptance criteria:

- Only configured prefixes are forwarded. Paths that aren't already clean are refused before matching.
- Each client gets its own token bucket. The client is the connection's IP, unless the connection is from a trusted load balancer, in which case it's read from `X-Forwarded-For`.
- Hop-by-hop headers, and any header named in `Connection`, never reach the upstream or the client.
- `X-Forwarded-For` is set correctly: client-supplied values are dropped unless they came through a trusted proxy.
- Timeouts everywhere: dial, TLS, response headers, and the whole request. Upstream errors don't leak internal addresses.

## What you'll use

- Token buckets with `golang.org/x/time/rate`: [Day 47](../../phase-4-concurrency/day-047-rate-limiting/). A mutex around the map of buckets: [Day 45](../../phase-4-concurrency/day-045-mutexes/).
- Contexts with deadlines: [Day 44](../../phase-4-concurrency/day-044-context-cancel/).
- Status codes `429`, `502` and `504`: [Day 51](../../phase-5-backend/day-051-status-codes/) and [reference/http.md](../../reference/http.md).
- `net/netip` prefixes for the trusted networks, as on [Day 85](../day-085-link-shortener/).
- New today: `httputil.ReverseProxy` with a `Rewrite` func. It hands you `pr.In` (what the client sent) and `pr.Out` (what you'll send), with hop-by-hop and `X-Forwarded-*` headers already removed from `Out`.

```go
rp := &httputil.ReverseProxy{
	Rewrite: func(pr *httputil.ProxyRequest) {
		pr.Out.URL.Scheme, pr.Out.URL.Host = "http", "10.0.0.5:9000"
		pr.SetXForwarded() // X-Forwarded-For: the client's IP
	},
}
```

## Why it matters at work

Almost every production service sits behind a proxy, and three mistakes show up in incident reviews again and again. Trusting `X-Forwarded-For` from anyone lets an attacker pick a fresh IP for every request and walk past your rate limit, or appear to come from an internal address. A proxy without timeouts holds a goroutine and a connection for every request to a stuck upstream until it runs out. And a proxy that forwards to whatever host the request names is an SSRF hole into your private network. An allowlist, a trusted-proxy list and explicit timeouts are what reviewers check for.

## Steps

Requirements. Meet them in any order. The tests in `starter/` check each one.

Must-haves:

1. `StripHopByHop` removes the listed headers and anything named in `Connection`.
2. `ClientIP` follows its doc comment. Read `X-Forwarded-For` from the right, skipping trusted addresses, and stop at anything that isn't an IP.
3. `New` validates the config, sorts routes longest prefix first, and builds a `ReverseProxy` with a `Rewrite` func, an `ErrorHandler` and a `Transport` with dial, TLS and response-header timeouts. Set the transport's `Proxy` to `nil` so `HTTP_PROXY` in the environment can't redirect traffic.
4. `ServeHTTP` refuses unclean paths with `400`, unknown prefixes with `404`, and empty buckets with `429` and `Retry-After`. Then it adds the timeout to the context, limits the body, and forwards.
5. `Rewrite` replaces the prefix with the upstream's path, keeps the query, sets `Host` to the upstream, strips hop-by-hop headers, keeps the inbound `X-Forwarded-For` only from a trusted peer, then calls `SetXForwarded`.
6. `ErrorHandler`: `504` for a deadline, `502` for anything else, `413` for a body over the limit. Log the real error; send a generic one.
7. `cmd/proxy` runs it from flags, with timeouts on the `http.Server` too.
8. Run the tests, with the race detector too:

```bash
go test ./phase-8-portfolio/day-090-rate-limited-proxy/starter
go test -race ./phase-8-portfolio/day-090-rate-limited-proxy/starter
```

Stretch goals:

- Evict buckets that haven't been used for 10 minutes, so a flood of new IPs can't grow the map forever.
- Different limits per route, and a stricter one for `POST`.
- Prometheus metrics: requests, 429s and upstream latency per route.
- Retries for idempotent methods only (`GET`, `HEAD`), with a budget, and a circuit breaker per upstream.
- Share the limits between several proxy instances with Redis ([Day 78](../../phase-7-production/day-078-caching/)).

## The kata: Token Buckets

Simulate per-client token buckets from a log of request times, in whole-number arithmetic, and print which requests were allowed. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 90
```

## When you're stuck

- **`panic: interface conversion: interface {} is nil, not proxy.route`** in `Rewrite`: `Rewrite` reads the matched route from the request's context. Put it there in `ServeHTTP` with `context.WithValue` before calling `rp.ServeHTTP`.
- **The upstream sees `X-Forwarded-For: 6.6.6.6, 41.90.64.7`**: you copied the inbound header for every client. Copy it only when the peer is a trusted proxy.
- **The upstream sees no `X-Forwarded-For` at all**: with `Rewrite`, the proxy removes it from `Out` first. Call `pr.SetXForwarded()`.
- **`/mpesa/../admin` reaches the upstream**: you matched the prefix on the raw path. Compare the path with `path.Clean` of itself first, and refuse it if they differ.
- **The slow-upstream test hangs or takes 10 seconds**: the timeout isn't on the request. Wrap `r.Context()` with `context.WithTimeout` and pass `r.WithContext(ctx)` on.
- **`502` instead of `504` for the slow upstream**: check `errors.Is(err, context.DeadlineExceeded)` in the error handler.
- **The rate-limit test gives a new bucket for each request**: your key includes the port. Use `ClientIP`.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- Give it its own repo with a README: an architecture diagram (client, proxy, upstreams), the config, and the security choices (allowlist, trusted proxies, header hygiene) with one sentence of why for each.
- Add a GIF of `hey` or `curl` in a loop hitting the limit and getting `429`s.
- Run it on a small VPS or Fly.io in front of a demo API, and link it.
- Write "What I'd do next": bucket eviction, per-route limits, metrics, Redis-backed limits.
- Add GitHub Actions running `go test -race` and `govulncheck`, with the CI badge.
