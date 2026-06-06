// Package benchmark measures MCP server performance.
package benchmark

import (
	"math"
	"sort"
	"time"
)

// Stats is a set of latency samples with computed statistics.
type Stats struct {
	Samples    []time.Duration
	Min        time.Duration
	Max        time.Duration
	Mean       time.Duration
	Median     time.Duration
	P75        time.Duration
	P90        time.Duration
	P95        time.Duration
	P99        time.Duration
	Stddev     time.Duration
	Throughput float64 // requests/second
	Total      time.Duration
	Errors     int
}

// Compute returns summary statistics for a set of latency samples.
func Compute(samples []time.Duration, totalDur time.Duration, errs int) Stats {
	if len(samples) == 0 {
		return Stats{Total: totalDur, Errors: errs}
	}
	sorted := make([]time.Duration, len(samples))
	copy(sorted, samples)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var sum time.Duration
	for _, d := range sorted {
		sum += d
	}
	mean := sum / time.Duration(len(sorted))

	var sqSum float64
	for _, d := range sorted {
		dd := float64(d - mean)
		sqSum += dd * dd
	}
	stddev := time.Duration(math.Sqrt(sqSum / float64(len(sorted))))

	tp := 0.0
	if totalDur > 0 {
		tp = float64(len(sorted)) / totalDur.Seconds()
	}

	return Stats{
		Samples:    sorted,
		Min:        sorted[0],
		Max:        sorted[len(sorted)-1],
		Mean:       mean,
		Median:     percentile(sorted, 50),
		P75:        percentile(sorted, 75),
		P90:        percentile(sorted, 90),
		P95:        percentile(sorted, 95),
		P99:        percentile(sorted, 99),
		Stddev:     stddev,
		Throughput: tp,
		Total:      totalDur,
		Errors:     errs,
	}
}

// percentile returns the p-th percentile from a sorted slice.
func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(p)/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
