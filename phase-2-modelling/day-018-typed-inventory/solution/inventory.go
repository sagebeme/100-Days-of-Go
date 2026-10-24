// Package typedinventory keeps a shop's stock with generic helpers: one Sum, Map,
// Filter and GroupBy that work for any type, and a Shelf that holds any kind of item.
package typedinventory

import (
	"cmp"
	"slices"
)

// Ksh is an amount in shillings. It's a new type whose underlying type is int.
type Ksh int

// Number is a constraint: the types Sum can add. The ~ means "this type or any type
// whose underlying type is this", so Ksh counts as an int.
type Number interface {
	~int | ~int64 | ~float64
}

// Sum adds up xs. An empty or nil slice sums to 0.
//
//	Sum([]int{1, 2, 3})       → 6
//	Sum([]Ksh{230, 185})      → Ksh(415)
//	Sum([]float64{0.5, 1.25}) → 1.75
func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

// Map returns a new slice holding f(x) for every x in xs, in order.
//
//	Map([]string{"unga", "sukari"}, strings.ToUpper) → []string{"UNGA", "SUKARI"}
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

// Filter returns the elements of xs for which keep returns true, in order. It
// returns an empty, non-nil slice when nothing is kept.
func Filter[T any](xs []T, keep func(T) bool) []T {
	out := []T{}
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

// GroupBy puts each element of xs in a group named by key(x). Inside a group the
// elements keep their order from xs.
//
//	GroupBy([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
//	→ map[bool][]int{false: {1, 3}, true: {2, 4}}
func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	groups := map[K][]T{}
	for _, x := range xs {
		k := key(x)
		groups[k] = append(groups[k], x)
	}
	return groups
}

// SortedKeys returns m's keys in increasing order. The constraint cmp.Ordered allows
// any type that < works on: numbers and strings.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Shelf holds items of one type. Take hands out the item that was Put first, so old
// stock sells before new stock.
type Shelf[T any] struct {
	items []T
}

// Put adds item to the back of the shelf.
func (s *Shelf[T]) Put(item T) {
	s.items = append(s.items, item)
}

// Take removes and returns the oldest item. On an empty shelf it returns T's zero
// value and false.
func (s *Shelf[T]) Take() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	item := s.items[0]
	s.items = s.items[1:]
	return item, true
}

// Len reports how many items are on the shelf.
func (s *Shelf[T]) Len() int {
	return len(s.items)
}

// Item is one line of stock.
type Item struct {
	Name     string
	Category string
	Price    Ksh // for one
	Qty      int
}

// StockValue returns what all the stock is worth: the sum of Price × Qty. Write it
// with Map and Sum.
func StockValue(items []Item) Ksh {
	return Sum(Map(items, func(it Item) Ksh { return it.Price * Ksh(it.Qty) }))
}

// LowStock returns the names of items with fewer than min left, in their original
// order. Write it with Filter and Map.
func LowStock(items []Item, min int) []string {
	low := Filter(items, func(it Item) bool { return it.Qty < min })
	return Map(low, func(it Item) string { return it.Name })
}
