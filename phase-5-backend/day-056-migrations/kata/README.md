# Kata: Migration Order

Numbered migrations run in one fixed line. Some tools let migrations name what they depend on instead: "add_fare needs create_routes", and anything without a rule between them may run in either order. Your job is to pick the order.

You get `n` migration names and `m` rules `a b`, meaning `a` must run before `b`. Print an order that follows every rule. When several migrations are ready to run, run the one whose name comes first alphabetically (byte order). That makes the answer unique. If the rules form a loop, so no order can follow them all, print `CYCLE`.

**Input**

- Line 1: `n` and `m` (1 ≤ n ≤ 1 000, 0 ≤ m ≤ 3 000)
- The next `n` lines: one name each, 1 to 20 characters from `a`–`z`, `0`–`9` and `_`. Names are all different.
- The next `m` lines: `a b`, two of those names, different from each other. A rule can appear more than once.

**Output**

The `n` names in order, one per line, or the single line `CYCLE`.

**Example**

```
Input                              Output
4 3                                create_routes
create_trips                       add_fare
create_routes                      create_trips
create_bookings                    create_bookings
add_fare
create_routes create_trips
create_trips create_bookings
create_routes add_fare
```

At first only `create_routes` has nothing it waits for. After it, both `add_fare` and `create_trips` are ready, and `add_fare` comes first alphabetically.

**Limits:** 2 seconds per test case.

**Go tip**

This is a topological sort, Kahn's algorithm: count how many rules point at each name (its in-degree), start with the names at 0, and each time you run one, lower the count of everything it points to. If you run out of ready names before printing all `n`, there's a cycle. Keep a repeated rule as two edges; it adds 1 to the in-degree twice and is taken away twice, so it works out. With `n` at most 1 000, picking the smallest ready name by sorting or scanning each time is fast enough; a heap (Day 55's kata) is the general answer.

Judge it from the repo root with `go run ./tools/kata 56`.
