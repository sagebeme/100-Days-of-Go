package routes

// Fastest returns the journey with the fewest minutes, using Dijkstra's algorithm
// with container/heap. Among equally fast journeys it picks the one with fewer
// transfers. A journey from a stop to itself has one stop and no legs.
func (n *Network) Fastest(from, to string) (Journey, error) {
	// Your code here
	return Journey{}, nil
}

// FewestTransfers returns the journey with the fewest matatu changes. Among those it
// picks the fastest.
func (n *Network) FewestTransfers(from, to string) (Journey, error) {
	// Your code here
	return Journey{}, nil
}
