// Package sortingshowdown sorts with merge sort, quicksort and custom comparators.
package sortingshowdown

import (
	"cmp"
	"slices"
)

// Scorer is a player on the league's top scorers table.
type Scorer struct {
	Name  string
	Club  string
	Goals int
}

// MergeSort returns a new ascending slice: split in half, sort each half, merge.
func MergeSort(a []int) []int {
	if len(a) <= 1 {
		return slices.Clone(a)
	}
	mid := len(a) / 2
	left, right := MergeSort(a[:mid]), MergeSort(a[mid:])
	out := make([]int, 0, len(a))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] { // <= keeps equal values in order: merge sort is stable
			out = append(out, left[i])
			i++
		} else {
			out = append(out, right[j])
			j++
		}
	}
	out = append(out, left[i:]...)
	return append(out, right[j:]...)
}

// QuickSort sorts a in place: pick a pivot, move smaller values to the left
// and larger ones to the right, then sort each side.
func QuickSort(a []int) {
	for len(a) > 1 {
		// Hoare's scheme with the middle value as the pivot. Sorted input and
		// all-equal input both split near the middle, so neither is O(n²).
		pivot := a[len(a)/2]
		i, j := 0, len(a)-1
		for i <= j {
			for a[i] < pivot {
				i++
			}
			for a[j] > pivot {
				j--
			}
			if i <= j {
				a[i], a[j] = a[j], a[i]
				i++
				j--
			}
		}
		// Now a[:j+1] holds values <= pivot and a[i:] values >= pivot.
		// Recurse into the smaller side and loop on the bigger one, so the
		// recursion is never deeper than about log n.
		if j+1 < len(a)-i {
			QuickSort(a[:j+1])
			a = a[i:]
		} else {
			QuickSort(a[i:])
			a = a[:j+1]
		}
	}
}

// SortScorers sorts by goals, most first, then by name.
func SortScorers(table []Scorer) {
	slices.SortFunc(table, func(x, y Scorer) int {
		if c := cmp.Compare(y.Goals, x.Goals); c != 0 { // y before x: descending
			return c
		}
		return cmp.Compare(x.Name, y.Name)
	})
}
