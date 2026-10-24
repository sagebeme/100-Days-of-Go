// Package twopointers solves array and string problems by walking two indexes.
package twopointers

// IsPalindrome reports whether s reads the same both ways, looking only at
// ASCII letters and digits and ignoring case. Everything else (spaces,
// punctuation) is skipped. An empty string is a palindrome.
//
//	IsPalindrome("Kayak")                → true
//	IsPalindrome("Was it a rat I saw?")  → true
func IsPalindrome(s string) bool {
	// Your code here.
	return false
}

// PairWithSum looks in sorted (ascending) for two positions i < j with
// sorted[i] + sorted[j] == target. It returns them and true, or 0, 0, false.
// Any valid pair is accepted. It must be O(n) and use no map.
//
//	PairWithSum([]int{50, 100, 200, 450, 800}, 650) → 2, 3, true
func PairWithSum(sorted []int, target int) (int, int, bool) {
	// Your code here.
	return 0, 0, false
}

// MergeSorted merges two ascending slices into one new ascending slice in
// O(len(a) + len(b)). Neither input is changed. Two empty inputs give an
// empty (non-nil) slice.
//
//	MergeSorted([]int{1, 4, 9}, []int{2, 4, 10}) → [1 2 4 4 9 10]
func MergeSorted(a, b []int) []int {
	// Your code here.
	return nil
}

// Dedup removes repeats from an ascending slice in place and returns the
// shortened slice, which shares sorted's backing array.
//
//	Dedup([]int{20, 20, 50, 50, 50, 100}) → [20 50 100]
func Dedup(sorted []int) []int {
	// Your code here.
	return nil
}
