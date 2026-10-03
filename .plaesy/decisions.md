---
title: "Decisions Index"
description: "Index of project decisions with traceability and access timestamps"
updatedAt: "2026-10-03T16:35:00.000Z"
---

# Decisions Index

**Project:** plaesy/spec-kit — the config-driven spec-kit fork
**Last Updated:** 2026-10-03
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

- [Embedding-based semantic duplicate search](decisions/embedding-based-semantic-duplicate-search.md) — `plaesy search` command; doc-comments for all 15 languages; fail-closed staleness check · 2026-10-03
- [Extract bloated always_load instructions](decisions/extract-always-load-bloat.md) — Self-contained sections move to scope_load; fictional content gets deleted · 2026-10-03
- [Self-provisioning Plaesy home](decisions/self-provisioning-plaesy-home.md) — FindHome extracts the embedded source tree when no home is found · 2026-10-02
- [Canonical platform ids: claude and kilo](decisions/canonical-platform-ids-claude-kilo.md) — Renamed from claude_code/kilo_code; old forms kept as aliases · 2026-10-02
- [Coverage ratchet runs on ubuntu-latest only](decisions/coverage-ratchet-single-os.md) — Several packages fork on GOOS; one baseline can't fit every OS · 2026-10-02
- [Full-autonomy policy + role-handoff chaining](decisions/full-autonomy-role-handoff.md) — Enumerated hard-stops replace "ask when unsure"; no hop limit · 2026-10-02
- [Deleted the markdown baseline](decisions/deleted-markdown-baseline.md) — Replaced the baseline file with zero-tolerance absence semantics · 2026-09-30
- [Accepted mdlint coverage dip](decisions/accepted-mdlint-coverage-dip.md) — Kept coverage at 88.6% → 87.9% with a recorded reason rather than raising the bar · 2026-09-30
- [Treated craft metrics as proxies](decisions/craft-metrics-as-proxies.md) — Ratchet heading presence, body length, and duplication — not whether an example is good · 2026-09-30
- [Kept prompts/ as source and regenerated the mirror](decisions/prompts-source-mirror-regenerated.md) — Verified byte-identical: the runtime loads .kilo/commands/ · 2026-09-30
- [Rebuilt memory.md as the index it is required to be](decisions/rebuilt-memory-index.md) — It held ~145 lines of rules inline, which is why .plaesy/memory/ was empty · 2026-09-30

---

## Superseded / Deprecated Decisions

None yet.

---

## Reference Information

**Decision files**: `.plaesy/decisions/{topic}.md` (flat structure, one topic per file)
**Index**: `.plaesy/decisions.md` (this file)
**Related**: `.plaesy/memory.md` — knowledge base for patterns learned from decisions

---

**Status:** Index rebuilt 2026-09-30. Session decisions migrated from
`.plaesy/context.md` into topic files; the context section is removed because
the decisions index is the single source of truth for decision history.
2026-10-02 added three entries for the self-provisioning-home,
claude/kilo-rename, and coverage-ratchet work that got CI green.
2026-10-03 entry refreshed for the semantic-search decision: doc-comment
coverage completed across all 15 languages, staleness policy resolved as
fail-closed, and the two silent-failure traps recorded there (regex drift from
`symbols.go`, attribute lines between a comment and its declaration).

**Next:** Add project-specific decision entries as they arise; every link above must resolve.
