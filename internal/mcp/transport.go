// Package mcp — transport utilities.
package mcp

import (
	"fmt"
	"strings"
)

// ParseCommand splits a shell-style command string into Command and Args.
// E.g. "npx -y some-server /tmp" → ("npx", ["-y", "some-server", "/tmp"]).
func ParseCommand(cmdline string) (string, []string, error) {
	parts := strings.Fields(cmdline)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}
	return parts[0], parts[1:], nil
}
