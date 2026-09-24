# Components

> **Component count**: 15 internal packages — at the threshold for flat-file
> layout (see `docs/architecture.md`). Each entry lists purpose, key types/functions,
> and implementing paths (`file:line`).

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
their file layout (core, instructions, prompts, chatmodes destinations).

- **`Load(path)`** — parses platform.json (`scripts/internal/config/config.go:82`).
- **`DetectPlatform(path)`** — first platform whose detection files exist in CWD
  (`scripts/internal/config/config.go:218`).
- **`GetMappingValue`/`GetMappingExcludes`** — reads plaesy.mapping sections
  (`scripts/internal/config/config.go:197-211`).
- **`GetCleanDirs`** — derives cleanup target directories for a platform
  (`scripts/internal/config/config.go:262`).

**Implementing path**: `scripts/internal/config/config.go`

## `internal/scaffold`

Ports `plaesy init` (formerly `plaesy-init.sh`): creates the `.plaesy/` directory structure and copies
platform-specific files. Driven by `platform.json`.

- **`Init(opts)`** — orchestrates full initialization: resolve root → load config
  → normalize platform → create structure → copy platform files
  (`scripts/internal/scaffold/init.go:28`).
- **`CreateStructure`** — creates `.plaesy/` core directories + copies
  instructions/prompts/chatmodes/templates
  (`scripts/internal/scaffold/structure.go`).
- **`SetupPlatformConfig`** — copies platform-specific core + prompt files
  (`scripts/internal/scaffold/platform.go:17`).

**Implementing path**: `scripts/internal/scaffold/`

## `internal/detectstack`

Ports `plaesy detect-stack` (formerly `detect-stack.sh/.ps1`): maps a target project to relevant
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

## `internal/aiheaders`

Ports `inject-ai-headers.sh`: injects platform-specific YAML front-matter
headers into prompt/chatmode/instructions files.

- **`ValidPlatforms`** — 12 supported platform names: copilot, cursor, windsurf,
  claude, chatgpt, gemini, trae-ai, qwen-code, codex-cli, opencode-cli,
  local-ai, manual (`scripts/internal/aiheaders/aiheaders.go:22-25`).
- **`Run(o *Options)`** — full injection workflow: load header YAML → find
  prompt files → inject/merge headers (`scripts/internal/aiheaders/aiheaders.go:582`).
- **`FindPromptFiles`** — walks target dir matching `*.prompt.md`,
  `*.chatmode.md`, `*.instructions.md` (`scripts/internal/aiheaders/aiheaders.go:129`).
- **`HasHeader`/`GetFrontMatterKey`** — YAML front-matter detection + key
  extraction (`scripts/internal/aiheaders/aiheaders.go:216-268`).

**Implementing path**: `scripts/internal/aiheaders/`

## `internal/featurepath`

Ports `plaesy create-new-feature` / `plaesy get-feature-paths` /
`plaesy check-task-prerequisites` (formerly `create-new-feature.sh`,
`get-feature-paths.sh`, and `check-task-prerequisites.sh`): feature-branch workflow helpers.

- **`CreateNewFeature(desc)`** — computes next NNN-slug number, creates git
  branch, seeds `specs/<branch>/spec.md` from template
  (`scripts/internal/featurepath/create.go:33`).
- **`GetPathsReport()`** — prints shell-sourceable `REPO_ROOT=` /
  `FEATURE_DIR=` / etc. (`scripts/internal/featurepath/prereq.go`).
- **`CheckTaskPrerequisites()`** — validates current feature has `plan.md`,
  reports available design docs (`scripts/internal/featurepath/prereq.go`).

**Implementing path**: `scripts/internal/featurepath/`

## `internal/taskmanage`

Ports `plaesy task-manage` (formerly `plaesy-task-manage.sh`): task lifecycle management for
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

Ports `plaesy update-agent-context` (formerly `update-agent-context.sh`): syncs AI agent context files with the current
feature's `plan.md`.

- **`Update(agentType)`** — writes/updates platform-specific agent context
  (`scripts/internal/agentcontext/update.go:92`).

**Implementing path**: `scripts/internal/agentcontext/`

## `internal/validate`

Ports `plaesy validate-memory` / `validate-docx` / `validate-pptx` /
`validate-xlsx` (formerly `plaesy-validate-memory.sh`, `validate-docx.sh`, etc.):
file format and self-containment validation.

- **`Memory(repoRoot)`** — scans `.plaesy/memory/*.md` for external-path
  references (`~/.claude/`, `/tmp/`, `C:\Users\...` pattern), reporting
  matched lines per file (`scripts/internal/validate/memory.go:57`).
- **`Docx`/`Pptx`/`Xlsx`** — OOXML structural validation: paragraph/table/slide
  counts, slide title placeholders, sheet names (`scripts/internal/validate/`).

**Implementing path**: `scripts/internal/validate/`

## `internal/imagegen`

Ports `plaesy generate-image` (formerly `generate-image.sh`): generates image assets via OpenAI or Gemini APIs
using only the Go standard library.

- **`Generate(opts)`** — calls provider image-generation API, decodes base64
  response, writes to `opts.Out` (`scripts/internal/imagegen/imagegen.go:30`).
- **`Options`** — prompt, provider ("openai" or "gemini"), size, output path
  (`scripts/internal/imagegen/imagegen.go:18-23`).

**Implementing path**: `scripts/internal/imagegen/imagegen.go`
