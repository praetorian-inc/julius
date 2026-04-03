# Changelog

All notable changes to Julius will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2026-04-02

### Fixed (Probes) — False Positive Audit

Comprehensive audit of all 63 probes to eliminate false positives when scanning arbitrary
internet infrastructure. 40 probes modified across 43 files (307 lines added, 493 removed).

**Added `require: all` (AND gate)** to 22 probes that had generic OR blocks which could
independently match unrelated sites:
- aws-bedrock, azure-openai, baseten-truss, bentoml, bifrost, databricks-model-serving,
  flowise, gradio, h2ogpt, helicone, koboldcpp, langflow, librechat, nvidia-nim, ollama,
  omniroute, openwebui, portkey-ai-gateway, powerinfer, quivr, ragflow, sillytavern

**Rewrote probes** that were fundamentally too generic or broken:
- **deepspeed-mii**: Old probe matched 350k+ Werkzeug servers; new probe uses hardcoded
  `owned_by: "deepspeed-mii"` from source code
- **modal**: Old probe (specificity 30) matched any OpenAI-compat server with CORS; new
  probe uses Modal's `"modal-http: invalid function call"` error on invalid paths (specificity 95)
- **petals**: Old probe targeted wrong service (health monitor, not chat server) with
  generic unnamed checks; new probe uses hardcoded template strings from source
- **sillytavern**: Old probe matched any page mentioning "SillyTavern"; new probe uses
  `/manifest.json` + `/version` endpoint AND gate
- **librechat**: Old probe matched any page mentioning "LibreChat" or listing "openAI";
  new probe AND's HTML title with `/api/config` containing `librechat.ai`
- **replicate**: Old probe required status 200 but Replicate returns 401 without auth —
  probe never matched; new probe uses RFC 7807 error format on Replicate-specific paths
- **together-ai**: Old probe checked for `"api.together.xyz"` in wrong response body —
  probe never matched; new probe uses correct endpoint

