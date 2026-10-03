# Changelog

All notable changes to Plaesy Constitution Kit are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project uses [Semantic Versioning](https://semver.org/) once it reaches 1.0.0.

## [Unreleased]

### Added

- `plaesy analyze` — generate AI-optimized project analysis under `.plaesy/analysis/`.
- `plaesy config` — read and validate `scripts/configs/platform.json`
  (`get-mapping`, `get-excludes`, `get-clean-dirs`, `get-platform`, `show`).
- `plaesy context` — keep the agent context files in sync with `plan.md`.
- `plaesy features` — list, create and inspect feature branches under `.plaesy/specs/`.
- `plaesy graph` — build or query a lightweight knowledge graph of a project,
  emitting `project.graph.json`, `project.html` and `reports.md`.
- `plaesy images` — generate image assets through a provider API.
- `plaesy install` / `plaesy uninstall` — install or remove the CLI binary in a
  well-known bin directory.
- `plaesy platforms` — `detect`, `list`, `show` and `get` for the AI platform a
  project targets.
- `plaesy reload` — refresh the generated `.plaesy/` files while preserving
  `memory.md`, `context.md` and `state.json`.
- `plaesy search` — embedding-based semantic search over extracted symbols, for
  finding existing code that serves the same purpose under different wording.
  `--index` builds the index; a query refuses to run against an index older than
  the source it was built from.
- `plaesy stack` — inspect a target project's technology stack, and install the
  instruction files that stack selects.
- `plaesy status` — show the CLI version and environment status.
- `plaesy tasks` — the task lifecycle across
  `.plaesy/tasks/{backlog,todo,doing,done,blocked}`, with `move`, `complete`,
  `block`, `unblock` and prerequisite checks.
- `plaesy trim` — three-layer token compression for command output, memory and
  instruction files.
- `plaesy validate` — validate a constitution, memory, Markdown (against a
  ratcheted baseline) or an OOXML document.
- `plaesy clean` — remove the framework's own files and directories at three
  levels, with `--dry-run`, backup control and platform-specific guidance.
- `plaesy upgrade` and `plaesy repair` — registered as explicit stubs that
  report "not yet implemented" rather than being absent from the command list.

### Fixed

- `platform.json` keyed six platforms by their short names (`claude`, `cursor`,
  `copilot`, `kilo`, `trae`, `windsurf`) while the `--ai` flag help, the README,
  the docs and `promptExtension()` all name the long form. The long form —
  the spelling the documentation tells users to type — resolved to nothing:
  `plaesy platforms show claude_code` printed `Name: Unknown`, and
  `plaesy clean --ai claude_code` removed nothing while reporting that it had.
  The keys are now the canonical ids (`claude_code`, `cursor_ai`,
  `github_copilot`, `kilo_code`, `trae_ai`, `windsurf_ai`) and the alias table
  maps the shorthands onto them, so `promptExtension()` reaches its
  `github_copilot` case and installs Copilot prompts as `.prompt.md`.
- `plaesy init --ai` resolved aliases against its own local table, which
  returned the same short ids, so `--ai claude_code` wrote a platform file for a
  platform the config does not declare. It now shares the single alias table in
  `internal/config` and adds only the dashed spellings.

## [0.0.1] - Initial release

- Initial public commit of the Plaesy Spec-Kit framework: prompts, instructions,
  agents/roles, checklists, templates, and bash/PowerShell automation scripts.
