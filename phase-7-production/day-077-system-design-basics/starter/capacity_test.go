package systemdesignbasics

import (
	"errors"
	"math"
	"slices"
	"testing"
	"time"
)

// near reports whether two floats agree to within a millionth, since float sums
// rarely land exactly.
func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestAverageAndPeakRPS(t *testing.T) {
	tests := []struct {
		users, perUser int
		factor         float64
		avg, peak      float64
	}{
		{864_000, 10, 5, 100, 500},
		{86_400, 1, 1, 1, 1},
		{200_000, 6, 4, 13.888889, 55.555556},
		{0, 10, 5, 0, 0},
	}
	for _, tt := range tests {
		if got := AverageRPS(tt.users, tt.perUser); math.Abs(got-tt.avg) > 1e-5 {
			t.Errorf("AverageRPS(%d, %d) = %v, want %v", tt.users, tt.perUser, got, tt.avg)
		}
		if got := PeakRPS(tt.users, tt.perUser, tt.factor); math.Abs(got-tt.peak) > 1e-5 {
			t.Errorf("PeakRPS(%d, %d, %v) = %v, want %v", tt.users, tt.perUser, tt.factor, got, tt.peak)
		}
	}
}

func TestStorageBytes(t *testing.T) {
	tests := []struct {
		records, size, days, replicas int
		want                          int64
	}{
		{50_000, 500, 365, 3, 27_375_000_000},
		{1, 1, 1, 1, 1},
		{0, 500, 365, 3, 0},
		{10_000_000, 1_000, 3_650, 3, 109_500_000_000_000}, // too big for 32 bits
	}
	for _, tt := range tests {
		if got := StorageBytes(tt.records, tt.size, tt.days, tt.replicas); got != tt.want {
			t.Errorf("StorageBytes(%d, %d, %d, %d) = %d, want %d", tt.records, tt.size, tt.days, tt.replicas, got, tt.want)
		}
	}
}

func TestBandwidthMbps(t *testing.T) {
	tests := []struct {
		rps  float64
		size int
		want float64
	}{
		{500, 2_000, 8},
		{1, 125_000, 1},
		{0, 2_000, 0},
		{55.5, 4_096, 1.818624},
	}
	for _, tt := range tests {
		if got := BandwidthMbps(tt.rps, tt.size); !near(got, tt.want) {
			t.Errorf("BandwidthMbps(%v, %d) = %v, want %v", tt.rps, tt.size, got, tt.want)
		}
	}
}

func TestConcurrency(t *testing.T) {
	tests := []struct {
		rps     float64
		latency time.Duration
		want    float64
	}{
		{500, 200 * time.Millisecond, 100},
		{100, time.Second, 100},
		{1_000, 50 * time.Millisecond, 50},
		{10, 0, 0},
	}
	for _, tt := range tests {
		if got := Concurrency(tt.rps, tt.latency); !near(got, tt.want) {
			t.Errorf("Concurrency(%v, %v) = %v, want %v", tt.rps, tt.latency, got, tt.want)
		}
	}
}

func TestServersNeeded(t *testing.T) {
	tests := []struct {
		peak, perServer, headroom float64
		want                      int
	}{
		{500, 200, 0.3, 4},
		{500, 250, 0, 2},
		{501, 250, 0, 3},
		{1, 1_000, 0.5, 1},
		{0, 1_000, 0.3, 1},
		{500, 0, 0.3, 0},
		{500, 200, 1, 0},
	}
	for _, tt := range tests {
		if got := ServersNeeded(tt.peak, tt.perServer, tt.headroom); got != tt.want {
			t.Errorf("ServersNeeded(%v, %v, %v) = %d, want %d", tt.peak, tt.perServer, tt.headroom, got, tt.want)
		}
	}
}

func ms(ns ...int) []time.Duration {
	out := make([]time.Duration, len(ns))
	for i, n := range ns {
		out[i] = time.Duration(n) * time.Millisecond
	}
	return out
}

func TestPercentile(t *testing.T) {
	tests := []struct {
		samples []time.Duration
		p       float64
		want    time.Duration
	}{
		{ms(40, 10, 30, 20), 50, 20 * time.Millisecond},
		{ms(40, 10, 30, 20), 75, 30 * time.Millisecond},
		{ms(40, 10, 30, 20), 76, 40 * time.Millisecond},
		{ms(40, 10, 30, 20), 100, 40 * time.Millisecond},
		{ms(40, 10, 30, 20), 1, 10 * time.Millisecond},
		{ms(7), 99, 7 * time.Millisecond},
		{ms(5, 5, 5, 900), 95, 900 * time.Millisecond},
	}
	for _, tt := range tests {
		got, err := Percentile(tt.samples, tt.p)
		if err != nil || got != tt.want {
			t.Errorf("Percentile(%v, %v) = %v, %v, want %v, nil", tt.samples, tt.p, got, err, tt.want)
		}
	}
	if _, err := Percentile(nil, 50); !errors.Is(err, ErrNoSamples) {
		t.Errorf("Percentile(nil, 50) error = %v, want ErrNoSamples", err)
	}
}

func TestPercentileLeavesInputAlone(t *testing.T) {
	samples := ms(40, 10, 30, 20)
	before := slices.Clone(samples)
	Percentile(samples, 50)
	Summarise(samples)
	if !slices.Equal(samples, before) {
		t.Errorf("samples became %v, want %v: sort a copy, not the caller's slice", samples, before)
	}
}

func TestSummarise(t *testing.T) {
	// 100 samples: 1ms, 2ms ... 100ms, shuffled by reversing.
	var samples []time.Duration
	for i := 100; i >= 1; i-- {
		samples = append(samples, time.Duration(i)*time.Millisecond)
	}
	got, err := Summarise(samples)
	want := LatencySummary{P50: 50 * time.Millisecond, P95: 95 * time.Millisecond, P99: 99 * time.Millisecond}
	if err != nil || got != want {
		t.Errorf("Summarise(1..100ms) = %+v, %v, want %+v, nil", got, err, want)
	}

	// A slow tail: the average hides it, p99 shows it.
	tail := append(ms(20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20), 2000*time.Millisecond)
	got, err = Summarise(tail)
	want = LatencySummary{P50: 20 * time.Millisecond, P95: 20 * time.Millisecond, P99: 2000 * time.Millisecond}
	if err != nil || got != want {
		t.Errorf("Summarise(19×20ms + 2s) = %+v, %v, want %+v, nil", got, err, want)
	}

	if _, err := Summarise(nil); !errors.Is(err, ErrNoSamples) {
		t.Errorf("Summarise(nil) error = %v, want ErrNoSamples", err)
	}
}
