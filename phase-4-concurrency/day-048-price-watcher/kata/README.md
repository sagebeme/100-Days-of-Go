# Kata: Price Board

Your price watcher feeds a live board. Shops change prices and stop selling things all day, and customers ask for the cheapest place to buy an item right now. Process the updates and answer the questions.

**Input**

- Line 1: `q`, the number of operations (1 ≤ q ≤ 200 000)
- The next `q` lines, one operation each:
  - `set SHOP ITEM PRICE`: `SHOP` now sells `ITEM` for `PRICE` shillings (1 ≤ PRICE ≤ 10^9). This replaces any earlier price that shop had for that item.
  - `del SHOP ITEM`: `SHOP` stops selling `ITEM`. If it didn't sell it, nothing changes.
  - `best ITEM`: ask for the cheapest offer for `ITEM`.

Shop and item names are 1 to 10 lowercase letters and digits.

**Output**

For each `best`, one line: the shop and the price of the cheapest offer, or `none` if no shop sells the item. If several shops share the lowest price, print the one whose name comes first in alphabetical order.

**Example**

```
Input                   Output
8                       njeri 230
set njeri unga 230      kona 225
set kona unga 245       njeri 230
best unga               none
set kona unga 225
best unga
del kona unga
best unga
best sukari
```

**Limits:** 2 seconds per test case. Scanning every shop on every `best` can be 200 000 × 100 000 steps when one item is sold everywhere.

**Go tip**

Keep the true prices in a map keyed by shop and item, `map[[2]string]int`. For each item, also keep a min-heap of `(price, shop)` pairs (Day 46's `container/heap`). A `set` pushes a new pair and never removes old ones. On `best`, look at the top: if the map no longer agrees with it (the price changed, or the shop stopped selling), pop it and look again. Each pair is popped at most once, so the whole run stays fast. That's called lazy deletion.

Judge it from the repo root with `go run ./tools/kata 48`.
