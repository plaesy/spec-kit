---
description: "Generate or refresh .plaesy/memory/design.md — the project design-system source of truth (Google Labs DESIGN.md format, YAML tokens + rationale prose)"
---

# `/spec:design` command instructions

This is the **design**-scoped entry point into `/spec`. It produces a real
`.plaesy/memory/design.md` file conforming to the **DESIGN.md** format
specification published by Google Labs
([`google-labs-code/design.md`](https://github.com/google-labs-code/design.md),
spec version `alpha`) — not a design *description*, and not a summary of the
file the user could have written themselves.

**Why this is a `/spec` scope and not an `instructions/` concern.** The
project rule is that stack-specific scaffolding belongs in `instructions/`
because it is knowledge an agent loads, not an asset. `design.md` is the
opposite: it is a *named deliverable file* with a fixed path, a fixed schema and
a validation step — the same class of thing as `/create:tasks` writing task
files into `.plaesy/tasks/`. The *rules for filling it in* stay in
`.plaesy/instructions/how-to-create-designmd.md`; this command generates it.

## Usage Format

```bash

/spec:design --init "Acme Dashboard — dense internal analytics tool for ops teams"
/spec:design --refresh --for .plaesy/memory/design.md
/spec:design --from-code --scan src/ --out .plaesy/memory/design.md
/spec:design --from-figma --out .plaesy/memory/design.md
/spec:design --lint
/spec:design --export tailwind --out theme.css
```

| Flag | Default | Meaning |
|---|---|---|
| `--init` / description | — | Brand/product direction in one line: what the product is, who it is for, how it should feel. Used with `--init` to author a new file |
| `--refresh` | false | Improve an existing `design.md` **in place**: reconcile stale tokens against the code, keep verified ones, do not rewrite blindly |
| `--from-code` | false | Extract the palette/spacing/type actually in use (CSS variables, Tailwind config, theme files, styled-components) rather than inventing values |
| `--scan` | `.` | Path to scan for `--from-code` |
| `--from-figma` | false | Pull canonical token values from a Figma source via the `figma-use` MCP tool, if the user has design files |
| `--for` | `.plaesy/memory/design.md` | Path of the file to read (with `--refresh`) or to derive context from |
| `--out` | `.plaesy/memory/design.md` | Output path. A second file (`design.dark.md`) is the documented way to carry a dark palette — see *Dark mode* below |
| `--dials` | asked, not guessed | `energy`, `rhythm`, `motion` as `1\|2\|3`. If the user has no brand direction, take the conservative default and label the output a **draft** |
| `--lint` | false | Run the spec linter against the file and report findings |
| `--export` | none | `tailwind` (v3 JSON / v4 CSS) or `dtcg` (W3C Design Tokens) — emit the token file the code actually consumes |
| `--overwrite` | false | Required to replace an existing file. Without it, write a new path or `design.v2.md` |

## Protocol

### Step 0: Decide Extract vs Author

This is the step that decides whether the output is worth anything, and it is
not a formality.

- **Extract** (any of `--refresh`, `--from-code`, `--from-figma`, or an
  existing `design.md`) — the project already has a visual identity. Read the
  *actual* values in use: CSS custom properties, Tailwind `theme` block,
  theme object, `styled-components` ThemeProvider, SVG fills in the component
  library. The file documents what the project **does**, not a generic starter
  palette.
- **Author** (`--init` with nothing to extract) — a brand-new project. Ask for
  brand direction. Do **not** fill in placeholder black-and-white defaults and
  call it done; a template with `#000000` left in every row is a draft that
  looks finished, which is the worst of both. Load
  `.plaesy/instructions/frontend-design.md` and run its two-pass
  plan-then-build process (token-system brainstorm, then checked against the
  five AI-slop clusters it lists) **before** writing Step 2's YAML — the dials
  in Step 3 record the result of that pass, they are not a substitute for it.
- If neither is possible (no code, no brand direction, no Figma), write the
  file with the `omitted` front-matter field marking what is undecided and
  state plainly in the report that it is a **skeleton awaiting direction**.
  An honest empty file beats a plausible fabricated one.

### Step 1: Resolve the Canonical Source

In priority order, stopping at the first that exists:

1. `figma-use` MCP (with `--from-figma`) — the design system of record
2. `.plaesy/roles/designer.md` — recorded brand decisions
3. `.plaesy/instructions/brandkit.md` — brand/product positioning
4. Source code (`--from-code`/`--scan`) — the values actually in effect
5. `.plaesy/templates/design.template.md` — only as a **skeleton**, never as a
   source of truth for values

Read the field-by-field rules in
`.plaesy/instructions/how-to-create-designmd.md` before writing. It owns the
token shapes, the section order and the liveliness dials; this command owns the
process.

### Step 2: Write the Tokens (YAML front matter)

Tokens are **normative** — they are the machine contract, and they must be
real values. The spec is at
[`docs/spec.md`](https://github.com/google-labs-code/design.md/blob/main/docs/spec.md);
**when this command and the spec disagree, the spec wins.**

| Field | Required | Notes |
| --- | --- | --- |
| `name` | yes | Product / design-system name |
| `version` | no | `"alpha"` — the spec version this targets |
| `description` | no | One line |
| `omitted` | no | Array of intentionally-skipped sections, each optionally with a `reason`. **Use this instead of leaving a section silently empty** — it is the spec's own mechanism for exactly this |
| `colors` | no | `map<string, Color>`: hex (`#RRGGBB` is the recommended default), named, `rgb()`/`hsl()`, or wide-gamut `oklch()`/`lab()`. Internally converted to sRGB for contrast checks |
| `typography` | no | `map<string, Typography>` — `fontFamily`, `fontSize`, `fontWeight`, `lineHeight`, `letterSpacing`, `fontFeature`, `fontVariation` |
| `spacing` | no | `map<string, Dimension \| number>` — unit suffixes are limited to `px`, `em`, `rem`; a unitless number is a multiplier/ratio |
| `rounded` | no | `map<string, Dimension>`. The spec's field name is `rounded`, **not** `radii` — do not "correct" it |
| `components` | no | `map<string, map<string, value>>`. Permitted properties are exactly: `backgroundColor`, `textColor`, `typography`, `rounded`, `padding`, `size`, `height`, `width` |

Reference other tokens with `{path.to.token}` (e.g.
`backgroundColor: "{colors.primary}"`) instead of repeating a literal. This is
how a token change propagates, and it is how the linter detects a broken
reference.

### Step 3: Write the Rationale (markdown body)

Eight `##` sections, in this order, each at most once:

1. **Overview** — brand personality and emotional tone in a sentence or two,
   not a feature list. Also state, one line each, *why* each liveliness dial
   was set.
2. **Colors** — the semantic **role** of each palette, not a restatement of the
   hex values already in the front matter.
3. **Typography** — font strategy and which level to reach for when.
4. **Layout** — grid model and spacing rhythm.
5. **Elevation & Depth** — how hierarchy is communicated (shadow, or borders
   and contrast for a flat system).
6. **Shapes** — corner-radius and general form language.
7. **Components** — conventions per component *family*, not the full inventory.
   Link to `docs/components.md` (from `/doc`) for the exhaustive list.
8. **Do's and Don'ts** — guardrails phrased as pairs an agent can pattern-match
   against generated code.

Omit a section via the `omitted` front-matter field rather than leaving an empty
heading. A duplicate heading is a **hard error** in the spec, not a warning.

### Step 4: Validate (`--lint`)

Run the official linter and report its output verbatim:

```bash

npx @google/design.md lint .plaesy/memory/design.md
```

On Windows/PowerShell use the `designmd` alias, which avoids the `.md`
file-association conflict:

```bash

npx -p @google/design.md designmd lint .plaesy/memory/design.md
```

It checks token references that do not resolve, WCAG AA contrast on every
component `backgroundColor`/`textColor` pair (4.5:1 for normal text), orphaned
tokens, section order, and unknown keys.

**If the linter cannot run** (no Node, no network), do not report success and do
not silently skip it. Say the linter was unavailable, then check by hand what
you can and name what you could not verify — in particular the contrast pairs,
which are the checks most often wrong and most often skipped.

**Contrast is not optional.** Convert the pairs to sRGB and verify them. A
light/dark pair that was hand-picked and never checked is a defect, and
`/assess:design` will flag it later at higher cost.

### Step 5: Export (optional, `--export`)

```bash

# Tailwind v4 @theme block
npx @google/design.md export --format css-tailwind .plaesy/memory/design.md > theme.css

# Tailwind v3 theme config
npx @google/design.md export --format json-tailwind .plaesy/memory/design.md > tailwind.theme.json

# W3C Design Tokens (DTCG)
npx @google/design.md export --format dtcg .plaesy/memory/design.md > tokens.json
```

Export is what turns the file from documentation into something the build
consumes. Note in the report when it was run, and when it was not.

### Step 6: Report & Index

Report the output path, the resolved canonical source, the lint result, and
whether the file is a **deliverable** or a **draft** (a draft is one with
unresolved brand direction, placeholder tokens, or omitted sections with no
reason). Never let a draft be reported as finished.

Then wire it in: `design.md` is read by `/implement:design` and audited by
`/assess:design`, so name both in the "Next" line.

## Dark mode

The spec has **no native dark-mode namespace** at version `alpha`. Two
project-level options, in order of preference:

1. **A second file** — `design.md` (light) + `design.dark.md` (dark), with a
   line in each Overview naming its counterpart. Simple, and each file lints
   on its own.
2. **A project-extended structure** — the spec preserves unknown top-level
   sections, so a `themes:` block is legal. Choose this only if a single file
   is a hard requirement, and document the extension in the file.

Reserve `adr`, `brd` and `spec` for future siblings under `doc/`; do not name
them in prose as though they exist. A selector documented in the corpus is
bound to its file by `TestNestedSubCommandNamesResolveToFiles`, and naming an
unbuilt scope fails that guard — which is the correct outcome.

**Known deviation, flagged not fixed.** `templates/design.template.md` currently
nests the palette as `colors.light.*` / `colors.dark.*`. The spec defines
`colors` as a flat `map<string, Color>` and states that a reference must point
at a **primitive**, not a group — so `colors.light.primary` is a group path.
This template predates the command and changing it is a committed design
judgment, not an auto-fix. Default to the template's existing shape when
`--refresh`-ing a project that already uses it (changing the shape silently
would break every `{colors...}` reference in the codebase), and report the
deviation in the output so the owner can decide.

## Programmatic Invocation

Any prompt that needs the design system mid-run calls this directly:

- `/assess:design` (Mode 1) — generates or updates the Design Spine
- `/implement:design` — ensures the file exists before building UI
- `/create:images` and `/create:storyboard` — read it for art direction
- `/doc` — links to it as the design-system section of the documentation

Do not re-describe token generation inline. Call the command.

## What Counts as Done

- **A file exists at a stated path.** This command's deliverable is a file. If
  brand direction is missing, the honest deliverable is a marked draft, not a
  report that nothing could be produced
- **Every token is a real value.** No `#000000` placeholder row left unedited
  from the template, and no token the project does not actually use
- **The linter ran, or its absence is reported.** A file that was never linted
  and a file that linted clean are different claims; say which one this is
- **The eight sections appear once each, in order** — or are listed in
  `omitted` with a reason
- **Contrast was verified computationally** for every component color pair, or
  the unverified ones are named
- **An existing file is not overwritten** without `--overwrite`
- **The file is labeled a draft or a deliverable**, and the reason is stated

## Error Recovery

- **Linter unavailable** → report it, run the checks you can by hand, name the
  ones you could not verify. Never report a pass you did not observe
- **`--from-figma` requested but no Figma access** → say so, fall back to
  `--from-code` if code exists, otherwise produce a draft skeleton
- **Existing `design.md` has a duplicate or out-of-order section** → that is a
  spec error, not a style preference; fix it and say it was fixed
- **No code, no Figma, no brand direction** → write the skeleton with `omitted`
  entries carrying reasons, seed a neutral placeholder palette from
  `.plaesy/templates/design.template.md`, label the whole draft
  `ASSUMED — no brand direction found`, and record it. Do not invent a
  bespoke palette and present it as if it were derived

## Success Criteria

A design document is complete when:

- ✅ `.plaesy/memory/design.md` written (or refreshed in place) and reported by path
- ✅ Front matter parses as valid YAML and every token holds a real value
- ✅ Token references use `{path.to.token}`; none dangle
- ✅ Eight sections present once each in spec order, or declared in `omitted`
- ✅ Linter run and its findings reported verbatim (or its unavailability stated)
- ✅ WCAG AA contrast verified for component color pairs
- ✅ Liveliness dials set to non-default values derived from actual brand
      direction, each justified in one line under Overview — or the file is
      explicitly a draft
- ✅ Canonical source named in the report
- ✅ Deliverable vs draft status stated with the reason
- ✅ `/implement:design` and `/assess:design` wired in via the "Next" line

## Output Format

```text
✅ Design Specification Created — /spec:design
├─ Output: [.plaesy/memory/design.md]
├─ Source of truth: [figma-use | designer.md | brandkit.md | source code (--scan <path>) | template skeleton]
├─ Mode: [extract | author]
├─ Status: [deliverable | draft — <reason>]
├─ Dials: energy=<n> rhythm=<n> motion=<n>  (<reason each, or "defaults — no brand direction">)
├─ Sections: [8/8 | 6/8 — omitted: <names + reasons>]
├─ Lint: [clean | N findings (<severity>) | unavailable — <what was checked by hand instead>]
├─ Contrast: [verified, N pairs | N pairs unverified — named]
└─ Export: [none | theme.css | tailwind.theme.json | tokens.json]

Next: /implement:design builds UI against these tokens; /assess:design audits code against them.

Note: Use /assess:design to audit the codebase against this file, or /spec:design --refresh to keep it current.
```

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`

**Research sources** (retrieved 2026-09-29):

- https://github.com/google-labs-code/design.md — format specification, CLI, linter
- https://github.com/google-labs-code/design.md/blob/main/docs/spec.md — authoritative field list, section order, consumer behavior for unknown content
- https://www.designtokens.org/tr/2025.10/format/ — W3C Design Token Format, the model the token schema follows
-  https://andrew.ooo/posts/google-labs-design-md-ai-agent-design-spec-review — independent review; `lint`/`diff`/`export` command reference, alpha-version limitations (no dark-mode namespace,
  7-property component model, no Figma export)
