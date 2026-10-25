# Day 96: Discord Bot

## The brief

The course's study group, "Go Nairobi Juniors" (a made-up server), meets on Discord. Njeri, who runs it, wants a bot that posts a kata every morning, gives hints one at a time without spoiling them for everyone, and checks answers in private. She has one condition: "I want to change the rules without opening Discord to test them." So the logic must sit behind an interface, and Discord is only a thin layer on top.

```
/kata                         →  **Day 93: Time Travel Store** ... Answer with /submit day:93 answer:...   (everyone sees it)
/hint day:93                  →  Hint 1 of 3: Keep every version.              (only you see it)
/hint day:93                  →  Hint 2 of 3: Binary search the times.
/submit day:93 answer:3200    →  Not quite. Try /hint day:93.
/submit day:93 answer:4500    →  Correct! Day 93 solved with 2 hints.

POST /interactions  {"type":1}           (signed by Discord)  →  200 {"type":1}
POST /interactions  bad X-Signature-Ed25519                   →  401
```

**Acceptance criteria**

- All command rules live in `Bot.Handle(ctx, Command) Response` and are tested with plain values.
- Today's kata changes at midnight, local time, and cycles through every kata.
- Hints are counted per user and per kata. Hints and answers are ephemeral.
- The HTTP endpoint rejects any request whose Ed25519 signature doesn't verify, answers Discord's PING, and turns commands into `Command` values.
- No Discord library, and no token in the code.

## What you'll use

- Interfaces you satisfy without saying so, from [Day 16: Payment Methods](../../phase-2-modelling/day-016-payment-methods/).
- Config from env, never committed, from [Day 21: Config Loader](../../phase-2-modelling/day-021-config-loader/) and [env-and-config.md](../../reference/env-and-config.md).
- Handlers and `httptest`, from [Day 49](../../phase-5-backend/day-049-hello-server/) and [Day 62: Integration Tests](../../phase-5-backend/day-062-integration-tests/).
- Verifying a signed request before reading it, from [Day 95: Webhook Relay](../day-095-webhook-relay/).
- New today: `crypto/ed25519`. HMAC (Day 95) needs a shared secret. Ed25519 is public-key: Discord signs with a private key it keeps, and you verify with the public key from the developer portal. Nobody who reads your config can forge a request.

```go
msg := append([]byte(timestamp), body...)
ok := ed25519.Verify(publicKey, msg, sig) // sig is the hex-decoded X-Signature-Ed25519
```

## Steps

Requirements. The tests in `starter/` check every must-have; the doc comments give the exact wording.

**Must-haves**

1. `NewBot` and `Today`: sort the katas by day; pick `(YearDay-1) % len` in the clock's own location. A bot with no katas must not panic.
2. `Handle` for `/kata`, `/hint` and `/submit`, with the replies from its doc comment. `/kata` never shows the answer.
3. `SameAnswer`: compare word by word with `strings.Fields`, so extra spaces don't matter but case does.
4. `MemoryProgress`: two maps behind one mutex. The bot only talks to the `Progress` interface.
5. `VerifyRequest`: hex-decode the signature, check its length, and verify `timestamp + body`. Bad input returns `false`, never a panic.
6. `ParseCommand`: read `data.name`, `data.options` and the user from `member.user` or `user`. JSON numbers decode into `any` as `float64`, so write them back as `"93"`.
7. `NewInteractionsHandler`: the statuses in its doc comment. Verify the raw body before parsing it.
8. Run the tests:

```bash
go test ./phase-8-portfolio/day-096-discord-bot/starter
```

**Stretch goals**

