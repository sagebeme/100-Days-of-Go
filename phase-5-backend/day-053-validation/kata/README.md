# Kata: IPv4 Check

Rate limits, logs and allow-lists all work on client IP addresses, and some of those addresses come from headers anyone can type into. Validate them before you trust them.

A valid IPv4 address here is exactly four parts separated by `.`, where each part:

- is 1 to 3 characters, every one an ASCII digit `0` to `9` (no signs, spaces or other characters)
- has no leading zero, unless the part is exactly `0`
- is at most 255

Print `VALID` or `INVALID` for each line.

**Input**

- Line 1: `n`, the number of lines (1 ≤ n ≤ 10 000)
- The next `n` lines: one candidate each, up to 30 bytes. A line can be empty, and can contain spaces or any other characters.

**Output**

`n` lines, each `VALID` or `INVALID`.

**Example**

```
Input           Output
5               VALID
192.168.0.1     VALID
41.90.64.7      INVALID
256.1.1.1       INVALID
10.0.0          INVALID
10.01.0.1
```

`256` is too big, `10.0.0` has three parts, and `01` has a leading zero.

**Limits:** 2 seconds per test case.

**Go tip**

`strings.Split("1..2.3", ".")` gives `["1" "" "2" "3"]`, so empty parts show up and you can reject them. Check the characters yourself before converting: `strconv.Atoi` accepts `+4` and `-4`, which aren't valid here. The standard library has `net/netip.ParseAddr`, which is what you'd use at work, but write the rules yourself today. Read with a `bufio.Scanner` so empty lines and spaces survive.

Judge it from the repo root with `go run ./tools/kata 53`.
