// Package mockhackerranktest is a timed mock test: six problems, easiest first.
// Problem 7, the hardest, is the kata.
package mockhackerranktest

// Problem 1 (easy).
//
// MissingReceipt: a till printed receipts numbered 1 to n+1, and one is lost.
// nums holds the other n numbers in any order, each once. Return the missing
// number. n can be 10^6 and the numbers must not be sorted.
//
//	MissingReceipt([]int{3, 1, 4}) → 2
func MissingReceipt(nums []int) int {
	// Your code here.
	return 0
}

// Problem 2 (easy).
//
// FirstUnique returns the index of the first byte in s that appears exactly
// once in s, or -1 if there is none. len(s) can be 10^6.
//
//	FirstUnique("mamboleo") → 1    'm' repeats, 'a' doesn't
func FirstUnique(s string) int {
	// Your code here.
	return 0
}

// Problem 3 (easy to medium).
//
// BestTrade: prices[i] is the price of a Safaricom share on day i. Buy on one
// day and sell on a later day. Return the biggest profit, or 0 if no trade
// makes money. len(prices) can be 10^6.
//
//	BestTrade([]int{20, 18, 25, 17, 22, 30, 21}) → 13    buy at 17, sell at 30
func BestTrade(prices []int) int {
	// Your code here.
	return 0
}

// Interval is a booking from Start to End, inclusive. Start <= End.
type Interval struct {
	Start, End int
}

// Problem 4 (medium).
//
// MergeBookings merges bookings that overlap or touch ([1 3] and [3 5] become
// [1 5]; [1 3] and [4 5] stay apart) and returns them sorted by Start. The
// input can be in any order and must not be changed. No bookings gives an
// empty (non-nil) slice.
//
//	MergeBookings([]Interval{{8, 10}, {1, 3}, {2, 6}, {15, 18}}) → [{1 6} {8 10} {15 18}]
func MergeBookings(bookings []Interval) []Interval {
	// Your code here.
	return nil
}

// Problem 5 (medium).
//
// CountShambas: grid is a satellite map, every row the same length. '#' is
// farmland and '.' is not. Farmland cells that touch up, down, left or right
// belong to the same shamba (diagonals don't count). Return the number of
// shambas. The grid can be 1000 × 1000.
//
//	CountShambas([]string{"##..", "#..#", "..##"}) → 2
func CountShambas(grid []string) int {
	// Your code here.
	return 0
}

// Problem 6 (hard).
//
// LongestRising returns the length of the longest strictly increasing
// subsequence of a: the most values you can pick, keeping their order, so
// each is bigger than the one before. len(a) can be 2 × 10^5, so O(n²) is too slow.
//
//	LongestRising([]int{10, 9, 2, 5, 3, 7, 101, 18}) → 4    2, 3, 7, 18
func LongestRising(a []int) int {
	// Your code here.
	return 0
}
