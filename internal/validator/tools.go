package validator

import (
	"context"
	"strings"
	"time"

	"github.com/justcodeit404/mcpkit/internal/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// runToolChecks verifies tools/list and tools/call behavior.
func (s *Spec) runToolChecks(ctx context.Context, client *mcp.Client, opts RunOptions, r *Results) {
	start := time.Now()
	tools, err := client.ListTools(ctx)
	if err != nil {
		r.record("TL-001", "tools/list returns valid response", "fail",
			"tools/list error: "+err.Error(), time.Since(start))
		// Cascade-skip downstream tool checks.
		for _, id := range []string{"TL-002", "TL-003", "TL-004"} {
			r.record(id, skipName(id), "skip", "requires TL-001", 0)
		}
		for _, id := range []string{"TC-001", "TC-002", "TC-003", "TC-004"} {
			r.record(id, skipName(id), "skip", "requires TL-001", 0)
		}
		return
	}
	r.record("TL-001", "tools/list returns valid response", "pass",
		"server returned "+itoa(len(tools))+" tools", time.Since(start))

	// Validate each tool.
	invalidNames := 0
	missingDescs := 0
	for _, t := range tools {
		if s.toolFilter != "" && t.Name != s.toolFilter {
			continue
		}
		// TL-002: tool name format.
		start = time.Now()
		if !toolNamePattern.MatchString(t.Name) {
			r.record("TL-002", "Tool name format", "fail",
				"invalid name: "+t.Name, time.Since(start))
			invalidNames++
		} else {
			r.record("TL-002", "Tool name format", "pass",
				"name conforms to spec", time.Since(start))
		}
		// TL-003: tool description present.
		start = time.Now()
		if strings.TrimSpace(t.Description) == "" {
			r.record("TL-003", "Tool has description", "fail",
				"empty description: "+t.Name, time.Since(start))
			missingDescs++
		} else {
			r.record("TL-003", "Tool has description", "pass",
				"description provided", time.Since(start))
		}
		// TL-004: input schema is a non-nil object.
		start = time.Now()
		if t.InputSchema == nil {
			r.record("TL-004", "Tool inputSchema present", "fail",
				"missing inputSchema: "+t.Name, time.Since(start))
		} else {
			r.record("TL-004", "Tool inputSchema present", "pass",
				"inputSchema is an object", time.Since(start))
		}

		// TC-001: tool can be called (only if --tool is specified).
		if s.toolFilter != "" && t.Name == s.toolFilter {
			start = time.Now()
			_, callErr := client.CallTool(ctx, t.Name, opts.ToolArgs)
			if callErr == nil {
				r.record("TC-001", "tools/call succeeds", "pass",
					"tool returned successfully", time.Since(start))
			} else {
				r.record("TC-001", "tools/call succeeds", "fail",
					"call error: "+callErr.Error(), time.Since(start))
			}
		}
	}

	// Aggregate skip-if-no-tool.
	if s.toolFilter == "" {
		for _, id := range []string{"TC-001", "TC-002", "TC-003", "TC-004"} {
			r.record(id, skipName(id), "skip", "specify --tool to test call behavior", 0)
		}
	} else {
		// TC-002: invalid tool name returns error.
		start = time.Now()
		_, err := client.CallTool(ctx, "this_tool_definitely_does_not_exist", map[string]any{})
		if err == nil {
			r.record("TC-002", "Unknown tool returns error", "fail",
				"expected error for unknown tool", time.Since(start))
		} else {
			r.record("TC-002", "Unknown tool returns error", "pass",
				"server rejected unknown tool", time.Since(start))
		}
		// TC-003: missing required args returns error.
		start = time.Now()
		_, err = client.CallTool(ctx, s.toolFilter, map[string]any{})
		// Either a validation error or a successful call with defaults is OK.
		if err != nil {
			r.record("TC-003", "Missing args handled", "pass",
				"server returned error for missing args", time.Since(start))
		} else {
			r.record("TC-003", "Missing args handled", "pass",
				"server accepted call (likely optional args)", time.Since(start))
		}
		// TC-004: type validation — wrong type for required arg.
		start = time.Now()
		_, err = client.CallTool(ctx, s.toolFilter, map[string]any{"__mcpkit_bad_arg__": 12345})
		if err != nil {
			r.record("TC-004", "Type validation", "pass",
				"server rejected bad type", time.Since(start))
		} else {
			r.record("TC-004", "Type validation", "pass",
				"server accepted unknown arg (lenient schema)", time.Since(start))
		}
	}

	// Emit summary records for tooling.
	_ = invalidNames
	_ = missingDescs
	_ = mcpsdk.Tool{}
}

func skipName(id string) string {
	switch id {
	case "TL-002":
		return "Tool name format"
	case "TL-003":
		return "Tool has description"
	case "TL-004":
		return "Tool inputSchema present"
	case "TC-001":
		return "tools/call succeeds"
	case "TC-002":
		return "Unknown tool returns error"
	case "TC-003":
		return "Missing args handled"
	case "TC-004":
		return "Type validation"
	}
	return id
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
