# Day 4: Times Tables Drill

## The brief

Mwalimu Achieng wants her Class 4 pupils to chant their times tables every morning, and she wants a printout for each table. She also sets bonus questions: what do all the answers in the 7 times table add up to? How many numbers between 1 and 100 can be shared equally among 3 friends? Write three functions that do the counting for her.

```
Table(3, 3)
  →  "3 x 1 = 3\n3 x 2 = 6\n3 x 3 = 9\n"

SumTable(7, 12)              →  546
CountDivisible(1, 100, 3)    →  33
```

## What you'll use

- `for` is Go's only loop. The classic form has three parts: start, condition, step
- An accumulator is a variable you create before the loop and add to inside it. It works for numbers (`total += x`) and for strings (`text += line`)
- `i++` adds one to `i`, and `+=` adds to a variable in place
- A loop whose condition is false from the start runs zero times. That's often exactly the edge case you want

```go
total := 0
for i := 1; i <= 5; i++ {
	total += i
}
// total is 15

text := ""
for i := 1; i <= 3; i++ {
	text += fmt.Sprintf("%d,", i)
}
// text is "1,2,3,"
```

## Steps

1. Open `starter/tables.go`. Keep the function names and parameters as they are.
2. `Table`: start with `text := ""`. Loop `i` from 1 to `upTo`. Each time round, add one line made with `fmt.Sprintf("%d x %d = %d\n", n, i, n*i)`. Return `text`. Add `import "fmt"`.
3. `SumTable`: the same loop, but add `n*i` to a `total` that starts at 0.
4. `CountDivisible`: loop `x` from `from` to `to`, and add one to a `count` each time `x % d == 0`.
5. Check the edge cases in your head. What happens when `upTo` is 0, or `from` is bigger than `to`? If your loop runs zero times you return the starting value, which is the right answer.
6. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-004-times-tables/starter
```

## The kata: Dry Spell

A farmer in Machakos records the rainfall every day. Find the longest run of days in a row with no rain at all. This is your first "streak" problem: keep a running count, reset it when the streak breaks, and remember the best you've seen. It comes up in interviews all the time. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 4
```

## When you're stuck

- **`expected '{', found '++'`**: you separated the three parts of your `for` with commas. Use semicolons: `for i := 1; i <= upTo; i++ {`.
- **`invalid operation: text += n * i (mismatched types string and int)`**: you added a number straight onto a string. Build the line with `fmt.Sprintf` first, then add that.
- **`Table(3, 3)` is missing the last line**: your condition is `i < upTo`. You want `i <= upTo` so the last row is included.
- **`Table(3, 3)` lines run together**: put `\n` at the end of the `Sprintf` template, so every line ends with a newline, including the last.
- **`CountDivisible(0, 10, 5) got 2, want 3`**: 0 counts. `0 % 5` is 0, so 0 is divisible by 5. Start the loop at `from`, not at 1.
- **Still stuck?** Open `solution/tables.go`, read it once, close it, then write your own version from memory.
