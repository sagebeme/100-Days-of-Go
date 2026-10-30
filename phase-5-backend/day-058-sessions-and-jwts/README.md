# Day 58: Sessions and JWTs

## The brief

Yesterday Twende (a made-up matatu SACCO app) learned to check a password. Now a passenger who has logged in needs to *stay* logged in without sending the password with every booking. Twende has two front ends, and each gets the tool that suits it:

- **The website** gets a **cookie session**: a random ID in a cookie, and the server remembers which passenger it belongs to.
- **The phone app** gets **tokens**: a JWT access token that lasts 15 minutes, and a refresh token that lasts 30 days and is swapped for a new one every time it's used. If an old refresh token is ever used twice, someone has stolen a copy, and every token from that login dies.

```
NewSessionCookie("twende_session", id, 2*time.Hour).String()
  →  "twende_session=K7Q...; Path=/; Max-Age=7200; HttpOnly; Secure; SameSite=Lax"

tokens.NewAccessToken("0712345678", "passenger")  →  "eyJhbGciOiJIUzI1NiIs...", nil
tokens.Parse(thatToken)                           →  &Claims{Role: "passenger", Subject: "0712345678", ...}, nil
tokens.Parse(sameToken, 16 minutes later)         →  nil, ErrInvalidToken
tokens.Parse(tokenWithAlgNone)                    →  nil, ErrInvalidToken

r1 := refresh.Issue("0712345678")
r2, user, _ := refresh.Rotate(r1)    →  new token, "0712345678", nil
refresh.Rotate(r1)                   →  ErrRefreshReused   (and r2 is dead too)

GET /me                               →  401  WWW-Authenticate: Bearer
GET /me  Authorization: Bearer junk   →  401  WWW-Authenticate: Bearer error="invalid_token"
GET /me  Authorization: Bearer <ok>   →  200
```

## What you'll use

**Cookie sessions.** The server makes a random ID with `crypto/rand.Text()` (Go 1.24 and later), stores `id → passenger` (in memory, Postgres or Redis), and sends the ID in a cookie. The browser sends it back on every request. Logging out deletes the row, which ends the session at once. Three cookie flags do most of the security work:

| Flag | What it does |
|---|---|
| `HttpOnly` | JavaScript can't read the cookie, so an XSS bug can't steal the session |
| `Secure` | sent only over HTTPS. Browsers treat `http://localhost` as secure, so it still works on your laptop |
| `SameSite=Lax` | not sent with cross-site `POST`s, which blocks most CSRF. `Strict` also drops it when someone follows a link to your site |

