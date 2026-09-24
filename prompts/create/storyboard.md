---
description: "Generate a visual storyboard (sequential narrative panels) from a user journey, interaction flow, or use case — produces real image assets, not descriptions"
---

# `/create:storyboard` command instructions

This is the **storyboard**-scoped entry point into `/create`. It produces real
saved image files (PNG by default), one per panel/scene, not descriptions.

## Usage Format

```bash
/create:storyboard "User signup flow: browse landing → enter email → verify → dashboard"
/create:storyboard "Mobile checkout: product review → add to cart → shipping → payment → confirmation"
/create:storyboard --for .plaesy/memory/design-spine.md --flow "user-onboarding" --panels 6
/create:storyboard --scenario "Error recovery when payment fails" --style flat-illustration --out assets/storyboards/payment-failure
```

| Flag | Default | Meaning |
|---|---|---|
| `--flow` / description | — | User journey, interaction flow, or use-case name (freeform text or reference to a named flow in design-spine.md) |
| `--panels` | `inferred from flow` (typically 4–8) | Number of sequential scenes to generate |
| `--scenario` | — | Alternative to description: a specific scenario/path within the flow (error case, edge case, happy path variant) |
| `--style` | Inferred from `.plaesy/memory/design-spine.md` / `instructions/brandkit.instructions.md`, else "clean flat UI illustration" | Art direction (sketch, wireframe, flat-illustration, realistic, etc.) |
| `--for` | — | Path to a design-spine/spec file to pull flow context from instead of freeform description |
| `--out` | `assets/storyboards/<slugified-flow>` | Base directory where panels are written (files named `panel-1.png`, `panel-2.png`, etc.) |
| `--alt-text` | auto-generated from panel description | Supply custom alt text per panel (delimited by `\n`); if omitted, auto-composed per panel |
| `--n` | `1` | Number of variants to generate (each variant set gets a `-variant-1`, `-variant-2` … directory) |

## Protocol

### Step 1: Resolve the Flow

- If `--for`/`--flow` given: read that file, extract the named flow's description/success metrics. If missing, ask for a one-line flow summary rather than guessing.
- If `--scenario` given: this is a variant within a larger flow (e.g. "error case" or "mobile variant"). Clarify the base flow first.
- Otherwise the description/argument *is* the flow.
- **Infer panel count** (if not `--panels` specified):
  - Happy path: 4–6 panels (minimal critical steps)
  - With error/alt paths: 6–8 panels
  - Complex flow: up to 10 panels (never exceed 12; split into multiple storyboards)
- Layer in project visual identity when it exists:
  - `.plaesy/memory/design-spine.md` — palette, tone, component conventions, UI patterns
  - `instructions/brandkit.instructions.md` (or its `.plaesy/memory/` copy) — brand identity (logos, colors, typography)
  - `.plaesy/roles/designer.md` / `.plaesy/roles/accessibility.md` — design principles, accessibility framing (readability, color contrast, never encode meaning by color alone)
- Compose a storyboard brief per panel, e.g.:
  ```
  Panel 1: User lands on homepage, sees hero + signup button
  Panel 2: Clicks "Sign Up" → enters email/password form
  Panel 3: Submits → sees verification email prompt
  Panel 4: Returns to app after email verification
  Panel 5: Dashboard appears with welcome message
  ```

### Step 2: Resolve the Provider

Read the provider config, in this order, stop at first match:
1. Env var `PLAESY_IMAGE_PROVIDER` (`openai` | `gemini`)
2. `.plaesy/scripts/configs/image-provider.json` → `{"provider": "openai"}`
3. Default: `openai`

Then confirm the matching API key env var is set (`OPENAI_API_KEY` for `openai`, `GEMINI_API_KEY` for `gemini`). **If it is not set: stop, tell the user exactly which env var to export and how, and offer the storyboard briefs from Step 1 as a manual fallback** (user can paste into any image tool to generate panels manually).

### Step 3: Generate Panel Images

For each panel, invoke `/create:images` (Programmatic Invocation pattern, see below) with:
- `brief`: the panel description from Step 1
- `style`: inherited from storyboard-level style
- `out`: `<base-out>/panel-<N>.png`

Run in **sequence, not parallel** — if any panel fails, stop and report which panel failed + the provider's error. Do not attempt retry loops; surface the error once and stop. The user can then fix the issue (API key, quota) and re-run the storyboard with the same command.

For `--n > 1`, generate each variant set in a separate directory (`-variant-1/`, `-variant-2/`, etc.), each with its own `panel-1.png`, `panel-2.png`, etc.

### Step 4: Generate Storyboard Index

Write a `.plaesy/memory/storyboards/{flow-name}.md` index file containing:
- Flow description (copied from Step 1 context)
- Success metrics (if from design-spine)
- Panel-by-panel breakdown:
  ```markdown
  ## Panel 1: Homepage
  ![Panel 1](../../assets/storyboards/flow-name/panel-1.png)
  *Alt text: [copied from panel generation]*
  
  User lands on homepage, sees value proposition and signup CTA.
  ```
