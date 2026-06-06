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
