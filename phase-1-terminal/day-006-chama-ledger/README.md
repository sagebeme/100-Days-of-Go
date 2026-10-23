# Day 6: Chama Ledger

## The brief

Mama Njeri's chama meets on the first Sunday of every month. Each member is meant to bring Ksh1,000. The treasurer writes every contribution in a counter book, and every meeting starts with the same three questions: how much have we collected, how did the balance grow, and who still owes? Write the treasurer's helpers.

```
Total([]int{1000, 1000, 500})                  →  2500
RunningBalance([]int{500, 1000, 250})           →  [500 1500 1750]
Defaulters([]string{"Njeri", "Atieno", "Wafula"},
           []int{1000, 600, 1000}, 1000)       →  [Atieno]
```

## What you'll use

- A slice is a list of values of one type: `[]int`, `[]string`
- `len(s)` is how many items it holds, and `s[i]` is the item at position `i`, counting from 0
- `append(s, x)` returns a new slice with `x` on the end. Always keep the result: `s = append(s, x)`
- `for i, x := range s` walks a slice and gives you each position and value. Write `_` for the one you don't need
- A slice declared with `var s []int` is `nil`. It has length 0, and you can still `append` to it

```go
var paid []int
paid = append(paid, 1000)
paid = append(paid, 500)
// paid is [1000 500], len(paid) is 2

for i, amount := range paid {
	fmt.Println(i, amount) // 0 1000, then 1 500
}
```

## Steps

1. Open `starter/ledger.go`. Keep the function names and parameters as they are.
2. `Total`: start a `total` at 0, `range` over the contributions and add each one.
3. `RunningBalance`: make an empty result with `var balances []int`. Keep a running `sum`. For each contribution, add it to `sum` and `append` the new `sum` to `balances`. Don't change the slice you were given.
4. `Defaulters`: `range` over `names` with the index, so you can read `paid[i]` for the same member. If they paid less than `due`, append their name to the result.
5. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-006-chama-ledger/starter
```

## The kata: Merry-Go-Round

In a merry-go-round chama the payout rotates: each month the person at the front gets paid and moves to the back. After `k` months, who is in what order? Rotating a list is a classic interview question, and the trick is not to do `k` moves one at a time. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 6
```

## When you're stuck

- **`append(balances, sum) (value of type []int) is not used`**: `append` doesn't change the slice in place. Keep its result: `balances = append(balances, sum)`.
- **`panic: runtime error: index out of range [3] with length 3`**: you read past the end of a slice. The last position is `len(s)-1`, so loop with `i < len(s)` or use `range`.
- **`declared and not used: i`**: `range` gave you an index you don't use. Write `for _, x := range s`.
- **`RunningBalance` gives `[500 1000 250]`**: you appended each contribution, not the running sum. Add to `sum` first, then append `sum`.
- **`RunningBalance changed its input`**: you wrote into the slice you were given, like `contributions[i] += ...`. Build a new slice with `append` instead.
- **Still stuck?** Open `solution/ledger.go`, read it once, close it, then write your own version from memory.
