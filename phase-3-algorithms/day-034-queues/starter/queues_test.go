package queues

import (
	"slices"
	"testing"
)

func TestQueue(t *testing.T) {
	var q Queue[string]
	if _, ok := q.Dequeue(); ok {
		t.Fatalf("Dequeue on an empty queue returned ok = true")
	}
	for _, name := range []string{"Akinyi", "Baraka", "Chebet"} {
		q.Enqueue(name)
	}
	if q.Len() != 3 {
		t.Fatalf("Len after 3 Enqueues = %d, want 3", q.Len())
	}
	for _, want := range []string{"Akinyi", "Baraka"} {
		if got, ok := q.Dequeue(); got != want || !ok {
			t.Errorf("Dequeue = %q, %v, want %q, true", got, ok, want)
		}
	}
	q.Enqueue("Dalmas")
	for _, want := range []string{"Chebet", "Dalmas"} {
		if got, ok := q.Dequeue(); got != want || !ok {
			t.Errorf("Dequeue = %q, %v, want %q, true", got, ok, want)
		}
	}
	if got, ok := q.Dequeue(); got != "" || ok || q.Len() != 0 {
		t.Errorf("Dequeue on an emptied queue = %q, %v (Len %d), want \"\", false (Len 0)", got, ok, q.Len())
	}
}

func TestQueueBig(t *testing.T) {
	// 10^6 customers through the queue, a few at a time. Removing from the
	// front with s = s[1:] works too, but deleting by shifting the slice would be O(n²).
	var q Queue[int]
	next := 0
	for i := range 1_000_000 {
		q.Enqueue(i)
		if i%3 != 0 {
			got, _ := q.Dequeue()
			if got != next {
				t.Fatalf("Dequeue = %d, want %d", got, next)
			}
			next++
		}
	}
	if q.Len() != 1_000_000-next {
		t.Errorf("Len = %d, want %d", q.Len(), 1_000_000-next)
	}
}

func TestDeque(t *testing.T) {
	var d Deque[int]
	if _, ok := d.PopFront(); ok {
		t.Fatalf("PopFront on an empty deque returned ok = true")
	}
	if _, ok := d.PopBack(); ok {
		t.Fatalf("PopBack on an empty deque returned ok = true")
	}
	if _, ok := d.Front(); ok {
		t.Fatalf("Front on an empty deque returned ok = true")
	}
	d.PushBack(2)
	d.PushBack(3)
	d.PushFront(1)
	d.PushFront(0) // 0 1 2 3
	if d.Len() != 4 {
		t.Fatalf("Len = %d, want 4", d.Len())
	}
	if f, _ := d.Front(); f != 0 {
		t.Errorf("Front = %d, want 0", f)
	}
	if b, _ := d.Back(); b != 3 {
		t.Errorf("Back = %d, want 3", b)
	}
	if v, ok := d.PopBack(); v != 3 || !ok {
		t.Errorf("PopBack = %d, %v, want 3, true", v, ok)
	}
	if v, ok := d.PopFront(); v != 0 || !ok {
		t.Errorf("PopFront = %d, %v, want 0, true", v, ok)
	}
	// Push past the buffer's size from both ends, so the ring wraps and grows.
	for i := range 100 {
		d.PushFront(-i)
		d.PushBack(100 + i)
	}
	if d.Len() != 202 {
		t.Fatalf("Len = %d, want 202", d.Len())
	}
	if f, _ := d.Front(); f != -99 {
		t.Errorf("Front = %d, want -99", f)
	}
	if b, _ := d.Back(); b != 199 {
		t.Errorf("Back = %d, want 199", b)
	}
	var all []int
	for d.Len() > 0 {
		v, _ := d.PopFront()
		all = append(all, v)
	}
	if all[0] != -99 || all[99] != 0 || all[100] != 1 || all[101] != 2 || all[102] != 100 || all[201] != 199 {
		t.Errorf("deque emptied in the wrong order: %v", all)
	}
}

