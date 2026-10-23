// Command verify proves the course works: every day's solution passes that day's tests,
// every kata solution passes its test cases, and (with -starters) every untouched starter
// still fails, so the tests really check something.
//
//	go run ./tools/verify            solutions and katas
//	go run ./tools/verify -starters  also check that the starters fail
//
// Tests live only in each day's starter/ folder. To test a solution, verify copies the
// starter's tests next to the solution's code in .verify/ (gitignored) and runs them there.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sagebeme/100-Days-of-Go/tools/internal/course"
	"github.com/sagebeme/100-Days-of-Go/tools/internal/judge"
)

func main() {
	starters := flag.Bool("starters", false, "also check that every untouched starter fails its tests")
	flag.Parse()

	root, err := course.Root()
	if err != nil {
		fail("%v", err)
	}
	days, err := course.Days(root)
	if err != nil {
		fail("%v", err)
	}
	scratch := filepath.Join(root, ".verify")
	if err := os.RemoveAll(filepath.Join(scratch, "days")); err != nil {
		fail("%v", err)
	}

	problems := 0
	tested, katas := 0, 0
	for _, d := range days {
		starter := filepath.Join(root, d.Dir, "starter")
		solution := filepath.Join(root, d.Dir, "solution")
		if course.Exists(starter) && course.Exists(solution) {
			dst := filepath.Join(scratch, "days", d.Name)
			if err := assemble(starter, solution, dst); err != nil {
				fmt.Printf("FAIL  day %d: %v\n", d.N, err)
				problems++
			} else {
				tested++
			}
		}
	}

	if tested > 0 {
		cmd := exec.Command("go", "test", "-count=1", "./.verify/days/...")
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("FAIL  solutions don't pass their tests:\n%s\n", out)
			problems++
		} else {
			fmt.Printf("ok    %d solutions pass their tests\n", tested)
		}
	}

	for _, d := range days {
		kata := filepath.Join(root, d.Dir, "kata")
		if !course.Exists(filepath.Join(kata, "solution")) {
			continue
		}
		katas++
		cases, err := judge.Cases(kata)
		if err != nil {
			fmt.Printf("FAIL  day %d kata: %v\n", d.N, err)
			problems++
			continue
		}
		bin, err := judge.Build(filepath.Join(kata, "solution"), filepath.Join(scratch, "bin"), d.Name+"-kata-solution")
		if err != nil {
			fmt.Printf("FAIL  day %d kata: %v\n", d.N, err)
			problems++
			continue
		}
		for _, r := range judge.Run(bin, cases, 2*time.Second) {
			if !r.Passed {
				fmt.Printf("FAIL  day %d kata case %s: %s\n", d.N, r.Name, r.Problem)
				problems++
			}
		}
	}
	fmt.Printf("ok    %d kata solutions checked\n", katas)

	if *starters {
		for _, d := range days {
			starter := filepath.Join(root, d.Dir, "starter")
			if !course.Exists(starter) {
				continue
			}
			cmd := exec.Command("go", "test", "-count=1", "./"+d.Dir+"/starter")
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			if err == nil {
				fmt.Printf("FAIL  day %d: the untouched starter passes its tests, so they check nothing\n", d.N)
				problems++
			} else if strings.Contains(string(out), "[build failed]") || strings.Contains(string(out), "[setup failed]") {
				fmt.Printf("FAIL  day %d: the starter doesn't compile; learners should start from code that builds:\n%s\n", d.N, out)
				problems++
			}
		}
		fmt.Println("ok    starters checked: each one compiles and fails its tests")
	}

	if problems > 0 {
		fmt.Printf("\n%d problem(s).\n", problems)
		os.Exit(1)
	}
	fmt.Printf("\nverify passed: %d days, %d katas\n", tested, katas)
}

// assemble copies the starter's tests (and testdata) and the solution's code into dst.
func assemble(starter, solution, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	tests, _ := filepath.Glob(filepath.Join(starter, "*_test.go"))
	if len(tests) == 0 {
		return fmt.Errorf("no tests in %s", starter)
	}
	for _, f := range tests {
		if err := copyFile(f, filepath.Join(dst, filepath.Base(f))); err != nil {
			return err
		}
	}
	code, _ := filepath.Glob(filepath.Join(solution, "*.go"))
	for _, f := range code {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		if err := copyFile(f, filepath.Join(dst, filepath.Base(f))); err != nil {
			return err
		}
	}
	if td := filepath.Join(starter, "testdata"); course.Exists(td) {
		return os.CopyFS(filepath.Join(dst, "testdata"), os.DirFS(td))
	}
	return nil
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
