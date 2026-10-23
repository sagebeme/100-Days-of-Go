// Package judge runs a kata program the way HackerRank does: feed each testdata/N.in on stdin,
// compare stdout with testdata/N.out, and fail anything slower than the time limit.
package judge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Result is one test case.
type Result struct {
	Name     string
	Passed   bool
	Elapsed  time.Duration
	Problem  string // why it failed: wrong line, timeout, crash
	TimedOut bool
}

// Cases lists the input files in dir/testdata, in natural order (2.in before 10.in).
func Cases(dir string) ([]string, error) {
	ins, err := filepath.Glob(filepath.Join(dir, "testdata", "*.in"))
	if err != nil {
		return nil, err
	}
	if len(ins) == 0 {
		return nil, fmt.Errorf("no test cases in %s", filepath.Join(dir, "testdata"))
	}
	sort.Slice(ins, func(i, j int) bool {
		a, b := filepath.Base(ins[i]), filepath.Base(ins[j])
		if len(a) != len(b) {
			return len(a) < len(b)
		}
		return a < b
	})
	return ins, nil
}

// Build compiles the main package in pkgDir into a binary inside outDir.
func Build(pkgDir, outDir, name string) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	bin := filepath.Join(outDir, name)
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = pkgDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed:\n%s", out)
	}
	return bin, nil
}

// Run feeds every case to bin and checks its output.
func Run(bin string, cases []string, limit time.Duration) []Result {
	results := make([]Result, 0, len(cases))
	for _, in := range cases {
		results = append(results, runOne(bin, in, limit))
	}
	return results
}

func runOne(bin, inPath string, limit time.Duration) Result {
	name := strings.TrimSuffix(filepath.Base(inPath), ".in")
	r := Result{Name: name}
	input, err := os.ReadFile(inPath)
	if err != nil {
		r.Problem = err.Error()
		return r
	}
	want, err := os.ReadFile(strings.TrimSuffix(inPath, ".in") + ".out")
	if err != nil {
		r.Problem = "missing expected output: " + err.Error()
		return r
	}

	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	start := time.Now()
	err = cmd.Run()
	r.Elapsed = time.Since(start)

	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		r.TimedOut = true
		r.Problem = fmt.Sprintf("took longer than %v", limit)
	case err != nil:
		r.Problem = fmt.Sprintf("crashed: %v\n%s", err, strings.TrimSpace(stderr.String()))
	default:
		r.Problem = Compare(stdout.String(), string(want))
		r.Passed = r.Problem == ""
	}
	return r
}

// Compare checks output line by line, ignoring trailing spaces and trailing blank lines,
// as HackerRank does. It returns "" when they match.
func Compare(got, want string) string {
	g, w := lines(got), lines(want)
	for i := 0; i < len(g) || i < len(w); i++ {
		var gl, wl string
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if i >= len(g) {
			return fmt.Sprintf("line %d: expected %q, but your output ended", i+1, wl)
		}
		if i >= len(w) {
			return fmt.Sprintf("line %d: expected no more output, got %q", i+1, gl)
		}
		if gl != wl {
			return fmt.Sprintf("line %d: expected %q, got %q", i+1, wl, gl)
		}
	}
	return ""
}

func lines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	parts := strings.Split(s, "\n")
	for i := range parts {
		parts[i] = strings.TrimRight(parts[i], " \t")
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
