# Phase 3 recap: Algorithms for HackerRank

Days 27 to 40. You learnt to read a problem's limits, pick a pattern that fits them, and write it fast enough to beat a time limit, ending with a timed mock test.

## What the phase covered

| Days | Topic |
|---|---|
| 27 | Big-O, counting operations, fast stdin and stdout |
| 28 to 30 | hashing, sets and frequency counts; two pointers; sliding windows |
| 31 to 32 | merge sort, quicksort and custom comparators; binary search on data and on the answer |
| 33 to 34 | stacks (brackets, undo); queues and deques |
| 35 to 36 | recursion and base cases; binary search trees and traversals |
| 37 to 39 | BFS and DFS on graphs; greedy choices; dynamic programming, memoised and tabulated |
| 40 | the capstone: a mock HackerRank test, seven problems in 90 minutes |

## Key ideas

1. **The limits tell you the algorithm.** About 10^8 simple steps a second: n = 10^5 rules out O(n²) and asks for O(n log n) or O(n).
2. **Read and write with buffers.** `bufio.Reader` and `bufio.Writer` (with a `Flush`) are the difference between passing and timing out on big inputs.
3. **A map trades memory for time.** "Have I seen this?" and "how many of each?" become O(1) per item.
4. **Sorted input unlocks two pointers and binary search.** Sorting first costs O(n log n) and often turns an O(n²) search into O(n) or O(log n).
5. **Windows move, they don't restart.** Add the new item on the right, drop items from the left while the window breaks the rule.
6. **Every recursion needs a base case, and deep recursion needs a stack.** Half a million levels will overflow; use an explicit stack or a queue.
7. **BFS finds the fewest steps; Dijkstra, with a heap, the lowest total cost.** Mark visited nodes, and skip stale heap entries.
8. **Greedy is fast but must be proven; DP is safe when choices overlap.** If a greedy choice can be beaten by a counter-example, write the table.

## Self-test

1. Predict the output:

   ```go
   xs := []int{10, 20, 20, 30}
   fmt.Println(sort.SearchInts(xs, 20), sort.SearchInts(xs, 25), sort.SearchInts(xs, 99))
   ```

2. Coins of 1, 3 and 4. How many coins does greedy (biggest first) use for 6, and how many does the best answer use?
3. A problem gives n up to 2 × 10^5 and asks for the number of pairs that sum to k. Which approach do you reject, and which do you use?
4. Why does BFS give the fewest matatu changes but not the shortest travel time?
5. Your recursive flood fill passes small tests and crashes on a 1000 × 1000 grid. Why, and what's the fix?

<details>
<summary>Answers</summary>

1. `1 3 4`: the index of the first value at least as big as the target, and `len(xs)` when there's none.
2. Greedy takes 4 + 1 + 1, three coins. The best is 3 + 3, two coins, which DP finds.
3. Checking every pair is O(n²), about 2 × 10^10 steps: too slow. Count values in a map as you go, adding the count of `k - x` for each `x`: O(n).
4. BFS counts edges, as if every hop cost the same. Travel time puts a different weight on each edge, which needs Dijkstra.
5. A million cells can mean a million nested calls, which overflows the stack. Use an explicit stack (a slice) or a queue.

</details>

## Build it without steps: Stage Rush Hour

Write a single program in the HackerRank format: read from stdin, write to stdout, one second for the biggest input. Write the brute force first, then the fast version, and test one against the other on random inputs.

The input is a day of boardings at one matatu stage: the first line has `n` (up to 2 × 10^5) and `q` (up to 2 × 10^5), then `n` lines `minute route` (minutes from 0 to 1439, in any order; a route is a short word like `46K`), then `q` queries, one per line:

- `busiest K`: the K routes with the most boardings, most first, ties in alphabetical order.
- `window W`: the most boardings in any W consecutive minutes.
- `between A B`: how many boardings happened from minute A to minute B inclusive.
- `gap`: the longest stretch of minutes with no boarding at all.

Done when:

- Your fast version matches the brute force on a few thousand random small cases.
- It answers the largest input you can generate in under a second, measured with `time`.
- You can say the time and memory complexity of each query type, and why it fits the limits.
- Input is read with `bufio.Reader` and output written through one `bufio.Writer`.
- You wrote it in one sitting without looking at the Phase 3 solutions; afterwards, note which day you'd have looked up.
