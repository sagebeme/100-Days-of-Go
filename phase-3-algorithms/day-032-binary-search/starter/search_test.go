package binarysearch

import "testing"

var prices = []int{50, 100, 100, 200, 500, 500, 500, 1000}

func TestSearch(t *testing.T) {
	tests := []struct {
		sorted []int
		x      int
		found  bool
	}{
		{nil, 5, false},
		{[]int{5}, 5, true},
		{[]int{5}, 6, false},
		{[]int{50, 100, 200, 500, 1000}, 200, true},
		{[]int{50, 100, 200, 500, 1000}, 50, true},
		{[]int{50, 100, 200, 500, 1000}, 1000, true},
		{[]int{50, 100, 200, 500, 1000}, 49, false},
		{[]int{50, 100, 200, 500, 1000}, 1001, false},
		{[]int{50, 100, 200, 500, 1000}, 150, false},
		{prices, 500, true},
		{[]int{-30, -20, -10}, -20, true},
	}
	for _, tt := range tests {
		got := Search(tt.sorted, tt.x)
		if !tt.found {
			if got != -1 {
				t.Errorf("Search(%v, %d) = %d, want -1", tt.sorted, tt.x, got)
			}
			continue
		}
		if got < 0 || got >= len(tt.sorted) || tt.sorted[got] != tt.x {
			t.Errorf("Search(%v, %d) = %d, which isn't an index of %d", tt.sorted, tt.x, got, tt.x)
		}
	}
}

func TestSearchBig(t *testing.T) {
	a := make([]int, 1_000_000)
	for i := range a {
		a[i] = 3 * i
	}
	// 200 000 searches: a linear scan would be 10^11 steps.
	for q := 0; q < 200_000; q++ {
		x := (q * 7919) % (3 * len(a))
		got := Search(a, x)
		want := -1
		if x%3 == 0 {
			want = x / 3
		}
		if got != want {
			t.Fatalf("Search(big, %d) = %d, want %d", x, got, want)
		}
	}
}

func TestFirstAtLeast(t *testing.T) {
	tests := []struct {
		sorted []int
		x      int
		want   int
	}{
		{nil, 5, 0},
		{[]int{10}, 5, 0},
		{[]int{10}, 10, 0},
		{[]int{10}, 11, 1},
		{prices, 0, 0},
		{prices, 50, 0},
		{prices, 100, 1},
		{prices, 101, 3},
		{prices, 500, 4},
		{prices, 1000, 7},
		{prices, 1001, 8},
		{[]int{7, 7, 7, 7}, 7, 0},
	}
	for _, tt := range tests {
		if got := FirstAtLeast(tt.sorted, tt.x); got != tt.want {
			t.Errorf("FirstAtLeast(%v, %d) = %d, want %d", tt.sorted, tt.x, got, tt.want)
		}
	}
}

func TestCountInRange(t *testing.T) {
	tests := []struct {
		sorted []int
		lo, hi int
		want   int
	}{
		{nil, 0, 100, 0},
		{prices, 100, 500, 6},
		{prices, 101, 499, 1},
		{prices, 500, 500, 3},
		{prices, 0, 10_000, 8},
		{prices, 1001, 5000, 0},
		{prices, 500, 100, 0},
		{prices, 0, 49, 0},
	}
	for _, tt := range tests {
		if got := CountInRange(tt.sorted, tt.lo, tt.hi); got != tt.want {
			t.Errorf("CountInRange(%v, %d, %d) = %d, want %d", tt.sorted, tt.lo, tt.hi, got, tt.want)
		}
	}
}

func TestIntSqrt(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 1},
		{4, 2},
		{26, 5},
		{99, 9},
		{100, 10},
		{999_999_999_999, 999_999},
		{1_000_000_000_000_000_000, 1_000_000_000},
		{999_999_999_999_999_999, 999_999_999},
	}
	for _, tt := range tests {
		if got := IntSqrt(tt.n); got != tt.want {
			t.Errorf("IntSqrt(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}
