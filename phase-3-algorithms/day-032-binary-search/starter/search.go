// Package binarysearch halves the search space: in sorted data, and over answers.
package binarysearch

// Search returns an index of x in sorted (ascending), or -1 if x isn't there.
// Write the loop yourself: no sort or slices package. It must be O(log n).
//
//	Search([]int{50, 100, 200, 500, 1000}, 200) → 2
func Search(sorted []int, x int) int {
	// Your code here.
	return -1
}

// FirstAtLeast returns the smallest index i with sorted[i] >= x, or len(sorted)
// if every value is smaller. With repeats, it's the first copy. O(log n).
//
//	FirstAtLeast([]int{50, 100, 100, 500}, 100) → 1
//	FirstAtLeast([]int{50, 100, 100, 500}, 101) → 3
func FirstAtLeast(sorted []int, x int) int {
	// Your code here.
	return 0
}

// CountInRange returns how many values in sorted lie between lo and hi,
// inclusive. If lo > hi it returns 0. O(log n).
//
//	CountInRange([]int{50, 100, 100, 500, 1000}, 100, 500) → 3
func CountInRange(sorted []int, lo, hi int) int {
	// Your code here.
	return 0
}

// IntSqrt returns the largest r with r*r <= n, for 0 <= n <= 10^18, by binary
// searching on the answer. No math package.
//
//	IntSqrt(26) → 5
func IntSqrt(n int) int {
	// Your code here.
	return 0
}
