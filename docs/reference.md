# CLI Reference

Complete reference for every `plaesy` subcommand. Flags and arguments are
sourced verbatim from `scripts/cmd/plaesy/*.go`.

## Root

```
plaesy [command] [flags]
```

Global flags: `-h, --help` (help for any command).

## Command Index

| Command | Short Description | Source File |
|---|---|---|
| `analyze` | Generate AI-optimized project analysis | `scripts/cmd/plaesy/analyze.go` |
| `check-task-prerequisites` | Validate feature has plan.md | `scripts/cmd/plaesy/prereq.go` |
| `clean` | Remove Plaesy framework files | `scripts/cmd/plaesy/clean.go` |
| `config` | Platform configuration management | `scripts/cmd/plaesy/config.go` |
| `create-new-feature` | Create feature branch + spec.md | `scripts/cmd/plaesy/create_feature.go` |
| `detect-stack` | List relevant instruction files | `scripts/cmd/plaesy/detect_stack.go` |
| `generate-image` | Generate image via provider API | `scripts/cmd/plaesy/generate_image.go` |
| `get-feature-paths` | Print current feature paths | `scripts/cmd/plaesy/feature_paths.go` |
| `graph` | Build/query knowledge graph | `scripts/cmd/plaesy/graph.go` |
| `init` | Scaffold new Plaesy project | `scripts/cmd/plaesy/init.go` |
| `inject-ai-headers` | Inject AI front-matter headers | `scripts/cmd/plaesy/inject_ai_headers.go` |
| `install` | Install binary to bin dir | `scripts/cmd/plaesy/install.go` |
| `repair` | Repair installation (stub) | `scripts/cmd/plaesy/install.go` |
| `status` | Show CLI version + environment | `scripts/cmd/plaesy/status.go` |
| `task-manage` | Task lifecycle management | `scripts/cmd/plaesy/task_manage.go` |
| `trim` | Token/context compression | `scripts/cmd/plaesy/trim.go` |
| `uninstall` | Remove installed binary | `scripts/cmd/plaesy/install.go` |
| `update-agent-context` | Sync agent context with plan.md | `scripts/cmd/plaesy/update_agent.go` |
| `upgrade` | Upgrade installation (stub) | `scripts/cmd/plaesy/install.go` |
| `validate-docx` | Validate .docx OOXML | `scripts/cmd/plaesy/validate.go` |
| `validate-memory` | Scan memory for external refs | `scripts/cmd/plaesy/validate.go` |
| `validate-pptx` | Validate .pptx OOXML | `scripts/cmd/plaesy/validate.go` |
| `validate-xlsx` | Validate .xlsx OOXML | `scripts/cmd/plaesy/validate.go` |

---

## `analyze`

```
plaesy analyze [project_path] [flags]
```

Generate AI-optimized project analysis under `.plaesy/analysis/`.

| Flag | Default | Description |
|---|---|---|
| `--no-graph` | `false` | Skip dependency graph build entirely |
| `--force` | `false` | Force full regeneration (bypass fingerprint check) |
| `--if-changed` | `false` | Hidden no-op — fingerprint fast path is always on |

**Example**: `plaesy analyze` — analyzes the current directory.

---

## `check-task-prerequisites`

```
plaesy check-task-prerequisites [--json]
```

Validate the current feature has a `plan.md` and report available design docs.

| Flag | Default | Description |
|---|---|---|
| `--json` | `false` | Output as JSON |

**Example**: `plaesy check-task-prerequisites`

---

## `clean`

```
plaesy clean [TARGET_DIR] [flags]
```

Remove Plaesy framework files and directories.

| Flag | Default | Description |
|---|---|---|
| `--yes`, `-y` | `false` | Auto-confirm deletion |
| `--dry-run` | `false` | Show what would be removed |
| `--level` | `safe` | Cleanup level: `safe`, `thorough`, `complete` |
| `--ai` | `""` | AI platform to clean (auto-detected if omitted) |
| `--verbose` | `false` | Show detailed progress |
| `--config` | `""` | Path to platform.json |
| `--backup` | `true` | Create backup before removal |
| `--no-backup` | `false` | Skip backup creation |

