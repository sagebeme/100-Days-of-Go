# Day 37: Graphs

## The brief

A matatu map is a graph: stops are the **nodes** and the rides or walks between them are the **edges**. "How do I get from Githurai to Ngong with the fewest changes?" is a shortest-path question, and "is there any way at all?" is a reachability question. Breadth-first search (BFS) and depth-first search (DFS) answer both, and they're the two graph algorithms every timed test expects.

Today's map uses real Nairobi stops and the routes listed on [nishukishe.com](https://nishukishe.com/directions) (checked October 2026). The connections are simplified for the exercise: one link per route, plus short walks between CBD stages. It is not a journey planner. Routes change, so confirm with the conductor.

| Route | Operator | Boards at (CBD) | Goes to |
|---|---|---|---|
| 46K | Citi Hoppa | Kencom / Ambassadeur | Kawangware |
| 111 | MCA Sacco | Railways | Ngong |
| 125 | City Shuttle | Railways | Ongata Rongai |
| 45 | Zamzam | Ronald Ngala Street | Githurai |
| 58 | Buruburu 58 Travellers Sacco | Ambassadeur | Buruburu |
| (number not shown) | Highrise Kibera Sacco | Railways | Kibera |

```
         Kawangware          Buruburu        Githurai
              |                  |               |
            Kencom ------- Ambassadeur ---- Ronald Ngala Street
              |
           Railways ---- Ngong
           /      \
     Kibera        Ongata Rongai

g.FewestHops("Githurai", "Ngong")  →  5, true
g.Path("Kibera", "Buruburu")       →  [Kibera Railways Kencom Ambassadeur Buruburu]
g.Reachable("Giza")                →  [Fahari Giza]       on a made-up village map
g.CountNetworks()                  →  2                   on the same village map
```

## What you'll use

### An adjacency list

Store each node's neighbours: `map[string][]string`. An undirected edge goes in both lists. For numbered nodes, `[][]int` is faster. A graph with V nodes and E edges takes O(V + E) memory.

### BFS: shortest paths when every edge costs the same

Start with the start node in a queue (Day 34). Take the front, add each neighbour you haven't seen, and remember where you came from. BFS reaches nodes in rings: everything 1 hop away, then 2 hops, and so on. So the first time it reaches a node, that's the fewest hops. Walk the `cameFrom` map back from the end to rebuild the path.

```go
cameFrom := map[string]string{start: ""}
queue := []string{start}
for len(queue) > 0 {
	stop := queue[0]
	queue = queue[1:]
	for _, next := range g.Neighbours(stop) {
		if _, seen := cameFrom[next]; !seen {
			cameFrom[next] = stop
			queue = append(queue, next)
		}
	}
}
```

Visiting neighbours in alphabetical order makes the answer the same every run. Map order is random, and when two routes are equally short, a different order picks a different one.

### DFS: go deep, then back up

DFS follows one road as far as it goes before trying the next. Write it with recursion, or with an explicit stack (Day 33) when the graph might be very deep. It's the simple choice for "can I get there?", "how many separate networks?", flood fill and cycle detection. It does **not** find shortest paths.

### Cost

Both visit each node once and look at each edge once (twice if undirected): O(V + E). A grid is a graph too: each cell is a node with up to four edges (the kata).

If edges have different costs, like minutes in traffic, BFS isn't enough. Dijkstra's algorithm uses a heap (Day 34) instead of a queue: O((V + E) log V). Day 40's kata needs it.

### Spotting it in a problem

- "**fewest** steps / moves / changes", "shortest path" with no weights, "minimum number of jumps": BFS.
- "is there a path", "connected", "**number of islands** / groups / provinces", "flood fill": DFS or BFS.
- A grid with walls, a maze, a board: a graph in disguise.
- "dependencies", "order of tasks", "prerequisites": a directed graph and topological sort, built from the same DFS.

### Brute force vs optimal

| Problem | Brute force | Better |
|---|---|---|
| shortest path, unweighted | try every route: exponential | BFS: O(V + E) |
| count networks | BFS from every node, comparing results: O(V·(V + E)) | one DFS per unvisited node: O(V + E) |
| shortest path in a grid (kata) | recursion over all routes: exponential | BFS over cells: O(R·C) |

See the graph rows in [reference/big-o.md](../../reference/big-o.md#patterns-and-their-complexity).

## Steps

1. Open `starter/graphs.go`. Write `AddRoute`. Make the map if it's `nil`, ignore `a == b`, skip the edge if `b` is already in `a`'s list (`slices.Contains`), then append each stop to the other's list.
2. Write `Neighbours`: clone the list (`slices.Clone`) and sort the copy, so you never reorder the graph itself.
3. Write `Path` with the BFS above. Return `nil` if `from` isn't in the map or `to` was never reached. Otherwise follow `cameFrom` from `to` back to `""`, then `slices.Reverse` the result.
4. Write `FewestHops` by calling `Path`: the number of hops is one less than the number of stops.
5. Write `Reachable` with a recursive DFS and a `seen` map. Collect the keys, sort them, return them.
6. Write `CountNetworks`: loop over every stop. Each one not yet seen starts a new network: count it, then mark everything it can reach. Try an explicit stack here instead of recursion.
7. Run the tests from the repo root:

```bash
go test ./phase-3-algorithms/day-037-graphs/starter
```

## The kata: Shortcut to the Stage

Fewest steps across a grid of streets and buildings to the stage. Grid BFS is probably the single most common graph problem in timed tests. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 37
```

## When you're stuck

- **`assignment to entry in nil map`:** `AddRoute` wrote into a `Graph` whose map was never made. Check `g.adj == nil` and make it.
- **`Path` returns the right stops backwards:** you built it from `to` to `from`. Reverse it.
- **`Path(Amani, Embe)` goes through Dawa:** you ranged over `g.adj[stop]` directly. Use `g.Neighbours(stop)`, which is sorted.
- **`FewestHops("Kencom", "Kencom")` is false:** the start is in `cameFrom` with `""` as its parent, so the path is just `[Kencom]`, 0 hops.
- **The program never ends:** you add stops to the queue without marking them seen, so they go round in circles. Mark a stop seen when you queue it.
- **`CountNetworks` is too big:** you counted every stop, not every unseen stop. Mark everything reachable before moving on.
- **Still stuck?** Open `solution/graphs.go`, read it once, close it, then write your own version from memory.
