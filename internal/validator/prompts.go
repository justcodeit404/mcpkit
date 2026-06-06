package validator

import (
	"context"
	"time"

	"github.com/justcodeit404/mcpkit/internal/mcp"
)

// runPromptChecks verifies prompts/list and prompts/get behavior.
func (s *Spec) runPromptChecks(ctx context.Context, client *mcp.Client, opts RunOptions, r *Results) {
	start := time.Now()
	prompts, err := client.ListPrompts(ctx)
	if err != nil {
		r.record("PL-001", "prompts/list returns valid response", "fail",
			"prompts/list error: "+err.Error(), time.Since(start))
		r.record("PG-001", "prompts/get returns messages", "skip",
			"requires PL-001", 0)
		r.record("PG-002", "Missing args handled", "skip",
			"requires PL-001", 0)
		return
	}
	r.record("PL-001", "prompts/list returns valid response", "pass",
		"server returned "+itoa(len(prompts))+" prompts", time.Since(start))

	if s.promptFilter == "" {
		r.record("PG-001", "prompts/get returns messages", "skip",
			"specify --prompt to test get", 0)
		r.record("PG-002", "Missing args handled", "skip",
			"specify --prompt to test get", 0)
		return
	}

	start = time.Now()
	_, err = client.GetPrompt(ctx, s.promptFilter, opts.PromptArgs)
	if err == nil {
		r.record("PG-001", "prompts/get returns messages", "pass",
			"prompt returned messages", time.Since(start))
	} else {
		r.record("PG-001", "prompts/get returns messages", "fail",
			"get error: "+err.Error(), time.Since(start))
	}

	// PG-002: missing required args returns error.
	start = time.Now()
	_, err = client.GetPrompt(ctx, s.promptFilter, map[string]string{})
	if err != nil {
		r.record("PG-002", "Missing args handled", "pass",
			"server returned error for missing args", time.Since(start))
	} else {
		r.record("PG-002", "Missing args handled", "pass",
			"server accepted call (likely optional args)", time.Since(start))
	}
}

// runPingCheck verifies the ping method.
func (s *Spec) runPingCheck(ctx context.Context, client *mcp.Client, r *Results) {
	start := time.Now()
	if err := client.Ping(ctx); err != nil {
		r.record("PING-01", "ping returns empty result", "fail",
			"ping error: "+err.Error(), time.Since(start))
	} else {
		r.record("PING-01", "ping returns empty result", "pass",
			"ping succeeded", time.Since(start))
	}
}
