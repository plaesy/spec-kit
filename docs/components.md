# Components

> **Component count**: 16 internal packages — one H2 per package below, so the
> count is verifiable by reading the file. An earlier "17" here did not match the
> 16 documented entries; that and a prior "15" are both gone. The "at the threshold
> for flat-file layout" note pointed at a rule in `docs/architecture.md` that does
> not exist — an unmeasurable claim, so it was removed rather than re-pointed.
> Each entry lists purpose, key types/functions, and implementing paths
> (`file:line`).
>
> `internal/quality` holds no runtime code; it exists so the repository's own gates
> have a home and run in the same `go test ./...` as everything else.
>
> To re-check the count: `ls -d scripts/internal/*/ | wc -l`.

## `internal/common`

Shared infrastructure used by every command: colored logging (console + optional
file via `PLAESY_LOG_FILE`), git-repo resolution, and environment validation.

- **`GetRepoRoot()`** — `git rev-parse --show-toplevel` with Windows MSYS2 path
  normalization (`scripts/internal/common/gitpaths.go:32`).
- **`GetFeaturePaths()`** — resolves `REPO_ROOT`, current branch, and feature
  spec/plan/tasks paths (`scripts/internal/common/gitpaths.go:77`).
- **`LogInfo`/`LogSuccess`/`LogWarning`/`LogError`/`LogDebug`** — leveled logging
  with optional file output (`scripts/internal/common/logging.go:53-72`).
- **`ValidateCommandExists`/`ValidateFileExists`** — environment checks
  (`scripts/internal/common/validate.go:10-49`).
- **`Version`** — set at build time via `-ldflags`; falls back to `0.0.0`
  (`scripts/internal/common/version.go:10`).

**Implementing path**: `scripts/internal/common/`

## `internal/config`

Loads and queries `scripts/configs/platform.json` — the centralized platform
configuration that maps AI platforms (Claude Code, Cursor, Copilot, etc.) to
their file layout (core, instructions, prompts, agents destinations).

- **`Load(path)`** — parses platform.json (`scripts/internal/config/config.go:82`).
- **`DetectPlatform(path)`** — first platform whose detection files exist in CWD
  (`scripts/internal/config/config.go:218`).
- **`GetMappingValue`/`GetMappingExcludes`** — reads plaesy.mapping sections
  (`scripts/internal/config/config.go:197-211`).
- **`GetCleanDirs`** — derives cleanup target directories for a platform
  (`scripts/internal/config/config.go:262`).
- **`NormalizePlatform`** — resolves a user-typed platform name (a shorthand, an
  id, or a display name) to the id `platform.json` declares. Shared with
  `internal/scaffold` so `plaesy init --ai claude` and `plaesy clean --ai claude`
  cannot mean different things (`scripts/internal/config/config.go`).
- **`HasPlatform` / `PlatformNames`** — the existence check and sorted id list
  that pair with `NormalizePlatform`; what an unknown-platform error lists.

**Implementing path**: `scripts/internal/config/config.go`

## `internal/scaffold`

Ports `plaesy init` (formerly `plaesy-init.sh`): creates the `.plaesy/` directory structure and copies
platform-specific files. Driven by `platform.json`.

- **`Init(opts)`** — orchestrates full initialization: resolve root → load config
  → normalize platform → create structure → copy platform files
  (`scripts/internal/scaffold/init.go:28`).
- **`CreateStructure`** — creates `.plaesy/` core directories + copies
  instructions/prompts/agents/templates
  (`scripts/internal/scaffold/structure.go`).
- **`SetupPlatformConfig`** — copies platform-specific core + prompt files
  (`scripts/internal/scaffold/platform.go:17`).
- **`copyFile`** — the shared copy primitive: skips an existing destination, and
  carries the source's permission bits over so a `0755` script stays executable
  (`scripts/internal/scaffold/copy.go:71`).

**Implementing path**: `scripts/internal/scaffold/`

## `internal/detectstack`

Ports `plaesy stack detect` (formerly `detect-stack.sh/.ps1`): maps a target project to relevant
`instructions/*.instructions.md` files using `instructions/mapping.json` as the
keyword registry.

- **`Detect(targetDir, plaesyRoot)`** — single-pass walk collecting
  always_load entries + keyword/extension/filename matches
  (`scripts/internal/detectstack/detectstack.go:108`).
- **`wordBoundaryRE(keyword)`** — case-insensitive `\b` regex per keyword to
  avoid false matches (e.g., "java" in "javascript")
  (`scripts/internal/detectstack/detectstack.go:50`).

**Implementing path**: `scripts/internal/detectstack/detectstack.go`

## `internal/analyze`

