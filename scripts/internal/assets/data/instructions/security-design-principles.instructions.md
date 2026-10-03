---
applyTo: '**'
description: 'Architectural security design principles (defense in depth, least privilege, fail-safe defaults, attack surface, threat modeling) applied when designing a system — the "why" underneath instructions/security-and-owasp.instructions.md''s implementation checklist.'
---

# Security Design Principles

The core rule: **assume any single control will eventually fail, and design so that
failure doesn't mean compromise.** No layer, credential, or component should be the
one thing standing between an attacker and the system.

This file sits one level above `.plaesy/instructions/security-and-owasp.md` — that file is the
implementation checklist (how to write a specific line of code against a specific
OWASP category); this file is the architectural reasoning applied when shaping a
system before any code is written. Don't duplicate the OWASP checklist here.

Sources (retrieved 2026-09-27): [Saltzer & Schroeder, "The Protection of Information
in Computer Systems" (1975)](https://www.jeremyjordan.me/security-design/) — the
foundational eight principles, still cited as current 50 years later —
[dwheeler.com — Secure Programs HOWTO: security design principles](https://dwheeler.com/secure-programs/Secure-Programs-HOWTO/follow-good-principles.html),
[Microsoft — Secure by Design](https://www.microsoft.com/en-us/securityengineering/sdl/practices/secure-by-design),
[SoftwareSecured — STRIDE Threat Model Explained](https://www.softwaresecured.com/post/stride-threat-modelling),
[StrongDM — What is an Attack Surface?](https://www.strongdm.com/blog/attack-surface).

## Defense in depth

Stack independent layers of control so one failing doesn't expose everything behind
it. — A system that relies on a single perimeter firewall as its only control means
one misconfigured rule, one exposed port, or one compromised edge box hands the
attacker the whole internal network; add network segmentation, per-service auth, and
least-privilege data access behind it, so no single layer's failure is fatal.

## Fail-safe defaults (deny by default)

Base access decisions on explicit permission, not on exclusion. — A resource that's
reachable unless explicitly blocked will eventually be reached through a case nobody
thought to block; a resource that's unreachable unless explicitly allowed fails
closed instead of open when a rule is missing.

```text
# BAD: fail-open — anything not explicitly denied gets through
iptables -P INPUT ACCEPT
iptables -A INPUT -s 10.0.0.0/8 -j DROP   # only known-bad is blocked

# GOOD: fail-closed — anything not explicitly allowed is dropped
iptables -P INPUT DROP
iptables -A INPUT -s 10.0.0.0/8 -p tcp --dport 443 -j ACCEPT   # only known-good is let in
```

## Least privilege

Every user, service, and process should hold only the access it needs for its
current task, no more. — A backend service account with full database admin rights,
used only to read one table, turns a single SQL-injection or credential leak into
full data-store compromise instead of a narrow read; scope the credential to exactly
the tables and operations that service actually performs.

## Complete mediation

Check every access to every protected resource, every time, at a point that can't be
bypassed. — Caching an authorization decision at request-start and reusing it for the
rest of a long-lived session lets a mid-session permission revocation (a fired
employee, a downgraded role) go unenforced until the session happens to end; check
authority at the point of each access, not once up front.

## Minimize attack surface (economy of mechanism)

Keep the system as simple and small as the requirement allows, and remove anything
not actively needed. — A production server left with SSH, an unused admin console,
and a debug endpoint all reachable "because it might be useful later" gives an
attacker three doors instead of one; disable and remove what isn't in active use, not
just leave it unauthenticated-but-present.

## Separation of privilege (separation of duties)

Require more than one condition, or more than one party, for the most sensitive
actions. — A single engineer with the credentials to both approve and deploy a
production database migration can push an unreviewed, destructive change alone,
whether by mistake or by a compromised account; requiring a second approver for that
class of action means one compromised credential isn't enough on its own.

## Open design (no security through obscurity)

Security should hold up even if the attacker knows exactly how the mechanism works —
don't rely on a secret design, only on a secret key/credential. — An admin panel
reachable at an unlisted URL with no auth check is "secure" only until the URL leaks
(a referrer header, a search index, a misconfigured proxy log) — at which point there
is no security left at all; put a real auth check on it regardless of how hard the
URL is to guess.

## Threat model before building, not after

Walk through how the system could be attacked (e.g. STRIDE — Spoofing, Tampering,
Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege) at
design/architecture-review time, per component and per data flow, before
implementation locks in the shape. — A design finalized without this pass tends to
surface the same class of flaw (missing auth on an internal-only-sounding endpoint,
an unsigned webhook payload) only after it's in production and already exploitable,
when it costs far more to fix than at the design stage.

## Applying these under real constraints

- Defense in depth and separation of duties both add operational cost and friction —
  scale how many layers/approvals a system carries to the actual sensitivity and
  blast radius of what it protects, not uniformly to every service regardless of risk.
- These are design-time principles; they inform *what* to build, while
  `.plaesy/instructions/security-and-owasp.md` governs *how* to write the code that
  implements it — use both together, don't treat one as a substitute for the other.
- When reviewing an architecture or design doc, name which principle is violated and
  the concrete compromise scenario it enables (as in the examples above), not just
  "this isn't secure."
