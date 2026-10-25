package dynamicprogramming

import (
	"strings"
	"testing"
)

func TestClimbWays(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{-1, 0},
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 5},
		{10, 89},
		{50, 20_365_011_074},
		{90, 4_660_046_610_375_530_309}, // plain recursion would make about 10^19 calls
	}
	for _, tt := range tests {
		if got := ClimbWays(tt.n); got != tt.want {
			t.Errorf("ClimbWays(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestFewestNotes(t *testing.T) {
	tests := []struct {
		notes  []int
		amount int
		want   int
	}{
		{[]int{1, 5, 10}, 0, 0},
		{nil, 0, 0},
		{nil, 10, -1},
		{[]int{1, 3, 4}, 6, 2},
		{[]int{200, 500}, 300, -1},
		{[]int{200, 500}, 1100, 4}, // 500 + 200 + 200 + 200
		{[]int{1000, 500, 200, 100, 50}, 1750, 4},
		{[]int{7}, 7, 1},
		{[]int{7}, 13, -1},
		{[]int{2, 5}, 3, -1},
		{[]int{1, 5, 10, 25}, 63, 6},
		{[]int{186, 419, 83, 408}, 6249, 20},
		{[]int{50, 200, 1000}, 100_000, 100},
	}
	for _, tt := range tests {
		if got := FewestNotes(tt.notes, tt.amount); got != tt.want {
			t.Errorf("FewestNotes(%v, %d) = %d, want %d", tt.notes, tt.amount, got, tt.want)
		}
	}
}

func TestLCS(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"", "abc", 0},
		{"abc", "", 0},
		{"a", "a", 1},
		{"a", "b", 0},
		{"KAMAU", "KAMUNYA", 4},
		{"ABCBDAB", "BDCABA", 4},
		{"matatu", "matatu", 6},
		{"abc", "cba", 1},
		{"AGGTAB", "GXTXAYB", 4},
		{strings.Repeat("ab", 1000), strings.Repeat("ba", 1000), 1999},
	}
	for _, tt := range tests {
		if got := LCS(tt.a, tt.b); got != tt.want {
			a, b := tt.a, tt.b
			if len(a) > 20 {
				a = a[:20] + "..."
			}
			if len(b) > 20 {
				b = b[:20] + "..."
			}
			t.Errorf("LCS(%q, %q) = %d, want %d", a, b, got, tt.want)
		}
	}
}

func TestBestLoad(t *testing.T) {
	tests := []struct {
		name     string
		weights  []int
		values   []int
		capacity int
		want     int
	}{
		{"nothing to carry", nil, nil, 50, 0},
		{"no room", []int{10}, []int{100}, 0, 0},
		{"classic", []int{10, 20, 30}, []int{60, 100, 120}, 50, 220},
		{"each parcel once", []int{10}, []int{60}, 100, 60},
		{"one too heavy", []int{60, 10}, []int{1000, 5}, 50, 5},
		{"exact fit", []int{25, 25, 50}, []int{30, 30, 50}, 50, 60},
		{"greedy by value per kg fails", []int{6, 5, 5}, []int{12, 9, 9}, 10, 18},
		{"many parcels", []int{1, 3, 4, 5, 2, 7, 6, 8}, []int{1, 4, 5, 7, 3, 9, 8, 10}, 15, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BestLoad(tt.weights, tt.values, tt.capacity); got != tt.want {
				t.Errorf("BestLoad(%v, %v, %d) = %d, want %d", tt.weights, tt.values, tt.capacity, got, tt.want)
			}
		})
	}
}

func TestBestLoadBig(t *testing.T) {
	// 100 parcels and a capacity of 10 000: 2^100 subsets, but only 10^6 table cells.
	w := make([]int, 100)
	v := make([]int, 100)
	for i := range w {
		w[i] = 100 + i
		v[i] = 1
	}
	// Every parcel pays the same, so the best is the most parcels: the 73
	// lightest weigh 100 + 101 + ... + 172 = 9928 kg, and one more is too many.
	if got := BestLoad(w, v, 10_000); got != 73 {
		t.Errorf("BestLoad(100 parcels, 10000) = %d, want 73", got)
	}
}
