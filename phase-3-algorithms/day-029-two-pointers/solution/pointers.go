// Package twopointers solves array and string problems by walking two indexes.
package twopointers

// IsPalindrome reports whether s reads the same both ways, looking only at
// ASCII letters and digits and ignoring case.
func IsPalindrome(s string) bool {
	i, j := 0, len(s)-1
	for i < j {
		if !alnum(s[i]) {
			i++
			continue
		}
		if !alnum(s[j]) {
			j--
			continue
		}
		if lower(s[i]) != lower(s[j]) {
			return false
		}
		i++
		j--
	}
	return true
}

func alnum(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

func lower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}

// PairWithSum finds i < j with sorted[i] + sorted[j] == target, walking in from both ends.
func PairWithSum(sorted []int, target int) (int, int, bool) {
	i, j := 0, len(sorted)-1
	for i < j {
		switch sum := sorted[i] + sorted[j]; {
		case sum == target:
			return i, j, true
		case sum < target:
			i++ // need more: only moving the left end up can help
		default:
			j-- // too much: move the right end down
		}
	}
	return 0, 0, false
}

// MergeSorted merges two ascending slices into a new ascending slice.
func MergeSorted(a, b []int) []int {
	out := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			out = append(out, a[i])
			i++
		} else {
			out = append(out, b[j])
			j++
		}
	}
	out = append(out, a[i:]...)
	return append(out, b[j:]...)
}

// Dedup removes repeats from an ascending slice in place.
func Dedup(sorted []int) []int {
	if len(sorted) == 0 {
		return sorted
	}
	w := 1 // sorted[:w] holds the distinct values so far
	for r := 1; r < len(sorted); r++ {
		if sorted[r] != sorted[w-1] {
			sorted[w] = sorted[r]
			w++
		}
	}
	return sorted[:w]
}
