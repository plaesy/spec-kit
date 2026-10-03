# CLI Reference

Complete reference for every `plaesy` subcommand. Flags and arguments are
sourced verbatim from `scripts/cmd/plaesy/*.go`.

## Root

```bash

plaesy [command] [flags]
```

Global flags: `-h, --help` (help for any command).

A noun parent (`config`, `context`, `features`, `images`, `platforms`, `stack`,
`tasks`) prints its help when run with no subcommand, and rejects a misspelled
one with `unknown command "…" for "plaesy …"` and a non-zero exit. That
distinction is deliberate: help text printed at exit 0 is indistinguishable
from a result, so `PLATFORM=$(plaesy config detect)` would have captured a page
of help and the calling script would have carried on.

## Command Index

| Command | Short Description | Source File |
|---|---|---|
| `analyze` | Generate AI-optimized project analysis | `scripts/cmd/plaesy/analyze.go` |
| `clean` | Remove Plaesy framework files | `scripts/cmd/plaesy/clean.go` |
| `config` | Read and validate scripts/configs/platform.json | `scripts/cmd/plaesy/config.go` |
| `context` | Agent context files kept in sync with plan.md | `scripts/cmd/plaesy/context.go` |
| `features` | List, create, and inspect feature branches | `scripts/cmd/plaesy/features.go` |
| `graph` | Build/query knowledge graph | `scripts/cmd/plaesy/graph.go` |
| `images` | Generate image assets | `scripts/cmd/plaesy/images.go` |
| `init` | Scaffold new Plaesy project | `scripts/cmd/plaesy/init.go` |
| `install` | Install binary to bin dir | `scripts/cmd/plaesy/install.go` |
| `platforms` | AI platforms plaesy can target | `scripts/cmd/plaesy/platforms.go` |
| `repair` | Repair installation (stub) | `scripts/cmd/plaesy/install.go` |
| `reload` | Refresh generated `.plaesy/` files, keeping memory and context; `--prune` lists orphans in the tool-owned trees, `--prune-apply` removes them | `scripts/cmd/plaesy/reload.go` |
| `search` | Embedding-based semantic search over extracted symbols; `--index` builds the index, a query refuses a stale one | `scripts/cmd/plaesy/search.go` |
| `stack` | Inspect a target project's technology stack | `scripts/cmd/plaesy/stack.go` |
| `status` | Show CLI version + environment | `scripts/cmd/plaesy/status.go` |
| `tasks` | Task lifecycle management | `scripts/cmd/plaesy/tasks.go` |
| `trim` | Token/context compression | `scripts/cmd/plaesy/trim.go` |
| `uninstall` | Remove installed binary | `scripts/cmd/plaesy/install.go` |
| `upgrade` | Upgrade installation (stub) | `scripts/cmd/plaesy/install.go` |
| `validate` | Run checks: constitution, memory, Markdown, OOXML | `scripts/cmd/plaesy/validate.go` |

---

## `analyze`

```bash

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

---

## `clean`

```bash

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

Levels decide what is actually removed:

| Level | Removes |
|---|---|
| `safe` | `.plaesy/` and the platform's mapped `core`/`instructions`/`prompts`/`agents` targets. `specs/` is kept. |
| `thorough` | The above plus `specs/` — the feature documents `create feature` writes. |
| `complete` | The above plus the `generic_ai` fallback's own files (`AI-INSTRUCTIONS.md`, `ai-config/`). |

`--backup` is on by default and covers every directory the chosen level removes.
A dry run lists what will go and what will be preserved.

**Example**: `plaesy clean --dry-run --level thorough`

---

---

## `config`

```bash

plaesy config [subcommand] [flags]
```

Centralized platform configuration management using
`scripts/configs/platform.json`.

Global flag: `--config` (path to platform.json; default: auto-resolved).

### Subcommands

