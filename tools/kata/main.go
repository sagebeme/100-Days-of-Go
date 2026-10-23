// Command kata judges a day's kata the way HackerRank does.
//
//	go run ./tools/kata 3             judge your kata/main.go for Day 3
//	go run ./tools/kata -solution 3   judge the reference solution instead
//	go run ./tools/kata -limit 1s 3   change the time limit per test case (default 2s)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sagebeme/100-Days-of-Go/tools/internal/course"
	"github.com/sagebeme/100-Days-of-Go/tools/internal/judge"
)

func main() {
	solution := flag.Bool("solution", false, "judge kata/solution instead of your kata/main.go")
	limit := flag.Duration("limit", 2*time.Second, "time limit per test case")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/kata [-solution] [-limit 2s] <day>")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	n, err := strconv.Atoi(strings.TrimLeft(strings.TrimPrefix(flag.Arg(0), "day-"), "0"))
	if err != nil {
		fail("%q isn't a day number. Try: go run ./tools/kata 3", flag.Arg(0))
	}

	root, err := course.Root()
	if err != nil {
		fail("%v", err)
	}
	day, err := course.Find(root, n)
	if err != nil {
		fail("%v", err)
	}
	kataDir := filepath.Join(root, day.Dir, "kata")
	if !course.Exists(kataDir) {
		fail("Day %d has no kata", n)
	}
	pkg := kataDir
	if *solution {
		pkg = filepath.Join(kataDir, "solution")
	}
	cases, err := judge.Cases(kataDir)
	if err != nil {
		fail("%v", err)
	}
	bin, err := judge.Build(pkg, filepath.Join(root, ".verify", "bin"), day.Name+"-kata")
	if err != nil {
		fail("%v", err)
	}

	passed := 0
	var total time.Duration
	for _, r := range judge.Run(bin, cases, *limit) {
		total += r.Elapsed
		if r.Passed {
			passed++
			fmt.Printf("  PASS  case %-3s %6s\n", r.Name, r.Elapsed.Round(time.Millisecond))
			continue
		}
		fmt.Printf("  FAIL  case %-3s %6s  %s\n", r.Name, r.Elapsed.Round(time.Millisecond), r.Problem)
	}
	fmt.Println()
	if passed != len(cases) {
		fmt.Printf("%d of %d cases passed. Read the first failure, fix it, run again.\n", passed, len(cases))
		os.Exit(1)
	}
	fmt.Printf("All %d cases passed in %s. Kata done.\n", passed, total.Round(time.Millisecond))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
