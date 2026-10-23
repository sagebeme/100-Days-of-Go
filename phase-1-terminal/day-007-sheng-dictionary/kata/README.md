# Kata: Trending Word

Your group chat never stops. Read every word sent today and print the one used the most, with how many times it came up. If two or more words are tied, print the one that comes first in alphabetical order.

**Input**

- Line 1: `n`, the number of words (1 ≤ n ≤ 200 000)
- The next lines: `n` words separated by spaces or newlines. Each word is 1 to 20 lowercase letters `a` to `z`

**Output**

One line: the word, a space, and its count.

**Example**

```
Input                              Output
7                                  poa 3
sasa poa niaje poa msee sasa poa
```

**Limits:** 2 seconds per test case.

**Go tip**

Count first, decide later. `fmt.Fscan` reads one word at a time into a `string`. After counting, `range` over the map and keep the best word. Strings compare alphabetically with `<`, which settles a tie:

```go
counts := make(map[string]int)
counts[word]++

for w, c := range counts {
	if c > bestCount || (c == bestCount && w < best) {
		best, bestCount = w, c
	}
}
```

A map's `range` order is random on purpose, so your tie rule must not depend on the order you visit the words.

Judge it from the repo root with `go run ./tools/kata 7`.
