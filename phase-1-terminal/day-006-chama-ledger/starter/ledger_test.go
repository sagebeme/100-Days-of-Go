package chamaledger

import (
	"slices"
	"testing"
)

func TestTotal(t *testing.T) {
	tests := []struct {
		contributions []int
		want          int
	}{
		{[]int{1000, 1000, 500}, 2500},
		{[]int{750}, 750},
		{[]int{0, 0, 0}, 0},
		{[]int{1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000}, 12000},
		{[]int{}, 0},
		{nil, 0},
	}
	for _, tt := range tests {
		got := Total(tt.contributions)
		if got != tt.want {
			t.Errorf("Total(%v)\n got: %d\nwant: %d", tt.contributions, got, tt.want)
		}
	}
}

func TestRunningBalance(t *testing.T) {
	tests := []struct {
		contributions []int
		want          []int
	}{
		{[]int{500, 1000, 250}, []int{500, 1500, 1750}},
		{[]int{1000}, []int{1000}},
		{[]int{0, 200, 0, 300}, []int{0, 200, 200, 500}},
		{[]int{}, []int{}},
		{nil, nil},
	}
	for _, tt := range tests {
		before := slices.Clone(tt.contributions)
		got := RunningBalance(tt.contributions)
		if len(got) != len(tt.want) || !slices.Equal(got, tt.want) {
			t.Errorf("RunningBalance(%v)\n got: %v\nwant: %v", before, got, tt.want)
		}
		if !slices.Equal(tt.contributions, before) {
			t.Errorf("RunningBalance changed its input: it was %v, now it's %v", before, tt.contributions)
		}
	}
}

func TestDefaulters(t *testing.T) {
	tests := []struct {
		names []string
		paid  []int
		due   int
		want  []string
	}{
		{[]string{"Njeri", "Atieno", "Wafula"}, []int{1000, 600, 1000}, 1000, []string{"Atieno"}},
		{[]string{"Njeri", "Atieno", "Wafula"}, []int{0, 999, 1000}, 1000, []string{"Njeri", "Atieno"}},
		{[]string{"Kamau", "Chebet"}, []int{1500, 1000}, 1000, nil},
		{[]string{"Kamau", "Chebet", "Odhiambo"}, []int{10, 20, 30}, 500, []string{"Kamau", "Chebet", "Odhiambo"}},
		{[]string{}, []int{}, 1000, nil},
	}
	for _, tt := range tests {
		got := Defaulters(tt.names, tt.paid, tt.due)
		if !slices.Equal(got, tt.want) {
			t.Errorf("Defaulters(%v, %v, %d)\n got: %v\nwant: %v", tt.names, tt.paid, tt.due, got, tt.want)
		}
	}
}
