# Kata: Statement Workers

A server renders `n` PDF statements with `k` workers, numbered 1 to `k`. The jobs are taken in order: whenever a worker is free, it takes the next job. If several workers are free at the same moment, the one with the smallest number goes first. Job `i` takes `t_i` seconds. At second 0 every worker is free.

Print the second when the last statement is done, and which worker rendered the most statements.

**Input**

- Line 1: `n` and `k` (1 ≤ n ≤ 100 000, 1 ≤ k ≤ 100 000)
- Line 2: `n` durations `t_1 ... t_n` in seconds (1 ≤ t_i ≤ 1 000 000)

**Output**

- Line 1: the second when every job is done
- Line 2: the number of the worker that rendered the most jobs, and how many it rendered. On a tie, print the smallest worker number.

**Example**

```
Input           Output
5 2             9
1 2 3 4 5       1 3
```

At 0, worker 1 takes job 1 (busy until 1) and worker 2 takes job 2 (until 2). At 1, worker 1 takes job 3 (until 4). At 2, worker 2 takes job 4 (until 6). At 4, worker 1 takes job 5 (until 9). Everything is done at 9, and worker 1 rendered 3 jobs.

**Limits:** 2 seconds per test case. Scanning all `k` workers for every job is `n × k` steps, up to 10^10. Too slow.

**Go tip**

Keep the workers in a min-heap ordered by (free time, worker number). The top is always the next worker to take a job: pop it, add the job's duration, push it back. Each job costs `O(log k)`. The standard library's `container/heap` works on any type with five methods:

```go
type worker struct{ free, id int }
type byFree []worker

func (h byFree) Len() int { return len(h) }
func (h byFree) Less(i, j int) bool {
	if h[i].free != h[j].free {
		return h[i].free < h[j].free
	}
	return h[i].id < h[j].id
}
func (h byFree) Swap(i, j int)  { h[i], h[j] = h[j], h[i] }
func (h *byFree) Push(x any)    { *h = append(*h, x.(worker)) }
func (h *byFree) Pop() any {
	old := *h
	w := old[len(old)-1]
	*h = old[:len(old)-1]
	return w
}
```

Then `heap.Init(&h)`, and `w := heap.Pop(&h).(worker)` and `heap.Push(&h, w)`. When the top only changes its free time, `h[0].free += t` followed by `heap.Fix(&h, 0)` is quicker than a pop and a push.

Judge it from the repo root with `go run ./tools/kata 46`.
