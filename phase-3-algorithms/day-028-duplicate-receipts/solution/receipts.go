// Package duplicatereceipts finds repeats and counts things with maps.
package duplicatereceipts

// FirstDuplicate returns the first code that shows up for the second time.
func FirstDuplicate(codes []string) (string, bool) {
	seen := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		if _, ok := seen[c]; ok {
			return c, true
		}
		seen[c] = struct{}{}
	}
	return "", false
}

// CountFrequencies returns how many times each item appears.
func CountFrequencies(items []string) map[string]int {
	counts := make(map[string]int, len(items))
	for _, it := range items {
		counts[it]++
	}
	return counts
}

// TopSeller returns the item sold most often, breaking ties alphabetically.
func TopSeller(items []string) (string, int) {
	best, bestCount := "", 0
	for item, n := range CountFrequencies(items) {
		if n > bestCount || (n == bestCount && item < best) {
			best, bestCount = item, n
		}
	}
	return best, bestCount
}

// TwoSum finds i < j with amounts[i] + amounts[j] == target in one pass.
func TwoSum(amounts []int, target int) (int, int, bool) {
	firstAt := make(map[int]int, len(amounts)) // value → first index it appeared at
	for j, x := range amounts {
		if i, ok := firstAt[target-x]; ok {
			return i, j, true
		}
		if _, ok := firstAt[x]; !ok {
			firstAt[x] = j
		}
	}
	return 0, 0, false
}
