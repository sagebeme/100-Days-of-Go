package mockinterview

import (
	"slices"
	"testing"
)

func TestPairForBudget(t *testing.T) {
	tests := []struct {
		prices []int
		budget int
		i, j   int
		ok     bool
	}{
		{[]int{350, 120, 500, 230}, 350, 1, 3, true},
		{[]int{100, 200, 300}, 500, 1, 2, true},
		{[]int{250, 250}, 500, 0, 1, true},        // the same price twice is fine
		{[]int{250}, 500, 0, 0, false},            // but not the same item twice
		{[]int{1, 9, 5, 5, 2, 8}, 10, 0, 1, true}, // smallest j wins
		{[]int{5, 3, 5, 5}, 10, 0, 2, true},       // then smallest i
		{[]int{7, 3, 4, 6}, 10, 0, 1, true},
		{[]int{-5, 15, 20}, 10, 0, 1, true}, // negative prices: refunds
		{[]int{0, 0}, 0, 0, 1, true},
		{[]int{}, 10, 0, 0, false},
		{[]int{1, 2, 3}, 100, 0, 0, false},
	}
	for _, tt := range tests {
		i, j, ok := PairForBudget(tt.prices, tt.budget)
		if ok != tt.ok || (ok && (i != tt.i || j != tt.j)) {
			t.Errorf("PairForBudget(%v, %d) = %d, %d, %v; want %d, %d, %v", tt.prices, tt.budget, i, j, ok, tt.i, tt.j, tt.ok)
		}
	}
}

func TestPairForBudgetLarge(t *testing.T) {
	prices := make([]int, 200000)
	for i := range prices {
		prices[i] = 2 * i // all even
	}
	prices[len(prices)-1] = 1
	i, j, ok := PairForBudget(prices, 3) // only 2 + 1 works
	if !ok || i != 1 || j != len(prices)-1 {
		t.Errorf("PairForBudget(200 000 prices, 3) = %d, %d, %v; want 1, %d, true", i, j, ok, len(prices)-1)
	}
}

func TestMergeShifts(t *testing.T) {
	tests := []struct {
		in   []Shift
		want []Shift
	}{
		{[]Shift{{360, 600}, {540, 720}, {900, 1080}}, []Shift{{360, 720}, {900, 1080}}},
		{[]Shift{{900, 1080}, {360, 600}, {540, 720}}, []Shift{{360, 720}, {900, 1080}}}, // any order
		{[]Shift{{360, 600}, {600, 660}}, []Shift{{360, 660}}},                           // touching
		{[]Shift{{360, 900}, {400, 500}}, []Shift{{360, 900}}},                           // inside another
		{[]Shift{{1, 2}, {3, 4}}, []Shift{{1, 2}, {3, 4}}},
		{[]Shift{{5, 10}}, []Shift{{5, 10}}},
		{nil, []Shift{}},
	}
	for _, tt := range tests {
		before := slices.Clone(tt.in)
		got := MergeShifts(tt.in)
		if len(got) != len(tt.want) || (len(got) > 0 && !slices.Equal(got, tt.want)) {
			t.Errorf("MergeShifts(%v) = %v, want %v", before, got, tt.want)
		}
		if !slices.Equal(tt.in, before) {
			t.Errorf("MergeShifts changed its argument from %v to %v", before, tt.in)
		}
	}
}

func TestLongestRun(t *testing.T) {
	const T, F = true, false
	tests := []struct {
		paid []bool
		k    int
		want int
	}{
		{[]bool{T, T, F, T, T, T, F, F, T}, 1, 6},
		{[]bool{T, T, F, T, T, T, F, F, T}, 0, 3},
		{[]bool{T, T, F, T, T, T, F, F, T}, 2, 7},
		{[]bool{F, F, F}, 0, 0},
		{[]bool{F, F, F}, 5, 3},
		{[]bool{T}, 0, 1},
		{nil, 3, 0},
	}
	for _, tt := range tests {
		if got := LongestRun(tt.paid, tt.k); got != tt.want {
			t.Errorf("LongestRun(%v, %d) = %d, want %d", tt.paid, tt.k, got, tt.want)
		}
	}
}

func TestLongestRunLarge(t *testing.T) {
	paid := make([]bool, 300000)
	for i := range paid {
		paid[i] = i%10 != 0 // a miss every 10th week
	}
	// With k misses, the best window runs from just after one miss to just before the
	// (k+1)th miss after it: 10(k+1) - 1 weeks.
	if got := LongestRun(paid, 50); got != 509 {
		t.Errorf("LongestRun(300 000 weeks, 50) = %d, want 509", got)
	}
}

func TestFewestNotes(t *testing.T) {
	kenya := []int{1, 5, 10, 20, 50, 100, 200, 500, 1000}
	tests := []struct {
		denoms []int
		amount int
		want   int
	}{
		{kenya, 1750, 4}, // 1000 + 500 + 200 + 50
		{kenya, 0, 0},
		{kenya, 3, 3},
		{[]int{50, 40, 1}, 80, 2}, // greedy would give 31
		{[]int{1, 40, 50}, 120, 3},
		{[]int{2}, 3, -1},
		{[]int{5, 7}, 1, -1},
		{[]int{5, 7}, 24, 4}, // 5+5+7+7
		{[]int{3, 7}, 11, -1},
		{[]int{1000, 500}, 100000, 100},
	}
	for _, tt := range tests {
		if got := FewestNotes(tt.denoms, tt.amount); got != tt.want {
			t.Errorf("FewestNotes(%v, %d) = %d, want %d", tt.denoms, tt.amount, got, tt.want)
		}
	}
}
