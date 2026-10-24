// Command chama splits contributions and works out loans for a chama.
//
//	go run ./phase-2-modelling/day-022-cli-flags/solution/cmd/chama split -total 12005 -members 5
package main

import (
	"os"

	cliflags "github.com/sagebeme/100-Days-of-Go/phase-2-modelling/day-022-cli-flags/solution"
)

func main() {
	os.Exit(cliflags.Run(os.Args[1:], os.Stdout, os.Stderr))
}
