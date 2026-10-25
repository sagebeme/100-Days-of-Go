package mockhackerranktest

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

// Run one problem at a time while you work, for example:
//
//	go test -run Problem3 ./phase-3-algorithms/day-040-mock-hackerrank-test/starter

func TestProblem1MissingReceipt(t *testing.T) {
	tests := []struct {
		nums []int
		want int
	}{
		{[]int{}, 1},
		{[]int{1}, 2},
		{[]int{2}, 1},
		{[]int{3, 1, 4}, 2},
		{[]int{1, 2, 3, 4}, 5},
		{[]int{5, 4, 3, 2}, 1},
	}
	for _, tt := range tests {
		if got := MissingReceipt(tt.nums); got != tt.want {
			t.Errorf("MissingReceipt(%v) = %d, want %d", tt.nums, got, tt.want)
		}
	}
	r := rand.New(rand.NewPCG(40, 1))
	big := r.Perm(1_000_001) // 0 .. 10^6
	missing := 0
	for i, x := range big {
		big[i] = x + 1 // 1 .. 10^6 + 1
		if x+1 == 777_777 {
			missing = i
		}
	}
	big = slices.Delete(big, missing, missing+1)
	if got := MissingReceipt(big); got != 777_777 {
		t.Errorf("MissingReceipt(10^6 receipts) = %d, want 777777", got)
	}
}

func TestProblem2FirstUnique(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", -1},
		{"a", 0},
		{"aa", -1},
		{"mamboleo", 1},
		{"aabbc", 4},
		{"abcabc", -1},
		{"Aa", 0}, // case matters
		{"leetcode", 0},
		{strings.Repeat("ab", 500_000) + "z", 1_000_000},
	}
	for _, tt := range tests {
		if got := FirstUnique(tt.s); got != tt.want {
			name := tt.s
			if len(name) > 20 {
				name = name[:20] + "..."
			}
			t.Errorf("FirstUnique(%q) = %d, want %d", name, got, tt.want)
		}
	}
}

func TestProblem3BestTrade(t *testing.T) {
	tests := []struct {
		prices []int
		want   int
	}{
		{nil, 0},
		{[]int{25}, 0},
		{[]int{20, 18, 25, 17, 22, 30, 21}, 13},
		{[]int{30, 25, 20, 15}, 0},
		{[]int{15, 15, 15}, 0},
		{[]int{1, 100}, 99},
		{[]int{9, 1, 8, 2, 10}, 9},
		{[]int{2, 4, 1}, 2}, // the cheapest day comes too late to use
	}
	for _, tt := range tests {
		if got := BestTrade(tt.prices); got != tt.want {
			t.Errorf("BestTrade(%v) = %d, want %d", tt.prices, got, tt.want)
		}
	}
	big := make([]int, 1_000_000)
	for i := range big {
		big[i] = 2_000_000 - i // falling all year...
	}
	big[len(big)-1] = 5_000_000 // ...then a jump on the last day
	if got, want := BestTrade(big), 5_000_000-(2_000_000-(len(big)-2)); got != want {
		t.Errorf("BestTrade(big) = %d, want %d", got, want)
	}
}

