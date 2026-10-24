package twopointers

import (
	"slices"
	"strings"
	"testing"
)

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"", true},
		{"a", true},
		{"Kayak", true},
		{"Was it a rat I saw?", true},
		{"KCA 121 ACK", true}, // a number plate, read both ways
		{"!!!", true},
		{"ab", false},
		{"Nairobi", false},
		{"race a car", false},
		{"0P", false},
		{"Step on no pets", true},
		{strings.Repeat("ab", 50_000) + strings.Repeat("ba", 50_000), true},
	}
	for _, tt := range tests {
		if got := IsPalindrome(tt.s); got != tt.want {
			name := tt.s
			if len(name) > 30 {
				name = name[:30] + "..."
			}
			t.Errorf("IsPalindrome(%q) = %v, want %v", name, got, tt.want)
		}
	}
}

func TestPairWithSum(t *testing.T) {
	tests := []struct {
		name   string
		sorted []int
		target int
		wantOK bool
	}{
		{"empty", nil, 10, false},
		{"one value", []int{5}, 10, false},
		{"ends", []int{1, 3, 5, 9}, 10, true},
		{"middle", []int{50, 100, 200, 450, 800}, 650, true},
		{"duplicates", []int{500, 500}, 1000, true},
		{"negatives", []int{-300, -100, 0, 250, 700}, -400, true},
		{"none", []int{1, 2, 4, 8}, 20, false},
		{"not by using one value twice", []int{1, 5, 9}, 10, true},
		{"can't reuse a single 5", []int{1, 5, 11}, 10, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, j, ok := PairWithSum(tt.sorted, tt.target)
			if ok != tt.wantOK {
				t.Fatalf("PairWithSum(%v, %d) ok = %v, want %v", tt.sorted, tt.target, ok, tt.wantOK)
			}
			if ok && !(0 <= i && i < j && j < len(tt.sorted) && tt.sorted[i]+tt.sorted[j] == tt.target) {
				t.Errorf("PairWithSum(%v, %d) = %d, %d: not two positions i < j that add up to %d", tt.sorted, tt.target, i, j, tt.target)
			}
		})
	}
}

func TestPairWithSumBig(t *testing.T) {
	a := make([]int, 200_000)
	for i := range a {
		a[i] = 2 * i
	}
	if _, _, ok := PairWithSum(a, 3); ok {
		t.Errorf("PairWithSum(200000 even numbers, 3) found a pair, want none")
	}
	i, j, ok := PairWithSum(a, 2*199_999+2*199_998)
	if !ok || a[i]+a[j] != 2*199_999+2*199_998 || i >= j {
		t.Errorf("PairWithSum(big, the two largest) = %d, %d, %v", i, j, ok)
	}
}

func TestMergeSorted(t *testing.T) {
	tests := []struct {
		a, b, want []int
	}{
		{nil, nil, []int{}},
		{[]int{1, 2}, nil, []int{1, 2}},
		{nil, []int{3}, []int{3}},
		{[]int{1, 4, 9}, []int{2, 4, 10}, []int{1, 2, 4, 4, 9, 10}},
		{[]int{5, 6, 7}, []int{1, 2}, []int{1, 2, 5, 6, 7}},
		{[]int{-5, 0}, []int{-7, -5, 3}, []int{-7, -5, -5, 0, 3}},
	}
	for _, tt := range tests {
		a, b := slices.Clone(tt.a), slices.Clone(tt.b)
		got := MergeSorted(a, b)
		if got == nil || !slices.Equal(got, tt.want) {
			t.Errorf("MergeSorted(%v, %v)\n got: %#v\nwant: %v", tt.a, tt.b, got, tt.want)
		}
		if !slices.Equal(a, tt.a) || !slices.Equal(b, tt.b) {
			t.Errorf("MergeSorted changed its inputs: %v %v, want %v %v", a, b, tt.a, tt.b)
		}
	}
}

func TestDedup(t *testing.T) {
	tests := []struct {
		in, want []int
	}{
		{[]int{}, []int{}},
		{[]int{7}, []int{7}},
		{[]int{20, 20, 50, 50, 50, 100}, []int{20, 50, 100}},
		{[]int{1, 2, 3}, []int{1, 2, 3}},
		{[]int{4, 4, 4, 4}, []int{4}},
		{[]int{-3, -3, 0, 0, 2}, []int{-3, 0, 2}},
	}
	for _, tt := range tests {
		in := slices.Clone(tt.in)
		got := Dedup(in)
		if !slices.Equal(got, tt.want) {
			t.Errorf("Dedup(%v) = %v, want %v", tt.in, got, tt.want)
		}
		if len(got) > 0 && &got[0] != &in[0] {
			t.Errorf("Dedup(%v) made a new slice; it should work in place", tt.in)
		}
	}
}
