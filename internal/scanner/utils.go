package scanner

import "fmt"

// stringer renders a struct as a JSON-ish string for regex scanning.
func stringer(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%+v", v)
}

// truncate returns the first n bytes of s, adding "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
