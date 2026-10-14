# plaesy images create

Generates an image asset through a configured provider API (`openai` or
`gemini`).

Source: `scripts/cmd/plaesy/images.go` (verb) + `image_create.go`
(subcommand), `scripts/internal/imagegen`.

| Subcommand | Documented in |
|---|---|
| `plaesy images create` | below |

> **Renamed.** This was `plaesy create image` (and before that
> `plaesy generate-image`). The settled shape is the same one `features create`
> and `tasks start` use: **resource plural, then verb**. `create` was left with a
> single child and was the only top-level verb in a tree that is otherwise
> resource-nouns (`config`, `features`, `tasks`, `context`, `validate`), so it
> went too. The old spellings were **removed**, not aliased.

## Purpose

Generate an image asset via a configured provider API. With no subcommand,
`plaesy images` prints the list of available subcommands.

## Quick Start

```bash

plaesy images create --prompt "a logo" --out logo.png --provider openai
plaesy images create --prompt "a logo" --out logo.png --size 1024x1024
```

## Options

| Flag | Default | Description |
|---|---|---|
| `--prompt` | `""` | Image prompt text (required) |
| `--out` | `""` | Output file path (required) |
| `--provider` | `""` | `openai` (default) or `gemini` |
| `--size` | `""` | e.g. `1024x1024` |
| `--help` / `-h` | | Print usage and exit |

Both flags marked required really are: omitting either prints usage and exits
non-zero, rather than exiting 0 having generated nothing. Parent directories of
`--out` are created if they do not exist.

See [docs/reference.md](../reference.md#images) for the full flag table, and
`prompts/create/images.md` for the image workflow that drives it.

## Troubleshooting

**"usage: plaesy images create --prompt <text> --out <file>" error** — a
required flag (`--prompt` or `--out`) was omitted.

**Provider API failure** — the configured provider returned an error; check
your API key and the `--provider` value (`openai` or `gemini`).

**Output path not writable** — parent directories are created automatically, so
this usually means a permissions problem on the target directory.

## Related Commands

- **`plaesy features create`** — creates the feature branch/spec that this command's output may document (see [features.md](./features.md))
- **`plaesy context update`** — updates AI context for the current feature (see [context-update.md](./context-update.md))
