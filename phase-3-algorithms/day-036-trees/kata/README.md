# Kata: Distant Cousins

At a family reunion in Murang'a, Shosho's family tree is written as one string. Each person has at most two children, a left one and a right one. The string lists the tree in **preorder**: `o` for a person, followed by their left child's whole family, then their right child's whole family. `.` marks a missing child.

```
ooo..o..oo...

        A              A = o, then A's left family, then A's right family
      /   \
     B     E           B = o, C's family "o..", D's family "o.."
    / \   /            E = o, F's family "o..", no right child "."
   C   D F
```

The distance between two relatives is the number of parent-child links on the path between them. Print the largest distance between any two people in the family.

**Input**

One line: the tree string, at most 2 × 10^5 + 1 characters. It describes at least one person and is a valid preorder: a tree with `n` people has exactly `2n + 1` characters.

**Output**

One line: the longest distance between two people. With one person, that's 0.

**Example**

```
Input           Output
ooo..o..oo...   4
```

C and F are furthest apart: C, B, A, E, F is 4 links.

**Limits:** 2 seconds per test case.

**Read the constraints.** The family can have 10^5 people, and the tree can be very deep. Working out the height of each person's family from scratch is O(n²) on a deep tree. Do it in one walk: one recursive function reads a subtree, returns its height, and on the way records `leftHeight + rightHeight`, the longest path that turns at this person. That's the tree's **diameter**, in O(n).

**Go tip**

You don't need to build `Node`s. Keep a position in the string and let the recursion read it:

```go
pos := 0
var height func() int
height = func() int {
	c := tree[pos]
	pos++
	if c == '.' {
		return 0
	}
	left, right := height(), height()
	...
}
```

The tree can be 10^5 levels deep. Go's stack grows to fit that, so plain recursion is fine here.

Judge it from the repo root with `go run ./tools/kata 36`.