Ports `plaesy analyze` (formerly `plaesy-analyze.sh`): comprehensive project analysis generating
`project.json`, `project.structure.json`, `overview.md`, and (optionally) the
dependency graph.

- **`Run(opts)`** — entry point: resolve paths → fingerprint check → walk →
  detect → generate → build graph (`scripts/internal/analyze/analyze.go:52`).
- **`shouldSkipAnalysis`** — fingerprint-based fast path: skips regeneration
  when the project fingerprint matches the last run
  (`scripts/internal/analyze/analyze.go:84`).
- **`detectAllFrameworks`/`detectAllLanguages`/`detectBuildSystems`** —
  technology detection from manifest files + source extensions
  (`scripts/internal/analyze/detect.go`).
- **`generateProjectJSON`/`generateProjectStructureJSON`/`generateOverviewMD`** —
  output generators (`scripts/internal/analyze/insights.go`,
  `scripts/internal/analyze/output.go`).

**Implementing path**: `scripts/internal/analyze/`

## `internal/graph`

Ports `plaesy graph` (formerly `plaesy-graph.sh`): lightweight knowledge-graph builder with query,
explain, path-query, impact-check, semantic-queue, and watch modes.

- **`Build(opts, paths)`** — scans project files, extracts references +
  mentions + symbols, runs community detection
  (`scripts/internal/graph/build.go`).
- **`SourceFingerprint`** — content hash of all scanned files; used for
  `--if-changed` / `--watch` incremental rebuilds
  (`scripts/internal/graph/build.go:186`).
- **`DoQuery`/`DoFuzzyQuery`/`DoExplain`/`DoPathQuery`/`DoImpactCheck`** —
  graph query operations (`scripts/internal/graph/query.go`).
- **`WriteReport`/`WriteHTML`** — renders `reports.md` (plain-language report)
  and `project.html` (force-directed visualization)
  (`scripts/internal/graph/render.go`).

**Implementing path**: `scripts/internal/graph/`

## `internal/cleaner`

Ports `plaesy clean` (formerly `plaesy-clean.sh`): safe removal of Plaesy framework files with three
levels (safe/thorough/complete), platform detection, backup, and dry-run
support.

- **`Level` / `ValidLevel`** — cleanup level enum (safe, thorough, complete)
  (`scripts/internal/cleaner/cleaner.go:20-34`).
- **`Options`** — target dir, auto-confirm, dry-run, backup, level, AI platform
  selection (`scripts/internal/cleaner/cleaner.go:45`).
- **`New(opts, configPath)`** — constructs cleaner with loaded config
  (`scripts/internal/cleaner/cleaner.go`).
- **`DetectAllPlatforms`** — scans target dir for all installed AI platform
  markers (`scripts/internal/cleaner/cleaner.go`).
- **The level decides what is deleted.** `safe` removes `.plaesy` and the
  platform's mapped targets and keeps `specs/`; `thorough` adds `specs/`;
  `complete` adds the `generic_ai` fallback's own files (`AI-INSTRUCTIONS.md`,
  `ai-config/`). The level used to be read only for the plan's banner, so every
  level deleted the same set and the default level destroyed feature documents.
  `CreateBackup` covers exactly what the chosen level will remove.

**Implementing path**: `scripts/internal/cleaner/`

## `internal/installer`

Manages the `plaesy` binary's lifecycle on the system (v1 model: the running
binary is the deliverable).

- **`InstallDir()`** — per-OS well-known bin directory:
  `%LOCALAPPDATA%\Plaesy\bin` (Windows) or `$HOME/.local/bin` (Unix)
  (`scripts/internal/installer/installer.go:33`).
- **`Install()`** — copies the running executable to `InstallDir`, with
  temp-file + rename for safe atomic replacement
  (`scripts/internal/installer/installer.go:123`).
- **`CollectStatus(version)`** — reports install dir, install path, whether
  installed, whether on PATH, and running-from path
  (`scripts/internal/installer/installer.go:198`).
- **`Uninstall()`** — removes the installed binary
  (`scripts/internal/installer/installer.go:246`).
- **`OnPath(dir)`** — case-insensitive PATH check on Windows, exact on Unix
  (`scripts/internal/installer/installer.go:74`).

**Implementing path**: `scripts/internal/installer/installer.go`

## `internal/trimmer`

Ports `plaesy trim` (formerly `plaesy-trim.sh`): token/context compression in two layers.

- **Layer 1** (`trim run`): runs a command, compresses its output.
- **Layer 2** (`trim compress`): heuristic prose compression of files
  (lite/full/ultra levels resolved from graph when omitted).
