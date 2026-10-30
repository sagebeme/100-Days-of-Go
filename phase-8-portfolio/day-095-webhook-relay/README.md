# Day 95: Webhook Relay

## The brief

Baraka's team at KopaPesa (a made-up lender) gets a webhook from their payment provider every time a customer repays. Today three services need it: the ledger, the SMS sender and the analytics job. The provider can call only one URL. Last month someone found that URL and posted a fake "Ksh 50,000 received", and when the ledger was down for ten minutes, the repayments in that window were lost. They want one relay in front: it checks every webhook is really from the provider, ignores repeats, and forwards each event to every service until it gets through.

```
POST /webhooks  Webhook-Signature: t=1792832400,v1=5f2c…   {"id":"evt_001",...}
                                        →  202  {"status":"queued"}
same request again                      →  200  {"status":"duplicate"}
same request, 10 minutes later          →  401  {"error":"invalid signature"}
body changed, old signature             →  401  {"error":"invalid signature"}

relay → ledger   500   (wait 1s)
relay → ledger   502   (wait 2s)
relay → ledger   200   →  Result{Attempts: 3, Err: nil}
relay → sms      404   →  Result{Attempts: 1, Err: "answered 404"}   (no point retrying)
```

**Acceptance criteria**

- A webhook is accepted only with a valid HMAC-SHA256 signature over its timestamp and body, compared in constant time, and only if the timestamp is within 5 minutes of now.
- An event ID seen in the last 24 hours is acknowledged but not forwarded again.
- Each event goes to every subscriber, signed with that subscriber's own secret, retried with exponential backoff on network errors, 408, 429 and 5xx.
- The clock, the sleep and the HTTP client are injected, so every test runs instantly and never touches the network.

## What you'll use

- Sentinel errors and `errors.Is`, from [Day 17: Failed Transactions](../../phase-2-modelling/day-017-failed-transactions/).
- `context` cancellation inside a retry loop, from [Day 44: Context](../../phase-4-concurrency/day-044-context-cancel/).
- Mutexes for the dedupe map, from [Day 45](../../phase-4-concurrency/day-045-mutexes/).
- Retries and idempotent workers, from [Day 79: Background Jobs](../../phase-7-production/day-079-background-jobs/).
- Body limits and status codes from [Day 53](../../phase-5-backend/day-053-validation/) and [http.md](../../reference/http.md). Security reasoning from [owasp-checklist.md](../../reference/owasp-checklist.md).
- New today: `crypto/hmac`. `hmac.New(sha256.New, secret)` gives a hash that only someone with the secret can produce, and `hmac.Equal` compares two MACs in constant time, so an attacker can't learn the right bytes one at a time by timing your `==`.

```go
h := hmac.New(sha256.New, secret)
h.Write([]byte("1792832400." + string(body)))
want := h.Sum(nil)
hmac.Equal(got, want) // never bytes.Equal or == for secrets
```

## Steps

Requirements. The tests in `starter/` check every must-have.

**Must-haves**

1. `Sign` and `Verify`, in the format of `SignatureHeader`. Sign `"<unix seconds>.<body>"`, not the body alone, so a fresh timestamp can't be glued onto an old signature. Accept any matching `v1`, which lets the provider rotate secrets without downtime.
2. Check the signature before the timestamp, and the timestamp both ways: a request from 6 minutes in the future is as suspicious as one from 6 minutes ago.
3. `Deduper`: remember IDs for a TTL and drop the expired ones, so memory stays bounded. `Forget` undoes a `Seen`.
4. `ServeHTTP`: the status codes in its doc comment. Read the body once with `http.MaxBytesReader`, verify the raw bytes, and only then parse JSON. Only remember an ID once the request is accepted, so a forged or refused copy can't block the real one.
5. Answer fast and deliver later: queue the event on a buffered channel, return `202`, and let `Run` do the slow work. If the queue is full, return `503` with `Retry-After` and let the provider retry.
6. `Deliver`: sign each forwarded request with the subscriber's secret, set `Webhook-Id`, and retry as the brief says, with `Backoff` doubling up to `MaxDelay`. Close every response body. Stop at once when `ctx` is cancelled.
7. `New` fills in the defaults from `Config`'s comments. Nothing reads the real clock or sleeps in tests.
8. Run the tests, with the race detector too:

```bash
go test ./phase-8-portfolio/day-095-webhook-relay/starter
go test -race ./phase-8-portfolio/day-095-webhook-relay/starter
```

**Stretch goals**

- A `cmd/relay` main that reads `WEBHOOK_SECRET` and the subscribers from env ([Day 21](../../phase-2-modelling/day-021-config-loader/)), runs `Run` in a goroutine and shuts down cleanly on Ctrl+C.
- Jitter: wait a random time between 0 and `Backoff(n)` ("full jitter") so a thousand retries don't land on a recovering service in the same second. Inject the random source.
- Honour a subscriber's `Retry-After` header on `429` and `503`.
- A dead-letter list for results that never got through, and a way to replay them.
- Keep the queue and the seen IDs in Redis ([Day 78](../../phase-7-production/day-078-caching/)) so a restart loses nothing.

## The kata: Replay Window

Given event IDs with timestamps, print `ACCEPT` or `DUPLICATE` for each, where a duplicate is the same ID accepted less than `W` seconds ago. It's LeetCode's "logger rate limiter", and it is your Deduper without the pruning. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 95
```

## When you're stuck

- **Every signature is `ErrBadSignature`, even your own**: you're signing different bytes from the ones you verify. Verify the raw body exactly as read, before any JSON decoding, and build the signed string the same way in both places: `strconv.FormatInt(t, 10) + "." + body`.
- **`encoding/hex: invalid byte`**: that's a malformed header, not a bad signature. Return `ErrMalformedSignature`; the handler answers `401` for both.
- **The forged-copy test fails: the real `evt_003` comes back `duplicate`**: you called `Seen` before `Verify`. Verify first.
- **Backoff goes negative for attempt 60**: `base << 59` overflowed. Double in a loop and stop as soon as you reach `max`.
- **`TestRunForwardsToEverySubscriber` hangs**: `Run` isn't reading the queue, or `ServeHTTP` isn't writing to it. `Run` needs a `select` on `ctx.Done()` and the queue channel, in a loop.
- **The race detector flags the dedupe map**: the handler runs on many goroutines at once. Lock around every read and write.
- **Still stuck?** Open `solution/signature.go` and `solution/relay.go`, read them once, close them, then write your own version from memory.

## Make it portfolio-ready

- **README**: a diagram (provider → relay → three services), the signature format with a worked example you can check with `openssl dgst -sha256 -hmac`, and the delivery guarantee in plain words: "at least once; subscribers dedupe by `Webhook-Id`".
- **Demo**: a script that signs a request with `curl` and shows `202`, then `200 duplicate`, then `401` for a tampered body. Record it with asciinema.
- **Deploy**: ship `cmd/relay` in a small Docker image ([Day 63](../../phase-5-backend/day-063-docker-and-ci/)) to a free tier, with two test subscribers on [webhook.site](https://webhook.site). Put the URL in the README, and the secrets only in the host's env settings.
- **Tests and CI**: `go test -race ./...` in GitHub Actions and the badge in the README. Add a fuzz test for `Verify` that checks it never panics on any header.
- **What you'd do next**: be honest about the gaps. The queue is in memory, so a crash loses queued events; the dedupe window resets on restart; one slow subscriber delays the others. Say how you'd fix each: a durable queue, Redis for seen IDs, a worker pool per subscriber ([Day 46](../../phase-4-concurrency/day-046-worker-pools/)).
