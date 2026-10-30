# Day 50: HTTP Methods

## The brief

The booking API can now list trips, add one, replace one, change one field of one, cancel one, and book a seat. Each of those is a different HTTP method on the same few paths, and the method tells everyone, from the browser to a cache to the app's retry logic, what the request is allowed to do.

The hard one is booking. A passenger on a matatu in a tunnel taps "Book seat 3", the request reaches the server, the answer never makes it back, and the app tries again. Today that retry must not book a second seat or charge a second time.

```
GET    /trips                          →  200  [{"id":1,"routeId":125,...}]
POST   /trips      {"routeId":111,...}  →  201  Location: /trips/2
PUT    /trips/1    {whole trip}         →  200  (twice: same result)
PATCH  /trips/1    {"seats":10}         →  200  only seats changed
DELETE /trips/1                         →  204  (twice: 404, same state)
DELETE /trips                           →  405  Allow: GET, HEAD, POST
HEAD   /trips/1                         →  200  headers only

POST /trips/1/bookings  Idempotency-Key: c0ffee  {"seat":3,...}  →  201  Location: /bookings/1
POST /trips/1/bookings  Idempotency-Key: c0ffee  {"seat":3,...}  →  201  same body, Idempotent-Replayed: true
POST /trips/1/bookings  Idempotency-Key: c0ffee  {"seat":5,...}  →  422  key reused for a different request
POST /trips/1/bookings  (no key)                 {"seat":3,...}  →  409  seat already booked
```

Route numbers in the examples are real Nairobi routes (125 to Ongata Rongai, 111 to Ngong, 58 to Buruburu), checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026. Routes change, so confirm with the conductor.

## What you'll use

Read the Methods section of [reference/http.md](../../reference/http.md) before you start. The two words that matter:

- **Safe**: the request changes nothing. `GET` and `HEAD` are safe, so a crawler or a link preview can send them freely. A `GET` that books a seat will be triggered by something you didn't expect.
- **Idempotent**: sending it twice leaves the server in the same state as sending it once. `GET`, `HEAD`, `PUT` and `DELETE` are. `POST` isn't, and `PATCH` isn't guaranteed to be.

| Request | Safe | Idempotent | Today |
|---|---|---|---|
| `GET /trips/1` | yes | yes | read |
| `PUT /trips/1` | no | yes | replace every field; a field you leave out is an error |
| `PATCH /trips/1` | no | not guaranteed | change only the fields sent |
| `DELETE /trips/1` | no | yes | the second one answers `404`, but the state is the same |
| `POST /trips/1/bookings` | no | **no** | unless the client sends an `Idempotency-Key` |

The Go pieces:

- **Method patterns**: `mux.HandleFunc("PATCH /trips/{id}", a.patchTrip)`. The mux answers `405` with an `Allow` header by itself when the path matches and the method doesn't, and a `GET` pattern answers `HEAD` too. The server drops the body of a `HEAD` answer for you.
- **Method values as handlers**: `a.patchTrip` is a method value, a function bound to `a`, so each handler can reach `a.store`.
- **Pointer fields for PATCH**: in `struct{ Seats *int }`, a nil pointer means "not sent", which is different from "sent as 0".
- **`r.Header.Get("Idempotency-Key")`** reads a request header. Header names aren't case-sensitive, and `Get` handles that.

```go
type tripPatch struct {
	Seats *int `json:"seats"`
}
var p tripPatch
json.NewDecoder(r.Body).Decode(&p) // {} leaves p.Seats nil; {"seats":0} makes it point to 0
if p.Seats != nil {
	t.Seats = *p.Seats
}
```

**How an idempotency key works.** The client makes up a random key per booking and sends the same key on every retry. The server keeps a map from key to the answer it gave. A repeat with the same key and the same body gets the saved answer, byte for byte, and books nothing. The same key with a different body means the client has a bug, and the [IETF draft for the header](https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/) suggests `422`. Payment APIs such as Stripe work this way. Today the map lives in memory and holds only successful answers. A real one is stored in the database next to the booking, scoped to the user, and expires after a day or so.

## Why it matters at work

