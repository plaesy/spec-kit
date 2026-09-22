---
title: "Graft assessment and Spec-Kit architecture gaps"
updatedAt: "2026-09-22T06:28:50Z"
---

# Graft Assessment and Spec-Kit Architecture Gaps

## Scope and evidence
- Assessed TrailHQ Graft against Plaesy Spec-Kit; no implementation was requested.
- Evidence came from direct source inspection, npm metadata, official TrailHQ documentation, targeted tests, and a full test run.
- Confidence is high for source, build, package, and test facts; medium for platform-failure root causes; low for marketing benchmark claims.

## Verified Graft patterns
- Tier-1 deterministic Tree-sitter graph extraction with a per-file cache keyed by extractor version and source fingerprint.
- Byte-identical rebuild invariant, query-time freshness refresh, lock protection, and nonfatal graph-only operation without network or LLM calls.
- Typed edges, wikilinks, lexical retrieval with deterministic graph reranking, MCP tools, Claude Code hooks, workspace federation, and blast-radius analysis.
- Explicit telemetry allowlisting and a local queue; public “No telemetry” wording needs reconciliation with the telemetry documentation.

## Verified results and risks
- Build passed.
- Full suite: 1063 tests, 984 passed, 74 failed.
- Targeted tests: graph traversal 14/14, graph ranking 15/15, graph seeding 7/12, and MCP server 4/5.
- Remaining failures cluster around Windows worktree seeding, fresh-worktree MCP advertisement, and locale-sensitive formatting; classify root causes before treating them as product defects.
- npm audit found one high-severity transitive `js-yaml@3.15.1` issue via `gray-matter@4.0.3`; patched `3.15.2` is available.
- Source package `0.19.0` is ahead of npm/latest Git tag `0.18.0`; do not treat unreleased source behavior as shipped behavior.
- Official speed and cost claims are self-reported product claims, not independent benchmarks.
- Security policy supports only the latest npm release while Graft is pre-1.0.

## Spec-Kit findings
- The analyzer has no cross-run `--if-changed` fast path even though the graph already has a fingerprint and skip-if-unchanged pattern to reuse.
- Bash graph fingerprinting uses GNU-only `find -printf`; macOS needs the analyzer's portable stat fallback.
- PowerShell analyzer hardcodes development tools and build systems and has narrower AI insights than Bash.
- `.plaesy/analysis/.edges.tsv` is tracked while nodes and fingerprint caches are not, creating an inconsistent cache policy.
- Graph documentation contradicts actual watch-mode support.
- CI skips analyzer functional coverage and lacks Bash/PowerShell parity and freshness assertions.
- PowerShell CLI dispatch forwards the `analyze` token as a project path.
- Prompt routers do not gate re-analysis on freshness.

## Decisions and next phase
- Prioritize a shared portable fingerprint plus analyzer `--if-changed`, then parity tests and CI wiring.
- Treat Graft telemetry, security-policy, version, and benchmark discrepancies as assessment caveats, not Spec-Kit requirements.
- Preserve existing uncommitted Spec-Kit changes; no assessment files were edited.
- The final assessment should be in Indonesian with scores, confidence, assumptions, missing information, the top five findings, and the next phase.
