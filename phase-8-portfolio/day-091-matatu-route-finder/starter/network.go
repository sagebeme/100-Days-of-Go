// Package routes finds journeys on a matatu network: the fastest one with Dijkstra's
// algorithm, and the one with the fewest changes of matatu.
package routes

import "errors"

// Walk is the route name for walking between two stops.
const Walk = "walk"

// Link joins two stops in both directions. Route is the matatu route that runs
// between them, or Walk.
type Link struct {
	A, B    string
	Route   string
	Minutes int
}

// Network is a set of stops and the links between them.
type Network struct {
	// Your fields here: an adjacency list (each stop's outgoing edges), and the stop
	// names in the order they were first seen.
}

// Errors.
var (
	ErrUnknownStop = errors.New("unknown stop")
	ErrNoJourney   = errors.New("no journey between those stops")
)

// NewNetwork builds a network. Every link needs two different, non-empty stop names,
// a route, and Minutes above 0.
func NewNetwork(links []Link) (*Network, error) {
	// Your code here
	return &Network{}, nil
}

// Stops returns every stop name, in the order they were first seen.
func (n *Network) Stops() []string {
	// Your code here
	return nil
}

// Has reports whether the network has a stop with that name.
func (n *Network) Has(stop string) bool {
	// Your code here
	return false
}

// Leg is one ride (or walk) on a single route, possibly past several stops.
type Leg struct {
	Route    string
	From, To string
	Minutes  int
}

// Journey is a way from one stop to another.
type Journey struct {
	Stops     []string // every stop passed, first and last included
	Legs      []Leg    // consecutive links on the same route are one leg
	Minutes   int
	Transfers int // matatu legs minus one; walking is not a matatu
}

// Describe writes a journey as directions, one leg per line:
//
//	Walk from Kencom to Railways (8 min)
//	Take the 111 from Railways to Ngong (70 min)
//	Total: 78 min, 0 transfers
func Describe(j Journey) string {
	// Your code here
	return ""
}
