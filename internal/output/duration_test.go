package output

import (
	"testing"
	"time"
)

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
