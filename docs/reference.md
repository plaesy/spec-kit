# CLI Reference

Index of every `plaesy` command. Each command has its own page under [commands/](commands/):
flags and arguments are sourced verbatim from `scripts/cmd/plaesy/*.go`.

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

| Command | Short Description | Source File | More |
|---|---|---|---|
| `analyze` | Generate AI-optimized project analysis | `scripts/cmd/plaesy/analyze.go` | [commands/analyze.md](commands/analyze.md) |
| `clean` | Remove Plaesy framework files | `scripts/cmd/plaesy/clean.go` | [commands/clean.md](commands/clean.md) |
| `config` | Read and validate configs/platform.json | `scripts/cmd/plaesy/config.go` | [commands/config.md](commands/config.md) |
| `context` | Agent context files kept in sync with plan.md | `scripts/cmd/plaesy/context.go` | [commands/context.md](commands/context.md) |
| `delegate` | Send a bounded sub-task to a configured model provider (CLI, Anthropic API, Ollama) | `scripts/cmd/plaesy/delegate.go` | [commands/delegate.md](commands/delegate.md) |
| `doctor` | Check health of all system components (reach platforms, search index, config) | `scripts/cmd/plaesy/doctor.go` | [commands/doctor.md](commands/doctor.md) |
| `eval` | Run search evaluation harness (Recall@k, MRR, nDCG) against a golden set | `scripts/cmd/plaesy/eval.go` | [commands/eval.md](commands/eval.md) |
| `features` | List, create, and inspect feature branches | `scripts/cmd/plaesy/features.go` | [commands/features.md](commands/features.md) |
| `images` | Generate image assets | `scripts/cmd/plaesy/images.go` | [commands/images.md](commands/images.md) |
| `init` | Scaffold new Plaesy project | `scripts/cmd/plaesy/init.go` | [commands/init.md](commands/init.md) |
| `install` | Install binary to bin dir | `scripts/cmd/plaesy/install.go` | [commands/install.md](commands/install.md) |
| `platforms` | AI platforms plaesy can target | `scripts/cmd/plaesy/platforms.go` | [commands/platforms.md](commands/platforms.md) |
| `reach` | Access external data sources (web, github, youtube, etc.) across multiple platforms; `configure` and `list` subcommands | `scripts/cmd/plaesy/reach.go` | [commands/reach.md](commands/reach.md) |
| `repair` | Repair installation (stub) | `scripts/cmd/plaesy/install.go` | [commands/repair.md](commands/repair.md) |
| `reload` | Refresh generated `.plaesy/` files, keeping memory and context; `--prune` lists orphans in the tool-owned trees, `--prune-apply` removes them | `scripts/cmd/plaesy/reload.go` | [commands/reload.md](commands/reload.md) |
| `search` | Search the project by meaning. Indexes are built by `plaesy analyze`; a query refuses a stale index | `scripts/cmd/plaesy/search.go` | [commands/search.md](commands/search.md) |
| `stack` | Inspect a target project's technology stack | `scripts/cmd/plaesy/stack.go` | [commands/stack.md](commands/stack.md) |
| `stats` | Show search index and reach cache statistics | `scripts/cmd/plaesy/stats.go` | [commands/stats.md](commands/stats.md) |
| `status` | Show CLI version + environment | `scripts/cmd/plaesy/status.go` | [commands/status.md](commands/status.md) |
| `tasks` | Task lifecycle management | `scripts/cmd/plaesy/tasks.go` | [commands/tasks.md](commands/tasks.md) |
| `trim` | Token/context compression | `scripts/cmd/plaesy/trim.go` | [commands/trim.md](commands/trim.md) |
| `uninstall` | Remove installed binary | `scripts/cmd/plaesy/install.go` | [commands/uninstall.md](commands/uninstall.md) |
| `upgrade` | Upgrade installation (stub) | `scripts/cmd/plaesy/install.go` | [commands/upgrade.md](commands/upgrade.md) |
| `validate` | Run checks: constitution, memory, Markdown, OOXML | `scripts/cmd/plaesy/validate.go` | [commands/validate.md](commands/validate.md) |
| `constitution` | Validate a generated constitution | `scripts/cmd/plaesy/validate.go` | [commands/validate.md](commands/validate.md) |
| `memory` | Scan .plaesy/memory/ for external references and validate self-containment | `scripts/cmd/plaesy/validate.go` | [commands/validate.md](commands/validate.md) |
| `markdown` | Lint Markdown with the built-in Go linter (no Node, no npm) | `scripts/cmd/plaesy/validate_markdown.go` | [commands/validate.md](commands/validate.md) |
| `assumptions` | Audit ASSUMED tags left by autonomous runs, optionally gated by age | `scripts/cmd/plaesy/validate.go` | [commands/validate.md](commands/validate.md) |
| `docx` | Validate a .docx file's OOXML structure | `scripts/cmd/plaesy/validate.go` | [commands/validate.md](commands/validate.md) |
| `pptx` | Validate a .pptx file's OOXML structure | `scripts/cmd/plaesy/validate.go` | [commands/validate.md](commands/validate.md) |
| `xlsx` | Validate a .xlsx file's OOXML structure | `scripts/cmd/plaesy/validate.go` | [commands/validate.md](commands/validate.md) |
| `design` | Lint produced frontend output against the design rules (offline, no browser) | `scripts/cmd/plaesy/validate_design.go` | [commands/validate.md](commands/validate.md) |
| `evals` | Run the instruction case files: phrase checks, lint fixtures, and optional model briefs | `scripts/cmd/plaesy/validate_evals.go` | [commands/validate.md](commands/validate.md) |

---

## Installation Commands

| Command | Purpose | Source |
|---|---|---|
| `plaesy install` | Copy binary to well-known bin dir | `scripts/cmd/plaesy/install.go` + `scripts/internal/installer/installer.go` |
| `plaesy status` | Show install location + PATH + version | `scripts/cmd/plaesy/status.go` |
| `plaesy uninstall` | Remove installed binary (y/N prompt) | `scripts/cmd/plaesy/install.go` |
| `plaesy repair` | Not yet implemented | `scripts/cmd/plaesy/install.go` |
| `plaesy upgrade` | Not yet implemented | `scripts/cmd/plaesy/install.go` |
