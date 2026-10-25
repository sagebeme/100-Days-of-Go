# Day 68: Talking to Go

## The brief

The route list moves to the server. A small Go API in `api/` answers `GET /api/routes?q=ngong`, and your React page searches it as the passenger types. Real networks are slow and sometimes fail, so the page needs three states: loading, error (with a Try again button), and results. And when the passenger types faster than the network answers, an old, slow answer must never replace a newer one.

```
GET /api/routes?q=ngong      →  200  [{"number":"111","operator":"MCA Sacco","boardsAt":"Railways","goesTo":"Ngong","fareKsh":100}]
GET /api/routes?q=mombasa    →  200  []
GET /api/routes?fail=1       →  500  {"error":"the route database is down"}

page loads          →  "Loading routes..."  then  "111 to Ngong, Ksh100" · "To Kibera, ask the conductor" · ...
server says 500     →  alert: "Couldn't load routes: the route database is down"  [Try again]
type "n", then "g"  →  requests for "", "n", "ng"; the first two are aborted; only "ng" is shown
```

The routes are real (checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026). Routes change, so confirm with the conductor.

## What you'll use

**`fetch`** returns a promise of a `Response`. It only rejects when there was no response at all (offline, DNS, CORS). A 404 or 500 *resolves*, so check `res.ok` yourself, like checking `resp.StatusCode` in Go.

```ts
const res = await fetch(`${API_URL}/api/routes?q=${encodeURIComponent(query)}`, { signal });
if (!res.ok) throw new ApiError(res.status, "...");
const data: unknown = await res.json();
```

**`useEffect`** runs code after React renders, again whenever a value in its dependency list changes. Return a cleanup function and React calls it before the next run and on unmount. An `AbortController` in the cleanup cancels the request this run started, which is the same idea as cancelling a `context.Context` in Go. The Go handler sees `r.Context().Done()` close and stops work.

```tsx
useEffect(() => {
  const controller = new AbortController();
  fetchRoutes(query, controller.signal).then(...).catch(...);
  return () => controller.abort();
}, [query]);
```

An aborted fetch rejects with an `AbortError`. That's not a failure to show the passenger, so check `controller.signal.aborted` in the `catch`. In development, `<StrictMode>` runs each effect twice on purpose, to catch effects with no cleanup. You'll see an aborted request in the Network tab; that's expected.

**State as a union.** `loading`, `error` and `done` can't be true together, so model them as `{ kind: "loading" } | { kind: "error"; message } | { kind: "done"; routes }`, not three booleans that can disagree.

**CORS.** The page comes from `http://localhost:5173` and the API from `http://localhost:8080`: different origins. The browser sends the request but hides the response from your JavaScript unless the API answers with `Access-Control-Allow-Origin: http://localhost:5173`. The Go side is the small `cors` function in `api/main.go`. CORS protects the *user's browser*, not your server: `curl` ignores it. Day 61 covers preflights and the rest; see also [http.md](../../reference/http.md).

## Why it matters at work

Most front-end bugs reported as "the page is broken" are unhandled network states: a spinner forever after a 500, an error shown for a request the page itself cancelled, or search results for the text before the last keystroke. Reviewers look for all three states and for cleanup in every effect that fetches. On the Go side, a CORS header set to `*` on an API that uses cookies is a finding in any security review; allow the exact origins you serve from, set from config.

## Steps

1. In `starter/api.ts`, write `fetchRoutes`: build the URL with `encodeURIComponent`, pass `{ signal }`, throw `ApiError` for a non-2xx answer (use the body's `error` if it's JSON with one, else `HTTP 404`), and return the array. Run the tests; `api.test.ts` should pass. The tests swap `fetch` for a fake with `vi.stubGlobal`, so no server is needed.
2. In `starter/App.tsx`, keep `query`, an `attempt` counter and a `status` union in state.
3. Write the `useEffect` above. Set `loading` at the start, `done` on success, `error` on failure unless the signal was aborted. Depend on `[query, attempt]`; Try again just bumps `attempt`.
4. Render the input labelled `Destination`, then exactly one of: `<p role="status">Loading routes...</p>`, `<div role="alert">` with the message and a Try again button, `No routes match.`, or `<ul aria-label="Routes">`.
5. Run the tests from `phase-6-frontend`:

```bash
npm test -- day-068
```

6. See it for real. In one terminal, from the repo root, start the API; in another, from `phase-6-frontend`, start the page and open http://localhost:5173:

```bash
go run ./phase-6-frontend/day-068-talking-to-go/api
npm run dev -- day-068-talking-to-go/starter
```

7. Break CORS on purpose: stop the API and start it with `ALLOWED_ORIGIN=http://localhost:3000`. The page shows "Failed to fetch" and the browser console says the request was blocked by CORS policy. `curl -i localhost:8080/api/routes` still works. Put it back.

## The kata: Latest Answer Wins

Given when each search request was sent and how long its answer took, work out what a page with abort shows, and what a page without it ends up showing. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 68
```

## When you're stuck

- **`Access to fetch at 'http://localhost:8080/...' from origin 'http://localhost:5173' has been blocked by CORS policy`**: the API isn't sending `Access-Control-Allow-Origin` for your origin. Check `ALLOWED_ORIGIN`, and that the page really is on port 5173.
- **`TypeError: Failed to fetch`** with no CORS message: the API isn't running, or it's on another port.
- **The spinner never goes away after a 500**: you only handled `.then`. A thrown `ApiError` needs a `.catch` that sets the error state.
- **"an abort is not an error" fails**: you show an alert for the `AbortError`. Return early in `catch` when `controller.signal.aborted`.
- **"the first request was aborted" fails**: your effect has no cleanup, or the controller is created outside the effect. Make a new one per run and abort it in the returned function.
- **`Unexpected token '<', "<!doctype "... is not valid JSON`**: you fetched a relative URL like `/api/routes`, so Vite answered with `index.html`. Use `API_URL`.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
