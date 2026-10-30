package routes

import (
	"container/heap"
	"fmt"
	"slices"
)

// state is where a traveller is and which matatu they're sitting in ("" when on foot
// or not yet boarded). Fewest transfers needs the route in the state: arriving at a
// stop on the 46K is not the same as arriving there on foot.
type state struct {
	stop, route string
}

type cost struct {
	boardings, minutes int
}

type item struct {
	s    state
	c    cost
	seq  int // insertion order, so equal costs pop in a fixed order
	less func(a, b cost) bool
}

// queue is a min-heap of items, ordered by the search's cost comparison.
type queue []*item

func (q queue) Len() int { return len(q) }
func (q queue) Less(i, j int) bool {
	if q[i].c != q[j].c {
		return q[i].less(q[i].c, q[j].c)
	}
	return q[i].seq < q[j].seq
}
func (q queue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *queue) Push(x any)   { *q = append(*q, x.(*item)) }
func (q *queue) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}

func byMinutes(a, b cost) bool {
	if a.minutes != b.minutes {
		return a.minutes < b.minutes
	}
	return a.boardings < b.boardings
}

func byBoardings(a, b cost) bool {
	if a.boardings != b.boardings {
		return a.boardings < b.boardings
	}
	return a.minutes < b.minutes
}

// Fastest returns the journey with the fewest minutes, using Dijkstra's algorithm
// with container/heap. Among equally fast journeys it picks the one with fewer
// transfers. A journey from a stop to itself has one stop and no legs.
func (n *Network) Fastest(from, to string) (Journey, error) {
	return n.search(from, to, byMinutes)
}

// FewestTransfers returns the journey with the fewest matatu changes. Among those it
// picks the fastest.
func (n *Network) FewestTransfers(from, to string) (Journey, error) {
	return n.search(from, to, byBoardings)
}

// search is Dijkstra's algorithm over (stop, current route) states, comparing costs
// with less. It works because no link has a negative cost: once a state is popped,
// nothing found later can reach it more cheaply.
func (n *Network) search(from, to string, less func(a, b cost) bool) (Journey, error) {
	for _, s := range []string{from, to} {
		if !n.Has(s) {
			return Journey{}, fmt.Errorf("%w: %q", ErrUnknownStop, s)
		}
	}

	start := state{stop: from}
	best := map[state]cost{start: {}}
	prev := map[state]step{}
	done := map[state]bool{}
	q := &queue{{s: start, less: less}}
	seq := 1

	for q.Len() > 0 {
		cur := heap.Pop(q).(*item)
		if done[cur.s] {
			continue // a stale entry: this state was already settled more cheaply
		}
		done[cur.s] = true
		if cur.s.stop == to {
			return trace(cur.s, start, prev), nil
		}
		for _, e := range n.adj[cur.s.stop] {
			next := state{stop: e.to}
			c := cost{boardings: cur.c.boardings, minutes: cur.c.minutes + e.minutes}
			switch {
			case e.route == Walk:
				// on foot: no matatu
			case e.route == cur.s.route:
				next.route = e.route // stay on board
			default:
				next.route = e.route
				c.boardings++
			}
			if old, seen := best[next]; seen && !less(c, old) {
				continue
			}
			best[next] = c
			prev[next] = step{prev: cur.s, route: e.route, minutes: e.minutes}
			heap.Push(q, &item{s: next, c: c, seq: seq, less: less})
			seq++
		}
	}
	return Journey{}, fmt.Errorf("%w: %s to %s", ErrNoJourney, from, to)
}

// step is how a state was reached: from which state, on which route, in how long.
type step struct {
	prev    state
	route   string
	minutes int
}

// trace walks back from end to start and builds the journey.
func trace(end, start state, prev map[state]step) Journey {
	stops := []string{end.stop}
	routes := []string{}
	minutes := []int{}
	for s := end; s != start; s = prev[s].prev {
		p := prev[s]
		stops = append(stops, p.prev.stop)
		routes = append(routes, p.route)
		minutes = append(minutes, p.minutes)
	}
	slices.Reverse(stops)
	slices.Reverse(routes)
	slices.Reverse(minutes)
	// buildJourney wants the route and minutes into stop i at index i.
	return buildJourney(stops, append([]string{""}, routes...), append([]int{0}, minutes...))
}
