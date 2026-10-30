# Day 76: Booking App

## The brief

This is the Phase 6 capstone. Twende needs its booking website: a passenger at Ambassadeur opens it on their phone, sees the next departures, picks the 58 to Buruburu, chooses a free seat, logs in if they haven't, fills in a name and a phone number for the SMS, and gets a booking reference. Everything that can go wrong on a real network will: trips that won't load, a seat someone else grabs a second earlier, a session that expires halfway through.

There's no walkthrough today. You get the API, the contract, the tests and some hints. Everything you need is from Days 65 to 75.

```
GET  /api/trips              →  58 to Buruburu, 07:45 from Ambassadeur, Ksh150, 3 seats left   [Book the 58 at 07:45]
                                111 to Ngong, 07:50 from Railways, Ksh100, full                 [Book the 111 at 07:50] (disabled)
GET  /api/trips/T1/seats     →  {"tripId":"T1","seatCount":14,"taken":[1,2,4,...]}
POST /api/login              →  200 {"accessToken":"...","user":{"phone":"0712345678","name":"Amina"}} + refresh cookie
POST /api/bookings           →  201 {"reference":"K7Q2M9XA","tripId":"T1","seat":3,"phone":"254712345678","totalKsh":150}
                             →  409 {"error":"Seat 3 was just taken"}
                             →  422 {"errors":{"phone":"Enter a Kenyan mobile number like 0712 345 678"}}
                             →  401 → refresh → retry, or back to the login page

Booked!
Reference K7Q2M9XA. Seat 3 on the 58 to Buruburu at 07:45. Ksh150. We'll SMS 254712345678.
```

The Go API in `api/` is Phase 5's booking API cut down to run in memory. Log in as `0712345678` with `matatu-2026` (a made-up demo account). The routes are real (checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026); the times and seats are made up. Routes change, so confirm with the conductor.

## What you'll use

- TypeScript types and unions for the API's shapes and results (Day 65)
- Components, props, state, lists and keys (Days 66 and 67)
- `fetch`, loading and error states, `useEffect` with cleanup (Day 68)
- A controlled form, validation mirrored from Go, 422 field errors (Day 69)
- An access token in memory, a refresh cookie, one shared refresh, 401 handling (Day 70)
- Labels, a radio group for seats, `aria-describedby`, focus on each new page (Day 72)
- Fakes at the edges: `fakeFetch` for the client, a fake `Client` for the App (Day 73)

`starter/api.ts` already has every type and error class; `starter/fakeFetch.ts` is Day 73's helper, finished. The tests use both, so keep their names.

## Steps

Requirements. Meet them in any order; the tests in `starter/` check each one.

1. **`validate.ts`**: `normalisePhone` and `validateBooking` follow the Go API's rules and messages exactly: name trimmed, 1 to 50 characters (`Enter your name`, `Name is too long (50 characters at most)`); phone as on Day 69 (`Enter a Kenyan mobile number like 0712 345 678`); a seat chosen (`Choose a seat`).
2. **`api.ts`, `createClient(baseUrl)`** returns a `Client`. Every request uses `credentials: "include"`.
   - `listTrips` and `getSeats(id)` return the JSON, and throw `ApiError(status, message)` for a non-2xx answer, using the body's `error` when there is one.
   - `login` keeps the access token in memory only and returns the user; a 401 throws `LoginError`. `restore` swaps the refresh cookie for a token (`null` if that fails). `logout` posts and forgets the token.
   - `book` sends `Authorization: Bearer <token>` and returns `{ kind: "booked" }` for 201, `{ kind: "invalid", errors }` for 422 and `{ kind: "conflict", message }` for 409, and throws `ApiError` otherwise. On a 401 it refreshes once and retries. If the refresh fails, it calls `onExpired()` and throws `SessionExpiredError`.
3. **The trips page** (the first page): `Loading trips...` while loading; a `<ul aria-label="Trips">` with one item per trip reading `58 to Buruburu, 07:45 from Ambassadeur, Ksh150, 3 seats left` (`1 seat left`, or `full`) and a button `Book the 58 at 07:45`, disabled when full. On failure, a `role="alert"` with `Couldn't load trips: MESSAGE` and a `Try again` button.
4. **A trip's page**: an `<h2>` like `58 to Buruburu at 07:45`, a `Back to trips` button, and a form with a `<fieldset>` named `Choose a seat` holding one radio per seat (`Seat 3`, or `Seat 2 (taken)` and disabled), then `Name` and `Phone` inputs. When logged in, the submit button reads `Book seat, Ksh150`; when logged out it's a `Log in to book` button instead.
5. **Logging in**: an `<h2>Log in</h2>` page with `Phone`, `Password` (`type="password"`) and a `Log in` button. Success returns to the page the passenger came from. Failure shows the error in `role="alert"` and clears the password. The header shows a `Log in` button, or `Habari, NAME` and `Log out`.
6. **Booking**: the browser's checks run first, and nothing is sent while they fail. Each field error is in an element its input (or the seat fieldset) points at with `aria-describedby`, and invalid inputs have `aria-invalid`. Send the name trimmed and the phone as typed. A 422 shows the server's messages the same way. A 409 shows `Seat 3 was just taken. Pick another seat.` in `role="alert"` and loads the seat map again. Anything else shows `Booking failed: MESSAGE`.
7. **Confirmation**: an `<h2>Booked!</h2>`, `Reference K7Q2M9XA. Seat 3 on the 58 to Buruburu at 07:45. Ksh150. We'll SMS 254712345678.`, and a `Book another trip` button back to the list.
8. **The session**: on load, `client.restore()` logs a returning passenger in. Set `client.onExpired` so that an expired session, from any request, logs the passenger out and shows the login page with `Your session has ended. Please log in again.`
9. **Focus**: each page's `<h2>` takes focus when the page appears (`tabIndex={-1}` and a ref).
10. Run the tests from `phase-6-frontend`, then use it for real with the Go API (repo root) and the dev server:

```bash
npm test -- day-076
go run ./phase-6-frontend/day-076-booking-app/api
npm run dev -- day-076-booking-app/starter
```

Book the last seats on the 58 in two browser windows to see the 409, and start the API with `ACCESS_TTL=20s` to watch a token refresh.

## The kata: Seats Together

Seat groups of passengers side by side in a matatu's rows, first come first served. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 76
```

## When you're stuck

- **Where do I start?** `validate.ts`, then `api.ts` (copy the shape of Day 70's `Auth`), then the App one page at a time, in the order of the requirements. Run `npm test -- day-076` after each.
- **Several pages in one component**: keep a `view` in state as a union, `{ name: "trips" } | { name: "trip"; trip } | { name: "login"; back } | { name: "done"; trip; booking }`, and `switch` on it. Storing `back` in the login view is how you return to the right page.
- **`Found multiple elements with the role "button" and name "Log in"`**: hide the header's `Log in` button while the login page is showing.
- **The trip page keeps an old seat map after a 409**: keep a reload counter in state and add it to the seat-loading effect's dependencies.
- **Name and phone should be filled in after logging in**: give the trip page a `key` that changes when the user does, and start its state from the user.
- **"the seat map is fetched again" counts 3**: the effect ran twice for one change. Check its dependencies don't include a function or object made fresh on every render.
- **`act(...)` warnings in the output**: a state update landed after the test finished. Await the thing the test is waiting for (`findBy...`), and ignore updates from effects that were cleaned up.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
