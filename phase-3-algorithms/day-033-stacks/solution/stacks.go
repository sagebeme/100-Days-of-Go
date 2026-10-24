// Package stacks uses last-in, first-out stacks for brackets, undo and "next warmer day".
package stacks

// Stack is a last-in, first-out stack of any type.
type Stack[T any] struct {
	items []T
}

// Push puts v on top.
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop removes and returns the top value.
func (s *Stack[T]) Pop() (T, bool) {
	v, ok := s.Peek()
	if ok {
		s.items = s.items[:len(s.items)-1]
	}
	return v, ok
}

// Peek returns the top value without removing it.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Len returns how many values are on the stack.
func (s *Stack[T]) Len() int {
	return len(s.items)
}

// Balanced reports whether the brackets in s match.
func Balanced(s string) bool {
	pairs := map[byte]byte{')': '(', ']': '[', '}': '{'}
	var open Stack[byte]
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '(', '[', '{':
			open.Push(c)
		case ')', ']', '}':
			top, ok := open.Pop()
			if !ok || top != pairs[c] {
				return false
			}
		}
	}
	return open.Len() == 0
}

// Editor is a text box with undo.
type Editor struct {
	text    string
	history Stack[string] // the text before each change
}

// Type adds s to the end of the text.
func (e *Editor) Type(s string) {
	e.history.Push(e.text)
	e.text += s
}

// Delete removes the last k bytes of the text.
func (e *Editor) Delete(k int) {
	e.history.Push(e.text)
	k = min(k, len(e.text))
	e.text = e.text[:len(e.text)-k]
}

// Undo reverses the most recent change.
func (e *Editor) Undo() {
	if prev, ok := e.history.Pop(); ok {
		e.text = prev
	}
}

// Text returns the current text.
func (e *Editor) Text() string {
	return e.text
}

// NextWarmer returns how many days each day waits for a warmer one.
func NextWarmer(temps []int) []int {
	wait := make([]int, len(temps))
	var cold Stack[int] // indexes of days still waiting; their temps only go down the stack
	for i, t := range temps {
		for {
			j, ok := cold.Peek()
			if !ok || temps[j] >= t {
				break
			}
			cold.Pop()
			wait[j] = i - j // today is the first day warmer than day j
		}
		cold.Push(i)
	}
	return wait
}
