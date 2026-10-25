# Kata: Component Tree

A React page is a tree: `App` renders `RouteList`, which renders a `RouteCard` for each route, which renders a `FareBadge`. React DevTools shows that tree with indentation. Build the same view from a list of "child parent" pairs.

**Input**

- Line 1: `n`, the number of components (1 ≤ n ≤ 3 000)
- Next `n` lines: `name parent`. Names are unique words. Exactly one component has parent `-`: that's the root. A pair can come before its parent's own line.

**Output**

- Every component, one per line, starting at the root. Indent each one by two spaces per level below the root. Print a component's children right after it (and their children after them), in the order the children appear in the input.
- Then `depth D`: the number of levels (a root on its own is depth 1).
- Then `leaves L`: the number of components with no children.

The tree is at most 50 levels deep.

**Example**

```
Input                Output
6                    App
App -                  Header
Header App             RouteList
RouteList App            RouteCard
RouteCard RouteList        FareBadge
FareBadge RouteCard    Footer
Footer App           depth 4
                     leaves 3
```

**Limits:** 2 seconds per test case.

**Go tip:** keep `children := map[string][]string{}` and append in input order. Then a recursive function prints one component and calls itself for each child with `depth+1`. `strings.Repeat("  ", depth)` makes the indent.

Judge it from the repo root with `go run ./tools/kata 66`.
