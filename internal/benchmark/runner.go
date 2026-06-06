package benchmark

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/justcodeit404/mcpkit/internal/output"
)

// Options configures a benchmark run.
type Options struct {
	Iterations  int
	Concurrency int
	Warmup      int
	Method      string
	ToolName    string
	ToolArgs    map[string]any
	ResourceURI string
	PromptName  string
	Buckets     []string
}

// Results holds the full benchmark output.
type Results struct {
	Method    string
	Stats     Stats
	Histogram []Bucket
	Edges     []time.Duration
}

// Runner executes a benchmark against an MCP server.
type Runner struct {
	opts Options
}

// New creates a Runner.
func New(opts Options) *Runner {
	if opts.Iterations <= 0 {
		opts.Iterations = 100
	}
	if opts.Warmup < 0 {
		opts.Warmup = 0
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}
	return &Runner{opts: opts}
}

// Run executes the benchmark and returns the results.
func (r *Runner) Run(ctx context.Context, cfg mcp.Config) (*Results, error) {
	// Parse buckets if provided.
	edges := DefaultBuckets
	if len(r.opts.Buckets) > 0 {
		edges = edges[:0]
		for _, b := range r.opts.Buckets {
			d, err := time.ParseDuration(b)
			if err != nil {
				return nil, fmt.Errorf("invalid bucket %q: %w", b, err)
			}
			edges = append(edges, d)
		}
		sort.Slice(edges, func(i, j int) bool { return edges[i] < edges[j] })
	}

	// Connect.
	client := mcp.NewClient(cfg)
	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Warmup.
	for i := 0; i < r.opts.Warmup; i++ {
		_ = r.invokeOnce(ctx, client)
	}

	// Measured run.
	concurrency := r.opts.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}

	var (
		mu      sync.Mutex
		samples []time.Duration
		errs    int64
		wg      sync.WaitGroup
		sem     = make(chan struct{}, concurrency)
	)

	start := time.Now()
	for i := 0; i < r.opts.Iterations; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			// Each goroutine creates its own client for true concurrency.
			c := mcp.NewClient(cfg)
			if err := c.Connect(ctx); err != nil {
				atomic.AddInt64(&errs, 1)
				return
			}
			defer c.Disconnect()

			d, err := r.invokeOnceTimed(ctx, c)
			if err != nil {
				atomic.AddInt64(&errs, 1)
				return
			}
			mu.Lock()
			samples = append(samples, d)
			mu.Unlock()
		}()
	}
	wg.Wait()
	total := time.Since(start)

	stats := Compute(samples, total, int(atomic.LoadInt64(&errs)))
	hist := Histogram(samples, edges)

	return &Results{
		Method:    r.opts.Method,
		Stats:     stats,
		Histogram: hist,
		Edges:     edges,
	}, nil
}

// invokeOnce runs a single iteration (no timing).
func (r *Runner) invokeOnce(ctx context.Context, c *mcp.Client) error {
	_, err := r.invokeOnceTimed(ctx, c)
	return err
}

// invokeOnceTimed runs a single method call and records its duration.
func (r *Runner) invokeOnceTimed(ctx context.Context, c *mcp.Client) (time.Duration, error) {
	start := time.Now()
	var err error
	switch r.opts.Method {
	case "ping":
		err = c.Ping(ctx)
	case "tools/list":
		_, err = c.ListTools(ctx)
	case "tools/call":
		if r.opts.ToolName == "" {
			return 0, fmt.Errorf("--tool required for tools/call")
		}
		_, err = c.CallTool(ctx, r.opts.ToolName, r.opts.ToolArgs)
	case "resources/list":
		_, err = c.ListResources(ctx)
	case "resources/read":
		if r.opts.ResourceURI == "" {
			return 0, fmt.Errorf("--resource required for resources/read")
		}
		_, err = c.ReadResource(ctx, r.opts.ResourceURI)
	case "prompts/list":
		_, err = c.ListPrompts(ctx)
	case "prompts/get":
		if r.opts.PromptName == "" {
			return 0, fmt.Errorf("--prompt required for prompts/get")
		}
		_, err = c.GetPrompt(ctx, r.opts.PromptName, nil)
	default:
		return 0, fmt.Errorf("unsupported method: %s", r.opts.Method)
	}
	return time.Since(start), err
}

// Metrics returns renderable metrics.
func (r *Results) Metrics() output.BenchMetrics {
	return output.BenchMetrics{
		Iterations: r.Stats.iterations(),
		Errors:     r.Stats.Errors,
		Min:        output.FormatDuration(r.Stats.Min),
		Max:        output.FormatDuration(r.Stats.Max),
		Mean:       output.FormatDuration(r.Stats.Mean),
		Median:     output.FormatDuration(r.Stats.Median),
		P75:        output.FormatDuration(r.Stats.P75),
		P90:        output.FormatDuration(r.Stats.P90),
		P95:        output.FormatDuration(r.Stats.P95),
		P99:        output.FormatDuration(r.Stats.P99),
		Stddev:     output.FormatDuration(r.Stats.Stddev),
		Throughput: fmt.Sprintf("%.1f req/s", r.Stats.Throughput),
		TotalDur:   output.FormatDuration(r.Stats.Total),
	}
}

// HistogramBuckets returns renderable histogram buckets.
func (r *Results) HistogramBuckets() []output.HistogramBucket {
	out := make([]output.HistogramBucket, len(r.Histogram))
	for i, b := range r.Histogram {
		out[i] = output.HistogramBucket{
			From:  output.FormatDuration(b.From),
			To:    output.FormatDuration(b.To),
			Count: b.Count,
		}
	}
	return out
}

// Raw returns a JSON-serializable representation.
func (r *Results) Raw() any {
	return map[string]any{
		"method":    r.Method,
		"stats":     r.Stats,
		"histogram": r.Histogram,
	}
}

func (s Stats) iterations() int {
	return len(s.Samples)
}
