# Day 34: Queues

## The brief

Monday morning at a bank in Thika: customers take a ticket and wait their turn. First in, first out. That's a queue. Today you'll build one, then a deque (a queue you can use from both ends), and use them for two classics: the sliding window maximum and a banking hall simulator with several tellers.

The simulator needs one more kind of queue. When a customer reaches the front, they go to whichever teller is free **first**. "Give me the smallest" is a priority queue, and Go has one in `container/heap`.

```
var q Queue[string]
q.Enqueue("Akinyi"); q.Enqueue("Baraka")
q.Dequeue()                                     →  "Akinyi", true
WindowMax([]int{3, 1, 4, 1, 5, 9, 2}, 3)        →  [4 4 5 9 9]
BankWaits([]int{0, 1, 2}, []int{5, 5, 5}, 2)    →  [0 0 3]
```

In `BankWaits`, Akinyi and Baraka go straight to the two tellers. Chebet arrives at minute 2 and waits until Akinyi's teller is free at minute 5: 3 minutes.

## What you'll use

### A queue from a slice

`append` adds at the back. Removing from the front with `s = s[1:]` is O(1) but never gives memory back. Today's `Queue` keeps a `head` index and, once more than half the slice is used up, moves the waiting values down to the start. Each value moves rarely, so `Dequeue` is O(1) amortised, like `append`. What you must never do is delete the front by shifting every value along: that's O(n) per removal and O(n²) for the whole queue ([reference/big-o.md](../../reference/big-o.md#common-go-operations)).

### A deque from a ring buffer

A ring buffer is a slice where the end wraps round to the start. Keep the index of the front and the count. The back is at `(front + n - 1) % len(buf)`. Pushing at the front moves `front` back one, wrapping: `(front - 1 + len(buf)) % len(buf)`. When it's full, copy the values in order into a slice twice the size.

### The sliding window maximum

Keep a deque of indexes whose values fall from front to back. For each new value, pop smaller (or equal) values off the back: they're older and no bigger, so they'll never be a maximum again. Push the new index. Pop the front if it has slid out of the window. The front is the maximum. Each index is pushed and popped once: O(n).

### A priority queue with container/heap

**New today.** `container/heap` keeps a slice arranged so the smallest value is always at index 0. Push and Pop are O(log n). You give it a type with five methods: `Len`, `Less`, `Swap` from `sort.Interface`, plus `Push` and `Pop`:

```go
type freeAt []int // minute each teller is next free

func (h freeAt) Len() int           { return len(h) }
func (h freeAt) Less(i, j int) bool { return h[i] < h[j] }
func (h freeAt) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *freeAt) Push(x any)        { *h = append(*h, x.(int)) }
func (h *freeAt) Pop() any          { old := *h; x := old[len(old)-1]; *h = old[:len(old)-1]; return x }
```

You don't call those methods yourself. You call `heap.Push(&h, x)`, `heap.Pop(&h)`, and `heap.Fix(&h, i)` after changing `h[i]`. A slice of all-equal values is already a valid heap, so `make(freeAt, tellers)` needs no `heap.Init`. Dijkstra's shortest path (Day 40's kata) is built on the same heap.

### Spotting it in a problem

- "**in the order they arrive**", "first come, first served", "simulate": a queue.
- "**shortest path**" on an unweighted graph or grid: BFS, which runs on a queue (Day 37).
- "**maximum / minimum of every window**": a monotonic deque.
- "the **smallest / earliest / cheapest** next", "the top k", "merge k sorted lists": a heap.

### Brute force vs optimal

| Problem | Brute force | Better |
|---|---|---|
| remove from the front n times | shift the slice: O(n²) | head index or ring buffer: O(n) |
| max of every window of k | rescan each window: O(n·k) | monotonic deque: O(n) |
| next free of t tellers, n customers | scan every teller: O(n·t) | heap: O(n log t) |
| sum of window maxima (kata) | O(n·k) | O(n) |

## Steps

1. Open `starter/queues.go`. Write `Queue`'s three methods. In `Dequeue`, after taking `items[head]`, set that slot to the zero value and move `head` on. Then, if `head > len(items)/2`, copy the waiting values to the start with `append(q.items[:0], q.items[q.head:]...)` and reset `head` to 0.
2. Write the deque. Start with a private `grow` method that doubles the buffer (at least 4 slots) when `n == len(buf)`, copying value `i` from `buf[(front+i)%len(buf)]`. Call it at the top of both pushes. Write `Front` and `Back` next, then build the pops on them.
3. Write `WindowMax` with a `Deque[int]` of indexes, as described above. Return `[]int{}` for a bad `k`.
4. Write `BankWaits` with a `freeAt` heap. Import `"container/heap"`. For each customer: the start is the later of their arrival and `free[0]`. The wait is start minus arrival. Then set `free[0]` to start plus service and call `heap.Fix(&free, 0)`.
5. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-034-queues/starter
```

## The kata: Busiest Windows

The sum of the longest queue in every window of `k` minutes. It's the sliding window maximum, HackerRank and LeetCode's favourite deque problem, with the answers added up so the output stays one line. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 34
```

## When you're stuck

- **`integer divide by zero`** in the deque: you used `% len(d.buf)` while the buffer was still empty. Call `grow` before every push, and check `n == 0` before every read.
- **The deque returns values in the wrong order after it grows:** copy in queue order, `buf[(front+i)%len(buf)]`, not `buf[i]`, and set `front = 0` afterwards.
- **`WindowMax` returns a value from outside the window:** pop the front when its index is `<= i-k`.
- **`WindowMax([]int{4, 4, 4, 4}, 3)` keeps old 4s:** pop from the back on `<=`, not just `<`.
- **`missing method Pop`** or **`does not implement heap.Interface`:** `Push` and `Pop` need pointer receivers (`*freeAt`), and you pass `&free` to the heap functions.
- **`BankWaits` makes a customer wait when a teller is free:** you used the teller who was free last. `free[0]` is the earliest, but only if you called `heap.Fix` after every change.
- **Still stuck?** Open `solution/queues.go`, read it once, close it, then write your own version from memory.
