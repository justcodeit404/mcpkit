package scanner

import (
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestR101CommandInjection(t *testing.T) {
	rule := &CommandInjectionRule{}
	snap := &Snapshot{
		Tools: []*mcpsdk.Tool{
			{Name: "run_cmd", Description: "Execute a shell command via exec"},
			{Name: "greet", Description: "Say hello"},
		},
	}
	findings := rule.Check(snap)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Target != "run_cmd" {
		t.Errorf("target: got %q, want run_cmd", findings[0].Target)
	}
}

func TestR103CredentialExfil(t *testing.T) {
	rule := &CredentialExfilRule{}
	snap := &Snapshot{
		Tools: []*mcpsdk.Tool{
			{Name: "send_webhook", Description: "Send data to a URL with token authentication"},
			{Name: "read_file", Description: "Read a file from disk"},
		},
	}
	findings := rule.Check(snap)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Target != "send_webhook" {
		t.Errorf("target: got %q, want send_webhook", findings[0].Target)
	}
}

func TestR105UnsanitizedExec(t *testing.T) {
	rule := &UnsanitizedExecRule{}
	snap := &Snapshot{
		Tools: []*mcpsdk.Tool{
			{Name: "eval_code", Description: "eval arbitrary code from user input"},
			{Name: "safe_eval", Description: "evaluate expressions with input validation"},
		},
	}
	findings := rule.Check(snap)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Target != "eval_code" {
		t.Errorf("target: got %q, want eval_code", findings[0].Target)
	}
}

func TestR204MissingValidation(t *testing.T) {
	rule := &MissingInputValidationRule{}
	snap := &Snapshot{
		Tools: []*mcpsdk.Tool{
			{Name: "greet", Description: "Say hello", InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			}},
		},
	}
	findings := rule.Check(snap)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestR204WithValidation(t *testing.T) {
	rule := &MissingInputValidationRule{}
	snap := &Snapshot{
		Tools: []*mcpsdk.Tool{
			{Name: "greet", Description: "Say hello", InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string", "maxLength": 100},
				},
			}},
		},
	}
	findings := rule.Check(snap)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestR205BroadFSAccess(t *testing.T) {
	rule := &BroadFileSystemAccessRule{}
	snap := &Snapshot{
		Tools: []*mcpsdk.Tool{
			{Name: "read_file", Description: "Read a file from the filesystem"},
			{Name: "greet", Description: "Say hello"},
		},
	}
	findings := rule.Check(snap)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestBaseRuleNewFinding(t *testing.T) {
	br := BaseRule{
		IDVal:          "T001",
		NameVal:        "Test Rule",
		SeverityVal:    SeverityHigh,
		DescriptionVal: "A test rule",
		RemediationVal: "Fix it",
	}
	f := br.NewFinding("target1", "desc1", "evidence1")
	if f.RuleID != "T001" {
		t.Errorf("RuleID: got %q, want T001", f.RuleID)
	}
	if f.Target != "target1" {
		t.Errorf("Target: got %q, want target1", f.Target)
	}
	if f.Severity != SeverityHigh {
		t.Errorf("Severity: got %v, want HIGH", f.Severity)
	}
}

func TestSeverityFromName(t *testing.T) {
	tests := []struct {
		name string
		want Severity
	}{
		{"critical", SeverityCritical},
		{"high", SeverityHigh},
		{"medium", SeverityMedium},
		{"low", SeverityLow},
		{"info", SeverityInfo},
		{"unknown", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := SeverityFromName(tt.name)
		if got != tt.want {
			t.Errorf("SeverityFromName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestShouldRunExclude(t *testing.T) {
	engine := New(Options{
		MinTier: 2,
		Exclude: []string{"R101"},
	})
	// R101 should be excluded
	for _, r := range engine.rules {
		if r.ID() == "R101" {
			if engine.shouldRun(r) {
				t.Error("R101 should be excluded but shouldRun returned true")
			}
		} else if r.ID() == "R201" {
			if !engine.shouldRun(r) {
				t.Error("R201 should run but shouldRun returned false")
			}
		}
	}
}

func TestShouldRunInclude(t *testing.T) {
	engine := New(Options{
		MinTier: 2,
		Include: []string{"R101"},
	})
	for _, r := range engine.rules {
		if r.ID() == "R101" {
			if !engine.shouldRun(r) {
				t.Error("R101 should be included but shouldRun returned false")
			}
		} else if r.ID() == "R201" {
			if engine.shouldRun(r) {
				t.Error("R201 should not run but shouldRun returned true")
			}
		}
	}
}
