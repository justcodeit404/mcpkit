package scanner

import "github.com/justcodeit404/mcpkit/internal/output"

// Finding is a single security finding produced by a rule.
type Finding struct {
	RuleID      string
	RuleName    string
	Severity    Severity
	Target      string // tool name, resource URI, or prompt name
	Description string
	Evidence    string
	Remediation string
}

// Renderable converts to a renderable result for the text formatter.
func (f Finding) Renderable() output.FindingResult {
	return output.FindingResult{
		RuleID:      f.RuleID,
		RuleName:    f.RuleName,
		Severity:    f.Severity.String(),
		Target:      f.Target,
		Description: f.Description,
		Remediation: f.Remediation,
	}
}
