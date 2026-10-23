# OWASP Top 10 checklist for a Go API

The [OWASP Top 10](https://owasp.org/Top10/) is a list of the ten most common and serious web security risks. This page uses the **2021 edition** (A01 to A10) and turns each item into a checklist for the matatu-booking API you build in Phase 5. Newer editions reorder and merge some items, but the defences below still apply.

For each item: what it is, how it shows up in Go code, and the fix. Days 57 to 61 build these defences one by one.

## The checklist at a glance

| | Risk | The Go fix in one line |
|---|---|---|
| A01 | Broken Access Control | check roles in one middleware, and scope every query to the user |
| A02 | Cryptographic Failures | bcrypt or argon2id for passwords, `crypto/rand` for tokens, TLS everywhere |
| A03 | Injection | `$1` placeholders in SQL, `html/template`, no shell in `os/exec` |
| A04 | Insecure Design | threat-model, enforce business limits on the server, idempotency keys |
| A05 | Security Misconfiguration | server timeouts, security headers, strict CORS, no debug endpoints |
| A06 | Vulnerable and Outdated Components | `govulncheck ./...` in CI, keep Go and modules updated |
| A07 | Identification and Authentication Failures | rate-limit logins, expire sessions, rotate refresh tokens |
| A08 | Software and Data Integrity Failures | commit `go.sum`, pin CI actions, verify webhook signatures |
| A09 | Security Logging and Monitoring Failures | log security events with `slog`, never log secrets |
| A10 | Server-Side Request Forgery | allowlist hosts, block private IPs at dial time |

## A01: Broken Access Control

**What it is.** A user can do or see something they shouldn't: another passenger's booking, an admin-only action.

**How it shows up in Go.** A handler that loads a row by the ID in the URL and never asks whose it is:

```go
// GET /bookings/{id}: any logged-in user can read any booking
db.QueryRowContext(ctx, `SELECT id, seats FROM bookings WHERE id = $1`, id)
```

**The fix.**

- [ ] Put role checks in **one** middleware, not copied into each handler.
- [ ] Scope every query to the logged-in user, so the database itself refuses other people's rows.
- [ ] Deny by default: a new route needs a role before it does anything.
- [ ] Return `404`, not `403`, when admitting a row exists would leak something.
- [ ] Never trust a user ID, role or price sent in the request body. Take them from the session and the database.

```go
func requireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := userFrom(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "log in first")
			return
		}
		if u.Role != role {
			writeError(w, http.StatusForbidden, "not allowed")
			return
		}
		next(w, r)
	}
}

mux.HandleFunc("POST /trips", requireRole("admin", createTrip))
```

```go
err := db.QueryRowContext(ctx,
	`SELECT id, seats FROM bookings WHERE id = $1 AND user_id = $2`,
	bookingID, user.ID,
).Scan(&b.ID, &b.Seats)
if errors.Is(err, sql.ErrNoRows) {
	writeError(w, http.StatusNotFound, "no such booking") // yours or nobody's
	return
}
```

Day 59 builds this.

## A02: Cryptographic Failures

**What it is.** Sensitive data is stored or sent without the right protection: plain-text passwords, guessable tokens, HTTP instead of HTTPS.

**How it shows up in Go.** Hashing passwords with `sha256.Sum256`, which is far too fast to resist guessing. Making tokens with `math/rand`, which is predictable. Comparing secrets with `==`, which can leak timing.

**The fix.**

- [ ] Hash passwords with bcrypt or argon2id from `golang.org/x/crypto`. Never store or log the plain password.
- [ ] Make every token, session ID and reset code with `crypto/rand`.
- [ ] Compare secrets with `crypto/subtle.ConstantTimeCompare` or `hmac.Equal`.
- [ ] Serve only HTTPS in production, with HSTS (see A05).
- [ ] Don't write your own crypto.

```go
import "golang.org/x/crypto/bcrypt"

hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// store string(hash)

err = bcrypt.CompareHashAndPassword(hash, []byte(attempt))
if err != nil {
	// wrong password: answer exactly as you would for an unknown phone number
}
```

bcrypt only uses the first 72 bytes of a password, and Go's version returns an error for longer ones. Set a maximum length in your validation.

argon2id is the newer choice. `argon2.IDKey(password, salt, time, memory, threads, keyLen)` takes memory in KiB. OWASP's [Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) suggests at least 19 MiB, 2 passes and 1 thread: `argon2.IDKey(pw, salt, 2, 19*1024, 1, 32)`. Make a fresh 16-byte salt with `crypto/rand` for each password and store it, and the parameters, next to the hash.

```go
import "crypto/rand"

token := rand.Text() // 26 random base32 characters, Go 1.24 and later
```

Day 57 builds password hashing. Day 58 builds tokens.

## A03: Injection

**What it is.** Input from a user is treated as code: SQL, HTML, JavaScript or a shell command. Cross-site scripting (XSS) is in this group.

**How it shows up in Go.** Building SQL with `fmt.Sprintf`, rendering HTML with `text/template`, or passing input to `sh -c`:

```go
// a phone number of  ' OR '1'='1  returns every user
q := fmt.Sprintf("SELECT * FROM users WHERE phone = '%s'", phone)
```

**The fix.**

- [ ] Use placeholders for every value. Postgres uses `$1`, `$2`. The driver sends values separately from the SQL, so they can't change it.
- [ ] For a dynamic `ORDER BY` column, pick from an allowlist. Placeholders work for values, not for names.
- [ ] Render HTML with `html/template`, which escapes by context. Never `text/template` for HTML.
- [ ] Don't wrap user input in `template.HTML`: that switches the escaping off.
- [ ] Call `exec.Command("program", arg1, arg2)` with separate arguments. Never `exec.Command("sh", "-c", userInput)`.
- [ ] Open user-named files with `os.OpenInRoot(dir, name)` (Go 1.24 and later), so `../../etc/passwd` can't escape the folder.

```go
// database/sql
row := db.QueryRowContext(ctx, `SELECT id, name FROM users WHERE phone = $1`, phone)

// pgx
row := pool.QueryRow(ctx, `SELECT id, name FROM users WHERE phone = $1`, phone)
```

```go
var page = template.Must(template.New("hi").Parse(`<p>Hello, {{.Name}}</p>`))
// Name "<script>alert(1)</script>" renders as &lt;script&gt;alert(1)&lt;/script&gt;
```

Day 55 writes the queries. Day 60 tries to break them.

## A04: Insecure Design

**What it is.** The code does exactly what was designed, but the design itself is unsafe. No amount of careful coding fixes a missing rule.

**How it shows up in Go.** A booking endpoint that lets one request book 500 seats. Two requests that book the same seat at the same moment. A password reset that never expires. A payment that's charged twice when the phone retries.

**The fix.**

- [ ] Threat-model each feature before you build it: who could misuse this, and how? (Day 82)
- [ ] Enforce business limits on the server: seats per booking, bookings per day, reset links that expire in 15 minutes.
- [ ] Let the database settle races: a unique constraint on `(trip_id, seat_no)` and a transaction, not "check, then insert".
- [ ] Accept an `Idempotency-Key` on payments and bookings (see [http.md](http.md)).
- [ ] Write tests for abuse, not just for the happy path.

## A05: Security Misconfiguration

**What it is.** The code is fine but the settings around it aren't: missing headers, debug pages left on, defaults that are open.

**How it shows up in Go.** `http.ListenAndServe` with no timeouts. `import _ "net/http/pprof"`, which adds `/debug/pprof/` to `http.DefaultServeMux`, served to the internet. CORS that allows `*`. Stack traces in error responses.

**The fix.**

- [ ] Use your own `http.ServeMux`, not `http.DefaultServeMux`, and an `http.Server` with timeouts (see [http.md](http.md)).
- [ ] If you need pprof, serve it on a separate port bound to `localhost`.
- [ ] Send security headers on every response.
- [ ] Allow only your own front-end origins in CORS. Never `*` with credentials.
- [ ] Protect cookie-based endpoints from cross-site requests with `http.CrossOriginProtection` (Go 1.25 and later).
- [ ] Return generic error messages. Log the details.
- [ ] Set cookies with `Secure`, `HttpOnly` and `SameSite`.

```go
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

handler := http.NewCrossOriginProtection().Handler(securityHeaders(mux))
```

| Header | What it stops |
|---|---|
| `Content-Security-Policy` | scripts from other sites; `frame-ancestors 'none'` stops your page being framed (clickjacking) |
| `Strict-Transport-Security` | the browser ever using plain HTTP for your site again |
| `X-Content-Type-Options: nosniff` | the browser guessing a file is a script when you said it wasn't |
| `Referrer-Policy` | your URLs leaking to other sites in the `Referer` header |

Day 61 builds CORS and the headers.

## A06: Vulnerable and Outdated Components

**What it is.** A library, the Go toolchain, or a base image you depend on has a known security hole.

**How it shows up in Go.** A `go.mod` that pins an old `golang.org/x/net`, or a server built with a Go version that is no longer supported.

**The fix.**

- [ ] Run `govulncheck` in CI. It reports only the vulnerabilities your code actually calls, so the list is short and real.
- [ ] Use a supported Go release. Go supports the two newest major versions.
- [ ] Turn on Dependabot or Renovate to open pull requests for updates.
- [ ] Use a small, current base image in Docker, and rebuild it often (Day 63).
- [ ] Remove modules you no longer use: `go mod tidy`.

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## A07: Identification and Authentication Failures

**What it is.** Logins that can be guessed, sessions that never end, or tokens that can be stolen and reused.

**How it shows up in Go.** A `/login` handler with no rate limit. A JWT with no expiry, or one signed with a short secret. A session ID that stays the same after login.

**The fix.**

- [ ] Rate-limit login, sign-up and reset per IP and per account, and answer `429` with `Retry-After`.
- [ ] Give the same answer for "wrong password" and "no such user".
- [ ] Short-lived access tokens (minutes), and refresh tokens that rotate on every use (Day 58).
- [ ] Issue a new session ID on login and delete it on logout.
- [ ] Check the JWT's algorithm, expiry and issuer. Reject `alg: none`.
- [ ] Set a minimum password length and check against known breached passwords.

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Every(time.Minute/5), 5) // about 5 tries a minute
if !limiter.Allow() {
	w.Header().Set("Retry-After", "60")
	writeError(w, http.StatusTooManyRequests, "too many attempts, try again in a minute")
	return
}
```

Keep one limiter per key (an IP or a phone number) in a map guarded by a mutex, and evict old ones. Day 47 builds a token bucket by hand.

## A08: Software and Data Integrity Failures

**What it is.** Code or data is trusted without checking where it came from: an unpinned dependency, a tampered build, a forged webhook.

**How it shows up in Go.** A CI workflow that runs a third-party action from a moving tag. An M-Pesa callback handler that believes any POST that arrives. Decoding untrusted `encoding/gob` data.

**The fix.**

- [ ] Commit `go.sum`. Go checks every module against it and against the public checksum database (`sum.golang.org`).
- [ ] Pin third-party GitHub Actions to a full commit SHA, and give workflows `permissions: contents: read` unless they need more.
- [ ] Verify webhook signatures with HMAC, compared with `hmac.Equal`, and reject old timestamps to stop replays (Day 95).
- [ ] For callbacks that have no signature, confirm the result by calling the provider's API yourself.
- [ ] Only decode formats from untrusted sources that can't run code: JSON, not `gob`.

## A09: Security Logging and Monitoring Failures

**What it is.** Attacks happen and nobody notices, because nothing was logged, or the logs are full of secrets that become a second leak.

**How it shows up in Go.** A failed login that logs nothing. A request logger that prints the `Authorization` header. `slog.Info("config", "cfg", cfg)` with the database password inside.

**The fix.**

- [ ] Log security events with `log/slog`: logins (good and bad), access denied, rate-limit hits, password changes, admin actions.
- [ ] Include who, what, when, the request ID and the IP. Not the password, token, full card number or session cookie.
- [ ] Redact secrets at the type level, so they can't be logged by accident.
- [ ] Alert on spikes: many `401`s from one IP, many `500`s after a deploy (Day 81).

```go
type Secret string

