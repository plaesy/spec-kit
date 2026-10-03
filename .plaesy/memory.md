---
title: "Memory & Knowledge Index"
description: "Central index for project memory, guidance, and reference"
updatedAt: "2026-10-03T19:58:00.000Z"
---

# Memory Index & Navigation

**Project:** plaesy/spec-kit — the config-driven spec-kit fork
**Last Updated:** 2026-10-03
**Purpose:** Navigation surface for durable project knowledge.

This file is an index only. Detail lives in `.plaesy/memory/[topic].md`.

---

## 📋 Guidance & Feedback

- [Prompt Authoring](memory/prompt-authoring.md) — source vs generated, dimension rules, routing, reload pruning
- [Quality Ratchets](memory/quality-ratchets.md) — markdown 0, prompt-craft floors, instrument bugs, linter defects
- [Testing Discipline](memory/testing-discipline.md) — red suites, mutation testing, `[build failed]` means 0 tests ran, new-command bookkeeping
- [Git and Session State](memory/git-and-state.md) — tracked vs generated `.plaesy/`, platform ids, publishing
- [Cross-Platform CI](memory/cross-platform-ci.md) — Getwd-after-Chdir traps, CRLF regex breaks, GOOS-gated tests, single-OS coverage ratchet
- [Instructions Maintenance](memory/instructions-maintenance.md) — .instructions.md→.md suffix trap, fake 5a./5b. ordered markers, maintainer-only steps leaking into prompts, reload-after-edit, go test\|tail hides exit code
- [Semantic Search Feature](memory/semantic-search-feature.md) — go.mod bump, instructions resync, CLI verbs, vet failures disable tests, parallel agents in one package

---

## ⚡ Quick Lookup

**"Which file do I edit, `prompts/` or `.kilo/commands/`?"** → `prompts/`; the copy is generated.
**"What is the markdown ceiling now?"** → Zero. The baseline file is gone on purpose.
**"How do I know a prompt edit did not break the corpus?"** → `./scripts/plaesy.exe validate markdown` and `go test ./internal/quality`.
**"A test is failing and I did not touch it — is it mine?"** → It may be. Read it before classifying it.
**"A package says `[build failed]` — did its tests run?"** → **No.** `go test` runs vet first, so vet errors suppress every test in the package. Re-run with `-vet=off`.
**"I added a `plaesy` command — am I done?"** → Three more edits: CHANGELOG line, `docs/reference.md` Command Index row, `docs/metadata.json` counts.
**"Which platform id do I use?"** → `claude`/`kilo` (not `claude_code`/`kilo_code`); see [the rename decision](decisions/canonical-platform-ids-claude-kilo.md).
**"Can I run the full test suite?"** → Not without explicit permission — except in `/fix` and `/assess`, whose exit criteria require it.
**"A test fails on one OS but not the others — is it the code or the test?"** → Often the test. See [Cross-Platform CI](memory/cross-platform-ci.md).
**"Did I lose a task file from git?"** → Check the `!.plaesy/` ordering in `.gitignore`.
**"Is there a stale command in the mirror?"** → `reload --prune` reports it; `--prune-apply` deletes it.
**"Can I delete a renamed prompt's old copy by hand?"** → Stopgap only; teach the tool to prune.

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

**Status:** Index rebuilt 2026-09-30; cross-platform-ci topic added 2026-10-02 after getting CI green for the first time since September; instructions-maintenance topic added 2026-10-03 after an instructions/ cleanup pass; semantic-search-feature topic extended 2026-10-03 after the doc-comment and staleness fan-out — now also covering parallel agents in one Go package, per-agent identifier prefixing, and verifying agent claims against disk. 2026-10-03 later the same day, after `/assess` scored 0 on a blocking gate: testing-discipline gained how to read a `go test ./...` result (`[build failed]` means zero tests ran) and the three bookkeeping edits a new top-level command owes; the semantic-search-feature vet note was corrected — its `git status` evidence did not establish pre-existence, and the breakage is not harmless because it disables two packages' tests.
**Next:** Add project-specific topics as they arise; every link above must resolve.
