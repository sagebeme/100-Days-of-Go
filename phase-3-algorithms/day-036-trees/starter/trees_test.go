package trees

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func build(keys ...int) *Node {
	var root *Node
	for _, k := range keys {
		root = Insert(root, k)
	}
	return root
}

func TestInsertShape(t *testing.T) {
	//        50
	//      /    \
	//    30      70
	//   /  \       \
	//  20   40      80
	root := build(50, 30, 70, 20, 40, 80, 30) // the second 30 is ignored
	if root == nil || root.Key != 50 {
		t.Fatalf("root = %v, want key 50", root)
	}
	if root.Left == nil || root.Left.Key != 30 || root.Right == nil || root.Right.Key != 70 {
		t.Fatalf("the root's children are wrong: want 30 on the left and 70 on the right")
	}
	if root.Left.Left == nil || root.Left.Left.Key != 20 || root.Left.Right == nil || root.Left.Right.Key != 40 {
		t.Errorf("30's children are wrong: want 20 and 40")
	}
	if root.Right.Left != nil || root.Right.Right == nil || root.Right.Right.Key != 80 {
		t.Errorf("70's children are wrong: want nothing on the left and 80 on the right")
	}
	if root.Left.Left.Left != nil || root.Left.Left.Right != nil {
		t.Errorf("the duplicate 30 was added to the tree")
	}
}

func TestContains(t *testing.T) {
	root := build(50, 30, 70, 20, 40, 80)
	for _, k := range []int{50, 30, 70, 20, 40, 80} {
		if !Contains(root, k) {
			t.Errorf("Contains(tree, %d) = false, want true", k)
		}
	}
	for _, k := range []int{0, 25, 45, 60, 75, 90, -50} {
		if Contains(root, k) {
			t.Errorf("Contains(tree, %d) = true, want false", k)
		}
	}
	if Contains(nil, 5) {
		t.Errorf("Contains(nil, 5) = true, want false")
	}
}

func TestInOrder(t *testing.T) {
	tests := []struct {
		name string
		keys []int
		want []int
	}{
		{"empty", nil, []int{}},
		{"one", []int{7}, []int{7}},
		{"mixed", []int{50, 30, 70, 20, 40, 80}, []int{20, 30, 40, 50, 70, 80}},
		{"duplicates and negatives", []int{0, -5, 5, -5, 0, 10}, []int{-5, 0, 5, 10}},
		{"sorted input", []int{1, 2, 3, 4}, []int{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InOrder(build(tt.keys...))
			if got == nil || !slices.Equal(got, tt.want) {
				t.Errorf("InOrder(tree of %v) = %#v, want %v", tt.keys, got, tt.want)
			}
		})
	}
}

func TestInOrderSortsRandomKeys(t *testing.T) {
	r := rand.New(rand.NewPCG(36, 36))
	keys := make([]int, 50_000)
	for i := range keys {
		keys[i] = r.IntN(1_000_000)
	}
	want := slices.Compact(slices.Sorted(slices.Values(keys)))
	if got := InOrder(build(keys...)); !slices.Equal(got, want) {
		t.Errorf("InOrder of 50000 random keys isn't the sorted distinct keys")
	}
}

func TestLevelOrder(t *testing.T) {
	tests := []struct {
		keys []int
		want [][]int
	}{
		{nil, [][]int{}},
		{[]int{9}, [][]int{{9}}},
		{[]int{50, 30, 70}, [][]int{{50}, {30, 70}}},
		{[]int{50, 30, 70, 20, 40, 80}, [][]int{{50}, {30, 70}, {20, 40, 80}}},
		{[]int{1, 2, 3}, [][]int{{1}, {2}, {3}}},
	}
	for _, tt := range tests {
		got := LevelOrder(build(tt.keys...))
		if got == nil || !slices.EqualFunc(got, tt.want, slices.Equal[[]int]) {
			t.Errorf("LevelOrder(tree of %v) = %v, want %v", tt.keys, got, tt.want)
		}
	}
}

func TestHeight(t *testing.T) {
	tests := []struct {
		keys []int
		want int
	}{
		{nil, 0},
		{[]int{1}, 1},
		{[]int{50, 30, 70}, 2},
		{[]int{50, 30, 70, 20, 40, 80}, 3},
		{[]int{50, 30, 20, 10}, 4},
		{[]int{4, 2, 6, 1, 3, 5, 7}, 3},
	}
	for _, tt := range tests {
		if got := Height(build(tt.keys...)); got != tt.want {
			t.Errorf("Height(tree of %v) = %d, want %d", tt.keys, got, tt.want)
		}
	}
}

func TestSortedInsertsMakeALine(t *testing.T) {
	// Members joining in number order: every new key goes to the right.
	// The tree is a linked list, so its height is n, not log n.
	keys := make([]int, 2000)
	for i := range keys {
		keys[i] = i
	}
	if got := Height(build(keys...)); got != 2000 {
		t.Errorf("Height after 2000 sorted inserts = %d, want 2000", got)
	}
}
