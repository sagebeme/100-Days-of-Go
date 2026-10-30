# Day 18: Typed Inventory

## The brief

A wholesaler in Gikomba keeps stock in a spreadsheet: unga, sukari, maziwa, prices, quantities. You've been asked for a few helpers. Add up a list of prices. Turn a list of items into a list of names. Keep only the items that are running low. Group them by category.

Written the Phase 1 way, you'd need `SumInts`, `SumKsh`, `SumFloats`, `FilterItems`, `FilterNames`... one copy per type. Today you write each helper once, for every type, with **generics**. Then you use them to answer the wholesaler's questions in one line each.

```
Sum([]Ksh{230, 185, 60})                        →  475
Map([]string{"unga", "sukari"}, strings.ToUpper) →  ["UNGA" "SUKARI"]
Filter([]int{5, 12, 3, 40}, atLeast8)            →  [12 40]
GroupBy([]int{1, 2, 3, 4}, isEven)               →  map[false:[1 3] true:[2 4]]
SortedKeys(map[string]int{"unga": 3, "chai": 2}) →  ["chai" "unga"]

var s Shelf[string]
s.Put("tomatoes from Monday"); s.Put("tomatoes from Tuesday")
s.Take()                                        →  "tomatoes from Monday", true

StockValue(stock)                               →  6015
LowStock(stock, 5)                              →  ["sukari" "chapati flour" "mala"]
```

## What you'll use

New today: **type parameters** and **constraints**.

- A **type parameter** goes in square brackets before the normal parameters: `func Sum[T Number](xs []T) T`. `T` stands for a type the caller chooses. Usually Go works it out from the arguments, so you write `Sum(prices)`, not `Sum[Ksh](prices)`. When it can't (`Sum(nil)`), you name it: `Sum[int](nil)`.
- A **constraint** says which types `T` may be, and so what you may do with a `T`. It's an interface:
  - `any` allows every type, but then you can only move values around, not add or compare them.
  - `comparable` allows types that work with `==`. Map keys must be `comparable`.
  - `cmp.Ordered`, from the standard library's `cmp` package, allows types that work with `<`: integers, floats and strings.
  - Your own: `type Number interface { ~int | ~int64 | ~float64 }`. The `|` lists the types; `~int` means "int or any type defined from int", which is how `Ksh` gets in.
- `type Ksh int` defines a **new type** with `int` underneath. `Ksh` and `int` don't mix without a conversion, `Ksh(it.Qty)`, which stops you adding shillings to a quantity by accident.
- A **generic type** has type parameters too: `type Shelf[T any] struct { items []T }`. Its methods use the same name: `func (s *Shelf[T]) Put(item T)`.
- **The zero value of `T`**: `T{}` doesn't compile, because `T` might be `int`. Write `var zero T` and return `zero`.
- **Functions as values**: `Map` and `Filter` take a function. You can pass a named one (`strings.ToUpper`) or write one in place: `func(n int) bool { return n >= 8 }`.

```go
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

Max(3, 7)           // 7
Max("chai", "unga") // "unga"
```

When should you reach for generics? When you'd otherwise copy the same function for several types and the body doesn't care which type it is. If the body needs a type switch, you probably want an interface ([Day 16](../day-016-payment-methods/)) instead.

## Steps

1. Open `starter/inventory.go`. The signatures are written. Read each one and say out loud what types it accepts.
2. Write `Sum`. Start from `var total T`, add every element, return it.
3. Write `Map` and `Filter`. Make the result with `make([]U, 0, len(xs))` in `Map` and `[]T{}` in `Filter`, so neither returns `nil` for an empty input.
4. Write `GroupBy`. Start with `groups := map[K][]T{}` and `append` each element to `groups[key(x)]`. Appending to a missing key works: its value is a `nil` slice, and `append` handles that.
5. Write `SortedKeys`. Collect the keys, then sort them with `slices.Sort(keys)` (import `"slices"`). It's generic too, which is why it works for both strings and `Ksh`.
6. Write the three `Shelf` methods. `Take` returns `s.items[0]` and keeps `s.items[1:]`.
7. Write `StockValue` with `Map` and `Sum`, and `LowStock` with `Filter` and `Map`. Each body fits on one or two lines.
8. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-018-typed-inventory/starter
```

## The kata: Best Sellers

Count what a duka sold today and print the top `k` items, with a tie-break. It's the interview classic "top k frequent elements", and the standard library's own generic `slices.SortFunc` does the sorting. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 18
```

## When you're stuck

- **`invalid operation: operator + not defined on total (variable of type T constrained by any)`**: `any` doesn't promise `+`. Use the `Number` constraint.
- **`Ksh does not satisfy Number (possibly missing ~ for int in Number)`**: the constraint says `int` but not `~int`.
- **`invalid composite literal type T`**: you wrote `T{}`. Use `var zero T` and return `zero`.
- **`cannot infer T`**: Go can't tell the type from a `nil` argument. Name it: `Sum[int](nil)`.
- **`cannot use generic type Shelf[T any] without instantiation`**: write `Shelf[string]`, not `Shelf`, when you declare one.
- **`mismatched types Ksh and int`**: convert the quantity, `it.Price * Ksh(it.Qty)`.
- **`Filter that keeps nothing = []int(nil)`**: start the result with `[]T{}`, not `var out []T`.
- **Still stuck?** Open `solution/inventory.go`, read it once, close it, then write your own version from memory.
