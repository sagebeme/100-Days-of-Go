# Kata: Request Line Parser

The first line of every HTTP/1.1 request is the request line: a method, a target and a version, with one space between each.

```
GET /trips?route=125&date=2026-11-02 HTTP/1.1
```

Go's server parses this for you. Today you write the parser yourself. For each line, print what a strict server would make of it.

Check the rules in this order and stop at the first one that fails:

1. Split the line on single spaces. There must be exactly 3 parts and none may be empty. Otherwise print `400`.
2. The method must be exactly one of `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`. Otherwise print `400`.
3. The target must start with `/`. Otherwise print `400`.
4. The version must be `HTTP/1.0` or `HTTP/1.1`. If it starts with `HTTP/` but is some other version, print `505` (HTTP Version Not Supported). If it doesn't start with `HTTP/` at all, print `400`.

A line that passes every rule prints the method, the path and the number of query parameters, separated by spaces. The path is the target up to the first `?`. The query is everything after that `?`. Split the query on `&` and count the pieces that aren't empty. A target with no `?` has 0 parameters.

**Input**

- Line 1: `n`, the number of request lines (1 ≤ n ≤ 5 000)
- The next `n` lines: one request line each, at most 200 characters. A line can start or end with a space, and it never contains a tab.

**Output**

`n` lines, one answer per request line.

**Example**

```
Input                                            Output
4                                                GET /trips 2
GET /trips?route=125&date=2026-11-02 HTTP/1.1    POST /bookings 0
POST /bookings HTTP/1.1                          400
get /trips HTTP/1.1                              505
GET /trips HTTP/2.0
```

`get` is lower case, and methods are case-sensitive, so line 3 is `400`.

**Limits:** 2 seconds per test case.

**Go tip**

Read whole lines with a `bufio.Scanner`, because `fmt.Fscan` would skip the spaces you need to check. `strings.Split` keeps empty pieces, which is what you want here; `strings.Fields` would hide a double space. `strings.Cut` splits on the first `?` only:

```go
sc := bufio.NewScanner(os.Stdin)
sc.Scan() // the line with n

parts := strings.Split("GET  /trips HTTP/1.1", " ") // ["GET" "" "/trips" "HTTP/1.1"]
path, query, found := strings.Cut("/trips?route=125", "?") // "/trips", "route=125", true
```

Judge it from the repo root with `go run ./tools/kata 49`.
