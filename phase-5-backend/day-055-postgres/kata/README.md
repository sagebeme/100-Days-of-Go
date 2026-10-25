# Kata: Connection Pool

Your API has a pool of `k` database connections. Queries arrive one by one. Each takes the connection that becomes free soonest. If that connection is free already, the query starts when it arrives; if not, it waits for it. Then it holds the connection for its whole duration.

Print when each query finishes, then the longest time any query waited. This is how you tell whether `MaxConns` is big enough for your traffic.

**Input**

- Line 1: `k` and `n`, the number of connections and of queries (1 ≤ k ≤ 1 000, 1 ≤ n ≤ 10 000)
- The next `n` lines: `t d`, the time the query arrives and how long it runs (0 ≤ t ≤ 10⁹, 0 ≤ d ≤ 10⁹). The arrival times never go down, and queries are served in this order.

**Output**

- `n` lines: the time each query finishes, in input order
- A last line: `max wait W`, where a query's wait is its start time minus its arrival time

**Example**

```
Input       Output
2 5         5
0 5         4
1 3         8
2 4         6
3 1         12
10 2        max wait 2
```

Both connections are busy when the third query arrives at time 2. The first one free is the second connection at time 4, so it starts at 4, finishes at 8 and waited 2. The fourth query gets the first connection at 5 and also waits 2.

**Limits:** 2 seconds per test case. Times can reach about 10¹³, which fits in Go's 64-bit `int`.

**Go tip**

Checking all `k` connections for every query is O(n·k), which is fine here. The textbook answer is a min-heap of "free at" times, which is O(n log k). `container/heap` (new today) needs a type with five methods; the smallest item is always at index 0:

```go
type freeAt []int

func (h freeAt) Len() int           { return len(h) }
func (h freeAt) Less(i, j int) bool { return h[i] < h[j] }
func (h freeAt) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *freeAt) Push(x any)        { *h = append(*h, x.(int)) }
func (h *freeAt) Pop() any          { old := *h; x := old[len(old)-1]; *h = old[:len(old)-1]; return x }

h := &freeAt{0, 0}                  // two connections, free at time 0
heap.Init(h)
soonest := heap.Pop(h).(int)
heap.Push(h, soonest+5)
```

Judge it from the repo root with `go run ./tools/kata 55`.
