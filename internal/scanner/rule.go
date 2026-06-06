// Package scanner runs security rule checks against an MCP server.
package scanner

// Severity represents a finding's severity tier.
type Severity int

const (
	SeverityCritical Severity = 1
	SeverityHigh     Severity = 2
	SeverityMedium   Severity = 3
	SeverityLow      Severity = 4
	SeverityInfo     Severity = 5
)

// String returns the lowercase name of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityCritical:
		return "critical"
	case SeverityHigh:
		return "high"
	case SeverityMedium:
		return "medium"
	case SeverityLow:
		return "low"
	case SeverityInfo:
		return "info"
	}
	return "unknown"
}

// SeverityFromName parses a severity name string into a Severity.
// Returns 0 (unknown) for unrecognized names.
func SeverityFromName(name string) Severity {
	switch name {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	case "info":
		return SeverityInfo
	}
	return 0
}

// BaseRule provides the metadata fields for all scanner rules.
// Embed this in rule structs so only Check() needs implementation.
type BaseRule struct {
	IDVal          string
	NameVal        string
	SeverityVal    Severity
	DescriptionVal string
	RemediationVal string
}

func (b BaseRule) ID() string          { return b.IDVal }
func (b BaseRule) Name() string        { return b.NameVal }
func (b BaseRule) Severity() Severity  { return b.SeverityVal }
func (b BaseRule) Description() string { return b.DescriptionVal }
func (b BaseRule) Remediation() string { return b.RemediationVal }

// NewFinding creates a Finding with metadata from the rule pre-filled.
func (b BaseRule) NewFinding(target, description, evidence string) Finding {
	return Finding{
		RuleID:      b.IDVal,
		RuleName:    b.NameVal,
		Severity:    b.SeverityVal,
		Target:      target,
		Description: description,
		Evidence:    evidence,
		Remediation: b.RemediationVal,
	}
}

// Rule is the interface implemented by every security rule.
type Rule interface {
	ID() string
	Name() string
	Severity() Severity
	Description() string
	Remediation() string
	// Check runs the rule against a server snapshot.
	// The rule should NOT modify the snapshot.
	Check(snap *Snapshot) []Finding
}
