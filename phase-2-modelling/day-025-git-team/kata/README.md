# Kata: Next Version

Conventional Commits pay off at release time: a tool can read the commit headers since the last release and choose the next version by itself. Write that tool.

A header is **valid** if it is: a type of one or more lowercase letters `a` to `z`; optionally a scope of lowercase letters, digits and hyphens in brackets, like `(day-24)`; optionally `!`; then `: ` (a colon and one space) and a description that doesn't start with a space. Ignore invalid headers completely.

Among the valid headers:

- any header with `!` means a **major** release: `v1.4.2` becomes `v2.0.0`
- otherwise, any `feat` means a **minor** release: `v1.4.2` becomes `v1.5.0`
- otherwise, any `fix` or `perf` means a **patch** release: `v1.4.2` becomes `v1.4.3`
- otherwise there's **none**, and the version stays the same

**Input**

- Line 1: the current version, `vMAJOR.MINOR.PATCH`, each number from 0 to 10^6
- Line 2: `n`, the number of commits (0 ≤ n ≤ 100 000)
- The next `n` lines: one commit header each, up to 100 characters. Headers may contain spaces.

**Output**

- Line 1: the next version
- Line 2: `major`, `minor`, `patch` or `none`

**Example**

```
Input                                   Output
v1.4.2                                  v1.5.0
4                                       minor
fix(ledger): stop negative balances
docs: explain .env.example
feat: add CSV export
chore: bump pgx
```

**Limits:** 2 seconds per test case.

**Go tip**

Headers have spaces, so read whole lines with a `bufio.Scanner`, not `fmt.Fscan`. Check the shape by hand with `strings.Cut(header, ": ")` and a loop over the bytes of the type and scope, or try the `regexp` package:

```go
var header = regexp.MustCompile(`^[a-z]+(\([a-z0-9-]+\))?(!)?: [^ ]`)
m := header.FindStringSubmatch(line) // nil if it doesn't match; m[2] is "!" or ""
```

This is how tools like semantic-release and release-please decide a version.

Judge it from the repo root with `go run ./tools/kata 25`.
