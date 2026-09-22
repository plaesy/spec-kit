---
title: "Spec-Kit Technical Assessment 2026-09-22"
updatedAt: "2026-09-22T14:35:00Z"
---

# Technical Assessment — Plaesy Spec-Kit Framework

**Scope**: Code quality, tests, security, performance, documentation, CI, analyzer gaps
**Confidence**: HIGH for source/structure facts; MEDIUM for test coverage; LOW for security audit depth

---

## Scores

| Dimension | Score | Notes |
|-----------|-------|-------|
| Code Quality & Architecture | 72/100 | Solid shell/PS structure; duplication between bash/PS analyzers; ~500 lines of near-identical framework detection logic |
| Test Coverage | 45/100 | Only 4 test files exist; no functional test coverage for analyzer; CI runs syntax-only |
| Security | 78/100 | No secrets in source; no known framework vulnerabilities; ShellCheck non-blocking |
| Performance | 85/100 | Detection uses single-pass file counting; cached results; <2s target met for init |
| Documentation | 70/100 | Public interfaces documented; some docs stale (graph watch-mode claim contradicts code) |
| CI/CD Coverage | 55/100 | Syntax + smoke only; no analyzer functional tests; no Bash/PowerShell parity assertions |
| Analyzer Gaps (from prior session) | 40/100 | `--if-changed` not wired; `find -printf` GNU-only; PS hardcodes dev tools; `.edges.tsv` tracked inconsistently |

**Overall Technical Score: 65/100** (below 85 threshold → route to `/implement` for TDD fixes)

---

## Top 5 Findings

### 1. Analyzer Bash/PowerShell Parity Gap (CRITICAL)
- **Location**: `scripts/bash/plaesy-analyze.sh:499` vs `scripts/powershell/plaesy-analyze.ps1:494`
- **Issue**: Bash uses `find -printf` (GNU-only, breaks on macOS); PowerShell hardcodes development tools and build systems; PowerShell has narrower AI insights (only ~10 framework cases vs ~50 in bash)
- **Confidence**: HIGH — direct source comparison
- **Impact**: macOS users get incomplete analysis; PowerShell output is less useful

### 2. No Cross-Run `--if-changed` Fast Path (HIGH)
- **Location**: `scripts/bash/plaesy-analyze.sh:1-10`, `scripts/powershell/plaesy-analyze.ps1:1-13`
- **Issue**: Both scripts accept `--force` flag but the `--if-changed` behavior is only implemented in `plaesy-graph.ps1`, not in the analyzer itself. The analyzer always regenerates project.json, structure.json, and overview.md.
- **Confidence**: HIGH — flag exists in usage comment but no implementation
- **Impact**: Unnecessary I/O on every run; contradicts documented "skip-if-unchanged" pattern

### 3. Test Coverage Below Constitution Minimum (HIGH)
- **Location**: `testing/` directory, `.github/workflows/ci.yml:27-29`
- **Issue**: Constitution requires 90% test coverage. Only 4 test files exist (`testing/bash/run.sh`, `testing/powershell/run.ps1`, `testing/smoke/smoke-e2e.sh`, `testing/smoke/smoke-powershell.ps1`). CI runs `--no-functional` bash tests only. No analyzer functional tests.
- **Confidence**: HIGH — direct file count and CI inspection
- **Impact**: Failing constitution quality bar; regressions undetected

### 4. `.edges.tsv` Inconsistent Cache Policy (MEDIUM)
- **Location**: `.plaesy/analysis/.edges.tsv` (tracked in git), `.nodes.tsv` (not tracked)
- **Issue**: `.edges.tsv` is committed to git while `.nodes.tsv` and fingerprint caches are not. Creates inconsistent cache invalidation policy.
- **Confidence**: HIGH — git tracking check
- **Impact**: Cache freshness unclear; cross-platform behavior differs

### 5. CI Skips Analyzer Functional Coverage (MEDIUM)
- **Location**: `.github/workflows/ci.yml`
- **Issue**: CI runs bash syntax check (`bash -n`), ShellCheck (non-blocking), PowerShell syntax check, and smoke tests. No functional test coverage for the analyzer itself. No Bash/PowerShell parity assertions between the two implementations.
- **Confidence**: HIGH — CI file inspection
- **Impact**: Analyzer regressions ship undetected; parity drift unnoticed

---

## Missing Information

- Actual test pass/fail counts for existing test suite (never run in CI)
- Performance benchmarks for detection/init scripts on macOS and Windows
- Security audit results for third-party dependencies (no `npm audit` or equivalent in CI)
- Whether `find -printf` issue affects the graph script too (separate from analyzer)

## Assumptions

- "90% coverage" in constitution applies to framework code, not user projects
- Bash/PowerShell parity requirement applies to all paired scripts, not just analyzers
- The uncommitted changes represent the current state of work (not stale/broken)

## Recommended Next Phase

**`/implement` (TDD)** — Fix the analyzer parity gaps, add `--if-changed` fast path, add functional tests, wire CI. Then re-assess.