# Kata: Cross Out k Digits

A shopkeeper in Gikomba has a price tag with a long number on it. A customer haggles, and the deal is this: the shopkeeper crosses out exactly `k` digits, the rest close up in the same order, and the customer pays what's left. The customer gets to choose which digits. Print the smallest number they can leave.

**Input**

- Line 1: `n` and `k` (1 ≤ n ≤ 2 × 10^5, 0 ≤ k ≤ n)
- Line 2: `n` digits with no spaces. The first digit may be `0`.

**Output**

One line: the smallest number left after crossing out exactly `k` digits, with leading zeros removed. If nothing is left, or only zeros, print `0`.

**Example**

```
Input           Output
7 3             1219
1432219
```

Cross out the 4, the 3 and one of the 2s. The digits left are `1219`.

More cases: `10200` with `k = 1` gives `200` (the `1` goes, and the leading `0` is dropped). `10` with `k = 2` gives `0`.

**Limits:** 2 seconds per test case.

**Read the constraints.** Trying every set of `k` digits is hopeless. Crossing out one digit at a time, rescanning from the start each time, is O(n·k): about 10^10 steps. The greedy is O(n): the leftmost digit that is bigger than the digit after it is the one to cross out, because an earlier digit is worth more. A stack does all `k` removals in one pass: before pushing a digit, pop bigger digits off the top while you still have crossings left. If any are left at the end, cross out the last digits, which are now the biggest.

**Go tip**

Keep the stack as a `[]byte`, and build the answer with `string(kept)`. `strings.TrimLeft(s, "0")` drops leading zeros. Print with a `bufio.Writer`: the answer can be 2 × 10^5 digits long, far too big for an `int`.

Judge it from the repo root with `go run ./tools/kata 38`.
