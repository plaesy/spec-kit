---
title: "Memory & Knowledge Index"
description: "Central index for project memory, guidance, and reference"
updatedAt: "2026-10-10T20:30:00.000Z"
---

# Memory Index & Navigation

**Project:** plaesy/spec-kit — the config-driven spec-kit fork
**Last Updated:** 2026-10-10
**Purpose:** Navigation surface for durable project knowledge.

This file is an index only. Detail lives in `.plaesy/memory/[topic].md`.

---

## 📋 Guidance & Feedback

- [Prompt Authoring](memory/prompt-authoring.md) — source vs generated, dimension rules, routing, reload pruning
- [Quality Ratchets](memory/quality-ratchets.md) — markdown 0, prompt-craft floors, instrument bugs, linter defects
- [Testing Discipline](memory/testing-discipline.md) — red suites, mutation testing, `[build failed]` means 0 tests ran, new-command bookkeeping, seams for network steps
- [Git and Session State](memory/git-and-state.md) — tracked vs generated `.plaesy/`, platform ids, publishing
- [Cross-Platform CI](memory/cross-platform-ci.md) — Getwd-after-Chdir traps, CRLF regex breaks, GOOS-gated tests, single-OS coverage ratchet
- [Instructions Maintenance](memory/instructions-maintenance.md) — .instructions.md→.md suffix trap, fake 5a./5b. ordered markers, maintainer-only steps leaking into prompts, reload-after-edit (reload targets the **current directory**), go test\|tail hides exit code
- [Semantic Search Feature](memory/semantic-search-feature.md) — go.mod bump, instructions resync, CLI verbs, vet failures disable tests, parallel agents in one package
- [CLI Restructure Patterns](memory/cli-restructure-patterns.md) — top-level verbs, subcommand consolidation, global vs scoped commands, config unification
- [Session 2026-10-10](memory/session-2026-10-10.md) — constitution, Search F2 eval harness, coverage gaps, HTTP fallback patterns

---

## ⚡ Quick Lookup

**"Which file do I edit, `prompts/` or `.kilo/commands/`?"** → `prompts/`; the copy is generated.
**"What is the markdown ceiling now?"** → Zero. The baseline file is gone on purpose.
**"How do I know a prompt edit did not break the corpus?"** → `./scripts/plaesy.exe validate markdown` and `go test ./internal/quality`.
**"A test is failing and I did not touch it — is it mine?"** → It may be. Read it before classifying it.
**"Do I have to run `plaesy search --index` after editing code?"** → No — `plaesy analyze` refreshes the index. `--index` rebuilds it alone.
**"A package says `[build failed]` — did its tests run?"** → **No.** `go test` runs vet first, so vet errors suppress every test in the package. Re-run with `-vet=off`.
**"I added a `plaesy` command — am I done?"** → Four more edits: CHANGELOG line, `docs/reference.md` Command Index row, `docs/metadata.json` counts, and the two CLI blocks in `README.md` (`TestReadmeCommandListMatchesTheReferenceIndex` catches the last one — it diffs README's bash/PowerShell blocks against the reference index, not just the reference against the command tree).
**"Does `make assets` make a prompt edit live?"** → No — it only syncs the go:embed copy (`scripts/internal/assets/data/`). A `prompts/*.md` edit also needs `plaesy reload --ai <platform>` per installed platform (check for `.claude/`, `.kilo/`, etc. at repo root) or it stays invisible to the runtime; `TestPromptMirrorMatchesSource` is the tripwire.
**"Which platform id do I use?"** → `claude`/`kilo` (not `claude_code`/`kilo_code`); see [the rename decision](decisions/canonical-platform-ids-claude-kilo.md).
**"Can I run the full test suite?"** → Not without explicit permission — except in `/fix` and `/assess`, whose exit criteria require it.
**"A test fails on one OS but not the others — is it the code or the test?"** → Often the test. See [Cross-Platform CI](memory/cross-platform-ci.md).
**"Did I lose a task file from git?"** → Check the `!.plaesy/` ordering in `.gitignore`.
**"Is there a stale command in the mirror?"** → `reload --prune` reports it; `--prune-apply` deletes it.
**"Can I delete a renamed prompt's old copy by hand?"** → Stopgap only; teach the tool to prune.
**"When should a new verb be a top-level command?"** → When it's a distinct user-facing action (`doctor`, `stats`, `search`), not a flag on an existing command. See [CLI Restructure Patterns](memory/cli-restructure-patterns.md).
**"How do I consolidate subcommands under a global command?"** → Move cross-cutting concerns (`doctor`, `stats`) to top level; keep domain-specific under their parent (`reach list`, `search index`).

---

## 🧑‍💼 Role Memory

Per-role history, one flat file per role under `memory/roles/[role].md` — the
one documented exception to "flat, no subfolders" (see
`.plaesy/instructions/plaesy.md` → Memory Hierarchy → Role Memory, and
`.plaesy/instructions/role-mapping.md`). Empty until a role's first
invocation writes its own file; add a link here (same
`- [Title](memory/roles/[role].md) — one-line description` pattern) only once
that file exists.

---

## 📌 Reference Information

**Platform-Agnostic Paths (after `plaesy init`):** instructions →
`.plaesy/instructions/[name].md`; agents → `.plaesy/roles/[name].md`;
scripts → `.plaesy/scripts/{powershell,bash}/`; configs →
`.plaesy/scripts/configs/`; analysis → `.plaesy/analysis/`; tasks →
`.plaesy/tasks/{backlog,todo,doing,done,blocked}/`; knowledge →
`.plaesy/memory/[topic].md`. Full table in `memory/git-and-state.md`.

---

## 📌 Configuration Reference

**Unified Config:** `.plaesy/config.json` — single file for all capabilities
- `reach` — 16 platforms, cache, rate limits, timeouts, retry
- `search` — embedder, vector store, text index, chunker, indexer, retriever, reranker
- `analyze` — auto-index, workers
- `graph` — include/exclude patterns, max file size

**ConfigManager[T]:** Generic, hot-reload (fsnotify), atomic writes, keyring for secrets, migration support.

**Reach Platforms (16):** web, youtube, github, rss, twitter, reddit, bilibili, xiaohongshu, linkedin, exa, v2ex, xueqiu, xiaoyuzhou, bosszhipin, facebook, instagram

---

**Status:** Index rebuilt 2026-10-10 with session-2026-10-10 topic. Constitution ratified (technical + internet_access dimensions). Search F2 hybrid eval harness completed with 20 tests. Coverage gaps identified: 7/12 packages below 90% constitutional threshold. Bleve upgraded to v2.6.1 fixing geo plugin build. Search config loading bug fixed (UnifiedConfig extraction). Next: coverage remediation per Constitution Rule 3, HTTP fallback adapters for 7 CLI-only platforms.

**Next:** Add project-specific topics as they arise; every link above must resolve.