package howfastisfast

import (
	"slices"
	"testing"
)

func TestSumTo(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{-5, 0},
		{0, 0},
		{1, 1},
		{2, 3},
		{10, 55},
		{365, 66795},
		{1_000_000_000, 500_000_000_500_000_000}, // a loop would do 10^9 additions
	}
	for _, tt := range tests {
		if got := SumTo(tt.n); got != tt.want {
			t.Errorf("SumTo(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestPrefixSums(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		want []int
	}{
		{"empty", []int{}, []int{0}},
		{"one day", []int{750}, []int{0, 750}},
		{"three days", []int{500, 200, 300}, []int{0, 500, 700, 1000}},
		{"refunds are negative", []int{100, -40, 0, 60}, []int{0, 100, 60, 60, 120}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrefixSums(tt.a); !slices.Equal(got, tt.want) {
				t.Errorf("PrefixSums(%v)\n got: %v\nwant: %v", tt.a, got, tt.want)
			}
		})
	}
}

func TestRangeSum(t *testing.T) {
	takings := []int{1200, 800, -300, 4500, 0, 950, 950}
	p := PrefixSums(takings)
	for l := 0; l < len(takings); l++ {
		for r := l; r < len(takings); r++ {
			want := rangeSumNaive(takings, l, r)
			if got := RangeSum(p, l, r); got != want {
				t.Errorf("RangeSum(prefix of %v, %d, %d) = %d, want %d", takings, l, r, got, want)
			}
		}
	}
}

func TestRangeSumBig(t *testing.T) {
	a := make([]int, 200_000)
	for i := range a {
		a[i] = 1_000_000_000 // Ksh 1 billion a day: totals need 64 bits
	}
	p := PrefixSums(a)
	if got, want := RangeSum(p, 0, len(a)-1), 200_000*1_000_000_000; got != want {
		t.Errorf("RangeSum over 200000 days = %d, want %d", got, want)
	}
	if got, want := RangeSum(p, 1000, 1000), 1_000_000_000; got != want {
		t.Errorf("RangeSum(p, 1000, 1000) = %d, want %d", got, want)
	}
}

func TestKnockoutRounds(t *testing.T) {
	tests := []struct {
		teams, want int
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 3},
		{16, 4},
		{17, 5},
		{47, 6}, // one team from each county
		{1 << 40, 40},
		{1<<40 + 1, 41},
	}
	for _, tt := range tests {
		if got := KnockoutRounds(tt.teams); got != tt.want {
			t.Errorf("KnockoutRounds(%d) = %d, want %d", tt.teams, got, tt.want)
		}
	}
}

// rangeSumNaive is the O(n) way: add up every day in the range.
func rangeSumNaive(a []int, l, r int) int {
	total := 0
	for i := l; i <= r; i++ {
		total += a[i]
	}
	return total
}

// Run the benchmarks with:
//
//	go test -bench . ./phase-3-algorithms/day-027-how-fast-is-fast/starter

func benchData() []int {
	a := make([]int, 100_000)
	for i := range a {
		a[i] = i % 1000
	}
	return a
}

func BenchmarkRangeSumNaive(b *testing.B) {
	a := benchData()
	for b.Loop() {
		rangeSumNaive(a, 10, len(a)-10)
	}
}

func BenchmarkRangeSumPrefix(b *testing.B) {
	a := benchData()
	p := PrefixSums(a) // built once: b.Loop starts the clock on its first call
	for b.Loop() {
		RangeSum(p, 10, len(a)-10)
	}
}
