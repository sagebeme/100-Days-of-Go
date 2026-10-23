# Kata: Top of the Class

Print the class in order of merit: highest score first. Students with the same score go in alphabetical order of their names.

**Input**

- Line 1: `n`, the number of students (1 ≤ n ≤ 10 000)
- The next `n` lines: a name and a score, separated by a space. A name is 1 to 12 lowercase letters `a`–`z` with no spaces. Two students can share a name. A score is a whole number from 0 to 100.

**Output**

`n` lines, each a name and a score separated by a space, in order of merit.

**Example**

```
Input           Output
4               baraka 85
amina 70        amina 70
baraka 85       chebet 70
chebet 70       kip 38
kip 38
```

**Limits:** 2 seconds per test case.

**Go tip**

Structs arrive on Day 15. Until then, keep two slices side by side, `names` and `scores`, and sort a slice of positions instead of the data itself:

```go
order := make([]int, n) // 0, 1, 2, ...
for i := range order {
	order[i] = i
}
sort.Slice(order, func(a, b int) bool {
	x, y := order[a], order[b]
	if scores[x] != scores[y] {
		return scores[x] > scores[y]
	}
	return names[x] < names[y]
})
```

Then print `names[order[i]]` and `scores[order[i]]` for each `i`.

Judge it from the repo root with `go run ./tools/kata 10`.