**Added stronger fingerprints** from source code research:
- **sglang, tensorrt-llm**: Added `owned_by` field checks from verified source code
- **groq**: Added `x-groq-region` response header (verified on Shodan, survives proxying)
- **lm-studio**: Added `/lmstudio-greeting` self-identification endpoint from source
- **dify**: Restored `<title>Dify</title>` check (Shodan confirms SSR in all versions,
  contrary to PR #45 comment); added subpath deployment detection

**Tightened existing probes** by adding service-specific field checks:
- **huggingface-tgi**: Added `model_dtype`, `max_batch_total_tokens`, `docker_label`
- **betterchatgpt**: Added unique meta description string
- **ray-serve**: Consolidated to single block with 4 Ray-specific field checks
- **localai**: Merged redundant blocks into single tighter block

**Removed dangerous generic blocks** that could match unrelated sites:
- cloudflare-ai-gateway block 2, tensorzero block 2, litellm blocks 1/3,
  triton-inference-server block 3, portkey-ai-gateway block 3, petals blocks 1-4

## [0.2.0] - 2026-03-24

### Breaking Changes

- **`scanner.NewScanner()` signature changed**: Now requires two additional parameters:
  - `maxResponseSize int64` - Maximum response body size in bytes
  - `tlsConfig *tls.Config` - TLS configuration (can be nil for defaults)

  **Migration:**
  ```go
  // Before
  s := scanner.NewScanner(timeout, concurrency)

  // After
  s := scanner.NewScanner(timeout, concurrency, scanner.DefaultMaxResponseSize, nil)
  ```

### Added (Probes)

- **30 new LLM service probes** bringing the total from 33 to 63:
  - **Self-hosted**: SGLang, BentoML, Baseten Truss, DeepSpeed-MII, MLC LLM, Petals, PowerInfer, Ray Serve, TensorRT-LLM, Triton Inference Server
  - **Cloud-managed**: AWS Bedrock, Azure OpenAI, Cloudflare AI Gateway, Databricks Model Serving, Fireworks AI, Google Vertex AI, Groq, Modal, Replicate, Together AI
  - **Gateways**: Bifrost, Helicone, OmniRoute, Portkey AI Gateway, TensorZero
  - **RAG/orchestration**: h2oGPT, Langflow, PrivateGPT, Quivr, RAGFlow

### Fixed (Probes)

- **SGLang**: Fixed `/server_info` match rules for cross-version compatibility (older versions lack `radix_eviction_policy`)
- **Ollama**: Fixed `/api/tags` false positive on Ollama-compatible servers (SGLang, etc.) by requiring `"families"` field
- **Dify**: Fixed detection for newer versions that render title via JavaScript; now uses `data-public-edition` + `/console/api` body data attributes
- **Flowise**: Updated title match from outdated string to `flowiseai.com` (present in all versions)
- **Bifrost**: Removed overly generic `/api/version` block that matched KoboldCpp, Open WebUI, and other services
- **DeepSpeed-MII**: Removed `/health` block entirely (`{"status":"ok"}` with uvicorn is too common across FastAPI apps)
- **Groq**: Removed generic `/openai/v1/models` 200-status block that matched KoboldCpp and other OpenAI-compat servers
- **AWS Bedrock, Cloudflare AI Gateway, Fireworks AI, Modal, OmniRoute**: Fixed `header.contains` rules that used header names in the `value` field without a `header` field, causing matches to always fail on HTTP/2

### Added

- `--max-response-size` flag to limit HTTP response body size (default 10MB)
  - Protects against memory exhaustion from large responses
- `--insecure` flag to skip TLS certificate verification
- `--ca-cert` flag to specify custom CA certificate file for enterprise PKI environments
- `loadProbes()` helper function to centralize probe loading logic

### Fixed

- Removed dead `matchedTargets` variable in `pkg/runner/probe.go`
- Eliminated duplicate probe loading code between `probe.go` and `list.go`

### Security

- Added response body size limiting via `io.LimitReader` to prevent OOM from malicious servers
- Added TLS configuration options for secure scanning in enterprise environments

## [0.1.2] - 2026-02-12

### Added

- New LLM service probe (33 total):
  - RAG/Orchestration: OpenClaw (formerly Clawdbot/Moltbot) - AI agent gateway and control plane on port 18789

## [0.1.1] - 2025-02-09

### Added

- 15 new LLM service probes (32 total):
  - Self-hosted: Aphrodite Engine, FastChat Controller, GPT4All, Jan, KoboldCpp, TabbyAPI, Text Generation WebUI
  - Gateway: Envoy AI Gateway
  - RAG/Orchestration: AstrBot, Dify, Flowise, HuggingFace Chat UI, LobeHub, NextChat, Onyx

## [0.1.0] - 2025-01-24

### Added

- Initial release of Julius LLM service fingerprinting tool
- Support for 19 LLM platforms:
  - Self-hosted: Ollama, vLLM, LocalAI, llama.cpp, Hugging Face TGI, LM Studio
  - Proxy/Gateway: LiteLLM, Kong AI Gateway
  - UI/Frontend: Open WebUI, LibreChat, Gradio, SillyTavern, BetterChatGPT
  - Enterprise: NVIDIA NIM, Salesforce Einstein, AnythingLLM
  - Generic: OpenAI-compatible
- HTTP-based probe system with YAML configuration
- Match rules: status, body.contains, body.prefix, header.contains, header.prefix
- Rule negation support with `not: true`
- Confidence scoring (high/medium/low)
- Model discovery and extraction via JQ expressions
- Augustus code generation integration
- Multiple input methods: single target, file, stdin
- Multiple output formats: table, JSON, JSONL
- Concurrent scanning with configurable concurrency
- Response caching with MD5 deduplication
- Port-based probe prioritization
- Probe validation command (`julius validate`)
- Probe listing command (`julius list`)
- Embedded probes compiled into binary
- Comprehensive test suite

### Technical

- Go 1.25.3 requirement
- Cobra CLI framework
- errgroup for bounded concurrency
- singleflight for request deduplication
- go-yaml for probe parsing
- gojq for model extraction
- tablewriter for formatted output

[Unreleased]: https://github.com/praetorian-inc/julius/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/praetorian-inc/julius/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/praetorian-inc/julius/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/praetorian-inc/julius/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/praetorian-inc/julius/releases/tag/v0.1.0
