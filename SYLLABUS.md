# Syllabus

One hundred days, eight phases. Each phase ends with a capstone that uses everything before it. Phase 1 (Days 1–14) is written. The other phases are planned and arrive one phase at a time.

Every day has two parts:

- **The project**: a small program with a story, a `starter/` folder, and tests that tell you when it works.
- **The kata**: a 5–15 minute puzzle in HackerRank format. Read from stdin, print to stdout, beat the time limit. `go run ./tools/kata <day>` judges it.

## Phase 1: Terminal Go

Learn: values, types, if and switch, loops, functions, slices, maps, strings and runes, packages, go test, reading input
Ends with: a text adventure in sheng

| Day | Project | What it teaches |
|---|---|---|
| 1 | M-Pesa SMS Builder | variables, constants, fmt.Sprintf |
| 2 | Bundle Calculator | integers, floats and converting between them |
| 3 | Matatu Fare Rules | if, else and switch |
| 4 | Times Tables Drill | for loops and accumulators |
| 5 | Fuel Split | functions, multiple return values, errors as values |
| 6 | Chama Ledger | slices, append and range |
| 7 | Sheng Dictionary | maps, the comma-ok idiom |
| 8 | Hashtag Maker | strings, runes and the strings package |
| 9 | Phone Number Cleaner | validating input, byte by byte |
| 10 | Exam Grader | sorting with sort.Slice, simple stats |
| 11 | Receipt Printer | formatting columns with fmt and text/tabwriter |
| 12 | Number Guessing Game | reading stdin with bufio, math/rand |
| 13 | Package It Up | packages, exported names, go doc, go vet |
| 14 | Capstone: Lost in Nairobi CBD | a text adventure driven by a map of rooms |

## Phase 2: Modelling and engineering habits

Learn: structs, methods, interfaces, errors, generics, table tests, modules, env variables and config, flags, files and JSON, git
Ends with: a CLI M-Pesa budget tracker

| Day | Project | What it teaches |
|---|---|---|
| 15 | Boda Boda Fleet | structs, methods, pointer vs value receivers |
| 16 | Payment Methods | interfaces and implicit satisfaction |
| 17 | Failed Transactions | custom errors, wrapping, errors.Is and errors.As |
| 18 | Typed Inventory | generics: type parameters and constraints |
| 19 | Table-Driven Tests | t.Run, subtests, test helpers, coverage |
| 20 | Go Modules | go.mod, semantic versions, adding and upgrading a dependency |
| 21 | Config Loader | env variables, 12-factor config, .env files, never committing secrets |
| 22 | CLI Flags | the flag package, subcommands, exit codes |
| 23 | Save and Load | files, encoding/json, struct tags |
| 24 | CSV Statements | encoding/csv, parsing a real M-Pesa statement |
| 25 | Git Like a Team | branches, small commits, pull requests, code review |
| 26 | Capstone: Pesa Budget CLI | tracks spending from statements, configured by env and flags |

## Phase 3: Algorithms for HackerRank

Learn: Big-O, arrays and strings, two pointers, sliding window, sorting, binary search, stacks, queues, hashing, recursion, trees, graphs, greedy, dynamic programming
Ends with: a timed mock HackerRank test

| Day | Project | What it teaches |
|---|---|---|
| 27 | How Fast Is Fast | Big-O, counting operations, fast stdin and stdout |
| 28 | Duplicate Receipts | hashing, sets, frequency counts |
| 29 | Two Pointers | pair sums, reversing, merging sorted lists |
| 30 | Sliding Window | best streaks and moving averages |
| 31 | Sorting Showdown | merge sort, quicksort, custom comparators |
| 32 | Binary Search | searching sorted data and searching the answer |
| 33 | Stacks | bracket checkers and undo history |
| 34 | Queues | a bank queue simulator and deques |
| 35 | Recursion | factorials to permutations, base cases |
| 36 | Trees | binary search trees, traversals |
| 37 | Graphs | BFS and DFS on a matatu route map |
| 38 | Greedy | change-making and scheduling |
| 39 | Dynamic Programming | memoisation, tabulation, classic problems |
| 40 | Capstone: Mock HackerRank Test | six problems, 90 minutes, judged locally |

## Phase 4: Concurrency

Learn: goroutines, channels, select, context, sync, worker pools, the race detector
Ends with: a concurrent price scraper

| Day | Project | What it teaches |
|---|---|---|
| 41 | Goroutines | starting work concurrently, sync.WaitGroup |
| 42 | Channels | sending, receiving, closing, ranging |
| 43 | Select | timeouts, multiplexing, default cases |
| 44 | Context | cancellation and deadlines |
| 45 | Mutexes | shared state, sync.Mutex, the race detector |
| 46 | Worker Pools | bounded concurrency, fan-out and fan-in |
| 47 | Rate Limiting | tickers and token buckets |
| 48 | Capstone: Price Watcher | scrapes many shops at once without hammering any |

