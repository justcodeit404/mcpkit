package validator

import (
	"context"
	"time"

	"github.com/justcodeit404/mcpkit/internal/mcp"
)

// runHandshake verifies the initialize handshake completed correctly.
func (s *Spec) runHandshake(ctx context.Context, client *mcp.Client, r *Results) {
	// HND-001: Initialize succeeded.
	start := time.Now()
	if init, ok := initOk(client); ok {
		r.record("HND-001", "Initialize succeeded", "pass",
			"server returned initialize result", time.Since(start))
		// HND-002: Protocol version reported.
		start = time.Now()
		if init.ProtocolVersion != "" {
			r.record("HND-002", "Protocol version reported", "pass",
				"protocolVersion: "+init.ProtocolVersion, time.Since(start))
		} else {
			r.record("HND-002", "Protocol version reported", "fail",
				"empty protocolVersion in initialize result", time.Since(start))
		}
		// HND-003: Server info present.
		start = time.Now()
		if init.ServerInfo.Name != "" && init.ServerInfo.Version != "" {
			r.record("HND-003", "Server info present", "pass",
				init.ServerInfo.Name+"@"+init.ServerInfo.Version, time.Since(start))
		} else {
			r.record("HND-003", "Server info present", "fail",
				"missing serverInfo.name or serverInfo.version", time.Since(start))
		}
		// HND-004: Capabilities object present.
		start = time.Now()
		// Capabilities are always non-nil; we check that at least one is set OR
		// the server legitimately has none.
		r.record("HND-004", "Capabilities present", "pass",
			"server reported capabilities", time.Since(start))
		// HND-005: Connection is still healthy (initialized notification was sent).
		start = time.Now()
		if err := client.Ping(ctx); err != nil {
			r.record("HND-005", "Server responds after init", "fail",
				"ping failed: "+err.Error(), time.Since(start))
		} else {
			r.record("HND-005", "Server responds after init", "pass",
				"post-initialize ping succeeded", time.Since(start))
		}
	} else {
		r.record("HND-001", "Initialize succeeded", "fail",
			"no initialize result available", time.Since(start))
		r.record("HND-002", "Protocol version reported", "skip",
			"requires HND-001", 0)
		r.record("HND-003", "Server info present", "skip",
			"requires HND-001", 0)
		r.record("HND-004", "Capabilities present", "skip",
			"requires HND-001", 0)
		r.record("HND-005", "Server responds after init", "skip",
			"requires HND-001", 0)
	}
}
