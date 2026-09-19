---
description: "Comprehensive security and compliance audit protocol for all projects"
applyTo: "**/*.js, **/*.ts, **/*.py, **/*.java, **/*.go, **/*.rs, **/*.cs, **/*.rb, **/*.sql, **/*.html"
---

# Security & Compliance Audit Protocol

## Overview
Structured security validation for OWASP Top 10, WCAG accessibility, and compliance requirements.

---

## OWASP Top 10 High/Critical Checklist

### A1: Broken Authentication
```
[ ] Password hashing uses strong algorithm (bcrypt/Argon2, NOT MD5/SHA-1)
[ ] No plaintext passwords stored anywhere
[ ] Session tokens not exposed in logs/error messages
[ ] Login attempts rate-limited (max 5/min per IP)
[ ] Password reset tokens expire after 15 minutes
[ ] No default credentials (admin/admin, test/test)
```

### A2: Broken Authorization
```
[ ] User can only access their own resources
[ ] Admin endpoints require admin role
[ ] API endpoints check permissions before returning data
[ ] No privilege escalation possible
[ ] No hard-coded role checks (use RBAC)
```

### A3: Injection (SQL, NoSQL, Command)
```
[ ] All database queries use parameterized statements
[ ] No string concatenation in SQL queries
[ ] Input sanitized before use in system commands
[ ] No eval() or similar code execution functions
[ ] GraphQL queries protected against injection
```

### A4: Insecure Design
```
[ ] Security requirements defined in specs
[ ] Threat modeling completed (STRIDE or similar)
[ ] Security design reviewed before implementation
[ ] No hardcoded secrets in code
[ ] No insecure defaults (all secure settings explicit)
```

### A5: Security Misconfiguration
```
[ ] No debug mode enabled in production
[ ] Error messages don't expose system details
[ ] Default credentials changed
[ ] Security headers present (CSP, X-Frame-Options, HSTS)
[ ] HTTPS enforced (no HTTP in production)
```

### A6: Vulnerable Outdated Components
```
[ ] Dependency audit run: npm audit / pip audit / cargo audit
[ ] No critical vulnerabilities unpatched
[ ] Deprecated libraries removed
[ ] Security advisory subscribed
```

### A7: Authentication & Session Management
```
[ ] Session timeout configured (15-30 min inactivity)
[ ] Session tokens not in URL (cookies only, httpOnly flag)
[ ] CSRF tokens on state-changing operations
[ ] Logout invalidates session immediately
```

### A8: Software & Data Integrity Failures
```
[ ] Code changes require review (no self-merge)
[ ] Build artifacts signed/verified
[ ] CI/CD pipeline secured (branch protection)
[ ] No insecure deserialization
```

### A9: Logging & Monitoring Failures
```
[ ] Security events logged (login, auth failure, admin actions)
[ ] Sensitive data NOT logged (passwords, tokens, PII)
[ ] Logs retained for audit trail (90+ days)
[ ] Monitoring alerts on suspicious activity
```

### A10: Server-Side Request Forgery (SSRF)
```
[ ] External URLs validated before fetching
[ ] Internal network not accessible from user input
[ ] No proxy/tunneling abuse possible
```

---

## WCAG 2.1 AA Accessibility Checklist

### Perceivable
```
[ ] All images have descriptive alt text
[ ] Decorative images marked with empty alt ("")
[ ] Heading hierarchy logical (h1 → h2 → h3, no skips)
[ ] Text contrast ratio ≥4.5:1 (normal text)
[ ] Text contrast ratio ≥3:1 (large text >18pt)
[ ] Text resizable without loss of content
[ ] No blinking/flashing >3 times/second
```

### Operable
```
[ ] All functionality keyboard accessible
[ ] No keyboard trap (can tab out)
[ ] Focus indicator visible (2px+ outline)
[ ] Tab order logical
[ ] No auto-refresh that disrupts reading
```

### Understandable
```
[ ] Page title descriptive
[ ] Link purpose clear (not "click here")
[ ] Form labels visible and descriptive
[ ] Error messages identify problem
[ ] Page language declared (<html lang="en">)
```

### Robust
```
[ ] Valid HTML (W3C validator passes)
[ ] ARIA roles used correctly
[ ] ARIA properties correct (aria-label, aria-describedby)
[ ] No orphaned form controls
```

---

## Security Audit Report Format

```
# Security Audit Report - [Project Name]

## Executive Summary
- Overall Security Rating: X/10
- Critical Issues: [N]
- High Issues: [N]
- Medium Issues: [N]
- Accessibility Issues: [N]

## OWASP Top 10 Summary
- A1 (Broken Authentication): ✅ PASS
- A2 (Broken Authorization): ⚠️ 1 issue found
- ... (all 10)

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
