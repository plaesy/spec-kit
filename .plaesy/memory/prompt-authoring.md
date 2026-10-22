---
name: Prompt Authoring
description: Durable rules for writing and routing prompts in this repo
updatedAt: "2026-09-30T19:40:00.000Z"
---

# Prompt Authoring

- **`prompts/` is the source; platform copies are generated.** An author edits
  `prompts/`. `.kilo/commands/`, `.claude/commands/` and friends are build
  artifacts produced by `plaesy init` / `plaesy reload`. Editing a copy is
  silently undone on the next reload. The runtime loads the *copy*, so a prompt
  edit that is not regenerated has changed nothing that an agent can see.
- **A prompt declares its dimension, it does not assume it.** The nine
  dimensions are `technical`, `design`, `business`, `marketing`, `legal`,
  `financial`, `product`, `management`, `operations`. The active one comes from
  the project's constitution. `technical` is the software instance and is named
  as a variant, never written as the default — see the Dimension Resolution
  table in `prompts/loop.md`.
- **A prompt must not contradict its own stop rule.** `prompts/loop.md` once
  said in three places that exhaustion ends the run while its Step 4 said the
  opposite. When a rule is restated in several sections, check that all copies
  agree before declaring the file done.
- **A selector that looks like `{family}:{dimension}` but is not one needs an
  explicit "not a dimension" row.** `/spec:design` has the same shape as
  `/assess:design` and `design` is a canonical dimension, so a routing reader
  would treat it as a dimension cell. The Command Coverage table in
  `dimension-mapping.instructions.md` says so in as many words, and says where
  a design-dimension finding *should* go. A shape that reads like something
  else is a defect even when the implementation is right.
- **Dropping a new file into `instructions/` does not wire it up.** A file
  there is orphaned — never copied to any consuming project's
  `.plaesy/instructions/` — unless it is also added to `always_load` or
  `scope_load` in `instructions/mapping.json` (`copyInstructions` reads only
  those two arrays; everything else in the file is documentation/intent per
  its own `field_semantics_note`). Confirmed twice now:
  `how-to-create-designmd.instructions.md` already carried a note warning
  about this; `frontend-design.instructions.md` hit the same gap on
  2026-10-03 and needed the same fix. Checklist when adding a new
  `*.instructions.md`: (1) add it to `mapping.json` `scope_load` (or a
  `frameworks`/`cross_cutting` entry if it should load conditionally), (2)
  run `go run ./internal/assets/gen` to resync the embed mirror, (3) run
  `./plaesy reload` on *this* repo so `.plaesy/instructions/` stops drifting
  from source (`TestInstalledInstructionsMatchSource` / idempotency test
  catch this but a `go build` does not), (4) bump
  `docs/metadata.json` → `counts.instructionFiles`
  (`TestMetadataCountsMatchTheTree` catches a stale count), (5) hard-wrap
  prose to the MD013 200-char ratchet — content is often pasted from an
  external source at full prose width and fails `plaesy validate markdown`
  on first commit.
- **The test for the scaffolding rule is *what the file is for*, not what it is
  made of.** Stack/tech-specific scaffolding belongs in
  `instructions/*.instructions.md` (auto-loads per file type via `mapping.json`),
  never a dedicated `/create` sub-router — user-confirmed 2026-09-29 after
  `/create:templates: {api,ci,infra,project}` was found baking one hardcoded
  stack (Node/pnpm, Terraform, GitHub Actions) into the only non-dimension-agnostic
  corner of the framework. Deleted; content moved into `api-design-principles`,
  `terraform`, `devops-core-principles` and new `monorepo-scaffolding`
  instructions. `/spec:design` (same day) ships a *generation* command that is
  not code scaffolding: it has a named deliverable, a fixed path, a fixed schema
  and a validation step. Fill-in rules stayed in
  `how-to-create-designmd.instructions.md`; only the generation moved. It was
  first built as `/create:doc:design` and moved once it was clear that
  *specification* and *asset generation* are different jobs wearing similar
  names — `/spec` authors what other commands read back, `/create` renders
  pixels and vectors, `/doc` describes what already exists. If a future `/create`
  scope wants to scaffold code, route it to instructions first.
- **External best-practice content belongs in `instructions/`/`templates/`, not
  duplicated into `prompts/*.md`.** Routers stay thin per the single-source-of-truth
  rule; only touch a prompt when it is an actual enforcement/creation point (a
  gate file like `quality-gates.instructions.md`, or a command that actually
  generates the asset in question).
- **`plaesy reload` does not prune — a renamed or deleted prompt leaves a
  working stale command in the mirror.** Found while moving
  `/create:doc:design` to `/spec:design`: `prompts/create/doc/design.md` was
  gone, `.kilo/commands/create/doc/design.md` remained, still titled with the
  old name and still invokable. Nothing failed, because the parity test walks
  the *source* and cannot see a file the source lacks.
  `TestMirrorHasNoFileMissingFromSource` now checks that direction. Fixed
  2026-09-30 by opt-in stale-file pruning: `reload --prune` reports,
  `--prune-apply` deletes. Source-missing/empty, shared deny-list,
  unnamed-platform and outside-root paths are protected; blanket rsync
  `--delete` remains prohibited.
- **`templates/design.template.md` deviates from the DESIGN.md spec and this is
  known, not forgotten.** It nests the palette as `colors.light.*` /
  `colors.dark.*`; the spec defines `colors` as a flat `map<string, Color>` and
  requires a token reference to point at a primitive, not a group. Changing the
  shape is a committed design judgment that would break every `{colors...}`
  reference in a project already using it, so `/spec:design` preserves the
  existing shape on `--refresh` and reports the deviation instead of silently
  migrating.
- **Orphan/reachability metrics do not apply to entry-point prompts.**
  `/save`, `/assess` and friends are called directly by a human, so having no
  incoming file reference is the expected condition, not a defect. Applying an
  orphan rate to them produces a number with no meaning.
