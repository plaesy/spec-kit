# OWASP Top 10 (2021)

Category labels are OWASP Top 10:2021 and are the only numbering used in this repo.
Each category states the **coding rule** (what to write) and the **audit criteria**
(what to check). A category with no rule here is not a gap: A04 and A09 are owned by
other files, as noted below.

Architectural principles (least privilege, complete mediation, fail-safe defaults,
threat modeling before build, attack surface) are in `principles.md`.

---

## A01: Broken Access Control

### Rule

- Prevent path traversal: sanitize file-upload and file-access input (e.g. `../../etc/passwd`), and use APIs that build paths securely.
- Least privilege and deny-by-default are architectural principles, owned by `principles.md`. Apply them; do not restate them here.

### Audit

```markdown
[ ] No privilege escalation possible
[ ] No hard-coded role checks (use RBAC)
[ ] Per-resource permission checked on every request, not only in the UI
```

---

## A02: Cryptographic Failures

### Rule

- Hashing: salted Argon2 or bcrypt for passwords. Never MD5 or SHA-1 for password storage.
- In transit: default to HTTPS for every network request.
- At rest: encrypt sensitive data (PII, tokens) with AES-256 or an equivalent standard algorithm.
- Secrets: never hardcode API keys, passwords, or connection strings. Read them from environment variables or a secrets manager (HashiCorp Vault, AWS Secrets Manager), with a clear placeholder.

  ```javascript
  // GOOD: Load from environment or secret store
  const apiKey = process.env.API_KEY;
  // TODO: Ensure API_KEY is securely configured in your environment.
  ```

  ```python
  # BAD: Hardcoded secret
  api_key = "sk_this_is_a_very_bad_idea_12345"
  ```

### Audit

```markdown
[ ] Passwords hashed with a memory-hard algorithm (argon2/bcrypt/scrypt), never MD5/SHA-1
[ ] TLS 1.2+ only; older protocol versions rejected
[ ] No secrets in source, images, or build artifacts
[ ] Encryption keys and secrets sourced from a secret store, not env files committed to git
```

---

## A03: Injection

### Rule

- No raw SQL: use parameterized queries or prepared statements only. Never concatenate or format user input into a query.
- Shell: use functions that escape arguments (e.g. `shlex` in Python) to prevent command injection.
- XSS: context-aware output encoding for user-controlled data. Prefer `.textContent` over `.innerHTML`. If `innerHTML` is necessary, sanitize with DOMPurify first.

### Audit

```markdown
[ ] No eval() or similar code execution functions
[ ] GraphQL queries protected against injection
```

---

## A04: Insecure Design

No coding rule in this file. Owned by `principles.md`.
The audit confirms that a design review happened and that threat modeling was done
before build. It does not restate the principle.

**Audit**: see `principles.md`. The secrets rule is in A02 above.

---

## A05: Security Misconfiguration

### Rule

- Secure defaults: disable verbose error messages and debug features in production.
- Security headers: `Content-Security-Policy` (CSP), `Strict-Transport-Security` (HSTS), `X-Content-Type-Options`.

### Audit

```markdown
[ ] Default credentials changed
[ ] Security headers present (CSP, X-Frame-Options, HSTS)
[ ] Debug mode and error-detail exposure disabled in production
```

---

## A06: Vulnerable and Outdated Components

### Rule

- Use the latest stable version of any new library. Run `npm audit`, `pip-audit`, or Snyk to find known vulnerabilities.
- **Deprecated libraries are prohibited.** Never suggest, recommend, or use a deprecated library, framework, API, or method. The list lives in `skills/security/SKILL.md`, section "Deprecated libraries".

### 6.2. Dependency Existence (Slopsquatting)

- **Check every new package name against its registry before it goes into a manifest.**
  Confirm the exact name exists, the publisher and repository are the ones you expect,
  and the release history is plausible. A name that a model suggested is not evidence
  that the package exists.
- **Why**: models sometimes invent package names. An attacker can register an invented
  name that does not exist yet, and an install that follows the generated text then pulls
  that package. In a study of 16 code-generating models, 576,000
  samples, hallucinated packages appeared in at least 5.2% of commercial-model samples and
  21.7% of open-source-model samples (Spracklen et al., arXiv:2406.10279, USENIX Security 2025).
- **Pin and commit the lockfile** for every new dependency. An install command copied from
  generated text is run only after the name in it has passed the check above.

### Audit

```markdown
[ ] No critical vulnerabilities unpatched
[ ] Security advisory subscribed
[ ] Every dependency name verified against its registry; lockfile committed
```

---

## A07: Identification and Authentication Failures

### Rule

- Session management: issue a new session ID on login (prevents fixation). Session cookies use `HttpOnly`, `Secure`, `SameSite=Strict`, and tokens never appear in the URL.
- Brute force: rate limiting and account lockout after repeated failed attempts (login, password reset).

### Audit

```markdown
[ ] No plaintext passwords stored anywhere
[ ] Session tokens not exposed in logs or error messages
[ ] Login attempts rate-limited (max 5/min per IP)
[ ] Password reset tokens expire after 15 minutes
[ ] No default credentials (admin/admin, test/test)
[ ] Session timeout configured (15-30 min inactivity)
[ ] CSRF tokens on state-changing operations
[ ] Logout invalidates the session immediately
```

OWASP 2021 treats authentication, session handling, and CSRF as one category, so they are
kept together here.

---

## A08: Software and Data Integrity Failures

### Rule

- Insecure deserialization: never deserialize untrusted data without validation. Prefer attack-resistant formats (JSON over Pickle in Python) with strict type checking.

### Audit

```markdown
[ ] Code changes require review (no self-merge)
[ ] Build artifacts signed and verified
[ ] CI/CD pipeline secured (branch protection)
```

---

## A09: Security Logging and Monitoring Failures

No coding rule in this file. The audit checklist below is the only statement of this category.

### Audit

```markdown
[ ] Security events logged (login, auth failure, admin actions)
[ ] Sensitive data NOT logged (passwords, tokens, PII)
[ ] Logs retained for audit trail (90+ days)
[ ] Monitoring alerts on suspicious activity
```

---

## A10: Server-Side Request Forgery (SSRF)

### Rule

- Treat user-supplied request URLs (e.g. webhooks) as untrusted. Validate host, port, and path against a strict allow-list.

### Audit

```markdown
[ ] External-URL allow-list enforced (host, port, path)
[ ] Internal network not reachable from user input
[ ] No proxy or tunneling abuse possible
```
