// Package trees builds a binary search tree and walks it.
package trees

// Node is one node of a binary search tree.
type Node struct {
	Key         int
	Left, Right *Node
}

// Insert adds key to the tree and returns the root.
func Insert(root *Node, key int) *Node {
	if root == nil {
		return &Node{Key: key}
	}
	switch {
	case key < root.Key:
		root.Left = Insert(root.Left, key)
	case key > root.Key:
		root.Right = Insert(root.Right, key)
	}
	return root
}

// Contains reports whether key is in the tree.
func Contains(root *Node, key int) bool {
	for root != nil {
		switch {
		case key < root.Key:
			root = root.Left
		case key > root.Key:
			root = root.Right
		default:
			return true
		}
	}
	return false
}

// InOrder returns the keys in ascending order.
func InOrder(root *Node) []int {
	keys := []int{}
	var walk func(n *Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		walk(n.Left)
		keys = append(keys, n.Key)
		walk(n.Right)
	}
	walk(root)
	return keys
}

// LevelOrder returns the keys level by level, using a queue of the current level.
func LevelOrder(root *Node) [][]int {
	levels := [][]int{}
	if root == nil {
		return levels
	}
	level := []*Node{root}
	for len(level) > 0 {
		var keys []int
		var next []*Node
		for _, n := range level {
			keys = append(keys, n.Key)
			if n.Left != nil {
				next = append(next, n.Left)
			}
			if n.Right != nil {
				next = append(next, n.Right)
			}
		}
		levels = append(levels, keys)
		level = next
	}
	return levels
}

// Height returns the number of nodes on the longest root-to-leaf path.
func Height(root *Node) int {
	if root == nil {
		return 0
	}
	return 1 + max(Height(root.Left), Height(root.Right))
}
