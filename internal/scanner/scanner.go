package scanner

import (
	"context"
	"sort"

	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/justcodeit404/mcpkit/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Snapshot is a server snapshot captured for static analysis.
type Snapshot struct {
	ServerInfo *mcpsdk.Implementation
	Tools      []*mcpsdk.Tool
	Resources  []*mcpsdk.Resource
	Prompts    []*mcpsdk.Prompt
}

// Options configures a scanner run.
type Options struct {
	MinTier         int
	Include         []string
	Exclude         []string
	Offline         bool
	ShowRemediation bool
}

// Results is the aggregate of all findings.
type Results struct {
	findings []Finding
	summary  output.ScanSummary
}

// New constructs the scanner engine with all default rules.
func New(opts Options) *Engine {
	return &Engine{
		opts: opts,
		rules: []Rule{
			// Tier 1 — CRITICAL
			&CommandInjectionRule{},
			&SystemPromptOverrideRule{},
			&CredentialExfilRule{},
			&ShellMetacharDefaultsRule{},
			&UnsanitizedExecRule{},
			// Tier 2 — HIGH
			&ImperativeLanguageRule{},
			&ToolNameShadowingRule{},
			&Base64PayloadRule{},
			&MissingInputValidationRule{},
			&BroadFileSystemAccessRule{},
		},
	}
}

// Engine orchestrates rule execution.
type Engine struct {
	opts  Options
	rules []Rule
}

// Run connects to the server (unless offline) and runs all registered rules.
func (e *Engine) Run(ctx context.Context, client *mcp.Client) (*Results, error) {
	var snap *Snapshot
	if !e.opts.Offline {
		s, err := client.Snapshot(ctx)
		if err != nil {
			return nil, err
		}
		snap = snapshotFromMCP(s)
	} else {
		// Offline mode — no tools/resources/prompts available.
		snap = &Snapshot{}
	}

	res := &Results{}
	for _, rule := range e.rules {
		if !e.shouldRun(rule) {
			continue
		}
		findings := rule.Check(snap)
		res.findings = append(res.findings, findings...)
	}

	// Sort: critical first, then by rule ID.
	sort.SliceStable(res.findings, func(i, j int) bool {
		if res.findings[i].Severity != res.findings[j].Severity {
			return res.findings[i].Severity < res.findings[j].Severity
		}
		return res.findings[i].RuleID < res.findings[j].RuleID
	})

	// Build summary.
	res.summary.Total = len(res.findings)
	for _, f := range res.findings {
		switch f.Severity {
		case SeverityCritical:
			res.summary.Critical++
		case SeverityHigh:
			res.summary.High++
		case SeverityMedium:
			res.summary.Medium++
		case SeverityLow:
			res.summary.Low++
		case SeverityInfo:
			res.summary.Info++
		}
	}
	return res, nil
}

func (e *Engine) shouldRun(r Rule) bool {
	if len(e.opts.Include) > 0 && !contains(e.opts.Include, r.ID()) {
		return false
	}
	if len(e.opts.Exclude) > 0 && contains(e.opts.Exclude, r.ID()) {
		return false
	}
	return int(r.Severity()) <= e.opts.MinTier
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func snapshotFromMCP(s *mcp.Snapshot) *Snapshot {
	return &Snapshot{
		ServerInfo: s.ServerInfo,
		Tools:      s.Tools,
		Resources:  s.Resources,
		Prompts:    s.Prompts,
	}
}

// ShouldFail returns true if findings exceed the fail-on threshold.
func (r *Results) ShouldFail(failOn string) bool {
	threshold := severityFromName(failOn)
	if threshold == 0 {
		return false
	}
	for _, f := range r.findings {
		if int(f.Severity) <= threshold {
			return true
		}
	}
	return false
}

func severityFromName(name string) int {
	switch name {
	case "critical":
		return 1
	case "high":
		return 2
	case "medium":
		return 3
	case "low":
		return 4
	case "info":
		return 5
	}
	return 0
}

// Renderable converts findings to a list of FindingResult for the text formatter.
func (r *Results) Renderable() []output.FindingResult {
	out := make([]output.FindingResult, len(r.findings))
	for i, f := range r.findings {
		out[i] = f.Renderable()
	}
	return out
}

// Raw returns a JSON-serializable representation.
func (r *Results) Raw() any {
	return map[string]any{
		"findings": r.findings,
		"summary":  r.summary,
	}
}

// Summary returns the scan summary.
func (r *Results) Summary() output.ScanSummary {
	return r.summary
}
