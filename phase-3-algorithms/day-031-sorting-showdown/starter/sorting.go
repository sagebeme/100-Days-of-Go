// Package sortingshowdown sorts with merge sort, quicksort and custom comparators.
package sortingshowdown

// Scorer is a player on the league's top scorers table.
type Scorer struct {
	Name  string
	Club  string
	Goals int
}

// MergeSort returns a new slice with a's values in ascending order. a is not
// changed. It must be O(n log n) in every case.
//
//	MergeSort([]int{5, 2, 9, 1}) → [1 2 5 9]
func MergeSort(a []int) []int {
	// Your code here.
	return nil
}

// QuickSort sorts a in place, in ascending order.
//
//	a := []int{5, 2, 9, 1}
//	QuickSort(a) // a is now [1 2 5 9]
func QuickSort(a []int) {
	// Your code here.
}

// SortScorers sorts the table in place: most goals first, then by name A to Z
// for players on the same number of goals.
func SortScorers(table []Scorer) {
	// Your code here.
}
