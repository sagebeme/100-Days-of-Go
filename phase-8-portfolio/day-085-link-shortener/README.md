# Day 85: Link Shortener

## The brief

Wanjiru runs Kiungo, a small agency in Nairobi that sends SMS campaigns for dukas and schools. An SMS is 160 characters, and a link like `https://duka.co.ke/sale?item=sufuria&size=5L&utm_source=sms` eats half of it. She wants her own shortener at `kiungo.ke`, with a count of how many people tapped each link so she can show clients the campaign worked. Last month a competitor's shortener was used to send phishing links, so she's firm: no `javascript:` links, and nothing that points into a private network.

```
POST /links  {"url":"https://duka.co.ke/sale?item=sufuria"}  →  201  Location: https://kiungo.ke/aB3xY9q
                                                                    {"code":"aB3xY9q","short":"https://kiungo.ke/aB3xY9q","url":"https://duka.co.ke/sale?item=sufuria"}
GET  /aB3xY9q                                                →  302  Location: https://duka.co.ke/sale?item=sufuria
GET  /links/aB3xY9q/stats                                    →  200  {"code":"aB3xY9q","url":"...","hits":3,"created":"2026-10-24T09:30:00Z"}
GET  /zzzzzzz                                                →  404  {"error":"not found"}
POST /links  {"url":"javascript:alert(1)"}                   →  422  {"error":"invalid url: scheme must be http or https"}
POST /links  {"url":"http://169.254.169.254/"}               →  422  {"error":"invalid url: must not point at a private or local address"}
POST /links  (the 11th in a second from one IP)              →  429  Retry-After: 1
```

Acceptance criteria:

- Codes are 7 base62 characters from `crypto/rand`. A collision never overwrites a link: generate a new code and try again, a few times, then fail cleanly.
- The redirect is a `302` and counts a hit. Hits are counted correctly under concurrent requests.
- Only `http` and `https` URLs with a public host are accepted, checked on the literal IP or on every address the name resolves to.
- `POST /links` is rate limited per client IP and answers `429` with a whole-second `Retry-After`.
- The store sits behind an interface, so Postgres can replace the in-memory map later.

## What you'll use

- `net/http` routing with methods and wildcards: [Day 49](../../phase-5-backend/day-049-hello-server/), [Day 50](../../phase-5-backend/day-050-http-methods/).
- Status codes and one JSON error shape: [Day 51](../../phase-5-backend/day-051-status-codes/), [reference/http.md](../../reference/http.md).
- Interfaces for the store and the limiter: [Day 16](../../phase-2-modelling/day-016-payment-methods/).
- A mutex around a map: [Day 45](../../phase-4-concurrency/day-045-mutexes/).
- Token buckets with `golang.org/x/time/rate`: [Day 47](../../phase-4-concurrency/day-047-rate-limiting/).
- SSRF thinking from [Day 60](../../phase-5-backend/day-060-owasp-top-10/) and A10 in [reference/owasp-checklist.md](../../reference/owasp-checklist.md).
- New today: `net/netip`, whose `Addr` has `IsPrivate`, `IsLoopback` and `IsLinkLocalUnicast` built in.

```go
addr, _ := netip.ParseAddr("169.254.169.254")
addr.IsLinkLocalUnicast() // true: the cloud metadata address
netip.MustParseAddr("::ffff:10.0.0.1").Unmap().IsPrivate() // true
```

## Why it matters at work

A shortener is an open door: anyone can make your domain point anywhere. That's why real ones are blocklisted by spam filters within weeks of launch. Refusing `javascript:` and `data:` stops stored XSS in any page that renders the long URL as a link. Refusing private targets matters as soon as you add a link preview, because then your server fetches the URL, and that is SSRF. Know the limit of today's check: a name can resolve to a public IP now and `127.0.0.1` a minute later (DNS rebinding). A server that fetches must check again at dial time, as the OWASP checklist says.

## Steps

