// Package mcp — transport factory.
package mcp

import (
	"fmt"
	"os/exec"
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

// BuildStdioCommand returns an *exec.Cmd for the given command and args.
func BuildStdioCommand(command string, args []string, env []string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	if len(env) > 0 {
		cmd.Env = append(cmd.Environ(), env...)
	}
	return cmd
}
