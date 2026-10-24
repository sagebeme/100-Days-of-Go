# Syllabus

One hundred days, eight phases. Each phase ends with a capstone that uses everything before it. The Status column says which days are written (**Done**) and which are still **Planned**.

Every day has two parts:

- **The project**: a small program with a story, a `starter/` folder, and tests that tell you when it works.
- **The kata**: a 5–15 minute puzzle in HackerRank format. Read from stdin, print to stdout, beat the time limit. `go run ./tools/kata <day>` judges it.

## Phase 1: Terminal Go

Learn: values, types, if and switch, loops, functions, slices, maps, strings and runes, packages, go test, reading input
Ends with: a text adventure in sheng

| Day | Project | What it teaches | Status |
|---|---|---|---|
| 1 | M-Pesa SMS Builder | variables, constants, fmt.Sprintf | Done |
| 2 | Bundle Calculator | integers, floats and converting between them | Done |
| 3 | Matatu Fare Rules | if, else and switch | Done |
| 4 | Times Tables Drill | for loops and accumulators | Done |
| 5 | Fuel Split | functions, multiple return values, errors as values | Done |
| 6 | Chama Ledger | slices, append and range | Done |
| 7 | Sheng Dictionary | maps, the comma-ok idiom | Done |
| 8 | Hashtag Maker | strings, runes and the strings package | Done |
| 9 | Phone Number Cleaner | validating input, byte by byte | Done |
| 10 | Exam Grader | sorting with sort.Slice, simple stats | Done |
| 11 | Receipt Printer | formatting columns with fmt and text/tabwriter | Done |
| 12 | Number Guessing Game | reading stdin with bufio, math/rand | Done |
| 13 | Package It Up | packages, exported names, go doc, go vet | Done |
| 14 | Capstone: Lost in Nairobi CBD | a text adventure driven by a map of rooms | Done |

## Phase 2: Modelling and engineering habits

Learn: structs, methods, interfaces, errors, generics, table tests, modules, env variables and config, flags, files and JSON, git
Ends with: a CLI M-Pesa budget tracker

| Day | Project | What it teaches | Status |
|---|---|---|---|
| 15 | Boda Boda Fleet | structs, methods, pointer vs value receivers | Done |
| 16 | Payment Methods | interfaces and implicit satisfaction | Done |
| 17 | Failed Transactions | custom errors, wrapping, errors.Is and errors.As | Done |
| 18 | Typed Inventory | generics: type parameters and constraints | Done |
| 19 | Table-Driven Tests | t.Run, subtests, test helpers, coverage | Done |
| 20 | Go Modules | go.mod, semantic versions, adding and upgrading a dependency | Done |
| 21 | Config Loader | env variables, 12-factor config, .env files, never committing secrets | Done |
| 22 | CLI Flags | the flag package, subcommands, exit codes | Done |
| 23 | Save and Load | files, encoding/json, struct tags | Planned |
| 24 | CSV Statements | encoding/csv, parsing a real M-Pesa statement | Planned |
| 25 | Git Like a Team | branches, small commits, pull requests, code review | Planned |
| 26 | Capstone: Pesa Budget CLI | tracks spending from statements, configured by env and flags | Planned |

## Phase 3: Algorithms for HackerRank

Learn: Big-O, arrays and strings, two pointers, sliding window, sorting, binary search, stacks, queues, hashing, recursion, trees, graphs, greedy, dynamic programming
Ends with: a timed mock HackerRank test

| Day | Project | What it teaches | Status |
|---|---|---|---|
| 27 | How Fast Is Fast | Big-O, counting operations, fast stdin and stdout | Done |
| 28 | Duplicate Receipts | hashing, sets, frequency counts | Done |
| 29 | Two Pointers | pair sums, reversing, merging sorted lists | Done |
| 30 | Sliding Window | best streaks and moving averages | Done |
| 31 | Sorting Showdown | merge sort, quicksort, custom comparators | Done |
| 32 | Binary Search | searching sorted data and searching the answer | Done |
| 33 | Stacks | bracket checkers and undo history | Done |
| 34 | Queues | a bank queue simulator and deques | Done |
| 35 | Recursion | factorials to permutations, base cases | Done |
| 36 | Trees | binary search trees, traversals | Done |
| 37 | Graphs | BFS and DFS on a matatu route map | Done |
| 38 | Greedy | change-making and scheduling | Done |
| 39 | Dynamic Programming | memoisation, tabulation, classic problems | Planned |
| 40 | Capstone: Mock HackerRank Test | six problems, 90 minutes, judged locally | Planned |

## Phase 4: Concurrency

Learn: goroutines, channels, select, context, sync, worker pools, the race detector
Ends with: a concurrent price scraper

| Day | Project | What it teaches | Status |
|---|---|---|---|
| 41 | Goroutines | starting work concurrently, sync.WaitGroup | Done |
| 42 | Channels | sending, receiving, closing, ranging | Done |
| 43 | Select | timeouts, multiplexing, default cases | Done |
| 44 | Context | cancellation and deadlines | Done |
| 45 | Mutexes | shared state, sync.Mutex, the race detector | Done |
| 46 | Worker Pools | bounded concurrency, fan-out and fan-in | Done |
| 47 | Rate Limiting | tickers and token buckets | Done |
| 48 | Capstone: Price Watcher | scrapes many shops at once without hammering any | Done |

