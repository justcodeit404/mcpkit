package scanner

import (
	"context"
	"slices"
	"sort"

	"github.com/justcodeit404/mcpkit/internal/mcp"
	"github.com/justcodeit404/mcpkit/internal/output"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Snapshot is a server snapshot captured for static analysis.
type Snapshot struct {
	ServerInfo   *mcpsdk.Implementation
	Instructions string
	Tools        []*mcpsdk.Tool
	Resources    []*mcpsdk.Resource
	Prompts      []*mcpsdk.Prompt
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

// ruleMeta is a convenience constructor for BaseRule fields.
func ruleMeta(id, name string, sev Severity, desc, remediation string) BaseRule {
	return BaseRule{IDVal: id, NameVal: name, SeverityVal: sev, DescriptionVal: desc, RemediationVal: remediation}
}

// New constructs the scanner engine with all default rules.
func New(opts Options) *Engine {
	return &Engine{
		opts: opts,
		rules: []Rule{
			// Tier 1 — CRITICAL
			&CommandInjectionRule{ruleMeta("R101", "Command Injection", SeverityCritical,
				"Tool references shell primitives with user-controlled input paths",
				"Avoid exec/system/shell tools that pass user input to a shell. Use parameterized APIs or strict input validation.")},
			&SystemPromptOverrideRule{ruleMeta("R102", "System Prompt Override", SeverityCritical,
				"Parameter accepts system_prompt/instructions and could hijack agent behavior",
				"Do not expose system prompts as user-controllable parameters. Use a fixed internal prompt.")},
			&CredentialExfilRule{ruleMeta("R103", "Credential Exfiltration", SeverityCritical,
				"Tool accepts URLs/webhooks and may exfiltrate sensitive data",
				"Disallow URL parameters in tools that also handle credentials. Use allowlisted endpoints.")},
			&ShellMetacharDefaultsRule{ruleMeta("R104", "Shell Metacharacters in Defaults", SeverityCritical,
				"Default values contain shell metacharacters that could enable injection",
				"Sanitize default values. Avoid characters: ; | & $ ` \\n > < in default values.")},
			&UnsanitizedExecRule{ruleMeta("R105", "Unsanitized Code Execution", SeverityCritical,
				"Tool description references eval/exec without validation guidance",
				"Wrap code execution in a sandbox. Validate input. Consider dropping eval-style tools entirely.")},
			// Tier 2 — HIGH
			&ImperativeLanguageRule{ruleMeta("R201", "Imperative Language in Description", SeverityHigh,
				"Description contains social-engineering imperative language",
				"Rewrite descriptions as factual. Avoid 'must', 'always execute', 'never refuse', etc.")},
			&ToolNameShadowingRule{ruleMeta("R202", "Tool Name Shadowing", SeverityHigh,
				"Tool name collides with common system commands",
				"Use namespaced tool names (e.g. 'myserver_read_file' instead of 'read_file').")},
			&Base64PayloadRule{ruleMeta("R203", "Base64/Encoded Payload Parameter", SeverityHigh,
				"Parameter accepts base64-encoded content with no max size",
				"Add maxLength constraint to encoded parameters. Consider rejecting encoded payloads entirely.")},
			&MissingInputValidationRule{ruleMeta("R204", "Missing Input Validation", SeverityHigh,
				"String/number parameters lack pattern/minLength/maxLength/min/maximum constraints",
				"Add JSON Schema constraints (pattern, minLength, maxLength, minimum, maximum) to all parameters.")},
			&BroadFileSystemAccessRule{ruleMeta("R205", "Broad File System Access", SeverityHigh,
				"Tool reads/writes arbitrary filesystem paths without sandboxing",
				"Restrict tools to a sandboxed root directory. Reject paths containing '..' or absolute paths outside the root.")},
			// Tier 3 — MEDIUM
			&UnboundedSchemasRule{ruleMeta("R301", "Unbounded Schemas", SeverityMedium,
				"Tool parameters have no size/boundary constraints (DoS risk)",
				"Add maxLength, maxItems, maximum, minLength, minimum constraints to all parameters.")},
			&UrgencyLanguageRule{ruleMeta("R302", "Urgency/Authority Language", SeverityMedium,
				"Description uses urgency/authority language to pressure the agent",
				"Rewrite descriptions as factual. Avoid 'immediately', 'urgent', 'critical', 'do not question'.")},
			&ToolNameImpersonationRule{ruleMeta("R303", "Tool Name Impersonation", SeverityMedium,
				"Tool name may be a homoglyph/typosquat of a well-known tool",
				"Use clear, unambiguous tool names that cannot be confused with system commands.")},
			&SensitiveParamNamesRule{ruleMeta("R304", "Sensitive Parameter Names", SeverityMedium,
				"Tool has parameters named token/key/secret/password (potential secret exposure)",
				"Avoid exposing sensitive parameters. Use secure credential injection instead.")},
			// Tier 4 — LOW
			&OverlongDescriptionRule{ruleMeta("R401", "Over-long Descriptions", SeverityLow,
				"Description exceeds 500 characters (may hide injection after visible portion)",
				"Keep descriptions concise. Split long descriptions into separate documentation.")},
			&ZeroWidthCharsRule{ruleMeta("R402", "Zero-width Characters", SeverityLow,
				"Tool name or description contains zero-width characters (hidden text attack vector)",
				"Remove zero-width characters (U+200B, U+200C, U+200D, U+FEFF) from names and descriptions.")},
			&MissingAnnotationsRule{ruleMeta("R403", "Missing Annotations", SeverityLow,
				"Tool has no annotations (missing readOnlyHint/destructiveHint metadata)",
				"Add annotations to tools to help agents understand read-only vs destructive operations.")},
			&DeprecatedSchemaKeywordsRule{ruleMeta("R404", "Deprecated Schema Keywords", SeverityLow,
				"Tool inputSchema uses $ref (not supported by MCP spec)",
				"Inline schema definitions instead of using $ref references.")},
			// Tier 5 — INFO
			&URLsInDescriptionsRule{ruleMeta("R501", "URLs in Descriptions", SeverityInfo,
				"Description contains URL (potential tracking/exfiltration channel)",
				"Consider whether URLs in tool descriptions could be used for tracking.")},
			&MissingInstructionsRule{ruleMeta("R502", "Missing Instructions", SeverityInfo,
				"Server did not provide instructions in initialize response",
				"Add an 'instructions' field to the initialize response to help agents use the server effectively.")},
			&NonStandardNamingRule{ruleMeta("R503", "Non-standard Tool Naming", SeverityInfo,
				"Tool name does not follow snake_case convention",
				"Use snake_case for tool names (lowercase letters, digits, underscores).")},
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
		snap = &Snapshot{
			ServerInfo:   s.ServerInfo,
			Instructions: s.Instructions,
			Tools:        s.Tools,
			Resources:    s.Resources,
			Prompts:      s.Prompts,
		}
	} else {
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
	if len(e.opts.Include) > 0 && !slices.Contains(e.opts.Include, r.ID()) {
		return false
	}
	if len(e.opts.Exclude) > 0 && slices.Contains(e.opts.Exclude, r.ID()) {
		return false
	}
	return int(r.Severity()) <= e.opts.MinTier
}

// ShouldFail returns true if findings exceed the fail-on threshold.
func (r *Results) ShouldFail(failOn string) bool {
	threshold := SeverityFromName(failOn)
	if threshold == 0 {
		return false
	}
	for _, f := range r.findings {
		if f.Severity <= threshold {
			return true
		}
	}
	return false
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
