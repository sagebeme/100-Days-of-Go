# Day 54: Middleware

## The brief

Friday, 17:40. A passenger calls: "Nimelipa lakini sina seat!" She paid, but has no booking. You open the logs and find nothing useful. No line per request, no way to tell her request apart from the thousand others, and one handler panicked at 17:38 and printed a stack trace to a terminal nobody was watching.

Today every request gets the same treatment before any handler runs. Each one gets an ID that comes back in a header, so support can ask for it. Each one leaves exactly one structured log line with its status and timing. A panic becomes a clean `500` and a logged stack trace instead of a dropped connection. And the logs never hold a token or a full phone number, because logs get read by more people than your database does.

```
POST /bookings  X-Request-ID: trace-42  {"passenger":"Achieng","phone":"254712345678"}
  →  201  X-Request-ID: trace-42

log:
{"time":"...","level":"INFO","msg":"booking created","request_id":"trace-42","phone":"2547*****678"}
{"time":"...","level":"INFO","msg":"request","method":"POST","path":"/bookings","status":201,"bytes":20,"duration":184000,"request_id":"trace-42"}

GET /trips/1   (the handler panics)
  →  500  {"error":"something went wrong"}
{"level":"ERROR","msg":"panic","value":"assignment to entry in nil map","stack":"goroutine 7 [running]:...","request_id":"9f86d081..."}
{"level":"ERROR","msg":"request","method":"GET","path":"/trips/1","status":500,...}
```

## What you'll use

- **Middleware** is a function from handler to handler: `func(next http.Handler) http.Handler`. It does something, calls `next.ServeHTTP(w, r)`, and can do more after. It wraps every route at once without touching any of them.
- **Order matters.** `Chain(h, a, b)` is `a(b(h))`: `a` sees the request first and the response last. RequestID goes outside so everything else can read the ID. Logger goes outside Recover so it logs the `500` that Recover writes.
- **Wrapping the `ResponseWriter`.** Go never tells a middleware which status the handler sent. A small struct that embeds `http.ResponseWriter` and overrides `WriteHeader` and `Write` records it. Embedding (Day 15) passes every other method through, and an `Unwrap` method lets `http.ResponseController` reach the real writer.
- **`context.WithValue`** carries the request ID to the handlers. Use an unexported key type, `type ctxKey struct{}`, so no other package can clash with it. Context values are for request-scoped data like this, not for passing a function's normal arguments.
- **`log/slog`** (new today) writes structured logs: a message plus key-value attributes, as JSON with `slog.NewJSONHandler`. Log tools can then search `status=500` or `request_id=trace-42` instead of grepping text.
- **`recover`** (new today) inside a deferred function stops a panic and returns its value. `net/http` already recovers panics so one request can't crash the server, but it only prints to stderr and drops the connection. Yours sends a proper answer and logs where the request ID is.

```go
func Timer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "path", r.URL.Path, "duration", time.Since(start))
	})
}
```

## Why it matters at work

Middleware is where a team puts the rules every endpoint must follow, so nobody has to remember them per handler: logging, request IDs, recovery, later auth (Day 58), rate limits (Day 60), CORS and security headers (Day 61). Request IDs are how support, front-end and back-end engineers talk about the same failed request, and in a system of several services the same ID is passed along so one search finds the whole story. The other half is what *not* to log. Logs are copied to log services, shared in chats and kept for months, so OWASP's A09 item says never log secrets (see [reference/owasp-checklist.md](../../reference/owasp-checklist.md)). Kenya's Data Protection Act 2019 treats a phone number as personal data. Log the path, not the query string, which can carry tokens. Never log headers, which carry `Authorization` and cookies. Mask phone numbers. A reviewer who sees `slog.Info("req", "headers", r.Header)` will stop the merge.

## Steps

1. Open `starter/middleware.go`. Write `Chain`: loop over `mws` from the last to the first, wrapping `h` each time, then return it.
2. Write `RequestIDFrom`: `ctx.Value(ctxKey{})` with a type assertion to `string`. The comma-ok form returns `""` when there's nothing there.
3. Write `RequestID`. Read `X-Request-ID`. Keep it only if it's 1 to 64 characters, each a letter, a digit, `-`, `_` or `.`. That rule also stops someone sending a newline to forge a fake log line. Otherwise make one: 16 bytes from `crypto/rand.Read`, then `hex.EncodeToString`. Set the response header, then call `next` with `r.WithContext(context.WithValue(r.Context(), ctxKey{}, id))`.
4. Write a `recorder` struct that embeds `http.ResponseWriter` and has `status` and `bytes` fields. `WriteHeader` saves the first status. `Write` sets status to 200 if it's still 0, then adds up the bytes. Add `Unwrap() http.ResponseWriter`.
5. Write `Logger`. Note the start time, call `next` with a recorder, then log `"request"` with `logger.LogAttrs` and the six attributes from the doc comment. A status still at 0 means 200. Use `slog.LevelError` for 5xx and `slog.LevelInfo` otherwise. Log `r.URL.Path`, never `r.URL.String()` or `r.RequestURI`, which include the query.
6. Write `Recover`. Defer a function that calls `recover()`. If it's nil, return. If it's `http.ErrAbortHandler`, panic again. Otherwise log `"panic"` at Error level with `value`, `stack` (from `debug.Stack()`) and `request_id`, then send a 500 with `{"error":"something went wrong"}` as JSON.
7. Write `MaskPhone` from its doc comment.
8. Open `starter/app.go` and finish `New` with `Chain(mux, RequestID, Logger(logger), Recover(logger))`.
9. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-054-middleware/starter
```

## The kata: Latency Percentiles

The duration in every log line is only useful when you sum it up, and averages lie: ninety fast requests and ten that hang average out to "fine". The kata computes p50, p90 and p99 from a list of request durations, with the nearest-rank method and whole-number rounding. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 54
```

## When you're stuck

- **The chain runs in reverse**: you wrapped from the first middleware to the last, so the first ended up innermost. Loop from `len(mws)-1` down to 0.
- **The logged status is 0**: the handler called `Write` without `WriteHeader`, or wrote nothing at all. Treat 0 as 200 in both places.
- **The client gets 200 but the log says 404**: your recorder's `WriteHeader` saved the status but didn't call `rec.ResponseWriter.WriteHeader(status)`.
- **`request_id` is empty in the log**: RequestID must be outside Logger in the chain, and you must call `next` with the new request from `r.WithContext`, not the old `r`.
- **The test sees the query string's token in the log**: log `r.URL.Path`.
- **The panic test crashes the whole test run**: `recover()` only works when called directly inside a deferred function. `defer recover()` doesn't work; `defer func() { p := recover() ... }()` does.
- **Only one log line after a panic**: Logger is inside Recover in your chain, so the panic skipped Logger's log call. Put Recover last, closest to the handler.
- **Still stuck?** Open `solution/middleware.go` and `solution/app.go`, read them once, close them, then write your own version from memory.
