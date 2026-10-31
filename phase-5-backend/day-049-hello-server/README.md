# Day 49: Hello, Server

## The brief

Phase 5 builds one thing: a booking API for matatus. A passenger picks a route, a trip and a seat, and books it from their phone. Over the next sixteen days it gets methods, status codes, JSON, validation, a database, logins and a deploy.

Today it answers its first requests. It's a small server, but it's built the way a production server is: it reads its port from the environment, it has timeouts, and when you stop it, it finishes the requests it already started before it exits.

```
GET /                 →  200  Karibu! The matatu booking API is running.
GET /health           →  200  ok
GET /hello/Wanjiku    →  200  Habari, Wanjiku!
GET /nope             →  404  404 page not found
POST /health          →  405  Method Not Allowed
```

In a terminal:

```
$ PORT=9000 go run ./phase-5-backend/day-049-hello-server/solution/cmd/server
2026/10/24 08:15:02 listening on [::]:9000 (development)
^C2026/10/24 08:15:40 stopped cleanly

$ curl -i localhost:9000/hello/Wanjiku
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff

Habari, Wanjiku!
```

## What you'll use

- **A handler** is any function with the signature `func(w http.ResponseWriter, r *http.Request)`. `r` is what the client sent. You write the answer to `w`.
- **Headers** are the `Name: value` lines before the body, like the `Content-Type` line in the `curl -i` output above. `r.Header.Get("Name")` reads one from the request; `w.Header().Set("Name", "value")` sets one on the response. Set them before the first `w.Write`, because that sends the headers, and anything set afterwards is silently dropped. [The headers section of http.md](../../reference/http.md#headers) lists every header the course uses and the day you practise it.
- **`http.ServeMux`** picks the handler for each request from a pattern. Since Go 1.22 a pattern can start with a method and contain `{name}` wildcards, which `r.PathValue("name")` reads. `{$}` means "this exact path and nothing below it". Without it, `"GET /"` would match every path.
- **`http.Server`** holds the handler, the address and the timeouts. `http.ListenAndServe` makes one with no timeouts at all, so a slow client can hold a connection open forever. Don't use it for a real server.
- **`net.Listen` and `srv.Serve`** are the two halves of `ListenAndServe`: open the port, then serve on it. Splitting them lets a test listen on port 0, which asks the operating system for any free port.
- **`signal.NotifyContext`** (new today) gives you a context that's cancelled when the process gets Ctrl+C (`os.Interrupt`) or `SIGTERM`, which Docker and hosting platforms send to stop your app. It's the [Day 44](../../phase-4-concurrency/day-044-context-cancel/) context idea, plugged into the operating system.
- **`srv.Shutdown(ctx)`** closes the listener, waits for running requests to finish, and gives up when `ctx` expires. `Serve` then returns `http.ErrServerClosed`, which means "stopped on purpose", not a failure.

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Habari, %s!\n", r.PathValue("name"))
})

ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
<-ctx.Done() // blocks until Ctrl+C
```

Keep [reference/http.md](../../reference/http.md) open for the whole phase. The config follows [reference/env-and-config.md](../../reference/env-and-config.md) and [Day 21](../../phase-2-modelling/day-021-config-loader/): read it once at startup, fail fast, and pass it down.

## Why it matters at work

Every deploy stops the old copy of your server and starts a new one. A server without graceful shutdown drops whatever it was doing at that moment: a passenger's booking, an M-Pesa callback. With it, the platform sends `SIGTERM`, the server stops taking new connections, finishes what's running and exits, and nobody notices the deploy. Missing timeouts are the other thing a reviewer will flag on sight. Go's `net/http` has none by default, and a few clients that open connections and never finish sending can use up a server's connections (an attack called Slowloris). `ReadHeaderTimeout` is the one that stops it, and `go vet` won't warn you that it's missing.

## Steps

1. Open `starter/server.go`. Write `LoadConfig`. Start from the defaults, `Config{Port: 8080, AppEnv: "development"}`. Read `PORT` with `os.Getenv`; if it isn't empty, parse it with `strconv.Atoi` and check it's from 1 to 65535. Read `APP_ENV` and accept only `development`, `staging` or `production`. Collect each problem in a `[]error`, and at the end return `Config{}` and `errors.Join(errs...)` if there were any. Each message names the variable and quotes the bad value with `%q`.
2. Write `Addr`: `":"` followed by the port.
3. Write `Routes`. Register `"GET /{$}"`, `"GET /health"` and `"GET /hello/{name}"` on the mux with `mux.HandleFunc`. Each handler sets `Content-Type: text/plain; charset=utf-8` and `X-Content-Type-Options: nosniff` with `w.Header().Set`, then writes its text with `fmt.Fprint(w, ...)`. A small `writeText(w, s)` helper saves repeating it.
4. Write `NewServer`: return a `&http.Server{...}` with `Addr`, `Handler` and the four timeouts from the doc comment.
5. Write `Run`. Start `srv.Serve(ln)` in a goroutine that sends its error on a channel with a buffer of 1. Then `select` on that channel and on `ctx.Done()`. If `Serve` returned first, return its error. If the context was cancelled, make a new context with `context.WithTimeout(context.Background(), shutdownTimeout)` and call `srv.Shutdown` with it. Return Shutdown's error wrapped with `%w`, so callers can check it with `errors.Is`. Finally read the channel: `http.ErrServerClosed` means all is well, so return `nil`.
6. Write `main` in `starter/cmd/server/main.go`, following the comment there.
7. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-049-hello-server/starter
```

8. Run your server, and from a second terminal ask it something. Then press Ctrl+C in the first terminal and watch it stop cleanly:

```bash
go run ./phase-5-backend/day-049-hello-server/starter/cmd/server
curl -i localhost:8080/hello/Otieno
curl -i -X POST localhost:8080/health
```

## The kata: Request Line Parser

Go's server reads `GET /trips?route=125 HTTP/1.1` and fills in `r.Method`, `r.URL.Path` and the query for you. In the kata you write that parser yourself, strictly: wrong spacing, an unknown method or an unsupported version gets the right error code. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 49
```

## When you're stuck

- **`listen tcp :8080: bind: address already in use`**: another program, often your server from earlier, already has the port. Stop it, or run with `PORT=9000`.
- **Every path answers with your home page**: you registered `"GET /"`. A pattern ending in `/` matches everything below it. Use `"GET /{$}"`.
- **`panic: pattern "GET /health" ... conflicts with pattern ...`**: you registered the same pattern twice. Each method and path pair can be registered once.
- **`r.PathValue("name")` returns `""`**: the name in the pattern and the name you ask for must match exactly, `{name}` and `"name"`.
- **`TestRunWaitsForInFlightRequests` says Run returned while a request was still running**: you returned straight after cancelling, or you called `srv.Close()`, which cuts connections off. `Shutdown` is the graceful one.
- **`Run returned ... want nil` after a clean shutdown**: `Serve` always returns an error, and after `Shutdown` it's `http.ErrServerClosed`. Check for it with `errors.Is` and return `nil`.
- **Shutdown gives up straight away**: you passed it `ctx`, which is already cancelled. Give it a fresh context from `context.Background()`.
- **`LoadConfig` errors with an empty `PORT`**: an empty value means "use the default". Only parse it when it isn't `""`.
- **Still stuck?** Open `solution/server.go` and `solution/cmd/server/main.go`, read them once, close them, then write your own version from memory.
