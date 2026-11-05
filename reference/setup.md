# Setting up your machine

Everything the course needs, and when you need it:

| Tool | Version | Needed from |
|---|---|---|
| Go | 1.26 or newer (`go.mod` says `go 1.26.0`) | Day 1 |
| Git | any recent one | Day 1 |
| Node.js and npm | 24 or newer (CI uses 24) | Phase 6, Day 65 |
| Docker, for Postgres 16 and Redis 7 | any recent one | Day 55 (optional before that) |

Phases 1 to 4 need only Go. Don't install the rest until you get there.

## Go

Any Go from 1.26.0 up works. If your Go is 1.21 or newer but older than 1.26, the `go` command downloads Go 1.26 for this repo by itself the first time you run it here (that's `GOTOOLCHAIN=auto`, the default). Installing 1.26 is still simpler.

### Linux

Download the archive from [go.dev/dl](https://go.dev/dl/) and unpack it into `/usr/local`, replacing any old copy:

```bash
curl -LO https://go.dev/dl/go1.26.0.linux-amd64.tar.gz   # or the newest 1.26.x; linux-arm64 on ARM
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc   # ~/.zshrc if you use zsh
source ~/.bashrc
go version
```

Your distro's package (`apt install golang-go` and so on) is often several versions behind. Check `go version` if you use it.

### macOS

Either the `.pkg` installer from [go.dev/dl](https://go.dev/dl/), or Homebrew:

```bash
brew install go
go version
```

### Windows

Run the `.msi` installer from [go.dev/dl](https://go.dev/dl/), or:

```powershell
winget install GoLang.Go
```

Close and reopen your terminal so it picks up the new `PATH`, then run `go version`. The course's commands are written for bash. They work in PowerShell except where they set a variable in front of a command (`DATABASE_URL=... go test`): in PowerShell, set it first with `$env:DATABASE_URL = "..."`. WSL avoids that.

### WSL (Windows Subsystem for Linux)

Inside WSL you're on Linux: follow the Linux steps in the WSL terminal. Two things to know:

- **Clone the repo into your Linux home** (`~/100-Days-of-Go`), not under `/mnt/c/...`. Files on the Windows drive are many times slower for `go build`, `go test` and `npm`.
- **Docker:** install Docker Desktop on Windows and turn on *Settings → Resources → WSL integration* for your distro, or install Docker Engine inside WSL. Either way, `docker run` in WSL works and the ports reach `localhost`.

## Node.js (Phase 6)

Phase 6 (Days 65 to 76) is TypeScript and React, tested with Vitest. Its tools are pinned in [`phase-6-frontend/package-lock.json`](../phase-6-frontend/package-lock.json): TypeScript, Vite, Vitest, React Testing Library, and Playwright for Day 73's optional browser test. You only install Node; npm installs the rest.

A version manager makes it easy to have Node 24 and switch later:

```bash
# Linux, macOS, WSL: nvm (https://github.com/nvm-sh/nvm)
nvm install 24
node --version

# Windows: fnm or nvm-windows, or the installer from https://nodejs.org
winget install Schniz.fnm
fnm install 24
```

Then, once:

```bash
cd phase-6-frontend && npm ci
```

[phase-6-frontend/README.md](../phase-6-frontend/README.md) has the commands for running a day's tests, the dev server and the builds.

## Postgres and Redis (Phases 5 and 7)

Days 55 and 56 have integration tests that run against a real Postgres, Day 62's tests run on a real Postgres as well as in memory, and Day 78 has one that runs against a real Redis. Without the env variable that points at the database, those tests are **skipped** (Day 62's run in memory only), and everything else still runs. The other days use in-memory fakes, so you can do the whole course without a database and only miss those few tests.

The easiest way to get both is Docker, with the same versions CI uses.

### Postgres 16

This is the database every Phase 5 README and `.env.example` uses: user `gopher`, password `gopher`, database `matatu`. It's the same command as [Day 55](../phase-5-backend/day-055-postgres/README.md#running-postgres):

```bash
docker run --name matatu-db -e POSTGRES_USER=gopher -e POSTGRES_PASSWORD=gopher -e POSTGRES_DB=matatu -p 5432:5432 -d postgres:16-alpine
docker exec matatu-db pg_isready -U gopher          # wait until it says "accepting connections"
export DATABASE_URL='postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable'
```

`docker stop matatu-db` stops it and `docker start matatu-db` brings it back with your data. `docker rm -f matatu-db` deletes it. Day 55 also loads its schema into it; [Day 56](../phase-5-backend/day-056-migrations/README.md) and [Day 62](../phase-5-backend/day-062-integration-tests/README.md) explain how their tests use it (Day 62 makes a fresh schema per test, so it never touches your data).

CI starts a different but equivalent one: `postgres:16-alpine` with only `POSTGRES_PASSWORD=postgres`, and `DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`. The tests work with either. The `gopher` one matches what the lessons say.

### Redis 7

```bash
docker run --name twende-redis -p 6379:6379 -d redis:7-alpine
docker exec twende-redis redis-cli ping               # PONG
export REDIS_ADDR=localhost:6379
go test ./phase-7-production/day-078-caching/starter -run Real
```

[Day 78](../phase-7-production/day-078-caching/README.md) says `redis:7`; `redis:7-alpine` (what CI uses) is the same Redis in a smaller image. Its other tests use miniredis, an in-memory Redis, and need nothing running.

### The env variables the days read

| Variable | Read by | Example |
|---|---|---|
| `DATABASE_URL` | Days 55 and 56 (integration tests skip without it), 62 (in memory without it), 63's config (required in production) | `postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable` |
| `REDIS_ADDR` | Day 78's `Real` test (skipped without it) | `localhost:6379` |
| `PORT` | the Phase 5 APIs (default `8080`) | `8080` |
| `APP_ENV` | Days 49 and 63 | `development` |
| `JWT_SECRET` | Days 58, 63 and 64 (required; never commit a real one) | `openssl rand -base64 32` |
| `JWT_ISSUER`, `ACCESS_TOKEN_TTL`, `BCRYPT_COST`, `ADMIN_PHONES`, `CORS_ALLOWED_ORIGINS`, `RATE_LIMIT_RPS`, `RATE_LIMIT_BURST`, `CLIENT_IP_HEADER`, `SHUTDOWN_TIMEOUT` | Day 64's config; the README explains each | |
| `READY_TIMEOUT` | Day 63 | `2s` |
| `ADDR`, `ALLOWED_ORIGIN` | the Phase 6 `api/` servers (defaults `:8080`, `http://localhost:5173`) | |
| `TWENDE_*` (`TWENDE_ADDR`, `TWENDE_API_KEY`, `TWENDE_ENV`, ...) | Day 84's config, and Day 83's secrets (`TWENDE_DB_PASSWORD`) | |

Copy `.env.example` to `.env` for your own values. [reference/env-and-config.md](env-and-config.md) explains how the course reads them and keeps secrets out of git.

## Common errors

**`go: command not found`** (or `'go' is not recognized` on Windows). Go isn't on your `PATH`. On Linux and WSL, add `/usr/local/go/bin` as in the Linux steps and open a new terminal. On Windows and macOS, close and reopen the terminal after installing. `which go` (or `where go` on Windows) shows which one you're running.

**`go: go.mod requires go >= 1.26.0 (running go 1.2x; GOTOOLCHAIN=local)`**. Your Go is older than the course's and automatic downloads are off. Install Go 1.26, or run `go env -u GOTOOLCHAIN` so it can fetch it.

**`go: cannot find main module`, `directory . outside main module`, `package ... is not in std`**. You're outside the repo, or the path has no `./`. Run commands from the repo root (the folder with `go.mod`) and start paths with `./`: `go test ./phase-1-terminal/day-001-mpesa-sms/starter`.

**GOPATH confusion.** You don't need to set `GOPATH`, and the repo doesn't have to live in `~/go/src`: this is a Go module, so clone it anywhere. If an old tutorial had you set `GO111MODULE=off`, undo it with `go env -u GO111MODULE` (and remove it from your shell profile). `~/go` is still where Go keeps downloaded modules and `go install`ed programs.

**Everything fails when I run `go test ./...`**. That's expected: every starter you haven't done yet fails on purpose. Test one day at a time, as the [README](../README.md#how-each-day-works) shows.

**`connect: connection refused` on `localhost:5432` (or `6379`)**. Nothing is listening there. `docker ps` shows whether the container is running; `docker start matatu-db` restarts a stopped one. Postgres takes a few seconds after `docker run` before it accepts connections: wait for `pg_isready`. In WSL with Docker Desktop, check WSL integration is on.

**`password authentication failed for user "gopher"`** or **`database "matatu" does not exist`**. You're reaching a different Postgres: one installed on your machine already listening on 5432, or an old container made with other settings (the user and password are only set the first time a container starts). Check with `docker ps`, and either use the next fix or `docker rm -f matatu-db` and run it again.

**`SSL is not enabled on the server`**. Add `?sslmode=disable` to `DATABASE_URL`. The local container has no TLS.

**`port is already allocated` / `address already in use`**. Something else already uses the port: often a Postgres or Redis installed on your machine, or yesterday's API still running. Find it with `ss -ltnp | grep 5432` (Linux), `lsof -i :5432` (macOS) or `netstat -ano | findstr :5432` (Windows). Stop it, or map the container to another port and point the URL there:

```bash
docker run --name matatu-db -e POSTGRES_USER=gopher -e POSTGRES_PASSWORD=gopher -e POSTGRES_DB=matatu -p 55432:5432 -d postgres:16-alpine
export DATABASE_URL='postgres://gopher:gopher@localhost:55432/matatu?sslmode=disable'
```

The same goes for the APIs: `PORT=8081 go run ...` (Phase 5) or `ADDR=:8081` (Phase 6's `api/` servers), and Vite picks the next free port by itself if 5173 is taken.

**`Cannot find module 'vitest'`, `vitest: not found`, `sh: tsc: not found`**. Run `npm ci` in `phase-6-frontend` first, and run `npm test` from that folder, not the repo root.

**`npm ci` or `npm test` fails with an `EBADENGINE` warning or a `SyntaxError` inside `node_modules`**. Your Node is probably too old: `node --version` should be 24 or newer.
