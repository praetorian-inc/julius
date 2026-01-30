# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in Julius, please report it responsibly:

1. **Do NOT** open a public GitHub issue
2. Email security@praetorian.com with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact assessment
   - Any suggested fixes (optional)

We will respond within 48 hours and work with you to understand and address the issue.

## Security Considerations for Julius

### What Julius Does

Julius is an HTTP-based LLM service fingerprinting tool. It performs passive identification by:

- Sending standard HTTP GET/POST requests to target endpoints
- Analyzing HTTP response status codes, headers, and body content
- Matching response patterns against known LLM service signatures
- Reporting identified services with confidence levels

### What Julius Does NOT Do

Julius is designed as a **passive identification tool**. It does NOT:

| Action | Julius Behavior |
|--------|-----------------|
| Exploit vulnerabilities | No - only sends standard HTTP requests |
| Attempt authentication bypass | No - does not try credentials or auth attacks |
| Perform denial of service | No - respects rate limits and uses reasonable concurrency |
| Modify or delete data | No - only reads responses, never writes |
| Execute code on targets | No - no payload delivery or code execution |
| Brute force credentials | No - no authentication attempts |
| Exfiltrate data | No - only reports service type and available models |
| Persist on systems | No - stateless operation with no implants |

### Network Behavior

Julius generates HTTP traffic that resembles normal API client behavior:

- Standard HTTP methods (GET, POST)
- Common API paths (`/api/tags`, `/health`, `/v1/models`)
- JSON content types
- No malformed requests or fuzzing payloads
- Configurable concurrency and timeouts

### Responsible Use Guidelines

Julius is intended exclusively for:

1. **Authorized penetration testing** - With written permission from system owners
2. **Security assessments** - As part of legitimate security audits
3. **Asset inventory** - On networks you own or administer
4. **Research and education** - In controlled lab environments

### Before Scanning

**Always ensure you have explicit authorization** before scanning any targets:

- Obtain written permission from the system owner
- Verify the scope of authorized testing
- Document authorization for your records
- Follow your organization's security testing policies

### Legal Considerations

Unauthorized scanning of computer systems may violate:

- Computer Fraud and Abuse Act (CFAA) in the United States
- Computer Misuse Act in the United Kingdom
- Similar cybercrime laws in other jurisdictions

**You are responsible for ensuring your use of Julius is lawful and authorized.**

## Security Best Practices for LLM Deployments

If Julius identifies exposed LLM services on your network, consider:

### Immediate Actions

1. **Verify authorization** - Confirm the service should be network-accessible
2. **Check authentication** - Ensure API endpoints require authentication
3. **Review network segmentation** - Limit exposure to authorized networks
4. **Audit access logs** - Check for unauthorized access attempts

### Recommended Configuration

| Service | Security Recommendation |
|---------|------------------------|
| Ollama | Bind to localhost (`OLLAMA_HOST=127.0.0.1`) unless API access is required |
| vLLM | Use `--api-key` flag or reverse proxy authentication |
| LiteLLM | Configure authentication via environment variables |
| LocalAI | Enable API key authentication |
| Open WebUI | Use built-in authentication system |

### Common Misconfigurations

- **Default bindings to 0.0.0.0** - Exposes service to all network interfaces
- **Missing authentication** - Allows unauthorized model access
- **Exposed management APIs** - Enables model loading/unloading by attackers
- **Public cloud deployments** - Internet-accessible LLM endpoints without auth

## Dependencies

Julius uses these third-party dependencies:

| Dependency | Purpose | License |
|------------|---------|---------|
| spf13/cobra | CLI framework | Apache 2.0 |
| goccy/go-yaml | YAML parsing | MIT |
| itchyny/gojq | JQ expression parsing | MIT |
| stretchr/testify | Testing | MIT |
| olekukonko/tablewriter | Table output | MIT |
| golang.org/x/sync | Concurrency utilities | BSD-3 |

All dependencies are regularly reviewed for security vulnerabilities.

## Security Updates

Security updates will be released as patch versions (e.g., 0.1.1) and announced via:

- GitHub Releases
- GitHub Security Advisories (for critical issues)
