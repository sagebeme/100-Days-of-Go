package sortingshowdown

import (
	"math/rand/v2"
	"slices"
	"testing"
)

var sortCases = []struct {
	name string
	in   []int
}{
	{"empty", []int{}},
	{"one", []int{42}},
	{"two backwards", []int{2, 1}},
	{"small", []int{5, 2, 9, 1}},
	{"duplicates", []int{3, 1, 3, 3, 0, 1}},
	{"negatives", []int{-5, 10, 0, -50, 7}},
	{"already sorted", []int{1, 2, 3, 4, 5, 6}},
	{"reverse sorted", []int{6, 5, 4, 3, 2, 1}},
	{"all equal", []int{7, 7, 7, 7}},
}

func bigCases() map[string][]int {
	r := rand.New(rand.NewPCG(31, 31)) // fixed seed: the same numbers every run
	random := make([]int, 200_000)
	for i := range random {
		random[i] = r.IntN(1_000_000) - 500_000
	}
	sorted := make([]int, 200_000)
	for i := range sorted {
		sorted[i] = i
	}
	equal := make([]int, 200_000)
	return map[string][]int{"200000 random": random, "200000 sorted": sorted, "200000 equal": equal}
}

func TestMergeSort(t *testing.T) {
	for _, tt := range sortCases {
		t.Run(tt.name, func(t *testing.T) {
			in := slices.Clone(tt.in)
			want := slices.Sorted(slices.Values(tt.in))
			got := MergeSort(in)
			if !slices.Equal(got, want) {
				t.Errorf("MergeSort(%v) = %v, want %v", tt.in, got, want)
			}
			if !slices.Equal(in, tt.in) {
				t.Errorf("MergeSort changed its input to %v", in)
			}
		})
	}
	for name, in := range bigCases() {
		want := slices.Sorted(slices.Values(in))
		if got := MergeSort(in); !slices.Equal(got, want) {
			t.Errorf("MergeSort(%s): not sorted correctly", name)
		}
	}
}

func TestQuickSort(t *testing.T) {
	for _, tt := range sortCases {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Clone(tt.in)
			want := slices.Sorted(slices.Values(tt.in))
			QuickSort(got)
			if !slices.Equal(got, want) {
				t.Errorf("QuickSort(%v) gave %v, want %v", tt.in, got, want)
			}
		})
	}
	// Sorted input is the classic trap: a first-element pivot makes quicksort O(n²) here.
	for name, in := range bigCases() {
		want := slices.Sorted(slices.Values(in))
		QuickSort(in)
		if !slices.Equal(in, want) {
			t.Errorf("QuickSort(%s): not sorted correctly", name)
		}
	}
}

func TestSortScorers(t *testing.T) {
	table := []Scorer{
		{"Otieno", "Gor Mahia", 12},
		{"Wekesa", "AFC Leopards", 15},
		{"Achieng", "Kenya Police", 12},
		{"Mwangi", "Tusker", 9},
		{"Baraka", "Gor Mahia", 15},
		{"Zawadi", "Bandari", 0},
	}
	want := []Scorer{
		{"Baraka", "Gor Mahia", 15},
		{"Wekesa", "AFC Leopards", 15},
		{"Achieng", "Kenya Police", 12},
		{"Otieno", "Gor Mahia", 12},
		{"Mwangi", "Tusker", 9},
		{"Zawadi", "Bandari", 0},
	}
	SortScorers(table)
	if !slices.Equal(table, want) {
		t.Errorf("SortScorers gave\n%v\nwant\n%v", table, want)
	}

	var empty []Scorer
	SortScorers(empty) // must not panic
}
