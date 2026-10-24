package slidingwindow

import (
	"slices"
	"strings"
	"testing"
)

func TestBestStreak(t *testing.T) {
	tests := []struct {
		name  string
		sales []int
		k     int
		want  int
	}{
		{"empty", nil, 1, 0},
		{"k is zero", []int{5, 6}, 0, 0},
		{"k too big", []int{5, 6}, 3, 0},
		{"k is one", []int{300, 800, 100}, 1, 800},
		{"pairs", []int{300, 800, 100, 900, 200}, 2, 1100},
		{"whole list", []int{300, 800, 100}, 3, 1200},
		{"best at the end", []int{1, 1, 1, 50, 60}, 2, 110},
		{"all refunds", []int{-50, -10, -30, -20}, 2, -40},
		{"same every day", []int{7, 7, 7, 7}, 3, 21},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BestStreak(tt.sales, tt.k); got != tt.want {
				t.Errorf("BestStreak(%v, %d) = %d, want %d", tt.sales, tt.k, got, tt.want)
			}
		})
	}
}

func TestBestStreakBig(t *testing.T) {
	a := make([]int, 200_000)
	for i := range a {
		a[i] = i % 1000
	}
	// The best 100 000-day window: an O(n·k) answer does 10^10 additions here.
	want := 0
	for _, x := range a[100_000:] {
		want += x
	}
	if got := BestStreak(a, 100_000); got != want {
		t.Errorf("BestStreak(big, 100000) = %d, want %d", got, want)
	}
}

func TestMovingAverage(t *testing.T) {
	tests := []struct {
		prices []int
		k      int
		want   []float64
	}{
		{nil, 1, []float64{}},
		{[]int{100}, 2, []float64{}},
		{[]int{100}, 0, []float64{}},
		{[]int{180, 190, 200, 210}, 2, []float64{185, 195, 205}},
		{[]int{180, 190, 200, 210}, 4, []float64{195}},
		{[]int{1, 2, 3, 4}, 1, []float64{1, 2, 3, 4}},
		{[]int{1, 2, 4}, 2, []float64{1.5, 3}},
	}
	for _, tt := range tests {
		got := MovingAverage(tt.prices, tt.k)
		if got == nil || !slices.Equal(got, tt.want) {
			t.Errorf("MovingAverage(%v, %d) = %#v, want %v", tt.prices, tt.k, got, tt.want)
		}
	}
}

func TestLongestUniqueRun(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"a", 1},
		{"aaaa", 1},
		{"ugaliugali", 5},
		{"abcabcbb", 3},
		{"pwwkew", 3},
		{"abba", 2},
		{"chapati na chai", 6}, // "pati n"
		{strings.Repeat("ab", 100_000), 2},
		{strings.Repeat("abcdefghijklmnopqrstuvwxyz", 5000), 26},
	}
	for _, tt := range tests {
		if got := LongestUniqueRun(tt.s); got != tt.want {
			name := tt.s
			if len(name) > 20 {
				name = name[:20] + "..."
			}
			t.Errorf("LongestUniqueRun(%q) = %d, want %d", name, got, tt.want)
		}
	}
}

func TestShortestToReach(t *testing.T) {
	tests := []struct {
		amounts []int
		target  int
		want    int
	}{
		{nil, 100, 0},
		{[]int{100}, 100, 1},
		{[]int{99}, 100, 0},
		{[]int{200, 500, 100, 1000, 300}, 1200, 2},
		{[]int{200, 500, 100, 1000, 300}, 1300, 2},
		{[]int{200, 500, 100, 1000, 300}, 2100, 5},
		{[]int{200, 500, 100, 1000, 300}, 2101, 0},
		{[]int{1, 1, 1, 1, 50}, 50, 1},
		{[]int{2, 3, 1, 2, 4, 3}, 7, 2},
	}
	for _, tt := range tests {
		if got := ShortestToReach(tt.amounts, tt.target); got != tt.want {
			t.Errorf("ShortestToReach(%v, %d) = %d, want %d", tt.amounts, tt.target, got, tt.want)
		}
	}
}

func TestShortestToReachBig(t *testing.T) {
	a := make([]int, 200_000)
	for i := range a {
		a[i] = 1
	}
	if got := ShortestToReach(a, 150_000); got != 150_000 {
		t.Errorf("ShortestToReach(200000 ones, 150000) = %d, want 150000", got)
	}
}
