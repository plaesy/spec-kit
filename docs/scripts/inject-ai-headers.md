# plaesy inject-ai-headers

Injects platform-specific AI instruction headers (YAML front-matter +
guidance) into prompt, chatmode, and instruction files.

For the header file format, naming convention, `{{DESCRIPTION}}` placeholder,
and how headers are authored, see
[templates/ai-headers/README.md](../../templates/ai-headers/README.md). This
page covers command usage only.

Source: `scripts/cmd/plaesy/inject_ai_headers.go` + `scripts/internal/aiheaders/aiheaders.go`.

## Quick Start

```bash
plaesy inject-ai-headers --ai claude --target ./prompts

plaesy inject-ai-headers --ai claude --target . --merge --backup

plaesy inject-ai-headers --ai copilot --target . --dry-run
```

## Options

| Flag | Purpose |
|---|---|
| `--ai <platform>` | AI platform (required) |
| `--target <dir>` | Target directory (required) |
| `--dry-run` | Preview without writing |
| `--force` | Overwrite existing headers |
| `--backup` | Backup originals before writing |
| `--merge` | Merge missing keys into existing front-matter, instead of skipping files that already have a header |
| `--list-only` | List file → header mapping only, no writes |
| `--pattern <glob>` | Extra include glob pattern (repeatable; default: `*.prompt.md`, `*.chatmode.md`, `*.instructions.md`) |
| `--exclude <fragment>` | Exclude a path fragment (repeatable) |
| `--headers-dir <dir>` | Directory containing header YAML files (default: `<repo-root>/templates/ai-headers`) |
| `--help` | Show help |

**Supported platforms**: `copilot`, `cursor`, `windsurf`, `claude`, `chatgpt`,
`gemini`, `trae-ai`, `qwen-code`, `codex-cli`, `opencode-cli`, `local-ai`,
`manual`.

## Behavior

- **Per-file-type header lookup**: resolves `templates/ai-headers/<platform>.<type>.yaml`
  (type = `prompts`/`chatmodes`/`instructions`/`generic`, detected from
  path/filename) before falling back to `<platform>.header.yaml` then
  `manual.header.yaml`.
- **`{{DESCRIPTION}}` replacement**: extracts each target file's own
  `description` (or derives one from the filename) and substitutes it into
  the header before injection.
- **`--merge`**: when set and a file already has front-matter, merges the
  header's keys into it instead of skipping the file, preserving existing
  values and adding only missing keys.
- **`.plaesy-headers.json` summary file**: written to the target directory
  after a non-dry-run with `processed > 0`.
- **File discovery**: default patterns are `*.prompt.md`, `*.chatmode.md`,
  `*.instructions.md` (glob).
- **`--exclude`**: excludes files whose relative path contains the given
  fragment.

> **Migration note**: the old PowerShell script (`inject-ai-headers.ps1`) had
> several gaps relative to bash — `-Merge` was accepted but never wired up,
> `{{DESCRIPTION}}` substitution didn't happen, the `.plaesy-headers.json`
> summary write was commented out, and its per-file-type header lookup used a
> `.header.md` naming convention that didn't match the repo's actual
> `.yaml` files (so it always fell through to the generic header). None of
> those gaps exist in the Go port — it is a single, complete implementation
> matching the bash script's behavior (which was always the more complete of
> the two).

## Output example

```
[INFO] Starting header injection for AI platform: claude
[INFO] Using header file: claude.header.yaml
[INFO] Found 3 prompt file(s)
[SUCCESS] Injected header into idea.prompt.md
[INFO] Skipping security.chatmode.md (already has header, use --force to overwrite)
[SUCCESS] Injected header into dev.instructions.md

[SUCCESS] HEADER INJECTION COMPLETE
[SUCCESS] Processed: 2 files
[INFO] Skipped: 1 files
[INFO] Created configuration: .plaesy-headers.json
```

## Troubleshooting

- **"invalid AI platform" error**: check spelling against the supported list above.
- **Nothing gets injected**: run with `--dry-run` first, and confirm files actually match the default patterns (`*.prompt.md`, `*.chatmode.md`, `*.instructions.md`) or pass `--pattern` explicitly.
- **File skipped with "already has header"**: the command found YAML front-matter or a known header marker in the first lines. Use `--force` to overwrite, or `--merge` to add missing keys without clobbering existing ones.
- **Header file not found**: verify `templates/ai-headers/<platform>.header.yaml` (or the more specific `<platform>.<type>.yaml`) exists; the command falls back to `manual.header.yaml` when nothing else matches.
