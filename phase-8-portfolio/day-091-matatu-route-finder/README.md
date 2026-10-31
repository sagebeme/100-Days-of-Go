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
