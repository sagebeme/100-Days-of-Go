# Kata: Word Count

Count how many times each word appears. Print every distinct word with its count, the most common first. Words with the same count go in alphabetical order.

**Input**

- Line 1: `n`, the number of words (1 ≤ n ≤ 30 000)
- Then `n` words separated by spaces or newlines. Each word is 1 to 15 lowercase letters `a`–`z`.

**Output**

One line per distinct word: the word, a space, and its count.

**Example**

```
Input                         Output
7                             sasa 3
sasa msee sasa niaje          msee 2
msee sasa poa                 niaje 1
                              poa 1
```

**Limits:** 2 seconds per test case.

**Go tip**

Count into a `map[string]int`, then copy the keys into a slice and sort that. Maps never come out in a fixed order, so you have to sort:

```go
counts := map[string]int{}
counts[word]++ // a missing key starts at 0

words := make([]string, 0, len(counts))
for w := range counts {
	words = append(words, w)
}
sort.Slice(words, func(i, j int) bool { ... })
```

`fmt.Fscan(in, &word)` reads the next word whether it is after a space or a newline.

Judge it from the repo root with `go run ./tools/kata 13`.
