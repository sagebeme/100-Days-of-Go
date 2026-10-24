# Kata: TinyURL

The classic interview warm-up for a URL shortener. Long URLs arrive one at a time. The first time a URL is seen it gets the next number from a counter, and its short code is that number in base62. A URL seen before gets its old code back. Then answer lookups: which long URL does a code stand for?

Base62 uses the digits `0-9`, then `A-Z`, then `a-z`, so `0` is 0, `A` is 10, `a` is 36 and `10` is 62.

**Input**

- Line 1: `n` and `start` (1 ≤ n ≤ 100 000, 0 ≤ start ≤ 10^12). The first new URL gets the number `start`, the next new one `start + 1`, and so on.
- The next `n` lines: one URL each, 1 to 100 characters with no spaces.
- The next line: `q` (1 ≤ q ≤ 100 000).
- The next `q` lines: a code, 1 to 12 characters with no spaces. It may contain characters that aren't base62.

**Output**

- `n` lines: the code for each URL, in input order.
- `q` lines: the URL for each code, or `404` if no URL has that code. Codes are exact: `00` is not the same code as `0`.

**Example**

```
Input                         Output
5 125000                      WW8
https://duka.co.ke/sufuria    WW9
https://kiungo.ke/about       WW8
https://duka.co.ke/sufuria    WWA
https://mtoto.school/fees     WW9
https://kiungo.ke/about       https://duka.co.ke/sufuria
4                             https://kiungo.ke/about
WW8                           https://mtoto.school/fees
WW9                           404
WWA
WWB
```

125000 is `32·62² + 32·62 + 8`, and 32 is `W`, so the first URL is `WW8`. The third line is a repeat and gets `WW8` again. `WWB` was never handed out.

**Limits:** 2 seconds per test case.

**Go tip**

Two maps do it: `map[string]string` from URL to code and another from code to URL. Build the code with a byte slice, filling it from the right:

```go
buf := make([]byte, 0, 11)
for n > 0 {
	buf = append(buf, alphabet[n%62])
	n /= 62
}
slices.Reverse(buf)
```

Judge it from the repo root with `go run ./tools/kata 85`.
