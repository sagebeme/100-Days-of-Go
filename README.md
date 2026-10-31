# 100 Days of Go

One project a day, for 100 days. You start by building an M-Pesa confirmation text in your terminal and finish as a software engineer who can pass a HackerRank test, build and secure a Go API, put a React front end on it, and ship both.

Every day ships with tests. You run them and they tell you when you got it right, so you never have to compare your code with someone else's.

Every day also ends with a **kata**: a 5–15 minute puzzle in the exact format HackerRank uses. Do one a day and timed coding tests stop being scary.

## What you'll be able to do

| By the end of | You can |
|---|---|
| Phase 1 (Day 14) | write and test small Go programs, read input, use slices, maps and strings |
| Phase 2 (Day 26) | model data with structs and interfaces, handle errors properly, configure apps with **env variables**, work in git like a team |
| Phase 3 (Day 40) | pass a timed HackerRank test: Big-O, two pointers, sorting, binary search, trees, graphs, dynamic programming |
| Phase 4 (Day 48) | run work concurrently with goroutines, channels and context, without data races |
| Phase 5 (Day 64) | build a JSON API with the right **HTTP methods and status codes**, Postgres, auth, the **OWASP Top 10** defences, Docker and CI |
| Phase 6 (Day 76) | build an accessible React and TypeScript front end on your own API |
| Phase 7 (Day 84) | design systems, add caching and queues, observe and threat-model a service |
| Phase 8 (Day 100) | build sixteen projects from a brief, with no walkthrough |

The full day-by-day plan is in [SYLLABUS.md](SYLLABUS.md).

## What you need

| | |
|---|---|
| **Go** | 1.26 or newer ([go.dev/dl](https://go.dev/dl/)) |
| **An editor** | VS Code with the Go extension, GoLand, or whatever you already use |
| **Git** | to clone this repo and save your work |
| **Node.js** | 24 or newer, from Phase 6 only |
| **Time** | about an hour a day |

Check your Go version:

```bash
go version
```

## Getting started

```bash
git clone git@github.com:sagebeme/100-Days-of-Go.git
cd 100-Days-of-Go
```

There's nothing to install: Phase 1 uses only Go's standard library. Start Day 1:

```bash
cd phase-1-terminal/day-001-mpesa-sms
```

Read its `README.md`, write your code in `starter/`, and run the tests.

## How each day works

```
phase-1-terminal/day-001-mpesa-sms/
  README.md        the brief, what you'll use, steps, and help when you're stuck
  starter/         your code goes here, next to the tests that check it
  solution/        one working answer: read it only after you've tried
  kata/
    README.md      the day's HackerRank-style puzzle
    main.go        your answer
    solution/      one working answer
    testdata/      1.in, 1.out, 2.in, ...: the judge's test cases
    gen/           on many katas: makes one full-size input when you judge
```

Run a day's tests from the repo root:

```bash
go test ./phase-1-terminal/day-001-mpesa-sms/starter
```

Judge a day's kata:

```bash
go run ./tools/kata 1
```

The judge feeds each `.in` file to your program, compares what it prints with the `.out` file, and fails anything slower than 2 seconds, like HackerRank. `go run ./tools/kata -solution 1` runs the reference answer instead.

The files in `testdata/` are small, so you can read them. A full-size input can be megabytes, so many katas make one instead: `kata/gen` prints it, the reference solution answers it, and your program must match that answer within the same 2 seconds. It shows up as case `big`. That's the case that catches an answer that's right but too slow, just as HackerRank's hidden tests do. If `big` is the only case that fails, and it fails by timing out, your logic is probably right but too slow. Make it faster, and read the Limits paragraph in the kata's README for what to avoid.

Don't run `go test ./...` across the whole repo: every untouched starter fails on purpose, so you'll see a wall of red for days you haven't reached.

## Env variables and secrets

From Phase 2 the projects read their settings from env variables. `.env.example` lists them with safe placeholder values. Copy it to `.env` and fill in your own:

```bash
cp .env.example .env
```

`.env` is in `.gitignore`. Never commit it: anything pushed to GitHub is public, and bots scan new commits for keys within minutes. Day 21 covers this properly.

## Reference

- [HTTP methods, headers and status codes](reference/http.md)
- [Env variables and config](reference/env-and-config.md)
- [OWASP Top 10 checklist for Go](reference/owasp-checklist.md)
- [Big-O cheat sheet](reference/big-o.md)

## The course website

The `site/` folder is a website for the course: a landing page, the syllabus, and a lesson viewer that shows each day's README with a checklist that saves your progress in your browser. It's plain static files, so any static server works:

```bash
python3 -m http.server 8000
```

Then open `http://localhost:8000/site/`. After adding or renaming a day, refresh the site's list with `go run ./tools/sitedata`; `go run ./tools/sitedata -check` fails if it's out of date.

To put it online, import the repo into Netlify or Vercel. `netlify.toml` and `vercel.json` are already set up, and `/` goes to `/site/`.

## For contributors

```bash
gofmt -l .                        # prints nothing when everything is formatted
go vet ./...
go run ./tools/verify -starters   # every solution and kata passes; every starter still fails
go run ./tools/sitedata -check
```

CI runs the same checks on every push and pull request.
