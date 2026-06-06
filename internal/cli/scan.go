package cli

import (
	"fmt"
	"os"

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

	client, ctx, cancel, err := connectClient(flags)
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect()

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

	if failOn != "never" && results.ShouldFail(failOn) {
		return fmt.Errorf("security findings above %s threshold", failOn)
	}
	return nil
}