Levels: `safe` (framework files only), `thorough` (framework + specs),
`complete` (everything Plaesy-related — DANGEROUS).

**Example**: `plaesy clean --dry-run --level thorough`

---

## `config`

```
plaesy config [subcommand] [flags]
```

Centralized platform configuration management using
`scripts/configs/platform.json`.

Global flag: `--config` (path to platform.json; default: auto-resolved).

### Subcommands

| Subcommand | Args | Description |
|---|---|---|
| `detect-platform` | — | Detect current AI platform |
| `list-platforms` | — | List all available platforms |
| `get-platform-config` | `<platform> <key>` | Get platform field or mapping value |
| `get-mapping-value` | `<section> <mapping_type>` | Get mapping value with value/excludes structure |
| `get-mapping-excludes` | `<section> <mapping_type>` | Get exclude patterns for a mapping |
| `get-clean-files` | `[platform]` | Get files to clean for platform |
| `get-clean-dirs` | `[platform]` | Get directories to clean for platform |
| `get-plaesy-structure` | `<component>` | Get Plaesy structure config |
| `show-platform-info` | `[platform]` | Show detailed platform information |
| `validate` | — | Validate platform configuration |

**Example**: `plaesy config detect-platform`

---

## `create-new-feature`

```
plaesy create-new-feature <feature_description> [--json]
```

Create a new feature branch, directory structure, and `spec.md` from template.
Feature numbering: `NNN-slug` (next number auto-detected from `specs/`).

| Flag | Default | Description |
|---|---|---|
| `--json` | `false` | Output as JSON |

**Example**: `plaesy create-new-feature "Add dark mode toggle"`

---

## `detect-stack`

```
plaesy detect-stack [target-dir] [--plaesy-root <path>]
```

List instruction files relevant to a target project's tech stack, using
`instructions/mapping.json` as the keyword registry.

| Flag | Default | Description |
|---|---|---|
| `--plaesy-root` | `""` | Path to Plaesy repo root (`$PLAESY_ROOT`, then git root) |

**Example**: `plaesy detect-stack .` — prints one instruction filename per line.

---

## `generate-image`

```
plaesy generate-image --prompt <text> --out <path> [--provider <openai|gemini>] [--size <WxH>]
```

Generate an image asset via a configured provider API.

| Flag | Default | Description |
|---|---|---|
| `--prompt` | `""` | Image prompt text (required) |
| `--provider` | `""` | Image provider: `openai` (default) or `gemini` |
| `--size` | `""` | Image size, e.g. `1024x1024` |
| `--out` | `""` | Output file path (required) |

**Example**: `plaesy generate-image --prompt "a logo" --out logo.png --provider openai`

---

## `get-feature-paths`

```
plaesy get-feature-paths
```

Print shell-sourceable paths (`REPO_ROOT=`, `FEATURE_DIR=`, etc.) for the
current feature branch. No arguments. Reads from `internal/featurepath`
(`scripts/internal/featurepath/prereq.go`).

**Example**: `eval $(plaesy get-feature-paths)`

---

## `graph`

```
plaesy graph [project_path] [flags]
```

Build (or query) a lightweight knowledge graph of a project. Scans
source/doc/config files, extracts structural references + mentions + symbols,
runs community detection, and emits `project.graph.json`, `project.html`, and
`reports.md`.