| Subcommand | Args | Description |
|---|---|---|
| `get-mapping` | `<section> <mapping_type>` | Get mapping value with value/excludes structure |
| `get-excludes` | `<section> <mapping_type>` | Get exclude patterns for a mapping |
| `get-clean-files` | `[platform]` | Get files to clean for platform |
| `get-clean-dirs` | `[platform]` | Get directories to clean for platform |
| `get-structure` | `<component>` | Get Plaesy structure config |
| `validate` | — | Validate platform configuration |

These six read `scripts/configs/platform.json` and ask things about **the
file**. The four platform questions — which platform am I on, which exist, what
does one look like, what is its config value — used to sit here too, which put
two different objects under one noun: the config file, and the AI platform. They
are now `plaesy platforms`; see that section. With no subcommand, `plaesy config`
prints this list.

**Example**: `plaesy config validate`

---

---

## `context`

```bash

plaesy context update [claude|gemini|copilot|cursor|qwen|opencode]
```

### `context update`

Sync **existing** agent context files with the current feature's `plan.md`. With no
argument, updates every detected context file; if none exist, exits non-zero listing
the paths it looked for. It does not create a context file — the agent-file template
was removed on 2026-09-25. Delegates to `internal/agentcontext`
(`scripts/internal/agentcontext/update.go`).

It was `plaesy update-agent-context`: a three-word bare verb with no resource in
it. The resource is `context` and the action is `update`, matching
`features create` and `images create`. With no subcommand, `plaesy context`
prints this list. The old spelling was removed, not aliased.

**Example**: `plaesy context update claude`

---

---

## `features`

```bash

plaesy features [subcommand] [flags]
```

The feature-branch workflow: create one, list them, print the current one's
paths, and check it is ready for task generation. With no subcommand, lists the
features that exist.

| Subcommand | Args | Description |
|---|---|---|
| *(none)* | — | List every feature in `.plaesy/specs/` |
| `create` | `<feature_description>` | Branch, directories, and `spec.md` |
| `paths` | — | Shell-sourceable paths for the current feature |
| `validate` | `[--json]` | Is the current feature ready for tasks |

These were three top-level commands — `create-new-feature`, `get-feature-paths`
and `check-task-prerequisites` — in three different shapes, none of which said
they belonged together. All three spellings were **removed**, not aliased: they
had no deprecation window, and a leftover alias would be invisible, since
deprecated commands are hidden from help and from the documentation bar the
drift tests enforce.

### `features create`

```bash

plaesy features create <feature_description> [--json]
```

Creates the next `NNN-slug` branch, checks it out, and seeds
`.plaesy/specs/<branch>/spec.md`. Numbering is `highest + 1` across existing
`NNN-*` directories. Details: [prompts/create.md](../prompts/create.md) —
there is no `docs/scripts/create.md`; the `/create` prompt is the only
specification of this command.

| Flag | Default | Description |
|---|---|---|
| `--json` | `false` | Output as JSON |

**Example**: `plaesy features create "Add dark mode toggle"`

### `features` (list)

Lists every feature directory, ordered by number then name, with the design
documents present in each and `*` on the checked-out branch. A directory that
is not `NNN-slug` is skipped. On a project with no `.plaesy/specs/`, prints the
command to create one rather than an empty block.

**Example**: `plaesy features`

### `features paths`

```bash

plaesy features paths
```

Print shell-sourceable paths (`REPO_ROOT=`, `FEATURE_DIR=`, etc.) for the
current feature branch. No arguments. Reads from `internal/featurepath`.

**Example**: `eval $(plaesy features paths)`

### `features validate`

Checks the current feature has a `plan.md` and reports which design documents
are available. With no feature directory it is an error, not an empty report.

| Flag | Default | Description |
|---|---|---|
| `--json` | `false` | Output as JSON |

**Example**: `plaesy features validate --json`

---

---

## `graph`

```bash

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

---

## `images`

```bash

