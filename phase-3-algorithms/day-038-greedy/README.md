# Day 38: Greedy

## The brief

A greedy algorithm makes the choice that looks best right now and never takes it back. Paying Ksh1780 at a supermarket in Nanyuki, you hand over the biggest note that fits, then the next: 1000, 500, 200, 50, 20, 10. Six pieces, no planning. Greedy algorithms are short and fast, usually a sort and one loop. The hard part is knowing when looking best now really is best overall.

Today: making change with the fewest notes, booking the most events into a community hall, and getting a crowd home on the fewest bodas.

```
FewestNotes(1780)                                         →  map[10:1 20:1 50:1 200:1 500:1 1000:1]
MaxBookings([]Slot{{9, 12}, {10, 11}, {11, 13}, {12, 14}}) →  2      10–11, then 12–14
FewestBodas([]int{70, 50, 80, 50}, 130)                   →  2      80+50 and 70+50
```

## What you'll use

### Choose, commit, repeat

Most greedy solutions are: sort by the right key, then walk once and take each item if it still fits. The sort is O(n log n) and the walk O(n).

### Proving it: the exchange argument

Before you trust a greedy, try to prove it. The usual way: take any best answer, and show you can swap its first choice for the greedy choice without making it worse. Then repeat for the rest.

For hall bookings: let the greedy take the slot that **ends first**. In any best schedule, swap its first slot for that one. It ends no later, so everything after it still fits. Same number of bookings, so the greedy is never worse.

The rules that seem just as sensible fail. The tests include both:

| Rule | Counter-example | Greedy gets | Best |
|---|---|---|---|
| starts first | 1–10, 2–3, 4–5, 6–7 | 1 (takes 1–10) | 3 |
| shortest first | 1–5, 4–7, 6–10 | 1 (takes 4–7) | 2 |
| **ends first** | | always the best | |

### When greedy is wrong

Biggest-note-first is optimal for Kenya's notes and coins (1000, 500, 200, 100, 50, 20, 10, 5, 1). It's not optimal for every set. With coins of 1, 3 and 4 and an amount of 6, greedy pays 4 + 1 + 1 (three coins), but 3 + 3 is two. When a counter-example exists, you need dynamic programming (Day 39), which tries every option without repeating work. If you can't prove a greedy and can't break it on a few small cases, write a brute force and compare them on random inputs.

### Pairing from both ends

For the bodas: sort the passengers. The heaviest person must go on some boda. If the lightest person fits with them, put them together: nobody else fits better with the heaviest. Otherwise the heaviest rides alone. That's two pointers (Day 29) on top of a sort.

### Spotting it in a problem

- "**maximum** number of non-overlapping intervals / meetings / events": sort by end.
- "**minimum** number of boats / trucks / rooms", "assign", "pair up": sort, then pair from both ends, or use a heap.
- "make change", "fewest coins" with standard denominations. Non-standard coins: DP.
- "smallest / largest number by removing or rearranging digits": a greedy, often with a stack (the kata).
- If a small counter-example breaks your rule, it's not greedy. Stop and think DP.

### Brute force vs optimal

| Problem | Brute force | Greedy |
|---|---|---|
| most bookings | try every subset: O(2^n) | sort by end, one pass: O(n log n) |
| fewest bodas | try every pairing: worse than O(n!) | sort, two pointers: O(n log n) |
| fewest notes | try every combination | biggest first: O(number of denominations) |
| cross out k digits (kata) | remove one at a time: O(n·k) | stack: O(n) |

The "pick the best each time" row of [reference/big-o.md](../../reference/big-o.md#patterns-and-their-complexity) is this whole day.

## Steps

1. Open `starter/greedy.go`. Write `FewestNotes`: for each denomination, largest first, if `amount >= d`, record `amount / d` and keep `amount % d`. Start from `map[int]int{}` so an amount of 0 gives an empty map.
2. Write `MaxBookings`. Clone the slots (the caller's slice must not change), sort the clone by `End` with `slices.SortFunc` and `cmp.Compare`. Keep `free`, the hour the hall is next free, starting very low (`math.MinInt` works). Take each slot with `Start >= free` and move `free` to its `End`.
3. Write `FewestBodas`. Sort a copy of the weights with `slices.Sorted(slices.Values(weights))`. With `i` at the lightest and `j` at the heaviest: if `i < j` and they fit together, move `i` on. Always move `j` down and count a trip. Loop while `i <= j`.
4. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-038-greedy/starter
```

## The kata: Cross Out k Digits

Cross out `k` digits from a price tag to leave the smallest number. LeetCode's "Remove K Digits": a greedy you can prove in one sentence, run with a monotonic stack (Day 33). Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 38
```

## When you're stuck

- **`FewestNotes(0)` returns `nil`:** make the map before the loop.
- **`FewestNotes` lists denominations with a count of 0:** only record `d` when `amount >= d`.
- **`MaxBookings` changed its input:** you sorted `slots` itself. Sort a clone.
- **`MaxBookings` returns 2 for back-to-back slots 9–10, 10–11, 11–12:** you used `Start > free`. A slot may start at the hour the last one ends: use `>=`.
- **`FewestBodas([]int{30, 30, 30}, 100)` returns 1:** a boda carries at most two people. Each trip moves `j` down once and `i` up at most once.
- **`FewestBodas([]int{60}, 100)` returns 0:** loop while `i <= j`, so the last person on their own still gets a trip.
- **The kata prints `0200`:** trim the leading zeros, and print `0` if nothing is left.
- **Still stuck?** Open `solution/greedy.go`, read it once, close it, then write your own version from memory.
