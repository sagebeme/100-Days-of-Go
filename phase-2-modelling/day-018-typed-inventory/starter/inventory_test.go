package typedinventory

import (
	"reflect"
	"strings"
	"testing"
)

func TestSum(t *testing.T) {
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Errorf("Sum([]int{1, 2, 3}) = %d, want 6", got)
	}
	if got := Sum([]Ksh{230, 185, 60}); got != 475 {
		t.Errorf("Sum([]Ksh{230, 185, 60}) = %d, want 475", got)
	}
	if got := Sum([]float64{0.5, 1.25}); got != 1.75 {
		t.Errorf("Sum([]float64{0.5, 1.25}) = %v, want 1.75", got)
	}
	if got := Sum([]int64{9_000_000_000, 1}); got != 9_000_000_001 {
		t.Errorf("Sum([]int64{9000000000, 1}) = %d, want 9000000001", got)
	}
	if got := Sum[int](nil); got != 0 {
		t.Errorf("Sum[int](nil) = %d, want 0", got)
	}
}

func TestMap(t *testing.T) {
	got := Map([]string{"unga", "sukari"}, strings.ToUpper)
	if want := []string{"UNGA", "SUKARI"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Map(names, strings.ToUpper)\n got: %q\nwant: %q", got, want)
	}
	lengths := Map([]string{"chai", "mandazi", ""}, func(s string) int { return len(s) })
	if want := []int{4, 7, 0}; !reflect.DeepEqual(lengths, want) {
		t.Errorf("Map(words, len)\n got: %v\nwant: %v", lengths, want)
	}
	if got := Map([]int{}, func(n int) int { return n }); got == nil || len(got) != 0 {
		t.Errorf("Map of an empty slice = %#v, want an empty non-nil slice", got)
	}
}

func TestFilter(t *testing.T) {
	got := Filter([]int{5, 12, 3, 40, 8}, func(n int) bool { return n >= 8 })
	if want := []int{12, 40, 8}; !reflect.DeepEqual(got, want) {
		t.Errorf("Filter(n >= 8)\n got: %v\nwant: %v", got, want)
	}
	none := Filter([]string{"a", "b"}, func(string) bool { return false })
	if none == nil || len(none) != 0 {
		t.Errorf("Filter that keeps nothing = %#v, want []string{}", none)
	}
}

func TestGroupBy(t *testing.T) {
	got := GroupBy([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
	want := map[bool][]int{false: {1, 3, 5}, true: {2, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy(even)\n got: %v\nwant: %v", got, want)
	}
	byLetter := GroupBy([]string{"sukuma", "unga", "sukari", "ugali"}, func(s string) byte { return s[0] })
	wantLetter := map[byte][]string{'s': {"sukuma", "sukari"}, 'u': {"unga", "ugali"}}
	if !reflect.DeepEqual(byLetter, wantLetter) {
		t.Errorf("GroupBy(first letter)\n got: %q\nwant: %q", byLetter, wantLetter)
	}
	if got := GroupBy([]int{}, func(n int) int { return n }); got == nil || len(got) != 0 {
		t.Errorf("GroupBy of an empty slice = %#v, want an empty non-nil map", got)
	}
}

func TestSortedKeys(t *testing.T) {
	got := SortedKeys(map[string]int{"sukari": 1, "chai": 2, "unga": 3})
	if want := []string{"chai", "sukari", "unga"}; !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeys(string map)\n got: %q\nwant: %q", got, want)
	}
	nums := SortedKeys(map[Ksh]bool{500: true, 50: true, 1000: false})
	if want := []Ksh{50, 500, 1000}; !reflect.DeepEqual(nums, want) {
		t.Errorf("SortedKeys(Ksh map)\n got: %v\nwant: %v", nums, want)
	}
}

func TestShelf(t *testing.T) {
	var s Shelf[string]
	if _, ok := s.Take(); ok {
		t.Errorf("Take() on an empty shelf returned ok = true")
	}
	for _, crate := range []string{"tomatoes from Monday", "tomatoes from Tuesday", "tomatoes from Wednesday"} {
		s.Put(crate)
	}
	if s.Len() != 3 {
		t.Errorf("Len() after 3 Puts = %d, want 3", s.Len())
	}
	for _, want := range []string{"tomatoes from Monday", "tomatoes from Tuesday"} {
		got, ok := s.Take()
		if !ok || got != want {
			t.Errorf("Take() = %q, %v; want %q, true", got, ok, want)
		}
	}
	s.Put("tomatoes from Thursday")
	if got, ok := s.Take(); !ok || got != "tomatoes from Wednesday" {
		t.Errorf("Take() = %q, %v; want %q, true", got, ok, "tomatoes from Wednesday")
	}
	if s.Len() != 1 {
		t.Errorf("Len() = %d, want 1", s.Len())
	}

	var eggs Shelf[Item]
	eggs.Put(Item{Name: "eggs", Qty: 30})
	if got, ok := eggs.Take(); !ok || got.Qty != 30 {
		t.Errorf("Shelf[Item].Take() = %+v, %v; want the eggs, true", got, ok)
	}
	if got, ok := eggs.Take(); ok || got != (Item{}) {
		t.Errorf("Take() on an empty Shelf[Item] = %+v, %v; want Item{}, false", got, ok)
	}
}

var stock = []Item{
	{Name: "unga", Category: "flour", Price: 230, Qty: 12},
	{Name: "sukari", Category: "sugar", Price: 185, Qty: 3},
	{Name: "maziwa", Category: "dairy", Price: 60, Qty: 40},
	{Name: "chapati flour", Category: "flour", Price: 210, Qty: 0},
	{Name: "mala", Category: "dairy", Price: 75, Qty: 4},
}

func TestStockValue(t *testing.T) {
	// 2760 + 555 + 2400 + 0 + 300
	if got := StockValue(stock); got != 6015 {
		t.Errorf("StockValue(stock) = %d, want 6015", got)
	}
	if got := StockValue(nil); got != 0 {
		t.Errorf("StockValue(nil) = %d, want 0", got)
	}
}

func TestLowStock(t *testing.T) {
	got := LowStock(stock, 5)
	if want := []string{"sukari", "chapati flour", "mala"}; !reflect.DeepEqual(got, want) {
		t.Errorf("LowStock(stock, 5)\n got: %q\nwant: %q", got, want)
	}
	if got := LowStock(stock, 0); got == nil || len(got) != 0 {
		t.Errorf("LowStock(stock, 0) = %#v, want []string{}", got)
	}
}

func TestGroupStockByCategory(t *testing.T) {
	groups := GroupBy(stock, func(it Item) string { return it.Category })
	if got, want := SortedKeys(groups), []string{"dairy", "flour", "sugar"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("categories\n got: %q\nwant: %q", got, want)
	}
	if got := StockValue(groups["dairy"]); got != 2700 {
		t.Errorf("dairy stock value = %d, want 2700", got)
	}
}
