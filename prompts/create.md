---
description: "Generate real assets (images, tasks, storyboards) from specifications — router for /create:{scope}"
---
# `/create` command instructions

⚡ **Run with**: standard (single-agent; no multi-agent fan-out needed for one asset call)

## Usage Format

```bash
/create:images "a flat-style empty-state illustration for an empty inbox"
/create:images --for .plaesy/memory/design-spine.md --component "empty-state/inbox"
/create:tasks "requirements.md" --format markdown
/create:storyboard "User signup flow: browse landing → enter email → verify → dashboard"
```

`/create` is a **router**, same shape as `/assess`/`/improve`/`/fix`: it has no
scope-less behavior of its own. Today it has four scopes:

| Scope                | Purpose                                                                                                              |
| -------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `/create:images` | Turn a text description (or a design-spine/brandkit entry) into an actual saved image file, not just a prompt string |
| `/create:storyboard` | Generate a visual storyboard (sequential narrative panels) from a user journey/interaction flow — produces real image assets per panel, not descriptions |
| `/create:diagram` | Generate visual diagrams (architecture, flowchart, ERD, sequence, swimlane, mindmap) from natural language or code context — produces SVG and Mermaid markdown files |
| `/create:tasks` | Generate hierarchical task backlog (epics → user stories → acceptance criteria → technical tasks) from requirements specifications — produces structured Markdown + JSON backlog files |

More scopes (e.g. `/create:audio`, `/create:video`, `/create:template`) are added the same way if
the project ever needs them — this file stays a thin router; each scope has its
own command (e.g., `/create:images`, `/create:storyboard`, `/create:diagram`, `/create:tasks`).

## Why This Exists

Every other Plaesy prompt (`/implement:design`, `/improve:design`, `/doc`) can
describe what an asset *should* look like, but none of them actually call an
image-generation API and write bytes to disk — they stop at a text description
and leave the human to go generate it by hand. `/create:images` closes that
gap: given a description (typed by a user, or handed to it programmatically by
another prompt), it produces a real image file in the project and reports its
path, so the calling context can reference it immediately (in an `<img>`, a
Figma upload, a doc).

## Callable By Other Prompts

Any prompt that needs a concrete asset mid-run invokes the appropriate `/create:{scope}` directly with a structured call instead of re-describing generation itself:

- **Images**: `/implement:design` building a component that needs an icon/illustration, `/improve:design` replacing an outdated asset, `/doc` illustrating a concept
- **Diagrams**: `/implement:technical` documenting system architecture, `/assess:technical` analyzing existing systems, `/doc` explaining processes/data models
- **Tasks**: `/implement` decomposing requirements into sprint-ready stories, `/doc` generating task documentation from specifications

See "Programmatic Invocation" sections in the respective scope documentation.

## Anti-Patterns (NEVER Do These)

- ❌ Return only a text prompt and call it "done" — the point of this command is
  a saved file; if generation truly cannot run (no provider configured), say so
  explicitly and hand back the prompt as a documented fallback, don't blur the two
- ❌ Silently pick a different provider than the one configured — fail loudly and
  tell the user what to configure
- ❌ Generate an asset that ignores the project's `brandkit.instructions.md` /
  `.plaesy/memory/design-spine.md` when either exists — this is asset generation
  *for the project*, not generic stock art
- ❌ Overwrite an existing asset file without confirmation — write to a new
  path/version unless the user explicitly asked to replace one

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
