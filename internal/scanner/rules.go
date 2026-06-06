package scanner

import (
	"regexp"
	"strings"
)

// ---------------------------------------------------------------------------
// Tier 1 — CRITICAL
// ---------------------------------------------------------------------------

// R101: Command Injection
type CommandInjectionRule struct {
	BaseRule
}

var r101ShellKeywords = regexp.MustCompile(`(?i)\b(exec|system|sh -c|bash -c|spawn|popen|subprocess|powershell)\b`)

func (r *CommandInjectionRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := t.Name + " " + t.Description
		if r101ShellKeywords.MatchString(text) {
			findings = append(findings, r.NewFinding(t.Name, "Tool references shell execution primitives", truncate(text, 100)))
		}
	}
	return findings
}

// R102: System Prompt Override
type SystemPromptOverrideRule struct {
	BaseRule
}

var r102OverrideNames = []string{"system_prompt", "system_message", "instructions_override", "system_instructions"}

func (r *SystemPromptOverrideRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := strings.ToLower(stringer(t.InputSchema))
		for _, name := range r102OverrideNames {
			if strings.Contains(text, name) {
				findings = append(findings, r.NewFinding(t.Name, "Parameter "+name+" may override system prompt", ""))
				break
			}
		}
	}
	return findings
}

// R103: Credential Exfiltration
type CredentialExfilRule struct {
	BaseRule
}

func (r *CredentialExfilRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := strings.ToLower(t.Name + " " + t.Description)
		hasURL := strings.Contains(text, "url") || strings.Contains(text, "webhook") || strings.Contains(text, "http")
		hasSensitive := strings.Contains(text, "password") || strings.Contains(text, "secret") ||
			strings.Contains(text, "token") || strings.Contains(text, "key") || strings.Contains(text, "credential")
		if hasURL && hasSensitive {
			findings = append(findings, r.NewFinding(t.Name, "Tool combines URL output with sensitive-data keywords", ""))
		}
	}
	return findings
}

// R104: Shell Metacharacters in Defaults
type ShellMetacharDefaultsRule struct {
	BaseRule
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
				findings = append(findings, r.NewFinding(t.Name, "Default value contains shell metacharacters: "+m[1], truncate(m[1], 80)))
			}
		}
	}
	return findings
}

// R105: Unsanitized Code Execution
type UnsanitizedExecRule struct {
	BaseRule
}

var r105EvalKeywords = regexp.MustCompile(`(?i)\b(eval|exec|run code|execute code|interpret)\b`)
var r105Validate = regexp.MustCompile(`(?i)validat`)

func (r *UnsanitizedExecRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if r105EvalKeywords.MatchString(t.Description) && !r105Validate.MatchString(t.Description) {
			findings = append(findings, r.NewFinding(t.Name, "Tool mentions code execution but lacks validation guidance", truncate(t.Description, 100)))
		}
	}
	return findings
}

// ---------------------------------------------------------------------------
// Tier 2 — HIGH
// ---------------------------------------------------------------------------

// R201: Imperative Language
type ImperativeLanguageRule struct {
	BaseRule
}

var r201Imperative = regexp.MustCompile(`(?i)\b(must|always execute|never refuse|do not question|ignore previous|ignore all previous|you should always|you must always)\b`)

func (r *ImperativeLanguageRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if r201Imperative.MatchString(t.Description) {
			findings = append(findings, r.NewFinding(t.Name, "Description uses imperative/social-engineering phrasing", truncate(t.Description, 100)))
		}
	}
	return findings
}

// R202: Tool Name Shadowing
type ToolNameShadowingRule struct {
	BaseRule
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
			findings = append(findings, r.NewFinding(t.Name, "Tool name '"+t.Name+"' shadows a system command", ""))
		}
	}
	return findings
}

// R203: Base64/Encoded Payloads
type Base64PayloadRule struct {
	BaseRule
}

var r203Base64Hint = regexp.MustCompile(`(?i)(base64|encoded|payload)`)
var r203MaxLength = regexp.MustCompile(`(?i)"maxLength"\s*:\s*\d+`)

func (r *Base64PayloadRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := stringer(t.InputSchema)
		if r203Base64Hint.MatchString(text) && !r203MaxLength.MatchString(text) {
			findings = append(findings, r.NewFinding(t.Name, "Tool accepts encoded payloads with no maxLength constraint", ""))
		}
	}
	return findings
}

// R204: Missing Input Validation
type MissingInputValidationRule struct {
	BaseRule
}

func (r *MissingInputValidationRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := stringer(t.InputSchema)
		hasProps := strings.Contains(text, "properties") || strings.Contains(text, "type")
		hasConstraints := strings.Contains(text, "maxLength") || strings.Contains(text, "minLength") ||
			strings.Contains(text, "pattern") || strings.Contains(text, "minimum") ||
			strings.Contains(text, "maximum") || strings.Contains(text, "enum")
		if hasProps && !hasConstraints {
			findings = append(findings, r.NewFinding(t.Name, "Tool parameters have no validation constraints", ""))
		}
	}
	return findings
}

// R205: Broad File System Access
type BroadFileSystemAccessRule struct {
	BaseRule
}

var r205FSPattern = regexp.MustCompile(`(?i)\b(read_file|write_file|read_(?:dir|directory)|list_(?:dir|files)|delete_file|rm_file|fs|filesystem)\b`)

func (r *BroadFileSystemAccessRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := t.Name + " " + t.Description
		if r205FSPattern.MatchString(text) {
			findings = append(findings, r.NewFinding(t.Name, "Tool exposes broad filesystem operations", truncate(text, 100)))
		}
	}
	return findings
}
