// Package judge runs a kata program the way HackerRank does: feed each testdata/N.in on stdin,
// compare stdout with testdata/N.out, and fail anything slower than the time limit.
//
// A full-size input can be megabytes, too big to commit, so a kata can instead ship kata/gen: a
// program that prints one. BigCase builds that input and the reference solution's answer to it at
// judge time, and gen/big.sha256 records what the answer must be.
package judge

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// HasBigCase reports whether the kata in dir ships a generator for a full-size case.
func HasBigCase(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "gen", "main.go"))
	return err == nil
}

// BigCase builds dir/gen, writes its input to outDir/big.in, and runs the reference solution
// (already built as solutionBin) on it to write outDir/big.out. It returns the path of big.in,
// ready to pass to Run with the other cases. The solution gets a generous limit here: timing it
// is Run's job.
func BigCase(dir, solutionBin, outDir string) (string, error) {
	gen, err := Build(filepath.Join(dir, "gen"), outDir, "gen")
	if err != nil {
		return "", fmt.Errorf("kata/gen: %w", err)
	}
	inPath := filepath.Join(outDir, "big.in")
	input, err := exec.Command(gen).Output()
	if err != nil {
		return "", fmt.Errorf("kata/gen failed: %w", err)
	}
	if err := os.WriteFile(inPath, input, 0o644); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, solutionBin)
	cmd.Stdin = bytes.NewReader(input)
	want, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("the reference solution failed on the big case: %w", err)
	}
	return inPath, os.WriteFile(filepath.Join(outDir, "big.out"), want, 0o644)
}

// Fingerprint is the SHA-256 of output as Compare sees it: no trailing spaces, no trailing blank
// lines, \n line endings. gen/big.sha256 holds the fingerprint of the big case's right answer.
func Fingerprint(output []byte) string {
	sum := sha256.Sum256([]byte(strings.Join(lines(string(output)), "\n")))
	return hex.EncodeToString(sum[:])
}

// CheckFingerprint compares outDir/big.out, as written by BigCase, with dir/gen/big.sha256.
func CheckFingerprint(dir, outDir string) error {
	want, err := os.ReadFile(filepath.Join(dir, "gen", "big.sha256"))
	if err != nil {
		return fmt.Errorf("kata/gen has no big.sha256: %w", err)
	}
	got, err := os.ReadFile(filepath.Join(outDir, "big.out"))
	if err != nil {
		return err
	}
	if fp := Fingerprint(got); fp != strings.TrimSpace(string(want)) {
		return fmt.Errorf("the big case's answer has fingerprint %s, but gen/big.sha256 says %s: the generator or the solution changed", fp, strings.TrimSpace(string(want)))
	}
	return nil
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
