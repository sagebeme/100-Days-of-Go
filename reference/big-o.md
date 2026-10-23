# Big-O cheat sheet for HackerRank in Go

A reference for Days 27 to 40. Read the constraints first, pick an algorithm that fits them, then make sure your input and output aren't the slow part.

## What n allows

Rule of thumb: a judge runs about **10^8 simple operations per second** (an add, a compare, a slice index). Go is compiled, so it gets close to this. Most HackerRank problems give 1 to 2 seconds.

Work out your algorithm's cost with the largest n in the constraints, and compare it with 10^8.

| Largest n | Fastest algorithm you need | Examples |
|---|---|---|
| n ≤ 10 | O(n!) | every permutation |
| n ≤ 20 | O(2^n) | every subset, bitmask brute force |
| n ≤ 500 | O(n^3) | three nested loops, Floyd-Warshall |
| n ≤ 5,000 | O(n^2) | two nested loops, simple DP on pairs |
| n ≤ 10^6 | O(n log n) | sort, then scan; binary search in a loop; heap |
| n ≤ 10^8 | O(n) | one pass, two pointers, sliding window, prefix sums |
| n ≤ 10^18 | O(log n) or O(1) | binary search on the answer, fast power, a formula |

So if n is up to 2 × 10^5, an O(n^2) solution does 4 × 10^10 operations and times out. You need O(n log n) or better.

Also check the values, not just n. If a sum can pass about 2 × 10^9, it doesn't fit in `int32`. Go's `int` is 64 bits on the judge's machines, which holds up to about 9.2 × 10^18.

## Common Go operations

| Operation | Cost | Note |
|---|---|---|
| `s[i]`, `s[i] = x`, `len(s)` | O(1) | |
| `append(s, x)` | O(1) amortised | when the slice is full, Go copies it into a bigger one. Rare, so it averages out |
| `make([]int, 0, n)` then append | O(1) each, no copies | pre-size when you know n |
| `s[a:b]` | O(1) | shares the array, nothing is copied |
| `copy(dst, src)` | O(k) | k elements copied |
| insert or delete in the middle (`slices.Insert`, `slices.Delete`) | O(n) | shifts everything after it |
| `slices.Contains`, `slices.Index` | O(n) | a linear scan. Use a map if you do it in a loop |
| `m[k]`, `m[k] = v`, `delete(m, k)` | O(1) on average | |
| `for k, v := range m` | O(n) | order is random on purpose |
| `sort.Slice`, `sort.Ints`, `slices.Sort`, `slices.SortFunc` | O(n log n) | not stable. Use `sort.SliceStable` or `slices.SortStableFunc` if equal items must keep their order |
| `sort.Search`, `slices.BinarySearch` | O(log n) | the slice must already be sorted |
| `container/heap` Push, Pop | O(log n) | a priority queue |
| `len(str)` | O(1) | bytes, not characters |
| `str[i]` | O(1) | a byte |
| `[]rune(str)`, `utf8.RuneCountInString` | O(n) | |
| `str1 == str2` | O(n) | compares bytes |
| `str[a:b]` | O(1) | shares memory |
| `s += "x"` in a loop | O(n^2) total | strings can't change, so every `+=` copies the whole string |
| `strings.Builder` WriteString | O(1) amortised | O(n) total. Use this to build output |
| `strings.Split`, `strings.Fields`, `strings.Contains` | O(n) | |
| `strconv.Atoi`, `strconv.Itoa` | O(digits) | much faster than `fmt.Sscan` |

### strings.Builder vs +=

```go
// slow: O(n^2). 10^5 pieces copy about 5 × 10^9 bytes in total
s := ""
for _, w := range words {
	s += w + " "
}

// fast: O(n)
var b strings.Builder
for _, w := range words {
	b.WriteString(w)
	b.WriteByte(' ')
}
s := b.String()
```

## Fast input and output

With 10^6 numbers, `fmt.Scan` straight from `os.Stdin` and `fmt.Println` for each answer can take longer than your algorithm. Both are unbuffered: every call talks to the operating system.

### Reading: bufio.Scanner with a bigger buffer

```go
sc := bufio.NewScanner(os.Stdin)
sc.Buffer(make([]byte, 1024*1024), 1024*1024) // tokens up to 1 MB
sc.Split(bufio.ScanWords)                     // one number per Scan, across lines

readInt := func() int {
	sc.Scan()
	n, _ := strconv.Atoi(sc.Text())
	return n
}
```

