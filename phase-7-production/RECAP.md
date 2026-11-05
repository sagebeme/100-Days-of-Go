# Phase 7 recap: Production engineering

Days 77 to 84. You learnt what happens after "it works on my laptop": estimating load, caching, background jobs, contracts between services, seeing inside a running system, finding threats before attackers do, handling secrets, and writing it all down before changing code.

## What the phase covered

| Days | Topic |
|---|---|
| 77 | system design: drawing the system, back-of-the-envelope numbers, Little's law, percentiles |
| 78 | caching with Redis: cache-aside, TTLs, invalidating on write, stampede protection |
| 79 | background jobs: backoff with jitter, deduplication, idempotent workers, dead letters, the outbox |
| 80 | gRPC: protobuf contracts, status codes, deadlines, server streaming |
| 81 | observability: structured logs with request IDs, Prometheus metrics, health and readiness, SLOs |
| 82 | threat modelling: data-flow diagrams, trust boundaries, STRIDE, rating risks |
| 83 | secrets: mounted files and vaults, a type that won't print, least privilege, key rotation |
| 84 | the capstone: a design doc, then the hardened service it calls for |

## Key ideas

1. **Estimate before you build.** Requests a second at peak, requests in flight (rate × latency), storage a year, and the one box you can't just copy: usually the primary database.
2. **Percentiles, not averages.** The p99 is the passenger staring at a spinner; averages hide them.
3. **A cache needs a TTL, an invalidation and a fallback.** Delete on write, let the TTL catch what you miss, and treat a cache outage as a miss, not an error. Never trust cached numbers for a decision like a booking.
4. **Work the user doesn't wait for goes in a queue, and runs at least once.** So jobs need an idempotency key, backoff with jitter, a maximum number of attempts and a dead-letter list; the outbox closes the gap between commit and enqueue.
5. **Every call across a network has a deadline.** And a status code the caller can act on: retry `Unavailable`, not `InvalidArgument`.
6. **Logs, metrics and traces answer different questions.** One request (logs with request IDs), how much and how fast (RED metrics labelled by route pattern), and where the time went across services (traces). Never put secrets or high-cardinality labels in them.
7. **STRIDE every box and arrow, and look hardest at the arrows that cross a boundary unauthenticated.** Rate risks by likelihood and impact, and fix the worst first.
8. **Secrets are typed, scoped and rotatable.** A type that redacts itself, a token that reads only its own folder, and two keys at once so rotation doesn't log everyone out.

## Self-test

1. 864 000 daily users make 10 requests each, with a peak five times the average, and each request takes 200 ms. Roughly how many requests are in flight at peak?
2. Why does a write delete the cache key instead of writing the new value into it?
3. An SMS job can be delivered twice. What makes that safe?
4. A metric is labelled with the full request path. What happens on a busy day?
5. The M-Pesa callback is a plain HTTPS POST. Which STRIDE threat is the worst, and what do you check before marking a booking paid?
6. You rotate a signing key and every passenger is logged out. What did the keyring forget?

<details>
<summary>Answers</summary>

1. About 100: 100 requests a second on average, 500 at peak, and 500 × 0.2 s = 100 in flight (Little's law).
2. Two writers can finish in the opposite order to how they started and leave the older value cached. A delete can't be out of order: the next read loads what's true.
3. An idempotent worker: it records the job's key when it succeeds and skips a key it has already done.
4. Every path (`/trips/1`, `/trips/2`, ...) becomes its own time series, and the monitoring system runs out of memory. Label by route pattern.
5. Spoofing: anyone who finds the URL can claim a payment. Confirm the transaction with M-Pesa's own API, check the amount and phone match, accept each receipt once, and restrict the URL.
6. The previous key. Keep it for verifying until the tokens it signed have expired, and sign new tokens with the new one.

</details>

## Build it without steps: Chama Payouts

Your Chama API from the Phase 5 recap is about to send money: at the end of each month it pays one member the pot through an M-Pesa B2C payout, which is slow, can fail, and calls you back later. Treat it as a real launch.

Use only Phases 1 to 7. Use a fake payout gateway in tests; never real credentials.

Done when:

- `DESIGN.md` (one page, the Day 84 template) covers the goals, the numbers for December's peak, the design, at least two alternatives you turned down, and how you'd roll back. Someone else has read it.
- `THREATS.md` has a STRIDE table for the new boxes and arrows, the five worst rated by likelihood and impact, and a fix and an owner for each. The payout callback is one of them.
- A payout is written to an outbox in the same transaction as the decision to pay, and a worker sends it with backoff and jitter, an idempotency key, a maximum number of attempts and a dead-letter list. A test proves a crash between commit and send loses nothing and a redelivery pays nobody twice.
- The member list on the dashboard is cached in Redis with a TTL and invalidated on every write; with Redis stopped, everything still works.
- Every request logs one JSON line with a request ID; `/metrics` has RED metrics by route pattern and a counter of dead-lettered payouts; `/readyz` checks Postgres and Redis with timeouts.
- The gateway keys come from mounted files or env through a type that can't be logged, and the signing key can be rotated without logging anyone out.
- Every outbound call has a deadline, and the service drains cleanly on `SIGTERM`.
