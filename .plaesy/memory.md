---
title: "Memory & Knowledge Index"
description: "Central index for project memory, guidance, and reference"
updatedAt: "2026-09-25T02:50:00Z"
---

# Memory Index & Navigation

**Project:** Plaesy Constitution Kit (framework repo, dogfooding its own `plaesy init`) —
renamed from "Plaesy Spec-Kit" 2026-09-23, README scope only so far; see
[[assess-marketing-2026-09-23]] decision log for remaining rename scope (docs/
CHANGELOG/repo-rename still undecided). CLI is now a single Go binary — see
[[go-cli-migration-2026-09-23]]; the bash/PowerShell scripts referenced by
several notes below are deleted.
**Last Updated:** 2026-09-23 (Go CLI migration checkpoint; /continue session: design+marketing verified via Mode 3, README rebrand confirmed)
**Purpose:** Central index for project memory, guidance, and reference

---

## 🎯 Core Reference

- [Plaesy Spec-Kit Constitution](constitution.md) — Governing dimensions, quality bars, stack, and hard stops.
- [Install Scripts & Pre-Release (v0.0.2)](plaesy-install-scripts-v0.0.2.md) — Cross-platform install.sh/install.ps1, documentation audit, ready for release tag.
- [Go CLI Migration 2026-09-23](go-cli-migration-2026-09-23.md) — bash/PowerShell scripts replaced by a single Go binary (`scripts/cmd`, `scripts/internal`); supersedes several rules below that still reference `scripts/bash`/`scripts/powershell`.
- [Graft Assessment and Spec-Kit Architecture Gaps](graft-assessment-and-analyzer-gaps-2026-09-22.md) — Evidence, risks, and transfer candidates from the Graft assessment.
- [Technical Assessment 2026-09-22](assess-technical-2026-09-22.md) — Multi-dimensional technical assessment: code quality, tests, security, performance, CI, analyzer gaps.
- [Design Assessment 2026-09-22](assess-design-2026-09-22.md) — Prompt/instruction/template quality, orphaned files, CLI ergonomics, docs UX.
- [Business/Product Assessment 2026-09-22](assess-business-product-2026-09-22.md) — Viability, completeness, marketing readiness, roadmap, adoption signals.
- [Marketing Assessment 2026-09-23](assess-marketing-2026-09-23.md) — CRITICAL: name collision with github/spec-kit (138k★), no stated wedge; 53/100 Grade D.
- [Analyzer Implementation Fixes 2026-09-22](analyzer-implementation-fixes-2026-09-22.md) — `--if-changed` fast path, portable stat, PowerShell parity, functional tests, CI wiring.
- [Plaesy Analyze Overview Refactor](analysis-overview-md-refactor-2026-09-16.md) — Analyzer writes only analysis/overview.md and never context.md/memory.md.
- [Assess Dimension Coverage](assess-dimension-coverage.md) — Financial/marketing assess coverage and mapping registration fix.
- [Command Surface and Architecture](command-surface.md) — Prompt phases, removed commands, constitution, verifier, and task model.
- [Performance Measurement Feedback](feedback_measure_perf_multiple_rounds.md) — Use multiple alternating warm-cache rounds before reporting speedups.
- [Feedback: Verify Before Deleting References](feedback-verify-before-deleting-refs.md) — Search for another producer before deleting a referenced capability.
- [Full Repository Audit](full-repo-audit-2026-09-15.md) — 141-agent audit, confirmed findings, fixes, and deliberate gaps.
- [Generate Images and Static-Site Gap](generate-images-and-static-site-2026-09-18.md) — Image command, API scripts, and static-site detection limitation.
- [How to Write a Slash-Command Prompt](how-to-create-prompt.md) — Prompt shape, routing rules, and anti-patterns.
- [mapping.json Extension Audit](mapping-json-extensions-audit-2026-09-16.md) — Generic extension/filename detection and orphan audit.
- [Microsoft 365 Automation](office-automation.md) — Office instructions, templates, validation, and install wiring.
- [Overview File-Count Fix](overview-md-file-counts-fix-2026-09-16.md) — File-count fix and source-versus-generated script rule.
- [Detect-Stack Performance Feedback](perf_detect_stack_extension_scan_2026_09_16.md) — Cache-aware performance measurement evidence.
- [Plaesy Analyze Path Fix](plaesy-analyze-path-fix-2026-09-16.md) — Root path, chatmodes, and analyzer output-location fixes.
- [Cross-Dimensional Prompt Routing](prompt-dimension-generalization-2026-09-16.md) — Dimension routing and post-install path consistency.
- [Prompt Quality and design.md](prompt-quality-and-design-md.md) — Prompt fixes and the design-token source of truth.
- [Stack Detection Fixes](stack-detection-fixes-2026-09-17.md) — Office, language, framework, and Next.js detection fixes.

