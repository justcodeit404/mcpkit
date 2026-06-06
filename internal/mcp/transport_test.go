package mcp

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input    string
		wantCmd  string
		wantArgs []string
		wantErr  bool
	}{
		{"npx -y some-server /tmp", "npx", []string{"-y", "some-server", "/tmp"}, false},
		{"go run ./cmd/server", "go", []string{"run", "./cmd/server"}, false},
		{"./server", "./server", nil, false},
		{"", "", nil, true},
	}
	for _, tt := range tests {
		cmd, args, err := ParseCommand(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseCommand(%q): expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseCommand(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if cmd != tt.wantCmd {
			t.Errorf("ParseCommand(%q): got cmd %q, want %q", tt.input, cmd, tt.wantCmd)
		}
		if len(args) != len(tt.wantArgs) {
			t.Errorf("ParseCommand(%q): got %d args, want %d", tt.input, len(args), len(tt.wantArgs))
			continue
		}
		for i, a := range args {
			if a != tt.wantArgs[i] {
				t.Errorf("ParseCommand(%q): arg[%d] = %q, want %q", tt.input, i, a, tt.wantArgs[i])
			}
		}
	}
}
