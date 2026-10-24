# Kata: Compare Versions

Not every project uses strict semver. Plenty of tools number their releases `1.2`, `1.2.0.4` or `4.08`. Compare two such version numbers.

Split each version on `.`. Compare the parts from left to right **as whole numbers**, so `1.10` is newer than `1.9` and `1.01` equals `1.001` (leading zeros don't count). If one version has fewer parts, treat its missing parts as `0`, so `1.0` equals `1.0.0`.

**Input**

- Line 1: `q`, the number of pairs (1 ≤ q ≤ 100 000)
- The next `q` lines: two versions `a` and `b`, separated by a space. Each has 1 to 10 parts, and each part is 1 to 9 digits.

**Output**

For each pair, one line: `-1` if `a` is older than `b`, `1` if `a` is newer, `0` if they're the same.

**Example**

```
Input        Output
4            0
1.01 1.001   0
1.0 1.0.0    -1
0.1 1.1      1
1.10 1.9
```

**Limits:** 2 seconds per test case.

**Go tip**

`strings.Split(a, ".")` gives the parts and `strconv.Atoi` turns `"001"` into `1`. Loop up to the longer length and use 0 when an index is past the end of the shorter one:

```go
x := 0
if i < len(pa) {
	x, _ = strconv.Atoi(pa[i])
}
```

This is LeetCode's "Compare Version Numbers", a common warm-up question in interviews.

Judge it from the repo root with `go run ./tools/kata 20`.
