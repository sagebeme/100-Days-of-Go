// Command logstat summarises an access log.
//
//	go run ./phase-8-portfolio/day-092-log-analyser/starter/cmd/logstat -top 5 access.log
//	zcat access.log.gz | go run ./phase-8-portfolio/day-092-log-analyser/starter/cmd/logstat -
//
// Exit codes: 0 fine, 1 the file couldn't be read, 2 bad usage, 3 the -max-5xx
// threshold was passed (handy in a cron job or CI check).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	logstat "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-092-log-analyser/starter"
)

func main() {
	top := flag.Int("top", 10, "how many paths to list")
	asJSON := flag.Bool("json", false, "print the report as JSON")
	max5xx := flag.Float64("max-5xx", 100, "exit 3 if more than this percent of parsed lines were 5xx")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: logstat [flags] <file | ->")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 || *top < 0 {
		flag.Usage()
		os.Exit(2)
	}

	var in io.Reader = os.Stdin
	if name := flag.Arg(0); name != "-" {
		f, err := os.Open(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	rep, err := logstat.Analyse(in, *top)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(rep)
	} else {
		rep.Write(os.Stdout)
	}
	if rep.Parsed > 0 && float64(rep.Class(5))*100/float64(rep.Parsed) > *max5xx {
		os.Exit(3)
	}
}
