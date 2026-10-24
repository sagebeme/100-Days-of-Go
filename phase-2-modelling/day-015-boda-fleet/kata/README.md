# Kata: Rider Leaderboard

A boda boda SACCO pays a bonus to its best riders each week. You get every trip of the week, one per line. Add the trips up per rider and print the top `k`.

Rank riders by total earnings, highest first. If two riders earned the same, the one who rode more kilometres comes first. If that's equal too, sort by name in alphabetical order.

**Input**

- Line 1: `n` and `k`, the number of trips and how many riders to print (0 ≤ n ≤ 200 000, 1 ≤ k ≤ 1000)
- The next `n` lines: a rider's name, the trip's kilometres and its fare in shillings, separated by spaces. Names are 1 to 20 lowercase letters and digits. 1 ≤ km ≤ 100, 0 ≤ fare ≤ 10 000.

**Output**

The top `k` riders, one per line: name, total earnings, total kilometres and number of trips, separated by spaces. If there are fewer than `k` riders, print them all. If there are no trips, print nothing.

**Example**

```
Input              Output
5 2                otieno 550 16 2
otieno 4 150       achieng 350 12 2
achieng 10 300
otieno 12 400
mwangi 3 100
achieng 2 50
```

Otieno earned 150 + 400 = 550 over 16 km. Achieng earned 350 over 12 km. Mwangi's 100 doesn't make the top 2.

**Limits:** 2 seconds per test case.

**Go tip**

Keep one struct per rider in a map of pointers, so you can add to it in place, then copy the values into a slice to sort:

```go
type stats struct {
	name                string
	earnings, km, trips int
}
byName := map[string]*stats{}
s, ok := byName[name]
if !ok {
	s = &stats{name: name}
	byName[name] = s
}
s.earnings += fare
```

`sort.Slice(list, func(i, j int) bool { ... })` takes a "less" function. Compare earnings first, then km, then names.

Judge it from the repo root with `go run ./tools/kata 15`.
