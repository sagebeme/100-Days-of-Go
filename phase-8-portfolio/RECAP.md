# Phase 8 recap: Portfolio

Days 85 to 100. No walkthroughs: a brief, acceptance criteria, tests, and you. Sixteen projects, each a small version of something companies really run, then one of them shipped properly.

## What the phase covered

| Days | Projects |
|---|---|
| 85 to 87 | a link shortener with SSRF checks; a Kiswahili Wordle; a chat server with rooms and a hub |
| 88 to 90 | invoice PDFs written by hand; a static site generator; a rate-limited reverse proxy |
| 91 to 93 | a matatu route finder with Dijkstra; a streaming log analyser; a crash-safe key-value store |
| 94 to 96 | a job board with search and CSRF protection; a webhook relay; a Discord bot with Ed25519 |
| 97 to 99 | fixing an open-source issue; making a function 10x faster with pprof; a mock interview |
| 100 | the capstone: ship one project people can use, and a portfolio generated from one file |

## Key ideas

1. **Read the brief twice and turn it into tests.** Acceptance criteria are the spec; when the brief is vague, write down the decision you made.
2. **Untrusted input is the default.** URLs (SSRF), names (XSS), paths (`..`), headers (`X-Forwarded-For`), bodies (size, unknown fields), signatures (verify the raw bytes first). Most of Phase 8's bugs live here.
3. **Never block on someone else.** A slow client, a slow upstream or a slow subscriber gets a buffer, a deadline or a queue, so it can't freeze everyone else.
4. **Durability means order.** Write, sync, then say "saved"; write to a temporary file and rename; truncate a torn record instead of reading it.
5. **Measure before you optimise, and keep the results identical.** Benchmarks, `benchstat` and pprof decide; correctness tests and allocation guards keep you honest.
6. **Someone else's code: reproduce, find the cause, fix the smallest thing.** A failing test first, one idea per PR, the project's style, not yours.
7. **Interviews reward the thinking out loud.** Repeat the problem, ask about the edges, say the brute force, name the pattern, test by hand, state the complexity.
8. **Shipped beats perfect.** One deployed project with a README, a demo, tests and a green badge is worth more than sixteen folders.

## Self-test

1. Your link shortener checks that a URL's host isn't a private IP literal. Name two ways an attacker still reaches `127.0.0.1`.
2. Why does the chat hub use `select` with `default` when it sends to a client?
3. In the key-value store, why does `Compact` write to `data.wal.tmp` and rename it, instead of rewriting `data.wal` in place?
4. The webhook relay checks the event ID for duplicates before checking the signature. What can an attacker do?
5. What's wrong with "I made it faster" in a PR description for Day 98, and what should it say instead?
6. A hiring manager has two minutes. What must the first screen of your portfolio show?

<details>
<summary>Answers</summary>

1. A name that resolves to it (like `localtest.me`), and DNS rebinding (a name that resolves to a public address at check time and a private one later). Resolve and check every address, ideally at dial time.
2. So the hub never waits on a slow client: a full buffer means disconnect or drop, by policy, and the rest of the room carries on.
3. Rewriting in place destroys the only copy on disk while it works; a crash leaves an empty or partial log. A rename is atomic: a crash leaves the old log or the new one.
4. Send a forged copy first: it gets recorded as seen, and the real event is then answered "duplicate" and never forwarded. Verify first, and only remember IDs that were accepted.
5. It gives no evidence. Show `benchstat` before and after (time, bytes and allocations, with p-values), and say the correctness tests still pass.
6. Your name, the role you want and where you are, then two or three featured projects, each with a live link, the code and one sentence on the problem it solves.

</details>

## Build it without steps: Your Own Brief

Every Phase 8 day started with someone's problem. Now find your own: a duka owner, a chama, a school, a church, a friend's small business, or your own daily annoyance. Write the brief yourself, then build it.

Done when:

- `BRIEF.md` has the story in a paragraph, an example of input and output like the briefs in this phase, five to eight acceptance criteria that a test could check, and what's out of scope. The person you built it for has read it and agreed.
- Every acceptance criterion has at least one test, written before the code that makes it pass.
- A short threat list names the untrusted inputs and what you do about each.
- It runs somewhere other people can use it (a URL, a release binary, or a bot in a real server), with config from env and no secrets in the repository or its history.
- CI runs `gofmt`, `go vet` and `go test -race` and the badge is green.
- The README is written for a stranger with two minutes: what it is, the live link, a screenshot, how to run it in three commands, and the trade-offs you made.
- It's in your `projects.json` from Day 100, with a CV line that starts with a verb and names the hard part.
