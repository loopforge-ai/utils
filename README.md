# loopforge-ai/utils

Reusable Go utility packages for the [loopforge-ai](https://github.com/loopforge-ai) organization. Standard library only — zero external dependencies.

## Packages

| Package | Description |
|---------|-------------|
| `assert` | Test assertion helpers (`assert.That`) |
| `env` | Generic environment variable parsing (`env.Get[T]`) |
| `fs` | Filesystem utilities (atomic write, path safety) |
| `html` | HTTP server infrastructure (middleware, template renderer, constants) |
| `llm` | LLM `Completer` interface with Claude CLI and OpenAI HTTP adapters |
| `mcp` | Minimal Model Context Protocol (MCP) server over JSON-RPC stdio |
| `yaml` | YAML marshal/unmarshal (no external deps) |

## Install

```bash
go get github.com/loopforge-ai/utils
```

## Build & Test

```bash
go test ./...                          # unit tests only
go test ./... -tags=integration        # include integration tests
golangci-lint run ./...                # lint (must pass with 0 issues)
```