- A `Scanner`'s default limit is 64 KB per token (`bufio.MaxScanTokenSize`). With the default line splitter, a longer line makes `Scan` return `false` and the rest of the input is silently lost. Check `sc.Err()` if numbers go missing. `sc.Buffer` raises the limit.
- `bufio.ScanWords` gives one number at a time, so long lines don't matter.
- If you need whole lines, keep the default `ScanLines` and raise the buffer above the longest line in the constraints.

### Reading: bufio.Reader with fmt.Fscan

Simpler to write, a little slower, and fine for most problems:

```go
in := bufio.NewReader(os.Stdin)
var n int
fmt.Fscan(in, &n)
a := make([]int, n)
for i := range a {
	fmt.Fscan(in, &a[i])
}
```

Wrap `os.Stdin` once. Don't mix `fmt.Scan` and a `bufio` reader in the same program: the reader has already pulled input ahead into its buffer.

### Writing: bufio.Writer, and never forget Flush

```go
out := bufio.NewWriter(os.Stdout)
defer out.Flush()

for _, x := range answers {
	fmt.Fprintln(out, x)
}
```

Nothing reaches the judge until `Flush`. A missing `Flush` prints nothing or only part of the output, and the judge marks it wrong. `defer out.Flush()` right after creating it is the habit. If you call `os.Exit`, deferred calls don't run, so flush first.

| Method | Time for 10^6 numbers (roughly) |
|---|---|
| `fmt.Scan` and `fmt.Println`, unbuffered | seconds |
| `bufio.Reader` + `fmt.Fscan`, `bufio.Writer` | a few hundred ms |
| `bufio.Scanner` + `ScanWords` + `strconv.Atoi`, `bufio.Writer` | tens of ms |

## Patterns and their complexity

| If the problem says | Try | Cost |
|---|---|---|
| "have I seen this before?", duplicates, counts | a map or `map[T]struct{}` set (Day 28) | O(n) |
| "a pair that sums to k" in a sorted array | two pointers from both ends (Day 29) | O(n) |
| "longest / shortest subarray with...", "best k in a row" | sliding window (Day 30) | O(n) |
| "sum of a range", many range queries | prefix sums | O(n) to build, O(1) per query |
| "k-th smallest", "closest pair", "merge intervals" | sort first (Day 31) | O(n log n) |
| "smallest x such that..." and the answer is monotonic | binary search on the answer (Day 32) | O(log range × check) |
| matching brackets, "next greater element", undo | a stack, often a monotonic one (Day 33) | O(n) |
| process in arrival order, a sliding-window maximum | a queue or deque (Day 34) | O(n) |
| every arrangement or subset | recursion and backtracking (Day 35) | O(n!) or O(2^n): n must be small |
| sorted data with inserts and lookups, traversals | a binary search tree (Day 36) | O(log n) each if balanced, O(n) if not |
| shortest path, unweighted | BFS (Day 37) | O(V + E) |
| connected parts, cycles, flood fill | DFS (Day 37) | O(V + E) |
| shortest path, weighted, no negative edges | Dijkstra with `container/heap` | O((V + E) log V) |
| "the top k", repeatedly take the smallest | a heap | O(n log k) |
| "pick the best each time" and a swap argument proves it | greedy, usually after sorting (Day 38) | O(n log n) |
| "number of ways", "minimum cost", overlapping subproblems | dynamic programming (Day 39) | states × work per state |

## Before you submit

- [ ] Put the largest n from the constraints into your complexity. Is it under about 10^8?
- [ ] Buffered input, buffered output, and `Flush`.
- [ ] Could any sum or product overflow? Use `int` (64-bit), or take the modulo the problem asks for at every step.
- [ ] Recursion depth: Go's stack grows, so deep recursion is usually fine, but 10^6 levels of DFS is slow. Switch to an explicit stack.
- [ ] Test the edge cases: n = 1, all the same value, already sorted, reverse sorted, the largest values.
- [ ] Run `go run ./tools/kata <day>` to judge it with the time limit.

## Where the course uses this

| Day | What you meet |
|---|---|
| 27 | counting operations, the 10^8 rule, fast stdin and stdout |
| 28–30 | maps and sets, two pointers, sliding window |
| 31–32 | sorting and binary search |
| 33–34 | stacks, queues and deques |
| 35–37 | recursion, trees and graphs |
| 38–39 | greedy and dynamic programming |
| 40 | the timed mock test: all of it at once |
