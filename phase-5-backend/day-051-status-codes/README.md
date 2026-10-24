# Day 51: Status Codes

## The brief

The booking app's front end is being built by Nduta, and she has one request: "Tell me what went wrong in a way my code can act on." Seat taken? She shows the seat map again. Not logged in? She opens the login screen. Too many tries? She waits and retries. Server broke? She shows "try again" and doesn't blame the passenger.

She can only do that if each outcome has its own status code and every error has the same shape. Today you give her both.

```
GET    /trips/1                                    →  200  {"id":1,"routeId":125,...}
GET    /trips/abc                                  →  400  {"error":"id must be a positive whole number"}
GET    /trips/9                                    →  404  {"error":"not found"}
POST   /trips/1/bookings  {"seat":5,...}           →  201  Location: /bookings/2
POST   /trips/1/bookings  {"seat":5,               →  400  {"error":"body must be valid JSON"}
POST   /trips/1/bookings  {"seat":15,...}          →  422  {"error":"seat must be between 1 and 14"}
POST   /trips/1/bookings  {"seat":3,...}           →  409  {"error":"that seat is already booked"}
POST   /trips/1/bookings  (the 31st in a minute)   →  429  Retry-After: 30
DELETE /bookings/1        (no token)               →  401  WWW-Authenticate: Bearer realm="bookings"
DELETE /bookings/1        (a clerk's token)        →  403  {"error":"only an admin can cancel bookings"}
DELETE /bookings/1        (an admin's token)       →  204  (no body)
GET    /trips/1           (database is down)       →  500  {"error":"something went wrong"}
```

## What you'll use

The Status codes section of [reference/http.md](../../reference/http.md) has every code you'll meet. The choices that come up in review again and again:

| Question | Answer |
|---|---|
| Created something? | `201` and a `Location` header pointing at it |
| Done, nothing to send back? | `204`, and write no body at all |
| `400` or `422`? | `400` if you couldn't read the request (broken JSON, an ID that isn't a number). `422` if you read it fine and a value breaks a rule (seat 15 on a 14-seater) |
| `401` or `403`? | Would logging in fix it? `401`, with `WWW-Authenticate`. Logged in and still not allowed? `403` |
| `404` or `409`? | `404`: the thing isn't there. `409`: it's there, and your request clashes with its current state, like a seat someone just took |
| `429`? | The client is going too fast. Send `Retry-After` in whole seconds |
| `500`? | Your bug, not theirs. Log the real error; send a generic message |

The Go pieces:

- **Sentinel errors and `errors.Is`** (Day 17): the store returns `ErrNotFound` or `ErrSeatTaken`, possibly wrapped with `%w`, and `errors.Is` finds them through the wrapping.
- **A custom error type and `errors.As`**: `*InvalidError` carries a message that's safe to show. `errors.As(err, &invalid)` finds it and fills `invalid` in.
- **One function that maps errors to codes**, `StatusFor`, so every handler answers the same way. Handlers call one `fail(w, r, err)` and stop thinking about codes.

```go
var invalid *InvalidError
switch {
case errors.As(err, &invalid):
	return http.StatusUnprocessableEntity, invalid.Message
case errors.Is(err, ErrNotFound):
	return http.StatusNotFound, "not found"
}
```

`starter/store.go` is written for you: the types, a `Store` interface, a `Limiter` interface and an in-memory store. The API depends on the interfaces, so the tests can pass in a fake limiter and a store that fails like a dead database.

## Why it matters at work

Status codes are the contract between your API and every client, proxy and monitoring tool. Dashboards count 5xx as "our fault" and page someone at night, so a validation error sent as `500` wakes people up for nothing, and a real failure sent as `200 {"ok": false}` never alerts at all. Retry logic reads them too: clients retry `429` and `503` and never retry `409`. In review, the other thing people look for is a leak: `writeError(w, 500, err.Error())` sends your SQL, table names and database hostnames to whoever asked. OWASP lists that under A05, Security Misconfiguration (see [reference/owasp-checklist.md](../../reference/owasp-checklist.md)). Log the details, send the client a generic message.

## Steps

1. Read `starter/store.go`.
2. In `starter/api.go`, finish `InvalidError.Error` (return the message), then write `StatusFor` from its doc comment. Check `*InvalidError` with `errors.As` and the two sentinels with `errors.Is`.
3. Write `writeJSON` and `writeError`: set `Content-Type: application/json`, then `w.WriteHeader(status)`, then encode. The error body is a struct with one field, `Error string` tagged `json:"error"`.
4. Write `fail(w, r, err)`: get the status and message from `StatusFor`. For a 500, `log.Printf` the method, the path and the real error. Then `writeError` with the safe message.
5. Write a `pathID` helper as on Day 50: `400` when `{id}` isn't a whole number above 0.
6. `GET /trips/{id}` and `GET /bookings/{id}`: read from the store, `fail` on an error, `200` otherwise.
7. `POST /trips/{id}/bookings`, in this order:
   - Ask `a.limiter.Allow` with the client's IP: `net.SplitHostPort(r.RemoteAddr)` gives host and port. If it says no, set `Retry-After` to the wait in whole seconds, rounded up with `math.Ceil` and at least 1, and answer `429`. The limiter goes first so a flood of bad requests is still limited.
   - Read the ID and decode the body. Broken JSON is `400`.
   - Load the trip (`fail` turns `ErrNotFound` into `404`).
   - Return an `*InvalidError` for a seat outside 1 to `trip.Seats`, or a blank passenger or phone, and pass it to `fail`.
   - `a.store.Book`. On success set `Location: /bookings/{id}` and answer `201` with the booking.
8. `DELETE /bookings/{id}`: take the token from `Authorization: Bearer <token>` with `strings.CutPrefix`. No header or an unknown token: set `WWW-Authenticate: Bearer realm="bookings"` and answer `401`. A token whose role isn't `"admin"`: `403`. Otherwise cancel and answer `204` with `w.WriteHeader` only.
9. Register the four routes.
10. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-051-status-codes/starter
```

## The kata: Status Board

Status codes are also what you read when something goes wrong in production. The kata takes an access log and prints the count per class, the 5xx rate and the path that failed most, using whole-number maths so the percentage is exact. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 51
```

## When you're stuck

- **`StatusFor` returns 500 for a wrapped `ErrNotFound`**: you compared with `==`. `fmt.Errorf("trip 7: %w", ErrNotFound) == ErrNotFound` is false. Use `errors.Is`.
- **`errors.As` panics with `second argument to errors.As must be a non-nil pointer`**: pass the address of your variable, `&invalid`, where `invalid` is a `*InvalidError`.
- **The status is 200 but you called `writeError(w, 404, ...)`**: something wrote to `w` before, and the first write sends `200`. Look for a missing `return` after an earlier error.
- **The 429 test sees no `Retry-After`**: you set the header after `writeError`. Headers set after `WriteHeader` are dropped. Set it first.
- **`Retry-After` is `1.5` or `0`**: it must be whole seconds. Round up, and never send less than 1.
- **The limiter key is `192.0.2.1:1234`**: `r.RemoteAddr` includes the port, which changes on every connection. Split it off with `net.SplitHostPort`.
- **The leak test fails**: you sent `err.Error()` in a 500. Send `"something went wrong"` and log the error instead.
- **Still stuck?** Open `solution/api.go`, read it once, close it, then write your own version from memory.
