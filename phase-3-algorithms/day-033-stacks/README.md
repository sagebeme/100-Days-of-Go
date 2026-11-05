# Day 33: Stacks

## The brief

Plates at a kibanda: the last one washed goes on top, and the next one used comes off the top. That's a stack, last in, first out. It's the right tool whenever the most recent unfinished thing is the one you deal with next: the last open bracket, the last edit to undo, the last day still waiting for warmer weather.

You'll build a generic `Stack[T]`, then use it for a bracket checker, a text box with undo (HackerRank's "Simple Text Editor"), and the monotonic stack trick for Nairobi's next warmer day.

```
Balanced("fare(route[46K]{stage})")      →  true
Balanced("([)]")                         →  false
e.Type("Niko"); e.Type(" stage"); e.Delete(5); e.Undo()
e.Text()                                 →  "Niko stage"
NextWarmer([]int{22, 24, 21, 20, 25})    →  [1 3 2 1 0]
```

## What you'll use

### A generic stack

A slice with `append` to push and a re-slice to pop. Each is O(1). The type parameter from [Day 18](../../phase-2-modelling/day-018-typed-inventory/) makes one stack for every element type:

```go
type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
```

The methods have pointer receivers ([Day 15](../../phase-2-modelling/day-015-boda-fleet/)) because they change the stack. `Pop` on an empty stack returns `T`'s zero value and `false`: `var zero T` gives you one for any `T`.

### Brackets

Push every opening bracket. On a closing bracket, pop and check it matches. At the end the stack must be empty. `([)]` fails because `)` meets `[` on top.

### Undo

Before each change, push the old text. `Undo` pops and restores it. Storing whole copies is simple and fine for a test. A real editor stores only what changed (the typed text, or the deleted bytes) to save memory.

### The monotonic stack

For "next warmer day", keep a stack of days still waiting, coldest on top. When a warmer day `i` arrives, pop every day it beats and set its wait to `i - j`, then push `i`. The temperatures in the stack only go down, hence "monotonic". Each day is pushed once and popped at most once, so the inner loop costs O(n) in total, even though it's a loop inside a loop.

### Spotting it in a problem

- "**valid** / **balanced** brackets", "matching tags", "evaluate an expression".
- "**undo**", "back button", "the most recent".
- "**next greater** / next smaller element", "days until", "stock span", "largest rectangle in a histogram": a monotonic stack.
- "longest valid substring of brackets": a stack of indexes (the kata).
- Recursion ([Day 35](../day-035-recursion/)) runs on a stack too. If recursion gets too deep, an explicit stack replaces it.

### Brute force vs optimal

| Problem | Brute force | Stack |
|---|---|---|
| balanced brackets | delete `()` pairs again and again: O(n²) | O(n) |
| next warmer day | scan ahead from each day: O(n²) | monotonic stack: O(n) |
| longest balanced stretch (kata) | scan from each start: O(n²) | stack of indexes: O(n) |

With 200 000 days of falling temperatures, the scan-ahead version does 2 × 10^10 steps, way past [reference/big-o.md](../../reference/big-o.md#what-n-allows)'s budget. One of the tests does exactly that.

## Steps

1. Open `starter/stacks.go`. Write `Push`, `Peek`, `Pop` and `Len`. Write `Pop` by calling `Peek`, then shortening `s.items` if there was a value.
2. Write `Balanced` with a `Stack[byte]`. A map from each closing bracket to its opening one keeps the check to one line. Return `false` if a closer finds the stack empty, and check the stack is empty at the end.
3. Write `Type`, `Delete`, `Undo` and `Text`. `Type` and `Delete` push `e.text` onto `e.history` before changing it. `Delete` must not cut more than the text holds: `k = min(k, len(e.text))`.
4. Write `NextWarmer` with a `Stack[int]` of indexes and the monotonic loop above. Make the result with `make([]int, len(temps))` so days that never warm up stay 0.
5. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-033-stacks/starter
```

## The kata: Longest Balanced Stretch

The longest balanced run of brackets in a broken calculator's display. LeetCode files it under Hard ("Longest Valid Parentheses"), yet it is one pass with a stack of indexes. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 33
```

## Check yourself

1. Predict `Balanced("((")`, `Balanced(")(")` and `Balanced("ab")`.
2. Predict `NextWarmer([]int{30, 25, 20})` and explain why the stack's temperatures only go down.
3. `NextWarmer` has a loop inside a loop. Why is it still O(n)?
4. Predict `e.Text()` at the end, and why `Delete` uses `min(k, len(e.text))`:

   ```go
   var e Editor
   e.Type("abc")
   e.Delete(10)
   e.Undo()
   e.Undo()
   e.Undo()
   ```

<details>
<summary>Answers</summary>

1. `false` (two openers are still on the stack at the end), `false` (the `)` finds the stack empty), and `true` (no brackets at all is balanced).
2. `[0 0 0]`: it only gets colder, so no day ever has a warmer one ahead. A warmer day pops every colder day on top before it's pushed, so what's left below is never colder than what's on top.
3. Each index is pushed once and popped at most once. The inner loop's pops add up to at most n over the whole run.
4. `""`. The first undo brings back `"abc"`, the second the empty starting text, and the third has nothing to undo. Without `min`, deleting 10 bytes from a 3-byte string would slice out of range and panic.

</details>

## When you're stuck

- **`index out of range [-1]`:** you popped an empty slice. Check `len(s.items) == 0` first.
- **`cannot use zero (variable of type int) as T value`:** declare the zero value as `var zero T`, not `0` or `""`.
- **`Balanced("(")` is true:** you didn't check the stack is empty at the end.
- **`Balanced(")")` panics:** a closer arrived with nothing open. `Pop` returns `ok = false`: return `false`.
- **`Undo` brings back the wrong text:** push the text before you change it, not after.
- **`slice bounds out of range [:-95]`** in `Delete`: `k` was bigger than the text. Clamp it with `min`.
- **`NextWarmer` is slow on the big test:** you're scanning ahead from each day. Pop from the stack instead.
- **Still stuck?** Open `solution/stacks.go`, read it once, close it, then write your own version from memory.
