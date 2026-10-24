// Package howfastisfast compares slow and fast ways to answer the same question.
package howfastisfast

// SumTo returns 1 + 2 + ... + n, or 0 when n < 1. It must be O(1): no loop.
func SumTo(n int) int {
	if n < 1 {
		return 0
	}
	return n * (n + 1) / 2
}

// PrefixSums returns p with len(a)+1 entries, where p[i] is the sum of a[0] to a[i-1].
func PrefixSums(a []int) []int {
	p := make([]int, len(a)+1)
	for i, x := range a {
		p[i+1] = p[i] + x
	}
	return p
}

// RangeSum returns a[l] + ... + a[r] (0-based, inclusive) in O(1).
func RangeSum(prefix []int, l, r int) int {
	return prefix[r+1] - prefix[l]
}

// KnockoutRounds returns how many rounds a knockout tournament needs.
func KnockoutRounds(teams int) int {
	rounds := 0
	for teams > 1 {
		teams = (teams + 1) / 2 // the odd team out gets a bye
		rounds++
	}
	return rounds
}
