# Day 87: Chat Server

## The brief

Baraka manages a matatu sacco with routes 111 to Ngong and 125 to Ongata Rongai. His drivers and stage marshals coordinate on a dozen WhatsApp groups, and messages get lost. He wants one small chat service: a room per route, everyone in a room sees every message, and one driver with a dying phone on a weak network must never slow the room down for everyone else.

```
GET /ws?room=route-125&name=otieno      →  101 Switching Protocols
  server → otieno:   {"type":"join","room":"route-125","from":"otieno"}
GET /ws?room=route-125&name=wanjiku     →  101
  server → both:     {"type":"join","room":"route-125","from":"wanjiku"}
  otieno → server:   {"text":"Mat ya 125 imefika Railways"}
  server → both:     {"type":"message","room":"route-125","from":"otieno","text":"Mat ya 125 imefika Railways"}
GET /ws?room=route-125&name=two words   →  400
GET /ws (Origin: https://evil.example)  →  403
  wanjiku's buffer fills up             →  wanjiku is closed with 1008 "too slow"
```

Acceptance criteria:

- One hub goroutine owns the rooms. Handlers talk to it over channels, never through a shared map.
- A message reaches everyone in its room, the sender included, and no one outside it.
- Joins and leaves are announced. Empty rooms are deleted.
- The hub never waits on a client. A full buffer means disconnect or drop, chosen by policy.
- The server sets the sender's name and room itself, ignores empty or over-long messages, limits frame size, and refuses foreign origins.
- Shutting the hub down closes every socket with `1001 Going Away`.

## What you'll use

- Goroutines, channels and `select`: [Day 41](../../phase-4-concurrency/day-041-goroutines/), [Day 42](../../phase-4-concurrency/day-042-channels/), [Day 43](../../phase-4-concurrency/day-043-select-timeouts/).
- Context to stop the hub and time out writes: [Day 44](../../phase-4-concurrency/day-044-context-cancel/).
- `sync/atomic` for a counter read from another goroutine: [Day 45](../../phase-4-concurrency/day-045-mutexes/).
- `httptest.NewServer`, a real server on a random port: [Day 49](../../phase-5-backend/day-049-hello-server/).
- New today: `github.com/coder/websocket`. `Accept` upgrades the request, `wsjson.Read` and `wsjson.Write` move JSON, `SetReadLimit` caps a message.

```go
select {
case c.send <- m: // room in the buffer
default: // full: never wait, apply the policy
}
```

## Why it matters at work

Two bugs show up in every first chat server. The first is one slow client blocking the broadcast loop, so a single phone in a tunnel freezes the room for everyone: that's why the hub uses `select` with `default`. The second is cross-site WebSocket hijacking. A page on another site can open a WebSocket to your server, the browser may send your cookies with it, and CORS doesn't apply. The Origin check is what stops another site from opening a socket as your logged-in user. `coder/websocket` checks the Origin by default. Reviewers look for `InsecureSkipVerify: true` and ask why.

## Steps

Requirements. Meet them in any order. The tests in `starter/` check each one.

Must-haves:

1. `NewClient` makes the buffered channel. `Messages` and `Dropped` read it and the atomic counter.
2. `NewHub` and `Run`: one `for`/`select` loop over join, leave, broadcast, query and `ctx.Done()`. Rooms are a `map[string]map[*Client]bool` that only `Run` touches.
3. Delivery never blocks. Under `Disconnect`, a full client is removed and its channel closed. Under `DropMessages`, the message is skipped and counted.
4. `Join`, `Leave`, `Broadcast` and `Rooms` send to `Run` and never block once the hub has stopped. `Leave` twice must not close a channel twice.
5. On shutdown, `Stopped` is true before any client's channel is closed.
6. `ValidName` and `GET /ws`: `400` before upgrading for a bad room or name. After `Accept`, set the read limit and join the hub.
7. Two goroutines per connection. The read loop decodes `{"text":...}`, trims it, skips empty or over-`MaxText` messages and broadcasts. The write loop sends from `Messages()` with `WriteTimeout`, and closes the socket with `StatusPolicyViolation` (slow) or `StatusGoingAway` (shutdown) when the channel closes.
8. `cmd/chat` serves `/ws` and shuts down on Ctrl+C.
9. Run the tests from the repo root, with the race detector too:

```bash
go test ./phase-8-portfolio/day-087-chat-server/starter
go test -race ./phase-8-portfolio/day-087-chat-server/starter
```

Stretch goals:

- Unique names per room: `Join` replies with an error when the name is taken.
- Message history: the last 20 messages of a room are sent to anyone who joins.
- Pings every 30 seconds with `conn.Ping`, to find dead connections that never closed.
- A cap on rooms and members, and a per-client rate limit on messages.
- Run two instances behind a load balancer with Redis pub/sub between them ([Day 78](../../phase-7-production/day-078-caching/) has Redis).

## The kata: Delivery Count

Replay a log of joins, leaves and messages, and say how many people each message reached. It's the hub's bookkeeping as a stdin problem. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 87
```

## When you're stuck

- **`fatal error: concurrent map writes`**: a handler touched the rooms map. Only `Run` may. Everything else sends on a channel.
- **`panic: close of closed channel`**: a kicked client then left, and both paths closed its channel. Check the client is still in the room before you remove it.
- **The whole room freezes when one test client stops reading**: you wrote `c.send <- m` without `select`/`default`. That blocks the hub.
- **A test hangs in `Join` after the hub stopped**: the send has nobody to receive it. `select` on the hub's `done` channel too.
- **`failed to accept WebSocket connection: request Origin "evil.example" is not authorized for Host "..."`**: that's the Origin check working. In the real app, list your front end's host in `origins`.
- **Messages show up with `"from":"otieno"` that otieno never sent**: you trusted a `from` field in the client's JSON. Decode only `text` and fill in the name from the connection.
- **`failed to read: read limited at 4097 bytes`**: a client sent more than `MaxFrame`. The library closes that connection for you, with `1009 Message Too Big`.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- Put it in its own repo with a README covering the message format, the hub design (a small diagram of handlers, channels and the hub) and the slow-client policy.
- Record a demo GIF of two browser tabs on `cmd/chat`'s test page chatting in the same room.
- Deploy it on a host that supports WebSockets (Fly.io and Render do) and link the live page. Use `wss://` behind HTTPS.
- Write "What I'd do next": history, presence, Redis pub/sub for more than one instance, a React client ([Day 71](../../phase-6-frontend/day-071-live-seats/)).
- Add GitHub Actions running `go test -race` and show the CI badge.
