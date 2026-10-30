// Command verify proves the course works: every day's solution passes that day's tests,
// every kata solution passes its test cases, and (with -starters) every untouched starter
// still fails, so the tests really check something.
//
//	go run ./tools/verify            solutions and katas
//	go run ./tools/verify -starters  also check that the starters fail
//
// Tests live only in each day's starter/ folder. To test a solution, verify copies the
// starter's tests next to the solution's code in .verify/run-*/ (gitignored) and runs them there.
// Each run gets its own folder, so two runs at once don't trip over each other.
//
// Phase 6 days are TypeScript: their tests are *.test.ts(x) files in starter/. For those, verify
// copies starter/ and then solution/ over it into phase-6-frontend/.verify/run-*/<day> and runs Vitest
// there, so it needs `npm ci` to have been run in phase-6-frontend first.
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
	if err := os.MkdirAll(filepath.Join(root, ".verify"), 0o755); err != nil {
		fail("%v", err)
	}
	scratch, err := os.MkdirTemp(filepath.Join(root, ".verify"), "run-")
	if err != nil {
		fail("%v", err)
	}
	cleanups := []string{scratch}
	defer func() {
		for _, dir := range cleanups {
			os.RemoveAll(dir)
		}
	}()

	problems := 0
	tested, katas, tsTested := 0, 0, 0
	var tsRoot, tsScratch string // the folder holding package.json, and this run's copies under it
	for _, d := range days {
		starter := filepath.Join(root, d.Dir, "starter")
		solution := filepath.Join(root, d.Dir, "solution")
		if isTS(starter) {
			if tsScratch == "" {
				tsRoot = filepath.Dir(filepath.Join(root, d.Dir))
				if err := os.MkdirAll(filepath.Join(tsRoot, ".verify"), 0o755); err != nil {
					fail("%v", err)
				}
				if tsScratch, err = os.MkdirTemp(filepath.Join(tsRoot, ".verify"), "run-"); err != nil {
					fail("%v", err)
				}
				cleanups = append(cleanups, tsScratch)
			}
			dst := filepath.Join(tsScratch, d.Name)
			if err := assembleTS(starter, solution, dst); err != nil {
				fmt.Printf("FAIL  day %d: %v\n", d.N, err)
				problems++
			} else {
				tsTested++
			}
			continue
		}
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
		rel, _ := filepath.Rel(root, filepath.Join(scratch, "days"))
		cmd := exec.Command("go", "test", "-count=1", "./"+filepath.ToSlash(rel)+"/...")
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("FAIL  solutions don't pass their tests:\n%s\n", out)
			problems++
		} else {
			fmt.Printf("ok    %d solutions pass their tests\n", tested)
		}
	}

	if tsTested > 0 {
		// Vitest strips types without checking them, so a starter or solution with a type
		// error would still run. tsc checks every day folder, starters included.
		tsc := exec.Command("npx", "tsc", "--noEmit")
		tsc.Dir = tsRoot
		if out, err := tsc.CombinedOutput(); err != nil {
			fmt.Printf("FAIL  TypeScript doesn't typecheck:\n%s\n", out)
			problems++
		} else {
			fmt.Println("ok    TypeScript typechecks, starters included")
		}

		rel, _ := filepath.Rel(tsRoot, tsScratch)
		cmd := exec.Command("npx", "vitest", "run", filepath.ToSlash(rel))
		cmd.Dir = tsRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("FAIL  TypeScript solutions don't pass their tests (did you run npm ci in %s?):\n%s\n", filepath.Base(tsRoot), out)
			problems++
		} else {
			fmt.Printf("ok    %d TypeScript solutions pass their tests\n", tsTested)
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
			var cmd *exec.Cmd
			if isTS(starter) {
				dayDir := filepath.Join(root, d.Dir)
				rel, _ := filepath.Rel(filepath.Dir(dayDir), filepath.Join(dayDir, "starter"))
				cmd = exec.Command("npx", "vitest", "run", filepath.ToSlash(rel))
				cmd.Dir = filepath.Dir(dayDir)
			} else {
				cmd = exec.Command("go", "test", "-count=1", "./"+d.Dir+"/starter")
				cmd.Dir = root
			}
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
		for _, dir := range cleanups {
			os.RemoveAll(dir) // os.Exit skips deferred calls
		}
		os.Exit(1)
	}
	fmt.Printf("\nverify passed: %d days, %d katas\n", tested+tsTested, katas)
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

// isTS reports whether a starter folder holds TypeScript tests.
func isTS(starter string) bool {
	found := false
	filepath.WalkDir(starter, func(path string, e os.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		if e.IsDir() && e.Name() == "node_modules" {
			return filepath.SkipDir
		}
		name := e.Name()
		if strings.HasSuffix(name, ".test.ts") || strings.HasSuffix(name, ".test.tsx") {
			found = true
		}
		return nil
	})
	return found
}

// assembleTS copies the whole starter, then the whole solution over it, into dst.
func assembleTS(starter, solution, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.CopyFS(dst, os.DirFS(starter)); err != nil {
		return err
	}
	return filepath.WalkDir(solution, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(solution, path)
		target := filepath.Join(dst, rel)
		if e.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
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