func TestWindowMax(t *testing.T) {
	tests := []struct {
		a    []int
		k    int
		want []int
	}{
		{nil, 1, []int{}},
		{[]int{5}, 0, []int{}},
		{[]int{5}, 2, []int{}},
		{[]int{5}, 1, []int{5}},
		{[]int{3, 1, 4, 1, 5, 9, 2}, 3, []int{4, 4, 5, 9, 9}},
		{[]int{3, 1, 4, 1, 5, 9, 2}, 1, []int{3, 1, 4, 1, 5, 9, 2}},
		{[]int{3, 1, 4, 1, 5, 9, 2}, 7, []int{9}},
		{[]int{9, 8, 7, 6, 5}, 2, []int{9, 8, 7, 6}},
		{[]int{1, 2, 3, 4, 5}, 2, []int{2, 3, 4, 5}},
		{[]int{4, 4, 4, 4}, 3, []int{4, 4}},
		{[]int{-5, -1, -3}, 2, []int{-1, -1}},
	}
	for _, tt := range tests {
		got := WindowMax(tt.a, tt.k)
		if !slices.Equal(got, tt.want) || got == nil {
			t.Errorf("WindowMax(%v, %d) = %#v, want %v", tt.a, tt.k, got, tt.want)
		}
	}
}

func TestWindowMaxBig(t *testing.T) {
	n, k := 200_000, 100_000
	a := make([]int, n)
	for i := range a {
		a[i] = n - i // falling: every window's maximum is its first value
	}
	got := WindowMax(a, k) // checking each window from scratch is 10^10 steps
	if len(got) != n-k+1 || got[0] != n || got[len(got)-1] != k {
		t.Errorf("WindowMax(big): len %d, first %d, last %d; want %d, %d, %d", len(got), got[0], got[len(got)-1], n-k+1, n, k)
	}
}

func TestBankWaits(t *testing.T) {
	tests := []struct {
		name     string
		arrivals []int
		service  []int
		tellers  int
		want     []int
	}{
		{"nobody", []int{}, []int{}, 3, []int{}},
		{"one customer", []int{10}, []int{4}, 1, []int{0}},
		{"one teller, a queue forms", []int{0, 1, 2}, []int{5, 5, 5}, 1, []int{0, 4, 8}},
		{"two tellers", []int{0, 1, 2}, []int{5, 5, 5}, 2, []int{0, 0, 3}},
		{"teller free again before the next arrival", []int{0, 10}, []int{3, 3}, 1, []int{0, 0}},
		{"same arrival time", []int{0, 0, 0, 0}, []int{2, 4, 6, 8}, 2, []int{0, 0, 2, 4}},
		{"more tellers than customers", []int{0, 0}, []int{9, 9}, 5, []int{0, 0}},
		{"zero-minute service", []int{0, 0, 0}, []int{0, 0, 0}, 1, []int{0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BankWaits(tt.arrivals, tt.service, tt.tellers)
			if !slices.Equal(got, tt.want) {
				t.Errorf("BankWaits(%v, %v, %d) = %v, want %v", tt.arrivals, tt.service, tt.tellers, got, tt.want)
			}
		})
	}
}

func TestBankWaitsBig(t *testing.T) {
	// 200 000 customers arrive at once, 1000 tellers, 1 minute each.
	// Scanning every teller for each customer is 2 × 10^8 steps; a heap is far fewer.
	n, tellers := 200_000, 1000
	arrivals := make([]int, n)
	service := make([]int, n)
	for i := range service {
		service[i] = 1
	}
	got := BankWaits(arrivals, service, tellers)
	if got[0] != 0 || got[999] != 0 || got[1000] != 1 || got[n-1] != (n-1)/tellers {
		t.Errorf("BankWaits(big): got[0]=%d got[999]=%d got[1000]=%d got[n-1]=%d", got[0], got[999], got[1000], got[n-1])
	}
}