---

## 📋 Guidance & Feedback

### Established Rules (Follow These)

- Project-specific facts, decisions, and behavioral feedback for THIS repo belong here (`.plaesy/memory/`, `.plaesy/context.md`), not in an AI tool's own cross-session memory — so they travel with the repo on any machine/tool. Reserve a tool's own memory system for things that should follow the *user* across unrelated projects (their role, general collaboration preferences), never for spec-kit's own project knowledge. See [[feedback-ps1-bash-parity]] and [[feedback-no-graphify-naming]] as examples of rules migrated in from Claude's cross-session memory on 2026-09-19.
- Task tracking uses only `.plaesy/tasks/{status}/*.md` — never reintroduce a flat `backlog.md`
- Instructions install to `.plaesy/instructions/`, not `.plaesy/memory/` — `.plaesy/memory/` is for constitution.md, project memory.md, and /doc outputs only
- `mapping.json` never gets copied into a project's `.plaesy/` — it's a spec-kit-source-only registry. Technology detection for a per-project session must go through `plaesy detect-stack` (Go binary, `scripts/internal/detectstack`), never by re-reading/re-parsing `mapping.json` inline in a prompt
- Filename transform on install: `<name>.instructions.md` (spec-kit source) → `.plaesy/instructions/<name>.md` (per-project) — strip `.instructions` before checking whether a project already has a given instruction
- When a prompt needs to reference another prompt's behavior, use the command name (`/assess`, `/start`) — never the source file path (`prompts/assess.md`). Only `instructions/`, `chatmodes/`, `templates/`, `checklists/` get a stable post-install mirror under `.plaesy/`; `prompts/` maps to a different platform-specific folder per AI tool and has no such mirror. See [[prompt-dimension-generalization-2026-09-16]]
- Every execution prompt (`start`/`continue`/`loop`/`implement`/`optimize`/`doc`/`fix`) must read the active dimension(s) from `.plaesy/memory/constitution.md` rather than assuming `software` — `/assess` was already cross-dimensional, the others weren't until 2026-09-16. See [[prompt-dimension-generalization-2026-09-16]]
- `context.md` and `memory.md` (the top-level session/knowledge files) live at `.plaesy/context.md` / `.plaesy/memory.md` (repo root) — NOT inside `.plaesy/memory/`. Only individual topic files (`[topic].md`, this file included) go in the `.plaesy/memory/` subfolder. `core_directories` in `platform.json` must never list `"chatmodes"` (chatmode files install to `.plaesy/roles/` only; `mapping.chatmodes` is a separate, still-needed key). See [[plaesy-analyze-path-fix-2026-09-16]]
- `plaesy analyze` never writes to `context.md`/`memory.md` (not even appending) — its only auto-generated output is `.plaesy/analysis/overview.md`, fully replaced every run. See [[analysis-overview-md-refactor-2026-09-16]]
- `agents.instructions.md` must never appear in `mapping.json` (`always_load` or any category) — it's the platform `core` template (`platform.json` → `mapping.core.value`), copied per-platform as CLAUDE.md/AGENTS.md/etc. by `plaesy-init`'s own loop, and already excluded from the generic instructions copy. Any new `instructions/*.instructions.md` file MUST be registered in `mapping.json` to auto-install — except this one. See [[mapping-json-extensions-audit-2026-09-16]]
- A `mapping.json` category entry can carry an `"extensions"` array for file-existence detection (e.g. `.docx`, `.tf`, `.sql`) — the Go `detectstack` package reads it generically; add extensions there instead of hardcoding extension checks in scripts. See [[mapping-json-extensions-audit-2026-09-16]]
- A `mapping.json` category entry can also carry a `"filenames"` array (exact-filename presence, e.g. `next.config.js`) for when the reliable signal is a config file's name rather than content, and the tech's own name is too generic/prose-colliding for a keyword. Never embed a literal `"` inside a `"keywords"` entry unless you have re-checked `detectstack.go`'s keyword extraction handles escaping the way you expect. The Go port (`scripts/internal/detectstack`) scans `"."` (not the `.plaesy` destination) when installing instructions. See [[stack-detection-fixes-2026-09-17]]
- A scoped prompt that needs to actually execute something (not just reason in text) shells out to `plaesy <command>` (Go binary) — never fabricates a result path/output; a missing precondition (API key, tool on PATH) means stop and say exactly what's missing, never degrade silently or pretend success. See [[generate-images-and-static-site-2026-09-18]]
- Plain HTML sites with no generator config file (`astro.config.*`, `_config.yml`, etc.) are NOT auto-detected by `static-site` in `mapping.json` — `.html` alone is too broad/collision-prone a signal; deliberately left as a known limitation rather than a false-positive-prone heuristic. See [[generate-images-and-static-site-2026-09-18]]
- `plaesy analyze`'s fingerprint fast-path (skip regeneration when nothing changed) is the DEFAULT — `--force` is the only way to force full regeneration; `--if-changed` is kept as an accepted no-op flag for backward compatibility only. Any prompt citing `.plaesy/analysis/overview.md` as evidence (`/continue`, `/assess`) must call `plaesy analyze` first — it's cheap when unchanged, so call unconditionally rather than trying to detect staleness yourself. See [[graft-assessment-and-analyzer-gaps-2026-09-22]]
- Only human-reviewable analyzer output (`overview.md`, `project.json`, `project.structure.json`) is git-tracked — regenerated cache artifacts (`.plaesy/analysis/.edges.tsv`, `.nodes.tsv`, `.analysis-fingerprint`) are gitignored, never committed. See [[graft-assessment-and-analyzer-gaps-2026-09-22]]

---

## 📌 Reference Information

**Platform-Agnostic Paths (After plaesy init):**
- Instructions → `.plaesy/instructions/[name].md` (flat structure)
- Chatmodes → `.plaesy/roles/[name].md` (flat structure, platform-agnostic)
- Configs → `.plaesy/scripts/configs/` (trim-rules, platform.json — copied from `scripts/configs/`)
- Analysis → `.plaesy/analysis/` (written by `plaesy analyze`; NOT `.plaesy/memory/analysis/` — fixed 2026-09-15, see [[feedback-verify-before-deleting-refs]])
- Tasks → `.plaesy/tasks/{backlog,todo,doing,done,blocked}/`

---

## ⚡ Quick Lookup

**"How do I reference an instruction file?"**  
→ Use `.plaesy/instructions/[name].md` (flat, no subfolders)

**"Where are chatmodes stored?"**  
→ `.plaesy/roles/[name].md` (platform-agnostic, like instructions in .plaesy/instructions/)

---

**Status:** Memory index synchronized 2026-09-23; every indexed topic link resolves to a local file.
**Next:** Resume from `.plaesy/context.md`; naming/rename scope decision still open (README done, docs/CHANGELOG/repo-rename pending user decision) — ask before broadening scope.
