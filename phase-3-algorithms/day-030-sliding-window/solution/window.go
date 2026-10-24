// Package slidingwindow answers questions about runs of consecutive values in O(n).
package slidingwindow

// BestStreak returns the largest total of k consecutive days of sales.
func BestStreak(sales []int, k int) int {
	if k < 1 || k > len(sales) {
		return 0
	}
	sum := 0
	for _, x := range sales[:k] {
		sum += x
	}
	best := sum
	for i := k; i < len(sales); i++ {
		sum += sales[i] - sales[i-k] // slide: add the new day, drop the oldest
		best = max(best, sum)
	}
	return best
}

// MovingAverage returns the average of every window of k consecutive prices.
func MovingAverage(prices []int, k int) []float64 {
	if k < 1 || k > len(prices) {
		return []float64{}
	}
	out := make([]float64, 0, len(prices)-k+1)
	sum := 0
	for i, x := range prices {
		sum += x
		if i >= k {
			sum -= prices[i-k]
		}
		if i >= k-1 {
			out = append(out, float64(sum)/float64(k))
		}
	}
	return out
}

// LongestUniqueRun returns the length of the longest run with no repeated byte.
func LongestUniqueRun(s string) int {
	var last [256]int // last[c] is 1 + the index where byte c was last seen, 0 if never
	best, left := 0, 0
	for right := 0; right < len(s); right++ {
		c := s[right]
		if last[c] > left {
			left = last[c] // jump past the earlier copy of c
		}
		last[c] = right + 1
		best = max(best, right-left+1)
	}
	return best
}

// ShortestToReach returns the length of the shortest run whose total is at least target.
func ShortestToReach(amounts []int, target int) int {
	best, sum, left := 0, 0, 0
	for right, x := range amounts {
		sum += x
		for sum >= target { // shrink from the left while the window still reaches target
			if n := right - left + 1; best == 0 || n < best {
				best = n
			}
			sum -= amounts[left]
			left++
		}
	}
	return best
}
