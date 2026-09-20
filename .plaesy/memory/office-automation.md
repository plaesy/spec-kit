---
title: "Microsoft 365 Automation Instructions"
description: "PowerPoint/Excel/Word instruction files, starter templates, validation scripts, and how they wire into detect-stack.sh/implement.md"
updatedAt: "2026-09-15T00:00:00.000Z"
---

# Microsoft 365 Automation Instructions

## What was added (2026-09-15)
- `instructions/powerpoint.instructions.md`, `instructions/excel.instructions.md`,
  `instructions/word.instructions.md` — each follows the existing
  `reactjs.instructions.md`-style format (frontmatter `description`+`applyTo`,
  Development Standards, Starter Template code block, Implementation Process).
- Library choices, based on web research (see chat for sources): `python-pptx`
  (PowerPoint), `openpyxl` for template-fill/read + `XlsxWriter` for large
  write-only jobs (Excel), `python-docx` + `docxtpl` for tag-based templates (Word).
- PowerPoint instruction additionally got: a `Dependencies` section (pinned
  version floor), a note on Microsoft Graph API as the alternative when the
  target lives in OneDrive/SharePoint rather than a local file, and a validation
  step.
- `scripts/bash/validate-pptx.sh` + `scripts/powershell/validate-pptx.ps1` —
  parity pair (bash/PS1 both, per the repo's established parity rule): always
  runs structural validation via `python-pptx`, additionally headless-renders
  via LibreOffice (`soffice`) when it's on PATH. Framed in the instruction as a
  direct one-off step to run right after generation, **not** as something that
  requires a CI pipeline — user explicitly said usage here is one-time.
- Registered in `instructions/mapping.json` under `mappings.frameworks`
  (`powerpoint`, `excel`, `word` keys) and in `docs/instructions/README.md`
  under a new "📊 Microsoft 365 Automation Instructions" table.
- **Update (2026-09-15, follow-up session)**: Excel and Word were brought to the
  same completion level as PowerPoint — both instruction files now have a
  `Dependencies` section, a Graph API note, and a "Testing and Validation" step
  that calls a dedicated validator. Added `scripts/bash/validate-xlsx.sh` +
  `.ps1` twin (openpyxl structural check + optional LibreOffice render) and
  `scripts/bash/validate-docx.sh` + `.ps1` twin (python-docx structural check +
  leftover-`{{ }}`-placeholder check + optional LibreOffice render), matching
  the `validate-pptx.*` pattern exactly (parity rule).
- Also fixed: `docs/instructions/README.md` had a dangling link to a
  non-existent `docs/ARCHITECTURE.md`. Removed the link; the content it was
  pointing at was already inline in the same file (under "📍 Architecture:
  Spec-Kit vs Per-Project"), so no new file was needed.
  `docs/prompts/AUTHORING.md` was checked and found to have no actual
  references anywhere — the earlier note flagging it was stale.

## Integration fix: how `/implement` discovers instructions
User asked whether prompts are actually wired to instructions, which surfaced
a real gap and then a real bug in the fix:
- `prompts/implement.md` originally hardcoded a per-technology instruction list
  (Go → go.instructions.md, React → reactjs.instructions.md, ...) that had no
  path to ever include powerpoint/excel/word without editing the prompt again.
- First fix attempt: told `/implement` to read `mapping.json` and parse it
  itself. **Wrong** — `mapping.json` lives only in the spec-kit source
  (`instructions/mapping.json`); it is never copied into a project's `.plaesy/`,
  so a per-project session can't read it, and re-implementing the parser inline
  in a prompt duplicates logic that already exists as a real script.
- Actual fix: `prompts/implement.md` Step 1 now tells the agent to run
  `scripts/bash/detect-stack.sh <dir>` / `.ps1` twin (the script that already
  parses `mapping.json` against the target project and prints matched
  `*.instructions.md` names), then apply the filename transform before checking
  `.plaesy/instructions/`.
- **Filename transform, confirmed from `plaesy-init.sh`'s `copy_instructions`
  (line ~107)**: `basename "$file" .instructions.md` → `$basename.md`. So
  `go.instructions.md` (source) becomes `.plaesy/instructions/go.md` (per-project).
  Any prompt or script reasoning about "does the project already have this
  instruction" must strip `.instructions` before checking `.plaesy/instructions/`.
- If `detect-stack.sh` reports an instruction not yet present under
  `.plaesy/instructions/` (tech added after the project's last `plaesy init`),
  `/implement` should copy it over (same rename) before loading it, not skip it.

## Why this matters going forward
Any new instruction file added to `instructions/` only needs a `mapping.json`
entry to become discoverable by `/implement` — no prompt file edits required,
because detection is centralized in `detect-stack.sh`/`.ps1`, not duplicated
per-prompt. If a future session adds another prompt that needs technology
detection, point it at `detect-stack.sh` too rather than re-deriving the logic.
