# Day 3: Matatu Fare Rules

## The brief

Every Nairobi commuter knows the fare isn't fixed. It goes up with distance, it goes up in the rush hour, and the moment it starts raining the makanga shouts a new price. Write the fare rules as code, plus a little helper that tells you which route number goes where.

The fare rules for this project:

| Distance | Base fare |
|---|---|
| 0 km or less | Ksh0, and no extras: there's no trip |
| 1 to 5 km | Ksh50 |
| 6 to 15 km | Ksh80 |
| 16 to 30 km | Ksh120 |
| over 30 km | Ksh200 |

- Rush hour adds Ksh30. Rush hour is any trip starting at hour 6, 7 or 8 in the morning, or 17, 18 or 19 in the evening.
- Rain adds Ksh20.

```
Fare(3, 10, false)   →  50
Fare(12, 7, false)   →  110    (80 + 30 rush hour)
Fare(40, 18, true)   →  250    (200 + 30 + 20)
Route("Rongai")      →  "125"
Route("Mombasa")     →  "unknown"
```

*The fares are made up, and real makangas don't follow a table. The route numbers are real: 46K (Citi Hoppa, Kencom to Kawangware), 111 (Ngong), 125 (Rongai), 45 (Githurai) and 58 (Buruburu), checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026. Routes change, so confirm with the conductor.*

## What you'll use

- `if`, `else if` and `else` to choose between paths. The condition needs no brackets, but the braces are required
- `&&` (and), `||` (or) and the comparisons `<`, `<=`, `>`, `>=`, `==`, `!=`
- `switch` to pick one case out of many. It stops after the first matching case, so there's no `break`. A case can list several values, and a `switch` with no value after it works like a tidy chain of `if`s

```go
if hour >= 6 && hour <= 8 {
	fare = fare + 30
}

switch day {
case "Saturday", "Sunday":
	mood = "relaxed"
default:
	mood = "rushing"
}

switch {
case km <= 5:
	fare = 50
case km <= 15:
	fare = 80
}
```

## Steps

1. Open `starter/fare.go`. Keep the function names and parameters as they are.
2. In `Fare`, return 0 straight away if `distanceKm` is 0 or less.
3. Use a `switch` with no value to set the base fare from the distance table. Order matters: check the smallest band first.
4. Add 30 if the hour is in either rush hour. One `if` with `||` between the morning and evening checks is enough.
5. Add 20 if `raining` is true. You can write `if raining`, because it's already a `bool`.
6. In `Route`, `switch` on the destination. Kawangware is `"46K"`, Ngong is `"111"`, Rongai and Ongata Rongai are both `"125"`, Githurai is `"45"`, Buruburu is `"58"`, and anything else is `"unknown"`.
7. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-003-matatu-fare/starter
```

## The kata: Stage Takings

A makanga counts the money at the end of a trip. Children ride for Ksh30, adults for Ksh60 and seniors ride free. Read the passengers' ages and print the takings. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 3
```

## When you're stuck

- **`syntax error: unexpected newline, expected { after if clause`**: Go wants the opening `{` on the same line as the `if`, not on the next line.
- **`non-boolean condition in if statement`**: you wrote something like `if hour`. An `if` needs a true or false answer, so compare: `if hour == 7`.
- **`expected boolean expression, found assignment`**: you wrote `if hour = 7`. One `=` stores a value; two `==` compare.
- **`missing return`**: every path through the function has to end in a `return`. Put a `return fare` (or `return "unknown"`) at the bottom.
- **`Fare(3, 10, false)` got 80, want 50**: your cases are in the wrong order. In a `switch` the first true case wins, and 3 is also `<= 15`, so `distanceKm <= 5` must come first.
- **`Fare(5, 9, false) got 80, want 50`**: hour 9 is not rush hour. The morning rush is hours 6, 7 and 8.
- **`Fare(0, 7, true) got 50, want 0`**: return 0 before you add any extras.
- **Still stuck?** Open `solution/fare.go`, read it once, close it, then write your own version from memory.
