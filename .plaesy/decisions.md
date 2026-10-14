---
title: "Decisions Index"
description: "Index of project decisions with traceability and access timestamps"
updatedAt: "2026-10-10T20:30:00.000Z"
---

# Decisions Index

**Project:** plaesy/spec-kit — the config-driven spec-kit fork
**Last Updated:** 2026-10-08
**Purpose:** Index of project decisions with traceability and access timestamps

---

## How This Index Works

This file is an index only — it links to decision topic files under
`.plaesy/decisions/`. Each entry records:

- **Topic** — the decision subject (matches the filename)
- **Decision** — the chosen outcome in one line
- **References** — links to source material that informed or supports the decision
- **Accessed** — timestamp when the reference was last consulted

Add a link here only once the file it points at exists. Keep descriptions under
100 chars. No nested bullets or sub-links — flat list only.

---

## Active Decisions

- [Graph extension lists consolidated](decisions/graph-extension-consolidation.md) — One SourceExtensions list; .hh and .cxx collected · 2026-10-13
- [Analysis fingerprint records built components](decisions/fingerprint-component-tracking.md) — --no-graph records nothing; skips need the components · 2026-10-13

- [Delegation provider and token savers](decisions/delegation-provider-and-token-savers.md) — plaesy delegate + fallback chains; no gateway, no account pooling · 2026-10-13
- [Remove ONNX and HTTP embedders; rembed only](decisions/remove-onnx-and-http-embedders.md) — Pure-Go rembed is the only embedder; onnxruntime and cgo dropped · 2026-10-13
- [Chunk cap in model tokens (480)](decisions/chunk-cap-model-tokens.md) — Size counted by the e5 tokenizer; oversized text split; MaxChunkTokens was never wired · 2026-10-13
- [Hybrid matches before parent-child](decisions/hybrid-matches-before-parent-child.md) — Truncate to top-K before expanding siblings; siblings listed once, after matches · 2026-10-13
- [Retrieval prefix and cap on 101 labels](decisions/retrieval-prefix-and-cap-101-labels.md) — E5 prefixes help; cap 256 better for hybrid and vector; default change pending · 2026-10-13
- [Nara-first routing and no hard stops](decisions/nara-first-no-stop.md) — Nara decides all items under NN-01..07; no run stops for the user · 2026-10-14
- [One search verb; analyze builds the indexes; E5 everywhere](decisions/search-single-verb-and-e5-index.md) — `plaesy search "<text>"` only; graph removed; watch moved to analyze · 2026-10-13
- [Framework Governance Hardening](decisions/framework-governance-hardening-2026-10-10.md) — Background-dispatch scope, single-committer, Subagent Invocation Contract, MA-01, 4 docs-drift repairs · 2026-10-10
- [Search CLI Restructure](decisions/search-cli-restructure.md) — Global `doctor`/`stats`; reach doctor/stats removed; search stats subcommand removed · 2026-10-08
- [Embedding-based semantic duplicate search](decisions/embedding-based-semantic-duplicate-search.md) — `plaesy search`; `plaesy analyze` refreshes the index by default; fail-closed staleness · 2026-10-04
- [Extract bloated always_load instructions](decisions/extract-always-load-bloat.md) — Self-contained sections move to scope_load; fictional content gets deleted · 2026-10-03
- [Self-provisioning Plaesy home](decisions/self-provisioning-plaesy-home.md) — FindHome extracts the embedded source tree when no home is found · 2026-10-02
- [Canonical platform ids: claude and kilo](decisions/canonical-platform-ids-claude-kilo.md) — Renamed from claude_code/kilo_code; old forms kept as aliases · 2026-10-02
- [Coverage ratchet runs on ubuntu-latest only](decisions/coverage-ratchet-single-os.md) — Several packages fork on GOOS; one baseline can't fit every OS · 2026-10-02
- [Full-autonomy policy + role-handoff chaining](decisions/full-autonomy-role-handoff.md) — Hard-stop list superseded 2026-10-14; routing and no hop limit still apply · 2026-10-02
- [Deleted the markdown baseline](decisions/deleted-markdown-baseline.md) — Replaced the baseline file with zero-tolerance absence semantics · 2026-09-30
- [Accepted mdlint coverage dip](decisions/accepted-mdlint-coverage-dip.md) — Kept coverage at 88.6% → 87.9% with a recorded reason rather than raising the bar · 2026-09-30
- [Treated craft metrics as proxies](decisions/craft-metrics-as-proxies.md) — Ratchet heading presence, body length, and duplication — not whether an example is good · 2026-09-30
- [Kept prompts/ as source and regenerated the mirror](decisions/prompts-source-mirror-regenerated.md) — Verified byte-identical: the runtime loads .kilo/commands/ · 2026-09-30
- [Rebuilt memory.md as the index it is required to be](decisions/rebuilt-memory-index.md) — It held ~145 lines of rules inline, which is why .plaesy/memory/ was empty · 2026-09-30