plaesy images create --prompt <text> --out <path> [--provider <openai|gemini>] [--size <WxH>]
```

Generates an image asset via a configured provider API. With no subcommand,
prints the list below.

| Flag | Default | Description |
|---|---|---|
| `--prompt` | `""` | Image prompt text (required) |
| `--provider` | `""` | Image provider: `openai` (default) or `gemini` |
| `--size` | `""` | Image size, e.g. `1024x1024` |
| `--out` | `""` | Output file path (required) |

**Example**: `plaesy images create --prompt "a logo" --out logo.png --provider openai`

---

---

## `init`

```bash

plaesy init [directory] [--ai <platform>] [--plaesy-home <path>]
```

Scaffold a new Plaesy project: `.plaesy/` structure + AI platform files.

| Flag | Default | Description |
|---|---|---|
| `--ai` | `""` | AI platform (e.g. `claude`, `cursor_ai`) |
| `--plaesy-home` | `""` | Override Plaesy repo root (`PLAESY_HOME` env var) |

Supported platforms: `claude`, `cursor_ai`, `github_copilot`,
`windsurf_ai`, `cline`, `deepseek`, `kilo`, `qoder`, `trae_ai`,
`continue_dev`, `tabnine`, `codeium`, `codewhisperer`, `studio_bot`,
`replit_ghostwriter`, `llama_index`, `ollama`, `lm_studio`, `generic_ai`.

**Examples**:

```bash

plaesy init my-project --ai claude
plaesy init . --ai cursor_ai
```

---

---

## `install`

```bash

plaesy install
```

Install the `plaesy` CLI to a well-known bin directory. Copies the currently-
running binary to `%LOCALAPPDATA%\Plaesy\bin\plaesy.exe` (Windows) or
`$HOME/.local/bin/plaesy` (Unix). It also extracts the embedded Plaesy home
(`templates/`, `instructions/`, `prompts/`, `agents/`, `checklists/`,
`scripts/configs`) to `%LOCALAPPDATA%\Plaesy\.plaesy` (Windows) or
`$HOME/.local/share/plaesy/.plaesy` (Unix), so `plaesy init` works without a
spec-kit checkout of its own — `FindHome` falls back to this location when no
override, `PLAESY_HOME`, or ambient checkout is found. If the bin directory is
not on PATH, prints instructions for adding it (does not modify shell rc or
registry).

**No flags.** Reads its own executable path via `os.Executable()`.

---

---

## `platforms`

```bash

plaesy platforms [subcommand] [flags]
```

The AI platform a project targets: `claude`, `cursor_ai`,
`github_copilot`, and the rest of `scripts/configs/platform.json`.

| Subcommand | Args | Description |
|---|---|---|
| *(none)* | — | Print this list |
| `detect` | — | Which platform this project is on |
| `list` | — | Every platform plaesy knows about |
| `show` | `[platform]` | One platform in detail |
| `get` | `<platform> <key>` | A platform's configuration value |

These four were `plaesy config detect`, `config list`, `config get-platform`
and `config show`. `config` had ended up grouping commands by the **file** they
read (`scripts/configs/platform.json`) rather than by the **object** a user is
thinking about — nobody asks "what does platform.json say", they ask "which
platform am I on" and "which platforms exist". Grouping by resource is what
kubectl, `gh` and terraform do, and it completes the noun family this CLI now
has: `features`, `tasks`, `images`, `platforms`, `context`, `config`, `stack`.

The six that stayed in `config` — `get-mapping`, `get-excludes`,
`get-clean-files`, `get-clean-dirs`, `get-structure` and `validate` — genuinely
are about the file. `platforms get-structure` or `platforms validate` would
name a resource the command does not touch.

`--config <path>` overrides `platform.json` for all four.

**Example**: `plaesy platforms detect`

---

---

## `repair`

```bash

plaesy repair
```

Print "not yet implemented" message. See `docs/scripts/install.md` for upgrade
instructions.

---

---

## `stack`

```bash

