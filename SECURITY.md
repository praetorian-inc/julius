# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.0.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in Julius, please report it responsibly:

1. **Email**: security@praetorian.com
2. **Do not** open a public GitHub issue for security vulnerabilities

We will acknowledge receipt within 48 hours and work with you to understand and address the issue.

## Security Considerations

### What Julius Does

Julius performs HTTP-based service fingerprinting by sending standard web requests to target endpoints. It:

- Sends GET/POST requests to known LLM service endpoints
- Analyzes HTTP responses (status codes, headers, body content)
- Reports identified services with confidence levels

### What Julius Does NOT Do

Julius is designed as a passive identification tool. It does NOT:

- Exploit vulnerabilities
- Attempt authentication bypass
- Perform denial of service
- Modify or delete data
- Execute code on targets
- Brute force credentials

### Responsible Use

Julius is intended for:

- Authorized penetration testing
- Security assessments with proper authorization
- Asset inventory and discovery on networks you own or have permission to scan
- Research and education

**Always ensure you have explicit authorization before scanning any targets.**

## Dependency Security

We monitor dependencies for known vulnerabilities using:

- GitHub Dependabot
- Regular dependency updates

Report dependency-related security concerns to security@praetorian.com.
