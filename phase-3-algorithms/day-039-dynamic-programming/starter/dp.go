// Package dynamicprogramming solves each small problem once and reuses the answer.
package dynamicprogramming

// ClimbWays returns how many ways there are to climb n steps taking 1 or 2
// steps at a time. ClimbWays(0) is 1 (stand still) and a negative n gives 0.
// n is at most 90. Use memoisation: plain recursion takes years for n = 90.
//
//	ClimbWays(4) → 5    1111, 112, 121, 211, 22
func ClimbWays(n int) int {
	// Your code here.
	return 0
}

// FewestNotes returns the smallest number of notes that add up to exactly
// amount, using as many of each denomination as you like, or -1 if it can't
// be done. amount 0 needs 0 notes. Use tabulation.
//
//	FewestNotes([]int{1, 3, 4}, 6)    → 2     3 + 3 (greedy would say 4 + 1 + 1)
//	FewestNotes([]int{200, 500}, 300) → -1
func FewestNotes(notes []int, amount int) int {
	// Your code here.
	return 0
}

// LCS returns the length of the longest common subsequence of a and b: the
// longest string you can get from both by deleting some bytes (keeping the
// order of the rest).
//
//	LCS("KAMAU", "KAMUNYA") → 4    "KAMU"
func LCS(a, b string) int {
	// Your code here.
	return 0
}

// BestLoad returns the most a boda rider can earn in one trip. Parcel i
// weighs weights[i] kg and pays values[i] shillings. Each parcel is taken
// whole or left behind, and the load can't be over capacity kg.
//
//	BestLoad([]int{10, 20, 30}, []int{60, 100, 120}, 50) → 220   the 20 and 30 kg parcels
func BestLoad(weights, values []int, capacity int) int {
	// Your code here.
	return 0
}