**JWTs** (JSON Web Tokens) are `header.payload.signature`, each part base64url. The payload holds **claims**: `sub` (who), `iss` (who issued it), `aud` (who it's for), `iat` (issued at), `nbf` (not before), `exp` (expires) and `jti` (a unique ID). The server signs it with a secret, so it can check a token without a database lookup. That's the strength and the weakness: a JWT can't be cancelled before `exp`, so keep access tokens short.

```go
claims := Claims{Role: "passenger", RegisteredClaims: jwt.RegisteredClaims{
	Subject:   "0712345678",
	ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
}}
s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)

_, err = jwt.ParseWithClaims(s, &claims, func(*jwt.Token) (any, error) { return secret, nil },
	jwt.WithValidMethods([]string{"HS256"}), // pin the algorithm
	jwt.WithExpirationRequired(),
)
```

**Pin the algorithm.** The token's header says which algorithm signed it, and the attacker writes the header. Without `WithValidMethods`, a token with `"alg": "none"` or a different algorithm might be accepted. Always say which one you expect.

**Refresh tokens with rotation.** The refresh token is a random string, not a JWT, stored on the server, so it *can* be revoked. Each use returns a new one and marks the old one used. All tokens from one login form a **family**. If a used token comes back, either the app retried or a thief has a copy; you can't tell which, so revoke the family and make the passenger log in again. Store the SHA-256 of each token, not the token: a leaked table is then useless. A fast hash is fine here because the token is 26 random characters, not a guessable password.

**`errors.Join`, `os.LookupEnv` and `t.Setenv`** for config, as on [Day 21](../../phase-2-modelling/day-021-config-loader/). The secret comes from `JWT_SECRET` and has no default. See [env-and-config.md](../../reference/env-and-config.md).

### Cookies or tokens?

| | Cookie session | JWT access + refresh |
|---|---|---|
| Best for | a website on your own domain | phone apps, other services calling your API |
| Logout | instant: delete the row | the access token works until `exp`; revoke the refresh family |
| Stolen by XSS | not with `HttpOnly` | yes, if the browser app keeps it in `localStorage` |
| CSRF | yes: use `SameSite` and `http.CrossOriginProtection` ([Day 61](../day-061-cors-and-security-headers/)) | no, if sent in the `Authorization` header |
| Server lookup per request | yes | no, until the refresh |

## Why it matters at work

Authentication bugs are OWASP A07 ([owasp-checklist.md](../../reference/owasp-checklist.md)), and reviewers look for the same handful: a JWT library called without a pinned algorithm, a token with no `exp` or one that lasts a year, a default secret in the code, a cookie without `HttpOnly` and `Secure`, a refresh token that never changes. Production incidents follow the same pattern. A token leaks into a log or a screenshot, and if it lives for months, so does the attacker's access. Short access tokens, rotating refresh tokens and reuse detection turn a leak into a 15-minute problem. The `WWW-Authenticate` header matters too: the phone app reads `error="invalid_token"` to know it should refresh once and replay the request (see [http.md](../../reference/http.md)).

## Steps

1. Open `starter/auth.go`. The errors, `Claims`, `Tokens`, `RefreshStore`, the env helpers and the JSON helpers are given.
2. Write `LoadConfig`. Read `JWT_SECRET` and require at least 32 bytes. Use `stringEnv` and `durationEnv` for the rest, with the defaults in the doc comment, and check that `ACCESS_TOKEN_TTL` is between 1 minute and 1 hour. Collect every problem and return `errors.Join(errs...)`. Never put the secret in an error message.
3. Write `NewSessionID` with `rand.Text()`, and `NewSessionCookie` with `Path: "/"`, `MaxAge` in whole seconds, and the three flags.
4. Write `NewAccessToken`. Fill `Claims` with the role and the registered claims: `sub`, `iss`, `aud` (a `jwt.ClaimStrings`), `iat` and `nbf` at `t.Now()`, `exp` at `t.Now().Add(t.TTL)` and a random `jti`. Sign with `jwt.SigningMethodHS256`.
5. Write `Parse` with `jwt.ParseWithClaims` and these options: `WithValidMethods`, `WithIssuer`, `WithAudience`, `WithExpirationRequired`, `WithIssuedAt`, `WithLeeway(Leeway)` and `WithTimeFunc(t.Now)`. Also reject an empty subject. Wrap every failure: `fmt.Errorf("%w: %v", ErrInvalidToken, err)`.
6. Write `Issue`: lock the mutex, then call the given `add` with a new random family ID.
7. Write `Rotate`, holding the mutex the whole time. Look the token up by `hashToken`. Missing, revoked or expired (`!now.Before(expires)`) is `ErrInvalidRefresh`. Already used: mark every record in the family revoked and return `ErrRefreshReused`. Otherwise mark it used and `add` a new token in the same family.
8. Write `RequireAuth`. Take the token with `strings.CutPrefix(header, "Bearer ")`. No token: 401 with `WWW-Authenticate: Bearer`. A bad token: 401 with `Bearer error="invalid_token"`. A good one: put the claims in the context with `context.WithValue(r.Context(), ctxKey{}, claims)` and call `next`.
9. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-058-sessions-and-jwts/starter
go test -race ./phase-5-backend/day-058-sessions-and-jwts/starter
```

The second run checks that `Rotate` is safe when 20 requests use the same refresh token at once. Exactly one may win.

## The kata: Token Timestamps

Decode the payload of each token in a log and decide whether it was valid when it arrived: malformed, wrong algorithm, issued in the future, not yet valid, expired, or fine. It's what `jwt.ParseWithClaims` does for you, done by hand. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 58
```

## When you're stuck

- **`key is of invalid type`** from `SignedString` or `Parse`: HS256 needs the secret as a `[]byte`, not a `string`. Return `t.Secret` from the key func.
- **`token has invalid claims: token is expired`** in a test that shouldn't expire: you used `time.Now()` instead of `t.Now()`. Use the clock in the struct for both making and checking tokens.
- **The "issued in the future" case is accepted**: add `jwt.WithIssuedAt()`. Without it the library doesn't check `iat`.
- **The "no exp" case is accepted**: add `jwt.WithExpirationRequired()`. A missing `exp` is otherwise allowed.
- **`TestLoadConfig` passes alone but fails after `TestLoadConfigErrors`**: you read env variables at package level (in a `var`). Read them inside `LoadConfig`, so `t.Setenv` takes effect.
- **`20 goroutines rotated the same token and 3 succeeded`**: you unlocked between the check and the update. Hold the lock from the lookup until the new token is saved.
- **`WWW-Authenticate` is empty on the 401**: set headers before `writeError`, which calls `WriteHeader`. Headers set afterwards are ignored.
- **Still stuck?** Open `solution/auth.go`, read it once, close it, then write your own version from memory.
