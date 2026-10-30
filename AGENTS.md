# AGENTS.md

This is the canonical agent instruction file for this repository, read by every coding agent that works here. `CLAUDE.md` is a one-line pointer to this file, and `.gemini/settings.json` lists it as a context file. Architecture notes and the probe/rule reference that used to live here are in `docs/agents/architecture.md`; `README.md` and `CONTRIBUTING.md` carry the user-facing versions.

## Development Commands

```bash
# Run all tests (also: make test)
go test ./...

# Run tests for a specific package
go test ./pkg/rules/...

# Build the binary (make build writes to bin/julius instead)
go build -o julius ./cmd/julius

# Validate probe YAML files (schema-only; exactly one directory argument)
julius validate ./probes

# Run the per-probe behavioral fixture harness (subset of `go test ./...`)
go test ./pkg/scanner/...

# Test against a target
./julius probe https://target.example.com

# Verbose output for debugging (-q suppresses non-match output)
./julius probe -v https://target.example.com
```

`make help` lists the remaining `Makefile` targets (`make lint` runs golangci-lint, the same linter CI runs).

## Probe Fixtures

Every shipped probe must have a companion behavioral fixture in
`pkg/scanner/testdata/fixtures/<probe>.yaml` with **at least one positive and
one negative case**. The harness (`pkg/scanner/fixture_test.go`) serves canned
HTTP responses and runs the real probe through the real scanner to assert match
behavior; `TestEveryProbeHasFixture` fails the build if any probe lacks a
fixture or lacks either case. When adding a probe, add its fixture — see
`pkg/scanner/testdata/fixtures/README.md` for the format. `CONTRIBUTING.md`
and the pull-request template describe only manual live-service testing and do
not mention this requirement, so do not treat their checklists as complete.

The fixture is keyed by the probe's `name` field, not its filename. `julius
validate` does not check that `name` matches the filename; that remains a
convention (see `docs/agents/architecture.md`).
