# Day 60: OWASP Top 10

## The brief

Twende (a made-up matatu SACCO app) hired a security tester before launch. Her report has five findings, one for each file in today's folder. Each is a real bug class from the [OWASP Top 10](https://owasp.org/Top10/), and each has a standard Go fix. Keep [owasp-checklist.md](../../reference/owasp-checklist.md) open beside you: it has the full list and today's code follows it.

| File | Finding | OWASP |
|---|---|---|
| `sql.go` | trip search builds SQL with `fmt.Sprintf`: `' OR '1'='1` returns every row | A03 Injection |
| `access.go` | `PATCH /me` decodes the body into the whole profile: `{"role":"admin"}` makes you an admin | A01 Broken Access Control |
| `ssrf.go` | the webhook feature fetches any URL, including `http://169.254.169.254/`, the cloud's secret metadata | A10 SSRF |
| `xss.go` | the ticket page uses `text/template`: a passenger named `<script>...` runs code in the conductor's browser | A03 Injection (XSS) |
| `secrets.go` | startup logs the config, database password and JWT secret included | A09 Logging Failures |

```
BuildTripSearch(TripFilter{RouteID: 111, Name: "' OR '1'='1"})
  →  "SELECT ... WHERE t.route_id = $1 AND r.name ILIKE $2 ORDER BY t.departs ASC LIMIT $3",
     []any{111, "%' OR '1'='1%", 20}

PATCH /me {"name":"Wanjiku Kamau"}                 →  200
PATCH /me {"name":"W","role":"admin"}              →  400  {"error": "body must be JSON with only a name"}

CheckURL("https://api.safaricom.co.ke/oauth", allowed)     →  ok
CheckURL("http://169.254.169.254/latest/meta-data/", ...)  →  ErrBlocked
safeClient.Get("https://127.0.0.1:8443/")                  →  error wrapping ErrBlocked, nothing sent

RenderTicket(w, Ticket{Passenger: "<script>alert(1)</script>", ...})
  →  <h1>Tiketi ya &lt;script&gt;alert(1)&lt;/script&gt;</h1> ...

slog.Info("starting", "cfg", cfg)
  →  {"msg":"starting","cfg":{"port":8080,"database":"postgres://gopher:xxxxx@db.internal:5432/matatu"}}
```

## What you'll use

- **Placeholders** (`$1`, `$2`) for every SQL value. The driver sends the values separately, so they can never change the query. Placeholders work only for values, not column names, so a sort column comes from an allowlist map. You'll test that the query *builder* emits placeholders: no database needed.

```go
// broken: the input becomes part of the SQL
q := fmt.Sprintf("SELECT ... WHERE r.name ILIKE '%%%s%%'", name)
// fixed: the input is an argument
q, args := "SELECT ... WHERE r.name ILIKE $1", []any{"%" + name + "%"}
rows, err := db.QueryContext(ctx, q, args...)
```

- **A narrow request struct.** Decode only the fields a user may set, with `DisallowUnknownFields`, so `role` is rejected instead of copied. Take *whose* profile from the logged-in user, never from the body or URL.
- **`net/netip`**: `netip.Addr` has `IsLoopback`, `IsPrivate`, `IsLinkLocalUnicast` and friends. `Unmap()` turns `::ffff:127.0.0.1` back into `127.0.0.1`, so the IPv6 spelling can't sneak past. `netip.MustParsePrefix("100.64.0.0/10").Contains(ip)` checks a range.
- **A dialer `Control` hook.** `net.Dialer.Control` runs after DNS, with the real IP about to be dialled. Checking there catches `evil.example` that resolves to `127.0.0.1`, which a check on the hostname alone misses. `http.Client.CheckRedirect` checks every redirect too.
- **`html/template`** has the same API as `text/template` but escapes each value for where it lands. In HTML text `<` becomes `&lt;`. In an `href`, a `javascript:` URL becomes `#ZgotmplZ`, a harmless marker.
- **`slog.LogValuer`**, **`fmt.Stringer`** and **`json.Marshaler`**: give a type `LogValue()`, `String()` and `MarshalJSON()` methods and slog, fmt and encoding/json all call them instead of printing the raw value. `url.URL.Redacted()` swaps a URL's password for `xxxxx`.
- **`govulncheck`** reads your code and `go.sum` and reports known vulnerabilities in code you actually call (A06).

## Why it matters at work

