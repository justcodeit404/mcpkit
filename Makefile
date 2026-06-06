SHELL := /bin/bash
BINARY := mcpkit
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
  -X github.com/justcodeit404/mcpkit/internal/cli.Version=$(VERSION) \
  -X github.com/justcodeit404/mcpkit/internal/cli.Commit=$(COMMIT) \
  -X github.com/justcodeit404/mcpkit/internal/cli.BuildDate=$(BUILD_DATE) \
  -X github.com/justcodeit404/mcpkit/internal/mcp.ClientVersion=$(VERSION)

.PHONY: build install test lint format clean release-snapshot

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/mcpkit

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/mcpkit

test:
	go test -race -v ./...

lint:
	golangci-lint run

format:
	gofmt -s -w .
	goimports -w .

clean:
	rm -rf bin/ dist/

release-snapshot:
	goreleaser release --snapshot --clean
