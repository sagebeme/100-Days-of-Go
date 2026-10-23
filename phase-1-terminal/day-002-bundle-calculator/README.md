# Day 2: Bundle Calculator

## The brief

You have Ksh250 in airtime and a data bundle costs Ksh99. How many bundles can you buy, and how much is left over? Is a big bundle really cheaper per gigabyte than a small one? Write four small functions that answer those questions.

```
BundlesYouCanBuy(250, 99)  →  2
Change(250, 99)            →  52
GBToMB(1.5)                →  1536
PricePerGB(50, 350)        →  146.29
```

*The prices are made up. Real bundles change every few months, so check the menu before you trust any number here.*

## What you'll use

- `int` for whole numbers and `float64` for numbers with a decimal point
- `/` on two ints gives a whole number and throws away the rest: `7 / 2` is `3`
- `%` gives the remainder of a division: `7 % 2` is `1`
- Go never mixes `int` and `float64` for you. You convert by hand with `float64(x)` or `int(x)`, and `int(x)` cuts off the decimals without rounding
- `math.Round` rounds a `float64` to the nearest whole number

```go
shillings := 7
half := float64(shillings) / 2 // 3.5
whole := int(3.9)              // 3: the .9 is cut off, not rounded
cents := math.Round(2.567*100) / 100 // 2.57
```

## Steps

1. Open `starter/bundle.go`. Keep the function names and parameters exactly as they are.
2. `BundlesYouCanBuy`: return `budget / price`. Both are ints, so the answer is already a whole number of bundles.
3. `Change`: return what's left after those bundles. That's the remainder, `budget % price`.
4. `GBToMB`: one gigabyte is 1024 megabytes. Multiply `gb` by 1024, then convert the result to `int`. The conversion drops any part of a megabyte.
5. `PricePerGB`: convert `mb` to gigabytes with `float64(mb) / 1024`, then divide `float64(price)` by that. Round to 2 decimal places with `math.Round(x*100) / 100`, and add `import "math"` under the `package` line.
6. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-002-bundle-calculator/starter
```

## The kata: Data Diary

You log how many megabytes you use each day. Print the total and the average per day to 2 decimal places. It's the same int and float64 conversion you used in the project, inside a loop. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 2
```

## When you're stuck

- **`invalid operation: gb * mbPerGB (mismatched types float64 and int)`**: you multiplied a `float64` by an `int` variable. Convert the variable with `float64(mbPerGB)`, or write the number straight in: `gb * 1024`.
- **`cannot use gb * 1024 (value of type float64) as int value in return statement`**: wrap it in `int(...)` before you return it.
- **`undefined: math`**: add `import "math"` on the line after `package bundlecalculator`.
- **`PricePerGB(50, 350) got +Inf`**: you divided two ints before converting. `float64(mb / 1024)` does the int division first, gets 0, and then you divide by zero. Convert first: `float64(mb) / 1024`.
- **`PricePerGB(50, 350) got 146.28571428571428`**: you forgot to round. Use `math.Round(x*100) / 100`.
- **Still stuck?** Open `solution/bundle.go`, read it once, close it, then write your own version from memory.
