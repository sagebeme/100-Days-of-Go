// Package logstat reads web server access logs as a stream and reports status codes,
// the busiest paths and latency percentiles, in constant memory.
package logstat

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Entry is one parsed line of the log.
type Entry struct {
	IP        string
	User      string // "-" when not logged in
	Time      time.Time
	Method    string
	Path      string // the request target with any ?query removed
	Proto     string
	Status    int
	Bytes     int64
	Referer   string
	UserAgent string
	Latency   time.Duration // rounded to the millisecond
}

// TimeLayout is the timestamp format inside [ ].
const TimeLayout = "02/Jan/2006:15:04:05 -0700"

// ErrMalformed wraps every parse error.
var ErrMalformed = errors.New("malformed log line")

// ParseLine reads one line in nginx's "combined" format with the request time added
// at the end ($request_time, in seconds):
//
//	41.90.64.7 - - [24/Oct/2026:09:00:01 +0300] "POST /stk/push HTTP/1.1" 200 312 "-" "okhttp/4.12.0" 0.120
//
// The request must have a method, a target and a protocol. Status must be 100 to 599.
// Bytes may be "-", meaning 0. Latency must be 0 or more. Errors wrap ErrMalformed.
func ParseLine(line string) (Entry, error) {
	var e Entry
	bad := func(what string) (Entry, error) {
		return Entry{}, fmt.Errorf("%w: %s", ErrMalformed, what)
	}

	var ok bool
	rest := line
	if e.IP, rest, ok = strings.Cut(rest, " "); !ok || e.IP == "" {
		return bad("no IP")
	}
	if _, rest, ok = strings.Cut(rest, " "); !ok { // the ident field, always "-"
		return bad("no ident")
	}
	if e.User, rest, ok = strings.Cut(rest, " "); !ok {
		return bad("no user")
	}

	ts, rest, ok := between(rest, '[', ']')
	if !ok {
		return bad("no [time]")
	}
	t, err := time.Parse(TimeLayout, ts)
	if err != nil {
		return bad("bad time")
	}
	e.Time = t

	req, rest, ok := between(strings.TrimPrefix(rest, " "), '"', '"')
	if !ok {
		return bad("no quoted request")
	}
	parts := strings.Fields(req)
	if len(parts) != 3 {
		return bad("request isn't METHOD TARGET PROTO")
	}
	e.Method, e.Proto = parts[0], parts[2]
	e.Path, _, _ = strings.Cut(parts[1], "?")

	fields := strings.Fields(rest)
	if len(fields) < 2 {
		return bad("no status and bytes")
	}
	if e.Status, err = strconv.Atoi(fields[0]); err != nil || e.Status < 100 || e.Status > 599 {
		return bad("bad status")
	}
	if fields[1] != "-" {
		if e.Bytes, err = strconv.ParseInt(fields[1], 10, 64); err != nil || e.Bytes < 0 {
			return bad("bad bytes")
		}
	}

	// Skip past status and bytes, then read the two quoted fields and the latency.
	rest = strings.TrimLeft(rest, " ")
	rest = strings.TrimLeft(rest[len(fields[0]):], " ")
	rest = strings.TrimLeft(rest[len(fields[1]):], " ")
	if e.Referer, rest, ok = between(rest, '"', '"'); !ok {
		return bad("no quoted referer")
	}
	if e.UserAgent, rest, ok = between(strings.TrimPrefix(rest, " "), '"', '"'); !ok {
		return bad("no quoted user agent")
	}
	secs, err := strconv.ParseFloat(strings.TrimSpace(rest), 64)
	if err != nil || secs < 0 || math.IsInf(secs, 0) || math.IsNaN(secs) {
		return bad("bad request time")
	}
	e.Latency = time.Duration(math.Round(secs*1000)) * time.Millisecond
	return e, nil
}

// between reads s from an opening byte to the next closing byte. It returns what's
// inside and what's after the closing byte.
func between(s string, open, close byte) (inside, rest string, ok bool) {
	if len(s) == 0 || s[0] != open {
		return "", s, false
	}
	end := strings.IndexByte(s[1:], close)
	if end < 0 {
		return "", s, false
	}
	return s[1 : end+1], s[end+2:], true
}
