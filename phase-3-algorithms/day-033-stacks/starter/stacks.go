// Package stacks uses last-in, first-out stacks for brackets, undo and "next warmer day".
package stacks

// Stack is a last-in, first-out stack of any type. The zero value is an empty
// stack ready to use.
type Stack[T any] struct {
	items []T
}

// Push puts v on top.
func (s *Stack[T]) Push(v T) {
	// Your code here.
}

// Pop removes and returns the top value and true, or the zero value and false
// when the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	// Your code here.
	var zero T
	return zero, false
}

// Peek returns the top value without removing it, like Pop.
func (s *Stack[T]) Peek() (T, bool) {
	// Your code here.
	var zero T
	return zero, false
}

// Len returns how many values are on the stack.
func (s *Stack[T]) Len() int {
	// Your code here.
	return 0
}

// Balanced reports whether every (, [ and { in s is closed by the matching
// bracket in the right order. Other characters are ignored.
//
//	Balanced("fare(route[46K]{stage})") → true
//	Balanced("([)]")                    → false
func Balanced(s string) bool {
	// Your code here.
	return false
}

// Editor is a text box with undo. The zero value is an empty editor.
type Editor struct {
	text    string
	history Stack[string]
}

// Type adds s to the end of the text.
func (e *Editor) Type(s string) {
	// Your code here.
}

// Delete removes the last k bytes of the text (all of it if k is bigger).
func (e *Editor) Delete(k int) {
	// Your code here.
}

// Undo reverses the most recent Type or Delete that hasn't been undone yet.
// With nothing to undo, it does nothing.
func (e *Editor) Undo() {
	// Your code here.
}

// Text returns the current text.
func (e *Editor) Text() string {
	// Your code here.
	return ""
}

// NextWarmer returns, for each day, how many days you wait for a warmer day,
// or 0 if no warmer day comes. It must be O(n).
//
//	NextWarmer([]int{22, 24, 21, 20, 25}) → [1 3 2 1 0]
func NextWarmer(temps []int) []int {
	// Your code here.
	return nil
}
