# `trim`

```bash

plaesy trim [subcommand] [flags]
```

Token/context compression for command output, memory, and instruction files.

## Subcommands

| Subcommand | Short |
|---|---|
| `run -- <command...>` | Layer 1: run a command and compress its output |
| `compress --path <file\|dir>` | Layer 2: fast heuristic prose compression |
| `llm-queue --path <file>` | Export prose segments for AI rewriting |
| `apply-llm --path <file> --annotations <json>` | Merge rewritten segments back |
| `report` | Cumulative savings across layers |

## `trim compress` flags

| Flag | Default | Description |
|---|---|---|
| `--path` | `""` | File or directory to compress |
| `--level` | `""` | `lite`, `full`, `ultra` (auto-resolved from graph when omitted) |
| `--recurse` | `false` | Recurse into subdirectories |
| `--dry-run` | `false` | Report savings without writing |

## `trim run`

Uses `DisableFlagParsing` — passes all args after `--` to the target command,
then compresses output.

**Example**: `plaesy trim run -- echo "hello world"`

See [CLI Reference](../reference.md).
