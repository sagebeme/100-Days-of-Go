# HTTP: methods and status codes

A reference to keep open from Day 49 onwards. Every example uses the matatu-booking API you build in Phase 5: routes, trips, seats and bookings.

## A request and a response

```
POST /bookings HTTP/1.1                  <- method, path
Authorization: Bearer eyJhbGci...        <- headers
Content-Type: application/json

{ "tripId": 42, "seats": 2 }             <- body

HTTP/1.1 201 Created                     <- status code
Location: /bookings/9001
Content-Type: application/json

{ "id": 9001, "seats": 2 }
```

The **method** says what you want done, the **path** says to what, and the **status code** says how it went.

In Go, `r.Method` holds the method and `r.URL.Path` holds the path. Compare them with the constants in `net/http`, such as `http.MethodGet` and `http.MethodPost`, not with strings you type yourself. You set the status code with `w.WriteHeader(http.StatusCreated)`. If you never call it, the first `w.Write` sends `200 OK`.

## Methods

Two words decide how a method may be used:

- **Safe**: it doesn't change anything on the server. Browsers, caches and crawlers fetch safe requests freely, so a `GET` must never delete or book anything.
- **Idempotent**: sending it twice has the same effect as sending it once. This is what makes a retry over a bad connection harmless.

| Method | Go constant | Use it to | Safe | Idempotent | Body | Usual success |
|---|---|---|---|---|---|---|
| `GET` | `http.MethodGet` | read something | yes | yes | no | `200 OK` |
| `HEAD` | `http.MethodHead` | `GET` without the body: does it exist, how big is it? | yes | yes | no | `200 OK` |
| `OPTIONS` | `http.MethodOptions` | ask what's allowed; browsers send it before cross-site requests (CORS, Day 61) | yes | yes | no | `204 No Content` |
| `POST` | `http.MethodPost` | create something, or run an action | no | **no** | yes | `201 Created` + `Location`, or `200 OK` |
| `PUT` | `http.MethodPut` | replace something whole, at a path you know | no | yes | yes | `200 OK`, or `201` if it was created |
| `PATCH` | `http.MethodPatch` | change some fields of something | no | not guaranteed | yes | `200 OK` |
| `DELETE` | `http.MethodDelete` | remove something | no | yes | usually no | `204 No Content` |

On the booking API:

```
GET    /trips               the list
GET    /trips/42            one trip
POST   /trips               create one              -> 201, Location: /trips/43
PUT    /trips/42            replace it whole (send every field)
PATCH  /trips/42            { "fare": 150 }         change just the fare
DELETE /trips/42            -> 204
POST   /bookings            book seats: creates a booking, so POST
POST   /bookings/9001/pay   pay for it: an action, so POST
```

### Routing by method in Go 1.22 and later