func (Secret) LogValue() slog.Value { return slog.StringValue("[redacted]") }

// slog.Info("config", "jwtSecret", cfg.JWTSecret) now prints jwtSecret=[redacted]
```

`slog.LogValuer` works for `slog` only. `fmt.Println(secret)` still prints the real value, so also add a `String() string` method if the type might be printed.

Day 54 adds structured logging and request IDs.

## A10: Server-Side Request Forgery (SSRF)

**What it is.** Your server fetches a URL the user chose, and the user points it at something only your server can reach: `localhost`, the database, or the cloud metadata address `169.254.169.254`.

**How it shows up in Go.** `http.Get(r.FormValue("logoURL"))`. A webhook URL a user can set to anything.

**The fix.**

- [ ] Allow only the hosts you need, and only `https`.
- [ ] Block private, loopback and link-local IPs **at dial time**, so a DNS name that resolves to `127.0.0.1` is caught too.
- [ ] Check redirects as well, and limit how many you follow.
- [ ] Set a timeout and a maximum response size.

```go
var allowedHosts = map[string]bool{
	"api.safaricom.co.ke":     true,
	"sandbox.safaricom.co.ke": true,
}

func checkURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" || !allowedHosts[u.Hostname()] {
		return nil, errBlocked
	}
	return u, nil
}

var safeClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
			Control: func(network, address string, _ syscall.RawConn) error {
				ap, err := netip.ParseAddrPort(address)
				if err != nil {
					return err
				}
				ip := ap.Addr().Unmap()
				if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
					ip.IsUnspecified() || ip.IsMulticast() {
					return errBlocked
				}
				return nil
			},
		}).DialContext,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("too many redirects")
		}
		_, err := checkURL(req.URL.String())
		return err
	},
}
```

`Control` runs after DNS lookup, with the real IP about to be dialled. That's why it catches tricks a check on the hostname alone would miss.

Day 60 attacks and fixes this.

## After the Top 10: OWASP ASVS

The Top 10 is an awareness list: it tells you what goes wrong most often. When you want a standard to build and test against, use the [OWASP Application Security Verification Standard (ASVS)](https://owasp.org/www-project-application-security-verification-standard/).

- It's a list of specific, testable requirements, grouped by area: authentication, sessions, access control, validation, cryptography, logging and more.
- It has three levels. **Level 1** is the minimum for any app. **Level 2** is for apps that handle sensitive data, which includes payments, so it's the right target for the booking API. **Level 3** is for the most critical systems.
- Use it as a checklist in code review and threat modelling (Day 82), and to agree with a client what "secure" means before you start.

The OWASP [Cheat Sheet Series](https://cheatsheetseries.owasp.org/) has short, practical guides for most ASVS areas.

## Where the course uses this

| Day | What you meet |
|---|---|
| 57 | A02: bcrypt and argon2id, never storing plain text |
| 58 | A07: sessions, JWTs, expiry, refresh rotation |
| 59 | A01: roles, `401` vs `403`, checks in one place |
| 60 | the whole Top 10: injection, broken access control, SSRF |
| 61 | A05: CORS, CSP, HSTS and the other headers |
| 63 | A06 and A08: `govulncheck` and pinned actions in CI |
| 82 | A04: threat modelling with STRIDE |
