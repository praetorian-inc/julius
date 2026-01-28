# Security Policy

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
