---
name: security
description: "Secure coding and security audit: OWASP Top 10 (2021) rules, deprecated-library prohibition, dependency verification, WCAG 2.2 AA checks, and the audit report format. Use when writing security-sensitive code or auditing for security or accessibility issues."
metadata:
  applies-to: "**/*"
---

# Security

Two uses, one file:

- **Designing a system**: apply the architectural principles in `references/principles.md` (defense
  in depth, least privilege, fail-safe defaults, threat modeling before build, agent and
  tool trust boundaries).
- **Writing or reviewing code**: apply the rules in `references/owasp.md`, grouped by OWASP Top 10
  (2021) category, and the prohibitions below.
- **Running an audit** (`/assess`, `/fix`): run the checklists in `references/owasp.md` and the
  WCAG checklist below, then write the report in the format at the end.

Code must be secure by default. When in doubt, choose the more secure option and say
why.

When a mitigation is suggested, say what it protects against (e.g. "parameterized query
here to prevent SQL injection"). When flagging a vulnerability, give the fix and explain
the risk of the original pattern.

---

## Deprecated libraries

**Non-negotiable.** Never suggest, recommend, or use a deprecated library, framework, API,
or method.

### Prohibited

- **JS/Node.js**: `request` (use `fetch` or `axios`), `var` (use `const`/`let`), callbacks (use async/await), `XMLHttpRequest` (use `fetch`)
- **Python**: `urllib2` (use `requests`), `cgi` (use a modern framework), `optparse` (use `argparse`), `distutils` (use `setuptools`)
- **Java**: legacy `Date`/`Calendar` (use `java.time`), `StringTokenizer` (use `String.split()`), `Hashtable`/`Vector` (use `HashMap`/`ArrayList`)
- **Security**: MD5/SHA-1 for passwords (use bcrypt/Argon2), DES/3DES (use AES), RSA under 2048 bits (use 2048+)
- **Web**: jQuery for DOM (use vanilla JS or a modern framework), Bootstrap 3 or older (use 5+), AngularJS (use Angular 2+)

**Before suggesting any library**: verify its current status in official docs or the
GitHub repo. Check for "DEPRECATED", "legacy", or "no longer maintained" warnings.
Confirm recent commits (6-12 months) and active issue resolution, and check recent releases
and community adoption.

**If a deprecated library is encountered**: stop, do not use it. Identify why it is
deprecated and the risk, recommend a maintained alternative, and include migration
guidance where it applies.

```text
❌ DEPRECATED DETECTED: [Library Name] is deprecated because [reasons].
⚠️ RISKS: Security vulnerabilities, no maintenance, compatibility issues.
✅ RECOMMENDED ALTERNATIVE: [Modern Library] - [brief benefits].
📋 MIGRATION: [Simple migration steps or resources].
```

---

## WCAG 2.2 AA accessibility (UI projects)

Run only when the project has a UI.

### Perceivable

```markdown
[ ] All images have descriptive alt text
[ ] Decorative images marked with empty alt ("")
[ ] Heading hierarchy logical (h1 → h2 → h3, no skips)
[ ] Text contrast ratio ≥4.5:1 (normal text)
[ ] Text contrast ratio ≥3:1 (large text >18pt)
[ ] Text resizable without loss of content
[ ] No blinking or flashing more than 3 times per second
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
[ ] Error messages identify the problem
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

## Audit report format

```markdown
# Security Audit Report - [Project Name]

## Executive Summary
- Overall Security Rating: X/10
- Critical Issues: [N]
- High Issues: [N]
- Medium Issues: [N]
- Accessibility Issues: [N]

## OWASP Top 10 Summary
Labels are OWASP Top 10:2021 and match the headings in `references/owasp.md`. Do not renumber them.

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

## WCAG 2.2 AA Summary
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

## Integration with /assess

When `/assess` runs:

- Auto-check OWASP High/Critical items (`references/owasp.md`)
- Check WCAG AA compliance (UI projects only)
- Run a dependency audit (`npm audit`, `cargo audit`, or equivalent)
- Scan for hardcoded credentials

### Assessment impact

- OWASP High/Critical: -20 points per issue (fix before shipping)
- WCAG AA violation: -10 points per issue (fix before UI release)
