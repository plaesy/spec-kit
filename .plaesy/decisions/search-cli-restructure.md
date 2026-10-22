---
title: "Search CLI Restructure"
description: "Moved reach doctor/stats to global doctor/stats; removed search stats subcommand; consolidated under top-level verbs"
updatedAt: "2026-10-08T19:35:00.000Z"
---

# Decision: Search CLI Restructure

## Summary

Restructured `plaesy` CLI to use top-level verbs for cross-cutting concerns (`doctor`, `stats`) and keep domain-specific operations under their parent (`reach`, `search`).

---

## Problem

Before restructure:
- `plaesy reach doctor` — checked reach platform health only
- `plaesy reach stats` — showed reach cache stats only
- `plaesy search stats` — showed search index stats only
- `plaesy search --index` / `plaesy search "query"` — legacy code symbol search (flags)
- `plaesy search index` / `plaesy search query` — new document search (subcommands)
- `plaesy search stats` — duplicate stats subcommand

Issues:
1. **Discoverability**: User didn't know `reach doctor` existed or that it only checked reach
2. **Duplication**: Two `stats` commands showing different subsets of system state
3. **Inconsistency**: Cross-cutting health/stats hidden under domain commands
4. **Legacy confusion**: `search --index` (flag) vs `search index` (subcommand) for different things

---

## Decision

1. **Create `plaesy doctor`** — global health check for all capabilities (reach platforms + search index)
2. **Create `plaesy stats`** — global statistics for search index + reach cache
3. **Remove `plaesy reach doctor`** and `plaesy reach stats`
4. **Remove `plaesy search stats`** subcommand
5. **Keep `plaesy reach list/configure/query`** — domain-specific
6. **Keep `plaesy search index/query`** — domain-specific
7. **Preserve `plaesy search --index` / `plaesy search "query"`** — legacy code symbol search via flags
8. **Update help text**: `plaesy search` help references `plaesy stats`

---

## Rationale

**Top-level verbs for cross-cutting concerns:**
- `doctor` and `stats` are generic system operations, not reach-specific or search-specific
- Users expect `plaesy doctor` to check everything, not just reach
- Single command for "is the system healthy?" and "what's the system state?"

**Subcommands for domain-specific operations:**
- `reach list/configure` only make sense for reach platforms
- `search index/query` only make sense for document search
- `reach web "..."` direct query is the primary reach workflow

**Legacy preservation:**
- Code symbol search (`--index`, query) is a different feature from document search
- Flags on parent command avoid breaking existing scripts
- Clear separation: `--index` = code symbols, `index` = documents

---

## Implementation

**Files changed:**
- `scripts/cmd/plaesy/doctor.go` — new top-level command (merged reach doctor + search index check)
- `scripts/cmd/plaesy/stats.go` — new top-level command (search index + reach cache)
- `scripts/cmd/plaesy/reach.go` — removed doctor/stats subcommands and handlers
- `scripts/cmd/plaesy/search.go` — removed stats subcommand, updated help text
- `scripts/cmd/plaesy/search_index.go` — unchanged
- `scripts/cmd/plaesy/search_query.go` — unchanged

**Key patterns used:**
- `ConfigManager[T]` for unified config access
- `reach.NewRegistry` with cache/rate-limiter for health checks
- `os.Stat` for search index existence checks
- Build tags for Windows ONNX stub (`onnx_windows.go`)

---

## Alternatives Considered

1. **Keep `reach doctor` + add `search doctor`** — rejected: fragments health check across commands
2. **Keep both `reach stats` and `search stats`** — rejected: user runs two commands for full picture
3. **Rename `search --index` to `search build-symbol-index`** — rejected: breaks existing scripts; flags are fine for legacy
4. **Merge `reach` into `search`** — rejected: fundamentally different capabilities (external vs local)

---

## Consequences

**Positive:**
- Single `plaesy doctor` tells you if the whole system is healthy
- Single `plaesy stats` shows complete system state
- Clear mental model: global verbs at top, domain verbs under parents
- Help text guides users to right command

**Negative:**
- Breaking change for users who used `plaesy reach doctor` / `plaesy reach stats` / `plaesy search stats`
- Migration: update scripts/aliases/documentation

**Mitigation:**
- Help text shows new commands
- Old commands show "unknown command" with suggestions (`doctor`, `stats`)
- No functional loss — same checks, better organization

---

## Verification

- `plaesy doctor` — checks 16 reach platforms + search index (vector/text/document stores)
- `plaesy stats` — shows vector store, text index, document store sizes + reach cache stats
- `plaesy search index ./docs` — builds document index
- `plaesy search query "golang" --json` — hybrid search with citations
- `plaesy reach list` — 16 platforms
- `plaesy reach web "https://..."` — fetch via Jina Reader
- All unit tests pass: `go test ./internal/reach/... ./internal/config/...`

---

## References

- [CLI Restructure Patterns](memory/cli-restructure-patterns.md)
- [Semantic Search Feature](memory/semantic-search-feature.md)
- [Unified Config Migration](decisions/unified-config-migration.md)

---

## Access Log

- 2026-10-08 — Decision made and implemented in current session