// Package binarysearch halves the search space: in sorted data, and over answers.
package binarysearch

// Search returns an index of x in sorted, or -1.
func Search(sorted []int, x int) int {
	lo, hi := 0, len(sorted)-1 // x, if present, is in sorted[lo..hi]
	for lo <= hi {
		mid := lo + (hi-lo)/2
		switch {
		case sorted[mid] == x:
			return mid
		case sorted[mid] < x:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}
	return -1
}

// FirstAtLeast returns the smallest index i with sorted[i] >= x, or len(sorted).
func FirstAtLeast(sorted []int, x int) int {
	lo, hi := 0, len(sorted) // the answer is in [lo, hi]
	for lo < hi {
		mid := lo + (hi-lo)/2
		if sorted[mid] >= x {
			hi = mid // mid might be the answer: keep it
		} else {
			lo = mid + 1
		}
	}
	return lo
}

// CountInRange returns how many values lie in [lo, hi].
func CountInRange(sorted []int, lo, hi int) int {
	if lo > hi {
		return 0
	}
	return FirstAtLeast(sorted, hi+1) - FirstAtLeast(sorted, lo)
}

// IntSqrt returns the largest r with r*r <= n.
func IntSqrt(n int) int {
	lo, hi := 0, 1_000_000_000 // 10^9 squared is 10^18, the largest n
	for lo < hi {
		mid := lo + (hi-lo+1)/2 // round up so lo = mid always makes progress
		if mid <= n/mid {       // mid*mid <= n, without overflowing
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}
