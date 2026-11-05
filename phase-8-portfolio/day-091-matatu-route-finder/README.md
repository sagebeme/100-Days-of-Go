# Day 91: Matatu Route Finder

## The brief

Wairimu just moved from Nyeri to Kawangware and starts a job in Ngong next week. Every app she tries knows the roads but not the matatus. A friend at a Nairobi transport startup asks you to build the engine for their route planner: given the stages and the matatu routes between them, find the fastest journey, and also the one with the fewest changes, because changing matatus with a bag in the CBD at rush hour is worth ten extra minutes to most people.

```
Fastest("Kawangware", "Ngong")
  Take the 46K from Kawangware to Kencom (45 min)
  Walk from Kencom to Railways (8 min)
  Take the 111 from Railways to Ngong (70 min)
  Total: 123 min, 1 transfer

Fastest("Kencom", "Kibera")        →  walk to Railways, Highrise Kibera Sacco to Kibera: 38 min, 0 transfers
Fastest("Westlands", "Ngong")      →  error: unknown stop: "Westlands"
FewestTransfers("Mto", "Ziwa")     →  one ride on C, 30 min, 0 transfers   (Fastest is 20 min with 1 transfer)
```

The Nairobi data uses only stages and routes checked on [nishukishe.com](https://nishukishe.com/directions) in October 2026: the 46K (Citi Hoppa) from Kencom or Ambassadeur to Kawangware, the 111 (MCA Sacco) from Railways to Ngong, the 125 (City Shuttle) from Railways to Ongata Rongai, the 45 (Nicco Travellers) from Ronald Ngala Street to Githurai, the 58 (Buruburu 58 Travellers Sacco) from Ambassadeur to Buruburu, and Highrise Kibera Sacco from Railways to Kibera. **The connections between stops and all the travel times are simplified for this exercise.** Real matatus stop in many more places, and times depend on traffic. Routes change, so confirm with the conductor. The tests for fewest transfers use a second, made-up network.

Acceptance criteria:

- `Fastest` is Dijkstra's algorithm with `container/heap`, and on a tie it prefers fewer transfers.
- `FewestTransfers` minimises the number of matatus boarded, then the minutes.
- Consecutive links on the same route are one leg. Walking is a leg but not a matatu, so walking between two rides still counts as a transfer.
- Unknown stops and unreachable stops are clear errors.

## What you'll use

- Graphs as adjacency lists, and BFS: [Day 37](../../phase-3-algorithms/day-037-graphs/). Queues: [Day 34](../../phase-3-algorithms/day-034-queues/).
- Greedy choice, which is what Dijkstra is: [Day 38](../../phase-3-algorithms/day-038-greedy/).
- Big-O: Dijkstra with a binary heap is O((V + E) log V). See [reference/big-o.md](../../reference/big-o.md).
- New today: `container/heap`. You give it a type with `Len`, `Less`, `Swap`, `Push` and `Pop`, and `heap.Push` and `heap.Pop` keep the smallest item on top.

```go
type queue []*item
func (q queue) Less(i, j int) bool { return q[i].minutes < q[j].minutes }
// ... Len, Swap, Push, Pop
heap.Push(q, &item{stop: "Railways", minutes: 8})
next := heap.Pop(q).(*item) // the closest unsettled stop
```

The idea that makes fewest transfers work: search over **states**, not stops. A state is (stop, the matatu you're on). Arriving at Railways on the 125 and arriving there on foot are different states, because only one of them can carry on along the 125 for free. The cost is a pair (boardings, minutes), compared boardings first.

## Steps

Requirements. Meet them in any order. The tests in `starter/` check each one.

Must-haves:

1. `NewNetwork` validates each link and stores it in both directions. `Stops` keeps first-seen order; `Has` checks a name.
2. A heap of search items, with `container/heap`. Break ties in a fixed way (an insertion counter) so results never depend on map order.
3. `Fastest`: Dijkstra over (stop, route) states with cost (minutes, boardings). Skip stale entries when you pop them. Stop when you pop the destination.
4. `FewestTransfers`: the same search with cost (boardings, minutes). Boarding a route you're not already on adds one; walking puts you on no route.
5. Keep a `prev` map, walk it back from the destination, and merge consecutive links on the same route into legs. `Transfers` is matatu legs minus one, never below zero.
6. `Describe` prints the directions as shown, with "1 transfer" but "0 transfers".
7. `cmd/route` prints directions between two stops given on the command line, with `-fewest` for the other search.

```bash
go test ./phase-8-portfolio/day-091-matatu-route-finder/starter
go run ./phase-8-portfolio/day-091-matatu-route-finder/starter/cmd/route Kawangware Ngong
```

Stretch goals:

- Fares: add Ksh to each link and a third search for the cheapest journey, or minutes with a fare limit.
- A transfer penalty in minutes, so "fastest" can mean "fastest, counting 5 minutes to find the next matatu".
- Load the network from a CSV file ([Day 24](../../phase-2-modelling/day-024-csv-statements/)) and serve it as a JSON API: `GET /journeys?from=Kawangware&to=Ngong`.
- A* with a straight-line distance heuristic, once stops have coordinates.

## The kata: Within K Changes

The cheapest fare from one stage to another using at most `k` changes. Plain Dijkstra fails here, and the README shows the round-by-round fix. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 91
```

## Check yourself

1. Why does the search run over (stop, route) states, not just stops? What would a plain stop-by-stop search get wrong about transfers?
2. Predict the results on the made-up network: Mto to Soko on route A (10 min), Soko to Ziwa on B (10 min), Mto to Shule on C (15 min), Shule to Ziwa on C (15 min).

   ```go
   f, _ := n.Fastest("Mto", "Ziwa")
   g, _ := n.FewestTransfers("Mto", "Ziwa")
   fmt.Println(f.Minutes, f.Transfers, len(f.Legs), g.Minutes, g.Transfers, len(g.Legs))
   ```

3. Spot the problems. This finds the right number of minutes, but it does far more work than it needs on a big network, and it can't answer the brief:

   ```go
   for q.Len() > 0 {
       it := heap.Pop(q).(*item)
       for _, l := range n.adj[it.stop] {
           cost := it.minutes + l.Minutes
           if old, ok := best[l.To]; !ok || cost < old {
               best[l.To] = cost
               heap.Push(q, &item{stop: l.To, minutes: cost})
           }
       }
   }
   ```

   Hint: what happens when a stop has an old, more expensive entry still in the heap, and when you pop the destination?
4. Walking from Kencom to Railways between two matatus: why does it still count as a transfer, and why break ties with an insertion counter rather than leaving them to map order?

<details>
<summary>Answers</summary>

1. Arriving at a stop on the 125 and arriving there on foot are different situations: only one can carry on along the 125 without boarding again. Searching by stop alone keeps only the first way it reached each stop, so it can throw away the one that needs fewer boardings.
2. `20 1 2 30 0 1`. Fastest takes A then B (20 minutes, one change); FewestTransfers stays on C all the way (30 minutes, no change, one leg).
3. It doesn't skip stale entries: when it pops an item whose cost is worse than `best[it.stop]`, it should `continue`, not expand that stop again. And it never stops at the destination; once the destination is popped its cost is final, so return there. It also searches stops, not (stop, route) states, and keeps no `prev` map, so it can't count transfers, break ties on them, or print the legs.
4. Walking isn't a matatu, so you still board a second one: two matatu legs means one transfer. Ties broken by map order would give different answers on different runs, and tests (and passengers) need the same answer every time.

</details>

## When you're stuck

- **`panic: interface conversion: interface {} is *routes.item, not routes.item`**: you pushed pointers and asserted a value, or the other way round. Pick one.
- **The route is right but slower than it should be**: you marked a stop as done when you pushed it. Mark it when you pop it, and skip it if it was already done.
- **Fewest transfers gives the 20-minute route with 1 transfer**: your state is only the stop. Arriving at Ziwa on the B and arriving on the C must be separate states.
- **Walk, then the same route again, counts 0 transfers**: walking must reset the current route to none, so getting back on is a new boarding.
- **The legs list has `Kencom → Ambassadeur` and `Ambassadeur → Ronald Ngala Street` as two walks**: merge a link into the previous leg when the route is the same.
- **`Fastest("Ngong", "Ngong")` returns an error**: the start is the destination. Return a journey with one stop and no legs.
- **Still stuck?** Open `solution/`, read it once, close it, then write your own version from memory.

## Make it portfolio-ready

- Give it its own repo. The README must say where the route data comes from (nishukishe.com, with the date) and that connections and times are simplified.
- Draw the Nairobi network as a diagram (Mermaid works in GitHub READMEs) and show a sample of `cmd/route` output.
- Put a small web page or JSON API in front of it and deploy it, with the demo link at the top.
- Write "What I'd do next": real stop lists, fares, timetables or frequency, a transfer penalty.
- Add GitHub Actions running `go test` with the CI badge, and a benchmark (`go test -bench`) on a generated 10 000-stop network.
