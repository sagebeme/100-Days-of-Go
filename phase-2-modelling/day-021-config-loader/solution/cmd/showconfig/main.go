// Command showconfig loads the chama reminder service's config and prints it with the
// secrets hidden, or prints every problem and exits with status 1. It reads a .env
// file first if there is one: the path in its first argument, or ".env".
//
//	go run ./phase-2-modelling/day-021-config-loader/solution/cmd/showconfig phase-2-modelling/day-021-config-loader/.env
package main

import (
	"errors"
	"fmt"
	"os"

	configloader "github.com/sagebeme/100-Days-of-Go/phase-2-modelling/day-021-config-loader/solution"
)

func main() {
	path := ".env"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	fileValues := map[string]string{}
	f, err := os.Open(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// No .env file: fine, use the real environment only.
	case err != nil:
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	default:
		fileValues, err = configloader.ParseDotEnv(f)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config: %s: %v\n", path, err)
			os.Exit(1)
		}
	}

	cfg, err := configloader.Load(configloader.Layered(os.LookupEnv, configloader.FromMap(fileValues)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "config:\n%v\n", err)
		os.Exit(1)
	}
	fmt.Println(cfg)
}
