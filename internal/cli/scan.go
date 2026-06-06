package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/justcodeit404/mcpkit/internal/output"
	"github.com/justcodeit404/mcpkit/internal/scanner"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Security scan an MCP server for vulnerabilities",
	Long: `Scan runs security rules against an MCP server, checking tool definitions,
parameter schemas, descriptions, and naming for known attack vectors.

Examples:
  mcpkit scan --command "npx -y some-mcp-server"
  mcpkit scan --command "./server" --tier 1 --show-remediation
  mcpkit scan --command "./server" --fail-on critical`,
	RunE: runScan,
}

func init() {
	scanCmd.Flags().IntP("tier", "T", 2, "Minimum severity tier to report (1-5)")
	scanCmd.Flags().StringSliceP("include", "i", nil, "Specific rule IDs to run (default: all)")
	scanCmd.Flags().StringSliceP("exclude", "e", nil, "Rule IDs to skip")
	scanCmd.Flags().String("fail-on", "high", "Exit non-zero on: critical|high|medium|low|info|never")
	scanCmd.Flags().Bool("offline", false, "Scan only the static snapshot (skip live requests)")
	scanCmd.Flags().Bool("show-remediation", false, "Include remediation guidance in output")
}

func runScan(cmd *cobra.Command, _ []string) error {
	flags := bindSharedFlags(cmd)
	tier := getInt(cmd.Flags(), "tier")
	include := getStringSlice(cmd.Flags(), "include")
	exclude := getStringSlice(cmd.Flags(), "exclude")
	failOn := getString(cmd.Flags(), "fail-on")
	offline := getBool(cmd.Flags(), "offline")
	showRemediation := getBool(cmd.Flags(), "show-remediation")

	command, args, err := mcp.ParseCommand(flags.Command)
	if err != nil && flags.URL == "" {
		return fmt.Errorf("--command or --url is required: %w", err)
	}

	cfg := mcp.Config{
		Transport:       flags.Transport,
		URL:             flags.URL,
		Command:         command,
		Args:            args,
		Headers:         parseHeaders(flags.Headers),
		ProtocolVersion: flags.ProtocolVersion,
	}

	ctx, cancel := context.WithTimeout(context.Background(), parseDuration(flags.Timeout))
	defer cancel()

	client := mcp.NewClient(cfg)
	defer client.Disconnect()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	engine := scanner.New(scanner.Options{
		MinTier:         tier,
		Include:         include,
		Exclude:         exclude,
		Offline:         offline,
		ShowRemediation: showRemediation,
	})

	results, err := engine.Run(ctx, client)
	if err != nil {
		return err
	}

	if flags.Output == "json" {
		return output.NewJSONFormatter().Format(os.Stdout, results.Raw())
	}

	ren := output.NewTextFormatter()
	ren.NoColor = flags.NoColor
	if err := ren.Format(os.Stdout, output.ScanResultRenderable{
		Server:   serverName(client),
		Findings: results.Renderable(),
		Summary:  results.Summary(),
	}); err != nil {
		return err
	}

	// Determine exit code based on --fail-on.
	if failOn != "never" && results.ShouldFail(failOn) {
		client.Disconnect()
		os.Exit(1)
	}
	return nil
}