---

- [Late-interaction retrieval not adopted](decisions/late-interaction-not-adopted.md) — MaxSim gave +2 hit@1 of 19 and lost hit@3; 14x latency, 230x memory · 2026-10-13
- [Reranker deferred](decisions/reranker-deferred.md) — bge-reranker-v2-m3 is Apache-2.0 but needs a rembed patch and a 2.3 GB model · 2026-10-13

## Superseded / Deprecated Decisions

- [Release: native matrix instead of CGO_ENABLED=0 cross-compile](decisions/release-native-matrix-cgo.md) — Superseded 2026-10-13: onnxruntime_go removed, so cgo is not required · 2026-10-10

---

## Reference Information

**Decision files**: `.plaesy/decisions/{topic}.md` (flat structure, one topic per file)
**Index**: `.plaesy/decisions.md` (this file)
**Related**: `.plaesy/memory.md` — knowledge base for patterns learned from decisions

---

## Status

Index rebuilt 2026-09-30. Session decisions migrated from
`.plaesy/context.md` into topic files; the context section is removed because
the decisions index is the single source of truth for decision history.
2026-10-02 added three entries for the self-provisioning-home,
claude/kilo-rename, and coverage-ratchet work that got CI green.
2026-10-03 entry refreshed for the semantic-search decision: doc-comment
coverage completed across all 15 languages, staleness policy resolved as
fail-closed, and the two silent-failure traps recorded there (regex drift from
`symbols.go`, attribute lines between a comment and its declaration).
2026-10-04: that entry now also records `plaesy analyze` refreshing the semantic
index by default (`--no-index` opts out), which closes the loop the fail-closed
staleness check left open.
2026-10-08: Search CLI Restructure added — global `doctor`/`stats` verbs; `reach doctor/stats` removed; `search stats` subcommand removed.
2026-10-10: Framework Governance Hardening added — 8 fixes to instructions/prompts/agents/templates/checklists (background-dispatch scope, single-committer, Subagent Invocation Contract, MA-01) plus 4 pre-existing docs-drift test failures repaired (CHANGELOG, reference.md, metadata.json, README CLI blocks). Verified via `plaesy validate` + `go build` + `go test ./...`, all green.

- [Internet Access as a tenth dimension](decisions/internet-access-dimension.md) — superseded · 2026-10-14
- [Remove the internet_access dimension](decisions/remove-internet-access-dimension.md) — nine dimensions again; platform health stays in `plaesy doctor` · 2026-10-14
- [Revert without history reset](decisions/revert-without-history-reset.md) — pre-edit snapshots under `.plaesy/tmp`, no `git reset` · 2026-10-14
- [Research fan-out inside plaesy](decisions/research-fan-out-in-plaesy.md) — `researcher` subagents, no external skill · 2026-10-14

**Next:** Add project-specific decision entries as they arise; every link above must resolve.