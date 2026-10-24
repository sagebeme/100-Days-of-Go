// Command showconfig loads the chama reminder service's config and prints it with the
// secrets hidden, or prints every problem and exits with status 1. It reads a .env
// file first if there is one: the path in its first argument, or ".env".
//
//	go run ./phase-2-modelling/day-021-config-loader/starter/cmd/showconfig phase-2-modelling/day-021-config-loader/.env
package main

import (
	"fmt"
	"os"

	configloader "github.com/sagebeme/100-Days-of-Go/phase-2-modelling/day-021-config-loader/starter"
)

func main() {
	path := ".env"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	fmt.Println("reading", path)

	// Your code here:
	//  1. os.Open(path). If errors.Is(err, os.ErrNotExist), carry on with no file values.
	//     Any other error: print it to os.Stderr and os.Exit(1).
	//  2. configloader.ParseDotEnv the file, then close it.
	//  3. configloader.Load(configloader.Layered(os.LookupEnv, configloader.FromMap(values)))
	//  4. On an error print "config:" and the error to os.Stderr and os.Exit(1).
	//     Otherwise print the config with fmt.Println(cfg): its String method hides
	//     the secrets.
	cfg, err := configloader.Load(os.LookupEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config:\n%v\n", err)
		os.Exit(1)
	}
	fmt.Println(cfg)
}
