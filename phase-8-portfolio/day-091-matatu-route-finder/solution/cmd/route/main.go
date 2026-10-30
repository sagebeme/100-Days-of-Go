// Command route prints directions between two stops on the Nairobi network.
//
//	go run ./phase-8-portfolio/day-091-matatu-route-finder/solution/cmd/route Kawangware Ngong
//	go run ./phase-8-portfolio/day-091-matatu-route-finder/solution/cmd/route -fewest Kawangware Ngong
//
// The connections and times are simplified. Routes change, so confirm with the conductor.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	routes "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-091-matatu-route-finder/solution"
)

func main() {
	fewest := flag.Bool("fewest", false, "fewest transfers instead of fastest")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: route [-fewest] <from> <to>")
		fmt.Fprintln(os.Stderr, "stops:", strings.Join(routes.Nairobi().Stops(), ", "))
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	n := routes.Nairobi()
	find := n.Fastest
	if *fewest {
		find = n.FewestTransfers
	}
	j, err := find(flag.Arg(0), flag.Arg(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(routes.Describe(j))
	fmt.Println("Times are simplified. Routes change, so confirm with the conductor.")
}