| Flag | Default | Description |
|---|---|---|
| `--path` | `.` | Project directory to scan |
| `--outdir` | `.plaesy/analysis` | Output directory (relative to `--path`) |
| `--query` | `""` | Keyword query against node IDs |
| `--fuzzy-query` | `""` | Fuzzy/substring search against node IDs |
| `--search-depth` | `2` | Search depth for `--fuzzy-query` |
| `--path-query-from` | `""` | Shortest-path: source node |
| `--path-query-to` | `""` | Shortest-path: target node |
| `--explain` | `""` | Explain one node's edges and symbols |
| `--impact-check` | `""` | List nodes impacted by changing this node |
| `--impact-depth` | `2` | BFS depth for `--impact-check` |
| `--impact-visualize` | `false` | Write `impact-visualization.html` |
| `--semantic-queue` | `false` | Export inferred edges to `semantic-queue.json` |
| `--business-logic` | `false` | Export inferred edges to `business-logic-queue.json` |
| `--generate-paths` | `false` | Write `learning-paths.json` |
| `--apply-semantic` | `""` | Merge annotations file into `reports.md` |
| `--watch` | `false` | Rebuild automatically on source changes |
| `--watch-interval` | `3` | Poll interval (seconds) for `--watch` |
| `--if-changed` | `false` | Rebuild only if source changed since last build |

**Examples**:
```bash
plaesy graph                       # build graph for current dir
plaesy graph --path /path/to/proj  # build for specific project
plaesy graph --query "README"      # query the graph
plaesy graph --watch               # watch mode (3s poll)
```

---

## `init`

```
plaesy init [directory] [--ai <platform>] [--plaesy-home <path>]
```

Scaffold a new Plaesy project: `.plaesy/` structure + AI platform files.

| Flag | Default | Description |
|---|---|---|
| `--ai` | `""` | AI platform (e.g. `claude_code`, `cursor_ai`) |
| `--plaesy-home` | `""` | Override Plaesy repo root (`PLAESY_HOME` env var) |

Supported platforms: `claude_code`, `cursor_ai`, `github_copilot`,
`windsurf_ai`, `cline`, `deepseek`, `kilo_code`, `qoder`, `trae_ai`,
`continue_dev`, `tabnine`, `codeium`, `codewhisperer`, `studio_bot`,
`replit_ghostwriter`, `llama_index`, `ollama`, `lm_studio`, `generic_ai`.

**Examples**:
```bash
plaesy init my-project --ai claude_code
plaesy init . --ai cursor_ai
```

---

## `inject-ai-headers`

```
plaesy inject-ai-headers --ai <platform> --target <dir> [flags]
```

Inject platform-specific YAML front-matter headers into prompt/chatmode/
instructions files.

| Flag | Default | Description |
|---|---|---|
| `--ai` | `""` | AI platform (required: `copilot`, `cursor`, `windsurf`, `claude`, `chatgpt`, `gemini`, `trae-ai`, `qwen-code`, `codex-cli`, `opencode-cli`, `local-ai`, `manual`) |
| `--target` | `""` | Target directory (required) |
| `--headers-dir` | `""` | Header YAML files dir (default: `<repo-root>/templates/ai-headers`) |
| `--dry-run` | `false` | Show changes without writing |
| `--force` | `false` | Overwrite existing headers |
| `--backup` | `false` | Create `.backup.<timestamp>` copy |
| `--merge` | `false` | Merge into existing front-matter |
| `--list-only` | `false` | List files + header mapping, then exit |
| `--pattern` | (default globs) | Glob patterns (repeatable) |
| `--exclude` | — | Exclude glob (repeatable) |

**Example**: `plaesy inject-ai-headers --ai claude --target prompts --backup`

---

## `install`

```
plaesy install
```

Install the `plaesy` CLI to a well-known bin directory. Copies the currently-
running binary to `%LOCALAPPDATA%\Plaesy\bin\plaesy.exe` (Windows) or
`$HOME/.local/bin/plaesy` (Unix). If the directory is not on PATH, prints
instructions for adding it (does not modify shell rc or registry).

**No flags.** Reads its own executable path via `os.Executable()`.

---

## `repair`

```
plaesy repair
```

Print "not yet implemented" message. See `docs/scripts/install.md` for upgrade
instructions.

---

## `status`

```
plaesy status
```

Show Plaesy CLI version and environment status. Reports:
- git availability
- resolved git repository root (uses `GetRepoRoot()` with Windows path normalization)
- Banner with version (`internal/common.Version`, set at build time)

