// Package systemdesignbasics turns back-of-the-envelope capacity estimates for the Twende
// matatu booking service into functions you can test.
package systemdesignbasics

import (
	"errors"
	"math"
	"slices"
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
	return float64(dailyUsers) * float64(requestsPerUser) / SecondsPerDay
}

// PeakRPS returns the busiest-moment requests per second: the average multiplied by
// peakFactor (how many times busier than average the rush hour is).
//
//	PeakRPS(864_000, 10, 5) → 500
func PeakRPS(dailyUsers, requestsPerUser int, peakFactor float64) float64 {
	return AverageRPS(dailyUsers, requestsPerUser) * peakFactor
}

// StorageBytes returns how many bytes recordsPerDay records of bytesPerRecord each take
// after days days, stored on replicas copies.
//
//	StorageBytes(50_000, 500, 365, 3) → 27_375_000_000
func StorageBytes(recordsPerDay, bytesPerRecord, days, replicas int) int64 {
	return int64(recordsPerDay) * int64(bytesPerRecord) * int64(days) * int64(replicas)
}

// BandwidthMbps returns the outgoing bandwidth in megabits per second when the service
// answers rps requests per second with responses of bytesPerResponse bytes.
// One megabit is 1 000 000 bits.
//
//	BandwidthMbps(500, 2_000) → 8
func BandwidthMbps(rps float64, bytesPerResponse int) float64 {
	return rps * float64(bytesPerResponse) * 8 / 1_000_000
}

// Concurrency uses Little's law (L = λ × W) to return how many requests are in flight on
// average when rps requests arrive each second and each one takes latency to finish.
//
//	Concurrency(500, 200*time.Millisecond) → 100
func Concurrency(rps float64, latency time.Duration) float64 {
	return rps * latency.Seconds()
}

// ServersNeeded returns how many servers handle peakRPS when one server handles
// perServer requests per second and you keep headroom (0.3 means 30%) spare for spikes
// and for losing a server. It always returns at least 1, and 0 when perServer is not positive.
//
//	ServersNeeded(500, 200, 0.3) → 4   (500 / (200 × 0.7) = 3.57, rounded up)
func ServersNeeded(peakRPS, perServer, headroom float64) int {
	if perServer <= 0 {
		return 0
	}
	usable := perServer * (1 - headroom)
	if usable <= 0 {
		return 0
	}
	n := int(math.Ceil(peakRPS / usable))
	return max(n, 1)
}

// Percentile returns the p-th percentile (0 < p ≤ 100) of samples using the nearest-rank
// method: sort the samples and take the one at position ceil(p/100 × n), counting from 1.
// It doesn't change samples. It returns ErrNoSamples for an empty slice.
//
//	Percentile([]time.Duration{40ms, 10ms, 30ms, 20ms}, 50) → 20ms
func Percentile(samples []time.Duration, p float64) (time.Duration, error) {
	if len(samples) == 0 {
		return 0, ErrNoSamples
	}
	sorted := slices.Clone(samples)
	slices.Sort(sorted)
	return nearestRank(sorted, p), nil
}

// LatencySummary holds the three percentiles people quote for a service.
type LatencySummary struct {
	P50, P95, P99 time.Duration
}

// Summarise returns the p50, p95 and p99 of samples, sorting a copy only once.
// It returns ErrNoSamples for an empty slice.
func Summarise(samples []time.Duration) (LatencySummary, error) {
	if len(samples) == 0 {
		return LatencySummary{}, ErrNoSamples
	}
	sorted := slices.Clone(samples)
	slices.Sort(sorted)
	return LatencySummary{
		P50: nearestRank(sorted, 50),
		P95: nearestRank(sorted, 95),
		P99: nearestRank(sorted, 99),
	}, nil
}

// nearestRank reads the p-th percentile from an already sorted, non-empty slice.
func nearestRank(sorted []time.Duration, p float64) time.Duration {
	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	rank = min(max(rank, 1), len(sorted))
	return sorted[rank-1]
}
