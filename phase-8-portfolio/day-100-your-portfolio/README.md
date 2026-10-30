# Day 100: Your Portfolio

## The brief

The client today is you. A hiring manager in Nairobi, Kigali or Berlin opens your CV, clicks one link, and spends about two minutes deciding whether to talk to you. What they find there has to be one project they can try and a page that makes the rest easy to scan. Sixteen half-finished folders won't do it.

So today has two parts. First, you pick one project from the course and ship it properly, so a stranger can use it. Second, you build a small Go tool that keeps your portfolio in one `projects.json` file and generates the portfolio page, the README badge table and your CV lines from it, so they never drift apart.

```
$ go run ./phase-8-portfolio/day-100-your-portfolio/starter/cmd/portfolio -in projects.json -out site
wrote site/index.html
wrote site/BADGES.md
wrote site/CV.md

BADGES.md:
| Project | Day | CI | Demo |
|---|---|---|---|
| [Webhook Relay](https://github.com/you/webhook-relay) | 95 | [![CI](.../actions/workflows/ci.yml/badge.svg)](.../actions/workflows/ci.yml) | [demo](https://...) |

CV.md:
- Webhook Relay: Built an HMAC-verified webhook relay in Go with replay protection... (https://github.com/you/webhook-relay)
```

**Acceptance criteria**

- One course project lives in its own public repository, deployed, with a README, a demo, tests and a green CI badge.
- `projects.json` is validated: every problem is reported at once, a typo in a field name is an error, and only `https://github.com/<owner>/<repo>` repos and `https://` demos get in.
- The page is generated with `html/template`, featured projects first, and works on a phone.
- The badge table and the CV lines come from the same file and the same order.

## What you'll use

- `encoding/json` and struct tags, from [Day 23: Save and Load](../../phase-2-modelling/day-023-save-load/).
- `errors.Join` to report every problem at once, from [Day 21: Config Loader](../../phase-2-modelling/day-021-config-loader/).
- Flags and exit codes, from [Day 22: CLI Flags](../../phase-2-modelling/day-022-cli-flags/).
- `html/template`, from [Day 94: Job Board](../day-094-job-board/).
- A Dockerfile and GitHub Actions, from [Day 63: Docker and CI](../../phase-5-backend/day-063-docker-and-ci/).
- Every project you've built. The best candidates to ship are the capstones and the Phase 8 projects: [Matatu Booking API](../../phase-5-backend/day-064-matatu-booking-api/), [Link Shortener](../day-085-link-shortener/), [Chat Server](../day-087-chat-server/), [Rate-Limited Proxy](../day-090-rate-limited-proxy/), [Matatu Route Finder](../day-091-matatu-route-finder/), [Key-Value Store](../day-093-key-value-store/), [Job Board](../day-094-job-board/), [Webhook Relay](../day-095-webhook-relay/) and [Discord Bot](../day-096-discord-bot/).

## Steps

Requirements, in two parts. Part A is the one that gets you hired; do it even if you run out of time for Part B.

**Part A: ship one project**

