// Package greedy makes the best-looking choice at each step, and proves it's safe.
package greedy

import (
	"cmp"
	"slices"
)

// Denominations are Kenya's notes and coins in shillings, largest first.
var Denominations = []int{1000, 500, 200, 100, 50, 20, 10, 5, 1}

// FewestNotes takes as many of the biggest note as fit, then the next, and so on.
// That's optimal for Kenya's denominations, though not for every set of coins.
func FewestNotes(amount int) map[int]int {
	counts := map[int]int{}
	for _, d := range Denominations {
		if amount >= d {
			counts[d] = amount / d
			amount %= d
		}
	}
	return counts
}

// Slot is a booking of the community hall from Start up to, but not including, End.
type Slot struct {
	Start, End int
}

// MaxBookings keeps the slot that ends first, then the next one that starts
// after it ends, and so on.
func MaxBookings(slots []Slot) int {
	byEnd := slices.Clone(slots)
	slices.SortFunc(byEnd, func(a, b Slot) int { return cmp.Compare(a.End, b.End) })
	count, free := 0, -1<<62 // free: the hour the hall is next free
	for _, s := range byEnd {
		if s.Start >= free {
			count++
			free = s.End
		}
	}
	return count
}

// FewestBodas pairs the heaviest passenger left with the lightest, if they fit.
func FewestBodas(weights []int, limit int) int {
	w := slices.Sorted(slices.Values(weights))
	trips := 0
	i, j := 0, len(w)-1
	for i <= j {
		if i < j && w[i]+w[j] <= limit {
			i++ // the lightest rides along
		}
		j-- // the heaviest always goes
		trips++
	}
	return trips
}
