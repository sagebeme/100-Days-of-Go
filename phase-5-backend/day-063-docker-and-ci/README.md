# Day 63: Docker and CI

## The brief

The Twende (made-up) API works on your laptop. Tomorrow it goes on a server you'll never log in to, built and shipped by a robot on every push. Today you get it ready for that: a container image that's small and doesn't run as root, config that's checked before the server opens a port, the two endpoints a hosting platform uses to look after it, and a GitHub Actions workflow that tests, scans, builds and deploys.

```
GET /healthz   →  200 {"status":"ok"}                                   (the process is alive)
GET /readyz    →  200 {"status":"ready","checks":{"db":"ok"}}           (send it traffic)
GET /readyz    →  503 {"status":"not ready","checks":{"db":"failed"}}   (database down: no traffic, no restart)
GET /readyz    →  503 {"status":"draining"}                             (shutting down: finish what you have)

APP_ENV=production ./api
  config:
  DATABASE_URL is required in production
  JWT_SECRET is required in production
  (exit status 1)

docker build -f phase-5-backend/day-063-docker-and-ci/solution/Dockerfile -t twende-api .
docker images twende-api   →  about 13 MB
```

## What you'll use

- **A multi-stage Dockerfile.** The first stage has the whole Go toolchain (around a gigabyte) and builds the binary. The second stage copies in only the binary. Go makes this easy: with `CGO_ENABLED=0` the binary is static and needs no C libraries.
- **Distroless.** `gcr.io/distroless/static-debian12:nonroot` has CA certificates, time zone data and a user called `nonroot` (UID 65532), and nothing else: no shell, no package manager, nothing for an attacker to use. `scratch` is even emptier, but then you must copy in certificates yourself.
- **Not root.** `USER nonroot:nonroot`. If someone breaks into the process, they aren't root in the container.
- **Layer caching.** `COPY go.mod go.sum` and `RUN go mod download` before `COPY . .`, so changing code doesn't download every module again.
- **Liveness and readiness.** `/healthz` says "the process can answer". If it fails, the platform restarts the container. It checks nothing else, because restarting won't fix a database outage. `/readyz` says "I can do real work": every dependency check passes. If it fails, the platform stops sending traffic but doesn't restart. During shutdown it answers 503 so the load balancer moves traffic away first.
- **A `Checker` interface** with `Name()` and `Check(ctx)`. `NewCheck("db", db.PingContext)` adapts any function. The checks run at the same time, one goroutine each, under one `context.WithTimeout`. A check must return when its context ends.
- **`sync/atomic`**: `atomic.Bool` is a boolean that's safe to read and write from many goroutines without a mutex. New today. It fits a single flag like "draining".
- **Graceful shutdown.** `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` gives a context that ends on Ctrl-C or on the `SIGTERM` a platform sends before it stops a container. Then `srv.Shutdown(ctx)` stops accepting new connections and waits for running requests, up to a timeout.
- **Config at startup**, as in [env-and-config.md](../../reference/env-and-config.md): read everything once, collect every problem with `errors.Join`, exit before opening a port. Production requires secrets that development can run without, but nothing secret ever has a default.
- **GitHub Actions.** A workflow is a YAML file in `.github/workflows/`. Jobs run on fresh virtual machines; `needs:` orders them; `services:` starts containers such as Postgres next to a job; `${{ secrets.NAME }}` reads an encrypted secret.

## Why it matters at work

Most teams' first review question about a new service is "how does it run?", and this day is the answer: a Dockerfile that a reviewer can read in a minute, a CI pipeline that blocks a merge when tests, `go vet` or `govulncheck` fail, and health endpoints the platform team can point their load balancer at. The details get checked. A container running as root, an image of 1 GB with a compiler inside, secrets printed in CI logs, a workflow with write permissions it doesn't need (A08 in [owasp-checklist.md](../../reference/owasp-checklist.md)), or a liveness check that pings the database, which turns one database blip into every container restarting at once. Graceful shutdown matters on the first deploy: without it, every release drops the requests that were in flight, and some of them were M-Pesa payments.

## Steps

