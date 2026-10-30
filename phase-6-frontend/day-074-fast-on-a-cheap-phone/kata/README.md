# Kata: Initial Bundle

A bundler starts at the entry module, `main`, and follows imports. A **static** import (`import x from "./x"`) goes into the first download. A **lazy** import (`lazy(() => import("./x"))`) starts a separate chunk that downloads only when needed. Given the module graph, work out what a passenger on a slow phone downloads before the page can start, what can wait, and which modules are dead weight.

- **Initial:** `main` and every module reachable from it through static imports only.
- **Later:** modules reachable from `main` through any mix of imports, but not initial.
- **Unused:** modules not reachable from `main` at all.

**Input**

- Line 1: `n m`: modules (1 ≤ n ≤ 5 000) and imports (0 ≤ m ≤ 10 000)
- Next `n` lines: `name bytes` (one word; 1 ≤ bytes ≤ 100 000). One module is called `main`.
- Next `m` lines: `from to kind`, where `kind` is `static` or `lazy`. Imports can form cycles.

**Output**

```
initial <total bytes of the initial modules>
modules <how many initial modules>
later <total bytes of the later modules>
unused <unused module names in sorted byte order, separated by spaces, or - if none>
```

**Example**

```
Input                      Output
6 5                        initial 182
main 40                    modules 3
react 130                  later 230
TripBoard 12               unused oldChart
SeatMap 90
leaflet 140
oldChart 60
main react static
main TripBoard static
TripBoard SeatMap lazy
SeatMap leaflet static
SeatMap react static
```

`main`, `react` and `TripBoard` load first: 182 bytes. `SeatMap` and its `leaflet` wait behind the lazy import: 230 bytes. `react` is already loaded, so it isn't counted again. Nothing imports `oldChart`.

**Limits:** 2 seconds per test case.

**Go tip:** two breadth-first searches from `main`: the first follows only static edges, the second follows both. Keep the edges in two maps, `static := map[string][]string{}` and `all := map[string][]string{}`, and a `seen` map per search.

Judge it from the repo root with `go run ./tools/kata 74`.