## Phase 5: Backend engineering

Learn: net/http, HTTP methods and status codes, routing, JSON APIs, validation, Postgres, migrations, auth, OWASP Top 10, CORS, security headers, logging, Docker, CI, deploying
Ends with: a deployed matatu-booking API

| Day | Project | What it teaches | Status |
|---|---|---|---|
| 49 | Hello, Server | net/http, handlers, ServeMux, ports | Done |
| 50 | HTTP Methods | GET, POST, PUT, PATCH, DELETE, safety and idempotency | Done |
| 51 | Status Codes | 2xx to 5xx, choosing the right one, error bodies | Done |
| 52 | JSON API | encoding requests and responses, content types | Done |
| 53 | Validation | rejecting bad input early, limiting body size | Done |
| 54 | Middleware | logging, recovery, request IDs, slog | Done |
| 55 | Postgres | database/sql, pgx, connection pools | Planned |
| 56 | Migrations | versioned schema changes, rolling back | Planned |
| 57 | Passwords | hashing with bcrypt and argon2, never storing plain text | Done |
| 58 | Sessions and JWTs | cookies vs tokens, expiry, refresh rotation | Done |
| 59 | Roles and Permissions | 401 vs 403, authorisation checks in one place | Done |
| 60 | OWASP Top 10 | injection, broken access control, SSRF, and how Go code avoids them | Done |
| 61 | CORS and Security Headers | origins, preflight, CSP, HSTS | Done |
| 62 | Integration Tests | httptest, a real test database, testing the whole request | Planned |
| 63 | Docker and CI | a small image, GitHub Actions, secrets in CI | Planned |
| 64 | Capstone: Matatu Booking API | seats, bookings and payments, deployed | Planned |

## Phase 6: Front end with TypeScript and React

Learn: TypeScript, React, Vite, fetching from your Go API, forms, auth flows, WebSockets, accessibility, front-end tests
Ends with: a full-stack booking app

| Day | Project | What it teaches | Status |
|---|---|---|---|
| 65 | TypeScript for Go Developers | types, interfaces, unions, async | Planned |
| 66 | First React App | Vite, components, props | Planned |
| 67 | State and Events | useState, lists, keys | Planned |
| 68 | Talking to Go | fetch, loading and error states, CORS in practice | Planned |
| 69 | Forms | controlled inputs, validation on both sides | Planned |
| 70 | Login Flow | storing tokens safely, protected routes | Planned |
| 71 | Live Seats | WebSockets from Go to React | Planned |
| 72 | Accessible by Default | semantics, keyboard use, contrast, axe | Planned |
| 73 | Front-End Tests | Vitest, Testing Library, Playwright | Planned |
| 74 | Fast on a Cheap Phone | bundle size, caching, slow 3G | Planned |
| 75 | Deploying the Front End | static hosting, env variables at build time | Planned |
| 76 | Capstone: Booking App | the Phase 5 API with a React front end, end to end | Planned |

## Phase 7: Production engineering

Learn: system design, caching, queues, gRPC, observability, threat modelling, code review, secrets
Ends with: a design doc and a hardened service

| Day | Project | What it teaches | Status |
|---|---|---|---|
| 77 | System Design Basics | load, latency, bottlenecks, drawing the system | Planned |
| 78 | Caching | Redis, cache-aside, invalidation | Planned |
| 79 | Background Jobs | queues, retries, idempotent workers | Planned |
| 80 | gRPC | protobuf, service contracts, streaming | Planned |
| 81 | Observability | metrics, traces, structured logs, health checks | Planned |
| 82 | Threat Modelling | STRIDE on your own API | Planned |
| 83 | Secrets Management | rotating keys, vaults, least privilege | Planned |
| 84 | Capstone: Hardened Service | a design doc, then the changes it calls for | Planned |

## Phase 8: Portfolio

Learn: no walkthrough: a brief, tests, and you
Ends with: sixteen projects of your own

| Day | Project | What it teaches | Status |
|---|---|---|---|
| 85 | Link Shortener | a full service from a one-paragraph brief | Done |
| 86 | Kiswahili Wordle | a terminal game with a word list | Planned |
| 87 | Chat Server | WebSockets and rooms | Planned |
| 88 | Invoice PDF | generating documents | Planned |
| 89 | Static Site Generator | Markdown to HTML | Planned |
| 90 | Rate-Limited Proxy | reverse proxies and limits | Planned |
| 91 | Matatu Route Finder | shortest paths on a real map | Planned |
| 92 | Log Analyser | streaming big files | Planned |
| 93 | Key-Value Store | a tiny database with a write-ahead log | Planned |
| 94 | Job Board | full stack with search | Planned |
| 95 | Webhook Relay | signatures, retries, replay protection | Planned |
| 96 | Discord Bot | a bot that runs your katas | Planned |
| 97 | Fix an Open-Source Issue | reading someone else's code | Planned |
| 98 | Make It 10x Faster | profiling with pprof | Planned |
| 99 | Mock Interview | a timed HackerRank test and a system design question | Planned |
| 100 | Capstone: Your Portfolio | ship one project people can use, with a write-up | Planned |
