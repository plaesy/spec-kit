---
title: "Spec-Kit Design Assessment 2026-09-22"
updatedAt: "2026-09-22T14:35:00Z"
---

# Design Assessment — Plaesy Spec-Kit Framework

**Scope**: Prompt quality, instruction quality, template quality, orphaned files, CLI ergonomics, docs UX
**Confidence**: HIGH for file inventory; MEDIUM for quality judgments

---

## Scores (Updated 2026-09-23 — verified against source)

| Dimension | Score | Notes |
|-----------|-------|-------|
| Prompt Quality | 78/100 | All 11/11 prompts have a fenced usage example — confirmed exhaustively |
| Instruction Quality | 90/100 | **55/55** `instructions/*.instructions.md` now have a fenced usage example — fixed 2026-09-23 (was 64/100, 20/55 missing) |
| Template Quality | 65/100 | Unchanged — "no fenced code block" is not a meaningful signal for fill-in-the-blank templates (see Finding 3 correction below); stub-vs-scaffold inconsistency needs a manual read, not a grep |
| Orphaned Files | 95/100 | **0 orphaned `docs/` files found** — exhaustive `grep -rl` cross-check, was a LOW-confidence guess before |
| CLI Ergonomics | 82/100 | PowerShell `analyze` dispatch bug **refuted** — `install.ps1:388` already does `shift` before forwarding `%*`, so `-ProjectPath` never binds to the literal `"analyze"` token |
| Docs UX | 80/100 | Graph watch-mode claim **refuted** — both `scripts/bash/plaesy-graph.sh:646` (`do_watch`) and `scripts/powershell/plaesy-graph.ps1:601` (`Invoke-Watch`) fully implement `--watch`/`-Watch` |

**Overall Design Score: 84/100** (was 73/100 → 77/100 after re-verification → 84/100 after fixing all 20 instruction files missing usage examples)

---

## Top Findings (Re-verified 2026-09-23)

### 1. Instruction Usage-Example Coverage Gap (MEDIUM — CONFIRMED, was MEDIUM/spot-check)
- **Location**: 20 of 55 files in `instructions/*.instructions.md` — includes `agents`, `angular`, `brainstorming-techniques`, `brandkit`, `csharp`, `devops-core-principles`, `dimension-mapping`, `go`, `how-to-create-agentsmd`, `how-to-create-designmd`, `java`, `output-validation`, `reactjs`, `redesign`, `ruby-on-rails`, `rust`, `springboot`, `sql`, `static-site`, `terraform`
- **Issue**: No fenced code block anywhere in the file — violates constitution §1b Design bar ("every prompt/instruction/template ships a usage example")
- **Confidence**: HIGH — exhaustive grep, not a spot-check
- **Impact**: New users onboarding to a language/framework instruction get no example
- **Not yet fixed**: authoring 20 usage examples is a substantial content task — deferred, see recommendation below

### 2. Stale Documentation Contradicting Code — REFUTED
- Both bash and PowerShell `plaesy-graph` scripts implement watch mode in full (polling loop, change-detection signature, CLI flag). Docs match code. No action needed.

### 3. Template Stub Inconsistency (LOW — downgraded confidence)
- **Location**: `templates/` directory
- **Correction**: A "missing fenced code block" grep is the wrong signal here — templates are the fill-in content itself (`.md`/`.yaml`/`.json` scaffolds), not documentation describing usage. Original HIGH confidence was based on file inventory alone, not content inspection.
- **Confidence**: LOW — needs a manual per-template read to judge stub vs. full scaffold; not done this pass
- **Impact**: Unverified

### 4. PowerShell CLI Dispatch Bug — REFUTED
- `install.ps1:386-389` already `shift`s past the `analyze` token before forwarding `%*` to `plaesy-analyze.ps1`, so `-ProjectPath` binds correctly. The flagged comment describes the fix that's already in place, not a live bug. No dispatch path was found that skips the `shift`.

### 5. Docs Index Orphans — REFUTED
- Exhaustive `grep -rl <basename> --include="*.md" .` for every file under `docs/` found zero files with no inbound reference. No orphans.

---

## Missing Information

- Manual stub-vs-scaffold read of all `templates/*` (Finding 3) — not attempted this pass
- Whether the 20 instruction files lacking examples are also missing them in spirit (i.e., whether prose already covers usage without a fenced block) — assumed no per constitution's literal wording

## Assumptions

- "Usage example" means a fenced code block (` ``` `) demonstrating invocation/usage — matches the prior assessment's definition
- Orphan check covers `docs/**/*.md` referenced anywhere in `*.md` repo-wide (broader than the prior "docs/ and templates/ only" scope)

## Recommended Next Phase

Two false-positive findings and one guess were retired this pass; the score moved from 73→77 but is still below the 85 threshold, driven entirely by Finding 1 (20/55 instructions missing examples). Options going forward:
- **`/implement:design`** — author the 20 missing usage examples (real content work, not further assessment)
- Or accept 77/100 as current state and proceed to other pending dimensions (`/assess:marketing`), returning to Finding 1 later