# Contributing to Julius

Thank you for your interest in contributing to Julius, the LLM service fingerprinting tool! This guide covers adding new LLM service probes, extending Julius with new rule types, and submitting contributions.

## Table of Contents

- [Quick Start](#quick-start)
- [Adding a Probe](#adding-a-probe)
- [Probe Reference](#probe-reference)
- [Adding a Rule Type](#adding-a-rule-type)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)

## Quick Start

1. Fork and clone the repository
2. Install Go 1.24+
3. Run tests: `go test ./...`
4. Make your changes
5. Submit a pull request

## Adding a Probe

Probes are YAML files in `probes/` that define how to identify an LLM service. Each probe specifies HTTP requests to send and rules to match against responses.

### When to Add a New Probe

Add a new probe when:

- A new LLM serving platform emerges (e.g., new inference server)
- An existing service changes its API signatures
- You discover a better way to identify a service with higher confidence

### Example: Ollama Probe

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

### Probe Fields Explained

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier matching the filename (without `.yaml`) |
| `description` | Yes | Human-readable description of the LLM service |
| `category` | Yes | Service category: `self-hosted`, `proxy`, `ui`, `enterprise`, `rag` |
| `port_hint` | No | Default port for the service (helps prioritize probes) |
| `api_docs` | No | Link to the service's API documentation |
| `probes` | Yes | List of HTTP probe definitions |
| `models` | No | Configuration for extracting available model names |

### Confidence Levels

Choose confidence based on how unique the signature is:

| Level | When to Use |
|-------|-------------|
| `high` | Signature is unique to this service (e.g., specific header, unique response body) |
| `medium` | Signature is fairly specific but could match similar services |
| `low` | Signature is generic (e.g., just a 200 status code) |

**Rule of thumb**: If another LLM service could reasonably match the same rules, use `medium` or lower.

### Validating Your Probe

Before submitting, validate your probe:

```bash
# Check YAML syntax and structure
julius validate ./probes

# Test against a live instance
julius probe -v https://your-test-instance:port

# Test for false positives against other services
julius probe -v https://different-llm-service:port
```

### Common Mistakes

1. **Too generic rules**: Using only `status: 200` matches too many services
2. **Missing negation**: Not using `not: true` when needed to exclude false positives
3. **Wrong confidence**: Setting `high` confidence for generic signatures
4. **Untested probes**: Not validating against live services

## Probe Reference

### Top-Level Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `name` | Yes | - | Unique identifier, should match filename |
| `description` | Yes | - | Human-readable description |
| `category` | Yes | - | Service category |
| `port_hint` | No | - | Default port for the service |
| `api_docs` | No | - | Link to API documentation |
| `probes` | Yes | - | List of probe definitions |
| `models` | No | - | Model extraction configuration |
| `require` | No | `any` | Match mode: `any` (first match wins) or `all` (all must match) |

### Probe Definition Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `type` | No | `http` | Probe type (currently only `http` supported) |
| `path` | Yes | - | HTTP path to request |
| `method` | No | `GET` | HTTP method |
| `headers` | No | - | Request headers as key-value pairs |
| `body` | No | - | Request body (for POST/PUT) |
| `match` | Yes | - | List of match rules (all must match) |
| `confidence` | No | `medium` | Match confidence level |

### Match Rules

| Type | Fields | Description | Example |
|------|--------|-------------|---------|
| `status` | `value` | HTTP status code equals value | `value: 200` |
| `body.contains` | `value` | Response body contains string | `value: '"models":'` |
| `body.prefix` | `value` | Response body starts with string | `value: '{"object":'` |
| `header.contains` | `header`, `value` | Header contains string | `header: Content-Type`, `value: json` |
| `header.prefix` | `header`, `value` | Header starts with string | `header: Server`, `value: nginx` |
| `content-type` | `value` | Content-Type header matches | `value: application/json` |

### Rule Negation

All rules support `not: true` to negate the match:

```yaml
match:
  - type: body.contains
    value: "OpenAI"
    not: true  # Match if body does NOT contain "OpenAI"
```

### Model Extraction

The `models` section defines how to extract available model names:

```yaml
models:
  path: /api/tags        # Endpoint returning model list
  method: GET            # HTTP method
  extract: ".models[].name"  # JQ expression to extract model names
```

The `extract` field uses [JQ syntax](https://jqlang.github.io/jq/manual/) for parsing JSON responses.

## Adding a Rule Type

To add a new match rule type (e.g., `body.regex`):

### Steps

1. Create a new file in `pkg/rules/` (e.g., `rule_body_regex.go`)
2. Define a struct implementing the `Rule` interface
3. Register the rule type in `init()`
4. Add tests in `pkg/rules/rules_test.go`

### Rule Interface

```go
type Rule interface {
    Match(resp *http.Response, body []byte) bool
}
```

### Example Implementation

Use existing rules as reference. Here's the pattern:

```go
package rules

import (
    "fmt"
    "net/http"
    "regexp"
)

func init() {
    Register("body.regex", NewBodyRegexRule)
}

type BodyRegexRule struct {
    BaseRule
    Pattern *regexp.Regexp
}

func (r BodyRegexRule) Match(resp *http.Response, body []byte) bool {
    result := r.Pattern.Match(body)
    if r.Not {
        return !result
    }
    return result
}

func NewBodyRegexRule(raw *RawRule) (Rule, error) {
    val, err := toString(raw.Value)
    if err != nil {
        return nil, fmt.Errorf("body.regex %w", err)
    }
    pattern, err := regexp.Compile(val)
    if err != nil {
        return nil, fmt.Errorf("body.regex invalid pattern: %w", err)
    }
    return &BodyRegexRule{
        BaseRule: BaseRule{Type: raw.Type, Not: raw.Not},
        Pattern:  pattern,
    }, nil
}
```

### Key Points

- Embed `BaseRule` for common fields (`Type`, `Not`)
- `Match()` returns true if the rule matches, respecting `Not` for negation
- Constructor parses `RawRule` from YAML and returns typed rule
- `Register()` in `init()` makes the rule available by name

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./pkg/rules/...

# Run tests with coverage
go test -cover ./...
```

### Validating Probes

```bash
# Validate all probes in default directory
julius validate ./probes

# Validate probes in custom directory
julius validate /path/to/probes
```

### Writing Tests

We use [testify](https://github.com/stretchr/testify) for assertions:

- Use `assert` for non-fatal checks (test continues on failure)
- Use `require` for fatal checks (test stops on failure)

**Example test for a new rule:**

```go
func TestBodyRegexRule(t *testing.T) {
    tests := []struct {
        name     string
        rule     RawRule
        body     string
        expected bool
    }{
        {"matches", RawRule{Type: "body.regex", Value: "model.*v1"}, `{"model": "v1.0"}`, true},
        {"no match", RawRule{Type: "body.regex", Value: "^foo"}, "hello world", false},
        {"negated", RawRule{Type: "body.regex", Value: "error", Not: true}, "success", true},
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

### Testing New Probes

For new probes, test manually:

1. **Positive test**: Run against a live instance of the service
2. **False positive test**: Run against other LLM services to ensure no incorrect matches
3. **Edge cases**: Test with authentication enabled, different configurations, etc.

## Code Style

### Go Code

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and testable

### YAML Probes

- Use lowercase names with hyphens (e.g., `huggingface-tgi`)
- Include `api_docs` link when available
- Order probes by confidence (highest first)
- Use descriptive string values in match rules

## Submitting Changes

### External Contributors

1. Fork the repository
2. Create a branch for your changes (`git checkout -b feature/my-feature`)
3. Make your changes
4. Ensure tests pass: `go test ./...`
5. Ensure probes validate: `julius validate ./probes`
6. Commit with clear messages
7. Push to your fork
8. Submit a pull request

### Pull Request Guidelines

- Provide a clear description of changes
- Reference any related issues
- Include test results for new probes
- Update documentation as needed

### Praetorian Team

1. Create a branch for your changes
2. Make changes and test
3. Submit a pull request for review

## Questions?

- Open a [GitHub issue](https://github.com/praetorian-inc/julius/issues) for bugs or feature requests
- Check existing issues before creating new ones
- Use issue templates when available