- Links to each panel file
- Notes on what each panel demonstrates (happy path vs. error case, if applicable)
- If multiple variants generated, note which scenario each variant represents

### Step 5: Validate & Report

- Confirm all output files exist and are non-zero bytes (the `/create:images` invocation exits non-zero on any failure — surface its stderr verbatim)
- Report: 
  - Output directory and file count (e.g., "6 panels written to assets/storyboards/user-signup/")
  - Storyboard index file path (`.plaesy/memory/storyboards/{flow-name}.md`)
  - The composed panel briefs (for reproducibility)
  - Any alt-text warnings (e.g., if auto-generation fell back to generic descriptions)
- If generation was skipped due to missing API key, report the prompts as a manual fallback

## Programmatic Invocation (Called By Other Prompts)

A prompt that needs a storyboard mid-run (`/implement:design`, `/improve:design`, `/doc`) calls this protocol directly:

```
CALL /create:storyboard
  flow: "<flow name or description>"
  panels: <number of panels, or omit to auto-infer>
  style: "<optional; omit to inherit project brand>"
  scenario: "<optional; specific variant/path>"
  out: "<base directory the calling prompt will reference>"
RETURNS
  directory: <written directory path, or null if generation was skipped — see below>
  panel_files: [<path-1>, <path-2>, …]
  index_file: <path to .plaesy/memory/storyboards/{flow-name}.md>
  panel_briefs: <list of composed descriptions, for reproducibility>
```

If no provider key is configured, the call returns `directory: null` and `panel_briefs` still populated — the calling prompt must treat this as "storyboard pending, manual generation needed" and say so in its own output, never silently drop the storyboard or fabricate paths that don't exist.

## Anatomy of a Storyboard (Output Structure)

```
assets/storyboards/
├── {flow-name}/
│   ├── panel-1.png
│   ├── panel-2.png
│   ├── panel-3.png
│   └── …
├── {flow-name}-variant-1/
│   ├── panel-1.png
│   ├── panel-2.png
│   └── …
└── {flow-name}-variant-2/
    └── …

.plaesy/memory/storyboards/
├── {flow-name}.md
└── [index of all generated storyboards]
```

Update `.plaesy/memory/storyboards.md` to index all storyboards (same pattern as `.plaesy/memory/components.md`):
```markdown
# Storyboards

## User Signup Flow
- [Full flow](storyboards/user-signup.md) — 6-panel happy path + verification
- Scenarios: [Error: invalid email](storyboards/user-signup.md#error-invalid-email), [Resend verification](storyboards/user-signup.md#resend-verification)

## Payment Checkout
- [Full flow](storyboards/payment-checkout.md) — 7-panel payment + confirmation
- Scenarios: [Card declined](storyboards/payment-checkout.md#card-declined), [Expired session](storyboards/payment-checkout.md#expired-session)
```

## Anti-Patterns (NEVER Do These)

- ❌ Generate only one or two panels and call it a storyboard — a storyboard is a **sequence** (minimum 4 panels; maximum 12 per storyboard; split into multiple if needed)
- ❌ Fabricate or hallucinate file paths without `/create:images` actually having written them
- ❌ Retry the provider call in a loop on failure — surface the error once, stop
- ❌ Embed API keys in prompts, logs, or reports
- ❌ Generate all panels in parallel — run sequentially so a failure is clear and isolated to one panel
- ❌ Skip the storyboard index file — it must exist and link all panels, so `/implement:design`, `/doc`, and future updates can reference it
- ❌ Use generic/stock storyboard templates instead of layering project visual identity (design tokens, brand colors, UI patterns) into each panel
- ❌ Generate a storyboard for a use case that doesn't involve **visual narrative** — text-only processes belong in a runbook, not a storyboard (use `/doc` for that)

## Design-Spine Integration

If the project has a `.plaesy/memory/design-spine.md` with defined flows/scenarios:

1. **Before generating**, check for the flow name in the Design Spine (e.g., does it list "user-signup" as a defined flow?)
2. If found, extract its Success Metrics, UI patterns, and component conventions
3. Layer these constraints into each panel's visual composition
4. If a flow is **missing** from the Design Spine but the user requests it, that's a finding for `/assess:design` — note it in the storyboard index as a gap and proceed with best-practice defaults

## Success Criteria

A storyboard is complete when:
- ✅ 4–12 sequential panels generated (image files exist)
- ✅ Each panel demonstrates one clear step in the flow
- ✅ Visual style consistent across all panels (brand identity applied)
- ✅ Alt text present on every panel (auto or provided)
- ✅ Storyboard index file written to `.plaesy/memory/storyboards/{flow-name}.md`
- ✅ `.plaesy/memory/storyboards.md` updated to index the new storyboard
- ✅ Output directory reported (e.g., `assets/storyboards/user-signup/`)

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
