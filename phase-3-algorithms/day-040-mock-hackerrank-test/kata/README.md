# Kata (Problem 7): Jam Session

Friday, 5 p.m., and you have to drive across a city grid from the top-left block to the bottom-right one. Each block is a digit from `1` to `9`: the minutes it takes to drive into that block in this traffic. You can move up, down, left or right, one block at a time. Driving into a block costs its minutes. The starting block costs nothing, because you're already there.

What is the fewest number of minutes to reach the bottom-right block?

The city is made up for this kata.

**Input**

- Line 1: `R` and `C` (1 ≤ R, C ≤ 1000)
- The next `R` lines: `C` digits each, from `1` to `9`, with no spaces

**Output**

One line: the fewest minutes from the top-left block to the bottom-right block. If the grid is one block, that's 0.

**Example**

```
Input           Output
3 4             13
1163
1381
2136
```

One best route goes down, down, right, right, right: 1 + 2 + 1 + 3 + 6 = 13.

**Limits:** 2 seconds per test case.

**Read the constraints.** There are up to 10^6 blocks.

- BFS (Day 37) counts steps, not minutes, so it gives wrong answers when costs differ.
- A DP that only moves right and down is wrong too: sometimes the fast road winds back left or up around a jam. One test is built to catch exactly that.
- Relaxing every block again and again until nothing changes is correct but can take O(V²).

Dijkstra's algorithm is O(V log V) here: keep a priority queue (a heap, Day 34) of blocks by their best known time. Pop the smallest. Its time is final, because every cost is positive. Try its four neighbours and push any you've improved.

**Go tip**

Use `container/heap` with an `item{time, cell int}` and number cells `r*C + c`. You don't need to delete or update entries in the heap. Push a new entry when you improve a block, and when you pop an entry whose time is worse than the block's best, skip it.

Judge it from the repo root with `go run ./tools/kata 40`.
