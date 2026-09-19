---
title: "Memory & Knowledge Index"
description: "Central index for project memory, guidance, and reference"
updatedAt: "2006-01-02T15:04:05.999Z"
---

# Memory Index & Navigation

**Project:** Plaesy Spec-Kit (framework repo, dogfooding its own `plaesy init`)  
**Last Updated:** 2026-09-18 (/generate:images + static-site gap)  
**Purpose:** Central index for project memory, guidance, and reference

---

## 🎯 Core Reference

- [design.md Spec Alignment](memory/designmd-spec-alignment-2026-09-19.md) — Aligned `.plaesy/memory/design.md`'s template with the real Google Labs DESIGN.md spec (fields, 8-section order), not a blog summary; new `instructions/how-to-create-designmd.instructions.md` (2026-09-19)
- [Plaesy Graph Tool](memory/project-plaesy-graph-tool.md) — Custom knowledge-graph builder (PS1/bash parity), feature set, known limitations (migrated from Claude cross-session memory 2026-09-19)
- [New /generate:images Command + Static-Site Framework Gap](memory/generate-images-and-static-site-2026-09-18.md) — new `prompts/generate.md` router + `prompts/generate/images.md` scoped protocol, backed by real API-calling scripts (`scripts/{bash,powershell}/generate-image.*`, OpenAI `gpt-image-1`/Gemini, requires API key env var, fails loudly if unset); new `instructions/static-site.instructions.md` (plain HTML/CSS/JS + Astro/Eleventy/Hugo/Jekyll) registered in `mapping.json` + `/improve` Artifact Type table; plain-HTML-with-no-config sites still not auto-detected (deliberate, documented limitation) (2026-09-18)
- [Stack Detection Fixes (Office/Go/Java/Rust/Dart/Next.js)](memory/stack-detection-fixes-2026-09-17.md) — root cause of the reported Word/Excel/PowerPoint install bug: bash `copy_instructions()` scanned the empty `.plaesy` dir instead of the project root, plus a `jq`-gating bug that skipped detection entirely when jq wasn't installed; full mapping.json audit found Go/Java/Rust/Dart/Next.js also undetected on realistic manifests; new `"filenames"` field added to both detect-stack scripts for Next.js; 2 dead-code blocks removed (`--all-instructions` was a no-op); PowerShell `HashSet` → `Get-ChildItem -Include` gotcha documented (2026-09-17)
- [Colon-Scope Sub-Commands (/assess:design etc.)](memory/colon-scope-commands-2026-09-17.md) — `:scope` syntax (`/assess:design`, `/fix:technical`, …) was usage-text only, no backing file anywhere; added 58 wrapper files under `prompts/{assess,fix,implement,optimize,loop,improve}/`; caught + fixed a source-path cross-reference mistake before saving (2026-09-17)
- [mapping.json Extension-Based Detection + Orphan Audit](memory/mapping-json-extensions-audit-2026-09-16.md) — new generic `extensions` field (file-existence detection) for word/excel/powerpoint/csharp/terraform/sql; fixed bash's missing `*.csproj` scan and 1 real orphan file; **`agents.instructions.md` must never be added to mapping.json** (it's the platform core template, handled separately) (2026-09-16)
- [analyze: overview.md Replaces context.md/memory.md Appending](memory/analysis-overview-md-refactor-2026-09-16.md) — `plaesy analyze` now only writes `.plaesy/analysis/overview.md` (always replaced); `context.md`/`memory.md` are no longer touched by analyze at all (2026-09-16)
- [memory.md File Counts Hardcoded to 0](memory/overview-md-file-counts-fix-2026-09-16.md) — `generate_overview_md`/`New-OverviewMd` never counted files (unlike `generate_project_json`); fixed in `scripts/` source. **Rule**: always edit `scripts/{bash,powershell}/*`, never `.plaesy/scripts/**` (gitignored `plaesy init` output, edits there are lost) (2026-09-16)
- [plaesy-analyze Path Fix + chatmodes core_directories Bug](memory/plaesy-analyze-path-fix-2026-09-16.md) — `.plaesy/chatmodes` empty-folder bug (core_directories vs mapping.chatmodes confusion) fixed; plaesy-analyze.sh/.ps1 were writing context.md/memory.md to `.plaesy/memory/` instead of `.plaesy/` root (breaking every other part of the framework that reads them); switched from skip-if-exists to append-dated-section on rerun (2026-09-16)
- [Cross-Dimensional Generalization + Path Consistency](memory/prompt-dimension-generalization-2026-09-16.md) — start/continue/loop/implement/optimize/doc/fix generalized beyond software-only; `:{dimension}` syntax standardized across optimize/fix/implement/loop; fixed `prompts/*.md` path references that break post-install (2026-09-16)
- [Command Surface & Architecture Decisions](memory/command-surface.md) — Current 9 prompts, removed commands (/clarify, /design, /research, /evolve, /flow) and where they went, constitution, verifier separation, task system, and the plaesy init bugs fixed 2026-09-15
- [Microsoft 365 Automation Instructions](memory/office-automation.md) — PowerPoint/Excel/Word instruction files, starter templates, validation scripts, mapping.json registration, and the `/implement` → `detect-stack.sh` wiring fix (2026-09-15)
- [Assess Dimension Coverage & mapping.json Install Bug](memory/assess-dimension-coverage.md) — New assess-financial/marketing instruction files, and the root-cause fix: 8 assess-* + 4 foundational files were never registered in mapping.json so never installed (2026-09-15)
- [Prompt Quality Fixes + design.md System](memory/prompt-quality-and-design-md.md) — 4 confirmed bugs fixed across prompts/*.md (live web research grounded), plus new `.plaesy/memory/design.md` design-token source-of-truth system following Google's 2026-04-10 DESIGN.md spec (2026-09-15)
- [Full Repo Audit (ultracode workflow)](memory/full-repo-audit-2026-09-15.md) — 141-agent workflow audit, 41 confirmed findings, ~20 fixed same session (detect-stack.sh/.ps1 false-positive matching, missing @nara chatmode, systemic docs/ broken links, orphaned checklists wired in, dead code removed, bash/PS1 parity gaps); 3 items flagged for user decision, not auto-fixed (2026-09-15)
- [Feedback: Verify Before Deleting Refs](memory/feedback-verify-before-deleting-refs.md) — a "doc says X, this script doesn't do X" finding needs a repo-wide grep for another producer before deleting X, not just fixing the one script checked (2026-09-15)
- [How to Write a Plaesy Slash-Command Prompt](memory/how-to-create-prompt.md) — required shape, "say routing logic once" rule + evidence, anti-patterns found/fixed across prompts/*.md this session; moved here from `instructions/` per the rule below (2026-09-16)

---

## 📋 Guidance & Feedback

### Established Rules (Follow These)

- Project-specific facts, decisions, and behavioral feedback for THIS repo belong here (`.plaesy/memory/`, `.plaesy/context.md`), not in an AI tool's own cross-session memory — so they travel with the repo on any machine/tool. Reserve a tool's own memory system for things that should follow the *user* across unrelated projects (their role, general collaboration preferences), never for spec-kit's own project knowledge. See [[feedback-ps1-bash-parity]] and [[feedback-no-graphify-naming]] as examples of rules migrated in from Claude's cross-session memory on 2026-09-19.
- Task tracking uses only `.plaesy/tasks/{status}/*.md` — never reintroduce a flat `backlog.md`
- Instructions install to `.plaesy/instructions/`, not `.plaesy/memory/` — `.plaesy/memory/` is for constitution.md, project memory.md, and /doc outputs only
- `mapping.json` never gets copied into a project's `.plaesy/` — it's a spec-kit-source-only registry. Technology detection for a per-project session must go through `scripts/bash/detect-stack.sh`/`.ps1` (which parses it), never by re-reading/re-parsing `mapping.json` inline in a prompt
- Filename transform on install: `<name>.instructions.md` (spec-kit source) → `.plaesy/instructions/<name>.md` (per-project) — strip `.instructions` before checking whether a project already has a given instruction
- When a prompt needs to reference another prompt's behavior, use the command name (`/assess`, `/start`) — never the source file path (`prompts/assess.md`). Only `instructions/`, `chatmodes/`, `templates/`, `checklists/` get a stable post-install mirror under `.plaesy/`; `prompts/` maps to a different platform-specific folder per AI tool and has no such mirror. See [[prompt-dimension-generalization-2026-09-16]]
- Every execution prompt (`start`/`continue`/`loop`/`implement`/`optimize`/`doc`/`fix`) must read the active dimension(s) from `.plaesy/memory/constitution.md` rather than assuming `software` — `/assess` was already cross-dimensional, the others weren't until 2026-09-16. See [[prompt-dimension-generalization-2026-09-16]]
- `context.md` and `memory.md` (the top-level session/knowledge files) live at `.plaesy/context.md` / `.plaesy/memory.md` (repo root) — NOT inside `.plaesy/memory/`. Only individual topic files (`[topic].md`, this file included) go in the `.plaesy/memory/` subfolder. `core_directories` in `platform.json` must never list `"chatmodes"` (chatmode files install to `.plaesy/roles/` only; `mapping.chatmodes` is a separate, still-needed key). See [[plaesy-analyze-path-fix-2026-09-16]]
- `plaesy analyze` never writes to `context.md`/`memory.md` (not even appending) — its only auto-generated output is `.plaesy/analysis/overview.md`, fully replaced every run. See [[analysis-overview-md-refactor-2026-09-16]]
- `agents.instructions.md` must never appear in `mapping.json` (`always_load` or any category) — it's the platform `core` template (`platform.json` → `mapping.core.value`), copied per-platform as CLAUDE.md/AGENTS.md/etc. by `plaesy-init`'s own loop, and already excluded from the generic instructions copy. Any new `instructions/*.instructions.md` file MUST be registered in `mapping.json` to auto-install — except this one. See [[mapping-json-extensions-audit-2026-09-16]]
- A `mapping.json` category entry can carry an `"extensions"` array for file-existence detection (e.g. `.docx`, `.tf`, `.sql`) — both `detect-stack.sh`/`.ps1` read it generically already; add extensions there instead of hardcoding extension checks in either script. See [[mapping-json-extensions-audit-2026-09-16]]
- A `mapping.json` category entry can also carry a `"filenames"` array (exact-filename presence, e.g. `next.config.js`) for when the reliable signal is a config file's name rather than content, and the tech's own name is too generic/prose-colliding for a keyword. Never embed a literal `"` inside a `"keywords"` entry — bash's naive quote-stripping mangles it silently. In `detect-stack.ps1`, never pass a `HashSet[string]` directly to `Get-ChildItem -Include` — pipe through `ForEach-Object { $_ }` to an array first or it silently matches nothing. In `plaesy-init.sh`'s `copy_instructions()`, always scan `"."`, never its own `$target_dir` param (that's the `.plaesy` destination). See [[stack-detection-fixes-2026-09-17]]
- A scoped prompt that needs to actually execute something (not just reason in text) shells out to a real script under `scripts/{bash,powershell}/` — never fabricates a result path/output; a missing precondition (API key, tool on PATH) means stop and say exactly what's missing, never degrade silently or pretend success. See [[generate-images-and-static-site-2026-09-18]]
- Plain HTML sites with no generator config file (`astro.config.*`, `_config.yml`, etc.) are NOT auto-detected by `static-site` in `mapping.json` — `.html` alone is too broad/collision-prone a signal; deliberately left as a known limitation rather than a false-positive-prone heuristic. See [[generate-images-and-static-site-2026-09-18]]

---

## 📌 Reference Information

**Platform-Agnostic Paths (After plaesy init):**
- Instructions → `.plaesy/instructions/[name].md` (flat structure)
- Chatmodes → `.plaesy/roles/[name].md` (flat structure, platform-agnostic)
- Scripts → `.plaesy/scripts/{powershell,bash}/`
- Configs → `.plaesy/scripts/configs/`
- Analysis → `.plaesy/analysis/` (written by `plaesy analyze`; NOT `.plaesy/memory/analysis/` — fixed 2026-09-15, see [[feedback-verify-before-deleting-refs]])
- Tasks → `.plaesy/tasks/{backlog,todo,doing,done,blocked}/`

---

## ⚡ Quick Lookup

**"How do I reference an instruction file?"**  
→ Use `.plaesy/instructions/[name].md` (flat, no subfolders)

**"Where are chatmodes stored?"**  
→ `.plaesy/roles/[name].md` (platform-agnostic, like instructions in .plaesy/instructions/)

---

**Status:** Memory index updated 2026-09-19 (design.md spec alignment + cross-session memory migrated in-repo)  
**Next:** Not yet committed (see context.md `## Next`) — four pending batches: colon-scope prompt wrappers, mapping.json/detect-stack/plaesy-init fixes, `/generate:images` + `static-site.instructions.md`, and this session's design.md spec alignment (5 files) + 4 new `.plaesy/memory/*.md` files. Re-run `plaesy init` on this repo to mirror `prompts/{fix,implement,optimize,loop,improve}/` into `.claude/commands/`. `/generate:images` still needs a real smoke test with an API key set.