func TestProblem4MergeBookings(t *testing.T) {
	tests := []struct {
		name string
		in   []Interval
		want []Interval
	}{
		{"none", nil, []Interval{}},
		{"one", []Interval{{4, 9}}, []Interval{{4, 9}}},
		{"example", []Interval{{8, 10}, {1, 3}, {2, 6}, {15, 18}}, []Interval{{1, 6}, {8, 10}, {15, 18}}},
		{"touching merges", []Interval{{1, 3}, {3, 5}}, []Interval{{1, 5}}},
		{"one apart stays apart", []Interval{{1, 3}, {4, 5}}, []Interval{{1, 3}, {4, 5}}},
		{"inside another", []Interval{{1, 10}, {2, 3}, {4, 5}}, []Interval{{1, 10}}},
		{"chain", []Interval{{5, 6}, {1, 2}, {2, 3}, {3, 5}}, []Interval{{1, 6}}},
		{"single points", []Interval{{7, 7}, {7, 7}, {8, 8}}, []Interval{{7, 7}, {8, 8}}},
		{"negative hours", []Interval{{-5, -1}, {-2, 0}}, []Interval{{-5, 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := slices.Clone(tt.in)
			got := MergeBookings(tt.in)
			if got == nil || !slices.Equal(got, tt.want) {
				t.Errorf("MergeBookings(%v) = %#v, want %v", tt.in, got, tt.want)
			}
			if !slices.Equal(before, tt.in) {
				t.Errorf("MergeBookings changed its input to %v", tt.in)
			}
		})
	}
	big := make([]Interval, 200_000)
	for i := range big {
		big[i] = Interval{2 * i, 2*i + 2} // each touches the next: one big booking
	}
	slices.Reverse(big)
	if got := MergeBookings(big); !slices.Equal(got, []Interval{{0, 400_000}}) {
		t.Errorf("MergeBookings(200000 touching) = %d bookings, want [{0 400000}]", len(got))
	}
}

func TestProblem5CountShambas(t *testing.T) {
	tests := []struct {
		grid []string
		want int
	}{
		{nil, 0},
		{[]string{"."}, 0},
		{[]string{"#"}, 1},
		{[]string{"##..", "#..#", "..##"}, 2},
		{[]string{"#.#", ".#.", "#.#"}, 5}, // diagonals don't join
		{[]string{"###", "#.#", "###"}, 1}, // a ring is one shamba
		{[]string{"#.#.#"}, 3},
		{[]string{"#", ".", "#", "#"}, 2},
	}
	for _, tt := range tests {
		if got := CountShambas(tt.grid); got != tt.want {
			t.Errorf("CountShambas(%q) = %d, want %d", tt.grid, got, tt.want)
		}
	}
	// 1000 × 1000, a snake of farmland that winds through every row: one shamba,
	// and a recursive flood fill would go 500 000 cells deep.
	rows := make([]string, 1000)
	for r := range rows {
		switch {
		case r%2 == 0:
			rows[r] = strings.Repeat("#", 1000)
		case r%4 == 1:
			rows[r] = strings.Repeat(".", 999) + "#"
		default:
			rows[r] = "#" + strings.Repeat(".", 999)
		}
	}
	if got := CountShambas(rows); got != 1 {
		t.Errorf("CountShambas(1000 × 1000 snake) = %d, want 1", got)
	}
}

func TestProblem6LongestRising(t *testing.T) {
	tests := []struct {
		a    []int
		want int
	}{
		{nil, 0},
		{[]int{5}, 1},
		{[]int{10, 9, 2, 5, 3, 7, 101, 18}, 4},
		{[]int{7, 7, 7, 7}, 1}, // strictly increasing: equal values don't count
		{[]int{1, 2, 3, 4, 5}, 5},
		{[]int{5, 4, 3, 2, 1}, 1},
		{[]int{0, 1, 0, 3, 2, 3}, 4},
		{[]int{-3, -1, -2, 0}, 3},
		{[]int{3, 10, 2, 1, 20}, 3},
	}
	for _, tt := range tests {
		if got := LongestRising(tt.a); got != tt.want {
			t.Errorf("LongestRising(%v) = %d, want %d", tt.a, got, tt.want)
		}
	}
	// 2 × 10^5 values: the O(n²) DP does 2 × 10^10 steps and this test times out.
	big := make([]int, 200_000)
	for i := range big {
		big[i] = (i%1000)*1000 + i/1000 // 200 rising runs of 1000, interleaved
	}
	if got := LongestRising(big); got != 1000+199 {
		t.Errorf("LongestRising(big) = %d, want %d", got, 1000+199)
	}
}
