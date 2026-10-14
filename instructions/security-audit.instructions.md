---
description: "Comprehensive security and compliance audit protocol for all projects"
applyTo: "**/*.js, **/*.ts, **/*.py, **/*.java, **/*.go, **/*.rs, **/*.cs, **/*.rb, **/*.sql, **/*.html"
---

# Security & Compliance Audit Protocol

## Overview

Structured security validation for OWASP Top 10, WCAG accessibility, and compliance requirements.

---

## OWASP Top 10 (2021) High/Critical Checklist

**Category labels are OWASP Top 10:2021, and that is the only numbering used
anywhere in this repo.** `.plaesy/instructions/security-and-owasp.md` uses the same
labels, so "A01" means one thing across both files. An earlier revision of this
file used a private A1-A10 scheme whose numbers meant different things
(`A1` was Broken Authentication, which is `A07`; `A2` was Broken Authorization,
which is `A01`) — reading a finding against the wrong file checked the wrong
category. That scheme is gone.

This checklist verifies, it does not re-teach. Each category below states only the
audit-observable criteria that no other file already fixes; the remediation rule itself
lives with its owner and is referenced, not restated:

- **Coding rules per OWASP category**: `instructions/security-and-owasp.instructions.md`
- **Architectural principles** (least privilege, complete mediation, fail-safe
  defaults, threat modeling before build, attack surface):
  `.plaesy/instructions/security-design-principles.md`
  (source `instructions/security-design-principles.instructions.md`)

### A07: Identification and Authentication Failures

Password hashing, session-cookie flags, and HTTPS enforcement are fixed by
`instructions/security-and-owasp.instructions.md` (§2, §5). This section and the
A07 continuation below were one section in the old private scheme; OWASP 2021
treats authentication, session handling, and CSRF as one category, so they are
split here only to keep each checklist short. Remaining audit criteria:

```markdown

[ ] No plaintext passwords stored anywhere
[ ] Session tokens not exposed in logs/error messages
[ ] Login attempts rate-limited (max 5/min per IP)
[ ] Password reset tokens expire after 15 minutes
[ ] No default credentials (admin/admin, test/test)
```

### A01: Broken Access Control

Per-resource permission checking and explicit allow rules are fixed by
`instructions/security-and-owasp.instructions.md` (§1, A01), on top of the least-privilege
and complete-mediation principles in
`.plaesy/instructions/security-design-principles.md`. Remaining audit criteria:

```markdown

[ ] No privilege escalation possible
[ ] No hard-coded role checks (use RBAC)
```

### A02: Cryptographic Failures

The old private scheme had no slot for this category, so nothing audited it.
Encryption, hashing, and key handling rules are fixed by
`instructions/security-and-owasp.instructions.md` (§2). Remaining audit criteria:

```markdown

[ ] Passwords hashed with a memory-hard algorithm (argon2/bcrypt/scrypt), never MD5/SHA-1
[ ] TLS 1.2+ only; older protocol versions rejected
[ ] No secrets in source, images, or build artifacts
[ ] Encryption keys and secrets sourced from a secret store, not env files committed to git
```

### A03: Injection (SQL, NoSQL, Command)

Parameterized statements and shell-argument escaping are fixed by
`instructions/security-and-owasp.instructions.md` (§3). Remaining audit criteria:

```markdown

[ ] No eval() or similar code execution functions
[ ] GraphQL queries protected against injection
```

### A04: Insecure Design

No checklist of its own: every item this section used to carry — security requirements
in specs, STRIDE threat modeling, design review before implementation, no hardcoded
secrets, no insecure defaults — is a principle owned by
`.plaesy/instructions/security-design-principles.md`
(source `instructions/security-design-principles.instructions.md`), with the secrets
rule in `instructions/security-and-owasp.instructions.md` (§2). An audit confirms the
design review actually happened; it does not restate the principle.

### A05: Security Misconfiguration

Debug mode, error-detail exposure, and HTTPS enforcement are fixed by
`instructions/security-and-owasp.instructions.md` (§2, §4). Remaining audit criteria:

```markdown

[ ] Default credentials changed
[ ] Security headers present (CSP, X-Frame-Options, HSTS)
```

### A06: Vulnerable and Outdated Components

Dependency scanning and the prohibition on deprecated libraries are fixed by
`instructions/security-and-owasp.instructions.md` (§4, §4.1 — the non-negotiable
deprecation list lives there). Remaining audit criteria:

```markdown

[ ] No critical vulnerabilities unpatched
[ ] Security advisory subscribed
```

### A07: Identification and Authentication Failures (cont.)

