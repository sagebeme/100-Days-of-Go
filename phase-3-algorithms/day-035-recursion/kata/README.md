# Kata: Rumour Mill

In a WhatsApp group in Nakuru, a rumour starts with one person. Every round, each person who knows it forwards it to `a` new people. After `b` rounds, `a^b` people have just heard it. The numbers get silly fast, so print them modulo 1 000 000 007.

**Input**

- Line 1: `q`, the number of questions (1 ≤ q ≤ 10^5)
- The next `q` lines: two integers `a` and `b` (0 ≤ a ≤ 10^9, 0 ≤ b ≤ 10^18)

**Output**

For each question, one line: `a^b mod 1 000 000 007`. Take `0^0` to be 1.

**Example**

```
Input           Output
3               1024
2 10            1
3 0             125
5 3
```

**Limits:** 2 seconds per test case.

**Read the constraints.** `b` can be 10^18. A loop that multiplies `b` times would run for centuries. Fast power is O(log b), about 60 steps: `a^b = (a^(b/2))²`, times one more `a` when `b` is odd, and `a^0 = 1`.

**Watch out for overflow.** Reduce modulo 1 000 000 007 after every multiplication. Two numbers below 10^9 + 7 multiply to under 1.1 × 10^18, which fits in Go's 64-bit `int` (up to about 9.2 × 10^18). Reduce `a` first too: it can be 10^9, but so can the modulus.

**Go tip**

```go
const mod = 1_000_000_007
func power(a, b int) int // recursive: power(a, b/2), then square it
```

Print with a `bufio.Writer`: there can be 10^5 lines of output.

Judge it from the repo root with `go run ./tools/kata 35`.
