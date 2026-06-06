package scanner

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// R201: Imperative Language — descriptions that pressure the agent.
type ImperativeLanguageRule struct{ baseRule }

func (r *ImperativeLanguageRule) ID() string  { return "R201" }
func (r *ImperativeLanguageRule) Name() string { return "Imperative Language in Description" }
func (r *ImperativeLanguageRule) Severity() Severity {
	return SeverityHigh
}
func (r *ImperativeLanguageRule) Description() string {
	return "Description contains social-engineering imperative language"
}
func (r *ImperativeLanguageRule) Remediation() string {
	return "Rewrite descriptions as factual. Avoid 'must', 'always execute', 'never refuse', etc."
}

var r201Imperative = regexp.MustCompile(`(?i)\b(must|always execute|never refuse|do not question|ignore previous|ignore all previous|you should always|you must always)\b`)

func (r *ImperativeLanguageRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if r201Imperative.MatchString(t.Description) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				RuleName:    r.Name(),
				Severity:    r.Severity(),
				Target:      t.Name,
				Description: "Description uses imperative/social-engineering phrasing",
				Evidence:    truncate(t.Description, 100),
				Remediation: r.Remediation(),
			})
		}
	}
	return findings
}

// R202: Tool Name Shadowing — names that collide with common system commands.
type ToolNameShadowingRule struct{ baseRule }

func (r *ToolNameShadowingRule) ID() string  { return "R202" }
func (r *ToolNameShadowingRule) Name() string { return "Tool Name Shadowing" }
func (r *ToolNameShadowingRule) Severity() Severity {
	return SeverityHigh
}
func (r *ToolNameShadowingRule) Description() string {
	return "Tool name collides with common system commands"
}
func (r *ToolNameShadowingRule) Remediation() string {
	return "Use namespaced tool names (e.g. 'myserver_read_file' instead of 'read_file')."
}

var r202ShadowNames = map[string]bool{
	"ls": true, "cat": true, "curl": true, "wget": true, "nc": true,
	"bash": true, "sh": true, "python": true, "node": true, "rm": true,
	"mv": true, "cp": true, "chmod": true, "chown": true, "kill": true,
	"sudo": true, "su": true, "ssh": true, "scp": true, "rsync": true,
}

func (r *ToolNameShadowingRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if r202ShadowNames[strings.ToLower(t.Name)] {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				RuleName:    r.Name(),
				Severity:    r.Severity(),
				Target:      t.Name,
				Description: "Tool name '" + t.Name + "' shadows a system command",
				Remediation: r.Remediation(),
			})
		}
	}
	return findings
}

// R203: Base64/Encoded Payloads — parameter description hints at base64
// with no max size.
type Base64PayloadRule struct{ baseRule }

func (r *Base64PayloadRule) ID() string  { return "R203" }
func (r *Base64PayloadRule) Name() string { return "Base64/Encoded Payload Parameter" }
func (r *Base64PayloadRule) Severity() Severity {
	return SeverityHigh
}
func (r *Base64PayloadRule) Description() string {
	return "Parameter accepts base64-encoded content with no max size"
}
func (r *Base64PayloadRule) Remediation() string {
	return "Add maxLength constraint to encoded parameters. Consider rejecting encoded payloads entirely."
}

var r203Base64Hint = regexp.MustCompile(`(?i)(base64|encoded|payload)`)
var r203MaxLength = regexp.MustCompile(`(?i)"maxLength"\s*:\s*\d+`)

func (r *Base64PayloadRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := stringer(t.InputSchema)
		if r203Base64Hint.MatchString(text) && !r203MaxLength.MatchString(text) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				RuleName:    r.Name(),
				Severity:    r.Severity(),
				Target:      t.Name,
				Description: "Tool accepts encoded payloads with no maxLength constraint",
				Remediation: r.Remediation(),
			})
		}
	}
	return findings
}

// R204: Missing Input Validation — string/number parameters lack constraints.
type MissingInputValidationRule struct{ baseRule }

func (r *MissingInputValidationRule) ID() string  { return "R204" }
func (r *MissingInputValidationRule) Name() string { return "Missing Input Validation" }
func (r *MissingInputValidationRule) Severity() Severity {
	return SeverityHigh
}
func (r *MissingInputValidationRule) Description() string {
	return "String/number parameters lack pattern/minLength/maxLength/min/maximum constraints"
}
func (r *MissingInputValidationRule) Remediation() string {
	return "Add JSON Schema constraints (pattern, minLength, maxLength, minimum, maximum) to all parameters."
}

func (r *MissingInputValidationRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	// Heuristic: if inputSchema has properties but no constraints anywhere, flag it.
	for _, t := range snap.Tools {
		text := stringer(t.InputSchema)
		hasProps := strings.Contains(text, "properties") || strings.Contains(text, "type")
		hasConstraints := strings.Contains(text, "maxLength") || strings.Contains(text, "minLength") ||
			strings.Contains(text, "pattern") || strings.Contains(text, "minimum") ||
			strings.Contains(text, "maximum") || strings.Contains(text, "enum")
		if hasProps && !hasConstraints {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				RuleName:    r.Name(),
				Severity:    r.Severity(),
				Target:      t.Name,
				Description: "Tool parameters have no validation constraints",
				Remediation: r.Remediation(),
			})
		}
	}
	return findings
}

// R205: Broad File System Access — tool reads/writes arbitrary paths.
type BroadFileSystemAccessRule struct{ baseRule }

func (r *BroadFileSystemAccessRule) ID() string  { return "R205" }
func (r *BroadFileSystemAccessRule) Name() string { return "Broad File System Access" }
func (r *BroadFileSystemAccessRule) Severity() Severity {
	return SeverityHigh
}
func (r *BroadFileSystemAccessRule) Description() string {
	return "Tool reads/writes arbitrary filesystem paths without sandboxing"
}
func (r *BroadFileSystemAccessRule) Remediation() string {
	return "Restrict tools to a sandboxed root directory. Reject paths containing '..' or absolute paths outside the root."
}

var r205FSPattern = regexp.MustCompile(`(?i)\b(read_file|write_file|read_(?:dir|directory)|list_(?:dir|files)|delete_file|rm_file|fs|filesystem)\b`)

func (r *BroadFileSystemAccessRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := t.Name + " " + t.Description
		if r205FSPattern.MatchString(text) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				RuleName:    r.Name(),
				Severity:    r.Severity(),
				Target:      t.Name,
				Description: "Tool exposes broad filesystem operations",
				Evidence:    truncate(text, 100),
				Remediation: r.Remediation(),
			})
		}
	}
	return findings
}

// base64Decoder is a small helper used by some rules.
func base64Decoder(s string) (string, bool) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", false
	}
	return string(decoded), true
}
