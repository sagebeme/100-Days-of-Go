package stacks

import (
	"slices"
	"strings"
	"testing"
)

func TestStack(t *testing.T) {
	var s Stack[string]
	if _, ok := s.Pop(); ok {
		t.Fatalf("Pop on an empty stack returned ok = true")
	}
	if _, ok := s.Peek(); ok {
		t.Fatalf("Peek on an empty stack returned ok = true")
	}
	for _, v := range []string{"ugali", "sukuma", "nyama"} {
		s.Push(v)
	}
	if s.Len() != 3 {
		t.Fatalf("Len after 3 pushes = %d, want 3", s.Len())
	}
	if v, ok := s.Peek(); v != "nyama" || !ok {
		t.Errorf("Peek = %q, %v, want %q, true", v, ok, "nyama")
	}
	if s.Len() != 3 {
		t.Errorf("Peek changed the stack: Len = %d, want 3", s.Len())
	}
	for _, want := range []string{"nyama", "sukuma", "ugali"} {
		if v, ok := s.Pop(); v != want || !ok {
			t.Errorf("Pop = %q, %v, want %q, true", v, ok, want)
		}
	}
	if v, ok := s.Pop(); v != "" || ok {
		t.Errorf("Pop on an emptied stack = %q, %v, want \"\", false", v, ok)
	}

	var nums Stack[int] // the same code works for any type
	for i := range 100_000 {
		nums.Push(i)
	}
	if v, _ := nums.Pop(); v != 99_999 || nums.Len() != 99_999 {
		t.Errorf("after 100000 pushes and a pop: got %d with Len %d", v, nums.Len())
	}
}

func TestBalanced(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"", true},
		{"no brackets at all", true},
		{"()", true},
		{"([]{})", true},
		{"fare(route[46K]{stage})", true},
		{"(", false},
		{")", false},
		{"([)]", false},
		{"(()", false},
		{"())(", false},
		{"}{", false},
		{"[{()}]", true},
		{strings.Repeat("(", 100_000) + strings.Repeat(")", 100_000), true},
		{strings.Repeat("(", 100_000) + strings.Repeat(")", 99_999), false},
	}
	for _, tt := range tests {
		if got := Balanced(tt.s); got != tt.want {
			name := tt.s
			if len(name) > 30 {
				name = name[:30] + "..."
			}
			t.Errorf("Balanced(%q) = %v, want %v", name, got, tt.want)
		}
	}
}

func TestEditor(t *testing.T) {
	var e Editor
	check := func(step, want string) {
		t.Helper()
		if got := e.Text(); got != want {
			t.Errorf("after %s: Text() = %q, want %q", step, got, want)
		}
	}
	e.Undo()
	check("Undo on a new editor", "")
	e.Type("Niko")
	check(`Type("Niko")`, "Niko")
	e.Type(" stage")
	check(`Type(" stage")`, "Niko stage")
	e.Delete(5)
	check("Delete(5)", "Niko ")
	e.Type("kwa mat")
	check(`Type("kwa mat")`, "Niko kwa mat")
	e.Undo()
	check("Undo", "Niko ")
	e.Undo()
	check("Undo", "Niko stage")
	e.Delete(100)
	check("Delete(100)", "")
	e.Undo()
	check("Undo", "Niko stage")
	e.Undo()
	e.Undo()
	check("two more Undos", "")
	e.Undo()
	check("Undo with nothing left", "")
}

func TestNextWarmer(t *testing.T) {
	tests := []struct {
		temps, want []int
	}{
		{[]int{}, []int{}},
		{[]int{25}, []int{0}},
		{[]int{22, 24, 21, 20, 25}, []int{1, 3, 2, 1, 0}},
		{[]int{20, 20, 20}, []int{0, 0, 0}},
		{[]int{18, 19, 20, 21}, []int{1, 1, 1, 0}},
		{[]int{30, 25, 20, 15}, []int{0, 0, 0, 0}},
		{[]int{-2, -5, 0}, []int{2, 1, 0}},
	}
	for _, tt := range tests {
		got := NextWarmer(tt.temps)
		if !slices.Equal(got, tt.want) {
			t.Errorf("NextWarmer(%v) = %v, want %v", tt.temps, got, tt.want)
		}
	}
}

func TestNextWarmerBig(t *testing.T) {
	// Falling for 200 000 days, then one hot day. Checking every later day is O(n²) here.
	n := 200_000
	temps := make([]int, n)
	for i := range temps {
		temps[i] = n - i
	}
	temps[n-1] = n + 1
	got := NextWarmer(temps)
	if got[0] != n-1 || got[n-2] != 1 || got[n-1] != 0 {
		t.Errorf("NextWarmer(big): got[0] = %d, got[n-2] = %d, got[n-1] = %d, want %d, 1, 0", got[0], got[n-2], got[n-1], n-1)
	}
}
