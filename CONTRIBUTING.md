# Contributing to mcpkit

Thanks for your interest in making mcpkit better! 🎉

## Development Setup

```bash
# Clone
git clone https://github.com/justcodeit404/mcpkit
cd mcpkit

# Build
make build

# Run
./bin/mcpkit --help

# Run tests
make test

# Lint
make lint
```

## Project Structure

```
mcpkit/
├── cmd/mcpkit/main.go            # Entry point
├── internal/
│   ├── cli/                      # Cobra command definitions
│   ├── mcp/                      # MCP client wrapper
│   ├── scanner/                  # Security scanner
│   ├── validator/                # Protocol compliance tests
│   ├── benchmark/                # Performance benchmarking
│   ├── probe/                    # Interactive REPL
│   └── output/                   # Text/JSON formatters
```

## Adding a New Security Rule

1. Add your rule struct in `internal/scanner/rules.go`
2. Embed `BaseRule` and implement only `Check(snap *Snapshot) []Finding`
3. Register it in `internal/scanner/scanner.go` `New()` function
4. Add a unit test in `internal/scanner/rules_test.go`
5. Update the README to list the new rule

## Adding a New Compliance Check

1. Add your check in the appropriate file under `internal/validator/` (handshake, tools, resources, prompts)
2. Use the `r.record()` helper to record a `CheckResult`
3. Update the README's checklist

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add benchmark histogram for tools/call
fix: scanner R203 false positive on optional parameters
docs: update README with new rules
test: add test for tier-1 R101 command injection
```

## Release Process

1. Update version in `internal/cli/root.go`
2. Update CHANGELOG
3. Tag `v0.1.0` and push
4. CI builds binaries, creates GitHub release, updates Homebrew tap

## Code Style

- Format with `gofmt -s -w .`
- Lint with `golangci-lint run`
- 100-character line limit
- All exported items need godoc comments
- Prefer small, focused functions (< 50 lines)
- Tests use table-driven style

## Questions?

Open a [GitHub Discussion](https://github.com/justcodeit404/mcpkit/discussions) or reach out on the MCP Discord.
