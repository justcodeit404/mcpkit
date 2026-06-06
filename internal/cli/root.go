package cli

import (
	"github.com/spf13/cobra"
)

// Version information set by goreleaser at build time.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// Execute runs the root command. Returns nil on success, or an error.
// Exit code handling belongs in main(), not here.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "mcpkit",
	Short: "The Swiss Army knife for MCP development",
	Long: `mcpkit is a comprehensive CLI toolkit for MCP (Model Context Protocol) developers.

Test, scan, bench, fuzz, and probe your MCP servers with a single, fast,
zero-dependency binary. Built for developers who want to ship reliable,
secure MCP servers.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to mcp.json config file")
	rootCmd.PersistentFlags().String("transport", "stdio", "Transport type: stdio|sse|streamable-http")
	rootCmd.PersistentFlags().String("command", "", "Command to launch MCP server (stdio transport)")
	rootCmd.PersistentFlags().String("url", "", "Server URL (HTTP transports)")
	rootCmd.PersistentFlags().StringSliceP("header", "H", nil, "Custom HTTP headers (key:value)")
	rootCmd.PersistentFlags().String("protocol-version", "2025-11-25", "MCP protocol version to negotiate")
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format: text|json")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-error output")
	rootCmd.PersistentFlags().Duration("timeout", 30_000_000_000, "Connection timeout") // 30s
	rootCmd.PersistentFlags().String("log-file", "", "Write debug log to file")

	rootCmd.AddCommand(probeCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(benchCmd)
}
