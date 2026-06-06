package scanner

import "fmt"

// fmtSprintf is a thin wrapper to keep `fmt` import in one place.
// This lets us consolidate formatting logic and avoid import cycles
// when stringer() is called from multiple rule files.
func fmtSprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
