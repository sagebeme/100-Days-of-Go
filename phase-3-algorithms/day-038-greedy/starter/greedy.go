// Package greedy makes the best-looking choice at each step, and proves it's safe.
package greedy

// Denominations are Kenya's notes and coins in shillings, largest first.
var Denominations = []int{1000, 500, 200, 100, 50, 20, 10, 5, 1}

// FewestNotes returns how many of each note and coin make up amount with the
// fewest pieces: denomination → count, listing only the ones used. An amount
// below 1 gives an empty (non-nil) map.
//
//	FewestNotes(1780) → map[10:1 20:1 50:1 200:1 500:1 1000:1]
func FewestNotes(amount int) map[int]int {
	// Your code here.
	return nil
}

// Slot is a booking of the community hall from Start up to, but not
// including, End (hours on a 24-hour clock). Start < End.
type Slot struct {
	Start, End int
}

// MaxBookings returns the largest number of slots the hall can host without
// two of them overlapping. A slot may start at the hour another one ends.
// The input order doesn't matter, and the slice must not be changed.
//
//	MaxBookings([]Slot{{9, 12}, {10, 11}, {11, 13}, {12, 14}}) → 2
func MaxBookings(slots []Slot) int {
	// Your code here.
	return 0
}

// FewestBodas returns the smallest number of boda trips that carry everyone.
// A boda takes one or two passengers whose total weight is at most limit kg.
// Every weight is between 1 and limit. The slice must not be changed.
//
//	FewestBodas([]int{70, 50, 80, 50}, 130) → 2    80+50 and 70+50
func FewestBodas(weights []int, limit int) int {
	// Your code here.
	return 0
}
