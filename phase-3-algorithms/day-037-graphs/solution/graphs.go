// Package graphs explores a map of stops with breadth-first and depth-first search.
package graphs

import (
	"slices"
	"sort"
)

// Graph is an undirected graph of stops.
type Graph struct {
	adj map[string][]string
}

// AddRoute connects stops a and b in both directions.
func (g *Graph) AddRoute(a, b string) {
	if a == b {
		return
	}
	if g.adj == nil {
		g.adj = map[string][]string{}
	}
	if slices.Contains(g.adj[a], b) {
		return
	}
	g.adj[a] = append(g.adj[a], b)
	g.adj[b] = append(g.adj[b], a)
}

// Neighbours returns the stops directly connected to stop, alphabetically.
func (g *Graph) Neighbours(stop string) []string {
	out := slices.Clone(g.adj[stop])
	sort.Strings(out)
	return out
}

// bfs walks outwards from start, one ring of stops at a time, and returns the
// stop each one was first reached from. start maps to "".
func (g *Graph) bfs(start string) map[string]string {
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
	return cameFrom
}

// Path returns the stops on a shortest route, or nil.
func (g *Graph) Path(from, to string) []string {
	if _, ok := g.adj[from]; !ok {
		return nil
	}
	cameFrom := g.bfs(from)
	if _, ok := cameFrom[to]; !ok {
		return nil
	}
	var path []string
	for stop := to; stop != ""; stop = cameFrom[stop] {
		path = append(path, stop)
	}
	slices.Reverse(path) // we walked it backwards, from to to from
	return path
}

// FewestHops returns the number of connections on a shortest route.
func (g *Graph) FewestHops(from, to string) (int, bool) {
	path := g.Path(from, to)
	if path == nil {
		return 0, false
	}
	return len(path) - 1, true
}

// Reachable returns every stop reachable from start, alphabetically.
func (g *Graph) Reachable(start string) []string {
	if _, ok := g.adj[start]; !ok {
		return []string{}
	}
	seen := map[string]bool{}
	var visit func(stop string)
	visit = func(stop string) {
		seen[stop] = true
		for _, next := range g.adj[stop] {
			if !seen[next] {
				visit(next)
			}
		}
	}
	visit(start)
	out := make([]string, 0, len(seen))
	for stop := range seen {
		out = append(out, stop)
	}
	sort.Strings(out)
	return out
}

// CountNetworks returns the number of connected groups of stops.
func (g *Graph) CountNetworks() int {
	seen := map[string]bool{}
	count := 0
	for stop := range g.adj {
		if seen[stop] {
			continue
		}
		count++ // a stop nothing so far has reached: a new network
		stack := []string{stop}
		seen[stop] = true
		for len(stack) > 0 { // depth-first, with an explicit stack this time
			cur := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, next := range g.adj[cur] {
				if !seen[next] {
					seen[next] = true
					stack = append(stack, next)
				}
			}
		}
	}
	return count
}
