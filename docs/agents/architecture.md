# Julius architecture and probe reference

Relocated from `AGENTS.md` (2026-09-02). This material is derivable from the
tree — `README.md` § "How It Works" / "Architecture" and `CONTRIBUTING.md`
§ "Probe Reference" / "Adding a Rule Type" cover the same ground — so it is
kept here for reference rather than loaded at every session start.

## Project Overview

Julius is an LLM service fingerprinting tool that identifies what AI server
software (Ollama, vLLM, LiteLLM, etc.) is running on network endpoints. It sends
HTTP probes and matches response signatures to identify LLM services. As of
2026-09-02 there are 64 probe files under `probes/`: one per service plus the
generic fallback `probes/openai-compatible.yaml` (specificity 1). The earlier
"17+" figure in `AGENTS.md` was stale.

## Architecture

### Core Flow

1. Target URLs are normalized and validated
2. Probes are loaded (embedded YAML + optional filesystem probes) and sorted so
   probes whose `port_hint` matches the target port run first
3. Scanner sends HTTP requests with response caching (`singleflight` + MD5
   dedup, see `pkg/scanner/cache.go`)
4. Rules match against HTTP responses (status, body, headers)
5. Results are ranked by specificity (1-100, higher = more specific match)
6. Models are optionally extracted via JQ expressions (gojq)

### Key Packages

- **`pkg/runner/`** - CLI command implementations (probe, list, validate)
- **`pkg/scanner/`** - HTTP client, response caching, model extraction, and the
  probe fixture harness
- **`pkg/rules/`** - Match rule engine with plugin-style registration
- **`pkg/probe/`** - Probe loader (embedded + filesystem YAML)
- **`pkg/types/`** - Core data structures (Probe, Request, Result)
- **`pkg/output/`** - Output formatters (table, JSON, JSONL)
- **`cmd/julius/`** - CLI entrypoint

### Rule System

Rules are registered via `init()` functions in `pkg/rules/rule_*.go` files
(for example `pkg/rules/rule_body_contains.go`):

```go
func init() {
    Register("body.contains", NewBodyContainsRule)
}
```

To add a new rule type:

1. Create `pkg/rules/rule_<name>.go` implementing the `Rule` interface
2. Register it in `init()` with a unique type name
3. Add tests in `pkg/rules/rules_test.go`

Available rule types (verified against the `Register` calls in `pkg/rules/`):
`status`, `body.contains`, `body.prefix`, `header.contains`, `header.prefix`,
`content-type`. All rules support negation with `not: true`.

### Probe YAML Structure

Probes in `probes/` define service detection. Key fields:

- `name` - Unique identifier. By convention it matches the filename without
  `.yaml`; `julius validate` does not enforce this, and the fixture harness
  keys on `name`, not the filename.
- `specificity` - 1-100 ranking (100=exact match, 50=default, 1=generic
  fallback). `0` or omitted is treated as 50; the named constants are in
  `pkg/types/probe.go`.
- `require` - `any` (default, first match wins) or `all` (all requests must
  match)
- `requests` - HTTP probes with match rules; each request needs at least one
  match rule. An empty `path` is allowed and sends the request to the target
  URL exactly as supplied.
- `models` - Optional JQ extraction config for model discovery
- `augustus` - Optional generator config for downstream tooling

`CONTRIBUTING.md` § "Probe Reference" documents the remaining fields
(`description`, `category`, `port_hint`, `api_docs`).
