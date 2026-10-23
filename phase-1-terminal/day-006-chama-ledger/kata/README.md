# Kata: Merry-Go-Round

A merry-go-round chama pays one member each month. The member at the front of the list gets the payout and moves to the back. Given the list and a number of months `k`, print the list as it stands after `k` months.

**Input**

- Line 1: `n` and `k`, separated by a space (1 ≤ n ≤ 100 000, 0 ≤ k ≤ 1 000 000 000)
- Line 2: `n` member numbers, separated by spaces (1 ≤ number ≤ 1 000 000)

**Output**

One line: the `n` member numbers after `k` months, separated by spaces.

**Example**

```
Input             Output
5 2               3 4 5 1 2
1 2 3 4 5
```

**Limits:** 2 seconds per test case. `k` can be a billion, so moving one member at a time, `k` times, is too slow.

**Go tip**

After `n` months the list is back where it started, so only `k % n` months matter. Read everything into a slice first. `make([]int, n)` creates a slice of `n` zeros that you can fill with `fmt.Fscan(in, &members[i])`. Then print starting from position `k % n` and wrap round to the start:

```go
start := k % n
for i := 0; i < n; i++ {
	member := members[(start+i)%n]
	// print member
}
```

Printing 100 000 numbers one `fmt.Print` at a time is slow. Write them to a `bufio.Writer` and call `Flush()` at the end.

Judge it from the repo root with `go run ./tools/kata 6`.
