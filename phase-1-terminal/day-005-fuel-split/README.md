# Day 5: Fuel Split

## The brief

Four friends are driving from Nairobi to Naivasha for the weekend. Before they leave, someone has to work out what the fuel will cost and how much each person sends on M-Pesa. Whatever doesn't split evenly goes to the car's owner, who is driving anyway.

Some inputs make no sense: a car that does 0 km per litre, a negative distance, a trip split among nobody. Your functions must say so with an error instead of returning a silly number.

```
TripCost(100, 15, 180)  →  1200, nil
TripCost(50, 12, 177)   →  738, nil      (737.5 rounded up)
TripCost(10, 0, 180)    →  0, error

Split(738, 4)           →  184, 2, nil
Split(500, 0)           →  0, 0, error
```

*Fuel prices change every month. The numbers here are only for practice.*

## What you'll use

- A function can return more than one value. List the types in brackets: `func f() (int, error)`
- Go has no exceptions. A function that can fail returns an `error` as its last value, and `nil` means "no error"
- `errors.New("message")` makes a fixed error, and `fmt.Errorf("can't use %d", x)` makes one with values in it
- When you return an error, return zero for the other values
- You can name the return values, as in `func Split(cost, people int) (share, leftover int, err error)`. The names document what each value means. You still return them with a normal `return`
- The caller checks the error straight away with `if err != nil`

```go
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("can't divide by zero")
	}
	return a / b, nil
}

result, err := Divide(10, 2)
if err != nil {
	fmt.Println("problem:", err)
}
```

## Steps

1. Open `starter/fuel.go`. Keep the function names, parameters and return types as they are.
2. In `TripCost`, check the bad inputs first. If `kmPerLitre` is 0 or less, or `distanceKm` or `pricePerLitre` is negative, return `0` and an error. Use `fmt.Errorf` so the message shows the bad value, for example `fmt.Errorf("km per litre must be more than 0, got %d", kmPerLitre)`.
3. The cost is `distanceKm * pricePerLitre / kmPerLitre`, rounded up to the next shilling. Integer division rounds down, so add `kmPerLitre - 1` before you divide: `(distanceKm*pricePerLitre + kmPerLitre - 1) / kmPerLitre`. Try it on paper with 8850 and 12.
4. Return the cost and `nil`.
5. In `Split`, return `0, 0` and an error if `people` is less than 1 or `cost` is negative. `errors.New` is fine here. Add `"errors"` to your imports.
6. Otherwise return `cost / people`, `cost % people` and `nil`.
7. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-005-fuel-split/starter
```

## The kata: Fuliza Check

Read a list of M-Pesa balances, some of them negative because of Fuliza, and print the smallest and the largest. Finding the minimum and maximum in one pass is a classic warm-up question in coding tests. Write it with a helper function that returns two values. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 5
```

## When you're stuck

- **`not enough return values`** with **`have (int)`** and **`want (int, error)`**: every `return` must give both values. Write `return cost, nil`.
- **`assignment mismatch: 1 variable but TripCost returns 2 values`**: you wrote `cost := TripCost(...)`. Catch both: `cost, err := TripCost(...)`.
- **`undefined: errors`**: add `"errors"` to the import list, next to `"fmt"`.
- **`TripCost(10, 0, 180)` panics with `integer divide by zero`**: check `kmPerLitre` before you divide, and return early.
- **`TripCost(50, 12, 177) got 737, want 738`**: you rounded down. Add `kmPerLitre - 1` before dividing.
- **`Split(500, 0)` should return an error, got nil**: the check has to come before the division, and it has to `return`.
- **Still stuck?** Open `solution/fuel.go`, read it once, close it, then write your own version from memory.
