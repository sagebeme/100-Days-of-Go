package logstat

import (
	"io"
	"time"
)

// MaxTracked is the slowest latency the histogram tells apart. Anything slower is
// counted as MaxTracked.
const MaxTracked = 60 * time.Second

// Histogram counts latencies in 1 ms buckets from 0 to MaxTracked. It uses the same
// 480 KB however many lines it sees, and its percentiles are exact to the
// millisecond, which is all the log records anyway.
type Histogram struct {
	// Your fields here: an array of MaxTracked/time.Millisecond + 1 counts, and a total.
}

// Add counts one latency.
func (h *Histogram) Add(d time.Duration) {
	// Your code here
}

// Count is how many latencies were added.
func (h *Histogram) Count() int64 {
	// Your code here
	return 0
}

// Percentile returns the p-th percentile (0 < p ≤ 100) by the nearest-rank method:
// the smallest latency that at least p% of all latencies are at or below. With no
// data it returns 0.
func (h *Histogram) Percentile(p float64) time.Duration {
	// Your code here
	return 0
}

// PathCount is one row of the busiest paths.
type PathCount struct {
	Path  string
	Count int
}

// Report is what Analyse finds.
type Report struct {
	Lines       int   // non-blank lines read
	Parsed      int   // lines that parsed
	Malformed   int   // lines that didn't
	BadLines    []int // line numbers of the first MaxBadLines malformed lines
	Status      map[int]int
	TopPaths    []PathCount // busiest first; ties by path
	Bytes       int64
	First, Last time.Time // earliest and latest timestamps
	P50, P95    time.Duration
	P99         time.Duration
}

// MaxBadLines is how many malformed line numbers a Report keeps.
const MaxBadLines = 5

// MaxLineBytes is the longest line Analyse will parse. Longer lines are skipped as
// malformed without being held in memory.
const MaxLineBytes = 64 << 10

// Class returns how many responses had a status in the hundreds class c (2 for 2xx).
func (r Report) Class(c int) int {
	// Your code here
	return 0
}

// Analyse reads a log line by line and reports on it. Blank lines are skipped.
// Malformed lines, including lines over MaxLineBytes, are counted and skipped, never
// fatal. Only a read error stops it. topN is how many paths to keep.
func Analyse(r io.Reader, topN int) (Report, error) {
	// Your code here
	return Report{}, nil
}

// Write prints the report as a small table: counts, the time range, bytes, each
// status class, p50, p95, p99 and the top paths.
func (r Report) Write(w io.Writer) error {
	// Your code here
	return nil
}
