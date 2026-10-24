# Day 21: Config Loader

## The brief

A chama in Nakuru wants a service that sends members an M-Pesa payment prompt on contribution day. It talks to a Postgres database and to Safaricom's Daraja API, which gives every app a consumer key and a consumer secret. The first version had them typed into `main.go`. Then the repo went public.

Build the config loader the rewrite needs. Every setting comes from an env variable. Required ones have no defaults. Optional ones do. Every value is checked at startup, and all the problems are reported in one go, so you fix them in one go. For local development it reads a `.env` file, which never gets committed. And when the config is printed, the secrets stay hidden.

```
Load(lookup)   with PORT=9000, DEBUG=true and the three required variables
  →  Config{Env: "development", Port: 9000, RequestTimeout: 5s, Debug: true, ...}, nil

Load(lookup)   with APP_ENV=production, PORT=abc, DEBUG=true and nothing else
  →  error:
     PORT must be a whole number, got "abc"
     DATABASE_URL is required
     DARAJA_CONSUMER_KEY is required
     DARAJA_CONSUMER_SECRET is required
     DEBUG can't be true when APP_ENV is production

ParseDotEnv(strings.NewReader("PORT = 8081\nexport DEBUG=true\n"))
  →  map[DEBUG:true PORT:8081], nil

fmt.Println(cfg)
  →  APP_ENV=production
     PORT=8080
     DATABASE_URL=postgres://chama:xxxxx@db.internal:5432/chama
     DARAJA_CONSUMER_KEY=****
     ...
```

Read [reference/env-and-config.md](../../reference/env-and-config.md) before you start. It's the long version of today, and you'll come back to it in Phases 5 to 7.

## What you'll use

- **`os.LookupEnv(key)`** returns the value and whether it's set. Today's rule: set to `""` counts as not set, because that's what people mean when they write `PORT=` in a file.
- **Injecting the lookup.** `Load` doesn't call `os.LookupEnv` itself. It takes a `LookupFunc`, a function with the same shape. `main` passes `os.LookupEnv`; tests pass a function backed by a map. The tests never touch your real environment, so they pass on every machine and can run in parallel. This is the most useful idea of the day.

```go
type LookupFunc func(key string) (string, bool)

cfg, err := Load(os.LookupEnv) // the real program
cfg, err := Load(FromMap(map[string]string{"PORT": "9000"})) // a test
```

- **Functions that return functions.** `FromMap` and `Layered` build a `LookupFunc` and return it. The function they return can use their parameters later. Go calls that a closure.
- **Parsing with good errors**: `strconv.Atoi` for whole numbers, `time.ParseDuration` for `5s` or `1m30s`, `strconv.ParseBool` for `true`, `false`, `1`, `0`. Every error names the variable and quotes the bad value.
- **`errors.Join(errs...)`** joins many errors into one, with a newline between them. It returns `nil` if there are none, so `return cfg, errors.Join(errs...)` works on the happy path too.
- **`io.Reader`**: an interface (Day 16) for anything you can read bytes from: a file, a network connection, or `strings.NewReader("...")` in a test. `bufio.NewScanner` takes any `io.Reader`.
- **`url.Parse` and `(*url.URL).Redacted()`** from `net/url`: `Redacted` returns the URL with the password replaced by `xxxxx`.
- **`String() string`**: any type with this method satisfies `fmt.Stringer`, and `fmt.Println(cfg)` calls it. That's where you hide the secrets, so nobody can print them by accident.
- **`os.Args`**: the words typed after the command. `os.Args[1]` is the first one. Day 22 does this properly with the `flag` package.

## Why it matters at work

Config bugs are some of the most common outages, and they're avoidable. A reviewer will look for four things in a PR like this. Does the service fail at startup, with a clear message, when a variable is missing or wrong, instead of starting up and failing on the first real request? Are secrets left without defaults, kept out of the repo, and kept out of the logs? Does `os.Getenv` appear only in one place, so it's clear what the service depends on? Is `.env.example` up to date with every variable the code reads? A leaked Daraja secret means anyone can start payment prompts as your business. The fix starts with rotating the key, not deleting the commit. The reference page has the full checklist.

## Steps

1. Open `starter/config.go` and read the doc comment on `Load`. It is the spec, down to every error message.
2. Write `FromMap` first. Return a function literal: `return func(key string) (string, bool) { v, ok := m[key]; return v, ok }`.
3. Write `Layered`: loop over the lookups and return the first value found.
4. Write `Load` one variable at a time. Start with `var errs []error`. For each optional variable, set the default, then override it if the variable is set and valid; if it's invalid, `append` an error and carry on. For the three required ones, a loop over a small table of `{key, &field}` saves repeating yourself. End with the production check and `return cfg, errors.Join(errs...)`.
5. Write `ParseDotEnv`. Count lines as you scan, so errors can say `line 3:`. `strings.Cut(line, "=")` splits at the first `=` only, which keeps `s3cr=t` whole.
6. Write `String` with `fmt.Sprintf`. Use `url.Parse` and `Redacted()` for the database URL.
7. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-021-config-loader/starter
```

8. Finish `starter/cmd/showconfig/main.go`, then try it for real. Copy the example file and run it:

```bash
cp phase-2-modelling/day-021-config-loader/.env.example phase-2-modelling/day-021-config-loader/.env
go run ./phase-2-modelling/day-021-config-loader/starter/cmd/showconfig phase-2-modelling/day-021-config-loader/.env
PORT=abc APP_ENV=production go run ./phase-2-modelling/day-021-config-loader/starter/cmd/showconfig phase-2-modelling/day-021-config-loader/.env
echo $?    # 1
```

9. Prove `.env` won't be committed: `git status --short` doesn't list it, and `git check-ignore -v phase-2-modelling/day-021-config-loader/.env` names the `.gitignore` rule that ignores it.

## The kata: Expand the .env

Read a `.env` file whose values refer to each other with `${NAME}`, and print every variable fully expanded. It's careful string walking, the same work a real `.env` library does. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 21
```

## When you're stuck

- **`TestLoadReportsEverything` shows only the first problem**: you returned as soon as one check failed. Append to `errs` and keep going; return once, at the end.
- **The errors come out in a different order**: check the variables in the order the doc comment lists them. The production check goes last.
- **`PORT=""` gives `PORT must be a whole number, got ""`**: an empty value counts as not set. Check `ok && v != ""`.
- **`REQUEST_TIMEOUT=5` is accepted**: it isn't, by `time.ParseDuration` (`missing unit in duration`). Make sure you return the error rather than ignore it.
- **`DARAJA_CONSUMER_SECRET` shows as `'s3cr`**: you split on every `=`. `strings.Cut` splits once.
- **`panic: runtime error: invalid memory address or nil pointer dereference` in your `showconfig`**: you called a `LookupFunc` that is `nil`. Finish `FromMap` and `Layered` before you run the command.
- **Your `.env` values are ignored when the variable is also exported in your shell**: that's on purpose. The real environment wins, so a value you set for one run beats the file.
- **`git status` lists `.env`**: you named it something else, like `env` or `local.env`. The ignore rules cover `.env` and `.env.*`, except `.env.example`.
- **Still stuck?** Open `solution/config.go` and `solution/cmd/showconfig/main.go`, read them once, close them, then write your own version from memory.
