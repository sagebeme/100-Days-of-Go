// Package dynamicprogramming solves each small problem once and reuses the answer.
package dynamicprogramming

// ClimbWays counts the ways to climb n steps, remembering each answer.
func ClimbWays(n int) int {
	memo := map[int]int{}
	var ways func(n int) int
	ways = func(n int) int {
		if n < 0 {
			return 0
		}
		if n == 0 {
			return 1
		}
		if w, ok := memo[n]; ok {
			return w
		}
		// The last move was 1 step or 2 steps.
		w := ways(n-1) + ways(n-2)
		memo[n] = w
		return w
	}
	return ways(n)
}

// FewestNotes fills a table: best[x] is the fewest notes that make x.
func FewestNotes(notes []int, amount int) int {
	if amount < 0 {
		return -1
	}
	const impossible = -1
	best := make([]int, amount+1)
	for x := 1; x <= amount; x++ {
		best[x] = impossible
		for _, n := range notes {
			if n <= x && best[x-n] != impossible && (best[x] == impossible || best[x-n]+1 < best[x]) {
				best[x] = best[x-n] + 1
			}
		}
	}
	return best[amount]
}

// LCS fills a table: t[i][j] is the LCS of a[:i] and b[:j].
func LCS(a, b string) int {
	t := make([][]int, len(a)+1)
	for i := range t {
		t[i] = make([]int, len(b)+1)
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				t[i][j] = t[i-1][j-1] + 1 // use the matching byte
			} else {
				t[i][j] = max(t[i-1][j], t[i][j-1]) // drop a byte from one of them
			}
		}
	}
	return t[len(a)][len(b)]
}

// BestLoad is 0/1 knapsack: best[c] is the most you can earn with capacity c.
func BestLoad(weights, values []int, capacity int) int {
	if capacity < 0 {
		return 0
	}
	best := make([]int, capacity+1)
	for i, w := range weights {
		// Go down, so best[c-w] still means "without parcel i": each parcel is used once.
		for c := capacity; c >= w; c-- {
			best[c] = max(best[c], best[c-w]+values[i])
		}
	}
	return best[capacity]
}
