// Package queues uses first-in, first-out queues, deques and a priority queue.
package queues

// Queue is a first-in, first-out queue. The zero value is an empty queue.
type Queue[T any] struct {
	items []T
	head  int // items[head:] are still waiting
}

// Enqueue adds v at the back. O(1) amortised.
func (q *Queue[T]) Enqueue(v T) {
	// Your code here.
}

// Dequeue removes and returns the value at the front and true, or the zero
// value and false if the queue is empty. O(1) amortised.
func (q *Queue[T]) Dequeue() (T, bool) {
	// Your code here.
	var zero T
	return zero, false
}

// Len returns how many values are waiting.
func (q *Queue[T]) Len() int {
	// Your code here.
	return 0
}

// Deque is a double-ended queue: add and remove at both ends in O(1). The
// zero value is an empty deque.
type Deque[T any] struct {
	buf   []T // a ring buffer
	front int // index in buf of the first value
	n     int // number of values
}

// PushBack adds v at the back.
func (d *Deque[T]) PushBack(v T) {
	// Your code here.
}

// PushFront adds v at the front.
func (d *Deque[T]) PushFront(v T) {
	// Your code here.
}

// PopFront removes and returns the front value, or the zero value and false.
func (d *Deque[T]) PopFront() (T, bool) {
	// Your code here.
	var zero T
	return zero, false
}

// PopBack removes and returns the back value, or the zero value and false.
func (d *Deque[T]) PopBack() (T, bool) {
	// Your code here.
	var zero T
	return zero, false
}

// Front returns the front value without removing it, or the zero value and false.
func (d *Deque[T]) Front() (T, bool) {
	// Your code here.
	var zero T
	return zero, false
}

// Back returns the back value without removing it, or the zero value and false.
func (d *Deque[T]) Back() (T, bool) {
	// Your code here.
	var zero T
	return zero, false
}

// Len returns how many values the deque holds.
func (d *Deque[T]) Len() int {
	// Your code here.
	return 0
}

// WindowMax returns the largest value in every window of k consecutive values,
// in order. If k < 1 or k > len(a) it returns an empty slice. It must be O(n).
//
//	WindowMax([]int{3, 1, 4, 1, 5, 9, 2}, 3) → [4 4 5 9 9]
func WindowMax(a []int, k int) []int {
	// Your code here.
	return nil
}

// BankWaits simulates a banking hall with the given number of tellers
// (at least 1). Customer i arrives at minute arrivals[i] (arrivals never go
// down) and needs service[i] minutes. Customers are served in arrival order,
// each by whichever teller is free first. It returns how long each customer
// waits before a teller starts serving them.
//
//	BankWaits([]int{0, 1, 2}, []int{5, 5, 5}, 2) → [0 0 3]
func BankWaits(arrivals, service []int, tellers int) []int {
	// Your code here.
	return nil
}
