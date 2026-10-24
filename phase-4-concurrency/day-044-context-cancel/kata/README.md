# Kata: Cancel Tree

A booking server makes contexts as it works. Context `i` is made from a parent context `p_i`, which was made before it, or from `context.Background()` when `p_i` is 0. Cancelling a context cancels it and everything made from it, all the way down. A context that's already cancelled stays cancelled and isn't counted again.

For each cancel, print how many contexts it newly cancelled.

**Input**

- Line 1: `n`, the number of contexts (1 ≤ n ≤ 200 000)
- Line 2: `n` numbers `p_1 ... p_n` (0 ≤ p_i < i)
- Line 3: `m`, the number of cancels (1 ≤ m ≤ 200 000)
- The next `m` lines: a context `x` to cancel (1 ≤ x ≤ n)

**Output**

`m` lines: how many contexts each cancel newly cancelled.

**Example**

```
Input           Output
6               2
0 1 1 2 0 5     0
4               2
2               1
4
1
6
```

Context 1 has children 2 and 3, context 2 has child 4, and context 5 has child 6. Cancelling 2 cancels 2 and 4. Cancelling 4 changes nothing: it already is. Cancelling 1 newly cancels 1 and 3; 2 and 4 already were. Cancelling 6 cancels just 6.

**Limits:** 2 seconds per test case. At the full limits, walking the whole subtree on every cancel can mean 200 000 × 200 000 steps, so skip what is already cancelled.

**Go tip**

Build a list of children for each context, `children := make([][]int, n+1)`. When a walk reaches a context that's already cancelled, stop there: everything under it is cancelled too. Then each context is visited at most once in the whole run. A tree can be one long chain 200 000 deep; Go's stacks grow, so recursion works, but a loop with your own stack (a slice) works just as well.

Judge it from the repo root with `go run ./tools/kata 44`.
