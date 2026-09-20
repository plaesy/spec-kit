---
applyTo: '*'
description: "Comprehensive secure coding instructions for all languages and frameworks, based on OWASP Top 10 and industry best practices."
---
# Secure Coding and OWASP Guidelines

## Instructions
All generated/reviewed/refactored code must be secure by default - security-first mindset. When in doubt, choose the more secure option and explain why. Based on OWASP Top 10 and industry best practices.

### 1. A01: Broken Access Control & A10: Server-Side Request Forgery (SSRF)
- **Least privilege**: default to the most restrictive permissions; explicitly check user rights against the resource's required permissions
- **Deny by default**: access granted only via an explicit allow rule
- **Validate SSRF-prone URLs**: treat user-supplied request URLs (e.g. webhooks) as untrusted - strict allow-list validation of host, port, path
- **Prevent path traversal**: sanitize file-upload/file-access input (e.g. `../../etc/passwd`), use APIs that build paths securely

### 2. A02: Cryptographic Failures
- **Strong, modern algorithms**: salted hashing via Argon2 or bcrypt - never MD5/SHA-1 for password storage
- **Data in transit**: default to HTTPS for network requests
- **Data at rest**: encrypt sensitive data (PII, tokens) with AES-256 or similarly strong standard algorithms
- **Secret management**: never hardcode secrets (API keys, passwords, connection strings) - read from environment variables or a secrets manager (HashiCorp Vault, AWS Secrets Manager), with a clear placeholder/comment
  ```javascript
  // GOOD: Load from environment or secret store
  const apiKey = process.env.API_KEY; 
  // TODO: Ensure API_KEY is securely configured in your environment.
  ```
  ```python
  # BAD: Hardcoded secret
  api_key = "sk_this_is_a_very_bad_idea_12345" 
  ```

### 3. A03: Injection
- **No raw SQL**: parameterized queries/prepared statements only - never string concat/formatting from user input
- **Sanitize CLI input**: use functions that handle argument escaping (e.g. `shlex` in Python) to prevent shell injection
- **Prevent XSS**: context-aware output encoding for user-controlled data - prefer `.textContent` over `.innerHTML`; if `innerHTML` is necessary, sanitize with DOMPurify first

### 4. A05: Security Misconfiguration & A06: Vulnerable Components
- **Secure defaults**: disable verbose error messages/debug features in production
- **Security headers**: `Content-Security-Policy` (CSP), `Strict-Transport-Security` (HSTS), `X-Content-Type-Options`
- **Up-to-date dependencies**: suggest latest stable version for new libraries; run `npm audit`, `pip-audit`, or Snyk for known vulnerabilities

### 4.1. STRICT PROHIBITION: Deprecated Libraries and Components
**NON-NEGOTIABLE**: never suggest/recommend/use deprecated libraries, frameworks, APIs, or methods.

**Prohibited**:
- **JS/Node.js**: `request` lib (use `fetch`/`axios`), `var` (use `const`/`let`), callbacks (use async/await), `XMLHttpRequest` (use `fetch`)
- **Python**: `urllib2` (use `requests`), `cgi` (modern frameworks), `optparse` (use `argparse`), `distutils` (use `setuptools`)
- **Java**: legacy `Date`/`Calendar` (use `java.time`), `StringTokenizer` (use `String.split()`), `Hashtable`/`Vector` (use `HashMap`/`ArrayList`)
- **Security**: MD5/SHA-1 for passwords (use bcrypt/Argon2), DES/3DES (use AES), RSA <2048 bits (use 2048+)
- **Web**: jQuery for DOM (vanilla JS/modern frameworks), Bootstrap ≤3 (use 5+), AngularJS (use Angular 2+)

**Before suggesting any library**: verify current status via official docs/GitHub; check for "DEPRECATED"/"legacy"/"no longer maintained" warnings; confirm recent commits (6-12 months) and active issue resolution; check recent releases and community adoption.

**If a deprecated library is encountered**: stop (don't use it) -> identify why it's deprecated and the risk -> recommend a modern, maintained alternative -> include migration guidance if applicable.

**Example Response Template**:
```
❌ DEPRECATED DETECTED: [Library Name] is deprecated because [reasons].
⚠️ RISKS: Security vulnerabilities, no maintenance, compatibility issues.
✅ RECOMMENDED ALTERNATIVE: [Modern Library] - [brief benefits].
📋 MIGRATION: [Simple migration steps or resources].
```

### 5. A07: Identification & Authentication Failures
- **Secure session management**: new session ID on login (prevents fixation); session cookies with `HttpOnly`, `Secure`, `SameSite=Strict`
- **Brute-force protection**: rate limiting and account lockout after repeated failed attempts (login, password reset)

### 6. A08: Software and Data Integrity Failures
- **Prevent insecure deserialization**: never deserialize untrusted data without validation; prefer attack-resistant formats (JSON over Pickle in Python) with strict type checking

## General Guidelines
- **Be explicit**: when suggesting a security mitigation, state what it protects against (e.g. "parameterized query here to prevent SQL injection")
- **Educate in code review**: when flagging a vulnerability, provide the fix AND explain the risk of the original pattern
