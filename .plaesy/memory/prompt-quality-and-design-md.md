---
title: "Prompt Quality Fixes + design.md System"
updatedAt: "2026-09-15T00:00:00.000Z"
---

# Prompt Quality Fixes + design.md System

## Part 1: prompts/*.md deep review (fork + live web research)
Reviewed all 9 files in `prompts/`. Sources: Anthropic "Building Effective Agents"
(anthropic.com/engineering/building-effective-agents), GitHub `github/spec-kit`
(category comparator), 2026 Claude Code slash-command convention writeups.

Confirmed bugs found and fixed:
- `continue.md`: "Project States" numbered list skipped from 6 to 8 (no #7) — fixed,
  renumbered 1-9, and renamed the item to "Quality Issues Persist After Fix" to
  disambiguate it from the similar #6.
- `fix.md`: had **two different "Completion Format" sections** (line 74 and 168)
  with inconsistent fields — merged into one canonical block (Root cause, Fix,
  Tests, Prevention, Confidence, Modified Files, Tests Added, Documentation).
- `doc.md`: Output Contracts table had an empty orphaned row (`| design.md | |`,
  line 169) — confirmed no `### docs/design.md` section exists anywhere in the
  file (unlike overview.md/architecture.md/etc. which each have one), so the row
  was removed rather than inventing content for a doc that isn't actually
  produced by `/doc`.
- `loop.md`: `--parallel 4 (default: 8)` — "8" was unsupported/fabricated anywhere
  else in the file; changed to `(default: sequential)`.

Grades given (informational, not acted on further): loop.md A, doc.md/implement.md/
save.md/start.md A-, optimize.md/assess.md B+, continue.md B, fix.md C+ (pre-fix).

Cross-file duplication check: 4 of 5 files with "Autonomous Routing" sections
(fix, doc, implement, optimize) already had a "See also" cross-reference to
`.plaesy/instructions/dimension-mapping.md`; only `assess.md` was missing it —
added. Did NOT delete the per-command routing logic itself in any file — each
command's post-completion routing is legitimately different (state after
`/optimize` ≠ state after `/fix`), so collapsing them would have risked breaking
mandatory execution logic for a stylistic win only.

## Part 2: `.plaesy/memory/design.md` — new design-system source of truth
User correctly identified a real gap: `/assess:design` (`assess-design.instructions.md`)
audited component/token/dark-mode consistency but had nothing to audit code
*against* — no persistent, project-wide file defining what the tokens should be.

Web research confirmed this is an emerging 2026 standard, not invented: **Google
Labs published the `DESIGN.md` spec on 2026-04-10** as part of Google Stitch —
YAML front matter (machine-readable tokens: colors, typography, spacing,
components) + Markdown body (human rationale/intent), with a
`npx @google/design.md lint` CI validator for WCAG contrast + structural rules.
Source: https://dev.to/aws-builders/agentsmd-skillmd-designmd-how-ai-instructions-split-into-three-layers-d0g
(also covers the emerging AGENTS.md/SKILL.md/DESIGN.md three-layer split).

Implemented, following that exact convention:
- `templates/design.template.md` — new template, output path
  `.plaesy/memory/design.md`. YAML front matter (colors light/dark, typography,
  spacing scale, radii, components) + Markdown body (Overview/intent, Token
  Rationale, Component Conventions, States & Edge Cases, Accessibility Notes,
  Dark Mode Notes). No mapping.json registration needed — `copy_templates()`
  copies the whole `templates/` dir flat, unconditionally, unlike instructions.
- `instructions/assess-design.instructions.md`: added **Step 0** — check
  `.plaesy/memory/design.md` exists; missing = a finding (route to
  `/assess:design` Design Spine Mode 1 to generate it); exists = it becomes the
  reference Steps 3 (tokens) and 4 (dark mode) audit code against, not a vague
  "are tokens used" heuristic. Also updated Success Criteria hard-stop line.
- `prompts/assess.md` Design Spine Handoff Spec row (UI/UX column, ~line 111):
  now explicitly names `.plaesy/memory/design.md` as the first handoff artifact.
- `prompts/implement.md` instruction-loading section: added step 5a — any
  UI/frontend task reads `.plaesy/memory/design.md` first and must reuse its
  tokens; if the file doesn't exist, flag it and recommend generating it via
  `/assess:design` rather than inventing tokens inline.

## Established rule for future sessions
`.plaesy/memory/design.md` is the project's single source of truth for design
tokens once a UI/frontend project exists. `/assess:design` (Mode 1) generates/
updates it, `/assess:design` (Mode 2) audits code against it, `/implement`
reads it before writing UI code. Don't let a new UI feature invent its own
color/spacing/typography values — trace them to this file or add to it via the
proper Design Spine flow, not inline.
