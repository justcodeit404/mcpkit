package validator

import (
	"testing"
)

func TestToolNamePattern(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"read_file", true},
		{"get-weather", true},
		{"tool.name", true},
		{"A", true},
		{"a123", true},
		{"", false},
		{"has space", false},
		{"has/slash", false},
		{"has@at", false},
	}
	for _, tt := range tests {
		got := toolNamePattern.MatchString(tt.name)
		if got != tt.want {
			t.Errorf("toolNamePattern.MatchString(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestToolNamePatternMaxLength(t *testing.T) {
	// 128 chars should pass
	long := ""
	for i := 0; i < 128; i++ {
		long += "a"
	}
	if !toolNamePattern.MatchString(long) {
		t.Error("128-char name should match")
	}
	// 129 chars should fail
	long += "a"
	if toolNamePattern.MatchString(long) {
		t.Error("129-char name should not match")
	}
}

func TestSkipName(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"TL-002", "Tool name format"},
		{"TC-001", "tools/call succeeds"},
		{"UNKNOWN", "UNKNOWN"},
	}
	for _, tt := range tests {
		got := skipName(tt.id)
		if got != tt.want {
			t.Errorf("skipName(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestResultsSummary(t *testing.T) {
	r := &Results{}
	r.record("HND-001", "test", "pass", "ok", 0)
	r.record("HND-002", "test", "fail", "bad", 0)
	r.record("HND-003", "test", "skip", "n/a", 0)

	// Compute summary manually
	for _, c := range r.checks {
		r.summary.Total++
		switch c.Status {
		case "pass":
			r.summary.Passed++
		case "fail":
			r.summary.Failed++
		case "skip":
			r.summary.Skipped++
		}
	}

	if r.summary.Total != 3 {
		t.Errorf("Total: got %d, want 3", r.summary.Total)
	}
	if r.summary.Passed != 1 {
		t.Errorf("Passed: got %d, want 1", r.summary.Passed)
	}
	if r.summary.Failed != 1 {
		t.Errorf("Failed: got %d, want 1", r.summary.Failed)
	}
	if r.summary.Skipped != 1 {
		t.Errorf("Skipped: got %d, want 1", r.summary.Skipped)
	}
}

func TestShouldRun(t *testing.T) {
	spec := &Spec{methodFilter: "tools/list"}
	if !spec.shouldRun("tools/list") {
		t.Error("shouldRun(tools/list) should be true with filter=tools/list")
	}
	if spec.shouldRun("ping") {
		t.Error("shouldRun(ping) should be false with filter=tools/list")
	}

	spec2 := &Spec{}
	if !spec2.shouldRun("anything") {
		t.Error("shouldRun should return true when no filter is set")
	}
}
