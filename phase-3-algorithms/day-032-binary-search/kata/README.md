# Kata: Boda Deliveries

A boda boda rider in Eldoret has `n` parcels to deliver in a fixed order. Each day he loads the next parcels in the list onto his bike until the next one would take him over the bike's capacity, then rides. He never skips a parcel or changes the order. He must finish within `d` days.

What is the smallest capacity, in kg, that lets him finish in `d` days or fewer?

**Input**

- Line 1: `n` and `d` (1 ≤ d ≤ n ≤ 10^5)
- Line 2: `n` parcel weights in kg, separated by spaces (1 ≤ weight ≤ 10^4)

**Output**

One line: the smallest capacity that works.

**Example**

```
Input           Output
6 3             6
3 2 2 4 1 4
```

With capacity 6 the days are `3 2` | `2 4` | `1 4`. With 5 he needs four days: `3 2` | `2` | `4 1` | `4`.

**Limits:** 2 seconds per test case.

**Read the constraints.** The answer is somewhere between the heaviest parcel and the total, which can be 10^9. Trying every capacity from the heaviest up, with an O(n) check each time, can be 10^8 checks of 10^5 parcels. But if a capacity works, every bigger one works too. That makes the answer searchable: binary search over capacities, about 30 checks of O(n) each.

**Go tip**

Write the check first, then the search:

```go
func daysNeeded(weights []int, capacity int) int
```

Judge it from the repo root with `go run ./tools/kata 32`.
