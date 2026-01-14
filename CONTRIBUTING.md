# Contributing to Julius

This guide covers adding new probes and extending julius with new rule types.

## Table of Contents

- [Adding a Probe](#adding-a-probe)
- [Probe Reference](#probe-reference)
- [Adding a Rule Type](#adding-a-rule-type)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)

## Adding a Probe

Probes are YAML files in `probes/` that define how to identify an LLM service.

### Example: Ollama

```yaml
# probes/ollama.yaml
name: ollama
description: Ollama local LLM server
category: self-hosted
port_hint: 11434
api_docs: https://github.com/ollama/ollama/blob/main/docs/api.md

probes:
  - type: http
    path: /api/tags
    method: GET
    match:
      - type: status
        value: 200
      - type: body.contains
        value: '"models":'
    confidence: high

  - type: http
    path: /
    method: GET
    match:
      - type: status
        value: 200
      - type: body.contains
        value: "Ollama is running"
    confidence: medium

models:
  path: /api/tags
  method: GET
  extract: ".models[].name"
```

**Key points:**

- `name` - Unique identifier, matches filename
- `port_hint` - Julius tries probes matching the target port first
- `probes` - List of HTTP requests to try, ordered by confidence (high first)
- `match` - All rules must match for the probe to succeed
- `confidence` - How certain the match is (`high`, `medium`, `low`)
- `models` - Optional endpoint to extract available model names (uses jq syntax)

### Validating Your Probe

1. Check YAML syntax:

```bash
julius validate ./probes
```

2. Test against a live instance of the service to confirm it matches.

3. Test against other LLM services (or non-LLM HTTP services) to ensure no false positives. Your match rules should be specific enough to avoid matching unrelated services.

## Probe Reference

### Top-Level Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier, should match filename |
| `description` | Yes | Human-readable description |
| `category` | Yes | Service category (e.g., `self-hosted`, `rag-orchestration`) |
| `port_hint` | No | Default port for the service |
| `api_docs` | No | Link to API documentation |
| `probes` | Yes | List of probe definitions |
| `models` | No | Model extraction configuration |

### Probe Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `type` | No | `http` | Probe type |
| `path` | Yes | - | HTTP path to request |
| `method` | No | `GET` | HTTP method |
| `headers` | No | - | Request headers |
| `body` | No | - | Request body |
| `match` | Yes | - | List of rules (all must match) |
| `confidence` | No | `medium` | Match confidence (`high`, `medium`, `low`) |

### Match Rules

| Type | Fields | Description |
|------|--------|-------------|
| `status` | `value` | HTTP status code equals value |
| `body.contains` | `value` | Response body contains string |
| `body.prefix` | `value` | Response body starts with string |
| `header.contains` | `header`, `value` | Header contains string |
| `header.prefix` | `header`, `value` | Header starts with string |

All rules support `not: true` to negate the match.

## Adding a Rule Type

To add a new match rule type (e.g., `body.regex`):

### Steps

1. Create a new file in `pkg/rules/` (e.g., `rule_body_regex.go`)
2. Define a struct implementing the `Rule` interface
3. Register the rule type in `init()`
4. Add tests in `pkg/rules/rules_test.go`

### Example: Existing Rule

Use `rule_body_contains.go` as a reference:

```go
package rules

import (
	"fmt"
	"net/http"
	"strings"
)

func init() {
	Register("body.contains", NewBodyContainsRule)
}

type BodyContainsRule struct {
	BaseRule
	Value string
}

func (r BodyContainsRule) Match(resp *http.Response, body []byte) bool {
	result := strings.Contains(string(body), r.Value)
	if r.Not {
		return !result
	}
	return result
}

func NewBodyContainsRule(raw *RawRule) (Rule, error) {
	val, err := toString(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("body.contains %w", err)
	}
	return &BodyContainsRule{
		BaseRule: BaseRule{Type: raw.Type, Not: raw.Not},
		Value:    val,
	}, nil
}
```

**Key points:**

- Embed `BaseRule` for common fields (`Type`, `Not`)
- `Match()` returns true if the rule matches, respecting `Not` for negation
- Constructor parses `RawRule` from YAML and returns typed rule
- `Register()` in `init()` makes the rule available by name

## Testing

### Running Tests

```bash
go test ./...
```

### Validating Probes

```bash
julius validate ./probes
```

### Writing Tests

We use [testify](https://github.com/stretchr/testify) for assertions. Prefer `assert` for non-fatal checks and `require` for fatal checks that should stop the test.

**For new rules**, add test cases to `pkg/rules/rules_test.go`:

```go
func TestBodyContainsRule(t *testing.T) {
	tests := []struct {
		name     string
		rule     RawRule
		body     string
		expected bool
	}{
		{"matches", RawRule{Type: "body.contains", Value: "hello"}, "hello world", true},
		{"no match", RawRule{Type: "body.contains", Value: "foo"}, "hello world", false},
		{"negated match", RawRule{Type: "body.contains", Value: "foo", Not: true}, "hello world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := tt.rule.ToRule()
			require.NoError(t, err)

			resp := &http.Response{StatusCode: 200, Header: http.Header{}}
			result := rule.Match(resp, []byte(tt.body))
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

**For new probes**, test manually against live services as described in [Adding a Probe](#adding-a-probe).

## Submitting Changes

### External Contributors

1. Fork the repository
2. Create a branch for your changes
3. Ensure tests pass: `go test ./...`
4. Submit a pull request

### Praetorian Team

1. Create a branch for your changes
2. Ensure tests pass: `go test ./...`
3. Submit a pull request
