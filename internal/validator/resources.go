package validator

import (
	"context"
	"strconv"
	"time"

	"github.com/justcodeit404/mcpkit/internal/mcp"
)

// runResourceChecks verifies resources/list and resources/read behavior.
func (s *Spec) runResourceChecks(ctx context.Context, client *mcp.Client, r *Results) {
	start := time.Now()
	resources, err := client.ListResources(ctx)
	if err != nil {
		r.record("RL-001", "resources/list returns valid response", "fail",
			"resources/list error: "+err.Error(), time.Since(start))
		r.record("RR-001", "resources/read returns content", "skip",
			"requires RL-001", 0)
		return
	}
	r.record("RL-001", "resources/list returns valid response", "pass",
		"server returned "+strconv.Itoa(len(resources))+" resources", time.Since(start))

	if s.resourceFilter == "" {
		r.record("RR-001", "resources/read returns content", "skip",
			"specify --resource to test read", 0)
		return
	}

	start = time.Now()
	_, err = client.ReadResource(ctx, s.resourceFilter)
	if err != nil {
		r.record("RR-001", "resources/read returns content", "fail",
			"read error: "+err.Error(), time.Since(start))
	} else {
		r.record("RR-001", "resources/read returns content", "pass",
			"resource read successfully", time.Since(start))
	}
}
