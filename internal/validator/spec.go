// Package validator runs protocol compliance checks against an MCP server.
package validator

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/justcodeit404/mcpkit/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RunOptions configures a validator run.
type RunOptions struct {
	ToolArgs   map[string]any
	PromptArgs map[string]string
	FailFast   bool
}

// Results aggregates all check results from a run.
type Results struct {
	checks  []output.CheckResult
	summary output.TestSummary
}

// Spec orchestrates spec compliance testing.
type Spec struct {
	methodFilter    string
	toolFilter      string
	resourceFilter  string
	promptFilter    string
}

// NewSpec returns a fresh Spec.
func NewSpec() *Spec {
	return &Spec{}
}

// SetFilter restricts checks to a single method.
func (s *Spec) SetFilter(method string) {
	s.methodFilter = method
}

// SetToolFilter restricts tool-call checks to a specific tool name.
func (s *Spec) SetToolFilter(name string) { s.toolFilter = name }

// SetResourceFilter restricts resource-read checks to a specific URI.
func (s *Spec) SetResourceFilter(uri string) { s.resourceFilter = uri }

// SetPromptFilter restricts prompt-get checks to a specific name.
func (s *Spec) SetPromptFilter(name string) { s.promptFilter = name }

// Run executes the full compliance suite.
func (s *Spec) Run(ctx context.Context, client *mcp.Client, opts RunOptions) (*Results, time.Duration, error) {
	start := time.Now()
	results := &Results{}

	// Run handshake checks.
	if s.shouldRun("initialize") {
		s.runHandshake(ctx, client, results)
	}

	// Tool checks.
	if s.shouldRun("tools/list") || s.shouldRun("tools/call") {
		s.runToolChecks(ctx, client, opts, results)
	}

	// Resource checks.
	if s.shouldRun("resources/list") || s.shouldRun("resources/read") {
		s.runResourceChecks(ctx, client, results)
	}

	// Prompt checks.
	if s.shouldRun("prompts/list") || s.shouldRun("prompts/get") {
		s.runPromptChecks(ctx, client, opts, results)
	}

	// Ping.
	if s.shouldRun("ping") {
		s.runPingCheck(ctx, client, results)
	}

	// Compute summary.
	for _, c := range results.checks {
		results.summary.Total++
		switch c.Status {
		case "pass":
			results.summary.Passed++
		case "fail":
			results.summary.Failed++
		case "skip":
			results.summary.Skipped++
		}
	}
	return results, time.Since(start), nil
}

func (s *Spec) shouldRun(method string) bool {
	return s.methodFilter == "" || s.methodFilter == method
}

func (r *Results) Checks() []output.CheckResult            { return r.checks }
func (r *Results) Summary() output.TestSummary             { return r.summary }
func (r *Results) Renderable() []output.CheckResult        { return r.checks }

// Raw returns a JSON-serializable representation.
func (r *Results) Raw() any {
	return map[string]any{
		"checks":  r.checks,
		"summary": r.summary,
	}
}

// record is a helper to add a check result.
func (r *Results) record(id, name, status, msg string, dur time.Duration) {
	r.checks = append(r.checks, output.CheckResult{
		ID:       id,
		Name:     name,
		Status:   status,
		Message:  msg,
		Duration: formatDuration(dur),
	})
}

// formatDuration renders a duration as a compact human-readable string.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000.0)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// toolNamePattern is the MCP spec rule for tool names:
// 1-128 chars, [A-Za-z0-9_.-] only.
var toolNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,128}$`)

// initOk checks the initialize response from a connected client.
func initOk(c *mcp.Client) (*mcpsdk.InitializeResult, bool) {
	r := c.InitializeResult()
	if r == nil {
		return nil, false
	}
	return r, true
}
