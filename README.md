<img width="2752" height="1536" alt="julius" src="https://github.com/user-attachments/assets/aca4e0d4-313e-4428-8856-06f07599d283" />

# Julius

**Simple LLM service identification.** Translate IP:Port to Ollama, vLLM, LiteLLM, or 15+ other AI services in seconds.

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/github/license/praetorian-inc/julius)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/praetorian-inc/julius/ci.yml?branch=main)](https://github.com/praetorian-inc/julius/actions)

## The Problem

You've found an open port during a security assessment. Is it Ollama? vLLM? LiteLLM? A Hugging Face endpoint? Something else entirely?

**Julius answers that question in seconds.**

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

## Supported LLM Services

Julius identifies 17 LLM platforms across self-hosted, enterprise, and UI categories:

| Service | Category |
|---------|----------|
| [Ollama](https://ollama.ai) | Self-hosted |
| [vLLM](https://github.com/vllm-project/vllm) | Self-hosted |
| [LiteLLM](https://github.com/BerriAI/litellm) | Proxy/Gateway |
| [LocalAI](https://localai.io) | Self-hosted |
| [LM Studio](https://lmstudio.ai) | Desktop |
| [Hugging Face TGI](https://huggingface.co/docs/text-generation-inference) | Self-hosted |
| [llama.cpp](https://github.com/ggerganov/llama.cpp) | Self-hosted |
| [Open WebUI](https://github.com/open-webui/open-webui) | UI/Frontend |
| [AnythingLLM](https://anythingllm.com) | RAG Platform |
| [NVIDIA NIM](https://developer.nvidia.com/nim) | Enterprise |
| [Kong AI Gateway](https://konghq.com) | Gateway |
| [LibreChat](https://librechat.ai) | Chat UI |
| [Gradio](https://gradio.app) | Web UI |
| [SillyTavern](https://sillytavernai.com) | Chat UI |
| [BetterChatGPT](https://github.com/ztjhz/BetterChatGPT) | Chat UI |
| [Salesforce Einstein](https://www.salesforce.com/einstein/) | Enterprise |
| OpenAI-compatible | Generic |

## Usage

### Single Target

```bash
julius probe https://target.example.com
```

### Multiple Targets

```bash
julius probe https://target1.example.com https://target2.example.com
julius probe -f targets.txt
cat targets.txt | julius probe -
```

### Output Formats

```bash
# Table (default)
julius probe https://target.example.com

# JSON
julius probe -o json https://target.example.com

# JSONL (for piping)
julius probe -o jsonl https://target.example.com
```

### List Available Probes

```bash
julius list
```

## How It Works

Julius sends HTTP probes to targets and matches responses against service signatures:

```
Target → Probes → Scanner → Rules → Match
```

Each probe defines:
- **Endpoint** - Path to request (e.g., `/api/tags` for Ollama)
- **Match rules** - Status codes, headers, body patterns
- **Confidence** - How certain the identification is

When all rules match, Julius reports the service with its confidence level.

## Architecture

```
cmd/julius/     CLI entrypoint
pkg/
  runner/       Command execution
  scanner/      HTTP client and response matching
  rules/        Match rules (status, body.contains, header.prefix, etc.)
  output/       Table and JSON formatters
  probe/        Probe loading and sorting
probes/         YAML probe definitions (add your own here)
```

## Adding Custom Probes

Create a YAML file in `probes/` to detect new services:

```yaml
name: my-llm-service
description: My custom LLM service
category: self-hosted
port_hint: 8080

probes:
  - path: /health
    match:
      - type: status
        value: 200
      - type: body.contains
        value: '"service":"my-llm"'
    confidence: high
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full probe specification.

## FAQ

### What is LLM service fingerprinting?

LLM service fingerprinting identifies what LLM server software (Ollama, vLLM, LiteLLM, etc.) is running on a network endpoint. This differs from model fingerprinting, which identifies which AI model generated a piece of text.

### Is Julius safe for penetration testing?

Yes. Julius only performs standard HTTP requests - the same as a web browser. It does not exploit vulnerabilities or modify data. Always ensure you have authorization before scanning targets.

### How do I add support for a new LLM service?

Create a YAML probe file in the `probes/` directory. See [CONTRIBUTING.md](CONTRIBUTING.md) for the full specification and examples.

### Why "Julius"?

Named after Julius Caesar - the original fingerprinter of Roman politics.

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:

- Adding new LLM service probes
- Creating new match rule types
- Testing guidelines

## License

[MIT](LICENSE) - Praetorian, Inc.
