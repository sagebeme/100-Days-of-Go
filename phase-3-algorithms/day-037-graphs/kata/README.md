# Kata: Shortcut to the Stage

You're late for the last mat home. The town is a grid of blocks: `.` is a street you can walk, `#` is a building, `S` is where you are and `T` is the stage. Each step moves you one block up, down, left or right. What is the fewest number of steps to the stage?

The town is made up for this kata.

**Input**

- Line 1: `R` and `C`, the number of rows and columns (1 ≤ R, C ≤ 1000)
- The next `R` lines: `C` characters each, from `.`, `#`, and exactly one `S` and one `T`

**Output**

One line: the fewest steps from `S` to `T`, or `-1` if the stage can't be reached.

**Example**

```
Input           Output
4 8             8
S..#....
.#...##.
...#..T.
##......
```

**Limits:** 2 seconds per test case.

**Read the constraints.** A grid can have 10^6 cells. Trying every route with plain recursion explodes exponentially, and depth-first search with a visited set finds **a** route, not the shortest one. Breadth-first search visits cells in order of distance from `S`, so the first time it reaches `T` is the shortest way: O(R × C).

**Go tip**

Number each cell `r*C + c` so one `[]int` holds every distance, with `-1` for "not reached". Use a slice as the queue with a `head` index instead of `queue = queue[1:]`. Two small arrays give the four directions:

```go
dr := [4]int{-1, 1, 0, 0}
dc := [4]int{0, 0, -1, 1}
```

Mark a cell as reached when you add it to the queue, not when you take it out, or the same cell joins the queue many times.

Judge it from the repo root with `go run ./tools/kata 37`.
