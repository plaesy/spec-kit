---
title: "CLI Restructure Patterns"
description: "Top-level verbs, subcommand consolidation, global vs scoped commands, config unification"
updatedAt: "2026-10-08T19:30:00.000Z"
---

# CLI Restructure Patterns

Lessons from restructuring `plaesy reach`, `plaesy search`, and adding `plaesy doctor`/`plaesy stats`.

---

## Principle: Distinct Verbs → Top-Level Commands

A genuinely new user-facing verb (`doctor`, `stats`, `search`, `reach`) gets its own top-level command, not flags on an existing command.

**Before (anti-pattern):**
```
plaesy reach doctor       # health check flag on reach
plaesy reach stats        # stats flag on reach
plaesy search --index     # semantic index flag on search
plaesy search stats       # stats subcommand on search
```

**After (pattern):**
```
plaesy doctor             # global health check (reach + search)
plaesy stats              # global statistics (search index + reach cache)
plaesy search index       # document index (subcommand)
plaesy search query       # document search (subcommand)
plaesy reach list         # platform list (subcommand)
plaesy reach configure    # credentials (subcommand)
plaesy reach web "..."    # query (direct)
```

---

## Principle: Cross-Cutting Concerns → Global Commands

If a command spans multiple capabilities, promote it to top level.

| Concern | Was under | Now at |
|---------|-----------|--------|
| Health check | `reach doctor` | `plaesy doctor` |
| Statistics | `reach stats`, `search stats` | `plaesy stats` |

**Rationale:** User shouldn't need to know `doctor` lives under `reach` when it also checks search index. `stats` for search index and reach cache are related (both about system state).

---

## Principle: Domain-Specific → Keep Under Parent

Commands that only make sense within a domain stay as subcommands.

| Command | Parent | Reason |
|---------|--------|--------|
| `list` | `reach` | Only reach has "platforms" to list |
| `configure` | `reach` | Credentials are per-platform |
| `index` | `search` | Only search builds document index |
| `query` | `search` | Only search queries document index |

---

## Principle: Legacy vs New — Preserve, Don't Break

Legacy workflows (`plaesy search --index` for code symbols) stay as flags on the parent. New workflows (`plaesy search index` for documents) become subcommands.

```go
// Parent command handles both
cmd := &cobra.Command{
    Use: "search [command] [query]",
    RunE: func(cmd, args) {
        if args[0] == "index" || args[0] == "query" {
            return cmd.Help() // subcommand takes over
        }
        return runLegacySearch(cmd, args) // flags --index, --top
    },
}
cmd.AddCommand(newSearchIndexCmd(), newSearchQueryCmd())
```

---

## Principle: Unified Config — Single Source of Truth

All capabilities read/write `.plaesy/config.json` via `ConfigManager[T]`.

- Generic: `ConfigManager[ReachConfig]`, `ConfigManager[SearchConfig]`
- Hot-reload: fsnotify watches file, notifies watchers
- Atomic writes: temp file + rename
- Keyring: zalando/go-keyring for credentials
- Migration: `MigrateFrom(legacyPath, converter)` for legacy→unified

No capability owns its own config file. Migration is one-way (legacy → unified) with bidirectional conversion functions for backwards compatibility.

---

## Pattern: Windows-Specific Stubs

ONNX Runtime not available on Windows via `onnxruntime-go`. Solution:

```go
// onnx.go (build tag !windows)
package embedder
func NewONNXEmbedder(cfg Config) (*ONNXEmbedder, error) { ... }

// onnx_windows.go (build tag windows)
package embedder
func NewONNXEmbedder(cfg Config) (*ONNXEmbedder, error) {
    return nil, fmt.Errorf("ONNX unavailable on Windows; use HTTP embedder with local TEI/Ollama")
}
type ONNXEmbedder struct{} // stub implementing Embedder interface
```

User gets actionable error, not silent CGO failure.

---

## Checklist for New Top-Level Command

- [ ] Distinct verb (not a flag on existing command)
- [ ] Cross-cutting? → top level. Domain-specific? → subcommand.
- [ ] Help text references related commands (`Use "plaesy stats" for...`)
- [ ] Reads/writes unified config via `ConfigManager[T]`
- [ ] Windows build passes (no CGO-only deps without stubs)
- [ ] Unit tests for core logic; integration via CLI smoke test
- [ ] Added to `docs/reference.md` Command Index
- [ ] `docs/metadata.json` count updated
- [ ] CHANGELOG entry

---

## Anti-Patterns to Avoid

1. **Flag proliferation on builder commands** — `plaesy graph --semantic-*` was hard to discover
2. **Duplicate subcommands** — `reach stats` and `search stats` both showed system state
3. **Hidden cross-cutting commands** — user didn't know `reach doctor` checked search index
4. **Separate config files** — `.plaesy/reach/config.json`, `.plaesy/search/config.json` → migration hell
5. **CGO without Windows fallback** — build breaks silently on Windows

---

## Related Decisions

- [Search CLI Restructure](decisions/search-cli-restructure.md) — this restructure
- [Canonical Platform IDs](decisions/canonical-platform-ids-claude-kilo.md) — verb naming consistency
- [Unified Config Migration](decisions/unified-config-migration.md) — single config file

---

## Related Memory Topics

- [Semantic Search Feature](memory/semantic-search-feature.md) — original search implementation patterns
- [Testing Discipline](memory/testing-discipline.md) — `[build failed]` means vet blocked tests
- [Instructions Maintenance](memory/instructions-maintenance.md) — reload targets cwd, not repo root