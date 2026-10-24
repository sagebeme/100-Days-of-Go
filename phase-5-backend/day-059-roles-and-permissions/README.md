# Day 59: Roles and Permissions

## The brief

Twende (a made-up matatu SACCO app) now has three kinds of user. **Passengers** list trips and book seats. **Conductors** check passengers in at the stage, so they can read any booking. **Admins** create trips. Last week a passenger changed the number in `/bookings/2` to `/bookings/3` and saw a stranger's name and phone number. Today you put every "may they?" decision in one place and close that hole.

```
GET  /trips                    no token          →  401  WWW-Authenticate: Bearer
POST /trips                    Wanjiku (passenger) →  403  {"error": "you're not allowed to do that"}
POST /trips                    Amina (admin)       →  201
GET  /bookings/2               Wanjiku, her own    →  200
GET  /bookings/2               Otieno, not his     →  404  {"error": "no such booking"}
GET  /bookings/999             Otieno, no such     →  404  {"error": "no such booking"}
GET  /bookings/2               Kamau (conductor)   →  200
POST /bookings/2/checkin       Kamau               →  204
POST /bookings {"tripId":1,"seat":9,"phone":"0712345678"}  Otieno  →  400 (you can't book in someone else's name)
```

## What you'll use

- **401 or 403.** `401 Unauthorized` means "who are you?": no token, or a bad one. Logging in fixes it, and the response carries `WWW-Authenticate: Bearer`. `403 Forbidden` means "I know who you are, and no." See [http.md](../../reference/http.md).
- **A policy table.** One `map[Role]map[Action]bool` says what each role may do. `Can(role, action)` reads it. Anything not in the table is denied, including a role nobody has heard of. This is deny by default.
- **Middleware in two layers.** `Authenticate` finds the user and puts it in the request context; it never rejects. `Require(action, handler)` wraps each route and does the 401 or 403. Every route names its action next to its path, so a reviewer can read the whole policy in `NewServer`.
- **Object-level checks.** `Require` answers "may passengers read bookings?" It can't answer "is *this* booking yours?", because it hasn't loaded the booking. That check goes in the handler, after the lookup, with `CanViewBooking(user, booking)`. Forgetting it is called IDOR (insecure direct object reference) or BOLA (broken object-level authorisation). It's the most common API security bug there is.
- **404, not 403, for other people's things.** If `/bookings/3` answered 403, an attacker would learn that booking 3 exists. Answer exactly as you would for a missing one.
- **Never trust identity from the body.** The booking's name and phone come from the logged-in user. A body with a `phone` field is rejected by `DisallowUnknownFields`.

```go
var permissions = map[Role]map[Action]bool{
	Passenger: {ListTrips: true, CreateBooking: true, ReadBooking: true},
	// ...
}

mux.Handle("POST /trips", Require(CreateTrip, http.HandlerFunc(s.createTrip)))
```

`mux.Handle` takes an `http.Handler`; `http.HandlerFunc(f)` turns a function with the right signature into one. A method value like `s.createTrip` works as that function.

## Why it matters at work

Broken access control is number one on the OWASP Top 10 (A01 in [owasp-checklist.md](../../reference/owasp-checklist.md)), and it's rarely a missing login. It's a handler that loads a row by the ID in the URL and never asks whose it is, or a new route someone added without the role check because the check lived inside each handler and they copied the wrong one. Putting the policy in one table and one middleware means a reviewer reads one file, and a missing `Require` stands out in the route list. Teams often add a test like `TestRoutesByRole`, a grid of every route against every role, so a change to the policy shows up as a failing row instead of a bug report from a passenger.

## Steps

1. Open `starter/roles.go`. The roles, actions, types, `StaticTokens`, `Authenticate`, the `MemStore`, the JSON helpers and three of the handlers are given.
2. Fill in `permissions`. Passengers: list trips, create bookings, read bookings. Conductors: list trips, read bookings, read any booking, check in. Admins: every action. Write each one out; don't add a "star means everything" rule, so a new action is denied until someone grants it.
3. Write `Can`. One line: reading a missing key of a nil inner map is safe and gives `false`.
4. Write `Require`. Get the user with `UserFrom`. None: set `WWW-Authenticate: Bearer`, then `writeError` 401. `Can` says no: log `"access denied"` with the user ID, role, action and path, then 403. Otherwise call `next`.
5. Write `CanViewBooking`: true with `ReadAnyBooking`, or with `ReadBooking` when the booking's phone is the user's phone.
6. Write `getBooking`. Parse `r.PathValue("id")` with `strconv.Atoi`. Load the booking. If it's missing **or** `CanViewBooking` says no, answer the same 404.
7. Write `createBooking`. Decode a `bookingRequest` with the given `decode`. 404 for an unknown trip, 422 for a seat outside `1..t.Seats`. Take `Passenger` and `Phone` from the logged-in user. Answer 201 with a `Location` header.
8. Write `NewServer`: one `mux.Handle` per route with its `Require`, as in the brief, then wrap the whole mux in `Authenticate`.
9. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-059-roles-and-permissions/starter
```

## The kata: Permission Matrix

Answer "may this user do this?" from a list of grants, users and roles that inherit from other roles, including inheritance chains that loop. It's `Can` with role hierarchy, the way bigger systems store policy. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 59
```

## When you're stuck

- **Every route answers 401, even with a good token**: `Authenticate` isn't wrapping the mux, so no user is ever in the context. `NewServer` should return `Authenticate(auth, mux)`.
- **Every route answers 404**: `NewServer` still returns an empty mux. Register the routes.
- **`GET /bookings/2` as Otieno answers 200**: the role check passed (passengers may read bookings), and nothing checked whose booking it is. Call `CanViewBooking` after loading it.
- **`someone else's booking and a missing one must look the same`**: you used 403 or a different message for "not yours". Use the same 404 and the same text for both.
- **`WWW-Authenticate = ""` on the 401**: set the header before `writeError`. Once `WriteHeader` runs, header changes are ignored.
- **`cannot use s.createTrip (value of type func(...)) as http.Handler value`**: wrap it, `http.HandlerFunc(s.createTrip)`.
- **The booking has an empty passenger name**: you took the name from the request body. Use `UserFrom(r.Context())`.
- **Still stuck?** Open `solution/roles.go`, read it once, close it, then write your own version from memory.
