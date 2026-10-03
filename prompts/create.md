---
description: "Generate real assets (images, diagrams, tasks, storyboards) from specifications — router for /create:{scope}"
---

# `/create` command instructions

⚡ **Run with**: standard (single-agent; no multi-agent fan-out needed for one asset call)

## Usage Format

```bash

# Asset generation
/create:images "a flat-style empty-state illustration for an empty inbox"
/create:storyboard "User signup flow: landing → email → verify → dashboard"
/create:diagram "Architecture: frontend → API gateway → microservices → database"
/create:tasks "requirements.md" --format markdown
```

`/create` is a **router**, same shape as `/assess`/`/improve`/`/fix`: it has no
scope-less behavior of its own. Today it has four scopes.

Specifications — a document that is the source of truth another command reads
back — are **not** a `/create` scope. `/spec:design` is the one that shipped;
see `prompts/spec.md` for the family and the dividing line.

Boilerplate/scaffolding (APIs, IaC, CI pipelines, monorepo layout) is **not**
a `/create` scope. It is stack-specific knowledge, not an asset-generation
call, so it lives in `instructions/` and loads automatically for the file
being written: `instructions/api-design-principles.instructions.md` (OpenAPI/
GraphQL), `instructions/terraform.instructions.md` (IaC), `instructions/
devops-core-principles.instructions.md` (CI/CD pipelines), `instructions/
monorepo-scaffolding.instructions.md` (Node/pnpm workspace layout).
`/implement:technical` reads the instruction file matching what it is
scaffolding — see `.plaesy/instructions/mapping.json` for the detection rules.

### Asset Generation (Real Binary/File Output)

| Scope                | Purpose                                                                                                              |
| -------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `/create:images` | Turn a text description (or a `.plaesy/memory/design.md` / brandkit entry) into an actual saved image file, not just a prompt string |
| `/create:storyboard` | Generate a visual storyboard (sequential narrative panels) from a user journey/interaction flow — produces real image assets per panel, not descriptions |
| `/create:diagram` | Generate visual diagrams (architecture, flowchart, ERD, sequence, swimlane, mindmap) from natural language or code context — produces SVG and Mermaid markdown files |
| `/create:tasks` | Generate hierarchical task backlog (epics → user stories → acceptance criteria → technical tasks) from requirements specifications — produces structured Markdown + JSON backlog files |

#### Provider Resolution

Scopes that generate image assets resolve a provider the same way, in this
order, stopping at the first match:

1. Env var `PLAESY_IMAGE_PROVIDER` (`openai` | `gemini`)
2. Default: `openai`
3. If the env var was absent, the default is unconfirmed — name the resolved
   provider in the report so a wrong pick is visible immediately

No project-wide override file exists: `scripts/internal/imagegen/imagegen.go`
reads only the env var, then falls back to `openai`. To pin a provider per
project, set `PLAESY_IMAGE_PROVIDER` in the project's environment — there is
no `.plaesy/config/image-provider.json`.

Then confirm the matching API key env var is set (`OPENAI_API_KEY` for
`openai`, `GEMINI_API_KEY` for `gemini`). Centralized here so `/create:images`
and `/create:storyboard` can't drift on the env var name or override meaning.

More scopes (e.g. `/create:audio`, `/create:video`) are added the same way if
the project ever needs them — this file stays a thin router; each scope has its
own command (e.g., `/create:images`).

`/create:tasks` writes structured text files and `/spec:design` writes a
specification, yet only the latter lives under `/spec`. The dividing line is
**what the file is for**: a specification is a source of truth a later command
reads back (`/implement:design` reads `design.md`; `/assess:design` audits
against it), while `/create:tasks` output is consumed by a human planning a
sprint. `/spec` is the home for the first kind.

## Why This Exists

Every other Plaesy prompt (`/implement:design`, `/improve:design`, `/doc`) can
describe what an asset *should* look like, but none calls an image-generation
API and writes bytes to disk — they stop at a text description and leave the
human to generate it by hand. `/create:images` closes that gap: given a
description (typed by a user, or handed programmatically by another prompt),
it produces a real image file and reports its path, so the calling context can
reference it immediately (in an `<img>`, a Figma upload, a doc).

The specification side of the same gap — `.plaesy/memory/design.md`, which
four commands read and only `/assess:design` Mode 1 could write — is
`/spec:design`, not a scope here. It was built this way first and moved when it
turned out that “specification” and “asset generation” are different
jobs wearing a similar name.

## Callable By Other Prompts

Any prompt that needs a concrete asset mid-run invokes the appropriate `/create:{scope}` directly with a structured call instead of re-describing generation itself:

- **Images**: `/implement:design` building a component that needs an icon/illustration, `/improve:design` replacing an outdated asset, `/doc` illustrating a concept
- **Diagrams**: `/implement:technical` documenting system architecture, `/assess:technical` analyzing existing systems, `/doc` explaining processes/data models
- **Tasks**: `/implement` decomposing requirements into sprint-ready stories, `/doc` generating task documentation from specifications

Specifications are called via `/spec:{scope}` (see `prompts/spec.md`) — not
from here.

See "Programmatic Invocation" sections in the respective scope documentation.

## What Counts as Done

- **Produce a saved file.** This command's deliverable is a file on disk. When
  generation cannot run (no provider configured), say so explicitly and hand back
  the prompt as a documented fallback — a fallback and a result are not the same thing
- **Fail loudly when the configured provider is unavailable**, naming exactly what
  to configure. A silent substitution produces an artifact that does not match the
  project's own setup
- **Layer `.plaesy/instructions/brandkit.md` and `.plaesy/memory/design.md` into the
  asset** when either exists — this is asset generation *for the project*, and
  generic art drops the project's own identity
- **Write to a new path or version** unless the user explicitly asked to replace
  the existing file
- **Report deliverable vs draft, with the reason.** A file assembled from
  placeholders is a draft, and a draft that reads as finished is worse than no
  file — an agent reads it as authoritative
- **Say when validation could not run.** A file that was never validated and one
  that validated clean are different claims

## Output Format

```text

✅ Asset Created — /create:{scope}
├─ Type: [images | storyboard | diagram | tasks]
├─ Provider: [resolved provider, if asset generation — name it even when it's the unconfirmed default]
├─ Output: [saved file path(s)]
├─ Brandkit/Design applied: [yes, source file | n/a]
└─ Validation: [result, or "not run — <reason>"]

Fallback (if generation could not run):
⚠️ Provider unavailable — [what to configure]
└─ Handed back as: [documented prompt/spec the user or caller can act on]

Next: [what the calling context — or the user — should do with the file]
```

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