1. Open `starter/config.go` and write `Load`, following its doc comment. `intEnv` and `durationEnv` are given. Check `DATABASE_URL` with `url.Parse` for a `postgres` or `postgresql` scheme and a host. Don't put any value in an error message; name the variable instead.
2. Open `starter/health.go`. Write `NewCheck` to return a `checkFunc`.
3. Write `SetDraining` with `h.draining.Store(true)`, and `Liveness`, which always answers 200 `{"status":"ok"}`. Use the given `writeJSON`, which also sets `Cache-Control: no-store`.
4. Write `Readiness`. While `h.draining.Load()` is true, answer 503 `draining` without running any checks. Otherwise make `ctx` with `context.WithTimeout(r.Context(), h.Timeout)`, start one goroutine per check, collect `"ok"` or `"failed"` into a map behind a mutex, log each failure with `h.Log.Error`, and wait with a `sync.WaitGroup`. Any failure is 503 `not ready`.
5. Write `Register` for `GET /healthz` and `GET /readyz`. Run the tests:

```bash
go test -race ./phase-5-backend/day-063-docker-and-ci/starter
```

6. Finish `starter/cmd/api/main.go`, following its comments. Run it, then press Ctrl-C and watch it shut down cleanly. Try `APP_ENV=production` to see the config errors.

```bash
go run ./phase-5-backend/day-063-docker-and-ci/starter/cmd/api
curl -i localhost:8080/readyz
```

7. Write `starter/Dockerfile` from its outline, then build it from the **repo root** (the build needs `go.mod`) and run it:

```bash
docker build -f phase-5-backend/day-063-docker-and-ci/starter/Dockerfile -t twende-api .
docker run --rm -p 8080:8080 twende-api
docker images twende-api                                       # the size
docker inspect --format '{{.Config.User}}' twende-api          # nonroot:nonroot
```

   In your own repo, add a `.dockerignore` (with `.git`, `node_modules` and `.env`) so the build doesn't send everything to Docker, and so a local `.env` never ends up in an image.

8. Write `starter/ci.yml` from its outline. It's for the booking API once it has its own repository ([Day 64](../day-064-matatu-booking-api/)), where it goes in `.github/workflows/api.yml`. Check it parses: `python3 -c "import yaml; yaml.safe_load(open('phase-5-backend/day-063-docker-and-ci/starter/ci.yml'))"`. Then compare with `solution/ci.yml`.

### Secrets in CI

- Add secrets under the repository's **Settings, Secrets and variables, Actions**. Put deploy secrets in an **environment** (here `production`), which can also require a reviewer to approve each deploy.
- Pass a secret to a step through `env:`, then use the variable (`"$DEPLOY_HOOK"`). Don't paste `${{ secrets.X }}` into the middle of a shell command, and never `echo` it.
- Pull requests from forks get no secrets, so the `test` job mustn't need any. The Postgres password in the workflow is fine in plain text: that database lives for a few minutes and only the job can reach it.
- `permissions: contents: read` at the top means a compromised step can't push code or edit releases. The workflow only uses GitHub's own `actions/*`; pin any third-party action to a full commit SHA.

## The kata: Build Order

Given CI jobs and their `needs:`, print the order to run them in, smallest name first when several are ready, or `CYCLE`. A topological sort with a min-heap. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 63
```

## When you're stuck

- **`go: go.mod requires go >= 1.26.0 (running go 1.2x)`** during `docker build`: the builder image is older than `go.mod` needs. Use `golang:1.26` or newer.
- **`COPY failed: file not found in build context`** or **`go.mod file not found`**: you ran `docker build` from the day folder. Run it from the repo root and pass the Dockerfile with `-f`.
- **`exec /api: no such file or directory`** on distroless: the binary links to C libraries. Build with `CGO_ENABLED=0`.
- **`x509: certificate signed by unknown authority`** when the API calls M-Pesa from a `scratch` image: there are no CA certificates. Use distroless, which has them.
- **The HEALTHCHECK fails** with `curl: not found`: distroless has no curl. That's why the binary has its own `-healthcheck` flag.
- **`GET /readyz` is still waiting after 2s**: the checks don't get a context with a deadline, or you run them one after another. One goroutine each, all sharing `ctx`.
- **`DATA RACE` on the results map**: two goroutines write to it at once. Lock a mutex around the write.
- **Ctrl-C kills the server at once with no "stopped cleanly"**: `ListenAndServe` returns `http.ErrServerClosed` as soon as `Shutdown` starts. Don't treat that as a failure and exit; wait for `Shutdown` to return.
- **Still stuck?** Open the matching file in `solution/`, read it once, close it, then write your own version from memory.