Requirements. Meet them in any order. The tests in `starter/` check each one.

Must-haves:

1. `Encode` and `Decode` convert between `uint64` and base62. `Decode` refuses empty strings, bad characters and overflow with `ErrBadCode`.
2. `ValidCode` accepts 1 to 16 base62 characters. `RandomCode` returns `CodeLength` characters from `crypto/rand`.
3. `MemoryStore` implements `Store` and is safe for concurrent use. `Create` never overwrites.
4. `PublicAddr` refuses loopback, private, link-local, carrier-grade NAT, unspecified, multicast and broadcast addresses, including IPv4 written as IPv6.
5. `ValidateURL` follows its doc comment. Every refusal is a `*URLError`.
6. `Server.Shorten` retries on `ErrCodeTaken` up to `MaxTries`, then returns `ErrNoFreeCode`.
7. `New` fills in defaults and registers the three routes. Bad JSON is `400`, a body over `MaxBodyBytes` is `413`, a refused URL is `422`, unknown codes are `404`, and any other error is a logged `500` with a generic message.
8. `IPLimiter` keeps one `rate.Limiter` per key. `POST /links` checks it first.
9. `starter/cmd/kiungo` runs the server with timeouts set on `http.Server`.
10. Run the tests from the repo root, with the race detector too:

```bash
go test ./phase-8-portfolio/day-085-link-shortener/starter
go test -race ./phase-8-portfolio/day-085-link-shortener/starter
go run ./phase-8-portfolio/day-085-link-shortener/starter/cmd/kiungo
curl -i -d '{"url":"https://go.dev"}' localhost:8080/links
```

Stretch goals:

- A Postgres store using pgx, with a unique constraint on the code. `ErrCodeTaken` becomes the unique-violation error. Run it for real with `docker run --rm -e POSTGRES_PASSWORD=dev -p 5432:5432 postgres:17`.
- Custom codes (`{"url":..., "code":"sale24"}`) with a list of reserved words like `links` and `admin`.
- Expiring links, and evicting idle buckets from `IPLimiter` so its map doesn't grow forever.
- A preview endpoint that fetches the page title, with the SSRF check at dial time in a `net.Dialer.Control` func.

## The kata: TinyURL

The interview version of today: hand out base62 codes from a counter, give a repeated URL its old code, and answer lookups. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 85
```

## When you're stuck

- **`Encode(0)` returns `""`**: the loop never runs for 0. Handle it first.
- **`Decode("LygHa16AHYG")` returns a small number**: it overflowed and wrapped. Before `n*62 + d`, check `n > (math.MaxUint64 - d) / 62`.
- **`javascript:alert(1)` passes**: you checked for a host but not the scheme. Allow `http` and `https` only, and refuse everything else. A blocklist always misses one.
- **`::ffff:127.0.0.1` passes**: that's IPv4 inside IPv6. Call `Unmap()` before the checks.
- **`two-faced.example` passes**: you checked the first address only. A name can have many; refuse if any one is private.
- **The race detector reports a data race in `Hit`**: reading, adding and writing back the count must all happen under the same lock.
- **`GET /links` gives 404, not 405**: that's expected. `GET /{code}` matches it with the code `links`. Reserve the word if you add custom codes.
- **The rate limit test passes alone but every client shares one bucket**: your key includes the port. Use `net.SplitHostPort(r.RemoteAddr)`. Never key on `X-Forwarded-For` unless a proxy you control sets it.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- Move it to its own repo with a README: what it does, the three endpoints with `curl` examples, and how to run it.
- Add a short demo GIF of `curl -i` creating a link and following it, or a screenshot of the stats JSON.
- Deploy it (Fly.io, Render or a small VPS) behind HTTPS and put the live link at the top of the README.
- Write "What I'd do next": the Postgres store, link previews with dial-time SSRF checks, abuse reporting.
- Add GitHub Actions running `go vet`, `go test -race` and `govulncheck`, and show the CI badge.
