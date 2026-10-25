// Package mockhackerranktest is a timed mock test: six problems, easiest first.
// Problem 7, the hardest, is the kata.
package mockhackerranktest

import (
	"cmp"
	"slices"
	"sort"
)

// MissingReceipt: the numbers 1..n+1 add up to (n+1)(n+2)/2. Subtract the ones we have.
func MissingReceipt(nums []int) int {
	m := len(nums) + 1
	total := m * (m + 1) / 2
	for _, x := range nums {
		total -= x
	}
	return total
}

// FirstUnique counts every byte, then finds the first with a count of 1.
func FirstUnique(s string) int {
	var count [256]int
	for i := 0; i < len(s); i++ {
		count[s[i]]++
	}
	for i := 0; i < len(s); i++ {
		if count[s[i]] == 1 {
			return i
		}
	}
	return -1
}

// BestTrade keeps the cheapest price so far and the best profit selling today.
func BestTrade(prices []int) int {
	if len(prices) == 0 {
		return 0
	}
	best, cheapest := 0, prices[0]
	for _, p := range prices[1:] {
		best = max(best, p-cheapest) // sell today, having bought at the cheapest day before
		cheapest = min(cheapest, p)
	}
	return best
}

// Interval is a booking from Start to End, inclusive.
type Interval struct {
	Start, End int
}

// MergeBookings sorts by start, then grows the last merged booking while the next one touches it.
func MergeBookings(bookings []Interval) []Interval {
	sorted := slices.Clone(bookings)
	slices.SortFunc(sorted, func(a, b Interval) int { return cmp.Compare(a.Start, b.Start) })
	merged := []Interval{}
	for _, b := range sorted {
		if n := len(merged); n > 0 && b.Start <= merged[n-1].End {
			merged[n-1].End = max(merged[n-1].End, b.End)
			continue
		}
		merged = append(merged, b)
	}
	return merged
}

// CountShambas flood-fills each new piece of farmland with an explicit stack.
func CountShambas(grid []string) int {
	if len(grid) == 0 {
		return 0
	}
	rows, cols := len(grid), len(grid[0])
	seen := make([]bool, rows*cols)
	count := 0
	for r := range rows {
		for c := range cols {
			if grid[r][c] != '#' || seen[r*cols+c] {
				continue
			}
			count++
			stack := []int{r*cols + c}
			seen[r*cols+c] = true
			for len(stack) > 0 {
				cell := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				cr, cc := cell/cols, cell%cols
				for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
					nr, nc := cr+d[0], cc+d[1]
					if nr < 0 || nr >= rows || nc < 0 || nc >= cols || grid[nr][nc] != '#' || seen[nr*cols+nc] {
						continue
					}
					seen[nr*cols+nc] = true
					stack = append(stack, nr*cols+nc)
				}
			}
		}
	}
	return count
}

// LongestRising keeps tails[k]: the smallest last value of any rising
// subsequence of length k+1 seen so far. tails is sorted, so each value
// finds its place by binary search: O(n log n).
func LongestRising(a []int) int {
	var tails []int
	for _, x := range a {
		i := sort.SearchInts(tails, x) // the first tail >= x
		if i == len(tails) {
			tails = append(tails, x) // x extends the longest one
		} else {
			tails[i] = x // x is a smaller ending for length i+1
		}
	}
	return len(tails)
}
