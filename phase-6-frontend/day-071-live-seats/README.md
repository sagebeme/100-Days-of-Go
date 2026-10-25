# Day 71: Live Seats

## The brief

Two passengers look at the last free seat on the 58 to Buruburu. Both tap it. Without live updates, both see "free" and one of them gets a nasty surprise at the stage. Today the Go server pushes every seat change over a **WebSocket**, so every open page redraws within a moment. Phones drop connections all the time on the move, so the page reconnects by itself, waiting a little longer after each failure.

```
connect ws://localhost:8080/ws/seats
  ←  {"type":"snapshot","version":7,"taken":[false,true,false,...]}     the whole map, once
  →  {"type":"hold","seat":3}                                           you tap seat 3
  ←  {"type":"seat","version":8,"seat":3,"taken":true}                  every page gets this
  →  {"type":"hold","seat":3}                                           someone else taps it too
  ←  {"type":"error","message":"Seat 3 is already taken"}               only they get this

connection drops  →  "Reconnecting..."  →  retry after 1 s, 2 s, 4 s, 8 s, 10 s, 10 s ...
reconnects        →  "Live", a fresh snapshot, and the wait goes back to 1 s
```

The seat map and the other "passengers" are simulated.

## What you'll use

**A WebSocket** starts as an HTTP request and turns into a two-way connection that stays open. Either side can send a message at any time, so the server can push changes instead of the page asking every few seconds.

```ts
const ws = new WebSocket("ws://localhost:8080/ws/seats");
ws.onopen = () => ws.send(JSON.stringify({ type: "hold", seat: 3 }));
ws.onmessage = (ev) => console.log(ev.data);   // a string
ws.onclose = () => { /* the browser never reconnects for you */ };
```

**On the Go side**, `github.com/coder/websocket` does the handshake. Each connection gets two goroutines: one reads holds, one writes updates from a channel. Only the writer writes, so messages never interleave. A slow client's channel fills up, so the hub drops that client instead of blocking everyone; it reconnects and gets a fresh snapshot. See `api/main.go`.

**Origin checks are yours.** CORS doesn't apply to WebSockets: any page on any site can open one to your server, with the user's cookies. `websocket.Accept` refuses origins not in `OriginPatterns`, which here comes from `ALLOWED_ORIGIN`. Never use `InsecureSkipVerify` in production.

**Reconnecting with backoff.** Retrying at once in a loop hammers a server that's already struggling, and thousands of phones doing it together keep it down. Wait 1 s, then 2 s, 4 s and so on up to a cap, and reset after a good connection. Real clients also add a little randomness (jitter) so they don't all retry in step; it's left out here so the tests can check exact times.

**Versions.** Every change has a version number. The page ignores an update that isn't newer than what it has. Within one connection messages arrive in order, so a single version for the whole map is enough here, and every reconnect starts again from a snapshot. The kata looks at the harder case.

**`useRef`** keeps the connection across renders without causing a re-render, so a click handler can call `conn.current.send(...)`.

## Steps

1. In `starter/seats.ts`, write `parseMessage`: `JSON.parse` in a `try`, then check `type` and each field's `typeof`. Anything odd gives `null`. Never trust data from the network, even your own server's.
2. Write `applyMessage` with a `switch` on `msg.type`. Return new objects and arrays; copy `taken` with `[...state.taken]` before changing a seat.
3. In `starter/socket.ts`, write `backoff`: `Math.min(1000 * 2 ** attempt, 10_000)`.
4. Write `connectSeats`. Inside, a `start()` function makes a socket and sets its three handlers. `onopen`: mark open, reset `attempt`, report `live`. `onclose`: unless `close()` was called, report `reconnecting` and `setTimeout(start, backoff(attempt++))`. `close()` sets a `stopped` flag, clears the timer and closes the socket. The tests use a fake socket and fake timers, so they run instantly.
5. In `starter/App.tsx`, call `connect` in a `useEffect` with `[]`, store the result in a ref, and return a cleanup that closes it. Use `setSeats((s) => applyMessage(s, msg))`: the function form, because messages arrive between renders.
6. Render the status, the free count and one button per seat. Here `key={i}` is right: the index *is* the seat number and never moves.
7. Run the tests from `phase-6-frontend`:

```bash
npm test -- day-071
```

8. Watch it live: `go run ./phase-6-frontend/day-071-live-seats/api` (repo root) and `npm run dev -- day-071-live-seats/starter`. Open the page in two windows and tap a seat in one. Stop the API with Ctrl+C: both pages say Reconnecting. Start it again and they come back on their own. In DevTools, Network, then WS, you can read every message.

## The kata: Last Write Wins

Updates arrive out of order; keep the newest version of each seat. Instructions are in `kata/README.md`.

```bash
go run ./tools/kata 71
```

## When you're stuck

- **`WebSocket connection to 'ws://localhost:8080/ws/seats' failed`** and the API logs nothing: the API isn't running, or the URL says `http://` instead of `ws://`.
- **The API answers 403**: the page's origin isn't the one in `ALLOWED_ORIGIN`. Check the port.
- **Two connections open in development**: `<StrictMode>` mounts, unmounts and mounts again. With a cleanup that calls `close()`, the first one closes at once. Without one, you leak a socket.
- **The page reconnects after you leave it**: `onclose` also fires when *you* close. Check the `stopped` flag before scheduling a retry.
- **Seats flicker back to an old state**: you applied an update without checking its version, or changed `state.taken` in place.
- **"no new socket before 1 s" fails**: you called `start()` straight from `onclose`. Schedule it with `setTimeout`.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.
