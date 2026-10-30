# Day 70: Login Flow

## The brief

Passengers want to see their bookings, and only theirs. Build the browser half of the token scheme from [Day 58](../../phase-5-backend/day-058-sessions-and-jwts/). Logging in gives a short-lived **access token**, which the page keeps in memory and sends with every API call, and a long-lived **refresh token**, which the Go server puts in an `HttpOnly` cookie that JavaScript can't touch. When the access token expires, the page quietly gets a new one. When the refresh token dies too, the passenger goes back to the login form with a clear message.

```
auth.login("0712345678", "matatu-2026")   →  { phone: "0712345678", name: "Amina" }   (token in memory, cookie set)
auth.login("0712345678", "1234")          →  throws LoginError "Wrong phone number or password"
auth.apiFetch("/api/me/bookings")         →  GET with Authorization: Bearer tok1                 →  200
  ... a minute later                      →  401 → POST /api/refresh → retry with Bearer tok2     →  200
  ... refresh cookie revoked              →  401 → refresh 401 → onExpired() → SessionExpiredError
reload the page, auth.restore()           →  POST /api/refresh with the cookie → logged in again
auth.logout()                             →  POST /api/logout, cookie cleared, token forgotten
```

The demo account and its password are made up and live only in `api/main.go`.

## What you'll use

**Where tokens live.** Anything in `localStorage` or `sessionStorage` can be read by any script on the page, so one XSS bug (or one bad npm package) sends a 30-day refresh token to an attacker. Keep the refresh token in a cookie with `HttpOnly` (scripts can't read it), `Secure` (HTTPS only) and `SameSite=Strict` (not sent from other sites). Keep the short access token in a JavaScript variable. A script running in your page could still use it while the page is open, which is why it lasts only minutes. After a reload it's gone, so `restore()` swaps the cookie for a new one.

```ts
class Auth {
  #token: string | null = null;   // a real private field: not on the object for anyone to read
}
fetch(url, { credentials: "include" }); // send and accept cookies on a cross-origin request
```

**Credentials across origins.** The page (`localhost:5173`) and API (`localhost:8080`) are different origins but the same *site*, so a `SameSite=Strict` cookie still travels. The request needs `credentials: "include"`, and the server must answer `Access-Control-Allow-Credentials: true` with its exact origin. `Access-Control-Allow-Origin: *` is refused by the browser when credentials are involved.

**401 handling in one place.** `apiFetch` adds the header, and on a 401 it refreshes and retries once. If several requests fail together, they must share a single refresh: the server rotates refresh tokens (each works once), so two parallel refreshes would make one of them fail and log the passenger out. Keep the running refresh in a promise and hand the same one to every caller.

**Protected routes.** `RequireAuth` shows its children only when there's a user. That's for the passenger's convenience, not security: anyone can edit the JavaScript in their browser. The Go API's 401 on `/api/me/bookings` is the real protection.

## Why it matters at work

Token storage comes up in every front-end security review, and "refresh token in `localStorage`" is a finding that blocks a release (see [owasp-checklist.md](../../reference/owasp-checklist.md)). So is a page that shows protected data based only on a client-side flag. The parallel-refresh race is a classic production bug: it only shows up when a page loads several things at once after a laptop wakes from sleep, and it looks like random logouts.

## Steps

1. In `starter/auth.ts`, write `login`: POST JSON with `credentials: "include"`, throw `LoginError` on 401, store `accessToken` in `#token`, return `user`.
2. Write `refresh` as a single flight: if `#refreshing` is set, return it; otherwise start the request, store the promise, and clear it in `.finally`. A failed refresh sets `#token` to `null` and resolves to `null`; it doesn't throw.
3. Write `apiFetch`: build `Headers` from `init.headers`, add `Authorization: Bearer ...`, fetch `API_URL + path`. On 401, `await this.refresh()`. If that gives a user, retry once. If not (or the retry is 401 too), clear the token, call `this.onExpired()`, and throw `SessionExpiredError`.
4. Write `logout`: POST `/api/logout`, and clear `#token` in `finally` so an offline logout still logs out locally. Run the tests: `auth.test.ts` should pass.
5. In `starter/App.tsx`, `RequireAuth` returns `children` or `fallback`.
6. `App`: on mount, set `auth.onExpired` and call `auth.restore()`, showing `Checking your session...` until it answers. Then render the login form or the logged-in view inside `RequireAuth`. Make the "session ended" handler with `useCallback(..., [])` so effects that use it don't re-run.
7. The login form clears the password after a failure. Use `type="password"` and `autoComplete="current-password"` so password managers work.
8. Run the tests from `phase-6-frontend`:

```bash
npm test -- day-070
```

9. Try it for real with `go run ./phase-6-frontend/day-070-login-flow/api` (repo root) and `npm run dev -- day-070-login-flow/starter`. Log in as `0712345678` / `matatu-2026`. In DevTools, open Application, then Cookies: the `refresh` cookie is ticked HttpOnly, and `document.cookie` in the console doesn't show it. Wait a minute and reload the bookings: the Network tab shows a 401, a refresh and a retry. Use Chrome or Firefox: they treat `http://localhost` as secure, so the `Secure` cookie is kept. Other browsers may drop it.

## The kata: Shared Refresh

Simulate a client that shares one refresh between all the requests that hit an expired token. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 70
```

## When you're stuck

- **The cookie never appears and every refresh is 401**: the login request needs `credentials: "include"`, and so does every request after it. Without it, the browser ignores `Set-Cookie` on a cross-origin response.
- **`The value of the 'Access-Control-Allow-Origin' header in the response must not be the wildcard '*' when the request's credentials mode is 'include'`**: the server must name the exact origin and send `Access-Control-Allow-Credentials: true`.
- **"number of POST /api/refresh" is 2**: each failing request started its own refresh. Store the promise in a field and return it while it's running.
- **The bookings load forever after the session ends**: a failed refresh must resolve to `null`, and `apiFetch` must then throw, so the page can react.
- **The bookings request runs again and again**: the function you pass to an effect's dependency list is new on every render. Wrap it in `useCallback`.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
