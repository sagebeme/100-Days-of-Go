// Package systemdesignbasics turns back-of-the-envelope capacity estimates for the Twende
// matatu booking service into functions you can test.
package systemdesignbasics

import (
	"errors"
	"time"
)

// SecondsPerDay is the number of seconds in a day, the divisor for every "per day" figure.
const SecondsPerDay = 24 * 60 * 60

// ErrNoSamples is returned by Percentile and Summarise when there are no samples.
var ErrNoSamples = errors.New("no latency samples")

// AverageRPS returns the average requests per second when dailyUsers each send
// requestsPerUser requests spread evenly over a day.
//
//	AverageRPS(864_000, 10) → 100
func AverageRPS(dailyUsers, requestsPerUser int) float64 {
	// Your code here
	return 0
}

// PeakRPS returns the busiest-moment requests per second: the average multiplied by
// peakFactor (how many times busier than average the rush hour is).
//
//	PeakRPS(864_000, 10, 5) → 500
func PeakRPS(dailyUsers, requestsPerUser int, peakFactor float64) float64 {
	// Your code here
	return 0
}

// StorageBytes returns how many bytes recordsPerDay records of bytesPerRecord each take
// after days days, stored on replicas copies.
//
//	StorageBytes(50_000, 500, 365, 3) → 27_375_000_000
func StorageBytes(recordsPerDay, bytesPerRecord, days, replicas int) int64 {
	// Your code here
	return 0
}

// BandwidthMbps returns the outgoing bandwidth in megabits per second when the service
// answers rps requests per second with responses of bytesPerResponse bytes.
// One megabit is 1 000 000 bits.
//
//	BandwidthMbps(500, 2_000) → 8
func BandwidthMbps(rps float64, bytesPerResponse int) float64 {
	// Your code here
	return 0
}

// Concurrency uses Little's law (L = λ × W) to return how many requests are in flight on
// average when rps requests arrive each second and each one takes latency to finish.
//
//	Concurrency(500, 200*time.Millisecond) → 100
func Concurrency(rps float64, latency time.Duration) float64 {
	// Your code here
	return 0
}

// ServersNeeded returns how many servers handle peakRPS when one server handles
// perServer requests per second and you keep headroom (0.3 means 30%) spare for spikes
// and for losing a server. It always returns at least 1, and 0 when perServer is not positive.
//
//	ServersNeeded(500, 200, 0.3) → 4   (500 / (200 × 0.7) = 3.57, rounded up)
func ServersNeeded(peakRPS, perServer, headroom float64) int {
	// Your code here
	return 0
}

// Percentile returns the p-th percentile (0 < p ≤ 100) of samples using the nearest-rank
// method: sort the samples and take the one at position ceil(p/100 × n), counting from 1.
// It doesn't change samples. It returns ErrNoSamples for an empty slice.
//
//	Percentile([]time.Duration{40ms, 10ms, 30ms, 20ms}, 50) → 20ms
func Percentile(samples []time.Duration, p float64) (time.Duration, error) {
	// Your code here
	return 0, nil
}

// LatencySummary holds the three percentiles people quote for a service.
type LatencySummary struct {
	P50, P95, P99 time.Duration
}

// Summarise returns the p50, p95 and p99 of samples, sorting a copy only once.
// It returns ErrNoSamples for an empty slice.
func Summarise(samples []time.Duration) (LatencySummary, error) {
	// Your code here
	return LatencySummary{}, nil
}
