# Day 83: Secrets Management

## The brief

Twende has a handful of secrets that could ruin the SACCO in an afternoon: the Postgres password, the M-Pesa Daraja keys and the key that signs passengers' login tokens ([Day 58](../../phase-5-backend/day-058-sessions-and-jwts/)). Right now they're env variables that every service can read, nobody has ever changed them, and last week one turned up in a log line. Today you load secrets from proper sources, make them impossible to log by accident, give each service only the secrets it needs, and rotate the signing key without logging out every passenger.

```
Chain{Env{Prefix: "TWENDE_"}, Files{FS: os.DirFS("/run/secrets")}}.Get(ctx, "DB_PASSWORD")
    →  TWENDE_DB_PASSWORD if it's set, otherwise the file /run/secrets/DB_PASSWORD

log.Info("config", "db_password", secret)   →  {"msg":"config","db_password":"[REDACTED]"}
fmt.Printf("%+v", cfg)                      →  {User:twende Password:[REDACTED]}

vault.Read("tok-booking",  "twende/booking-api/DB_PASSWORD")   →  ok
vault.Read("tok-booking",  "twende/payments/DARAJA_KEY")       →  ErrForbidden   (least privilege)

ring.Sign("user:42")           →  "k1.yG2q..."
ring.Rotate(Key{ID: "k2", ...})
ring.Verify("user:42", "k1.yG2q...")   →  nil    (old tokens still work)
ring.Sign("user:42")           →  "k2...."     (new tokens use the new key)
```

## What you'll use

**Where secrets come from.** Read [env-and-config.md](../../reference/env-and-config.md) first; today goes further.

- **Env variables** are simple and fine for local work. They're also visible to every process the app starts and are easily dumped into crash reports.
- **Mounted files.** Docker secrets appear as `/run/secrets/<name>`; a Kubernetes Secret volume puts one file per key in a folder. Only the container that mounts them can read them. Files often end with a newline, so strip exactly one.
- **A vault**: HashiCorp Vault or a cloud secrets manager. It keeps versions, logs every read, gives each service a token or identity with its own permissions, and can hand out short-lived database passwords. Today's `FakeVault` does the first three in memory.
- **`Chain`** tries sources in order. Stop on a real error such as "forbidden": falling through to an older copy elsewhere hides the problem.

**`fs.FS` for testable file reading.** `Files` reads from an `fs.FS`. Production passes `os.DirFS("/run/secrets")`; tests pass `fstest.MapFS{"DB_PASSWORD": {Data: []byte("pg-pass\n")}}`. Reject names containing a slash, `.` or `..`, so a name can never climb out of the folder.

**A type that won't print.** `type Secret string` with three methods:

```go
func (s Secret) String() string        { return "[REDACTED]" } // fmt %s, %v
func (s Secret) GoString() string      { return "[REDACTED]" } // fmt %#v
func (s Secret) LogValue() slog.Value  { return slog.StringValue("[REDACTED]") } // slog.LogValuer
func (s Secret) Reveal() string        { return string(s) }    // the one way out
```

Any `slog` handler calls `LogValue` before writing, so the secret is safe whichever logger someone uses. Code that really needs the value calls `Reveal()`, which a reviewer can search for.

**Least privilege.** Each service gets a token that can read only its own folder: the booking API reads `twende/booking-api/*`, the payments worker reads `twende/payments/DARAJA_KEY`. A leaked booking-API token can't pay anyone. Answer "forbidden", not "not found", for paths outside the grant, so a stolen token can't map what exists. Audit every attempt, especially the denied ones.

**Key rotation with two keys.** If you swap a signing key in one step, every token signed with the old key stops working at once. Instead, a `Keyring` keeps two:

1. Sign with the **current** key and put its ID in front: `k2.<hmac>`.
2. Verify with whichever key the ID names, current or **previous**.
3. To rotate, make a new key current. The old one becomes previous and keeps verifying until the tokens it signed have expired. The rotation after that retires it.

Compare MACs with `hmac.Equal`, which takes the same time however many bytes match, so an attacker can't guess a signature byte by byte from timings. JWT libraries do the same thing with the `kid` header.

**After a leak.** Rotate first, then look for damage, then clean up history and add scanning, in that order (the full list is in [env-and-config.md](../../reference/env-and-config.md#if-you-leak-a-secret)). A keyring and a vault make the first step a five-minute job instead of an outage.

## Why it matters at work

Leaked credentials are one of the most common ways companies are breached, and the leak is rarely dramatic: a key in a log line, in a Slack paste, in a test file, in a screenshot. Reviewers look for secrets typed as plain `string` in config structs, `fmt.Printf("%+v", cfg)` at startup, services that share one all-powerful database user, and keys with no plan for rotation. "How would we rotate this tonight if it leaked?" is a standard design-review question, and "redeploy everything and log everyone out" is not a good answer.

## Steps

1. Open `starter/secrets.go` and `starter/keyring.go`. The errors, types, `Put`, `Grant`, `Audit`, `Client` and `mac` are given.
2. Write `String`, `GoString` and `LogValue` on `Secret`.
3. Write `Env.Get`. Use `os.LookupEnv` when `Lookup` is nil. Not set or empty is `fmt.Errorf("%w: env %s", ErrNotFound, key)`.
4. Write `Files.Get`. Check the name, `fs.ReadFile`, turn `fs.ErrNotExist` into `ErrNotFound`, and trim one `"\n"` and then one `"\r"` with `strings.TrimSuffix`.
5. Write `Chain.Get`.
6. Write `allowed` (`strings.CutSuffix(p, "/*")` gives the folder; match `folder + "/"` as a prefix) and `FakeVault.Read`: audit first, then the grant check, then the latest version.
7. Write `Rotate`, `Sign` and `Verify`. `strings.Cut(sig, ".")` splits off the key ID. Hold the lock only while you choose the key.
8. Run the tests from the repo root:

```bash
go test -race ./phase-7-production/day-083-secrets-management/starter
```

## The kata: Leak Scanner

Blank out passwords, live API keys and bearer tokens in log lines before they leave the building, and count how many you caught. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 83
```

## When you're stuck

- **`fmt.Sprintf("%#v") printed the secret`**: `%#v` ignores `String`; it calls `GoString`.
- **`json log = {"daraja_key":"sk_live_..."}`**: `LogValue` must have exactly the signature `LogValue() slog.Value` for slog to find it.
- **`Get("windows") = "abc\r"`**: trim `"\n"` first, then `"\r"`.
- **`Get("two_newlines") = "keep"`**: you trimmed every newline with `TrimRight`. Remove exactly one.
- **`Get("sub/hidden") = "nope", want ErrBadName`**: `fs.FS` happily opens paths with slashes. Reject any name containing `/` or `\`.
- **`Read(tok-booking, twende/booking-api-admin/ROOT)` succeeds**: you matched the prefix `twende/booking-api` without the trailing slash.
- **`Read(tok-nobody, ...) = ErrNotFound, want ErrForbidden`**: check the grant before checking whether the secret exists.
- **`after one Rotate, an old-key signature = bad signature`**: `Rotate` must keep the old current key as `previous`.
- **The race detector complains in `TestKeyringIsSafeForGoroutines`**: every read of `current` and `previous` needs `r.mu.RLock()`, every write `r.mu.Lock()`.
- **Still stuck?** Open `solution/secrets.go` and `solution/keyring.go`, read them once, close them, then write your own version from memory.
