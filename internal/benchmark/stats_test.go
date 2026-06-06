package benchmark

import (
	"testing"
	"time"
)

func TestCompute(t *testing.T) {
	samples := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		20 * time.Millisecond,
	}
	stats := Compute(samples, 100*time.Millisecond, 0)
	if stats.Min != 1*time.Millisecond {
		t.Errorf("Min: got %v, want 1ms", stats.Min)
	}
	if stats.Max != 20*time.Millisecond {
		t.Errorf("Max: got %v, want 20ms", stats.Max)
	}
	if stats.Median != 4*time.Millisecond {
		t.Errorf("Median: got %v, want 4ms", stats.Median)
	}
	if stats.P95 != 20*time.Millisecond {
		t.Errorf("P95: got %v, want 20ms", stats.P95)
	}
}

func TestComputeEmpty(t *testing.T) {
	stats := Compute(nil, 0, 5)
	if stats.Errors != 5 {
		t.Errorf("Errors: got %d, want 5", stats.Errors)
	}
}

func TestHistogram(t *testing.T) {
	samples := []time.Duration{
		500 * time.Microsecond,
		2 * time.Millisecond,
		15 * time.Millisecond,
		300 * time.Millisecond,
	}
	buckets := Histogram(samples, DefaultBuckets)
	total := 0
	for _, b := range buckets {
		total += b.Count
	}
	if total != len(samples) {
		t.Errorf("bucket counts sum: got %d, want %d", total, len(samples))
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{500 * time.Nanosecond, "500ns"},
		{500 * time.Microsecond, "500.0µs"},
		{500 * time.Millisecond, "500.0ms"},
		{2500 * time.Millisecond, "2.50s"},
	}
	for _, c := range cases {
		got := FormatDuration(c.in)
		if got != c.want {
			t.Errorf("FormatDuration(%v): got %q, want %q", c.in, got, c.want)
		}
	}
}
