package greedy

import (
	"maps"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestFewestNotes(t *testing.T) {
	tests := []struct {
		amount int
		want   map[int]int
	}{
		{0, map[int]int{}},
		{-50, map[int]int{}},
		{1, map[int]int{1: 1}},
		{4, map[int]int{1: 4}},
		{50, map[int]int{50: 1}},
		{1780, map[int]int{1000: 1, 500: 1, 200: 1, 50: 1, 20: 1, 10: 1}},
		{2999, map[int]int{1000: 2, 500: 1, 200: 2, 50: 1, 20: 2, 5: 1, 1: 4}},
		{1_000_000, map[int]int{1000: 1000}},
	}
	for _, tt := range tests {
		got := FewestNotes(tt.amount)
		if got == nil || !maps.Equal(got, tt.want) {
			t.Errorf("FewestNotes(%d)\n got: %v\nwant: %v", tt.amount, got, tt.want)
		}
	}
}

func TestMaxBookings(t *testing.T) {
	tests := []struct {
		name  string
		slots []Slot
		want  int
	}{
		{"none", nil, 0},
		{"one", []Slot{{8, 17}}, 1},
		{"back to back", []Slot{{9, 10}, {10, 11}, {11, 12}}, 3},
		{"the long one loses", []Slot{{8, 20}, {9, 10}, {10, 11}, {11, 12}}, 3},
		{"overlaps", []Slot{{9, 12}, {10, 11}, {11, 13}, {12, 14}}, 2},
		{"unsorted", []Slot{{14, 16}, {9, 11}, {10, 15}, {11, 14}}, 3},
		{"all the same", []Slot{{9, 11}, {9, 11}, {9, 11}}, 1},
		{"starts first is wrong", []Slot{{1, 10}, {2, 3}, {4, 5}, {6, 7}}, 3},
		{"shortest first is wrong", []Slot{{1, 5}, {4, 7}, {6, 10}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := slices.Clone(tt.slots)
			if got := MaxBookings(tt.slots); got != tt.want {
				t.Errorf("MaxBookings(%v) = %d, want %d", tt.slots, got, tt.want)
			}
			if !slices.Equal(before, tt.slots) {
				t.Errorf("MaxBookings changed its input to %v", tt.slots)
			}
		})
	}
}

func TestMaxBookingsBig(t *testing.T) {
	r := rand.New(rand.NewPCG(38, 38))
	slots := make([]Slot, 200_000)
	for i := range slots {
		slots[i] = Slot{2 * i, 2*i + 1} // 200 000 slots that never overlap
	}
	r.Shuffle(len(slots), func(i, j int) { slots[i], slots[j] = slots[j], slots[i] })
	if got := MaxBookings(slots); got != 200_000 {
		t.Errorf("MaxBookings(200000 separate slots) = %d, want 200000", got)
	}
}

func TestFewestBodas(t *testing.T) {
	tests := []struct {
		weights []int
		limit   int
		want    int
	}{
		{nil, 100, 0},
		{[]int{60}, 100, 1},
		{[]int{70, 50, 80, 50}, 130, 2},
		{[]int{50, 50, 50, 50}, 100, 2},
		{[]int{100, 100, 100}, 100, 3},
		{[]int{30, 30, 30}, 100, 2}, // never three on one boda
		{[]int{40, 60, 90, 10}, 100, 2},
		{[]int{90, 85, 20, 15}, 100, 3}, // 90 can't take 15: 85+15, then 90 and 20 alone
		{[]int{95, 10, 50, 50, 5}, 100, 3},
	}
	for _, tt := range tests {
		before := slices.Clone(tt.weights)
		if got := FewestBodas(tt.weights, tt.limit); got != tt.want {
			t.Errorf("FewestBodas(%v, %d) = %d, want %d", tt.weights, tt.limit, got, tt.want)
		}
		if !slices.Equal(before, tt.weights) {
			t.Errorf("FewestBodas changed its input to %v", tt.weights)
		}
	}
}

func TestFewestBodasBig(t *testing.T) {
	w := make([]int, 200_000)
	for i := range w {
		w[i] = 1 + i%100 // 1..100 kg, a limit of 101 pairs everyone up
	}
	if got := FewestBodas(w, 101); got != 100_000 {
		t.Errorf("FewestBodas(big, 101) = %d, want 100000", got)
	}
}