1. **Choose.** Pick the project you can explain best and that someone can use in under a minute. A deployed web service with a URL beats a library; a library with a clear README beats a CLI nobody can install. Pick the one that matches the job you want: backend roles want the API, the relay or the KV store.
2. **Move it into its own repository**, with its own `go.mod` (Day 20). Keep the history honest: a fresh repository with a first commit called "Import from 100 Days of Go" is fine.
3. **Tests and CI.** `go vet`, `go test -race ./...` and `gofmt -l .` in `.github/workflows/ci.yml` (Day 63). The badge goes at the top of the README. Never commit a secret: config comes from env (Day 21).
4. **Deploy it** to a free tier (Render, Fly.io, Koyeb, or a static host for a CLI's docs), with a health check. Put the live link at the top of the README.
5. **README**: one line on what it is and who it's for; the live link and a screenshot or GIF; how to run it locally in three commands; the design in a short diagram; the trade-offs and what you'd do next. Write it for a stranger who has two minutes.
6. **Demo**: a 60 to 90 second screen recording that shows it working, linked from the README. Show the result first, then how.

**Part B: the portfolio tool**

7. Make the tests in `starter/` pass: `Load`, `Validate`, `RepoPath`, `Sorted`, `BadgeTable`, `CVLines` and `RenderPage`. The doc comments give the exact formats.
8. Finish `starter/cmd/portfolio`: flags `-in` and `-out`, exit status 1 with a clear message on any error.
9. Copy `starter/projects.example.json` to your own `projects.json`, fill it in with your real projects, and generate the site. Paste `BADGES.md` into your GitHub profile README.
10. Run the tests:

```bash
go test ./phase-8-portfolio/day-100-your-portfolio/starter
```

**Stretch goals**

- A GitHub Action that regenerates the site on every push to `projects.json` and publishes it with GitHub Pages.
- A dark mode for the page with `prefers-color-scheme`, and a check with Day 72's accessibility tools.
- An `og:image` and `og:description` per project, so links look good when shared on LinkedIn and WhatsApp.
- Ship a second project, then a third. Three solid ones beat ten thin ones.

### Portfolio site checklist

- [ ] Your name, the role you want, and where you are (or "remote"), in the first screen.
- [ ] Two or three featured projects, each with a live link, the code, and one sentence on the problem it solves.
- [ ] Every link works. Click them all on a phone, on mobile data.
- [ ] Every repository it links to has a README, a green CI badge and no secrets in its history.
- [ ] It loads fast on a cheap phone: no 5 MB hero image, no framework you don't need (Day 74).
- [ ] Contact: an email address and LinkedIn. GitHub is linked from every project.
- [ ] It has your own domain if you can afford one (a `.co.ke` or `.dev`; check current prices with a registrar), or a tidy `you.github.io`.
- [ ] No "lorem ipsum", no unfinished sections, no typos. Ask a friend to read it.

### A CV or LinkedIn line for each project

Good lines start with a verb, name the hard part, and give a number when there is one. Rewrite these with your own numbers:

| Project | Line |
|---|---|
| Pesa Budget CLI (Day 26) | Built a Go CLI that categorises M-Pesa statements and reports monthly spending, configured by env and flags. |
| Price Watcher (Day 48) | Wrote a concurrent price scraper in Go with worker pools and per-site rate limits. |
| Matatu Booking API (Day 64) | Built a REST API in Go and Postgres for seat booking with payments, auth and integration tests, deployed with Docker. |
| Booking App (Day 76) | Built a React and TypeScript front end for a Go booking API with live seat updates over WebSockets. |
| Hardened Service (Day 84) | Wrote a threat model for a Go service and shipped the fixes: metrics, structured logs, secret rotation. |
| Link Shortener (Day 85) | Built and deployed a URL shortener in Go from a one-paragraph brief, with tests and CI. |
| Kiswahili Wordle (Day 86) | Built a terminal word game in Go with a Kiswahili word list and table-driven tests for every game rule. |
| Chat Server (Day 87) | Built a WebSocket chat server in Go with rooms and safe concurrent broadcasting. |
| Invoice PDF (Day 88) | Generated PDF invoices in Go from structured data, with tests for totals and VAT. |
| Static Site Generator (Day 89) | Wrote a Markdown-to-HTML static site generator in Go. |
| Rate-Limited Proxy (Day 90) | Built a reverse proxy in Go with per-client token-bucket rate limiting. |
| Matatu Route Finder (Day 91) | Implemented shortest-path routing over a Nairobi stage graph in Go. |
| Log Analyser (Day 92) | Wrote a streaming log analyser in Go that handles files larger than memory. |
| Key-Value Store (Day 93) | Built a crash-safe key-value store in Go with a checksummed write-ahead log, recovery and compaction. |
| Job Board (Day 94) | Built a job board in Go with ranked search, pagination, html/template pages and CSRF protection. |
| Webhook Relay (Day 95) | Built an HMAC-verified webhook relay in Go with replay protection, deduplication and exponential backoff. |
| Discord Bot (Day 96) | Built a Discord bot in Go on the interactions endpoint with Ed25519 request verification, no SDK. |
| Open-source fix (Day 97) | Fixed a money-parsing bug in an open-source Go package, with a regression test and a PR that followed the project's contributing guide. |
| Make It 10x Faster (Day 98) | Made a Go statement parser 400x faster using pprof and benchmarks, guarded by allocation tests in CI. |

## The kata: Polish Plan

Choose polishing tasks, each with hours and impact, to get the most impact within your hours. It's the 0/1 knapsack, the classic dynamic programming interview question, and it's also today's real decision. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 100
```

## When you're stuck

- **`json: unknown field "repo_url"`**: that's `DisallowUnknownFields` doing its job. Use the field names from the struct tags.
- **Validate stops at the first problem**: collect errors in a slice and return `errors.Join(errs...)`. It returns `nil` for an empty slice.
- **`RepoPath` accepts `https://github.com/you/repo/tree/main`**: split the trimmed path on `/` and require exactly two parts.
- **The badge shows "no status"**: the workflow file isn't named `ci.yml`, or it hasn't run on the default branch yet. Push once and check the Actions tab.
- **The page shows `&lt;b&gt;` where you wanted bold**: that's correct. Text from the JSON is never trusted as HTML, even your own.
- **A demo link renders as `#ZgotmplZ`**: `html/template` replaced an unsafe URL such as `javascript:...`. `Validate` should have rejected it first.
- **You keep polishing and never ship**: set a deadline, ship, then improve it in public. A deployed project with known gaps beats a perfect one on your laptop.
- **Still stuck?** Open `solution/portfolio.go` and `solution/cmd/portfolio/main.go`, read them once, close them, then write your own version from memory.

## Make it portfolio-ready

This day is the portfolio, so the last checklist is the whole course's:

- **README**: your GitHub profile README holds the headline, the badge table from `BADGES.md`, and a link to the site.
- **Screenshots or demo**: the shipped project's recording, linked from the site and the CV.
- **Deploy link**: the site itself (GitHub Pages or any static host), and the shipped project's live URL.
- **Tests and CI**: this tool gets its own CI badge too. Green badges everywhere are the quickest signal that you finish things.
- **What you'd do next**: a short "Now" section on the site: what you're learning, the next project you'll ship, the open-source issue you're working on. Update it every month.

You started with a single M-Pesa SMS on Day 1. A hundred days later you can design, build, test, secure, speed up and ship a service, and explain how you did it. Ship it, and go and get the job.