## Phase 5: Backend engineering

Learn: net/http, HTTP methods and status codes, routing, JSON APIs, validation, Postgres, migrations, auth, OWASP Top 10, CORS, security headers, logging, Docker, CI, deploying
Ends with: a deployed matatu-booking API

| Day | Project | What it teaches |
|---|---|---|
| 49 | Hello, Server | net/http, handlers, ServeMux, ports |
| 50 | HTTP Methods | GET, POST, PUT, PATCH, DELETE, safety and idempotency |
| 51 | Status Codes | 2xx to 5xx, choosing the right one, error bodies |
| 52 | JSON API | encoding requests and responses, content types |
| 53 | Validation | rejecting bad input early, limiting body size |
| 54 | Middleware | logging, recovery, request IDs, slog |
| 55 | Postgres | database/sql, pgx, connection pools |
| 56 | Migrations | versioned schema changes, rolling back |
| 57 | Passwords | hashing with bcrypt and argon2, never storing plain text |
| 58 | Sessions and JWTs | cookies vs tokens, expiry, refresh rotation |
| 59 | Roles and Permissions | 401 vs 403, authorisation checks in one place |
| 60 | OWASP Top 10 | injection, broken access control, SSRF, and how Go code avoids them |
| 61 | CORS and Security Headers | origins, preflight, CSP, HSTS |
| 62 | Integration Tests | httptest, a real test database, testing the whole request |
| 63 | Docker and CI | a small image, GitHub Actions, secrets in CI |
| 64 | Capstone: Matatu Booking API | seats, bookings and payments, deployed |

## Phase 6: Front end with TypeScript and React

Learn: TypeScript, React, Vite, fetching from your Go API, forms, auth flows, WebSockets, accessibility, front-end tests
Ends with: a full-stack booking app

| Day | Project | What it teaches |
|---|---|---|
| 65 | TypeScript for Go Developers | types, interfaces, unions, async |
| 66 | First React App | Vite, components, props |
| 67 | State and Events | useState, lists, keys |
| 68 | Talking to Go | fetch, loading and error states, CORS in practice |
| 69 | Forms | controlled inputs, validation on both sides |
| 70 | Login Flow | storing tokens safely, protected routes |
| 71 | Live Seats | WebSockets from Go to React |
| 72 | Accessible by Default | semantics, keyboard use, contrast, axe |
| 73 | Front-End Tests | Vitest, Testing Library, Playwright |
| 74 | Fast on a Cheap Phone | bundle size, caching, slow 3G |
| 75 | Deploying the Front End | static hosting, env variables at build time |
| 76 | Capstone: Booking App | the Phase 5 API with a React front end, end to end |

## Phase 7: Production engineering

Learn: system design, caching, queues, gRPC, observability, threat modelling, code review, secrets
Ends with: a design doc and a hardened service

| Day | Project | What it teaches |
|---|---|---|
| 77 | System Design Basics | load, latency, bottlenecks, drawing the system |
| 78 | Caching | Redis, cache-aside, invalidation |
| 79 | Background Jobs | queues, retries, idempotent workers |
| 80 | gRPC | protobuf, service contracts, streaming |
| 81 | Observability | metrics, traces, structured logs, health checks |
| 82 | Threat Modelling | STRIDE on your own API |
| 83 | Secrets Management | rotating keys, vaults, least privilege |
| 84 | Capstone: Hardened Service | a design doc, then the changes it calls for |

## Phase 8: Portfolio

Learn: no walkthrough: a brief, tests, and you
Ends with: sixteen projects of your own

| Day | Project | What it teaches |
|---|---|---|
| 85 | Link Shortener | a full service from a one-paragraph brief |
| 86 | Kiswahili Wordle | a terminal game with a word list |
| 87 | Chat Server | WebSockets and rooms |
| 88 | Invoice PDF | generating documents |
| 89 | Static Site Generator | Markdown to HTML |
| 90 | Rate-Limited Proxy | reverse proxies and limits |
| 91 | Matatu Route Finder | shortest paths on a real map |
| 92 | Log Analyser | streaming big files |
| 93 | Key-Value Store | a tiny database with a write-ahead log |
| 94 | Job Board | full stack with search |
| 95 | Webhook Relay | signatures, retries, replay protection |
| 96 | Discord Bot | a bot that runs your katas |
| 97 | Fix an Open-Source Issue | reading someone else's code |
| 98 | Make It 10x Faster | profiling with pprof |
| 99 | Mock Interview | a timed HackerRank test and a system design question |
| 100 | Capstone: Your Portfolio | ship one project people can use, with a write-up |
