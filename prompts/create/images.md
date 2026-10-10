---
description: "Generate an actual UI/design image asset from a text prompt and save it into the project"
---

# `/create:images` command instructions

This is the **images**-scoped entry point into `/create`. It produces a real
saved file (PNG by default), not a description of one.

When the image depicts a UI (a screen mockup, a component, an app screenshot),
apply `.plaesy/instructions/ui-ux-design-principles.md` (hierarchy, feedback,
affordance) to what's drawn — the same bar as coded UI, just a different
medium.

## Usage Format

```bash

/create:images "<description>"
/create:images "<description>" --style flat-illustration --size 1024x1024 --out assets/images/empty-inbox.png
/create:images --for .plaesy/memory/design.md --component "empty-state/inbox"
```

| Flag | Default | Meaning |
|---|---|---|
| `--style` | inferred from `.plaesy/memory/design.md` / `.plaesy/instructions/brandkit.md` if present, else "clean flat UI illustration" | Art direction to layer onto the description |
| `--size` | `1024x1024` | Passed straight to the provider (use `1536x1024` / `1024x1536` for wide/tall UI assets) |
| `--out` | `assets/images/<slugified-description>.png` | Where the file is written, relative to project root |
| `--for` | — | Path to a design-context/spec file (e.g. `.plaesy/memory/design.md`) to pull component context from instead of a freeform description |
| `--component` | — | With `--for`, which component/section to generate the asset for |
| `--n` | `1` | Number of variants to generate (each gets a `-1`, `-2`, … suffix) |

## Protocol

### Step 0: Ensure Design Context Exists

- Check whether `.plaesy/memory/design.md` exists (project-wide design tokens +
  rationale, Google Labs DESIGN.md convention — YAML front matter + Markdown body;
  field list in `.plaesy/instructions/how-to-create-designmd.md`).
- **Missing** → seed the file from `.plaesy/templates/design.template.md`
  (a neutral, documented placeholder palette/tone), label it
  `ASSUMED — seeded design.md from template`, record it, and continue. Do
  **not** silently fall through to generic art direction as if project
  identity had been read — the seed must be visible and traceable, not quiet.
- **Exists** → use it as source of truth for palette, tone, and component
  conventions in Step 1.

### Step 1: Resolve the Brief

- If `--for`/`--component` given: read that file, extract the component's visual
  description/purpose. If missing entirely, derive a one-line brief from the
  component/file name and nearby context, label it `ASSUMED — <brief>`, and
  record it rather than guessing silently.
- Otherwise the CLI/programmatic argument *is* the brief.
- Layer in project visual identity when it exists, without being asked:
  - `.plaesy/memory/design.md` — palette, tone, component conventions
  - `.plaesy/instructions/brandkit.md` — brand identity (logo system, colors,
    typography) when the ask is brand/identity-shaped (logo, brand board); this
    skill is only installed when the project declares a brand kit, so if the file
    is missing, use the neutral placeholder palette (same one `design.md`
    seeds from) and label the output `ASSUMED — placeholder, not on-brand`
    rather than inventing brand assets — never substitute a generic
    illustration and silently call it on-brand
  - `.plaesy/roles/designer.md` / `.plaesy/roles/accessibility.md` — general
    design and a11y framing (contrast, never encode meaning by color alone)
- Compose the final generation prompt: `{brief} — {style}, {brand/tone cues},
  no embedded text unless explicitly requested, {size} composition`.

### Step 1a: Confirm Before Fabricating Identity-Bearing Content

If the brief asks for a logo/icon, an avatar or profile photo, or any image
meant to represent a real statistic or a real person — confirm the concept
with the user first rather than inventing one from assumption. When asking
isn't possible (rapid iteration, no user turn available), generate a clearly
labeled placeholder instead of a disguised final asset: a text-based mark
(product name in an appropriate typeface) or `[LOGO]` for a logo, an
initial-based or simple geometric avatar for a person, and skip the image
entirely (report `[REAL DATA]` in its place) rather than illustrating a
statistic that has no real source. Never report a placeholder as if it were
approved final art. (Convention: anti-slop R-23,
github.com/miqdadbadjuber/anti-slop, retrieved 2026-09-29.)

### Step 2: Resolve the Provider

Resolve the provider by the order in
[`/create` → Provider Resolution](../create.md#provider-resolution) — it is the
only copy, so do not restate it here.

Then confirm the matching API key env var is set (`OPENAI_API_KEY` for
`openai`, `GEMINI_API_KEY` for `gemini`). **If it is not set: stop, tell the
user exactly which env var to export and how, and offer the composed prompt
from Step 1 as a manual fallback they can paste into any image tool.** Do not
attempt the API call without a key — it will just fail noisily.

### Step 3: Generate

Run the `plaesy images create` command, never call the provider API ad hoc
inline — the command already handles encoding, sizing, and error surfacing:

```bash

plaesy images create --prompt "<composed prompt>" --provider openai --size 1024x1024 --out assets/images/empty-inbox.png
```

This is a single cross-platform command — no separate bash/PowerShell variant.

For `--n > 1`, invoke the command once per variant with a `-1`/`-2`/… suffix on
`--out`, not a single call — keeps each failure isolated and reportable.

### Step 4: Validate & Report

- Confirm the output file exists and is non-zero bytes (the script exits
  non-zero on failure — surface its stderr verbatim, don't paraphrase it).
- Write matching alt text (one sentence, describes function not just
  appearance — per `.plaesy/roles/accessibility.md`) next to the report; if the
  calling context is a component file, this is what goes in its `alt=`.
- Report: file path(s) written, the exact prompt used (so it's reproducible),
  and the alt text.

## Programmatic Invocation (Called By Other Prompts)

A prompt that needs an asset mid-run (`/implement:design`, `/improve:design`,
`/doc`) calls this protocol directly rather than re-describing generation:

```text

CALL /create:images
  brief: "<one-line visual description>"
  style: "<optional; omit to inherit project brand>"
  out: "<path the calling prompt will reference>"
RETURNS
  path: <written file path, or null if generation was skipped — see below>
  alt_text: <one sentence>
  prompt_used: <final composed prompt, for reproducibility>
```

If no provider key is configured, the call returns `path: null` and
`prompt_used` still populated — the calling prompt must treat this as "asset
pending, manual generation needed" and say so in its own output, never
silently drop the asset or fabricate a path that doesn't exist.

## What a Good Image Run Does

- **Report only paths the script actually wrote** — an invented path reaches the
  user as a broken link
- **Surface a provider failure once and stop** — a retry loop cannot recover a
  provider outage
- **Keep the OpenAI/Gemini API key out of the composed prompt, the logs, and the
  report**
- **Apply the full `.plaesy/instructions/brandkit.md` discipline** when the ask is
  brand-identity-shaped, rather than the generic illustration style path
- **Seed `.plaesy/memory/design.md` from the template at Step 0 when absent**,
  labeled `ASSUMED`, rather than falling through to generic art direction
  unlabeled

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
