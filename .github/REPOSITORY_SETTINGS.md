# Recommended GitHub Repository Settings

This document contains the recommended GitHub repository settings for optimal SEO and discoverability.

## About Section

Configure in: Repository Settings > General > About

**Recommended Description (12 words):**

```
LLM service fingerprinting tool - detect Ollama, vLLM, LiteLLM, and 17+ AI servers
```

**Alternative (shorter, 8 words):**

```
Detect Ollama, vLLM, LiteLLM and 17+ LLM services
```

## Topics (12 recommended)

Configure in: Repository Settings > General > Topics

Add these topics in order of importance:

### Primary (Technology + Purpose)
1. `llm-fingerprinting` - Unique niche keyword
2. `service-detection` - Core functionality
3. `security-tools` - Target audience

### Service-Specific (High Search Volume)
4. `ollama` - Most popular self-hosted LLM
5. `vllm` - Popular inference server
6. `litellm` - Popular LLM proxy
7. `huggingface` - Major AI platform

### Domain Keywords
8. `penetration-testing` - Target use case
9. `attack-surface` - Security context
10. `reconnaissance` - Security workflow

### Technical
11. `cli-tool` - Tool type
12. `golang` - Implementation language (optional - high competition)

## Website URL

Set to: `https://www.praetorian.com/` (or project-specific docs if available)

## Social Preview Image

The current banner image is good. Ensure it displays well at:
- 1280x640 pixels (recommended)
- Contains the project name and key value proposition

## Features to Enable

- [x] Issues
- [x] Discussions (optional - for community Q&A)
- [x] Projects (optional)
- [x] Wiki (optional - for extended documentation)
- [x] Sponsorship (via FUNDING.yml)

## Branch Protection

Recommended for `main` branch:
- Require pull request reviews
- Require status checks to pass (CI)
- Do not allow force pushes

## Actions

Ensure CI workflow runs on:
- Push to main
- Pull requests
- Release tags
