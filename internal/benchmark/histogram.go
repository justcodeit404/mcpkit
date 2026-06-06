package benchmark

import (
	"fmt"
	"time"
)

// Bucket is a single histogram bucket.
type Bucket struct {
	From  time.Duration
	To    time.Duration
	Count int
}

// DefaultBuckets is a sensible default histogram for sub-second latencies.
var DefaultBuckets = []time.Duration{
	1 * time.Millisecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	500 * time.Millisecond,
	1 * time.Second,
	5 * time.Second,
	10 * time.Second,
}

// Histogram computes bucket counts from a slice of samples.
func Histogram(samples []time.Duration, edges []time.Duration) []Bucket {
	if len(edges) == 0 {
		edges = DefaultBuckets
	}
	sortEdges := make([]time.Duration, len(edges))
	copy(sortEdges, edges)
	// edges already monotonic ascending
	buckets := make([]Bucket, len(sortEdges)+1)
	for i, e := range sortEdges {
		if i == 0 {
			buckets[i] = Bucket{From: 0, To: e}
		} else {
			buckets[i] = Bucket{From: sortEdges[i-1], To: e}
		}
	}
	buckets[len(sortEdges)] = Bucket{From: sortEdges[len(sortEdges)-1], To: time.Duration(1 << 62)}

	for _, s := range samples {
		for i := range buckets {
			if s >= buckets[i].From && s < buckets[i].To {
				buckets[i].Count++
				break
			}
		}
	}
	return buckets
}

// FormatDuration renders a duration in compact form.
func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.1fµs", float64(d.Nanoseconds())/1000.0)
	case d < time.Second:
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1_000_000.0)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
