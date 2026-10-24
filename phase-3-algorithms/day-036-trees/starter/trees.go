// Package trees builds a binary search tree and walks it.
package trees

// Node is one node of a binary search tree. Every key in Left is smaller than
// Key, and every key in Right is bigger. A nil *Node is an empty tree.
type Node struct {
	Key         int
	Left, Right *Node
}

// Insert adds key to the tree and returns the root (a new node if root is nil).
// A key that's already there is ignored.
//
//	var root *Node
//	root = Insert(root, 5040) // member numbers, in the order members joined
func Insert(root *Node, key int) *Node {
	// Your code here.
	return nil
}

// Contains reports whether key is in the tree. It walks one path from the
// root, so it's O(height).
func Contains(root *Node, key int) bool {
	// Your code here.
	return false
}

// InOrder returns the keys in ascending order: left subtree, node, right
// subtree. An empty tree gives an empty (non-nil) slice.
func InOrder(root *Node) []int {
	// Your code here.
	return nil
}

// LevelOrder returns the keys level by level, top first, each level left to
// right. An empty tree gives an empty (non-nil) slice.
//
//	tree with 50 at the root, 30 and 70 below it → [[50] [30 70]]
func LevelOrder(root *Node) [][]int {
	// Your code here.
	return nil
}

// Height returns the number of nodes on the longest path from the root down
// to a leaf. An empty tree has height 0 and a single node has height 1.
func Height(root *Node) int {
	// Your code here.
	return 0
}
