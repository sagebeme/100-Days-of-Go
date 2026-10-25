# Kata: Double Bookings

Before launch, the SACCO wants to replay last month's booking log from the old system against the new rule: one seat on one trip belongs to at most one active booking. Replay it and print what the new API would have answered.

Each line is one of:

- `BOOK id trip seat`: a new booking `id` for that seat. If an active booking already holds the seat, print `TAKEN` and that booking's ID, and the new booking is never created. Otherwise print `OK`.
- `CANCEL id`: if `id` is an active booking, cancel it, free its seat, and print `CANCELLED`. If `id` never succeeded or was already cancelled, print `UNKNOWN`.

After the last line, print how many bookings are still active.

**Input**

- Line 1: `n`, the number of operations (1 ≤ n ≤ 200 000)
- Next `n` lines: an operation. IDs and trip names have no spaces; `seat` is a whole number from 1 to 33. Each `BOOK` line has a new ID.

**Output**

`n + 1` lines: one answer per operation, then the count.

**Example**

```
Input                    Output
7
BOOK b1 46K-0630 3       OK
BOOK b2 46K-0630 3       TAKEN b1
BOOK b3 46K-0700 3       OK
CANCEL b1                CANCELLED
BOOK b4 46K-0630 3       OK
CANCEL b2                UNKNOWN
CANCEL b1                UNKNOWN
                         2
```

`b2` was refused, so cancelling it is `UNKNOWN`. Seat 3 on the 06:30 is free again after `b1` is cancelled, so `b4` gets it. `b3` and `b4` are left.

**Limits:** 2 seconds per test case.

**Go tip:** two maps that point at each other: `held[seat] = id` answers "who has this seat?" and `active[id] = seat` answers "which seat does this booking hold?". Cancelling deletes from both. A struct `seat{trip string; n int}` works as a map key. This is the same rule as the API's `UNIQUE (trip_id, seat)`, but a real database would also need cancelled rows kept, so the rule there becomes a unique index on active bookings only.

Judge it from the repo root with `go run ./tools/kata 64`.
