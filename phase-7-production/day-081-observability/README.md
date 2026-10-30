# Day 81: Observability

## The brief

Friday, 5:40 pm. A passenger tweets that Twende "keeps spinning" when she books the 125 to Rongai. Is it everyone or just her? Since when? Which part is slow? Without logs, metrics and health checks the honest answer is "no idea". Today you add all three to the API, so the next time someone asks, you can answer in a minute.

```
GET /trips/46K-0730   X-Request-ID: 7f3a...
  → 200, header X-Request-ID: 7f3a...
  log  {"level":"INFO","msg":"request","request_id":"7f3a...","method":"GET","path":"/trips/46K-0730","status":200,"duration_ms":12}

GET /metrics
  http_requests_total{code="200",method="GET",route="/trips/{id}"} 1832
  http_requests_total{code="500",method="POST",route="/bookings"} 4
  http_request_duration_seconds_bucket{method="POST",route="/bookings",le="0.5"} 512

GET /healthz  → 200 ok
GET /readyz   → 200 {"status":"ok","checks":{"db":"ok","redis":"ok"}}
GET /readyz   → 503 {"status":"fail","checks":{"db":"ok","redis":"fail"}}   (the details are in the log, not here)
```

## What you'll use

**Structured logs with `log/slog`.** One JSON object per event, with named fields, so you can search `status>=500 AND path="/bookings"` instead of reading text. `slog.New(slog.NewJSONHandler(os.Stdout, nil))` makes a logger; `log.LogAttrs(ctx, slog.LevelInfo, "request", slog.Int("status", 200))` writes a line at a level you choose.

- **Request IDs.** You met these in [Day 54](../../phase-5-backend/day-054-middleware/)'s middleware; today they get stricter. Every request gets an ID that appears in every log line it causes and in the `X-Request-ID` response header. The passenger's support ticket quotes it, and you find her exact request. Reuse an ID that the caller sent, so one ID follows a request through several services, but only if it's short and plain: a header with a newline in it could forge log lines.
- **No secrets in logs.** Log `r.URL.Path`, not `r.URL.String()`: query strings carry tokens and reset codes. Never log headers, bodies, passwords or full phone numbers. Logs are copied to many places and kept for months (A09 in [owasp-checklist.md](../../reference/owasp-checklist.md)).

**Metrics with Prometheus.** Prometheus scrapes `/metrics` every few seconds and stores numbers over time. Two kinds today:

- A **counter** only goes up: `http_requests_total`. Prometheus works out the rate.
- A **histogram** counts observations into buckets: how many requests took at most 0.05 s, at most 0.1 s, and so on, plus their sum and count. From the buckets Prometheus estimates p95 and p99 across every server, which you can't do by averaging each server's p95.
- **Labels and cardinality.** Each different set of label values is a separate time series. Label by route **pattern** (`/trips/{id}`), never by real path (`/trips/1`, `/trips/2`, ...), user ID or phone number, or one busy day creates millions of series and topples Prometheus.

```go
requests := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "..."}, []string{"method", "route", "code"})
reg.MustRegister(requests)
requests.WithLabelValues("GET", "/trips/{id}", "200").Inc()
```

**What to measure.** The **RED** method for anything that serves requests: **R**ate, **E**rrors, **D**uration. Today's two metrics give all three. The **USE** method for resources such as CPU, memory, disks and connection pools: **U**tilisation, **S**aturation (work waiting), **E**rrors.

**Health and readiness.** Kubernetes and load balancers ask two different questions:

- **Liveness** (`/healthz`): is the process stuck? If this fails, the server is restarted. Check nothing external: if the database is down, restarting every API server makes things worse.
- **Readiness** (`/readyz`): should this server get traffic right now? Check the dependencies it needs, each with a timeout. During shutdown, fail readiness first (`Drain`) so the load balancer stops sending new requests while the ones in flight finish ([Day 84](../day-084-hardened-service/)). Answer `ok` or `fail` only; the error text goes in the log, because it can name hosts, users and versions.

**Tracing, in concept.** Logs tell you what happened in one service; metrics tell you how much. A **trace** follows one request across services: the booking API calls the seat service ([Day 80](../day-080-grpc/)), which calls Postgres. Each step is a **span** with a start time, a duration and a parent, and they share one trace ID passed between services in the `traceparent` header (the W3C Trace Context standard). Tools draw the spans as a waterfall, so "the 2 s is all in the seat service's database call" is visible at a glance. **OpenTelemetry** is the vendor-neutral standard and Go SDK for producing traces, metrics and logs. Today's request ID is the simplest form of the same idea; we don't add the OpenTelemetry dependency here.

