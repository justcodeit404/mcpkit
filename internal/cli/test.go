package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/justcodeit404/mcpkit/internal/output"
	"github.com/justcodeit404/mcpkit/internal/validator"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run protocol compliance tests against an MCP server",
	Long: `Test runs the full MCP specification compliance suite against a server,
including handshake, tools, resources, and prompts checks.

Examples:
  mcpkit test --command "npx -y @modelcontextprotocol/server-filesystem /tmp"
  mcpkit test --command "go run ./cmd/myserver" --method tools/list
  mcpkit test --command "./server" --output json`,
	RunE: runTest,
}

func init() {
	testCmd.Flags().StringP("method", "m", "", "Test specific method only: initialize|tools/list|tools/call|resources/list|resources/read|prompts/list|prompts/get|ping")
	testCmd.Flags().StringP("tool", "t", "", "Tool to test (with --method=tools/call)")
	testCmd.Flags().String("resource", "", "Resource URI to test (with --method=resources/read)")
	testCmd.Flags().String("prompt", "", "Prompt name to test (with --method=prompts/get)")
	testCmd.Flags().String("tool-args", "{}", "JSON arguments for tool call")
	testCmd.Flags().String("prompt-args", "{}", "JSON arguments for prompt get")
	testCmd.Flags().Bool("fail-fast", false, "Stop on first failure")
	testCmd.Flags().IntP("retry", "r", 0, "Retry failed checks N times")
}

func runTest(cmd *cobra.Command, _ []string) error {
	flags := bindSharedFlags(cmd)
	method := getString(cmd.Flags(), "method")
	toolName := getString(cmd.Flags(), "tool")
	resourceURI := getString(cmd.Flags(), "resource")
	promptName := getString(cmd.Flags(), "prompt")
	toolArgs := getString(cmd.Flags(), "tool-args")
	promptArgs := getString(cmd.Flags(), "prompt-args")
	failFast := getBool(cmd.Flags(), "fail-fast")

	// Build client config from flags.
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

	spec := validator.NewSpec()
	if method != "" {
		spec.SetFilter(method)
	}
	if toolName != "" {
		spec.SetToolFilter(toolName)
	}
	if resourceURI != "" {
		spec.SetResourceFilter(resourceURI)
	}
	if promptName != "" {
		spec.SetPromptFilter(promptName)
	}

	var toolArgsMap map[string]any
	if err := json.Unmarshal([]byte(toolArgs), &toolArgsMap); err != nil {
		return fmt.Errorf("invalid --tool-args JSON: %w", err)
	}
	var promptArgsMap map[string]string
	if err := json.Unmarshal([]byte(promptArgs), &promptArgsMap); err != nil {
		return fmt.Errorf("invalid --prompt-args JSON: %w", err)
	}

	results, durMS, err := spec.Run(ctx, client, validator.RunOptions{
		ToolArgs:   toolArgsMap,
		PromptArgs: promptArgsMap,
		FailFast:   failFast,
	})
	if err != nil {
		return err
	}

	// Render.
	formatter := pickFormatter(flags.Output, flags.NoColor)
	if flags.Output == "json" {
		return formatter.Format(os.Stdout, results.Raw())
	}

	ren := output.NewTextFormatter()
	return ren.Format(os.Stdout, output.TestResultRenderable{
		Server:  serverName(client),
		Checks:  results.Renderable(),
		Summary: results.Summary(),
		DurMS:   durMS.Milliseconds(),
	})
}

func serverName(c *mcp.Client) string {
	if r := c.InitializeResult(); r != nil {
		return fmt.Sprintf("%s@%s", r.ServerInfo.Name, r.ServerInfo.Version)
	}
	return "unknown"
}

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

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func pickFormatter(name string, noColor bool) output.Formatter {
	switch name {
	case "json":
		return output.NewJSONFormatter()
	default:
		f := output.NewTextFormatter()
		f.NoColor = noColor
		return f
	}
}