Since Go 1.22, `http.ServeMux` patterns can start with a method and contain `{name}` wildcards. You don't need a router library for this course.

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /trips", listTrips)
mux.HandleFunc("GET /trips/{id}", getTrip)
mux.HandleFunc("POST /trips", createTrip)
mux.HandleFunc("PATCH /trips/{id}", updateTrip)
mux.HandleFunc("DELETE /trips/{id}", deleteTrip)
mux.HandleFunc("POST /bookings", createBooking)
```

Inside a handler, `r.PathValue("id")` returns the wildcard's text. It's always a string, so convert it and handle the error.

Three things the mux does for you:

- A `GET` pattern also matches `HEAD`.
- If the path matches but the method doesn't, the mux answers `405 Method Not Allowed` with an `Allow` header listing the methods that would work.
- A pattern ending in `/`, such as `"GET /static/"`, matches everything below it. Use `{$}` to match only the exact path: `"GET /{$}"` is the home page and nothing else.

### PUT or PATCH?

- `PUT` sends the **whole** thing. A field you leave out is gone. Sending the same `PUT` twice leaves the same result: idempotent.
- `PATCH` sends **only what changes**. `{ "fare": 150 }` touches the fare and nothing else. `{ "seatsLeft": { "decrement": 1 } }` is a `PATCH` that is *not* idempotent: twice means two seats.

Most APIs use `PATCH` for edit forms, because the form rarely has every field.

In Go, a `PATCH` body needs to tell "not sent" apart from "sent as zero". Use pointer fields: a nil `*int` means the client left it out.

```go
type tripPatch struct {
	Fare  *int    `json:"fare"`
	Route *string `json:"route"`
}
```

### Why `POST` twice is dangerous, and the fix

"Book 2 seats" sent twice books 4. A phone on a matatu sends it, the connection drops before the answer arrives, the app retries: two bookings and a double M-Pesa charge. The fix is an **idempotency key**: the client makes up a random ID per booking and sends it as an `Idempotency-Key` header. The server remembers keys it has seen and, for a repeat, returns the first answer instead of booking again. Payment APIs work this way, and Day 79 builds workers on the same idea.

`DELETE` twice is fine: the second one finds nothing to delete. It may answer `404`, but the *state* is the same, which is what idempotent means.

## Status codes

The first digit is the class. A client that doesn't know a code treats it like the `x00` of its class.

| Class | Means | Whose move |
|---|---|---|
| `1xx` | informational: "carry on" | nobody's, yet |
| `2xx` | it worked | done |
| `3xx` | look elsewhere | the client follows |
| `4xx` | **your** request is wrong | the client must change something |
| `5xx` | **the server** failed | not the client's fault; retrying may help |

`http.StatusText(code)` gives the standard phrase, such as `"Not Found"`.

### 1xx: informational

You'll rarely handle these yourself. Go's server sends `100 Continue` for you.

| Code | Go constant | When you see it |
|---|---|---|
| `100 Continue` | `http.StatusContinue` | The client asked "may I send this big upload?" first; the server says go ahead |
| `101 Switching Protocols` | `http.StatusSwitchingProtocols` | The connection becomes a WebSocket: live seats on Day 71 |
| `103 Early Hints` | `http.StatusEarlyHints` | "Start fetching this CSS while I build the page" |

### 2xx: success

| Code | Go constant | When your API returns it |
|---|---|---|
| `200 OK` | `http.StatusOK` | Here it is. `GET /trips/42`, or a `PATCH` that returns the updated trip |
| `201 Created` | `http.StatusCreated` | I made it. `POST /bookings`, with a `Location: /bookings/9001` header |
| `202 Accepted` | `http.StatusAccepted` | Got it, I'll do it later. An SMS to every passenger on a trip, queued |
| `204 No Content` | `http.StatusNoContent` | Done, nothing to show. `DELETE /trips/42`, `POST /logout`. Don't write a body |
| `206 Partial Content` | `http.StatusPartialContent` | Here's the slice you asked for, with `Range`. `http.ServeContent` handles it for you |

### 3xx: redirection

| Code | Go constant | When you see it |
|---|---|---|
| `301 Moved Permanently` | `http.StatusMovedPermanently` | It lives at a new address for good, and browsers remember it. That's why the link shortener on Day 85 uses `302`: with a `301`, repeat clicks never reach you to be counted |
| `302 Found` | `http.StatusFound` | Temporarily elsewhere. After a form login, "go to /dashboard" |
| `303 See Other` | `http.StatusSeeOther` | After a `POST`, go and `GET` this page instead, so refreshing doesn't book twice |
| `304 Not Modified` | `http.StatusNotModified` | You already have the latest copy (your `If-None-Match` ETag matched), so no body is sent. Saves data on slow connections |
| `307` / `308` | `http.StatusTemporaryRedirect` / `http.StatusPermanentRedirect` | Like `302` / `301`, but the method and body must stay the same: a redirected `POST` stays a `POST` |

Send one with `http.Redirect(w, r, "/dashboard", http.StatusSeeOther)`.

### 4xx: the client's mistake

| Code | Go constant | When your API returns it |
|---|---|---|
| `400 Bad Request` | `http.StatusBadRequest` | The request can't be read: broken JSON, a missing field |
| `401 Unauthorized` | `http.StatusUnauthorized` | **Who are you?** No token or session, or it expired. Logging in can fix it. Send `WWW-Authenticate` (Day 58) |
| `403 Forbidden` | `http.StatusForbidden` | **I know who you are, and no.** A passenger trying to create a trip (Day 59) |
| `404 Not Found` | `http.StatusNotFound` | Nothing there. Also used instead of `403` when even admitting it exists would leak something: another passenger's booking |
| `405 Method Not Allowed` | `http.StatusMethodNotAllowed` | The path exists but not for this method: `DELETE /health`. The mux sends it, with `Allow` |
| `406 Not Acceptable` | `http.StatusNotAcceptable` | You asked (in `Accept`) for a format I can't produce |
| `408 Request Timeout` | `http.StatusRequestTimeout` | You took too long to send the request |
| `409 Conflict` | `http.StatusConflict` | It clashes with what's there: that phone number already has an account; the seat was just taken |
| `410 Gone` | `http.StatusGone` | It existed and was removed for good: a cancelled trip's page |
| `412 Precondition Failed` | `http.StatusPreconditionFailed` | Someone else changed it since you read it (your `If-Match` ETag is stale), so your edit would overwrite theirs |
| `413 Content Too Large` | `http.StatusRequestEntityTooLarge` | The body is too big: a 40 MB photo for a profile picture. Go's constant still uses the old name |
| `415 Unsupported Media Type` | `http.StatusUnsupportedMediaType` | I only take JSON, and you sent a form |
| `422 Unprocessable Content` | `http.StatusUnprocessableEntity` | The JSON is fine but the values aren't: zero seats, a trip date in the past (Day 53). Go's constant uses the old name |
| `429 Too Many Requests` | `http.StatusTooManyRequests` | Slow down. Send `Retry-After`. Stops password guessing and booking bots (Day 47, Day 60) |

**400 or 422?** `400` when you can't even read the request, `422` when you read it and the values break a rule. Many APIs use `400` for both; pick one and stay consistent.

**401 or 403?** Will logging in (again) fix it? Then `401`. If not, `403`.

### 5xx: the server's mistake

| Code | Go constant | When you see it |
|---|---|---|
| `500 Internal Server Error` | `http.StatusInternalServerError` | A bug. Log the details; tell the client nothing about your code |
| `501 Not Implemented` | `http.StatusNotImplemented` | The server doesn't support that method at all |
| `502 Bad Gateway` | `http.StatusBadGateway` | A server in front (a proxy, a load balancer) got a bad answer from yours: your app crashed |
| `503 Service Unavailable` | `http.StatusServiceUnavailable` | Down or overloaded for now. Send `Retry-After`. Friday evening before a long weekend |
| `504 Gateway Timeout` | `http.StatusGatewayTimeout` | The proxy gave up waiting for your app: a slow query, or M-Pesa taking its time |

## Writing JSON, and JSON errors

Give every error the same shape, so the front end can show `body.error` without guessing. Two small helpers cover the whole API:

```go
type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}
```

The order matters: set headers, then call `WriteHeader`, then write the body. Headers set after `WriteHeader` are ignored, and calling `WriteHeader` twice logs "superfluous response.WriteHeader call".

`http.Error(w, msg, code)` exists too, but it sends plain text. Use it for quick tools, not for a JSON API.

For a `500`, write a generic message to the client and the real error to your log:

```go
if err != nil {
	slog.Error("saving booking", "err", err)
	writeError(w, http.StatusInternalServerError, "something went wrong")
	return
}
```

## Reading a JSON body safely

A handler that reads a body should limit its size, reject fields it doesn't know, and pick the status code from what went wrong:

```go
const maxBody = 1 << 20 // 1 MB

