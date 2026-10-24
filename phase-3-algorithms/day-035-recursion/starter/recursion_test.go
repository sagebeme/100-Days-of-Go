package recursion

import (
	"slices"
	"testing"
)

func TestFactorial(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{-1, 0},
		{0, 1},
		{1, 1},
		{2, 2},
		{5, 120},
		{10, 3_628_800},
		{20, 2_432_902_008_176_640_000},
	}
	for _, tt := range tests {
		if got := Factorial(tt.n); got != tt.want {
			t.Errorf("Factorial(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestPermutations(t *testing.T) {
	tests := []struct {
		s    string
		want []string
	}{
		{"", []string{""}},
		{"a", []string{"a"}},
		{"ab", []string{"ab", "ba"}},
		{"ba", []string{"ab", "ba"}},
		{"abc", []string{"abc", "acb", "bac", "bca", "cab", "cba"}},
		{"aab", []string{"aab", "aba", "baa"}},
		{"aaa", []string{"aaa"}},
		{"mat", []string{"amt", "atm", "mat", "mta", "tam", "tma"}},
	}
	for _, tt := range tests {
		if got := Permutations(tt.s); !slices.Equal(got, tt.want) {
			t.Errorf("Permutations(%q)\n got: %q\nwant: %q", tt.s, got, tt.want)
		}
	}
	got := Permutations("nairobi") // 7 letters, i twice: 7!/2 = 2520
	if len(got) != 2520 || got[0] != "abiinor" || got[len(got)-1] != "roniiba" {
		t.Errorf("Permutations(\"nairobi\"): %d results from %q to %q, want 2520 from \"abiinor\" to \"roniiba\"", len(got), got[0], got[len(got)-1])
	}
}

func TestCanPay(t *testing.T) {
	pocket := []int{1000, 500, 200, 50}
	tests := []struct {
		notes  []int
		amount int
		want   bool
	}{
		{nil, 0, true},
		{nil, 50, false},
		{pocket, 0, true},
		{pocket, 750, true},
		{pocket, 300, false},
		{pocket, 1750, true},
		{pocket, 1800, false},
		{pocket, 2000, false}, // can't use the 1000 twice
		{[]int{200, 200, 200}, 400, true},
		{[]int{50}, 50, true},
	}
	for _, tt := range tests {
		if got := CanPay(tt.notes, tt.amount); got != tt.want {
			t.Errorf("CanPay(%v, %d) = %v, want %v", tt.notes, tt.amount, got, tt.want)
		}
	}
	// 22 notes: 2^22 ≈ 4 million choices to try when nothing works. Still fast.
	many := make([]int, 22)
	for i := range many {
		many[i] = 100
	}
	if CanPay(many, 150) {
		t.Errorf("CanPay(22 × 100, 150) = true, want false")
	}
	if !CanPay(many, 2200) {
		t.Errorf("CanPay(22 × 100, 2200) = false, want true")
	}
}

func TestHanoi(t *testing.T) {
	if got := Hanoi(0, "A", "B", "C"); len(got) != 0 {
		t.Errorf("Hanoi(0) = %v, want no moves", got)
	}
	if got, want := Hanoi(1, "A", "B", "C"), []Move{{"A", "C"}}; !slices.Equal(got, want) {
		t.Errorf("Hanoi(1) = %v, want %v", got, want)
	}
	if got, want := Hanoi(2, "A", "B", "C"), []Move{{"A", "B"}, {"A", "C"}, {"B", "C"}}; !slices.Equal(got, want) {
		t.Errorf("Hanoi(2) = %v, want %v", got, want)
	}
	want3 := []Move{{"A", "C"}, {"A", "B"}, {"C", "B"}, {"A", "C"}, {"B", "A"}, {"B", "C"}, {"A", "C"}}
	if got := Hanoi(3, "A", "B", "C"); !slices.Equal(got, want3) {
		t.Errorf("Hanoi(3) = %v, want %v", got, want3)
	}
	// Play 10 discs and check every move is legal and they all end on C.
	pegs := map[string][]int{"A": {10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, "B": nil, "C": nil}
	moves := Hanoi(10, "A", "B", "C")
	if len(moves) != 1023 {
		t.Fatalf("Hanoi(10) made %d moves, want 1023", len(moves))
	}
	for i, m := range moves {
		from, to := pegs[m.From], pegs[m.To]
		if len(from) == 0 {
			t.Fatalf("move %d %v: peg %s is empty", i+1, m, m.From)
		}
		disc := from[len(from)-1]
		if len(to) > 0 && to[len(to)-1] < disc {
			t.Fatalf("move %d %v: disc %d onto smaller disc %d", i+1, m, disc, to[len(to)-1])
		}
		pegs[m.From], pegs[m.To] = from[:len(from)-1], append(to, disc)
	}
	if len(pegs["C"]) != 10 {
		t.Errorf("after Hanoi(10), peg C has %d discs, want 10", len(pegs["C"]))
	}
}
