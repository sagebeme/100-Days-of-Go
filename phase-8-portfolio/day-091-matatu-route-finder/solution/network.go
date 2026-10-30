// Package routes finds journeys on a matatu network: the fastest one with Dijkstra's
// algorithm, and the one with the fewest changes of matatu.
package routes

import (
	"errors"
	"fmt"
	"strings"
)

// Walk is the route name for walking between two stops.
const Walk = "walk"

// Link joins two stops in both directions. Route is the matatu route that runs
// between them, or Walk.
type Link struct {
	A, B    string
	Route   string
	Minutes int
}

// edge is one direction of a Link.
type edge struct {
	to      string
	route   string
	minutes int
}

// Network is a set of stops and the links between them.
type Network struct {
	adj   map[string][]edge
	stops []string // in the order they were first seen
}

// Errors.
var (
	ErrUnknownStop = errors.New("unknown stop")
	ErrNoJourney   = errors.New("no journey between those stops")
)

// NewNetwork builds a network. Every link needs two different, non-empty stop names,
// a route, and Minutes above 0.
func NewNetwork(links []Link) (*Network, error) {
	n := &Network{adj: map[string][]edge{}}
	for i, l := range links {
		if l.A == "" || l.B == "" || l.A == l.B || l.Route == "" || l.Minutes <= 0 {
			return nil, fmt.Errorf("link %d (%s-%s): want two different stops, a route and minutes above 0", i, l.A, l.B)
		}
		for _, s := range []string{l.A, l.B} {
			if _, ok := n.adj[s]; !ok {
				n.adj[s] = nil
				n.stops = append(n.stops, s)
			}
		}
		n.adj[l.A] = append(n.adj[l.A], edge{l.B, l.Route, l.Minutes})
		n.adj[l.B] = append(n.adj[l.B], edge{l.A, l.Route, l.Minutes})
	}
	return n, nil
}

// Stops returns every stop name, in the order they were first seen.
func (n *Network) Stops() []string { return append([]string(nil), n.stops...) }

// Has reports whether the network has a stop with that name.
func (n *Network) Has(stop string) bool {
	_, ok := n.adj[stop]
	return ok
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

// buildJourney turns a list of stops and the route used into each one into legs.
func buildJourney(stops []string, routes []string, minutes []int) Journey {
	j := Journey{Stops: stops}
	for i := 1; i < len(stops); i++ {
		j.Minutes += minutes[i]
		if k := len(j.Legs) - 1; k >= 0 && j.Legs[k].Route == routes[i] {
			j.Legs[k].To = stops[i]
			j.Legs[k].Minutes += minutes[i]
			continue
		}
		j.Legs = append(j.Legs, Leg{Route: routes[i], From: stops[i-1], To: stops[i], Minutes: minutes[i]})
	}
	rides := 0
	for _, l := range j.Legs {
		if l.Route != Walk {
			rides++
		}
	}
	j.Transfers = max(0, rides-1)
	return j
}

// Describe writes a journey as directions, one leg per line:
//
//	Walk from Kencom to Railways (8 min)
//	Take the 111 from Railways to Ngong (70 min)
//	Total: 78 min, 0 transfers
func Describe(j Journey) string {
	var b strings.Builder
	for _, l := range j.Legs {
		if l.Route == Walk {
			fmt.Fprintf(&b, "Walk from %s to %s (%d min)\n", l.From, l.To, l.Minutes)
		} else {
			fmt.Fprintf(&b, "Take the %s from %s to %s (%d min)\n", l.Route, l.From, l.To, l.Minutes)
		}
	}
	word := "transfers"
	if j.Transfers == 1 {
		word = "transfer"
	}
	fmt.Fprintf(&b, "Total: %d min, %d %s", j.Minutes, j.Transfers, word)
	return b.String()
}