- **Layer 2 LLM mode** (`trim llm-queue` / `trim apply-llm`): exports prose
  segments for AI rewriting, then merges results back.
- **`Rules`** — loaded from `plaesy-trim.rules` in the framework root
  (`scripts/internal/trimmer/rules.go`).

**Implementing path**: `scripts/internal/trimmer/`

## `internal/featurepath`

Ports `plaesy features create` / `plaesy features paths` /
`plaesy features validate` (formerly `create-new-feature.sh`,
`get-feature-paths.sh`, and `check-task-prerequisites.sh`): feature-branch workflow helpers.

- **`CreateNewFeature(desc)`** — computes next NNN-slug number, creates git
  branch, seeds `.plaesy/specs/<branch>/spec.md` from template
  (`scripts/internal/featurepath/create.go:33`).
- **`GetPathsReport()`** — prints shell-sourceable `REPO_ROOT=` /
  `FEATURE_DIR=` / etc. (`scripts/internal/featurepath/prereq.go`).
- **`CheckTaskPrerequisites()`** — validates current feature has `plan.md`,
  reports available design docs (`scripts/internal/featurepath/prereq.go`).

**Implementing path**: `scripts/internal/featurepath/`

## `internal/taskmanage`

Ports `plaesy tasks` (formerly `plaesy-task-manage.sh`): task lifecycle management for
`.plaesy/tasks/{backlog,todo,doing,done,blocked}/`.

- **`Statuses`** — the 5 valid status directories in order
  (`scripts/internal/taskmanage/taskmanage.go:17`).
- **`FindProjectRoot()`** — walks up from CWD looking for `.plaesy/`
  (`scripts/internal/taskmanage/taskmanage.go:31`).
- **`ListTasks` / `MoveTask` / `StartTask` / `CompleteTask` / `BlockTask` /
  `UnblockTask`** — task CRUD operations (`scripts/internal/taskmanage/taskmanage.go`).
- **`ReadTask`** — reads a task file by name, resolving status directory
  (`scripts/internal/taskmanage/taskmanage.go`).

**Implementing path**: `scripts/internal/taskmanage/`

## `internal/agentcontext`

Ports `plaesy context update` (formerly `update-agent-context.sh`): syncs
**existing** AI agent context files with the current feature's `plan.md`. It patches
only — the agent-file template was removed on 2026-09-25, so a missing context file
is an error rather than a scaffold.

- **`Update(agentType)`** — patches platform-specific agent context; updates every
  detected context file when `agentType` is empty, and exits non-zero if none exist
  (`scripts/internal/agentcontext/update.go:92`).

**Implementing path**: `scripts/internal/agentcontext/`

## `internal/validate`

Backs `plaesy validate memory` / `constitution` / `docx` / `pptx` / `xlsx` (all
formerly `plaesy-validate-memory.sh`, `validate-docx.sh`, etc., and all
originally six top-level commands before they were grouped under the single
`validate` verb): file format, self-containment, and constitution-contract
validation.

- **`Memory(repoRoot)`** — scans `.plaesy/memory/*.md` for external-path
  references (`~/.claude/`, `/tmp/`, `C:\Users\...` pattern), reporting
  matched lines per file (`scripts/internal/validate/memory.go:57`).
- **`Constitution(path)`** — validates a generated constitution: frontmatter
  completeness, ISO dates, no unfilled `{{ }}` placeholders, a coherent §1
  active-dimension scope (frontmatter must agree with the table), frontmatter
  version vs amendment log, unique rule IDs. Returns the declared version,
  active dimensions, and rule counts alongside the aggregated failure list
  (`scripts/internal/validate/constitution.go`).
- **`Docx`/`Pptx`/`Xlsx`** — OOXML structural validation: paragraph/table/slide
  counts, slide title placeholders, sheet names (`scripts/internal/validate/`).

**Implementing path**: `scripts/internal/validate/`

## `internal/mdlint`

Backs `plaesy validate markdown`. A Markdown linter on goldmark that replaced the
npm `markdownlint-cli2` + `@github/markdownlint-github` setup on 2026-09-25, so
this repository and every project that installs the framework can lint Markdown
with the same binary and without Node. Rule IDs and config keys are markdownlint's,
so a `.markdownlint.json` stays portable; 25 rules are implemented, including the
three GitHub accessibility rules (GHA001-GHA003) that only the npm ruleset had.
`RuleIDs`/`RuleDescription` back `--list-rules`, which exists so that "no
violations" is never mistaken for "markdownlint would be happy".

- **`LoadConfig`/`LoadBaseline`/`WriteBaseline`** — config loading that treats
  unknown rule names, unknown option keys, and `extends` as hard errors (a config
  that half-loads is a check that checks nothing); baseline loading that rejects a
  zero ceiling, since "no violations" is expressed by passing no baseline.
