# Env variables and config

A reference for Day 21 onwards. It covers how a Go program reads its settings, how to keep secrets out of git, and what to do when one leaks anyway.

## What an env variable is

Every running program gets a list of `NAME=value` strings from whoever started it: your shell, Docker, a CI runner, a hosting platform. That list is the **environment**.

```bash
PORT=9000 go run ./cmd/api        # set PORT for this one command
export DATABASE_URL=postgres://... # set it for everything this shell starts
env | sort                         # see the whole list
```

Env variables are always strings. Your program decides what `"9000"` or `"true"` means.

## Reading them in Go

| Function | Returns | Use it when |
|---|---|---|
| `os.Getenv("PORT")` | the value, or `""` if it isn't set | an empty value and a missing one mean the same thing to you |
| `os.LookupEnv("PORT")` | the value and `ok`, which is `false` if it isn't set | you need to tell "not set" apart from "set to empty" |
| `os.Environ()` | every `NAME=value` pair as a `[]string` | debugging. Don't log it: it contains your secrets |

```go
port := os.Getenv("PORT") // "" if missing

if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
	// set, possibly to ""
}
```

## Required or optional

Sort every setting into one of two kinds:

- **Required**: the program can't work without it, and there's no safe default. `DATABASE_URL`, `JWT_SECRET`. If it's missing, stop.
- **Optional**: there's a sensible default. `PORT` defaults to `8080`, `LOG_LEVEL` to `info`.

Never give a secret a default. A default `JWT_SECRET` ends up in production one day, and everyone who has read your code knows it.

## Parsing ints, durations and bools

Parsing can fail, and the error should name the variable and show the bad value. One small helper per type keeps it tidy:

```go
func intEnv(key string, def int) (int, error) {
	s, ok := os.LookupEnv(key)
	if !ok || s == "" {
		return def, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number, got %q", key, s)
	}
	return n, nil
}

func durationEnv(key string, def time.Duration) (time.Duration, error) {
	s, ok := os.LookupEnv(key)
	if !ok || s == "" {
		return def, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration like 10s or 2m, got %q", key, s)
	}
	return d, nil
}

func boolEnv(key string, def bool) (bool, error) {
	s, ok := os.LookupEnv(key)
	if !ok || s == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false, got %q", key, s)
	}
	return b, nil
}
```

| Parser | Accepts | Rejects |
|---|---|---|
| `strconv.Atoi` | `8080`, `-3` | `8080 `, `8k`, `""` |
| `time.ParseDuration` | `300ms`, `10s`, `2m`, `1h30m` | `10` (no unit), `2 days` |
| `strconv.ParseBool` | `1`, `t`, `T`, `TRUE`, `true`, `True`, and the same for false | `yes`, `on`, `enabled` |

## Fail fast at startup

Read and check all config once, in `main`, before you open a port. Collect every problem, so one run tells you everything that's wrong, then exit.

```go
type Config struct {
	Port            int
	DatabaseURL     string
	JWTSecret       string
	ShutdownTimeout time.Duration
	Debug           bool
}

func Load() (Config, error) {
	var cfg Config
	var errs []error

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET is required and must be at least 32 characters"))
	}

	var err error
	if cfg.Port, err = intEnv("PORT", 8080); err != nil {
		errs = append(errs, err)
	}
	if cfg.ShutdownTimeout, err = durationEnv("SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		errs = append(errs, err)
	}
	if cfg.Debug, err = boolEnv("DEBUG", false); err != nil {
		errs = append(errs, err)
	}

	return cfg, errors.Join(errs...)
}

func main() {
	cfg, err := Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config:\n%v\n", err)
		os.Exit(1)
	}
	// only now: connect to the database, start the server
}
```

`errors.Join` returns `nil` when every error is `nil`, and prints one error per line otherwise:

```
config:
DATABASE_URL is required
PORT must be a whole number, got "abc"
```

Crashing at startup is the good outcome. The alternative is a server that starts, passes its health check, and fails on the first real request at 2 a.m.

After `Load`, pass `cfg` (or the fields a function needs) down to the code that uses it. Don't call `os.Getenv` deep inside handlers: it hides what the program depends on and makes tests harder.

## The 12-factor rule

