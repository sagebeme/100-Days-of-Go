# Kata: Token Timestamps

The Twende gateway has a log of access tokens and the time each one arrived. Before anyone looks at signatures, it needs a quick sort: which tokens are even shaped right, and which were valid at the time on the clock? Check each one.

A token is three parts joined by dots: `header.payload.signature`. The header and payload are JSON objects, each encoded with **base64url without padding** (the alphabet with `-` and `_`, and no `=` at the end). Ignore the signature.

For each token, print the first verdict that applies:

1. `MALFORMED`: not exactly three parts, a part that isn't valid base64url, a header or payload that isn't a JSON object, an `exp`, `iat` or `nbf` that isn't a whole number, or no `exp` at all.
2. `BAD_ALG`: the header's `alg` isn't exactly `HS256` (including a missing `alg`).
3. `BAD_IAT`: `iat` is present and `iat > now + skew` (issued in the future).
4. `NOT_YET`: `nbf` is present and `now + skew < nbf`.
5. `EXPIRED`: `now >= exp + skew`.
6. `VALID` otherwise.

`skew` is the clock difference the gateway forgives, in seconds.

**Input**

- Line 1: `n skew` (1 ≤ n ≤ 100 000, 0 ≤ skew ≤ 300)
- Next `n` lines: `now token`, where `now` is a Unix time in seconds and `token` has no spaces

**Output**

`n` lines, one verdict each.

**Example**

```
Input
3 30
1000 eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIwNzEyMzQ1Njc4IiwiaWF0Ijo5MDAsImV4cCI6MTgwMH0.c2ln
1830 eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIwNzEyMzQ1Njc4IiwiaWF0Ijo5MDAsImV4cCI6MTgwMH0.c2ln
1000 not-a-token

Output
VALID
EXPIRED
MALFORMED
```

The payload decodes to `{"sub":"0712345678","iat":900,"exp":1800}`. At 1830, `1830 >= 1800 + 30`, so it has expired.

**Limits:** 2 seconds per test case.

**Go tip:** `base64.RawURLEncoding.DecodeString` is exactly the encoding JWTs use. Unmarshal into a struct with pointer fields, `Exp *int64 \`json:"exp"\``: a nil pointer means the claim was missing, and a value like `"200"` or `1.5` makes `json.Unmarshal` return an error. Read with `fmt.Fscan` on a `bufio.Reader`: it splits on spaces, so `now` and `token` come out separately.

Judge it from the repo root with `go run ./tools/kata 58`.
