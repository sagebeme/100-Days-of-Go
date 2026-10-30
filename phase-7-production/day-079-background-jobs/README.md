# Day 79: Background Jobs

## The brief

When a passenger books seat 3 on the 46K, Twende sends an SMS: "Seat 3, 46K Kencom to Kawangware, 07:30." The SMS gateway takes up to two seconds and sometimes times out. The passenger shouldn't wait for it, and a timeout shouldn't lose the message. So the booking handler only **enqueues** a job and answers straight away; a worker sends the SMS in the background, retries when the gateway fails, and gives up cleanly when retrying is pointless.

```
Enqueue("sms:booking-1042", "Seat 3 ...")   →  nil
Enqueue("sms:booking-1042", "Seat 3 ...")   →  ErrDuplicate     (the booking handler retried; no second SMS)

07:00:00.0  RunDue  →  gateway timeout   Attempts 1, retry at 07:00:00.5
07:00:00.5  RunDue  →  gateway timeout   Attempts 2, retry at 07:00:01.5
07:00:01.5  RunDue  →  sent

"sms:bad-number"  →  Permanent(no such phone number)  →  dead letters after 1 attempt
"sms:1050"        →  fails 3 times (MaxAttempts 3)     →  dead letters, someone looks at it
```

## What you'll use

- **A queue of jobs with a time on each.** A job runs when its `RunAt` has passed. A failed job goes back with a later `RunAt`.
- **Exponential backoff with jitter.** Wait 1 s, then 2 s, then 4 s, up to a cap, so a struggling gateway gets room to recover. Then add randomness: if a thousand SMSes failed in the same second, a thousand retries without jitter all arrive in the same second again. Today's `Backoff` takes half the ceiling plus a random share of the other half.
- **An injected clock and random source.** `Queue` asks a `Clock` for the time and `Options.Rand` for randomness. Production passes `SystemClock{}` and `rand.Float64`; tests pass a fake clock they move by hand and a function that always returns 0. No sleeps, and the same result every run.
- **Deduplication by key.** Each job has a key that names the work, like `sms:booking-1042`. `Enqueue` refuses a key it has seen before, so a retried HTTP request can't send two SMSes.
- **Idempotent workers.** Real queues deliver **at least once**: a worker can send the SMS and crash before it marks the job done, and the job runs again. So the work itself must be safe to repeat. `Idempotent` wraps a handler and skips a key that already succeeded.
- **A dead-letter list.** A job that fails `MaxAttempts` times, or fails with a `Permanent` error, stops retrying and is kept with its last error for a human to look at. Retrying a wrong phone number for ever only burns money.

```go
var p permanent
if errors.As(err, &p) { /* don't retry */ } // IsPermanent, given in the starter

slices.SortFunc(due, byRunAt)           // earliest first, enqueue order on a tie
j.RunAt = now.Add(Backoff(j.Attempts, base, max, rnd))
```

**Where real jobs live.** An in-process queue loses everything when the process restarts, so it's for small, losable work. Durable options:

- **Postgres.** A `jobs` table with `run_at`, `attempts` and `last_error`. Workers claim work with `SELECT ... WHERE run_at <= now() ORDER BY run_at LIMIT 10 FOR UPDATE SKIP LOCKED`, so two workers never take the same row. A unique index on the key gives you dedupe for free.
- **Redis.** A list (`LPUSH`/`BRPOP`) for simple queues, a sorted set scored by run time for delayed retries, or Redis Streams with consumer groups for acknowledgements.
- **The outbox pattern.** "Save the booking, then enqueue the SMS" has a gap: the process can die between the two, or the enqueue can fail after the commit. Instead, insert the job into an `outbox` table **in the same transaction** as the booking. Either both are saved or neither is. A separate worker reads the outbox and publishes or runs each job, marking it sent. Because it can publish twice, the consumer must be idempotent: the same lesson as today.

## Why it matters at work

Anything slow or flaky that the user doesn't need to wait for goes in a queue: emails, SMSes, M-Pesa callbacks, report generation, webhooks to partners. The bugs are always the same: duplicate charges or messages because a job ran twice, a retry storm that knocks the gateway over again the moment it recovers, and a poison job that fails for ever and blocks the jobs behind it. A reviewer asks "what's the idempotency key?", "is there jitter?", "what's the max attempts, and where do dead jobs go?" and "can the job be lost between the commit and the enqueue?". Having the dead-letter list on a dashboard with an alert is what turns a silent failure into a ticket.

## Steps

1. Open `starter/jobs.go`. `Clock`, `Job`, `Handler`, `Permanent`, `IsPermanent`, `Options`, `Queue`, `NewQueue` and `byRunAt` are given.
2. Write `Backoff`. Start the ceiling at `base` and double it `attempt - 1` times, stopping once it reaches `max` (so a huge attempt can't overflow). Cap it with `min`. Return `half + time.Duration(rnd()*float64(ceiling-half))`.
3. Write `Enqueue`. Under the lock: if `q.seen[key]`, return `ErrDuplicate`. Otherwise mark it seen and append a job with `RunAt: q.clock.Now()` and the next `seq`.
4. Write `RunDue`. Read the time once. Under the lock, split `pending` into due and not due, and keep the not-due ones. Unlock, sort the due ones with `byRunAt`, and run each with the handler. On an error: add one to `Attempts`, save `err.Error()` in `LastErr`, then under the lock either append to `dead` (permanent, or `Attempts >= MaxAttempts`) or set a new `RunAt` and append to `pending`. Return how many ran.
5. Write `Run`: a `for` loop with a `select` on `ctx.Done()` and `tick` ([Day 43](../../phase-4-concurrency/day-043-select-timeouts/)).
6. Write `Pending` (a sorted copy) and `DeadLetters` (a copy). Return copies so callers can't change the queue behind its lock.
7. Write `Idempotent`: a map of keys that succeeded, behind its own mutex. Don't hold the mutex while the handler runs.
8. Run the tests from the repo root:

```bash
go test -race ./phase-7-production/day-079-background-jobs/starter
```

## The kata: Retry Timeline

Simulate one SMS worker with flaky jobs, exponential backoff and a dead-letter limit, and print when each job finished. It's a priority-queue simulation, the kind interviews like. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 79
```

## When you're stuck

- **`Backoff(7, 1s, 1m, rnd=0) = 32s, want 30s`**: you capped after halving, or not at all. Cap the ceiling at `max` first, then take half.
- **`Backoff(100, ...)` returns a negative number or hangs**: doubling 99 times overflowed. Stop doubling once the ceiling reaches `max`.
- **`RunDue 1ms early ran 1 jobs, want 0`**: you compared with `Before` the wrong way round. A job is due when `!j.RunAt.After(now)`.
- **`after two failures ... want RunAt +1.5s`**: the new `RunAt` is the time `RunDue` started plus the backoff, not the old `RunAt` plus the backoff.
- **`DeadLetters = [] ...`**: check `j.Attempts >= q.opts.MaxAttempts` after adding one to `Attempts`.
- **`fatal error: all goroutines are asleep - deadlock!`** in a handler that enqueues: you held `q.mu` while calling the handler. Unlock first.
- **The race detector reports `Pending` or `DeadLetters`**: you returned `q.pending` itself. Return `slices.Clone(...)`.
- **Still stuck?** Open `solution/jobs.go`, read it once, close it, then write your own version from memory.
