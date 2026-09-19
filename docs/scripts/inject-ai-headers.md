# Inject AI Headers Script

Injects platform-specific AI instruction headers (YAML front-matter + guidance) into prompt, chatmode, and instruction files.

For the header file format, naming convention, `{{DESCRIPTION}}` placeholder, and how headers are authored, see [templates/ai-headers/README.md](../../templates/ai-headers/README.md). This page covers script usage only.

Scripts: `scripts/bash/inject-ai-headers.sh` and `scripts/powershell/inject-ai-headers.ps1`.

## Quick Start

```bash
# bash
./scripts/bash/inject-ai-headers.sh --ai claude --target ./prompts

./scripts/bash/inject-ai-headers.sh --ai claude --target . --merge --backup

./scripts/bash/inject-ai-headers.sh --ai copilot --target . --dry-run
```

```powershell
# PowerShell
.\scripts\powershell\inject-ai-headers.ps1 -AiPlatform claude -TargetDirectory .\prompts

.\scripts\powershell\inject-ai-headers.ps1 -AiPlatform claude -TargetDirectory . -Merge -Backup

.\scripts\powershell\inject-ai-headers.ps1 -AiPlatform copilot -TargetDirectory . -DryRun
```

## Options

| Purpose | bash | PowerShell |
|---|---|---|
| AI platform (required) | `--ai <platform>` | `-AiPlatform <platform>` |
| Target directory (required) | `--target <dir>` | `-TargetDirectory <dir>` |
| Preview without writing | `--dry-run` | `-DryRun` |
| Overwrite existing headers | `--force` | `-Force` |
| Backup originals before writing | `--backup` | `-Backup` |
| Merge missing keys into existing front-matter | `--merge` | `-Merge` (accepted but **not wired up** — see below) |
| List file → header mapping only, no writes | `--list-only` | `-ListOnly` |
| Extra include pattern (repeatable) | `--pattern <glob>` | `-Pattern <regex[]>` |
| Exclude path fragment (repeatable) | `--exclude <fragment>` | not supported |
| Show help | `--help` | `-Help` |

**Supported platforms (bash)**: `copilot`, `cursor`, `windsurf`, `claude`, `chatgpt`, `gemini`, `trae-ai`, `qwen-code`, `codex-cli`, `opencode-cli`, `local-ai`, `manual`.

**Supported platforms (PowerShell)**: the same set, plus framework-name aliases that fall back to the base platform's `.header.yaml`: `github_copilot`→copilot, `cursor_ai`→cursor, `windsurf_ai`→windsurf, `claude_code`→claude, `trae_ai`→trae-ai, `qwen_code`→qwen-code.

## Verified behavior differences

Both scripts were read in full to confirm behavior rather than trusting prior docs:

- **Per-file-type header lookup**: bash resolves `templates/ai-headers/<platform>.<type>.yaml` (type = `prompts`/`chatmodes`/`instructions`/`generic`, detected from path/filename) before falling back to `<platform>.header.yaml` then `manual.header.yaml`. PowerShell's `Get-HeaderFile` looks for `<platform>.<type>.header.md` (note: `.header.md`, not `.yaml` — this candidate does not match the repo's actual `<platform>.<type>.yaml` naming, so in practice PowerShell always falls through to `<platform>.header.yaml`).
- **`{{DESCRIPTION}}` replacement**: bash extracts each target file's own `description` (or derives one from the filename) and substitutes it into the header before injection. PowerShell injects the header file's raw content verbatim — no placeholder substitution.
- **`--merge` / `-Merge`**: bash actually calls its merge logic per file when `--merge` is set and a file already has front-matter. PowerShell defines `Merge-FrontMatter` but `Invoke-Main` never calls it, so `-Merge` currently has no effect.
- **`.plaesy-headers.json` summary file**: bash writes one to the target directory after a non-dry-run with `processed > 0`. The equivalent block in the PowerShell script is present but commented out, so it is not created.
- **File discovery**: bash's default patterns are `*.prompt.md`, `*.chatmode.md`, `*.instructions.md` (glob, via `find`). PowerShell's defaults are regexes matching the filename only: `\.prompt\.md$`, `\.chatmode\.md$`, `\.instruction\.md$` (singular "instruction").
- **`--exclude`**: only bash supports excluding path fragments; PowerShell has no equivalent flag.

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

(PowerShell prints the same log lines but omits the final "Created configuration" line — that step is disabled.)

## Troubleshooting

- **"Invalid AI platform" / parameter validation error"**: check spelling against the supported list above; PowerShell will reject unknown values before running (`ValidateSet`), bash exits with an error after printing help.
- **Nothing gets injected**: run with `--dry-run`/`-DryRun` first, and confirm files actually match the default patterns (`*.prompt.md`, `*.chatmode.md`, `*.instructions.md` for bash; the singular-`instruction` regex set for PowerShell) or pass `--pattern`/`-Pattern` explicitly.
- **File skipped with "already has header"**: the script found YAML front-matter or a known header marker in the first lines. Use `--force`/`-Force` to overwrite, or `--merge` (bash only — see above) to add missing keys without clobbering existing ones.
- **Expected `--merge`/`-Merge` to update a PowerShell target and nothing changed**: this is the known gap above — the flag is currently a no-op in the `.ps1` script.
- **Header file not found**: verify `templates/ai-headers/<platform>.header.yaml` (or the more specific `<platform>.<type>.yaml`) exists; both scripts fall back to `manual.header.yaml` when nothing else matches.
