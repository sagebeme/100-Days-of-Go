// Package recursion solves problems by calling a function on a smaller version of itself.
package recursion

import "slices"

// Factorial returns n!, recursively.
func Factorial(n int) int {
	if n < 0 {
		return 0
	}
	if n <= 1 { // base case: 0! = 1! = 1
		return 1
	}
	return n * Factorial(n-1)
}

// Permutations returns every distinct ordering of s, in alphabetical order.
func Permutations(s string) []string {
	letters := []byte(s)
	slices.Sort(letters) // sorted letters give the answers in alphabetical order
	used := make([]bool, len(letters))
	path := make([]byte, 0, len(letters))
	var out []string

	var build func()
	build = func() {
		if len(path) == len(letters) { // base case: a full ordering
			out = append(out, string(path))
			return
		}
		for i, c := range letters {
			if used[i] {
				continue
			}
			// Skip a letter equal to the one before it unless that one is in use:
			// otherwise "aab" would start with each 'a' and give "aab" twice.
			if i > 0 && c == letters[i-1] && !used[i-1] {
				continue
			}
			used[i] = true
			path = append(path, c)
			build()
			path = path[:len(path)-1] // undo the choice: backtrack
			used[i] = false
		}
	}
	build()
	return out
}

// CanPay reports whether some of the notes add up to exactly amount.
func CanPay(notes []int, amount int) bool {
	if amount == 0 {
		return true
	}
	if len(notes) == 0 || amount < 0 {
		return false
	}
	// Either the first note is used, or it isn't.
	return CanPay(notes[1:], amount-notes[0]) || CanPay(notes[1:], amount)
}

// Move is one move in the Tower of Hanoi: the top disc from peg From to peg To.
type Move struct {
	From, To string
}

// Hanoi returns the moves that take n discs from one peg to another.
func Hanoi(n int, from, via, to string) []Move {
	if n == 0 {
		return nil
	}
	// Move the n-1 smaller discs out of the way, move the biggest, then put
	// the smaller ones back on top of it.
	moves := Hanoi(n-1, from, to, via)
	moves = append(moves, Move{from, to})
	return append(moves, Hanoi(n-1, via, from, to)...)
}
