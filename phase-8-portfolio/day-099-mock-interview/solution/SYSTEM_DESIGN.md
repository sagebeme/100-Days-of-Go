# System design: Twende Seats (a sample answer)

One good answer, not the only one. Mark yours with the rubric in the day's README
before you read this.

## 1. Questions I'd ask, and the answers I'm assuming

- Are seats numbered, or is it first come, first seated? *Numbered: passengers pick a seat.*
- How long can a seat be held while the passenger pays? *3 minutes.*
- Are fares fixed per route, or do they change (more in December)? *Set per trip by the SACCO.*
- Refunds and cancellations? *Out of scope today; cancellations before departure free the seat.*
- One SACCO or many? *One for now, but don't make many impossible.*

## 2. Requirements

**Functional**

- List trips for a route and date, with free seats.
- Hold a seat, pay for it by STK push, get an SMS ticket with a code the conductor checks.
- The same flow by USSD for feature phones.
- Staff see the manifest for each trip.

**Non-functional**

- Never sell one seat twice. Correctness beats availability for the booking step.
- Booking answers in under 1 s at the 95th percentile; the payment itself takes as long as the passenger takes on their phone.
- Up 99.9% of the time. December matters most.
- Cheap: a SACCO, not a bank.

## 3. Numbers

- 50 000 bookings a day is about 0.6 a second on average. If 20% come in the busiest hour: 10 000 an hour, about 3 a second.
- December: 10x, so about 30 bookings a second at peak. Reads (browsing seats) maybe 10x that: 300 a second.
- A booking row is under 500 bytes: 50 000 × 365 × 0.5 KB ≈ 9 GB a year, before December.
- Conclusion: one Postgres primary handles this easily. The hard parts are correctness and payments, not scale.

## 4. API

```
GET  /routes/{id}/trips?date=2026-12-23     trips with free seat counts
GET  /trips/{id}/seats                      the seat map
POST /trips/{id}/holds  {seat, phone}       201 {hold_id, expires_at} or 409 seat taken
POST /holds/{id}/pay                        202: sends the STK push
POST /mpesa/callback                        from Safaricom: payment result (verify source)
GET  /bookings/{code}                       the ticket, for the conductor's app
```

Every POST from a client takes an `Idempotency-Key` header, so a retry on a bad network doesn't hold two seats.

## 5. Data model

```
trips     (id, route_id, departs_at, vehicle_id, fare_cents, seats)
holds     (id, trip_id, seat, phone, status, expires_at, mpesa_checkout_id)
          status: held | paid | expired | cancelled
          UNIQUE (trip_id, seat) WHERE status IN ('held', 'paid')
payments  (id, hold_id, mpesa_receipt UNIQUE, amount_cents, raw_callback, received_at)
```

The partial unique index is what stops double booking: the database refuses the second
`held` row for the same seat, however many servers race.

## 6. High-level design

```
 app / USSD gateway
        |
   [load balancer]
        |
   [API servers, stateless, 2+]  --->  [Postgres primary] ---> [read replica: seat maps]
        |            |                        ^
        |            +--> [job queue] --> [workers: SMS, expiry sweep, payment status checks]
        v                                     |
  M-Pesa Daraja (STK push)  ---- callback ----+--> POST /mpesa/callback
```

- API servers hold no state, so we add more in December.
- Workers send SMS, expire old holds, and ask M-Pesa about payments whose callback never came.
- The replica serves the seat map, which may be a second stale. The hold itself always goes to the primary.

## 7. Deep dives

**No double booking**

- `INSERT INTO holds ... status 'held'` inside one transaction. If the unique index fires, return `409` and the client refreshes the seat map.
- Expiry: a worker marks `held` rows past `expires_at` as `expired` every few seconds, which frees the seat. Reads also treat an expired hold as free, so a slow worker can't block a seat.
- Don't use a cache or an in-memory lock as the source of truth. They can be wrong after a restart; the database constraint can't.

**The payment callback**

- The callback can arrive late, twice, or never. Store it raw first, then process it.
- Idempotent: `mpesa_receipt` is unique, so a second copy of the same callback inserts nothing and changes nothing.
- Late payment for an expired hold: if the seat is still free, take it; if not, refund and send an SMS. Decide this with the SACCO, but decide it explicitly.
- Never came: a worker uses Daraja's STK push query API for holds stuck in `held` with a checkout ID, before expiring them.
- Only accept callbacks we can trust: a secret path, and the checkout ID must match a hold we created.

## 8. What breaks, and how it scales

- Postgres down: booking stops (we chose correctness). A standby with automatic failover limits it to a minute or two.
- M-Pesa slow or down: holds expire unpaid. Show a clear message and don't charge late.
- SMS provider down: the ticket is still in the app and the conductor's manifest; retry SMS from the queue.
- December: more API servers and workers, a bigger read replica, and load-test it in November with real December numbers.
- Later, many SACCOs: add `sacco_id` everywhere now, so sharding by SACCO is possible if it's ever needed.

## 9. Trade-offs I made

- Postgres constraints over a distributed lock: simpler and correct; fine at 30 writes a second.
- A 3-minute hold: long enough for a passenger to find their M-Pesa PIN, short enough that seats don't sit locked. Measure and tune.
- Seat maps from a replica may be a second stale. A passenger might pick a seat that just went and get a `409`. That's acceptable; selling it twice is not.
