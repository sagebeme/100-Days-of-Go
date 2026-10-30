package logstat

import (
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

var eat = time.FixedZone("", 3*60*60)

func TestParseLine(t *testing.T) {
	line := `105.163.1.20 - kamau [24/Oct/2026:09:00:05 +0300] "GET /statements/2026-10?page=2 HTTP/2.0" 404 - "https://pesa.example/home" "curl/8.5.0" 0.0035`
	got, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine: %v", err)
	}
	want := Entry{
		IP: "105.163.1.20", User: "kamau",
		Time:   time.Date(2026, 10, 24, 9, 0, 5, 0, eat),
		Method: "GET", Path: "/statements/2026-10", Proto: "HTTP/2.0",
		Status: 404, Bytes: 0,
		Referer: "https://pesa.example/home", UserAgent: "curl/8.5.0",
		Latency: 4 * time.Millisecond,
	}
	if !got.Time.Equal(want.Time) {
		t.Errorf("Time = %v, want %v", got.Time, want.Time)
	}
	got.Time = want.Time
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseLine =\n%+v\nwant\n%+v", got, want)
	}

	ua, err := ParseLine(`41.90.64.7 - - [24/Oct/2026:09:00:02 +0300] "GET /balance HTTP/1.1" 200 88 "-" "Mozilla/5.0 (Linux; Android 14)" 0.050`)
	if err != nil || ua.UserAgent != "Mozilla/5.0 (Linux; Android 14)" || ua.Bytes != 88 || ua.Latency != 50*time.Millisecond {
		t.Errorf("a user agent with spaces: got %+v, %v", ua, err)
	}
}

func TestParseLineMalformed(t *testing.T) {
	good := `41.90.64.7 - - [24/Oct/2026:09:00:01 +0300] "POST /stk/push HTTP/1.1" 200 312 "-" "okhttp/4.12.0" 0.120`
	bad := map[string]string{
		"empty":            "",
		"junk":             "this line is not a log line at all",
		"no time":          `41.90.64.7 - - "POST /stk/push HTTP/1.1" 200 312 "-" "ua" 0.1`,
		"bad date":         strings.Replace(good, "24/Oct", "31/Feb", 1),
		"short request":    strings.Replace(good, `"POST /stk/push HTTP/1.1"`, `"-"`, 1),
		"status too big":   strings.Replace(good, " 200 ", " 999 ", 1),
		"status not a num": strings.Replace(good, " 200 ", " OK ", 1),
		"bad bytes":        strings.Replace(good, " 312 ", " lots ", 1),
		"no user agent":    strings.TrimSuffix(good, ` "okhttp/4.12.0" 0.120`),
		"no latency":       strings.TrimSuffix(good, " 0.120"),
		"negative latency": strings.Replace(good, "0.120", "-1", 1),
		"cut off":          good[:40],
	}
	for name, line := range bad {
		if _, err := ParseLine(line); !errors.Is(err, ErrMalformed) {
			t.Errorf("%s: ParseLine error = %v, want ErrMalformed", name, err)
		}
	}
}

func TestHistogram(t *testing.T) {
	var h Histogram
	if h.Percentile(95) != 0 {
		t.Errorf("an empty histogram's p95 = %v, want 0", h.Percentile(95))
	}
	for i := 1; i <= 20; i++ {
		h.Add(time.Duration(i) * time.Millisecond)
	}
	h.Add(2 * time.Minute) // over MaxTracked: counted as MaxTracked
	if h.Count() != 21 {
		t.Errorf("Count = %d, want 21", h.Count())
	}
	tests := []struct {
		p    float64
		want time.Duration
	}{
		{50, 11 * time.Millisecond}, // rank ceil(10.5) = 11
		{95, 20 * time.Millisecond}, // rank ceil(19.95) = 20
		{99, MaxTracked},            // rank ceil(20.79) = 21
		{100, MaxTracked},           // the largest
		{1, 1 * time.Millisecond},   // rank ceil(0.21) = 1
	}
	for _, tt := range tests {
		if got := h.Percentile(tt.p); got != tt.want {
			t.Errorf("Percentile(%v) = %v, want %v", tt.p, got, tt.want)
		}
	}

	var exact Histogram
	for i := 1; i <= 1000; i++ {
		exact.Add(time.Duration(i) * time.Millisecond)
	}
	if got := exact.Percentile(99.9); got != 999*time.Millisecond {
		t.Errorf("p99.9 of 1..1000 ms = %v, want 999ms (rank exactly 999)", got)
	}
	if got := exact.Percentile(95); got != 950*time.Millisecond {
		t.Errorf("p95 of 1..1000 ms = %v, want 950ms", got)
	}
}

