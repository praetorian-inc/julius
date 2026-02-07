# Changelog

All notable changes to Julius will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[Unreleased]: https://github.com/praetorian-inc/julius/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/praetorian-inc/julius/releases/tag/v0.1.0
