// Package graphs explores a map of stops with breadth-first and depth-first search.
package graphs

// Graph is an undirected graph of stops. The zero value is an empty graph
// ready to use.
type Graph struct {
	adj map[string][]string // stop → the stops one ride or walk away
}

// AddRoute connects stops a and b in both directions. Adding the same
// connection twice has no effect, and a stop is never connected to itself.
func (g *Graph) AddRoute(a, b string) {
	// Your code here.
}

// Neighbours returns the stops directly connected to stop, in alphabetical
// order. A stop that isn't on the map has none.
func (g *Graph) Neighbours(stop string) []string {
	// Your code here.
	return nil
}

// FewestHops returns the smallest number of connections needed to get from
// one stop to another, and true. If to can't be reached, or either stop isn't
// on the map, it returns 0, false. From a stop to itself is 0 hops, if the
// stop is on the map. Use breadth-first search.
func (g *Graph) FewestHops(from, to string) (int, bool) {
	// Your code here.
	return 0, false
}

// Path returns the stops on a shortest route from one stop to another,
// including both ends, or nil if there's no route. When several routes are
// equally short, it returns the one a BFS finds when it visits neighbours in
// alphabetical order.
func (g *Graph) Path(from, to string) []string {
	// Your code here.
	return nil
}

// Reachable returns every stop you can get to from start, including start,
// in alphabetical order. A start that isn't on the map gives an empty
// (non-nil) slice. Use depth-first search.
func (g *Graph) Reachable(start string) []string {
	// Your code here.
	return nil
}

// CountNetworks returns the number of separate networks on the map: groups of
// stops connected to each other but not to any other group.
func (g *Graph) CountNetworks() int {
	// Your code here.
	return 0
}
