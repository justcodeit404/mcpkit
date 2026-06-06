# Changelog

All notable changes to mcpkit will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-07

### Added
- `mcpkit probe` — interactive REPL for exploring MCP servers
- `mcpkit test` — protocol compliance testing with 20 checks
- `mcpkit scan` — security scanner with 21 rules across 5 tiers (CRITICAL/HIGH/MEDIUM/LOW/INFO)
- `mcpkit bench` — performance benchmarking with percentile stats and concurrency support
- Support for stdio and streamable-http transports
- Beautiful terminal output via Charmbracelet lipgloss
- JSON output for CI/CD integration
- Cross-platform binaries (Linux, macOS, Windows on amd64/arm64)
- Unit tests for benchmark, scanner, mcp, output, validator packages
- `--version` flag with build-time version injection
- Bench `--concurrency` flag for parallel benchmarking

### Notes
- `mcpkit fuzz` deferred to v0.3.0
- `mcpkit new` and `mcpkit validate` deferred to v0.2.0
