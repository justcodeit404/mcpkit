package cli

import (
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

// transport flags helpers — commonly added to connection-oriented subcommands.
func addTransportFlags(flags *pflag.FlagSet) {
	flags.String("transport", "stdio", "Transport type: stdio|sse|streamable-http")
	flags.String("command", "", "Command to launch MCP server (for stdio transport)")
	flags.String("url", "", "Server endpoint URL (for HTTP transports)")
	flags.StringSliceP("header", "H", nil, "Custom HTTP headers in key:value format (repeatable)")
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

func getDuration(flags *pflag.FlagSet, name string) string {
	v, _ := flags.GetString(name)
	return v
}
