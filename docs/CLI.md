<!-- Generated from the live cobra command tree by 'make cli-docs'. Do not edit by hand. -->

# julius CLI reference

Every command, alias and flag below is derived from the cobra command tree, not from prose.
Schema version 1, surface hash `sha256:13a7761739dbaa91d852750c58caf6348442ab056715c755e97a60dfafbb1cfe`.

Regenerate with `make cli-docs` after adding, removing or renaming a command or a flag.

## Command index

| Command | Aliases | Description |
| --- | --- | --- |
| [`julius`](#julius) | *(none)* | Julius - LLM Service Fingerprinting Tool |
| [`julius list`](#julius-list) | *(none)* | List all available probe definitions |
| [`julius probe`](#julius-probe) | *(none)* | Probe targets to identify LLM services |
| [`julius validate`](#julius-validate) | *(none)* | Validate probe definition files |

## `julius`

Julius - LLM Service Fingerprinting Tool

- Usage: `julius`
- Aliases: *(none)*
- Requires a subcommand

### Flags

| Flag | Short | Type | Default | Description |
| --- | --- | --- | --- | --- |
| `--banner` |  | bool | `true` | Show ASCII banner |
| `--ca-cert` |  | string |  | Path to custom CA certificate file |
| `--concurrency` | `-c` | int | `10` | Maximum concurrent probe requests per target |
| `--insecure` |  | bool | `false` | Skip TLS certificate verification |
| `--max-response-size` |  | int64 | `10485760` | Maximum response body size in bytes (default 10MB) |
| `--no-color` |  | bool | `false` | Disable color output |
| `--output` | `-o` | string | `table` | Output format (table, json, jsonl) |
| `--probes-dir` | `-p` | string |  | Override probe definitions directory |
| `--proxy` |  | string |  | Proxy URL (e.g. socks5://127.0.0.1:1080, http://127.0.0.1:8080) |
| `--proxy-auth` |  | string |  | Proxy authentication (username:password) |
| `--quiet` | `-q` | bool | `false` | Suppress non-match output |
| `--timeout` | `-t` | int | `5` | HTTP timeout in seconds |
| `--verbose` | `-v` | bool | `false` | Verbose output |

## `julius list`

List all available probe definitions

- Usage: `julius list`
- Aliases: *(none)*

### Inherited flags

| Flag | Short | Type | Default | Description |
| --- | --- | --- | --- | --- |
| `--banner` |  | bool | `true` | Show ASCII banner |
| `--ca-cert` |  | string |  | Path to custom CA certificate file |
| `--concurrency` | `-c` | int | `10` | Maximum concurrent probe requests per target |
| `--insecure` |  | bool | `false` | Skip TLS certificate verification |
| `--max-response-size` |  | int64 | `10485760` | Maximum response body size in bytes (default 10MB) |
| `--no-color` |  | bool | `false` | Disable color output |
| `--output` | `-o` | string | `table` | Output format (table, json, jsonl) |
| `--probes-dir` | `-p` | string |  | Override probe definitions directory |
| `--proxy` |  | string |  | Proxy URL (e.g. socks5://127.0.0.1:1080, http://127.0.0.1:8080) |
| `--proxy-auth` |  | string |  | Proxy authentication (username:password) |
| `--quiet` | `-q` | bool | `false` | Suppress non-match output |
| `--timeout` | `-t` | int | `5` | HTTP timeout in seconds |
| `--verbose` | `-v` | bool | `false` | Verbose output |

## `julius probe`

Probe targets to identify LLM services

- Usage: `julius probe [targets...]`
- Aliases: *(none)*

### Flags

| Flag | Short | Type | Default | Description |
| --- | --- | --- | --- | --- |
| `--augustus` |  | bool | `false` | Include Augustus generator configs in output |
| `--base-paths` |  | string |  | Comma-separated path prefixes to prepend to probe paths (e.g., /api,/proxy) |
| `--file` | `-f` | string |  | Read targets from file |
| `--header` | `-H` | stringArray | `[]` | Custom HTTP header (e.g., "Authorization: Bearer token"). Can be specified multiple times |

### Inherited flags

| Flag | Short | Type | Default | Description |
| --- | --- | --- | --- | --- |
| `--banner` |  | bool | `true` | Show ASCII banner |
| `--ca-cert` |  | string |  | Path to custom CA certificate file |
| `--concurrency` | `-c` | int | `10` | Maximum concurrent probe requests per target |
| `--insecure` |  | bool | `false` | Skip TLS certificate verification |
| `--max-response-size` |  | int64 | `10485760` | Maximum response body size in bytes (default 10MB) |
| `--no-color` |  | bool | `false` | Disable color output |
| `--output` | `-o` | string | `table` | Output format (table, json, jsonl) |
| `--probes-dir` | `-p` | string |  | Override probe definitions directory |
| `--proxy` |  | string |  | Proxy URL (e.g. socks5://127.0.0.1:1080, http://127.0.0.1:8080) |
| `--proxy-auth` |  | string |  | Proxy authentication (username:password) |
| `--quiet` | `-q` | bool | `false` | Suppress non-match output |
| `--timeout` | `-t` | int | `5` | HTTP timeout in seconds |
| `--verbose` | `-v` | bool | `false` | Verbose output |

## `julius validate`

Validate probe definition files

- Usage: `julius validate [directory]`
- Aliases: *(none)*

### Inherited flags

| Flag | Short | Type | Default | Description |
| --- | --- | --- | --- | --- |
| `--banner` |  | bool | `true` | Show ASCII banner |
| `--ca-cert` |  | string |  | Path to custom CA certificate file |
| `--concurrency` | `-c` | int | `10` | Maximum concurrent probe requests per target |
| `--insecure` |  | bool | `false` | Skip TLS certificate verification |
| `--max-response-size` |  | int64 | `10485760` | Maximum response body size in bytes (default 10MB) |
| `--no-color` |  | bool | `false` | Disable color output |
| `--output` | `-o` | string | `table` | Output format (table, json, jsonl) |
| `--probes-dir` | `-p` | string |  | Override probe definitions directory |
| `--proxy` |  | string |  | Proxy URL (e.g. socks5://127.0.0.1:1080, http://127.0.0.1:8080) |
| `--proxy-auth` |  | string |  | Proxy authentication (username:password) |
| `--quiet` | `-q` | bool | `false` | Suppress non-match output |
| `--timeout` | `-t` | int | `5` | HTTP timeout in seconds |
| `--verbose` | `-v` | bool | `false` | Verbose output |