These five bugs are what security reviews and bug-bounty reports are made of, and every one has slipped past good developers in a hurry. In code review, `fmt.Sprintf` next to `SELECT`, `json.Decode` into a model with a `Role` field, `http.Get(userURL)`, `text/template` writing HTML, and `slog.Info("cfg", "cfg", cfg)` should each make you stop. The fixes are cheap when written in, expensive after an incident: a leaked database password means rotating it everywhere it runs, and a stolen metadata token can mean the whole cloud account. A test for each attack, like the ones here, keeps the fix from being undone by the next refactor.

## Steps

1. **SQL** (`sql.go`). Write `BuildTripSearch`. Look up `f.Sort` in `sortColumns`; unknown is `ErrBadSort`. For each filter that's set, append the value to `args` and a condition ending in `$` plus `len(args)` to a `where` slice. Join with `" AND "`. Add `ORDER BY`, `ASC` or `DESC`, then the limit (0 or less means 20, over 100 means 100) as the last placeholder. The tests give the exact strings.
2. **Access control** (`access.go`). Write `UpdateProfile`. 401 with `WWW-Authenticate: Bearer` if `PhoneFrom` finds nothing. Load the profile by that phone. Decode into `profileUpdate` with `DisallowUnknownFields` (400 on any error), check the name is 1 to 50 characters (422), set only `Name`, save, and answer 200 with the profile.
3. **SSRF** (`ssrf.go`). Write `IsBlockedIP`: `Unmap` first, then the `netip` methods, then `extraBlocked`. An invalid (zero) address is blocked. Write `CheckURL`: parse, then require `https`, no `u.User`, and `allowed[u.Hostname()]`. Write `NewSafeClient` with a `net.Dialer` whose `Control` parses the address with `netip.ParseAddrPort` and returns `ErrBlocked` for a blocked IP, plus `Timeout` and a `CheckRedirect` that allows at most 3 hops, each through `CheckURL`. The reference checklist's A10 section has the shape.
4. **XSS** (`xss.go`). Write `RenderTicket` with `html/template`. Parse the template from the doc comment once, in a package-level `var` with `template.Must`, and `Execute` it.
5. **Secrets** (`secrets.go`). Give `Secret` its three methods, each returning `[redacted]`. Write `RedactURL` with `url.Parse` and `Redacted()`. Write `Config.LogValue` with `slog.GroupValue(slog.Int("port", ...), slog.String("database", ...))`.
6. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-060-owasp-top-10/starter
```

7. Run `govulncheck` on the whole repo (this needs the network to download it and the vulnerability database):

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

"No vulnerabilities found" is the answer you want. If it reports one, it names the module, the fixed version and the line in your code that reaches it. Upgrade with `go get module@version` and run it again. [Day 63](../day-063-docker-and-ci/) runs it in CI on every push.

## The kata: Blocked Addresses

Parse IPv4 addresses by hand, reject malformed ones (including leading zeros, an old filter-bypass trick), and check each against a list of CIDR ranges with bit masks. It's what `IsBlockedIP` and `netip.Prefix` do inside. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 60
```

## When you're stuck

- **The query has `$1` twice, or `LIMIT $1` after two filters**: number each placeholder from `len(args)` right after appending, not from a counter you forget to bump.
- **`args` has `%rongai%` but the test wants it first**: append in the order the conditions appear: route, time, name, then the limit.
- **`TestBuildTripSearchSortAllowlist` fails for `SEATS`**: good instinct to be lenient, but the allowlist is exact. Unknown means error, not "fix it up".
- **`PATCH` with `"role"` answers 200**: you decoded into `Profile`. Decode into `profileUpdate`, and call `dec.DisallowUnknownFields()` before `Decode`.
- **`IsBlockedIP(::ffff:127.0.0.1) = false`**: call `ip.Unmap()` before the checks.
- **The safe client's error doesn't wrap `ErrBlocked`**: return `ErrBlocked` itself from `Control`. The `*url.Error` and `*net.OpError` around it both unwrap, so `errors.Is` finds it.
- **`href="javascript:alert(document.cookie)"` is still in the page**: you imported `text/template`. Change the import to `html/template`; nothing else changes.
- **`fmt.Sprintf("%v", cfg)` prints the password**: `String()` must have a value receiver, `func (Secret) String() string`, so it applies to the field inside a `Config` value.
- **`%!v(PANIC=String method: ...)`** or an infinite loop: inside `String`, don't print `s` with `fmt`, because that calls `String` again. Return the constant.
- **Still stuck?** Open the matching file in `solution/`, read it once, close it, then write your own version from memory.