The [Twelve-Factor App](https://12factor.net/config) says: **keep config in the environment, not in the code.** Config is anything that changes between deploys: database addresses, API keys, feature switches. The test is simple. Could you make the repo public right now without leaking a credential? If not, some config is in the code.

The same build then runs everywhere. Your laptop, CI and production differ only in their env variables.

## .env files for local development

Typing ten `export` lines every morning gets old. A `.env` file holds them instead:

```
# .env  (never committed)
DATABASE_URL=postgres://gopher:gopher@localhost:5432/matatu?sslmode=disable
JWT_SECRET=a-long-random-string-only-on-this-laptop
PORT=8080
```

**Go doesn't read `.env` files.** Something has to load it. You have three options:

1. Your shell: `set -a; source .env; set +a; go run ./cmd/api`.
2. A library: [`github.com/joho/godotenv`](https://github.com/joho/godotenv). Call `godotenv.Load()` at the top of `main`.
3. A tiny loader of your own, which is what Day 21 builds:

```go
// loadDotEnv sets variables from a KEY=value file, skipping blank lines and # comments.
// A variable that is already set wins, so the real environment beats the file.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil // no .env file: fine, use the real environment
	}
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s: bad line %q", path, line)
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, set := os.LookupEnv(key); !set {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}
```

A `.env` file is for your machine only. In Docker, CI and production, set real env variables instead.

### The .env.example convention

Commit a `.env.example` with every variable the app reads, fake values, and a comment for anything unclear. A new teammate copies it and fills it in:

```bash
cp .env.example .env
```

This repo has one at the root. Keep it in step with `Load`: when you add a variable to the code, add it to `.env.example` in the same commit.

## Never commit secrets

Put the real file in `.gitignore` before your first commit. This repo's `.gitignore` already has:

```
.env
.env.*
!.env.example
```

Check before you commit:

```bash
git status --short          # is .env listed? Then it isn't ignored
git diff --cached           # read what you're about to commit
git check-ignore -v .env    # prints the rule that ignores it
```

Secrets also leak through places that aren't `.env`: a test file with a real API key, a log line that prints the config, a screenshot, a pasted error in a chat. Treat them all the same way.

## If you leak a secret

Assume it's already been copied. Bots scan public GitHub pushes within minutes.

1. **Rotate first.** Revoke the leaked key or password and issue a new one, at the provider (M-Pesa Daraja, your database host, GitHub). Update the new value wherever the app runs. Until you do this, nothing else helps.
2. **Check for damage.** Look at the provider's logs for use you didn't make.
3. **Then purge it from history** if the repo is public or shared. Deleting the line in a new commit isn't enough: it's still in the old commits. Use [`git filter-repo`](https://github.com/newren/git-filter-repo) or [BFG](https://rtyley.github.io/bfg-repo-cleaner/), then force-push, and ask GitHub support to clear cached views if needed. Everyone else must re-clone.
4. **Stop it happening again.** Turn on GitHub secret scanning and push protection for the repo.

Purging without rotating is the common mistake. The old key still works for whoever already copied it.

## Secrets in CI

GitHub Actions keeps secrets encrypted under **Settings, Secrets and variables, Actions**. A workflow reads them with `${{ secrets.NAME }}`:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go test ./...
        env:
          DATABASE_URL: ${{ secrets.TEST_DATABASE_URL }}
```

- GitHub masks secret values in logs, but only exact matches. Don't print them, encode them, or echo parts of them.
- Workflows triggered by pull requests from forks don't get your secrets. That's on purpose.
- Give each job the least it needs: `permissions: contents: read` unless it must write.
- Non-secret settings (a region, a feature flag) go under **Variables**, read with `${{ vars.NAME }}`.

Day 63 sets this up for the booking API. Day 83 goes further: rotating keys on a schedule, vaults, and least privilege.

## Build-time vs runtime env for front ends

A Go server reads its env when it **starts**. A browser app can't: it's just files on a CDN, and the browser has no access to the server's environment. So front-end tools copy env values into the JavaScript **at build time**.

With Vite (Phase 6), only variables that start with `VITE_` are copied, and you read them with `import.meta.env`:

```ts
const apiURL = import.meta.env.VITE_API_URL;
```

**Everything with the `VITE_` prefix is public.** It's written into the bundle in plain text, and anyone can read it with the browser's dev tools. Never give a secret the prefix. A front end has no secrets: anything secret belongs in your Go API, which the front end calls.

| | Go API | Vite front end |
|---|---|---|
| When env is read | every time it starts | once, when you run `npm run build` |
| Change a value | restart the server | rebuild and redeploy |
| Can hold secrets | yes | **no**, everything is public |

Day 75 deploys the front end and sets `VITE_API_URL` on the host.

## Where the course uses this

| Day | What you meet |
|---|---|
| 21 | the config loader: `os.LookupEnv`, defaults, parsing, fail fast, `.env` |
| 26 | the budget CLI, configured by env and flags |
| 63 | Docker and CI: env in containers, GitHub Actions secrets |
| 75 | build-time env for the front end, `VITE_` is public |
| 83 | secrets management: rotation, vaults, least privilege |
