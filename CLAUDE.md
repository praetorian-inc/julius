# Julius - LLM Service Fingerprinting Tool

## Documentation

- [README.md](README.md) - Quick start, usage, architecture overview
- [CONTRIBUTING.md](CONTRIBUTING.md) - Adding probes, rule types, testing

## Development Workflow

```bash
# Run tests
go test ./...

# Validate probe YAML files
julius validate ./probes

# Build
go build -o julius ./cmd/julius

# Test against a target
./julius probe https://target.example.com
```

## Key Directories

- `probes/` - YAML probe definitions for LLM services
- `pkg/rules/` - Match rule implementations
- `pkg/scanner/` - HTTP client and response matching
- `pkg/runner/` - CLI commands

## Local Files

See `local/CLAUDE.md` for documentation on local research files (gitignored, may vary per developer).
