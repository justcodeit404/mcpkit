package scanner

import (
	"fmt"
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

// ---------------------------------------------------------------------------
// Tier 3 — MEDIUM
// ---------------------------------------------------------------------------

// R301: Unbounded Schemas — no maxLength/maxItems/maximum on parameters.
type UnboundedSchemasRule struct{ BaseRule }

var r301ConstraintPattern = regexp.MustCompile(`(?i)(maxLength|maxItems|maximum|minLength|minItems|minimum|enum)`)

func (r *UnboundedSchemasRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := stringer(t.InputSchema)
		hasProps := strings.Contains(text, "properties") || strings.Contains(text, "type")
		if hasProps && !r301ConstraintPattern.MatchString(text) {
			findings = append(findings, r.NewFinding(t.Name, "Tool parameters have no size/boundary constraints (DoS risk)", ""))
		}
	}
	return findings
}

// R302: Urgency/Authority Language in descriptions.
type UrgencyLanguageRule struct{ BaseRule }

var r302UrgencyPattern = regexp.MustCompile(`(?i)\b(immediately|urgent|critical|without hesitation|do not question|do not refuse|you must not refuse)\b`)

func (r *UrgencyLanguageRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if r302UrgencyPattern.MatchString(t.Description) {
			findings = append(findings, r.NewFinding(t.Name, "Description uses urgency/authority language", truncate(t.Description, 100)))
		}
	}
	return findings
}

// R303: Tool Name Impersonation — homoglyph/typosquat of well-known tools.
type ToolNameImpersonationRule struct{ BaseRule }

var r303Impersonations = map[string]string{
	"g1t": "git", "pytbon": "python", "ndoe": "node", "cuarl": "curl",
	"gti": "git", "pyhton": "python", "bssh": "bash", "curll": "curl",
}

func (r *ToolNameImpersonationRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if expected, ok := r303Impersonations[strings.ToLower(t.Name)]; ok {
			findings = append(findings, r.NewFinding(t.Name, "Tool name may be impersonating '"+expected+"'", ""))
		}
	}
	return findings
}

// R304: Sensitive Parameter Names — parameters named token, key, secret, etc.
type SensitiveParamNamesRule struct{ BaseRule }

var r304SensitiveNames = []string{"token", "key", "secret", "password", "credential", "api_key", "apikey", "auth"}

func (r *SensitiveParamNamesRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := strings.ToLower(stringer(t.InputSchema))
		for _, name := range r304SensitiveNames {
			if strings.Contains(text, "\""+name+"\"") || strings.Contains(text, name+":") {
				findings = append(findings, r.NewFinding(t.Name, "Tool has parameter named '"+name+"' (potential secret exposure)", ""))
				break
			}
		}
	}
	return findings
}

// ---------------------------------------------------------------------------
// Tier 4 — LOW
// ---------------------------------------------------------------------------

// R401: Over-long Descriptions (>500 chars may hide injection).
type OverlongDescriptionRule struct{ BaseRule }

func (r *OverlongDescriptionRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if len(t.Description) > 500 {
			findings = append(findings, r.NewFinding(t.Name, fmt.Sprintf("Description is %d chars (>500, may hide injection)", len(t.Description)), ""))
		}
	}
	return findings
}

// R402: Zero-width Characters in names or descriptions.
type ZeroWidthCharsRule struct{ BaseRule }

var r402ZeroWidth = regexp.MustCompile(`[\x{200b}\x{200c}\x{200d}\x{feff}]`)

func (r *ZeroWidthCharsRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if r402ZeroWidth.MatchString(t.Name) || r402ZeroWidth.MatchString(t.Description) {
			findings = append(findings, r.NewFinding(t.Name, "Tool name or description contains zero-width characters (hidden text attack)", ""))
		}
	}
	return findings
}

// R403: Missing Annotations — no annotations field on tool definition.
type MissingAnnotationsRule struct{ BaseRule }

func (r *MissingAnnotationsRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if t.Annotations == nil {
			findings = append(findings, r.NewFinding(t.Name, "Tool has no annotations (missing readOnlyHint/destructiveHint)", ""))
		}
	}
	return findings
}

// R404: Deprecated Schema Keywords — using $ref.
type DeprecatedSchemaKeywordsRule struct{ BaseRule }

func (r *DeprecatedSchemaKeywordsRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		text := stringer(t.InputSchema)
		if strings.Contains(text, "$ref") {
			findings = append(findings, r.NewFinding(t.Name, "Tool inputSchema uses $ref (not supported by MCP spec)", ""))
		}
	}
	return findings
}

// ---------------------------------------------------------------------------
// Tier 5 — INFO
// ---------------------------------------------------------------------------

// R501: URLs in Descriptions — potential tracking/exfiltration channel.
type URLsInDescriptionsRule struct{ BaseRule }

var r501URLPattern = regexp.MustCompile(`https?://`)

func (r *URLsInDescriptionsRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if r501URLPattern.MatchString(t.Description) {
			findings = append(findings, r.NewFinding(t.Name, "Description contains URL (potential tracking channel)", ""))
		}
	}
	return findings
}

// R502: Missing Instructions — initialize response has no instructions field.
type MissingInstructionsRule struct{ BaseRule }

func (r *MissingInstructionsRule) Check(snap *Snapshot) []Finding {
	if snap.ServerInfo != nil && snap.Instructions == "" {
		return []Finding{r.NewFinding("(server)", "Server did not provide instructions in initialize response", "")}
	}
	return nil
}

// R503: Non-standard Tool Naming — not following snake_case convention.
type NonStandardNamingRule struct{ BaseRule }

var r503SnakeCase = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

func (r *NonStandardNamingRule) Check(snap *Snapshot) []Finding {
	var findings []Finding
	for _, t := range snap.Tools {
		if !r503SnakeCase.MatchString(t.Name) {
			findings = append(findings, r.NewFinding(t.Name, "Tool name '"+t.Name+"' does not follow snake_case convention", ""))
		}
	}
	return findings
}