---

## `task-manage`

```
plaesy task-manage [subcommand]
```

Task lifecycle management for `.plaesy/tasks/{backlog,todo,doing,done,blocked}/`.

### Subcommands

| Subcommand | Args | Description |
|---|---|---|
| `list` | — | List all tasks across every status |
| `list-status` | `<status>` | List tasks in one status |
| `next` | — | Show the next task to work on |
| `start` | `<task>` | Move task to `doing` |
| `complete` | `<task>` | Move task to `done` |
| `block` | `<task>` | Move task to `blocked` |
| `unblock` | `<task>` | Unblock a task |
| `move` | `<task_file> <from> <to>` | Move between arbitrary statuses |
| `show` | `<task_file> [status]` | Display task content |

Valid statuses: `backlog`, `todo`, `doing`, `done`, `blocked`.

---

## `trim`

```
plaesy trim [subcommand] [flags]
```

Token/context compression for command output, memory, and instruction files.

### Subcommands

| Subcommand | Short |
|---|---|
| `run -- <command...>` | Layer 1: run a command and compress its output |
| `compress --path <file\|dir>` | Layer 2: fast heuristic prose compression |
| `llm-queue --path <file>` | Export prose segments for AI rewriting |
| `apply-llm --path <file> --annotations <json>` | Merge rewritten segments back |
| `report` | Cumulative savings across layers |

### `trim compress` flags

| Flag | Default | Description |
|---|---|---|
| `--path` | `""` | File or directory to compress |
| `--level` | `""` | `lite`, `full`, `ultra` (auto-resolved from graph when omitted) |
| `--recurse` | `false` | Recurse into subdirectories |
| `--dry-run` | `false` | Report savings without writing |

### `trim run`

Uses `DisableFlagParsing` — passes all args after `--` to the target command,
then compresses output.

**Example**: `plaesy trim run -- echo "hello world"`

---

## `uninstall`

```
plaesy uninstall
```

Remove the installed `plaesy` binary. Prompts `y/N` (default N). Does not
remove any project's `.plaesy/` directory — only the binary itself.

---

## `update-agent-context`

```
plaesy update-agent-context [claude|gemini|copilot|cursor|qwen|opencode]
```

Sync agent context files with the current feature's `plan.md`. Auto-detects
platform context files when no argument is given. Delegates to
`internal/agentcontext` (`scripts/internal/agentcontext/update.go`).

---

## `upgrade`

```
plaesy upgrade
```

Print "not yet implemented" message. Build a new binary from source and re-run
`plaesy install` (see `docs/scripts/install.md`).

---

## `validate-memory`

```
plaesy validate-memory
```

Scan `.plaesy/memory/*.md` for external, non-self-contained references
(`~/.claude/`, `/tmp/`, `C:\Users\...\.claude`). Reports files, patterns, and
line numbers. Exits non-zero if issues found.

---

## `validate-docx` / `validate-pptx` / `validate-xlsx`

```
plaesy validate-docx <file.docx>
plaesy validate-pptx <file.pptx> [expected_slide_count]
plaesy validate-xlsx <file.xlsx> [expected_sheet_name...]
```

Validate OOXML structure (`.docx`/`.pptx`/`.xlsx`) without rendering. Reports
paragraph/table/slide/sheet counts and structural issues.

---

## Installation Commands

| Command | Purpose | Source |
|---|---|---|
| `plaesy install` | Copy binary to well-known bin dir | `scripts/cmd/plaesy/install.go` + `scripts/internal/installer/installer.go` |
| `plaesy status` | Show install location + PATH + version | `scripts/cmd/plaesy/status.go` |
| `plaesy uninstall` | Remove installed binary (y/N prompt) | `scripts/cmd/plaesy/install.go` |
| `plaesy repair` | Not yet implemented | `scripts/cmd/plaesy/install.go` |
| `plaesy upgrade` | Not yet implemented | `scripts/cmd/plaesy/install.go` |
