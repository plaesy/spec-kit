---
title: "{{PROJECT_NAME}} Constitution"
version: "1.0.0"
ratified: "{{CURRENT_DATE}}"
last_amended: "{{CURRENT_DATE}}"
---

# {{PROJECT_NAME}} Constitution

**Purpose**: The governing decisions for this specific project. Every later phase
(`/assess`, `/implement`, `/optimize`) reads this file before acting — it is
the single source of truth for standards that would otherwise be re-decided (and
re-litigated) on every run.

> Generated once by `/start` (Phase 0). Amend deliberately via `/save` — do not let
> downstream phases silently drift from what's written here.

## 1. Quality Standards

- **Test coverage minimum**: {{TEST_COVERAGE_MIN|90}}%
- **Performance target**: {{PERFORMANCE_TARGET|<200ms p95 response time}}
- **Security bar**: {{SECURITY_BAR|zero known vulnerabilities at Sev-High or above}}
- **Documentation bar**: {{DOC_BAR|complete for all public interfaces}}

## 2. Technology Constraints

- **Approved stack**: {{APPROVED_STACK}}
- **Forbidden/deprecated**: {{FORBIDDEN_TECH|none identified yet}}
- **Validation source**: Context7 (or equivalent) citation required for any new dependency

## 3. Process Rules

- **TDD**: {{TDD_REQUIRED|Required — Red-Green-Refactor, tests before implementation}}
- **Review gate**: implementation is not "done" until a verifier pass (independent of
  the implementing agent/session) checks the diff against the spec — see `/implement`.
- **Ambiguity handling**: unresolved ambiguity blocks `/implement`; run `/assess` first
  (Mode 1 resolves spec ambiguity, not just external research).

## 4. Non-Negotiables (Hard Stops)

List anything that must never be violated, regardless of pressure to ship. Examples:
- {{NON_NEGOTIABLE_1|Never store secrets or credentials in source control}}
- {{NON_NEGOTIABLE_2|Never disable a failing test to make CI pass}}

## 5. Amendment Log

| Date | Change | Reason |
|------|--------|--------|
| {{CURRENT_DATE}} | Initial ratification | Project start |

---
**This file is authoritative.** If a prompt's hardcoded default (e.g. "90% coverage")
conflicts with this file, this file wins — update the prompt if the conflict is
systemic, don't silently override the constitution per-run.