func createBooking(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req bookingRequest
	if err := dec.Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "body must be 1 MB or less")
			return
		}
		writeError(w, http.StatusBadRequest, "body must be valid JSON")
		return
	}
	if req.Seats < 1 || req.Seats > 14 {
		writeError(w, http.StatusUnprocessableEntity, "seats must be between 1 and 14")
		return
	}

	id := 9001 // saved to the database on Day 55
	w.Header().Set("Location", "/bookings/"+strconv.Itoa(id))
	writeJSON(w, http.StatusCreated, map[string]int{"id": id, "seats": req.Seats})
}
```

`http.MaxBytesReader` stops reading after `n` bytes and returns an error of type `*http.MaxBytesError`. It also tells the server to close the connection, so a client can't keep streaming gigabytes at you. Without it, `json.NewDecoder(r.Body)` will read as much as the client sends.

Also set timeouts on the server. `http.ListenAndServe` has none, so a slow client can hold a connection open forever:

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout:       10 * time.Second,
	WriteTimeout:      10 * time.Second,
	IdleTimeout:       60 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```

## Which ones are safe to retry?

| Response | Retry? |
|---|---|
| network error, `502`, `503`, `504` | **yes**, with a growing wait (1 s, 2 s, 4 s...) and only for idempotent requests, or a `POST` with an idempotency key |
| `429` | yes, after `Retry-After` |
| `401` with `invalid_token` | refresh the token once, then replay (Day 58) |
| `500` | maybe once; it's probably a bug that will happen again |
| any other `4xx` | **no**: the same request gets the same answer. Fix the request |

## Where the course uses them

| Day | What you meet |
|---|---|
| 49 | your first server: handlers, `ServeMux`, `200` and `404` |
| 50 | routes per method, safety and idempotency, `405` with `Allow` |
| 51 | choosing the right status code, one shape for every error |
| 52 | JSON in and out, `Content-Type`, `201` with `Location`, `415` |
| 53 | validation, `http.MaxBytesReader`, `400`, `413` and `422` |
| 58 | sessions and JWTs: `401` with `WWW-Authenticate` |
| 59 | roles: `401`, `403` or `404`? |
| 60–61 | rate limits and `429`, CORS and `OPTIONS` |
| 71 | WebSockets: `101` |
| 85 | redirects: `302`, and why not `301` |

See also [owasp-checklist.md](owasp-checklist.md) for the security side of the same handlers.