plaesy stack [subcommand] [flags]
```

Inspect a target project's technology stack.

| Subcommand | Args | Description |
|---|---|---|
| *(none)* | — | Print this list |
| `detect` | `[target-dir]` | List instruction files relevant to that stack |

`stack detect` was the bare `plaesy detect-stack`. "detect-stack" fused the two
things the rest of this tree was un-bundling: it detects a technology *stack*,
from marker *files* in a target project. Both are nouns here — `platforms` is a
target, a `stack` is a target — and `detect` is a verb, so `stack detect` says
the same thing while matching `features validate`, `images create` and `context
update`. It also read as a typo'd `detect --stack`.

`stack` currently has one subcommand, which is usually a sign a parent is
premature. It earns the wrapper because the noun is the stable part: a future
`stack explain` or `stack check` then lands without another top-level rename.

### `stack detect`

List instruction files relevant to a target project's tech stack, using
`instructions/mapping.json` as the keyword registry.

| Flag | Default | Description |
|---|---|---|
| `--plaesy-root` | `""` | Path to Plaesy repo root (`$PLAESY_ROOT`, then git root) |
| `--install` | `false` | Copy each detected instruction file into `.plaesy/instructions/` (the `.instructions` suffix is stripped). Existing files are never overwritten. |

**Examples**:
`plaesy stack detect .` — prints one instruction filename per line.
`plaesy stack detect . --install` — prints the same list and copies each file in.

---

---

## `status`

```bash

plaesy status
```

Show Plaesy CLI version and environment status. Reports:

- git availability
- resolved git repository root (uses `GetRepoRoot()` with Windows path normalization)
- Banner with version (`internal/common.Version`, set at build time)

---

---

## `tasks`

```bash

plaesy tasks [subcommand]
```

Task lifecycle management for `.plaesy/tasks/{backlog,todo,doing,done,blocked}/`.

### Subcommands

| Subcommand | Args | Description |
|---|---|---|
| `list` | — | List all tasks across every status |
| `list-status` | `<status>` | List tasks in one status |
| `next` | — | Move the next backlog task into `todo` and report it |
| `start` | `<task>` | Move task to `doing` |
| `complete` | `<task>` | Move task to `done` |
| `block` | `<task>` | Move task to `blocked` |
| `unblock` | `<task>` | Unblock a task |
| `move` | `<task_file> <from> <to>` | Move between arbitrary statuses |
| `show` | `<task_file> [status]` | Display task content |

Valid statuses: `backlog`, `todo`, `doing`, `done`, `blocked`.

It was `plaesy task-manage`, which spelled the resource as `task` and then
repeated the idea in `-manage`: `tasks list` says everything `task-manage list`
does, one word shorter. The old spelling was removed, not aliased.

---

---

## `trim`

```bash

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

---

## `uninstall`

```bash

plaesy uninstall
```

Remove the installed `plaesy` binary. Prompts `y/N` (default N). Does not
remove any project's `.plaesy/` directory — only the binary itself.

---

---

## `upgrade`

```bash

plaesy upgrade
```

Print "not yet implemented" message. Build a new binary from source and re-run
`plaesy install` (see `docs/scripts/install.md`).

---

---

## `validate`

```bash

plaesy validate [target|file] [args...] [flags]
```

One verb for every check. Three ways to call it:

```bash

plaesy validate                      # every project-wide check, with a summary
plaesy validate constitution         # one target, with its own flags
plaesy validate deck.pptx            # a document, dispatched by extension
```

With no arguments it runs each project-wide check — constitution, memory,
Markdown, and assumptions — and prints what ran, what was skipped, and why. A check that does not
apply (no constitution yet, no Markdown config) is reported as `[SKIP]` with the
reason, never silently passed over: an omitted check that prints nothing reads as
a check that passed.

