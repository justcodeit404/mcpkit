# mcpkit

> **The Swiss Army knife for MCP development.**
> Test, scan, bench, fuzz, and probe your MCP servers with a single, fast, zero-dependency binary.

```
$ mcpkit scan --command "npx -y @modelcontextprotocol/server-filesystem /tmp"

🛠  mcpkit scan — MCP server security scan
  Server:           filesystem@<version>

Rule    Severity   Target          Finding
R202    ▲ HIGH     read_file       Tool name 'read_file' shadows a system command
R205    ▲ HIGH     read_file       Tool exposes broad filesystem operations

Summary
  2 findings: 0 critical, 2 high, 0 medium, 0 low, 0 info
```

## ✨ Why mcpkit?

The MCP ecosystem has 16,000+ servers and 150M+ SDK downloads — but no Go-native, single-binary toolkit for developers who want to ship reliable, secure MCP servers.

**mcpkit fills that gap.** It's the missing CLI for the MCP protocol, the way `curl` is the universal HTTP tool, or how `psql` is the canonical PostgreSQL client.

## 🚀 Quick Start

```bash
# Install
brew install mcpkit
# or
go install github.com/justcodeit404/mcpkit/cmd/mcpkit@latest

# Probe interactively
mcpkit probe --command "npx -y @modelcontextprotocol/server-filesystem /tmp"

# Run protocol compliance tests
mcpkit test --command "npx -y @modelcontextprotocol/server-filesystem /tmp"

# Security scan
mcpkit scan --command "npx -y @modelcontextprotocol/server-filesystem /tmp"

# Benchmark performance
mcpkit bench --command "./my-server" --method ping -n 1000
```

## 🧰 Commands

| Command | Description |
|---------|-------------|
| `mcpkit probe` | Interactive REPL for exploring MCP servers |
| `mcpkit test` | Protocol compliance testing (20 checks) |
| `mcpkit scan` | Security vulnerability scanning (10 rules, 2 tiers) |
| `mcpkit bench` | Performance benchmarking with percentile stats |
| `mcpkit fuzz` | Protocol fuzzing (coming in v0.3.0) |
| `mcpkit new` | Scaffold a new MCP server (coming in v0.2.0) |
| `mcpkit validate` | Validate `mcp.json` configuration (coming in v0.2.0) |

## ⚔️ Comparison

| Feature | MCP Inspector | mcp-server-doctor | MCPLint | **mcpkit** |
|---------|:---:|:---:|:---:|:---:|
| Language | Node.js | Node.js | Rust | **Go** |
| Single binary | ❌ | ❌ | ✅ | ✅ |
| Interactive REPL | ⚠️ Web UI | ❌ | ❌ | ✅ |
| Spec compliance | ⚠️ Partial | ⚠️ Partial | ✅ | ✅ |
| Security scanning | ❌ | ❌ | ⚠️ | ✅ |
| Benchmarking | ❌ | ⚠️ Basic | ❌ | ✅ |
| Protocol fuzzing | ❌ | ❌ | ❌ | ✅ (v0.3.0) |
| CI/CD JSON output | ⚠️ Partial | ⚠️ | ❌ | ✅ |
| Cross-platform | ⚠️ Limited | ⚠️ | ✅ | ✅ |
| Zero npm/node deps | ❌ | ❌ | ✅ | ✅ |

## 📦 Installation

```bash
# Homebrew (macOS/Linux)
brew install mcpkit

# Go install
go install github.com/justcodeit404/mcpkit/cmd/mcpkit@latest

# Direct binary download
# See https://github.com/justcodeit404/mcpkit/releases/latest

# Docker
docker run --rm -it ghcr.io/justcodeit404/mcpkit --help
```

## 🎯 Why mcpkit Wins

- **Zero dependencies** — single statically-linked binary, no Node.js, no Python
- **Fast** — Go + minimal memory; sub-millisecond startup
- **Beautiful output** — terminal UI designed with Charmbracelet lipgloss
- **CI/CD native** — JSON output for GitHub Actions, GitLab CI, Jenkins
- **Cross-platform** — Windows, macOS, Linux from the same source

## 🔍 What mcpkit test Checks (v0.1.0)

| Check | Category | What it verifies |
|-------|----------|------------------|
| HND-001..005 | Handshake | initialize succeeded, protocol version, server info, capabilities, post-init ping |
| TL-001..004 | Tools list | response valid, name format, description present, inputSchema present |
| TC-001..004 | Tools call | succeeds, unknown tool returns error, missing args handled, type validation |
| RL-001, RR-001 | Resources | list returns valid response, read returns content |
| PL-001, PG-001..002 | Prompts | list returns valid response, get returns messages, missing args handled |
| PING-01 | Core | ping returns empty result |

## 🛡️ What mcpkit scan Detects (v0.1.0)

**Tier 1 — CRITICAL (5 rules)**

- **R101** — Command Injection: Tool references shell primitives with user input
- **R102** — System Prompt Override: Parameter accepts system_prompt/instructions
- **R103** — Credential Exfiltration: Tool combines URL output with sensitive keywords
- **R104** — Shell Metacharacters in Defaults: Default values contain `; | & $ \``
- **R105** — Unsanitized Code Execution: eval/exec references without validation

**Tier 2 — HIGH (5 rules)**

- **R201** — Imperative Language: "must", "always execute", "ignore previous"
- **R202** — Tool Name Shadowing: Names collide with `ls`, `cat`, `curl`, `bash`, etc.
- **R203** — Base64 Payloads: Encoded parameters with no max size
- **R204** — Missing Input Validation: No JSON Schema constraints on parameters
- **R205** — Broad File System Access: Arbitrary path reads/writes

## 🧪 Development

```bash
# Build
make build

# Run tests
make test

# Lint
make lint

# Snapshot release
make release-snapshot
```

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, and browse [good-first-issues](https://github.com/justcodeit404/mcpkit/labels/good%20first%20issue) to get started.

## 📜 License

MIT — see [LICENSE](LICENSE).

## 🙏 Acknowledgments

- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) — the official Go SDK
- [Charmbracelet](https://charmbracelet.com/) — beautiful terminal UI
- [MCP Inspector](https://github.com/modelcontextprotocol/inspector) — inspiration
- [Stacklok MCP Security Checklist](https://stacklok.com/blog/the-mcp-security-checklist-what-to-verify-before-you-ship-an-mcp-server-to-production/) — security rule references

---

**Made with 🛠 for the MCP community.**