- **`Lint(root, paths, cfg, _)`** — walks Markdown, skips `DefaultIgnoreDirs`,
  resolves rule options, and returns per-file detail plus a per-rule histogram.
- **`Fix(root, paths, cfg, _)`** — applies the machine-fixable rules in place.
  Each spacing rule builds its plan from the same predicates its check reports on
  (`fixMD022`, `fixMD031`, `fixMD032`, `fixMD058`), and the plan is recomputed
  between the "insert above" and "insert below" passes because line indexes move
  under the first one. `fix_test.go` lints the output of every pass and fails if a
  fixable rule still reports. That test exists because the alternative was invisible:
  the fixer used to answer "does this line continue the block above?" by inspecting
  the wrong line, so `--fix` printed `[OK] nothing to fix` on a file the linter had
  just listed 20 fixable violations for, and a document is equally clean before and
  after a fix that does nothing.
- **`Result.Exceeds(baseline)`** — the ratchet: a run passes at or below the
  recorded ceiling and fails when the total rises. `TestRepoMarkdownStaysWithinRatchet`
  runs the same check inside `go test ./...`, the only CI job. A `nil` baseline is
  `--no-baseline`, which promises zero, so it is a ceiling of 0 rather than the
  absence of one — reading it as "nothing to compare" made the flag print its own
  violations and exit 0.
- **Fence and list state** — `doc.fenceOpens` distinguishes the fence that starts
  a block from the one that ends it, and MD029 compares an item's number with its
  position in its own list. Three rules (MD031, MD040, MD029) reported 1,757
  violations repo-wide that no markdownlint rule reports, and `--fix` was inserting
  a blank line *inside* code blocks to satisfy one of them.

**Implementing path**: `scripts/internal/mdlint/`

## `internal/quality`

No runtime code — the package exists to hold the repository's own quality gates, the
checks that run inside `go test ./...` and therefore inside the only CI job. Before
this, the coverage bar lived in the constitution as four hand-copied percentages, and
nothing failed when the code moved: 4 of 17 packages were listed, the listed number
for `cmd/plaesy` was stale, and 11 packages sat at 0.0% with no bar at all.

- **`TestCoverageDoesNotRegress`** — measures every package in the module with its own
  `go test -cover -count=1` run and compares the result against
  `scripts/coverage-baseline.json`. It fails on four separate conditions, each with a
  different fix: a package below its recorded number (add tests, or accept a lower bar
  deliberately), a package with no entry (record the measured number, and decide
  whether it deserves tests), an entry for a package that no longer exists (delete
  the claim), and an entry the tree has outgrown by more than 2 points (re-measure —
  a too-low bar is indistinguishable from a correct one, so nothing else would notice
  it going stale). Packages without test files measure 0.0% and stay in the file on
  purpose: excluding them would make a large new untested package invisible to the
  gate. `-update-coverage-baseline` re-records every number from a fresh measurement.
- **`compareCoverage`** — the comparison as a pure function, so every failure class is
  proven by fast deterministic tests rather than being discovered the day a package
  quietly loses its tests. Improvement is reported and never fails; a 0.05pp epsilon
  absorbs the one-decimal rounding `go test` prints, because a gate that trips on
  rounding noise gets switched off.
- **`TestCoverageBaselineIsWellFormed`** — guards the guard: a value outside 0-100 or a
  missing `measured_at` fails as a malformed baseline rather than as a regression
  nobody can act on.
- **Measurement caveat** — `go test ./... -cover` prints 25.1% for `cmd/plaesy` while
  the profile it writes for that same run says 22.9%. The baseline records the number
  the profile backs, measured one package at a time; the reasoning is in the
  `coverage-baseline.json` header.

The package excludes itself from the measurement: measuring the ratchet means running
it, and a test that runs itself recurses.

**Implementing path**: `scripts/internal/quality/`

## `internal/imagegen`

Ports `plaesy images create` (formerly `generate-image.sh`): generates image assets via OpenAI or Gemini APIs
using only the Go standard library.

- **`Generate(opts)`** — calls provider image-generation API, decodes base64
  response, writes to `opts.Out` (`scripts/internal/imagegen/imagegen.go:30`).
- **`Options`** — prompt, provider ("openai" or "gemini"), size, output path
  (`scripts/internal/imagegen/imagegen.go:18-23`).
- The tests fake both the HTTP transport and the API key, and assert that the fake
  key is the one that reached the request. A suite for a package that calls a paid
  API has to be able to prove it cannot spend money.

**Implementing path**: `scripts/internal/imagegen/imagegen.go`
