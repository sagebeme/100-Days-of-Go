// Package howfastisfast compares slow and fast ways to answer the same question.
package howfastisfast

// SumTo returns 1 + 2 + ... + n, or 0 when n < 1. It must be O(1): no loop.
//
//	SumTo(10) → 55
func SumTo(n int) int {
	// Your code here.
	return 0
}

// PrefixSums returns p with len(a)+1 entries, where p[i] is the sum of a[0] to a[i-1].
// p[0] is always 0.
//
//	PrefixSums([]int{500, 200, 300}) → [0 500 700 1000]
func PrefixSums(a []int) []int {
	// Your code here.
	return nil
}

// RangeSum returns a[l] + ... + a[r] (0-based, inclusive) in O(1), using
// prefix := PrefixSums(a). It assumes 0 <= l <= r < len(a).
//
//	RangeSum([0 500 700 1000], 1, 2) → 500
func RangeSum(prefix []int, l, r int) int {
	// Your code here.
	return 0
}

// KnockoutRounds returns how many rounds a knockout tournament with the given
// number of teams needs to find a winner. Each round halves the teams that are
// left (an odd team out gets a bye and goes through). 0 or 1 teams need 0 rounds.
//
//	KnockoutRounds(16) → 4
//	KnockoutRounds(17) → 5
func KnockoutRounds(teams int) int {
	// Your code here.
	return 0
}