Session-cookie flags (`HttpOnly`, `Secure`, `SameSite=Strict`, tokens not in the URL)
are fixed by `instructions/security-and-owasp.instructions.md` (§5). Remaining audit
criteria:

```markdown

[ ] Session timeout configured (15-30 min inactivity)
[ ] CSRF tokens on state-changing operations
[ ] Logout invalidates session immediately
```

### A08: Software and Data Integrity Failures

Insecure deserialization is fixed by
`instructions/security-and-owasp.instructions.md` (§6). Remaining audit criteria:

```markdown

[ ] Code changes require review (no self-merge)
[ ] Build artifacts signed/verified
[ ] CI/CD pipeline secured (branch protection)
```

### A09: Security Logging and Monitoring Failures

No rule for this category is stated in the coding or principle files, so the whole
checklist stays here:

```markdown

[ ] Security events logged (login, auth failure, admin actions)
[ ] Sensitive data NOT logged (passwords, tokens, PII)
[ ] Logs retained for audit trail (90+ days)
[ ] Monitoring alerts on suspicious activity
```

### A10: Server-Side Request Forgery (SSRF)

External-URL allow-listing (host, port, path) is fixed by
`instructions/security-and-owasp.instructions.md` (§1). Remaining audit criteria:

```markdown

[ ] Internal network not accessible from user input
[ ] No proxy/tunneling abuse possible
```

---

## WCAG 2.1 AA Accessibility Checklist

### Perceivable

```markdown

[ ] All images have descriptive alt text
[ ] Decorative images marked with empty alt ("")
[ ] Heading hierarchy logical (h1 → h2 → h3, no skips)
[ ] Text contrast ratio ≥4.5:1 (normal text)
[ ] Text contrast ratio ≥3:1 (large text >18pt)
[ ] Text resizable without loss of content
[ ] No blinking/flashing >3 times/second
```

### Operable

```markdown

[ ] All functionality keyboard accessible
[ ] No keyboard trap (can tab out)
[ ] Focus indicator visible (2px+ outline)
[ ] Tab order logical
[ ] No auto-refresh that disrupts reading
```

### Understandable

```markdown

[ ] Page title descriptive
[ ] Link purpose clear (not "click here")
[ ] Form labels visible and descriptive
[ ] Error messages identify problem
[ ] Page language declared (<html lang="en">)
```

### Robust

```markdown

[ ] Valid HTML (W3C validator passes)
[ ] ARIA roles used correctly
[ ] ARIA properties correct (aria-label, aria-describedby)
[ ] No orphaned form controls
```

---

## Security Audit Report Format

```markdown

# Security Audit Report - [Project Name]

## Executive Summary
- Overall Security Rating: X/10
- Critical Issues: [N]
- High Issues: [N]
- Medium Issues: [N]
- Accessibility Issues: [N]

## OWASP Top 10 Summary
Labels are OWASP Top 10:2021 and match the section headings above — A01 is Broken
Access Control, A07 is Identification & Authentication. Do not renumber them in
the report.

- A01 (Broken Access Control): ✅ PASS
- A02 (Cryptographic Failures): ⚠️ 1 issue found
- A03 (Injection): ✅ PASS
- A04 (Insecure Design): ⚠️ 1 issue found
- A05 (Security Misconfiguration): ✅ PASS
- A06 (Vulnerable and Outdated Components): ⚠️ 2 issues found
- A07 (Identification and Authentication Failures): ✅ PASS
- A08 (Software and Data Integrity Failures): ✅ PASS
- A09 (Security Logging and Monitoring Failures): ✅ PASS
- A10 (Server-Side Request Forgery): ✅ PASS

## WCAG 2.1 AA Summary
- Perceivable: ✅ PASS
- Operable: ⚠️ 2 issues found
- Understandable: ✅ PASS
- Robust: ✅ PASS

## Issues Found
1. [File:Line] SQL injection in user search endpoint
2. [File:Line] Weak password hashing (SHA-1)
...
```

---

## Integration with /assess Phase

When `/assess` runs:

- ✅ Auto-check OWASP High/Critical items
- ✅ Check WCAG AA compliance (if UI project)
- ✅ Dependency audit (npm audit / cargo audit)
- ✅ Secret scan for hardcoded credentials

**Assessment Impact**:

- OWASP High/Critical: -20 points per issue (critical fix before shipping)
- WCAG AA violations: -10 points per issue (must fix before UI release)

---

*Security & Compliance Audit Protocol v1.0.0*
*Framework: Spec Kit Constitutional Development Framework*
