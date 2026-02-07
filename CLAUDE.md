# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Julius is an LLM service fingerprinting tool that identifies what AI server software (Ollama, vLLM, LiteLLM, etc.) is running on network endpoints. It sends HTTP probes and matches response signatures to identify 17+ LLM services.

## Development Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/rules/...

# Build the binary
go build -o julius ./cmd/julius

# Validate probe YAML files
julius validate ./probes

# Test against a target
./julius probe https://target.example.com

# Verbose output for debugging
./julius probe -v https://target.example.com
```

## Architecture

### Core Flow
1. Target URLs are normalized and validated
2. Probes are loaded (embedded YAML + optional filesystem probes)
3. Scanner sends HTTP requests with response caching (`singleflight` + MD5 dedup)
4. Rules match against HTTP responses (status, body, headers)
5. Results are ranked by specificity (1-100, higher = more specific match)
6. Models are optionally extracted via JQ expressions

### Key Packages

- **`pkg/runner/`** - CLI command implementations (probe, list, validate)
- **`pkg/scanner/`** - HTTP client, response caching, model extraction
- **`pkg/rules/`** - Match rule engine with plugin-style registration
- **`pkg/probe/`** - Probe loader (embedded + filesystem YAML)
- **`pkg/types/`** - Core data structures (Probe, Request, Result)
- **`pkg/output/`** - Output formatters (table, JSON, JSONL)

### Rule System

Rules are registered via `init()` functions in `pkg/rules/rule_*.go` files:

```go
func init() {
    Register("body.contains", NewBodyContainsRule)
}
```

To add a new rule type:
1. Create `pkg/rules/rule_<name>.go` implementing the `Rule` interface
2. Register it in `init()` with a unique type name
3. Add tests in `pkg/rules/rules_test.go`

Available rule types: `status`, `body.contains`, `body.prefix`, `header.contains`, `header.prefix`, `content-type`

### Probe YAML Structure

Probes in `probes/` define service detection. Key fields:

- `name` - Unique identifier (must match filename)
- `specificity` - 1-100 ranking (100=exact match, 50=default, 1=generic fallback)
- `require` - `any` (default, first match wins) or `all` (all requests must match)
- `requests` - HTTP probes with match rules
- `models` - Optional JQ extraction config for model discovery
- `augustus` - Optional generator config for downstream tooling

## Local Files

See `local/CLAUDE.md` for documentation on local research files (gitignored, may vary per developer).
