# Kata: Derby Draws

Mashemeji Derby night. A fan wrote down every goal of a long friendly tournament between Gor Mahia and AFC Leopards as one string: `G` for a Gor goal and `A` for an AFC Leopards goal, in order. A stretch is any run of consecutive goals in the string. A stretch is a draw if both teams scored the same number of goals in it.

Count the stretches that are draws.

**Input**

- Line 1: `n`, the number of goals (1 ≤ n ≤ 2 × 10^5)
- Line 2: `n` characters, each `G` or `A`

**Output**

One line: the number of stretches (i, j), i ≤ j, where `G` and `A` appear equally often from position i to position j.

**Example**

```
Input           Output
4               4
GAGA
```

The draws are `GA` (positions 1–2), `AG` (2–3), `GA` (3–4) and `GAGA` (1–4).

**Limits:** 2 seconds per test case.

**Read the constraints.** There are about n²/2 = 2 × 10^10 stretches, so you can't look at each one. Keep a running score `diff` (+1 for `G`, −1 for `A`). A stretch is a draw exactly when `diff` at its end equals `diff` just before its start. So the question becomes: how many earlier positions had the same `diff`? A map answers that in O(1). The total is O(n), and the answer can pass 2^32, so use `int`.

**Go tip**

```go
seen := map[int]int{0: 1} // the score before the first goal
```

Judge it from the repo root with `go run ./tools/kata 28`.
