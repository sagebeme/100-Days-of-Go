# Day 10: Exam Grader

## The brief

Mr. Kamau teaches Form 3 chemistry in Nyeri. After every CAT he works out the class mean, the median, each student's grade and the order of merit by hand, with a calculator and a red pen. Write the four functions that do it for him.

```
Mean([]int{70, 80, 90})        →  80
Median([]int{90, 40, 70})      →  70
Median([]int{10, 40, 30, 20})  →  25
Grade(72)                      →  "B", nil
Grade(101)                     →  "", score 101 is outside 0 to 100
Rank(map[string]int{"Amina": 70, "Baraka": 85, "Chebet": 70})
                               →  ["Baraka", "Amina", "Chebet"]
```

Amina and Chebet tie on 70, so they go in alphabetical order. The grade bands are: A for 80 to 100, B for 65 to 79, C for 50 to 64, D for 40 to 49, E below 40.

## What you'll use

- `sort.Ints(s)` sorts a slice of ints in place, smallest first.
- `sort.Slice(s, less)` sorts any slice. You give it a function `less(i, j int) bool` that says whether `s[i]` should come before `s[j]`.
- `copy(dst, src)` copies one slice into another, so you can sort the copy and leave the original alone.
- `float64(total) / float64(n)` gives a real average. `total / n` with ints throws away the fraction.
- Map keys come out in a random order. To get them in a fixed order, put them in a slice and sort it.

```go
names := []string{"Kip", "Amina", "Baraka"}
sort.Slice(names, func(i, j int) bool {
	return names[i] < names[j]
})
// names is now ["Amina" "Baraka" "Kip"]

sorted := make([]int, len(scores))
copy(sorted, scores)
sort.Ints(sorted) // scores itself is unchanged
```

## Steps

1. Open `starter/grader.go`. Import `"fmt"` and `"sort"` as you need them.
2. `Mean`: if there are no scores, return 0. Otherwise add them up and divide, converting both numbers to `float64` first.
3. `Median`: return 0 for no scores. Make a copy with `make` and `copy`, then sort the copy with `sort.Ints`. With an odd count, return the middle one, `sorted[n/2]`. With an even count, return the average of `sorted[n/2-1]` and `sorted[n/2]`.
4. `Grade`: return an error from `fmt.Errorf` when the score is below 0 or above 100. Then use a `switch` with no value, checking from the top band down: `score >= 80` is `"A"`, `score >= 65` is `"B"`, and so on.
5. `Rank`: put every name from the map into a slice. Sort it with `sort.Slice`. In the `less` function, a higher score comes first. If the two scores are equal, the name that comes first in the alphabet goes first.
6. Run the tests from the repo root:

```bash
go test ./phase-1-terminal/day-010-exam-grader/starter
```

## The kata: Top of the Class

Read a class list of names and scores and print it as an order of merit: highest score first, ties in alphabetical order. A custom sort with two keys is one of the most common steps in a HackerRank problem. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 10
```

## When you're stuck

- **`Mean([50 51]) = 50, want 50.5`**: you divided two ints. Convert both to `float64` before dividing.
- **`Median changed its input`**: `sort.Ints` sorts the slice you give it in place. Sort a copy.
- **`index out of range [-1]`**: you used `sorted[n/2-1]` on an empty slice. Return 0 before you index.
- **`Grade(80)` gives `"B"`**: check the order of your cases and whether you used `>` or `>=`. A switch runs the first case that matches.
- **`Rank` gives a different order each run**: you returned the names straight from the map. Sort them before returning.
- **`invalid operation: scores[a] > scores[b] (mismatched types ...)`**: inside `less`, `i` and `j` are positions in the names slice. Look up the name first, `names[i]`, then its score.
- **Still stuck?** Open `solution/grader.go`, read it once, close it, then write your own version from memory.
