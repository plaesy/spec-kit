---
description: 'Instructions for writing .plaesy/memory/design.md'
---

Create or update `.plaesy/memory/design.md` for this repository — the project's
single source of truth for design tokens and visual identity.

Follows the **DESIGN.md** convention published by Google Labs (`google-labs-code/design.md`,
spec version `alpha`, 2026-04). It's the visual-identity counterpart to `AGENTS.md`
(project context) and `SKILL.md` (capabilities): AGENTS.md tells an agent what the
project is, DESIGN.md tells it what the project should look like. Read
[the spec](https://github.com/google-labs-code/design.md/blob/main/docs/spec.md) for
the authoritative field list if this file and the spec ever disagree — trust the spec.

Tokens are normative values; prose provides context. When a specific rule isn't
defined, agents should follow the markdown rationale for high-level decisions and
fall back to the closest existing token rather than inventing a new one inline.

## File shape

Two parts, in one file:

1. **YAML front matter** (`---` delimited) — machine-readable tokens. This is what a
   linter or another agent parses without touching prose.
2. **Markdown body** — human-readable rationale: why these values, not just what they are.

### Front matter fields

| Field | Required | Notes |
| --- | --- | --- |
| `name` | yes | Design system / product name |
| `version` | no | Spec version this file targets (currently `"alpha"`) |
| `description` | no | One-line system description |
| `omitted` | no | Array of intentionally-skipped sections, each with an optional reason — use this instead of leaving a section silently empty |
| `colors` | no | `map<string, Color>` — hex, named, functional (`rgb()`/`hsl()`), or wide-gamut CSS color values |
| `typography` | no | `map<string, Typography>` — each entry may set `fontFamily`, `fontSize`, `fontWeight`, `lineHeight`, `letterSpacing`, `fontFeature`, `fontVariation` |
| `spacing` | no | `map<string, Dimension \| number>` — layout spacing units |
| `rounded` | no | `map<string, Dimension>` — corner radius values (not `radii` — match the spec's field name) |
| `components` | no | `map<string, map<string, value>>` — per-component property tokens |

Reference one token from another with `{path.to.token}` (e.g. a component's border
color set to `{colors.primary}`) instead of repeating the literal value — this is how
the linter detects broken references and how an agent should propagate a token change.

Unknown token names are accepted as long as the value is syntactically valid, and
unknown top-level sections are preserved without error — so it's safe to extend this
file with a project-specific token group, but don't invent a *replacement* name for an
existing one (e.g. don't rename `rounded` to `radii`) since that breaks the shared
convention other tools expect.

### Body sections, in this exact order

Duplicate headings are a spec violation — each of these eight `##` sections appears
at most once, in this order. Omit a section (via `omitted` in front matter) rather
than leaving it as an empty heading:

1. **Overview** — brand personality / emotional tone in a sentence or two, not a
   feature list (e.g. "Architectural minimalism meets journalistic gravitas")
2. **Colors** — palette descriptions and the semantic role of each color (not a
   restatement of the hex values already in front matter)
3. **Typography** — font strategy and hierarchy levels (when to use h1 vs h2 vs body)
4. **Layout** — grid model and spacing rhythm
5. **Elevation & Depth** — how visual hierarchy/z-order is communicated (shadows,
   borders, layering)
6. **Shapes** — corner-radius and general form language
7. **Components** — per-component-family styling guidance (states covered, source
   location) — this is conventions, not a full inventory; link to `docs/components.md`
   (from `/doc`) for the exhaustive list rather than duplicating it here
8. **Do's and Don'ts** — concrete guardrails, phrased as pairs an agent can pattern-match
   against generated code (e.g. "Do use `{spacing.md}` between form fields. Don't
   introduce a one-off pixel value.")

## How to investigate before writing

- Read any existing `.plaesy/memory/design.md` first — improve in place, don't
  rewrite blindly. Preserve verified tokens, reconcile stale ones against the
  actual code.
- Search the codebase for the *actual* colors, spacing, and font values in use
  (CSS variables, theme files, Tailwind config, styled-components theme) rather than
  inventing values — this file documents what the project does, or is deciding to do,
  not a generic starter palette.
- Check `.plaesy/roles/designer.md` and any Figma/Pencil source (via the `figma-use`
  or `pencil` MCP tools, if the user has design files) for the canonical source before
  guessing at brand color/type choices.
- If this is a brand-new project with no visual identity yet, this file is being
  *authored*, not extracted — ask the user for brand direction rather than filling in
  placeholder black-and-white defaults and calling it done.

## Writing rules

- Every token needs a value that's actually used or intentionally decided — no
  placeholder `"#000000"` rows left unedited from the template.
- Keep the **Token Rationale** table (if present in the project's template) only for
  values whose *why* isn't obvious from the name; delete rows that need no explanation.
- WCAG AA contrast is checked by converting colors to sRGB — don't hand-pick a
  light/dark pair without verifying contrast; flag it as a finding for
  `/assess:design` if you can't verify computationally.
- Accessibility and dark-mode notes belong under their own sub-guidance only if they
  go **beyond** generic WCAG AA (already enforced by
  `.plaesy/instructions/assess-design.md`) — don't restate the generic rule.

## Maintaining it over time

Treat every edit like a release note: when the project publishes a design-system
update (new tokens, changed component, deprecated pattern), update this file in the
same change, not as a follow-up. `/assess:design` (Design Spine, Mode 1) generates or
updates this file; `/assess:design` (Mode 2) audits code against it; any UI-affecting
`/implement` run reads it first. A change here is a design decision, not a typo fix —
route non-trivial changes through `/optimize:design` so the audit re-runs.
