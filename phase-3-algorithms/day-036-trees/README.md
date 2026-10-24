# Day 36: Trees

## The brief

A SACCO in Meru keeps its members' numbers in a binary search tree. Each member is a node. Smaller numbers go to the left, bigger ones to the right. Looking a member up means walking down one path from the top, so with a million members in a balanced tree it's about 20 steps, not a million.

Today you'll build the tree, search it, walk it in sorted order and level by level, and measure its height. You'll also see its weak spot: members joining in number order turn the tree into a long line.

```
//        50
//      /    \
//    30      70
//   /  \       \
//  20   40      80

root := Insert(...)          // 50, 30, 70, 20, 40, 80
Contains(root, 40)           →  true
Contains(root, 45)           →  false
InOrder(root)                →  [20 30 40 50 70 80]
LevelOrder(root)             →  [[50] [30 70] [20 40 80]]
Height(root)                 →  3
```

## What you'll use

### Nodes and pointers

```go
type Node struct {
	Key         int
	Left, Right *Node
}
```

A tree is a node plus two smaller trees, which is why recursion (Day 35) fits it so well. `nil` is the empty tree and the base case of nearly every tree function:

```go
func Height(n *Node) int {
	if n == nil {
		return 0
	}
	return 1 + max(Height(n.Left), Height(n.Right))
}
```

`Insert` returns the root. That handles the empty tree: `Insert(nil, 5)` returns a new node. Inside, `root.Left = Insert(root.Left, key)` reattaches whatever the call returns.

### The traversals

| Order | Visits | Gives you |
|---|---|---|
| in-order | left, node, right | a BST's keys in sorted order |
| pre-order | node, left, right | a copy of the shape: the kata reads one |
| post-order | left, right, node | children before parents: heights, sizes, deleting |
| level order | one level at a time, top down | breadth-first search, with a queue (Day 34) |

### How fast is a BST?

Search, insert and delete each walk one path: O(height). A balanced tree has height about log₂ n. But insert sorted keys and every new key goes right, so the tree is a line of height n and every operation is O(n). One test inserts 2000 sorted keys and checks the height is 2000. Real libraries use self-balancing trees (red-black, AVL) to avoid this. Go's standard library has no tree; a sorted slice with binary search (Day 32) or a map usually does the job.

### Spotting it in a problem

- "binary tree", "root", "parent", "children", "leaf", "ancestor".
- "height", "depth", "diameter", "is it balanced": recurse, and return something from each subtree.
- "sorted order" from a BST: in-order.
- "level by level", "nearest", "minimum depth": level order.
- Values that must stay sorted while you insert: a BST, or a heap if you only need the smallest.

### Brute force vs optimal

| Problem | Brute force | Better |
|---|---|---|
| look up a key | scan every node: O(n) | walk one path: O(height), O(log n) when balanced |
| height of every subtree | recompute each one: O(n²) on a deep tree | one post-order walk: O(n) |
| tree diameter (kata) | BFS from every node: O(n²) | one walk returning heights: O(n) |

See the tree row in [reference/big-o.md](../../reference/big-o.md#patterns-and-their-complexity).

## Steps

1. Open `starter/trees.go`. Write `Insert` recursively: a new node for `nil`, otherwise recurse left or right and return `root`. Equal keys do nothing.
2. Write `Contains` with a loop instead of recursion: move `root` left or right until you find the key or reach `nil`.
3. Write `Height` recursively, as above.
4. Write `InOrder` with a recursive closure `walk` that appends to a `keys` slice. Start from `keys := []int{}` so an empty tree gives an empty slice, not `nil`.
5. Write `LevelOrder` with a slice of the current level's nodes. For each level, collect its keys and the next level's children, then move on.
6. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-036-trees/starter
```

## The kata: Distant Cousins

The two relatives furthest apart in a family tree, given as a preorder string. It's the tree diameter problem, a classic that rewards the "return one thing, record another" pattern of recursive tree code. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 36
```

## When you're stuck

- **`invalid memory address or nil pointer dereference`:** you read `root.Key` or `n.Left` without checking for `nil` first. Every tree function starts with the `nil` case.
- **`Insert` loses everything but the last key:** you returned the new node instead of `root`. Only return a new node when `root` is `nil`.
- **The tree is always one node:** you called `Insert(root.Left, key)` but didn't assign the result to `root.Left`.
- **`InOrder` returns `[]int(nil)` for an empty tree:** the test wants `[]int{}`. Start with `keys := []int{}`.
- **`LevelOrder` mixes levels together:** build the next level in a separate slice, and swap it in after the current level is done.
- **The kata answer is one too big:** the diameter counts links (edges), not people. A leaf has height 1 and its children 0, so `left + right` is already in links.
- **Still stuck?** Open `solution/trees.go`, read it once, close it, then write your own version from memory.