**SLOs and error budgets.** A **service level objective** is a target you promise yourselves, written on what users feel: "99.9% of booking requests succeed in under 500 ms, over 30 days". The **error budget** is what's left: 0.1% of 30 days is 43.2 minutes of failure. While budget remains, ship features. When it's spent, the team stops and works on reliability. Alert on the budget burning fast, not on every slow request.

**Testing metrics.** The tests use the client library's own `testutil` package: `testutil.GatherAndCompare(reg, expected, names...)` reads the registry the way Prometheus scrapes `/metrics` and diffs the named metrics against the text you expect, and `testutil.GatherAndCount` counts series. `NewMetrics` takes a clock function so the durations in the test are exact.

## Why it matters at work

On-call engineers live in these three tools. The first question in any incident is "what changed, and since when?", and a dashboard of RED metrics per route answers it. The second is "show me one failing request", and a request ID in a structured log answers it. Code review catches the expensive mistakes: a metric labelled by user ID (the monitoring bill explodes), a log line with a token in it (a security incident), a liveness check that pings the database (a small outage becomes a big one). Teams that agree SLOs argue less about whether to ship or fix, because the error budget decides.

## Steps

1. Open `starter/observe.go`. `RandomID`, the `recorder`, `Metrics`, `NewMetrics`, `Readiness`, `NewReadiness`, `Add` and `Drain` are given.
2. Write `validID`: 1 to 64 characters, each a letter, digit or hyphen.
3. Write `RequestID`: take a valid header or call `newID`, set the response header, and pass on `r.WithContext(context.WithValue(...))` with the unexported `ctxKey{}`. Write `RequestIDFrom` with a two-value type assertion ([Day 16](../../phase-2-modelling/day-016-payment-methods/)).
4. Write `Logging`. Wrap `w` in a `recorder`, call `next`, then log once with `LogAttrs`: Error level for 500 and above, otherwise Info. Use `rec.code()`.
5. Write `Instrument`. Time the call with `m.now()` before and after, then `Inc` the counter with the status as a string (`strconv.Itoa`) and `Observe` the seconds in the histogram.
6. Write `Live`.
7. Write `Readiness.ServeHTTP`. Copy the checks under the lock, then run each with `context.WithTimeout(r.Context(), rd.timeout)`. Log failures with the check name; put only `"ok"` or `"fail"` in the body. 503 if anything failed or `Drain` was called.
8. Run the tests from the repo root:

```bash
go test -race ./phase-7-production/day-081-observability/starter
```

9. Run it and look:

```bash
go run ./phase-7-production/day-081-observability/starter/cmd/observed
curl -i localhost:8080/trips
curl -s localhost:8080/metrics | grep http_
```

## The kata: Running Percentile

Answer "what's the p95 right now?" over a stream of latencies, in O(log n) per sample, with two heaps. It's the running-median interview question, generalised to any percentile. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 81
```

## When you're stuck

- **`RequestIDFrom = "", want "generated-1"`**: the handler got the old request. Pass `r.WithContext(ctx)` to `next`, not `r`.
- **`newline rejected: RequestIDFrom = "abc\nlevel=ERROR..."`**: you trusted the header. Validate it with `validID`.
- **`log status = <nil>`** or `log has no duration_ms`: the attribute names must match exactly: `request_id`, `method`, `path`, `status`, `duration_ms`.
- **`the log contains "SECRET-TOKEN"`**: you logged `r.URL.String()` or `r.RequestURI`. Use `r.URL.Path`.
- **`status 0`** in the log: the handler never called `WriteHeader`. Use `rec.code()`, which treats that as 200.
- **`3 request series, want 2`**, with a diff showing `route="/trips/1"` where `route="/trips/{id}"` was expected: you used `r.URL.Path` as the route label. Use the `route` argument.
- **`panic: duplicate metrics collector registration attempted`**: a test registered the metrics twice on one registry. Each test makes a fresh `prometheus.NewRegistry()`; don't use the default registry in tests.
- **`a hanging check hung the readiness probe`**: give each check its own `context.WithTimeout` and call `cancel()` after it.
- **`readyz body ... leaks the error text`**: write `"fail"` in the body and the error in the log.
- **Still stuck?** Open `solution/observe.go`, read it once, close it, then write your own version from memory.
