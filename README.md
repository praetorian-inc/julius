# Julius

LLM service fingerprinting tool.

## Quick Start

```bash
go install github.com/praetorian-inc/julius/cmd/julius@latest
julius probe https://target.example.com
```

```
+----------------------------+----------+------------+---------------+-------------+
|           TARGET           | SERVICE  | CONFIDENCE | MATCHED PROBE |  CATEGORY   |
+----------------------------+----------+------------+---------------+-------------+
| https://target.example.com | ollama   | high       | /api/tags     | self-hosted |
+----------------------------+----------+------------+---------------+-------------+
```

## Usage

Probe a single target:

```bash
julius probe https://target.example.com
```

Probe multiple targets:

```bash
julius probe https://target1.example.com https://target2.example.com
julius probe -f targets.txt
cat targets.txt | julius probe -
```

Output as JSON:

```bash
julius probe -o json https://target.example.com
```

List available probes:

```bash
julius list
```

## Architecture

Julius sends HTTP probes to targets and matches responses against rules to identify LLM services.

```
Target → Probes → Scanner → Rules → Output
```

Directory structure:

```
cmd/julius/     CLI entrypoint
pkg/
  runner/       Command execution
  scanner/      HTTP client and response matching
  rules/        Match rules (status, body.contains, header.prefix, etc.)
  output/       Table and JSON formatters
  probe/        Probe loading and sorting
probes/         YAML probe definitions
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on adding probes and extending julius.
