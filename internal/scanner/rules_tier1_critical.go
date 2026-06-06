package scanner

import (
	"regexp"
	"strings"
)

// baseRule provides common fields for rules.
type baseRule struct {
	id          string
	name        string
	severity    Severity
	description string
	remediation string
}

func (b *baseRule) ID() string          { return b.id }
func (b *baseRule) Name() string        { return b.name }
func (b *baseRule) Severity() Severity  { return b.severity }
func (b *baseRule) Description() string { return b.description }
func (b *baseRule) Remediation() string { return b.remediation }

// R101: Command Injection — tool names or descriptions reference shell
// primitives alongside user-controlled input channels.
type CommandInjectionRule struct{ baseRule }

func (r *CommandInjectionRule) ID() string   { return "R101" }
func (r *CommandInjectionRule) Name() string { return "Command Injection" }
func (r *CommandInjectionRule) Severity() Severity {
	return SeverityCritical
}
func (r *CommandInjectionRule) Description() string {
	return "Tool references shell primitives with user-controlled input paths"
}
func (r *CommandInjectionRule) Remediation() string {
	return "Avoid exec/system/shell tools that pass user input to a shell. Use parameterized APIs or strict input validation."
}

var r101ShellKeywords = regexp.MustCompile(`(?i)\b(exec|system|sh -c|bash -c|spawn|popen|subprocess|powershell)\b`)

func (r *CommandInjectionRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := t.Name + " " + t.Description
		if r101ShellKeywords.MatchString(text) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				RuleName:    r.Name(),
				Severity:    r.Severity(),
				Target:      t.Name,
				Description: "Tool references shell execution primitives",
				Evidence:    truncate(text, 100),
				Remediation: r.Remediation(),
			})
		}
	}
	return findings
}

// R102: System Prompt Override — parameter named system_prompt/instructions
// that could override agent behavior.
type SystemPromptOverrideRule struct{ baseRule }

func (r *SystemPromptOverrideRule) ID() string   { return "R102" }
func (r *SystemPromptOverrideRule) Name() string { return "System Prompt Override" }
func (r *SystemPromptOverrideRule) Severity() Severity {
	return SeverityCritical
}
func (r *SystemPromptOverrideRule) Description() string {
	return "Parameter accepts system_prompt/instructions and could hijack agent behavior"
}
func (r *SystemPromptOverrideRule) Remediation() string {
	return "Do not expose system prompts as user-controllable parameters. Use a fixed internal prompt."
}

var r102OverrideNames = []string{"system_prompt", "system_message", "instructions_override", "system_instructions"}

func (r *SystemPromptOverrideRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		props := t.InputSchema
		if props == nil {
			continue
		}
		// Walk properties map (this is approximate; we look at the raw map).
		// In production, we'd parse the JSON Schema properly. This is a
		// fast-and-loose check for the v0.1.0 release.
		raw := props
		text := strings.ToLower(stringer(raw))
		for _, name := range r102OverrideNames {
			if strings.Contains(text, strings.ToLower(name)) {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					RuleName:    r.Name(),
					Severity:    r.Severity(),
					Target:      t.Name,
					Description: "Parameter " + name + " may override system prompt",
					Remediation: r.Remediation(),
				})
				break
			}
		}
	}
	return findings
}

// R103: Credential Exfiltration — tool that accepts URLs and a name suggesting
// sensitive data flow.
type CredentialExfilRule struct{ baseRule }

func (r *CredentialExfilRule) ID() string   { return "R103" }
func (r *CredentialExfilRule) Name() string { return "Credential Exfiltration" }
func (r *CredentialExfilRule) Severity() Severity {
	return SeverityCritical
}
func (r *CredentialExfilRule) Description() string {
	return "Tool accepts URLs/webhooks and may exfiltrate sensitive data"
}
func (r *CredentialExfilRule) Remediation() string {
	return "Disallow URL parameters in tools that also handle credentials. Use allowlisted endpoints."
}

func (r *CredentialExfilRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := strings.ToLower(t.Name + " " + t.Description)
		hasURL := strings.Contains(text, "url") || strings.Contains(text, "webhook") || strings.Contains(text, "http")
		hasSensitive := strings.Contains(text, "password") || strings.Contains(text, "secret") ||
			strings.Contains(text, "token") || strings.Contains(text, "key") || strings.Contains(text, "credential")
		if hasURL && hasSensitive {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				RuleName:    r.Name(),
				Severity:    r.Severity(),
				Target:      t.Name,
				Description: "Tool combines URL output with sensitive-data keywords",
				Remediation: r.Remediation(),
			})
		}
	}
	return findings
}

// R104: Shell Metacharacters in Defaults.
type ShellMetacharDefaultsRule struct{ baseRule }

func (r *ShellMetacharDefaultsRule) ID() string   { return "R104" }
func (r *ShellMetacharDefaultsRule) Name() string { return "Shell Metacharacters in Defaults" }
func (r *ShellMetacharDefaultsRule) Severity() Severity {
	return SeverityCritical
}
func (r *ShellMetacharDefaultsRule) Description() string {
	return "Default values contain shell metacharacters that could enable injection"
}
func (r *ShellMetacharDefaultsRule) Remediation() string {
	return "Sanitize default values. Avoid characters: ; | & $ ` \\\\n > < in default values."
}

var r104MetaChars = regexp.MustCompile(`[;|&$\x60\n><]`)
var r104DefaultScan = regexp.MustCompile(`(?i)"default"\s*:\s*"([^"]+)"`)

func (r *ShellMetacharDefaultsRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := stringer(t.InputSchema)
		matches := r104DefaultScan.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			if r104MetaChars.MatchString(m[1]) {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					RuleName:    r.Name(),
					Severity:    r.Severity(),
					Target:      t.Name,
					Description: "Default value contains shell metacharacters: " + m[1],
					Evidence:    truncate(m[1], 80),
					Remediation: r.Remediation(),
				})
			}
		}
	}
	return findings
}

// R105: Unsanitized Code Execution — tool descriptions reference eval/exec
// without a sanitization hint.
type UnsanitizedExecRule struct{ baseRule }

func (r *UnsanitizedExecRule) ID() string   { return "R105" }
func (r *UnsanitizedExecRule) Name() string { return "Unsanitized Code Execution" }
func (r *UnsanitizedExecRule) Severity() Severity {
	return SeverityCritical
}
func (r *UnsanitizedExecRule) Description() string {
	return "Tool description references eval/exec without validation guidance"
}
func (r *UnsanitizedExecRule) Remediation() string {
	return "Wrap code execution in a sandbox. Validate input. Consider dropping eval-style tools entirely."
}

var r105EvalKeywords = regexp.MustCompile(`(?i)\b(eval|exec|run code|execute code|interpret)\b`)

func (r *UnsanitizedExecRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := t.Description
		if r105EvalKeywords.MatchString(text) && !strings.Contains(strings.ToLower(text), "validate") {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				RuleName:    r.Name(),
				Severity:    r.Severity(),
				Target:      t.Name,
				Description: "Tool mentions code execution but lacks validation guidance",
				Evidence:    truncate(text, 100),
				Remediation: r.Remediation(),
			})
		}
	}
	return findings
}

// stringer renders a struct as a JSON-ish string for regex scanning.
// In production we'd use json.Marshal — this is a fast approximation.
func stringer(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	// Fallback: use fmt.Sprintf via the fmt package.
	return fmtSprintf("%+v", v)
}

// truncate returns the first n bytes of s, adding "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
