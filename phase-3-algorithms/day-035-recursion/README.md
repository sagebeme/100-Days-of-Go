# Day 35: Recursion

## The brief

A recursive function solves a problem by calling itself on a smaller version of the same problem, until the problem is so small the answer is obvious. That obvious case is the **base case**, and forgetting it is the most common recursion bug.

You'll start small with factorials, then use recursion to try every possibility: every way to arrange the letters of a matatu's name sticker, every way to pay with the notes in your pocket, and the Tower of Hanoi, the puzzle that shows how a short recursive function can do 2^n − 1 moves.

```
Factorial(5)                                →  120
Permutations("aab")                         →  [aab aba baa]
CanPay([]int{1000, 500, 200, 50}, 750)      →  true     500 + 200 + 50
CanPay([]int{1000, 500, 200, 50}, 300)      →  false
Hanoi(2, "A", "B", "C")                     →  [{A B} {A C} {B C}]
```

## What you'll use

### Base case and recursive case

```go
func Factorial(n int) int {
	if n <= 1 {
		return 1 // base case: stops the calls
	}
	return n * Factorial(n-1) // recursive case: a smaller n each time
}
```

Each call waits on the stack ([Day 33](../day-033-stacks/)) until the smaller one returns. Go's stack grows as needed, so thousands of levels are fine. A million levels is slow, and an explicit stack is better there ([reference/big-o.md](../../reference/big-o.md#before-you-submit)).

### Include or exclude

For "can some of these add up to X", every note is either used or not. That's two recursive calls, each on the rest of the notes:

```go
return CanPay(notes[1:], amount-notes[0]) || CanPay(notes[1:], amount)
```

With n notes that's up to 2^n paths. Fine for n = 25, hopeless for n = 100. [Day 39](../day-039-dynamic-programming/) makes this kind of problem fast when the amounts are small.

### Backtracking

To build every permutation, choose a letter, recurse to fill the rest, then **undo** the choice and try the next letter. Keep a `used` slice and a `path` that grows and shrinks. To skip duplicates from repeated letters, sort the letters first, and skip a letter equal to the previous one if the previous one isn't in use.

A **closure** that calls itself has to be declared before it's assigned: `var build func()` then `build = func() { ... build() ... }`.

### Splitting into smaller copies

Hanoi with n discs is: move n−1 discs to the spare peg, move the biggest disc, move the n−1 discs back on top. Two calls on n−1 give 2^n − 1 moves. The kata does the same trick the clever way: `a^b` needs `a^(b/2)` only **once**, then squares it. One call on half the size is O(log b).

### Spotting it in a problem

- "**all** permutations / subsets / combinations", "every way to", "generate": backtracking.
- **n ≤ 10** (n! choices) or **n ≤ 20** (2^n choices). Small limits are the hint that brute-force recursion is the intended answer.
- Trees and nested structures ([Day 36](../day-036-trees/)): a tree is a node plus smaller trees.
- "divide in half": merge sort, binary search, fast power.
- If the same smaller problem gets solved again and again, it's dynamic programming (Day 39).

### Brute force vs optimal

| Problem | Brute force | Recursive idea |
|---|---|---|
| a^b mod m, b up to 10^18 | multiply b times: O(b) | fast power: O(log b) |
| permutations of n letters | there are n! of them, so O(n · n!) is the best possible | backtracking, skipping duplicates |
| subset sum, n ≤ 25 | O(2^n) | include/exclude, stopping when the amount goes negative |

Sometimes brute force is the answer. When the output itself has n! items, no algorithm is faster than listing them. Read the constraints to know when that's allowed ([reference/big-o.md](../../reference/big-o.md#what-n-allows): n ≤ 10 means O(n!) is OK).

## Steps

1. Open `starter/recursion.go`. Write `Factorial` recursively. Return 0 for a negative `n`, before the base case.
2. Write `CanPay`. Base cases first: amount 0 is payable, and with no notes left (or a negative amount) it isn't. Then the include-or-exclude line above.
3. Write `Hanoi`. For `n == 0` return `nil`. Otherwise: the moves for `n-1` discs from `from` to `via` (using `to` as the spare), then `Move{from, to}`, then the moves for `n-1` discs from `via` to `to`.
4. Write `Permutations` with backtracking: sort the letters with `slices.Sort`, then a recursive closure `build` that appends `string(path)` when `path` is full. Get it working on `"abc"` first, then add the duplicate check for `"aab"`.
5. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-035-recursion/starter
```

## The kata: Rumour Mill

`a^b mod 1 000 000 007` with `b` up to 10^18, for up to 10^5 questions. This is fast power, also called exponentiation by squaring: recursion that halves the problem instead of shrinking it by one. The modulo and the overflow check are what timed tests love to catch you on. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 35
```

## When you're stuck

- **`runtime: goroutine stack exceeds 1000000000-byte limit`** or **`fatal error: stack overflow`:** there's no base case, or the recursive call isn't getting smaller. Check that `n-1` (or `notes[1:]`) really is passed.
- **`Factorial(0)` returns 0:** your base case is `n == 1` only. Use `n <= 1`.
- **`Permutations("aab")` returns 6 answers:** add the duplicate check. Sort the letters first, or it won't work.
- **`Permutations` returns the same string 6 times:** you appended `path` itself instead of `string(path)`. A string copies the bytes; the slice keeps changing.
- **`CanPay(pocket, 2000)` returns true:** you passed `notes` instead of `notes[1:]` in the "use it" call, so the 1000 was used twice.
- **`Hanoi(3)` gives the right number of moves in the wrong order:** swap the `via` and `to` arguments in the first recursive call.
- **The kata prints negative numbers:** a product overflowed. Reduce modulo 10^9 + 7 after each multiplication, and reduce `a` before you start.
- **Still stuck?** Open `solution/recursion.go`, read it once, close it, then write your own version from memory.
