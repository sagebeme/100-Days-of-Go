# Kata: Lost Updates

`total += a` looks like one step, but the computer does it in three: **load** `total` into a register, **add** to the register, **store** the register back into `total`. When two goroutines interleave those steps, one store can overwrite the other, and an update is lost. That's the race you fixed today.

Simulate it. A shared counter `x` starts at 0. Goroutine `i` wants to add 1 to `x`, `c_i` times. Each of its increments is three steps in order: load (`r_i = x`), add (`r_i = r_i + 1`) and store (`x = r_i`). Every goroutine has its own register `r_i`. A schedule says which goroutine takes its next step, one step at a time.

Print the final value of `x`, and how many increments were lost.

**Input**

- Line 1: `k`, the number of goroutines (1 ≤ k ≤ 100 000)
- Line 2: `k` numbers `c_1 ... c_k`, the increments each goroutine does (1 ≤ c_i, and they add up to at most 100 000)
- Line 3: `L`, the length of the schedule, which is 3 × (c_1 + ... + c_k)
- Line 4: `L` goroutine numbers from 1 to `k`. Goroutine `i` appears exactly 3 × `c_i` times.

**Output**

- Line 1: the final value of `x`
- Line 2: the number of lost increments: (c_1 + ... + c_k) minus `x`

**Example**

```
Input           Output
2               1
1 1             1
6
1 2 1 1 2 2
```

Goroutine 1 loads 0. Goroutine 2 loads 0. Goroutine 1 adds (its register is 1) and stores: `x` is 1. Goroutine 2 adds (its register is 1, from the 0 it loaded) and stores: `x` is still 1. Two increments ran and one was lost.

**Limits:** 2 seconds per test case.

**Go tip**

Keep three slices indexed by goroutine: its register, and which step it does next (0 load, 1 add, 2 store). Each schedule entry does that goroutine's next step, then moves its step on with `step[g] = (step[g] + 1) % 3`.

Judge it from the repo root with `go run ./tools/kata 45`.
