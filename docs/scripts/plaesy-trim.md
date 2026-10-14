# plaesy trim

**Token and context compression for command output and memory/instruction files.** Go standard library only, no network calls.

Full usage guide: [instructions/context-engineering.instructions.md](../../instructions/context-engineering.instructions.md).

## Purpose

Cuts token spend at two points:

1. **Layer 1 — command output**: `run <command...>` executes a shell command and compresses its output before it reaches context (per-tool aware via `scripts/configs/plaesy-trim-rules.json` — git,
   npm, cargo, pytest, dotnet, docker, ...). Test-runner output collapses to failures + summary only.
2. **Layer 2 — memory/instruction files**: `compress` (fast, local, heuristic) or `llm-queue`/`apply-llm` (LLM-quality rewrite, done by the calling assistant, no API key).

Code blocks, shell commands, error messages/stack traces, and structured data (JSON/YAML) are always preserved byte-exact — only narrative prose is compressed.

`trim run` is a transparent wrapper: it prints the command's compressed output
and then exits with a non-zero status if the command itself failed, so
`plaesy trim run go build ./... && deploy` still stops on a failed build. The
output is printed before the failure is reported, because on a failed build the
output *is* the error message.

## Quick Start

```bash

plaesy trim run git status
plaesy trim compress --path instructions/plaesy.instructions.md
plaesy trim compress --path instructions --recurse
plaesy trim compress --path instructions --dry-run
plaesy trim compress --path README.md --level ultra
plaesy trim llm-queue --path instructions/plaesy.instructions.md
plaesy trim apply-llm --path instructions/plaesy.instructions.md --annotations ann.json
plaesy trim report
```

Run `plaesy trim --help` or `plaesy trim <subcommand> --help` for the exact current flag set (source of truth: `scripts/cmd/plaesy/trim.go`).

## Compression levels (Layer 2)

- `lite` — strip filler phrases, collapse repeated blank lines
- `full` (default for mid-traffic files) — `lite` + merge short redundant sentences, trim hedging
- `ultra` — `full` + telegram-style fragments (readability drops — low-traffic files only)

Level is auto-selected per file from its `plaesy analyze` in-degree (`degree` field in `.plaesy/analysis/project.graph.json`) when `--level` is omitted; falls back to `full` if that file doesn't exist
yet (run `plaesy analyze` first).

Every `compress` run backs up the original as `<file>.bak` and logs to `.plaesy/memory/token-stats.json`; `report` reads that log.

## LLM-mode round-trip (Layer 2)

1. `llm-queue -Path <file>` splits the file on fenced code blocks and writes prose segments ≥40 chars to `.plaesy/analysis/trim-queue.json`.
2. Hand that file to the calling assistant to rewrite each segment denser, same meaning, without touching code/commands/paths.
3. Save the reply as `annotations.json`, then `apply-llm -Path <file> -Annotations annotations.json` — segment indices must match the same, unmodified source file used for `llm-queue`.

## Type-aware filters

`trim run` picks a filter from the command name before the generic compression
runs. Each filter restructures the output so the lines an agent needs stay
exact:

| Command | Filter | What it keeps / drops |
|---|---|---|
| `git diff` | `git-diff` | Keeps file and hunk headers and every +/- line; drops `index` lines and context lines |
| `git status` | `git-status` | Keeps all state lines; drops `(use ...)` hints and blank lines |
| `git log` | `git-log` | One line per commit (7-character sha), then the first message line; drops Author/Date |
| `grep`, `rg` | `grep` | Keeps match lines exactly; drops blank lines and repeated `--` separators |
| `ls`, `find`, `tree` | `listing` | Drops the `total` line and blank lines |

A filter that fails or panics passes the raw output through and writes one
warning to stderr. An unknown command runs the generic path unchanged. Each
filter's savings are recorded as `layer1:<filter>`, so `plaesy trim report`
shows them separately. Measured savings are in
`.plaesy/decisions/delegation-provider-and-token-savers.md`.
