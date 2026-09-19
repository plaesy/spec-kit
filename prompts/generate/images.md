---
description: "Generate an actual UI/design image asset from a text prompt and save it into the project"
---

# `/generate:images` command instructions

This is the **images**-scoped entry point into `/generate`. It produces a real
saved file (PNG by default), not a description of one.

## Usage Format

```bash
/generate:images "<description>"
/generate:images "<description>" --style flat-illustration --size 1024x1024 --out assets/images/empty-inbox.png
/generate:images --for .plaesy/memory/design-spine.md --component "empty-state/inbox"
```

| Flag | Default | Meaning |
|---|---|---|
| `--style` | inferred from `.plaesy/memory/design-spine.md` / `instructions/brandkit.instructions.md` if present, else "clean flat UI illustration" | Art direction to layer onto the description |
| `--size` | `1024x1024` | Passed straight to the provider (use `1536x1024` / `1024x1536` for wide/tall UI assets) |
| `--out` | `assets/images/<slugified-description>.png` | Where the file is written, relative to project root |
| `--for` | — | Path to a design-spine/spec file to pull component context from instead of a freeform description |
| `--component` | — | With `--for`, which component/section to generate the asset for |
| `--n` | `1` | Number of variants to generate (each gets a `-1`, `-2`, … suffix) |

## Protocol

### Step 1: Resolve the Brief

- If `--for`/`--component` given: read that file, extract the component's visual
  description/purpose. If missing entirely, ask the user for a one-line brief
  rather than guessing.
- Otherwise the CLI/programmatic argument *is* the brief.
- Layer in project visual identity when it exists, without being asked:
  - `.plaesy/memory/design-spine.md` — palette, tone, component conventions
  - `instructions/brandkit.instructions.md` (or its `.plaesy/memory/` copy) —
    when the ask is brand/identity-shaped (logo, brand board), follow that
    skill's full art-direction discipline instead of a generic illustration
  - `.plaesy/roles/designer.md` / `.plaesy/roles/accessibility.md` — general
    design and a11y framing (contrast, never encode meaning by color alone)
- Compose the final generation prompt: `{brief} — {style}, {brand/tone cues},
  no embedded text unless explicitly requested, {size} composition`.

### Step 2: Resolve the Provider

Read the provider config, in this order, stop at first match:
1. Env var `PLAESY_IMAGE_PROVIDER` (`openai` | `gemini`)
2. `.plaesy/scripts/configs/image-provider.json` → `{"provider": "openai"}`
3. Default: `openai`

Then confirm the matching API key env var is set (`OPENAI_API_KEY` for
`openai`, `GEMINI_API_KEY` for `gemini`). **If it is not set: stop, tell the
user exactly which env var to export and how, and offer the composed prompt
from Step 1 as a manual fallback they can paste into any image tool.** Do not
attempt the API call without a key — it will just fail noisily.

### Step 3: Generate

Run the platform-appropriate script, never call the provider API ad hoc inline
— the scripts already handle encoding, sizing, and error surfacing:

```bash
# bash
scripts/bash/generate-image.sh --prompt "<composed prompt>" --provider openai --size 1024x1024 --out assets/images/empty-inbox.png
```

```powershell
# PowerShell
scripts/powershell/generate-image.ps1 -Prompt "<composed prompt>" -Provider openai -Size 1024x1024 -Out assets/images/empty-inbox.png
```

For `--n > 1`, invoke the script once per variant with a `-1`/`-2`/… suffix on
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

```
CALL /generate:images
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

## Anti-Patterns (NEVER Do These)

- ❌ Fabricate or hallucinate a file path without the script actually having
  written a file there
- ❌ Retry the provider call in a loop on failure — surface the error once, stop
- ❌ Embed the OpenAI/Gemini API key in the composed prompt, logs, or the report
  back to the user
- ❌ Generate brand/logo assets with the generic illustration style path instead
  of the full `brandkit.instructions.md` discipline when the ask is
  brand-identity-shaped

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