func TestAnalyseFixture(t *testing.T) {
	f, err := os.Open("testdata/access.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rep, err := Analyse(f, 3)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}

	if rep.Lines != 12 || rep.Parsed != 9 || rep.Malformed != 3 {
		t.Errorf("lines/parsed/malformed = %d/%d/%d, want 12/9/3", rep.Lines, rep.Parsed, rep.Malformed)
	}
	if want := []int{5, 10, 11}; !reflect.DeepEqual(rep.BadLines, want) {
		t.Errorf("BadLines = %v, want %v (line numbers count the blank line)", rep.BadLines, want)
	}
	if want := map[int]int{200: 5, 404: 1, 429: 1, 500: 1, 502: 1}; !reflect.DeepEqual(rep.Status, want) {
		t.Errorf("Status = %v, want %v", rep.Status, want)
	}
	if rep.Class(2) != 5 || rep.Class(4) != 2 || rep.Class(5) != 2 || rep.Class(3) != 0 {
		t.Errorf("classes 2xx/3xx/4xx/5xx = %d/%d/%d/%d, want 5/0/2/2", rep.Class(2), rep.Class(3), rep.Class(4), rep.Class(5))
	}
	if want := []PathCount{{"/stk/push", 4}, {"/balance", 3}, {"/health", 1}}; !reflect.DeepEqual(rep.TopPaths, want) {
		t.Errorf("TopPaths = %v, want %v", rep.TopPaths, want)
	}
	if rep.Bytes != 877 {
		t.Errorf("Bytes = %d, want 877", rep.Bytes)
	}
	if rep.P50 != 50*time.Millisecond || rep.P95 != 10*time.Second || rep.P99 != 10*time.Second {
		t.Errorf("p50/p95/p99 = %v/%v/%v, want 50ms/10s/10s", rep.P50, rep.P95, rep.P99)
	}
	first, last := time.Date(2026, 10, 24, 9, 0, 1, 0, eat), time.Date(2026, 10, 24, 9, 0, 11, 0, eat)
	if !rep.First.Equal(first) || !rep.Last.Equal(last) {
		t.Errorf("First/Last = %v/%v, want %v/%v", rep.First, rep.Last, first, last)
	}

	var out bytes.Buffer
	if err := rep.Write(&out); err != nil {
		t.Fatalf("Write: %v", err)
	}
	for _, want := range []string{"malformed", "5xx", "p95", "10s", "/stk/push"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("Write output has no %q:\n%s", want, out.String())
		}
	}
}

func TestAnalyseEdgeCases(t *testing.T) {
	good := `41.90.64.7 - - [24/Oct/2026:09:00:01 +0300] "GET /a HTTP/1.1" 200 1 "-" "ua" 0.001`

	rep, err := Analyse(strings.NewReader(""), 10)
	if err != nil || rep.Lines != 0 || rep.P95 != 0 || len(rep.TopPaths) != 0 {
		t.Errorf("empty input: %+v, %v", rep, err)
	}

	// No newline at the end, and Windows line endings.
	rep, _ = Analyse(strings.NewReader(good+"\r\n"+good), 10)
	if rep.Parsed != 2 || rep.Malformed != 0 {
		t.Errorf("CRLF and no final newline: parsed %d, malformed %d, want 2 and 0", rep.Parsed, rep.Malformed)
	}

	// A 1 MB line is skipped as malformed, and the next line still counts.
	huge := strings.Repeat("x", 1<<20)
	rep, err = Analyse(strings.NewReader(good+"\n"+huge+"\n"+good+"\n"), 10)
	if err != nil || rep.Parsed != 2 || rep.Malformed != 1 || !reflect.DeepEqual(rep.BadLines, []int{2}) {
		t.Errorf("a huge line: parsed %d, malformed %d, bad %v, err %v; want 2, 1, [2], nil", rep.Parsed, rep.Malformed, rep.BadLines, err)
	}

	// Only the first MaxBadLines line numbers are kept.
	rep, _ = Analyse(strings.NewReader(strings.Repeat("junk\n", 8)), 10)
	if rep.Malformed != 8 || len(rep.BadLines) != MaxBadLines {
		t.Errorf("8 junk lines: malformed %d, kept %d line numbers, want 8 and %d", rep.Malformed, len(rep.BadLines), MaxBadLines)
	}

	// A read error is returned, not hidden.
	_, err = Analyse(iotest.TimeoutReader(strings.NewReader(good+"\n"+good+"\n")), 10)
	if !errors.Is(err, iotest.ErrTimeout) {
		t.Errorf("a failing reader: error = %v, want iotest.ErrTimeout", err)
	}
}

func TestAnalyseStreams(t *testing.T) {
	// 100 000 lines generated on the fly: Analyse must read them as a stream.
	r := &repeatReader{line: []byte(`41.90.64.7 - - [24/Oct/2026:09:00:01 +0300] "GET /a HTTP/1.1" 200 1 "-" "ua" 0.007` + "\n"), left: 100_000}
	rep, err := Analyse(r, 1)
	if err != nil || rep.Parsed != 100_000 || rep.P95 != 7*time.Millisecond {
		t.Errorf("100 000 lines: parsed %d, p95 %v, err %v", rep.Parsed, rep.P95, err)
	}
}

type repeatReader struct {
	line []byte
	left int
	buf  []byte
}

func (r *repeatReader) Read(p []byte) (int, error) {
	if len(r.buf) == 0 {
		if r.left == 0 {
			return 0, io.EOF
		}
		r.buf = r.line
		r.left--
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}
