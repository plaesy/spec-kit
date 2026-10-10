# Changelog

All notable changes to Plaesy Constitution Kit are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project uses [Semantic Versioning](https://semver.org/) once it reaches 1.0.0.

## [Unreleased]

### Added

- `plaesy analyze` — generate AI-optimized project analysis under `.plaesy/analysis/`.
  It also refreshes the semantic-search index `plaesy search` queries, so a
  search after an analyze run is no longer refused as stale; `--no-index` skips
  that step, and `plaesy search --index` still rebuilds the index alone.
- `plaesy config` — read and validate `scripts/configs/platform.json`
  (`get-mapping`, `get-excludes`, `get-clean-dirs`, `get-platform`, `show`).
- `plaesy context` — keep the agent context files in sync with `plan.md`.
- `plaesy doctor` — check health of all system components (reach platforms,
  search index, and related CLI state).
- `plaesy eval` — run the search evaluation harness against a golden set.
- `plaesy features` — list, create and inspect feature branches under `.plaesy/specs/`.
- `plaesy graph` — build or query a lightweight knowledge graph of a project,
  emitting `project.graph.json`, `project.html` and `reports.md`.
- `plaesy images` — generate image assets through a provider API.
- `plaesy install` / `plaesy uninstall` — install or remove the CLI binary in a
  well-known bin directory.
- `plaesy platforms` — `detect`, `list`, `show` and `get` for the AI platform a
  project targets.
- `plaesy reach` — access external data sources across multiple platforms,
  with `configure` (store platform credentials) and `list` (available
  platforms) subcommands.
- `plaesy reload` — refresh the generated `.plaesy/` files while preserving
  `memory.md`, `context.md` and `state.json`.
- `plaesy search` — embedding-based semantic search over extracted symbols, for
  finding existing code that serves the same purpose under different wording.
  `plaesy analyze` builds the index and `--index` rebuilds it alone; a query
  refuses to run against an index older than the source it was built from.
  Also registered at top level as `plaesy index` (build/update the index) and
  `plaesy query` (search it directly), for scripting without the `search`
  prefix.
- `plaesy stack` — inspect a target project's technology stack, and install the
  instruction files that stack selects.
- `plaesy stats` — show search index and reach cache statistics.
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

- 1052 build artifacts were tracked in git: `scripts/.plaesy/` (this module's
  own self-test dogfood output — 1051 files, 509 of them binary `.gob.gz`
  embedding shards) and `scripts/plaesy` (a 34 MB binary). `.gitignore`'s
  `!.plaesy/` negation was unanchored, so it re-included every directory
  named `.plaesy` at any depth, not just the project root — and the nested
  one had no re-ignore rules of its own. `scripts/plaesy` was never covered
  by any rule at all (`/plaesy` anchors to the repo root only).
  `TestNoBuildArtifactIsTracked` caught both. Fixed: `!.plaesy/` →
  `!/.plaesy/` (root-anchored) plus an explicit unanchored `.plaesy/` rule so
  every *other* `.plaesy/` stays ignored, and a new anchored `/scripts/plaesy`
  rule. `git rm --cached` on both; files remain on disk, just untracked.
- CI and `go build` failed with compile errors in
  `internal/search/embedder/onnx.go`. Bumping `go.mod`'s
  `github.com/yalue/onnxruntime_go` from v1.24.0 (pinned) to v1.36.0 (the
  version the code appeared written for) did not fix it: neither version's
  `SessionOptions` is a struct with public fields — it's built via
  `NewSessionOptions()` plus setter methods — and `NewTensor`/`Run` take a
  different shape entirely (`NewTensor(Shape, []T)` not
  `NewTensor([][]int64, []int64)`; `Run(inputs, outputs []Value) error` not
  a map returning `([]Tensor, error)`). The file also imported `strings`
  and `time` without using either, confirming it had never compiled on a
  non-Windows target. Rewrote the session setup, tensor creation, and
  inference call against the real v1.36.0 API.
- Release workflow failed on every target with `onnxruntime_go: build
  constraints exclude all Go files` — that dependency is cgo-only (every
  non-test file does `import "C"`, no non-cgo fallback), and
  `release.yml` cross-compiled all 5 targets from one runner with
  `CGO_ENABLED=0`. Rebuilt the workflow as a 5-way matrix building each
  target natively (cgo on by default) on a runner matching its OS/arch; see
  `.plaesy/decisions/release-native-matrix-cgo.md`.
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
