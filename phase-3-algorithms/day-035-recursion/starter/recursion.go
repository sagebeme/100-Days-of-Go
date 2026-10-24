// Package recursion solves problems by calling a function on a smaller version of itself.
package recursion

// Factorial returns n! = n × (n-1) × ... × 1, with 0! = 1, for 0 <= n <= 20
// (21! doesn't fit in an int). It returns 0 for a negative n. Write it recursively.
//
//	Factorial(5) → 120
func Factorial(n int) int {
	// Your code here.
	return 0
}

// Permutations returns every ordering of the bytes in s, each one once (so
// repeated letters don't give repeated answers), in alphabetical order.
// Permutations("") is [""].
//
//	Permutations("abc") → [abc acb bac bca cab cba]
//	Permutations("aab") → [aab aba baa]
func Permutations(s string) []string {
	// Your code here.
	return nil
}

// CanPay reports whether some of the notes in your pocket, each used at most
// once, add up to exactly amount. amount 0 is always payable (use no notes).
// len(notes) is at most 25.
//
//	CanPay([]int{1000, 500, 200, 50}, 750) → true    500 + 200 + 50
//	CanPay([]int{1000, 500, 200, 50}, 300) → false
func CanPay(notes []int, amount int) bool {
	// Your code here.
	return false
}

// Move is one move in the Tower of Hanoi: the top disc from peg From to peg To.
type Move struct {
	From, To string
}

// Hanoi returns the shortest list of moves that takes n discs from peg from to
// peg to, using peg via as the spare. Only the top disc of a peg moves, and a
// disc never sits on a smaller one. It's 2^n - 1 moves. n = 0 gives no moves.
//
//	Hanoi(2, "A", "B", "C") → [{A B} {A C} {B C}]
func Hanoi(n int, from, via, to string) []Move {
	// Your code here.
	return nil
}