| Target | Scope | What it checks |
|---|---|---|
| `constitution` | project-wide | The generated constitution against the contract its template states |
| `memory` | project-wide | `.plaesy/memory/*.md` for external, non-self-contained references |
| `markdown` | project-wide | Markdown style, structure, and image/link accessibility |
| `docx` | file | `.docx` OOXML structure |
| `pptx` | file | `.pptx` OOXML structure, slide count, title placeholders |
| `xlsx` | file | `.xlsx` OOXML structure, sheet names |
| `assumptions` | project-wide | `ASSUMED —` tags left by autonomous runs, optionally gated by age |

`plaesy validate --list` prints the same table from the binary, so it cannot drift
from the code.

The three Office targets validate the container directly: the file must be a ZIP
holding `[Content_Types].xml` and its main part, every XML part must contain a
well-formed *root element* (bare text and empty parts are rejected), and the
document-specific rules follow — a docx needs at least one paragraph or table and
no leftover `{{ }}` template tags, a pptx's slide count comes from its
`sldIdLst` (and a missing title placeholder is a warning, not a failure), an
xlsx needs at least one sheet and its worksheets must parse. The headless-render
check the shell versions performed with LibreOffice is not available in this Go
port, and the command prints a notice saying so rather than doing less silently.

A `.md` argument routes to `markdown`; `.docx`/`.pptx`/`.xlsx` route to their own
validators. Extra expectations (a slide count, expected sheet names) are only
available through the explicit target form, because with several files there is
no unambiguous owner for a trailing argument.

Anything else is an error naming the valid targets, rather than a guess.

### Legacy spellings

`plaesy validate-constitution`, `validate-memory`, `validate-markdown`,
`validate-docx`, `validate-pptx`, and `validate-xlsx` still run and print where
they moved. They are deprecated and will be removed; the subcommand form is the
supported one.

### `validate constitution`

```bash

plaesy validate constitution [path]
```

Validate a generated project constitution (default
`.plaesy/memory/constitution.md`) against the contract
`templates/constitution.template.md` states about itself:

- frontmatter carries `title`, `version`, `ratified`, `last_amended`,
  `active_dimensions`; dates are ISO 8601
- no unfilled `{{ ... }}` placeholder survives
- every §1 Active Dimensions row is a known dimension marked exactly `yes`/`no`,
  and at least one is active
- frontmatter `active_dimensions` matches the §1 rows marked `yes` (the table is
  the source of the active scope)
- the frontmatter version matches the last amendment-log row
- rule IDs (`EV-`/`QG-`/`DOC-`/`SC-`/`RTM-`/`NN-`) are unique

Prints the declared version, active dimensions, and rule/non-negotiable counts.
Exits non-zero and lists every problem at once if the constitution is unfilled,
unratified, or internally inconsistent. Run it after `/start` (Phase 1, sub-step 1.0) and after
any amendment.

### `validate memory`

```bash

plaesy validate memory
```

Scan `.plaesy/memory/*.md` for external, non-self-contained references
(`~/.claude/`, `/tmp/`, `C:\Users\...\.claude`). Reports files, patterns, and
line numbers. Exits non-zero if issues found.

### `validate markdown`

```bash

plaesy validate markdown [paths...] [flags]
```

Lint Markdown with the linter built into this binary. No Node, no npm, no
`node_modules` — for this repository and for any project that installed the
framework. The rule IDs and config keys are markdownlint's, so
`.markdownlint.json` stays portable.

25 rules are implemented: MD001, MD003, MD004, MD007, MD009, MD010, MD012,
MD013, MD022, MD024, MD025, MD026, MD029, MD031, MD032, MD036, MD040, MD041,
MD046, MD047, MD048, MD058, plus the three GitHub accessibility rules GHA001
(alt text), GHA002 (alt text that is not just a file name) and GHA003 (link text
that describes its destination). `plaesy validate markdown --list-rules` prints
the set with descriptions; a clean run means "no implemented rule fired", not
"markdownlint would be happy".