Every phone network drops requests, and every mobile app and HTTP client library retries. If `POST /bookings` isn't protected, retries turn into double bookings and double M-Pesa charges, and the support line finds out before you do. In code review, people check that each handler's method matches what it does: no `GET` that changes data, `PUT` that really replaces, `DELETE` that's safe to repeat, and an idempotency key on any `POST` that moves money. Load balancers, proxies and [Day 79](../../phase-7-production/day-079-background-jobs/)'s job workers retry idempotent requests on their own, so a handler that lies about its method breaks in ways that are hard to trace.

## Steps

`starter/store.go` is already written: the `Trip` and `Booking` types, an in-memory `Store`, and the `writeJSON` and `writeError` helpers from reference/http.md. Read it first. You write the handlers in `starter/api.go`.

1. Write a helper `pathID(w, r) (int, bool)` that converts `r.PathValue("id")` with `strconv.Atoi`. If it isn't a whole number above 0, write a `400` and return false.
2. `GET /trips`: `writeJSON(w, 200, a.store.Trips())`. `Trips` never returns nil, so an empty list is `[]`, not `null`.
3. `POST /trips`: decode a body with `routeId`, `departs` and `seats`. Reject broken JSON or a missing field with `400`. Save it with `CreateTrip`, set `Location: /trips/{id}` and answer `201` with the trip.
4. `GET /trips/{id}`: `200` with the trip, or `404`.
5. `PUT /trips/{id}`: the same body and checks as `POST`, then `UpdateTrip` replaces every field. `404` if there's no such trip.
6. `PATCH /trips/{id}`: decode into a struct of pointer fields. Reject a sent `routeId` or `seats` below 1 with `400`. In `UpdateTrip`, copy only the non-nil fields.
7. `DELETE /trips/{id}`: `DeleteTrip`, then `w.WriteHeader(http.StatusNoContent)` with no body, or `404`.
8. `GET /bookings/{id}`: `200` or `404`.
9. `POST /trips/{id}/bookings`. Read the body with `io.ReadAll`. Write a function `book(tripID, body)` that returns the status, the `Location` and the value to send, without writing anything: `400` for broken JSON, a seat outside 1 to the trip's seats, or a missing passenger or phone; `404` for no trip; `409` for `ErrSeatTaken`; `201` and `/bookings/{id}` on success.
10. Now the key. With no `Idempotency-Key`, call `book` and write its answer. With one longer than 255 characters, answer `400`. Otherwise lock `a.mu` for the rest of the handler. If the key is in `a.seen` with the same trip ID and body, write the saved status, `Location`, `Content-Type` and body, plus `Idempotent-Replayed: true`. If it's there with a different request, answer `422`. If it's new, call `book`. On a 2xx, marshal the value with `json.Marshal`, add a `'\n'` so the bytes match what `writeJSON` sends, and save them. Then write the answer.
11. Register all eight routes in `Routes`.
12. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-050-http-methods/starter
```

## The kata: Allow Header

Today the mux works out `405` and the `Allow` header for you. In the kata you write that logic: match request paths against patterns with `{id}` wildcards and decide between `200`, `404` and `405` with the sorted list of allowed methods. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 50
```

## When you're stuck

- **`GET /trips` with no trips gives `null`**: you encoded a nil slice. `Store.Trips` makes an empty one, so use it as it is rather than declaring `var trips []Trip`.
- **`panic: pattern "POST /trips/{id}/bookings" ... conflicts`**: two patterns match the same requests with the same method. Check for a duplicate registration.
- **`HEAD` gets `405`**: you registered the route with a method the mux doesn't extend to `HEAD`. Only `GET` patterns answer `HEAD`.
- **PATCH `{"seats": 10}` resets the route to 0**: you decoded into the whole-trip struct and copied every field. Use pointer fields and copy only the non-nil ones.
- **The retry with the same key gets `409`**: you looked up the key after calling `book`. Check `a.seen` first.
- **The retry's body differs by a newline**: `json.Encoder` adds `'\n'` after the value and `json.Marshal` doesn't. Append one before saving.
- **`http: superfluous response.WriteHeader call`**: you wrote a status twice, often an error followed by a missing `return`. Every `writeError` in a handler should be followed by `return`.
- **Still stuck?** Open `solution/api.go`, read it once, close it, then write your own version from memory.
