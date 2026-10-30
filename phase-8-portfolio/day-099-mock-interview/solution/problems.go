// Package mockinterview holds the four timed problems of the mock interview. Each one
// is the kind of question a HackerRank screen or a first technical round asks.
package mockinterview

import (
	"cmp"
	"slices"
)

// Problem 1 (easy, 10 minutes).

// PairForBudget finds two different items whose prices add up to exactly budget. It
// returns their indexes i < j. If there are several pairs, it returns the one with the
// smallest j, and for that j the smallest i. ok is false when there is no pair.
//
// One pass with a map from price to the first index it appeared at: O(n).
func PairForBudget(prices []int, budget int) (i, j int, ok bool) {
	first := map[int]int{}
	for j, p := range prices {
		if i, found := first[budget-p]; found {
			return i, j, true
		}
		if _, seen := first[p]; !seen {
			first[p] = j
		}
	}
	return 0, 0, false
}

// Problem 2 (easy to medium, 15 minutes).

// Shift is a boda rider's shift, in minutes from midnight. Start < End.
type Shift struct {
	Start, End int
}

// MergeShifts merges shifts that overlap or touch (one ends when the next starts) and
// returns them sorted by Start. It doesn't change its argument. No shifts gives an
// empty result.
//
// Sort by start, then extend the last merged shift or start a new one: O(n log n).
func MergeShifts(shifts []Shift) []Shift {
	s := slices.Clone(shifts)
	slices.SortFunc(s, func(a, b Shift) int { return cmp.Compare(a.Start, b.Start) })
	out := []Shift{}
	for _, sh := range s {
		if n := len(out); n > 0 && sh.Start <= out[n-1].End {
			out[n-1].End = max(out[n-1].End, sh.End)
			continue
		}
		out = append(out, sh)
	}
	return out
}

// Problem 3 (medium, 20 minutes).

// LongestRun returns the length of the longest stretch of consecutive weeks in which a
// chama member missed at most k contributions. paid[w] is false for a missed week.
//
// A sliding window that holds at most k misses: O(n).
func LongestRun(paid []bool, k int) int {
	best, misses, left := 0, 0, 0
	for right, p := range paid {
		if !p {
			misses++
		}
		for misses > k {
			if !paid[left] {
				misses--
			}
			left++
		}
		best = max(best, right-left+1)
	}
	return best
}

// Problem 4 (medium, 20 minutes).

// FewestNotes returns the fewest notes and coins that add up to amount, using any
// number of each denomination, or -1 if it can't be done. Denominations are positive
// and may come in any order. Greedy (largest first) is wrong for some sets: with 1, 40
// and 50, the best for 80 is 40+40, not 50+1×30.
//
// Dynamic programming over amounts 0..amount: O(amount × len(denoms)).
func FewestNotes(denoms []int, amount int) int {
	if amount < 0 {
		return -1
	}
	const none = -1
	best := make([]int, amount+1) // best[a] = fewest for a, or none
	for a := 1; a <= amount; a++ {
		best[a] = none
		for _, d := range denoms {
			if d <= a && best[a-d] != none && (best[a] == none || best[a-d]+1 < best[a]) {
				best[a] = best[a-d] + 1
			}
		}
	}
	return best[amount]
}
