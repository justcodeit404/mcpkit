package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// sharedFlags holds all the common connection flags.
type sharedFlags struct {
	ConfigPath      string
	Transport       string
	Command         string
	URL             string
	Headers         []string
	ProtocolVersion string
	Output          string
	NoColor         bool
	Verbose         bool
	Quiet           bool
	Timeout         string // stored as string from pflag
	LogFile         string
}

// bindSharedFlags extracts flag values into a sharedFlags struct.
func bindSharedFlags(cmd *cobra.Command) sharedFlags {
	return sharedFlags{
		ConfigPath:      getString(cmd.Flags(), "config"),
		Transport:       getString(cmd.Flags(), "transport"),
		Command:         getString(cmd.Flags(), "command"),
		URL:             getString(cmd.Flags(), "url"),
		Headers:         getStringSlice(cmd.Flags(), "header"),
		ProtocolVersion: getString(cmd.Flags(), "protocol-version"),
		Output:          getString(cmd.Flags(), "output"),
		NoColor:         getBool(cmd.Flags(), "no-color"),
		Verbose:         getBool(cmd.Flags(), "verbose"),
		Quiet:           getBool(cmd.Flags(), "quiet"),
		Timeout:         getString(cmd.Flags(), "timeout"),
		LogFile:         getString(cmd.Flags(), "log-file"),
	}
}

func getString(flags *pflag.FlagSet, name string) string {
	v, _ := flags.GetString(name)
	return v
}

func getBool(flags *pflag.FlagSet, name string) bool {
	v, _ := flags.GetBool(name)
	return v
}

func getStringSlice(flags *pflag.FlagSet, name string) []string {
	v, _ := flags.GetStringSlice(name)
	return v
}

func getInt(flags *pflag.FlagSet, name string) int {
	v, _ := flags.GetInt(name)
	return v
}

// connectClient builds an MCP client from shared flags, connects, and returns
// the client along with a context that auto-cancels on timeout.
func connectClient(flags sharedFlags) (*mcp.Client, context.Context, context.CancelFunc, error) {
	command, args, err := mcp.ParseCommand(flags.Command)
	if err != nil && flags.URL == "" {
		return nil, nil, nil, fmt.Errorf("--command or --url is required: %w", err)
	}

	cfg := mcp.Config{
		Transport:       flags.Transport,
		URL:             flags.URL,
		Command:         command,
		Args:            args,
		Headers:         parseHeaders(flags.Headers),
		ProtocolVersion: flags.ProtocolVersion,
	}

	timeout := 30 * time.Second
	if d, err := time.ParseDuration(flags.Timeout); err == nil {
		timeout = d
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	client := mcp.NewClient(cfg)
	if err := client.Connect(ctx); err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("connect: %w", err)
	}
	return client, ctx, cancel, nil
}

// parseHeaders converts a slice of "key:value" strings into a map.
func parseHeaders(raw []string) map[string]string {
	m := map[string]string{}
	for _, h := range raw {
		for i, c := range h {
			if c == ':' {
				m[h[:i]] = h[i+1:]
				break
			}
		}
	}
	return m
}

// parseJSONArgs parses a JSON string into a map. Returns nil if empty.
func parseJSONArgs(raw string) (map[string]any, error) {
	if raw == "" || raw == "{}" {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("invalid JSON args: %w", err)
	}
	return m, nil
}
