# Kata: Best Sellers

At closing time a duka owner wants to know what sold most today. You get every item sold, one name per sale. Print the `k` items sold most often, with their counts. Items sold the same number of times go in alphabetical order.

**Input**

- Line 1: `n` and `k` (1 ≤ n ≤ 200 000, 1 ≤ k ≤ 1000)
- Line 2: `n` item names separated by spaces, each 1 to 20 lowercase letters and digits

**Output**

The top `k` items, one per line: the name and how many were sold. If fewer than `k` different items were sold, print them all.

**Example**

```
Input                                               Output
8 2                                                 unga 3
unga sukari unga maziwa sukari unga mkate maziwa    maziwa 2
```

Maziwa and sukari both sold twice. Maziwa comes first in the alphabet, so it takes second place.

**Limits:** 2 seconds per test case.

**Go tip**

This is "top k frequent elements", a favourite interview question. Count with a map, move the counts into a slice of structs and sort it. `slices.SortFunc` with `cmp.Compare` reads well for a two-level sort:

```go
slices.SortFunc(list, func(a, b count) int {
	if a.n != b.n {
		return cmp.Compare(b.n, a.n) // more sales first
	}
	return cmp.Compare(a.name, b.name)
})
```

`slices.SortFunc` and `cmp.Compare` are generic functions from the standard library, the same kind of code you wrote today.

Judge it from the repo root with `go run ./tools/kata 18`.
