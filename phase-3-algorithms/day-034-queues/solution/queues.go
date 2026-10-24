// Package queues uses first-in, first-out queues, deques and a priority queue.
package queues

import "container/heap"

// Queue is a first-in, first-out queue.
type Queue[T any] struct {
	items []T
	head  int // items[head:] are still waiting
}

// Enqueue adds v at the back.
func (q *Queue[T]) Enqueue(v T) {
	q.items = append(q.items, v)
}

// Dequeue removes and returns the front value.
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if q.head == len(q.items) {
		return zero, false
	}
	v := q.items[q.head]
	q.items[q.head] = zero // let the garbage collector have it
	q.head++
	// Once more than half the slice is used up, move the waiting values down.
	// Each value is moved at most once per halving, so this is O(1) amortised.
	if q.head > len(q.items)/2 {
		q.items = append(q.items[:0], q.items[q.head:]...)
		q.head = 0
	}
	return v, true
}

// Len returns how many values are waiting.
func (q *Queue[T]) Len() int {
	return len(q.items) - q.head
}

// Deque is a double-ended queue backed by a ring buffer.
type Deque[T any] struct {
	buf   []T
	front int
	n     int
}

// grow doubles the buffer when it's full, unrolling the ring into the new one.
func (d *Deque[T]) grow() {
	if d.n < len(d.buf) {
		return
	}
	bigger := make([]T, max(4, 2*len(d.buf)))
	for i := range d.n {
		bigger[i] = d.buf[(d.front+i)%len(d.buf)]
	}
	d.buf, d.front = bigger, 0
}

// PushBack adds v at the back.
func (d *Deque[T]) PushBack(v T) {
	d.grow()
	d.buf[(d.front+d.n)%len(d.buf)] = v
	d.n++
}

// PushFront adds v at the front.
func (d *Deque[T]) PushFront(v T) {
	d.grow()
	d.front = (d.front - 1 + len(d.buf)) % len(d.buf)
	d.buf[d.front] = v
	d.n++
}

// PopFront removes and returns the front value.
func (d *Deque[T]) PopFront() (T, bool) {
	v, ok := d.Front()
	if ok {
		var zero T
		d.buf[d.front] = zero
		d.front = (d.front + 1) % len(d.buf)
		d.n--
	}
	return v, ok
}

// PopBack removes and returns the back value.
func (d *Deque[T]) PopBack() (T, bool) {
	v, ok := d.Back()
	if ok {
		var zero T
		d.buf[(d.front+d.n-1)%len(d.buf)] = zero
		d.n--
	}
	return v, ok
}

// Front returns the front value without removing it.
func (d *Deque[T]) Front() (T, bool) {
	if d.n == 0 {
		var zero T
		return zero, false
	}
	return d.buf[d.front], true
}

// Back returns the back value without removing it.
func (d *Deque[T]) Back() (T, bool) {
	if d.n == 0 {
		var zero T
		return zero, false
	}
	return d.buf[(d.front+d.n-1)%len(d.buf)], true
}

// Len returns how many values the deque holds.
func (d *Deque[T]) Len() int {
	return d.n
}

// WindowMax returns the largest value in every window of k consecutive values.
func WindowMax(a []int, k int) []int {
	if k < 1 || k > len(a) {
		return []int{}
	}
	out := make([]int, 0, len(a)-k+1)
	var idx Deque[int] // indexes whose values go down from front to back
	for i, x := range a {
		for {
			b, ok := idx.Back()
			if !ok || a[b] > x {
				break
			}
			idx.PopBack() // a[b] can never be a maximum again: x is newer and at least as big
		}
		idx.PushBack(i)
		if f, _ := idx.Front(); f <= i-k {
			idx.PopFront() // slid out of the window
		}
		if i >= k-1 {
			f, _ := idx.Front()
			out = append(out, a[f])
		}
	}
	return out
}

// freeAt is a min-heap of the minutes at which each teller is next free.
type freeAt []int

func (h freeAt) Len() int           { return len(h) }
func (h freeAt) Less(i, j int) bool { return h[i] < h[j] }
func (h freeAt) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *freeAt) Push(x any)        { *h = append(*h, x.(int)) }
func (h *freeAt) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

// BankWaits returns how long each customer waits for a teller.
func BankWaits(arrivals, service []int, tellers int) []int {
	free := make(freeAt, tellers) // every teller is free at minute 0
	waits := make([]int, len(arrivals))
	for i, arrive := range arrivals {
		start := max(arrive, free[0]) // free[0] is the teller who is free first
		waits[i] = start - arrive
		free[0] = start + service[i]
		heap.Fix(&free, 0) // that teller is busy now: sink them back into place
	}
	return waits
}
