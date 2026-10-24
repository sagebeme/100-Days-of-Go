// Package duplicatereceipts finds repeats and counts things with maps.
package duplicatereceipts

// FirstDuplicate returns the first code that shows up for the second time,
// reading from the start, and true. If no code repeats it returns "", false.
//
//	FirstDuplicate([]string{"QFT1", "RKD4", "SAB0", "RKD4", "QFT1"}) → "RKD4", true
func FirstDuplicate(codes []string) (string, bool) {
	// Your code here.
	return "", false
}

// CountFrequencies returns how many times each item appears.
// An empty list gives an empty (non-nil) map.
//
//	CountFrequencies([]string{"unga", "sukari", "unga"}) → map[sukari:1 unga:2]
func CountFrequencies(items []string) map[string]int {
	// Your code here.
	return nil
}

// TopSeller returns the item sold most often and how many times it was sold.
// On a tie, it returns the item that comes first alphabetically.
// An empty list gives "", 0.
//
//	TopSeller([]string{"mandazi", "chai", "chai", "mandazi", "samosa"}) → "chai", 2
func TopSeller(items []string) (string, int) {
	// Your code here.
	return "", 0
}

// TwoSum finds two different positions i < j with amounts[i] + amounts[j] == target.
// If several pairs work, it returns the one with the smallest j, and for that j
// the smallest i. If none works it returns 0, 0, false. It must be O(n).
//
//	TwoSum([]int{300, 700, 150, 850}, 1000) → 0, 1, true
func TwoSum(amounts []int, target int) (int, int, bool) {
	// Your code here.
	return 0, 0, false
}
