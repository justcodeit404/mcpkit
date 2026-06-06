package cli

import (
	"fmt"
	"os"

	"github.com/justcodeit404/mcpkit/internal/benchmark"
	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/justcodeit404/mcpkit/internal/output"
	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Benchmark MCP server performance",
	Long: `Bench measures MCP server latency and throughput.

Examples:
  mcpkit bench --command "./server" --method ping -n 100
  mcpkit bench --command "./server" --method tools/call --tool greet -n 1000`,
	RunE: runBench,
}

func init() {
	benchCmd.Flags().IntP("iterations", "n", 100, "Number of iterations")
	benchCmd.Flags().IntP("concurrency", "C", 1, "Concurrent workers")
	benchCmd.Flags().IntP("warmup", "w", 10, "Warmup iterations (discarded)")
	benchCmd.Flags().String("method", "ping", "Method: ping|tools/list|tools/call|resources/list|resources/read|prompts/list|prompts/get")
	benchCmd.Flags().StringP("tool", "t", "", "Tool name (for --method=tools/call)")
	benchCmd.Flags().String("tool-args", "{}", "JSON args for tool call")
	benchCmd.Flags().String("resource", "", "Resource URI (for --method=resources/read)")
	benchCmd.Flags().String("prompt", "", "Prompt name (for --method=prompts/get)")
	benchCmd.Flags().StringSlice("histogram-buckets", nil, "Custom latency buckets (e.g. \"1ms,5ms,10ms,50ms,100ms\")")
}

func runBench(cmd *cobra.Command, _ []string) error {
	flags := bindSharedFlags(cmd)
	method := getString(cmd.Flags(), "method")
	iterations := getInt(cmd.Flags(), "iterations")
	concurrency := getInt(cmd.Flags(), "concurrency")
	warmup := getInt(cmd.Flags(), "warmup")
	toolName := getString(cmd.Flags(), "tool")
	toolArgs := getString(cmd.Flags(), "tool-args")
	resourceURI := getString(cmd.Flags(), "resource")
	promptName := getString(cmd.Flags(), "prompt")
	buckets := getStringSlice(cmd.Flags(), "histogram-buckets")

	toolArgsMap, err := parseJSONArgs(toolArgs)
	if err != nil {
		return err
	}

	client, ctx, cancel, err := connectClient(flags)
	if err != nil {
		return err
	}
	cancel()
	client.Disconnect() // bench creates its own connections per worker

	runner := benchmark.New(benchmark.Options{
		Iterations:  iterations,
		Concurrency: concurrency,
		Warmup:      warmup,
		Method:      method,
		ToolName:    toolName,
		ToolArgs:    toolArgsMap,
		ResourceURI: resourceURI,
		PromptName:  promptName,
		Buckets:     buckets,
	})

	results, err := runner.Run(ctx, mcp.Config{
		Transport:       flags.Transport,
		URL:             flags.URL,
		Command:         flags.Command,
		Headers:         parseHeaders(flags.Headers),
		ProtocolVersion: flags.ProtocolVersion,
	})
	if err != nil {
		return err
	}

	if flags.Output == "json" {
		return output.NewJSONFormatter().Format(os.Stdout, results.Raw())
	}

	ren := output.NewTextFormatter()
	ren.NoColor = flags.NoColor
	return ren.Format(os.Stdout, output.BenchResultRenderable{
		Server:    fmt.Sprintf("%s (method=%s)", flags.Command, method),
		Method:    method,
		Metrics:   results.Metrics(),
		Histogram: results.HistogramBuckets(),
	})
}