- Finish `starter/cmd/katabot` and run it against Discord (below).
- `/leaderboard`, using the kata's ranking rules.
- Reject a timestamp more than 5 minutes old, as on Day 95. Discord doesn't require it, but it's cheap replay protection.
- Keep progress in Postgres or Redis behind the same `Progress` interface.
- Run real code submissions? Read about sandboxes first (gVisor, Firecracker, the Go Playground's design). Running strangers' code on your server is how servers get taken over, which is why `/submit` checks an answer instead.

### Wiring it to Discord

1. Create an application at the [Discord developer portal](https://discord.com/developers/applications). From **General Information**, copy the **Application ID** and **Public Key**. From **Bot**, reset and copy the **token**. Put them in `.env` (already in `.gitignore`), never in code:

   ```bash
   DISCORD_APPLICATION_ID=...
   DISCORD_PUBLIC_KEY=...
   DISCORD_BOT_TOKEN=...
   ```

2. Register the commands once with Discord's REST API. Use a server (guild) command while testing, because it updates at once:

   ```bash
   curl -X PUT "https://discord.com/api/v10/applications/$DISCORD_APPLICATION_ID/guilds/$GUILD_ID/commands" \
     -H "Authorization: Bot $DISCORD_BOT_TOKEN" -H "Content-Type: application/json" \
     -d '[{"name":"kata","description":"Show a kata","options":[{"type":4,"name":"day","description":"Day number"}]},
          {"name":"hint","description":"Get the next hint","options":[{"type":4,"name":"day","description":"Day number","required":true}]},
          {"name":"submit","description":"Check your answer","options":[{"type":4,"name":"day","description":"Day number","required":true},{"type":3,"name":"answer","description":"Your answer","required":true}]}]'
   ```

   Option type 4 is an integer and 3 is a string.

3. **Interactions endpoint (what you built).** Run `katabot` somewhere Discord can reach over HTTPS: a deploy, or a tunnel such as `cloudflared tunnel --url http://localhost:8080`. Paste `https://<host>/interactions` into **Interactions Endpoint URL** and save. Discord sends a PING, plus requests with bad signatures, and saves the URL only if you answer PONG and reject the bad ones with `401`. Reply within 3 seconds; for slower work, reply with type 5 (deferred) and send the message later.
4. **The other way: the gateway.** A bot can instead keep a WebSocket open to Discord's gateway, send an IDENTIFY with the bot token, keep the heartbeat going, and receive `INTERACTION_CREATE` events. It then answers each one with `POST /interactions/{id}/{token}/callback`. That's more code and a process that never sleeps, but it also receives messages and reactions. Either way, `Bot.Handle` doesn't change. That's the point of the interface.
5. Invite the bot with an OAuth2 URL from the portal, with the `applications.commands` scope (add `bot` for the gateway).

## The kata: Kata Leaderboard

Process solve events and `TOP k` queries, ranking users by katas solved, then fewer hints, then name. A leaderboard with a multi-key sort is a common interview warm-up, and it's the `/leaderboard` stretch goal. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 96
```

## When you're stuck

- **Discord says "The specified interactions endpoint url could not be verified"**: your handler doesn't return `401` for a bad signature, doesn't answer `{"type":1}` to a PING, or isn't on HTTPS. The handler tests cover the first two.
- **Every signature fails, even in the tests**: you signed or verified the parsed JSON, not the raw bytes, or the body alone. The message is the timestamp header followed by the exact body bytes.
- **`panic: ed25519: bad public key length`**: `ed25519.Verify` panics on a key that isn't 32 bytes. Check `len(publicKey)` first.
- **`Options["day"]` is `"93.000000"` or empty**: a JSON number in an `any` is a `float64`. Format it with `strconv.FormatFloat(v, 'f', -1, 64)`.
- **`Today` is off by one around midnight**: use `now().YearDay()` on the time as given. Don't convert to UTC first, or 23:30 in Nairobi becomes the same day's 20:30 UTC while 01:00 becomes the day before.
- **Hints leak between users**: key the progress maps on both user and day, for example `user + "/" + strconv.Itoa(day)` or a struct key.
- **Still stuck?** Open `solution/bot.go` and `solution/interactions.go`, read them once, close them, then write your own version from memory.

## Make it portfolio-ready

- **README**: a screenshot or short GIF of the bot in a test server, the three commands, and a diagram: Discord → `/interactions` → `VerifyRequest` → `ParseCommand` → `Bot.Handle`.
- **Deploy**: run `katabot` on a free tier (it's one small binary in a Docker image, Day 63), set `DISCORD_PUBLIC_KEY` in the host's settings, and add a `/healthz` check. Put the invite link in the README so people can try it.
- **Tests and CI**: `go test -race ./...` in GitHub Actions and the badge. Note in the README that every command rule is tested without Discord, and how.
- **What you'd do next**: progress is lost on restart, so name the store you'd use. Mention rate limiting `/submit` (Day 47) so nobody brute-forces answers, and the deferred response for anything slower than 3 seconds.