| Flag | Effect |
|------|--------|
| `--config <path>` | Config file (default `<repo>/.markdownlint.json`) |
| `--baseline <path>` | Ratchet baseline (default `<repo>/.markdownlint-baseline.json`) |
| `--no-baseline` | Ignore the baseline; require zero violations |
| `--summary` | Print only the per-rule histogram |
| `--max-files <n>` | How many offending files to detail (0 = all, default 20) |
| `--fix` | Apply the machine-fixable rules in place |
| `--write-baseline` | Re-measure and record the current total as the new ceiling |
| `--list-rules` | List implemented rules and exit |

Two behaviours are deliberate and differ from the npm tooling:

- **Unknown rule names and unknown option keys are hard errors.** A config that
  half-loads is a check that reports success while checking nothing. `extends`
  is rejected outright — there is no registry to resolve it against, so the
  rules a project needs must live in the config file. `$`-prefixed keys are
  treated as human documentation and ignored.
- **The default run is a ratchet.** Pre-existing debt should not block unrelated
  work, but new violations must fail. The baseline records a ceiling and the
  per-rule counts behind it; a run passes while the total stays at or below the
  ceiling and fails when it rises. `TestRepoMarkdownStaysWithinRatchet` in
  `scripts/internal/mdlint` runs the same check inside `go test ./...`, which is
  the only CI job — that is what keeps Markdown linting alive without a second
  pipeline to forget. Lower the ceiling when you fix things; raising it to make
  a red build green defeats the point.

The ratchet is measured over the whole tree, so linting a single file compares
that file against the repository-wide ceiling. Pass `--no-baseline` when you
want zero tolerance for one file.

Machine-fixable rules: MD009, MD010, MD012, MD022, MD031, MD032, MD047, MD058.

`--fix` derives its decisions from the same predicates the corresponding check
reports on, and `internal/mdlint/fix_test.go` lints the output of every fix pass
and fails when a fixable rule still fires. That test is not ceremony: the fixer
once asked whether the wrong line continued the block above it, so a list below a
heading never got its blank line and the command reported `[OK] nothing to fix` on
a file the linter had just listed 20 fixable violations for. A fix pass that
changes nothing and a document that is already clean look identical from the
outside, so the pass is checked by what it leaves behind.

### `validate assumptions`

```bash

plaesy validate assumptions [--review-days <n>]
```

Scans the corpus for `ASSUMED —` tags: the label an autonomous run writes
instead of stopping to ask (`instructions/plaesy.instructions.md` rule 2). By
default it is an audit, not a gate — it always lists what it found and exits
`0`. It becomes a gate only when `.plaesy/state.json`'s
`assumptions_review_days` is a positive integer (the same opt-in convention as
`autonomous_loop.checkpoint_interval`), in which case a tag whose file has had
no commit in over that many days fails the run — surfacing it for a batched,
asynchronous human review rather than blocking the autonomous run that wrote
it in the first place.

| Flag | Effect |
|------|--------|
| `--review-days <n>` | Override `.plaesy/state.json`'s `assumptions_review_days` for this run (`0` = use the config, or informational-only if unset) |

### `validate docx` / `validate pptx` / `validate xlsx`

```bash

plaesy validate docx <file.docx>
plaesy validate pptx <file.pptx> [expected_slide_count]
plaesy validate xlsx <file.xlsx> [expected_sheet_name...]
```

Validate OOXML structure without rendering. Reports paragraph/table/slide/sheet
counts and structural issues.

---

## Installation Commands

| Command | Purpose | Source |
|---|---|---|
| `plaesy install` | Copy binary to well-known bin dir | `scripts/cmd/plaesy/install.go` + `scripts/internal/installer/installer.go` |
| `plaesy status` | Show install location + PATH + version | `scripts/cmd/plaesy/status.go` |
| `plaesy uninstall` | Remove installed binary (y/N prompt) | `scripts/cmd/plaesy/install.go` |
| `plaesy repair` | Not yet implemented | `scripts/cmd/plaesy/install.go` |
| `plaesy upgrade` | Not yet implemented | `scripts/cmd/plaesy/install.go` |
