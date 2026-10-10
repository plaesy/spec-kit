---
description: "Generate specification documents that are the project's source of truth (design system, architecture) — router for /spec:{scope}"
---

# `/spec` command instructions

⚡ **Run with**: standard (single-agent; one specification document per run)

## Usage Format

```bash

/spec:design --init "Acme Dashboard — dense internal analytics tool for ops teams"  # Generate or refresh .plaesy/memory/design.md — the project design-system source of truth (Google Labs DESIGN.md format, YAML tokens + rationale prose)
/spec:design --refresh --for .plaesy/memory/design.md
/spec:design --from-code --scan src/ --out .plaesy/memory/design.md
/spec:design --lint
/spec:design --export tailwind --out theme.css
```

`/spec` is a **router**, same shape as `/assess`/`/improve`/`/create`: it has
no scope-less behavior of its own. Today it has one scope.

| Scope | Purpose |
| --- | --- |
| `/spec:design` | Generate or refresh `.plaesy/memory/design.md`, the project design-system source of truth — a real DESIGN.md file (YAML tokens + rationale prose, Google Labs format) |

## What `/spec` Is For

`/spec` produces **specifications**: documents that are the source of truth a
later command reads back. That is the dividing line from its neighbours, and
it is worth stating because the names overlap:

- `/spec:{name}` — *author* a specification. Produces a file that other
  commands consume. `/implement:design` reads `design.md`; `/assess:design`
  audits code against it.
- `/create:{scope}` — *render an asset* from a description. Pixels, vectors,
  and generated backlog files.
- `/doc` — *describe* what already exists. Reads the codebase, writes
  documentation about it.

A project that has no `design.md` and runs `/implement:design` gets an agent
that invents a palette. A project that has a `design.md` and runs
`/spec:design --refresh` gets one that matches its code. That gap is why
`/spec` exists.

## What Is Not A Scope

- **Stack-specific scaffolding** (APIs, IaC, CI pipelines, monorepo layout) is
  not a `/spec` scope. It is knowledge an agent loads, not a deliverable, so
  it lives in `instructions/` and auto-loads per file type via
  `.plaesy/instructions/mapping.json` — see `api-design-principles`,
  `terraform`, `devops-core-principles`, `monorepo-scaffolding`.
- **Dimension routing** is not here either. `/spec:design` shares the shape
  `{family}:{dimension}` with `/assess:design`, `/implement:design` and the
  rest, and `design` *is* one of the nine canonical dimensions — but this
  selector names a **document type**, not a lens. See the Command Coverage
  table in `dimension-mapping.instructions.md`, which records the difference
  so a routing reader does not treat `/spec:design` as a dimension cell.

## When A New Scope Is Added

Add it as `/spec:{name}` → `prompts/spec/{name}.md`. The segment is the
directory in the path, and `TestNestedSubCommandNamesResolveToFiles` binds the
two for the three-segment shape. Name a reserved sibling in prose only as a
bare word (e.g. "an ADR scope") — a selector written in full is bound to a
file by a guard, and naming one that does not exist fails the build, which is
the correct outcome.

## Why This Exists

Four commands read `.plaesy/memory/design.md` (`/implement:design`,
`/assess:design`, `/create:images`, `/create:storyboard`) and only
`/assess:design` Mode 1 could produce it. A project that skipped the
assessment had no design context at all, and no prompt could create it on
demand. `/spec:design` closes that gap: given brand direction or an existing
codebase, it produces a real specification file and reports its path, its
validation status, and whether it is a deliverable or a draft.

## Callable By Other Prompts

Any prompt that needs a specification mid-run invokes `/spec:{scope}` directly
rather than re-describing generation:

- `/implement:design` — ensures `design.md` exists before building UI
- `/assess:design` (Mode 1) — generates or updates the Design Spine
- `/create:images` and `/create:storyboard` — read it for art direction
- `/doc` — links to it as the design-system section of the documentation

## What Counts as Done

- **A file exists at a stated path.** A specification is a file, not a
  description of one
- **Every token is a real value** — no placeholder left unedited from the
  template, and no value the project does not actually use
- **Validation ran, or its absence is reported.** A file that was never
  linted and one that linted clean are different claims
- **Deliverable vs draft is stated, with the reason.** A specification
  assembled from placeholders reads as authoritative to the next agent, which
  is worse than no file
- **An existing file is not overwritten** without an explicit request

## Output Format

```text
✅ Specification Created — /spec:{scope}
├─ Output: [saved file path(s)]
├─ Source of truth: [which canonical source was read]
├─ Status: [deliverable | draft — <reason>]
└─ Validation: [result, or "not run — <reason>"]

Next: [what the calling context — or the user — should do with the file]
```

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
