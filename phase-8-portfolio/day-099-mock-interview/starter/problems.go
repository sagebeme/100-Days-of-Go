// Package mockinterview holds the four timed problems of the mock interview. Each one
// is the kind of question a HackerRank screen or a first technical round asks.
package mockinterview

// Problem 1 (easy, 10 minutes).

// PairForBudget finds two different items whose prices add up to exactly budget. It
// returns their indexes i < j. If there are several pairs, it returns the one with the
// smallest j, and for that j the smallest i. ok is false when there is no pair.
func PairForBudget(prices []int, budget int) (i, j int, ok bool) {
	// Your code here
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
func MergeShifts(shifts []Shift) []Shift {
	// Your code here
	return nil
}

// Problem 3 (medium, 20 minutes).

// LongestRun returns the length of the longest stretch of consecutive weeks in which a
// chama member missed at most k contributions. paid[w] is false for a missed week.
func LongestRun(paid []bool, k int) int {
	// Your code here
	return 0
}

// Problem 4 (medium, 20 minutes).

// FewestNotes returns the fewest notes and coins that add up to amount, using any
// number of each denomination, or -1 if it can't be done. Denominations are positive
// and may come in any order. An amount of 0 needs 0 notes.
func FewestNotes(denoms []int, amount int) int {
	// Your code here
	return 0
}
