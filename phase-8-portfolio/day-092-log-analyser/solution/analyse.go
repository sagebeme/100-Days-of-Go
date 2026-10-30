package logstat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
	"text/tabwriter"
	"time"
)

// MaxTracked is the slowest latency the histogram tells apart. Anything slower is
// counted as MaxTracked.
const MaxTracked = 60 * time.Second

// Histogram counts latencies in 1 ms buckets from 0 to MaxTracked. It uses the same
// 480 KB however many lines it sees, and its percentiles are exact to the
// millisecond, which is all the log records anyway.
type Histogram struct {
	buckets [MaxTracked/time.Millisecond + 1]int64
	n       int64
}

// Add counts one latency.
func (h *Histogram) Add(d time.Duration) {
	ms := min(max(d/time.Millisecond, 0), MaxTracked/time.Millisecond)
	h.buckets[ms]++
	h.n++
}

// Count is how many latencies were added.
func (h *Histogram) Count() int64 { return h.n }

// Percentile returns the p-th percentile (0 < p ≤ 100) by the nearest-rank method:
// the smallest latency that at least p% of all latencies are at or below. With no
// data it returns 0.
func (h *Histogram) Percentile(p float64) time.Duration {
	if h.n == 0 || p <= 0 {
		return 0
	}
	// rank = ceil(p% of n). The tiny nudge stops 99.9 * 1000 / 100, which is a hair
	// over 999 in floating point, from rounding up to 1000.
	rank := int64(math.Ceil(p*float64(h.n)/100 - 1e-9))
	rank = min(max(rank, 1), h.n)
	var seen int64
	for ms, c := range h.buckets {
		seen += c
		if seen >= rank {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return MaxTracked
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
	n := 0
	for code, count := range r.Status {
		if code/100 == c {
			n += count
		}
	}
	return n
}

// Analyse reads a log line by line and reports on it. Blank lines are skipped.
// Malformed lines, including lines over MaxLineBytes, are counted and skipped, never
// fatal. Only a read error stops it. topN is how many paths to keep.
func Analyse(r io.Reader, topN int) (Report, error) {
	rep := Report{Status: map[int]int{}}
	paths := map[string]int{}
	var hist Histogram
	br := bufio.NewReaderSize(r, MaxLineBytes)

	for lineNo := 1; ; lineNo++ {
		line, tooLong, err := readLine(br)
		if err == io.EOF && line == "" && !tooLong {
			break
		}
		if err != nil && err != io.EOF {
			return rep, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if strings.TrimSpace(line) != "" || tooLong {
			rep.Lines++
			e, perr := ParseLine(line)
			if tooLong || perr != nil {
				rep.Malformed++
				if len(rep.BadLines) < MaxBadLines {
					rep.BadLines = append(rep.BadLines, lineNo)
				}
			} else {
				rep.Parsed++
				rep.Status[e.Status]++
				paths[e.Path]++
				rep.Bytes += e.Bytes
				hist.Add(e.Latency)
				if rep.First.IsZero() || e.Time.Before(rep.First) {
					rep.First = e.Time
				}
				if e.Time.After(rep.Last) {
					rep.Last = e.Time
				}
			}
		}
		if err == io.EOF {
			break
		}
	}

	for p, c := range paths {
		rep.TopPaths = append(rep.TopPaths, PathCount{p, c})
	}
	slices.SortFunc(rep.TopPaths, func(a, b PathCount) int {
		if a.Count != b.Count {
			return b.Count - a.Count
		}
		return strings.Compare(a.Path, b.Path)
	})
	if len(rep.TopPaths) > topN {
		rep.TopPaths = rep.TopPaths[:max(topN, 0)]
	}
	rep.P50, rep.P95, rep.P99 = hist.Percentile(50), hist.Percentile(95), hist.Percentile(99)
	return rep, nil
}

// readLine returns the next line without its line ending. If the line is longer than
// the reader's buffer, it reads and throws away the rest and reports tooLong. At the
// end of the input it returns io.EOF, possibly with a last line that had no newline.
func readLine(br *bufio.Reader) (line string, tooLong bool, err error) {
	b, err := br.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) {
		for errors.Is(err, bufio.ErrBufferFull) {
			_, err = br.ReadSlice('\n')
		}
		if err != nil && err != io.EOF {
			return "", true, err
		}
		return "", true, err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(string(b), "\n"), "\r")
	return line, false, err
}

// Write prints the report as a small table: counts, the time range, bytes, each
// status class, p50, p95, p99 and the top paths.
func (r Report) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "lines\t%d\n", r.Lines)
	fmt.Fprintf(tw, "parsed\t%d\n", r.Parsed)
	fmt.Fprintf(tw, "malformed\t%d", r.Malformed)
	if len(r.BadLines) > 0 {
		fmt.Fprintf(tw, " (first at lines %s)", joinInts(r.BadLines))
	}
	fmt.Fprintln(tw)
	if !r.First.IsZero() {
		fmt.Fprintf(tw, "from\t%s\n", r.First.Format(time.RFC3339))
		fmt.Fprintf(tw, "to\t%s\n", r.Last.Format(time.RFC3339))
	}
	fmt.Fprintf(tw, "bytes sent\t%d\n", r.Bytes)
	for c := 1; c <= 5; c++ {
		fmt.Fprintf(tw, "%dxx\t%d\n", c, r.Class(c))
	}
	fmt.Fprintf(tw, "p50\t%s\n", r.P50)
	fmt.Fprintf(tw, "p95\t%s\n", r.P95)
	fmt.Fprintf(tw, "p99\t%s\n", r.P99)
	fmt.Fprintln(tw, "top paths\t")
	for _, p := range r.TopPaths {
		fmt.Fprintf(tw, "  %d\t%s\n", p.Count, p.Path)
	}
	return tw.Flush()
}

func joinInts(ns []int) string {
	s := make([]string, len(ns))
	for i, n := range ns {
		s[i] = fmt.Sprint(n)
	}
	return strings.Join(s, ", ")
}
