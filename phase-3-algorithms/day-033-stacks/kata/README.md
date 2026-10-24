# Kata: Longest Balanced Stretch

Wanjiru's old calculator has a broken screen and only shows the brackets she typed: a long string of `(` and `)`. She wants the longest stretch of consecutive characters that is a balanced bracket sequence, where every `(` is closed by a later `)` and nothing closes before it opens.

**Input**

- Line 1: `n` (1 ≤ n ≤ 2 × 10^5)
- Line 2: `n` characters, each `(` or `)`

**Output**

One line: the length of the longest balanced stretch, or `0` if there is none.

**Example**

```
Input           Output
6               4
)()())
```

The stretch `()()` from position 2 to 5 is balanced. Nothing longer is.

**Limits:** 2 seconds per test case.

**Read the constraints.** Starting at every position and scanning right is O(n²), about 2 × 10^10 steps on deep nesting. A stack of indexes does it in one pass. Push the index of each `(`. On `)`, pop. If the stack is now empty, this `)` can't be matched: push its index as a new floor. Otherwise the stretch since the index on top is balanced, and its length is `i - top`. Start the stack with `-1` as the first floor.

**Go tip**

A plain `[]int` works as a stack here: `stack = append(stack, i)` to push and `stack = stack[:len(stack)-1]` to pop. Or reuse today's `Stack[int]`.

Judge it from the repo root with `go run ./tools/kata 33`.
