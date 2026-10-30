// Package logstat reads web server access logs as a stream and reports status codes,
// the busiest paths and latency percentiles, in constant memory.
package logstat

import (
	"errors"
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
	// Your code here
	return Entry{}, nil
}
