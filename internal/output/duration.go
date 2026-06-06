package output

import (
	"fmt"
	"time"
)

// FormatDuration renders a duration in compact human-readable form.
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
