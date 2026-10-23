# Day 7: Sheng Dictionary

## The brief

Your cousin has just arrived from the UK, and every conversation in Nairobi leaves her lost. "Msee, hii keja ni fiti" means nothing to her yet. Build her a pocket Sheng dictionary: look a word up, translate a list of words, count which words come up most, and list the ones the dictionary doesn't know yet.

```
dict := map[string]string{"msee": "guy", "keja": "house", "fiti": "fine", "doh": "money"}

Lookup(dict, "keja")                                →  "house", true
Lookup(dict, "mbogi")                               →  "", false
TranslateAll(dict, []string{"msee", "hii", "keja"}) →  [guy hii house]
WordCounts([]string{"poa", "msee", "poa"})          →  map[msee:1 poa:2]
Unknown(dict, []string{"msee", "rada", "rada"})     →  [rada]
```

*Sheng changes fast and differs from estate to estate. If your mtaa says it differently, add your own words.*

## What you'll use

- A map stores values under keys: `map[string]string` maps words to meanings, `map[string]int` maps words to counts
- Make an empty map with `make(map[string]int)` or `map[string]int{}`. You can read from a `nil` map, but writing to one panics
- `m[key] = value` stores a value, and `m[key]` reads it. A missing key gives the zero value: `""` for strings, `0` for ints
- The comma-ok idiom tells you whether the key was really there: `value, ok := m[key]`
- `m[key]++` adds one to a count, even the first time, because a missing key starts at 0

```go
counts := make(map[string]int)
counts["poa"]++
counts["poa"]++
// counts["poa"] is 2

meaning, ok := dict["mbogi"]
if !ok {
	fmt.Println("not in the dictionary yet")
}
```

## Steps

1. Open `starter/dictionary.go`. Keep the function names and parameters as they are.
2. `Lookup`: one line. `meaning, ok := dict[word]`, then return both.
3. `TranslateAll`: `range` over the words. If a word is in the dictionary, append its meaning to the result. If it isn't, append the word unchanged.
4. `WordCounts`: make a map with `make(map[string]int)`, add one for every word, and return it. Return the empty map, not `nil`, when there are no words.
5. `Unknown`: you need two things, the result slice and a "seen" set so you don't list a word twice. A `map[string]bool` makes a good set. Append a word only if it's not in `dict` and you haven't seen it yet, then mark it as seen.
6. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-007-sheng-dictionary/starter
```

## The kata: Trending Word

Read a stream of words from a group chat and print the word that came up the most. Counting things with a map is one of the most common moves in coding tests, so this pattern will come back again and again. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 7
```

## When you're stuck

- **`panic: assignment to entry in nil map`**: you wrote to a map you only declared with `var counts map[string]int`. Make it first: `counts := make(map[string]int)`.
- **`assignment mismatch: 2 variables but 1 value`**: you wrote `meaning, ok := Lookup(...)` against a function that returns one value, or `a, b := x` by mistake. The comma-ok form only works on a map read like `dict[word]`.
- **`Lookup(dict, "rada")` gives `"", true`**: you returned `true` yourself instead of the `ok` the map gave you.
- **`TranslateAll` drops words**: when a word isn't in the dictionary, append the word itself, so the sentence keeps its length.
- **`Unknown` lists `rada` twice**: check your seen set before you append, and set `seen[word] = true` after.
- **`WordCounts([]) returned nil, want an empty map you made with make`**: make the map at the top of the function and return it, even when no words went in.
- **Still stuck?** Open `solution/dictionary.go`, read it once, close it, then write your own version from memory.
