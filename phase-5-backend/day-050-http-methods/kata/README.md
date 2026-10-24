# Kata: Allow Header

Build the part of a router that decides between `200`, `404` and `405`.

You get a list of route patterns, each a method and a path. A path segment written as `{name}` is a wildcard: it matches any one segment. Every other segment must match exactly. A pattern matches a path only if both have the same number of segments. `/` has no segments.

For each request, print:

- `200` if a pattern with the request's method matches the path. A `GET` pattern also accepts `HEAD`.
- otherwise, if any pattern matches the path, `405` followed by a space and the allowed methods: every method of every matching pattern, plus `HEAD` when `GET` is among them, sorted alphabetically and joined with `, `.
- otherwise `404`.

**Input**

- Line 1: `n`, the number of patterns (1 ≤ n ≤ 100)
- The next `n` lines: a method and a path, separated by a space. No pattern appears twice.
- Then a line with `q`, the number of requests (1 ≤ q ≤ 3 000)
- The next `q` lines: a method and a path

Methods are upper-case words. Paths start with `/`, have at most 5 segments and never end with `/` (except `/` itself).

**Output**

One line per request.

**Example**

```
Input                  Output
4                      200
GET /trips             405 GET, HEAD, POST
POST /trips            200
GET /trips/{id}        405 DELETE, GET, HEAD
DELETE /trips/{id}     404
5
GET /trips
DELETE /trips
HEAD /trips/42
PUT /trips/42
GET /bookings
```

`/trips/42` matches `GET /trips/{id}` and `DELETE /trips/{id}`, so `PUT` gets `405` and the list `DELETE, GET, HEAD`. Nothing matches `/bookings`.

**Limits:** 2 seconds per test case.

**Go tip**

Split paths into segments with `strings.Split(strings.Trim(path, "/"), "/")`, and treat the empty string from `/` as zero segments. Collect the allowed methods in a `map[string]bool` so `GET` added twice still appears once, then sort the keys:

```go
allowed := map[string]bool{}
allowed["GET"], allowed["HEAD"] = true, true
names := make([]string, 0, len(allowed))
for m := range allowed {
	names = append(names, m)
}
sort.Strings(names)
strings.Join(names, ", ") // "GET, HEAD"
```

Judge it from the repo root with `go run ./tools/kata 50`.
