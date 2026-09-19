# Plaesy Common Functions Library

Shared utility functions and constants sourced/dot-included by other Plaesy Spec-Kit scripts. Two implementations exist, exporting a mostly-overlapping surface:

- `scripts/bash/common.sh` (bash 4+)
- `scripts/powershell/common.ps1` (PowerShell 5.1+, exports via `Export-ModuleMember`)

## Purpose

Provides consistent logging, error handling, validation, and (bash-only) git/feature-path helpers so individual scripts don't reimplement them.

## Exported functions and variables

| Bash | PowerShell | Purpose |
|------|-----------|---------|
| `log_debug` | *(none)* | Debug message, only when debug mode enabled |
| `log_info` | `Log-Info` | Informational message |
| `log_success` | `Log-Success` / `Write-Success` | Success message |
| `log_warning` | `Log-Warning` | Warning message (stderr in bash) |
| `log_error` | `Log-Error` | Error message (stderr in bash) |
| `log_message` | `Write-LogMessage` | Lower-level formatter the above wrap (`level`, `message` params) |
| *(none)* | `Write-Step` | Blue `[STEP]` message |
| *(none)* | `Write-Warning-Custom` | Warning message |
| *(none)* | `Write-Error-Custom` | Error message, then `exit 1` |
| `setup_error_handling` | *(none)* | Installs `ERR`/`EXIT` traps and enables `set -euo pipefail` |
| `handle_error` | *(none)* | Trap target; logs failing line/command, prints stack trace in debug mode |
| `cleanup_on_error` | *(none)* | Removes stale `/tmp/plaesy-*` files older than 60 min |
| *(none)* | `Invoke-ErrorHandling` | Logs error, optional stack trace via `Get-PSCallStack`, `exit $ExitCode` |
| `validate_command_exists` | `Test-CommandExists` | Check a command/executable is on PATH |
| `validate_file_exists` | `Test-FileExists` | Check a file exists |
| `validate_directory_exists` | `Test-DirectoryExists` | Check a directory exists |
| `validate_environment` | *(none)* | Checks bash>=4, `git`/`curl` present, warns if not in a git repo |
| `validate_not_root` | `Test-NotRoot` | Bash: errors if `EUID -eq 0`. PowerShell: returns `$true` if **not** Administrator |
| `validate_disk_space` | *(none)* | Errors if free space under `$HOME` is below the given MB |
| `time_execution` | *(none)* | Runs a command, logs elapsed time via `log_debug` |
| `safe_copy` | *(none)* | Copies a file, backing up any existing destination first |
| `print_banner` | `Write-Banner` | Prints the Plaesy ASCII banner with title/subtitle |
| `get_plaesy_version` | `Get-PlaesyVersion` | Reads `VERSION` file (repo root, then `~/.plaesy/VERSION`), falls back to `0.0.0` |
| `get_repo_root` | *(none)* | `git rev-parse --show-toplevel` |
| `get_current_branch` | *(none)* | `git rev-parse --abbrev-ref HEAD` |
| `check_feature_branch` | *(none)* | Validates branch name matches `NNN-*` |
| `get_feature_dir` | *(none)* | Builds `$repo_root/specs/$branch` |
| `get_feature_paths` | *(none)* | Emits `REPO_ROOT`, `CURRENT_BRANCH`, `FEATURE_DIR`, `FEATURE_SPEC`, `IMPL_PLAN`, `TASKS`, `RESEARCH`, `DATA_MODEL`, `QUICKSTART`, `CONTRACTS_DIR` as `KEY='value'` lines for `eval` |
| `check_file` / `check_dir` | *(none)* | Prints ✓/✗ existence status with a label |
| *(none)* | `Test-CommandExists`... | (see validation row above) |
| *(none)* | `Get-Timestamp` | Current time as `yyyy-MM-dd HH:mm:ss` |
| *(none)* | `Get-ScriptDirectory` | Directory of the invoking script |
| *(none)* | `Write-SectionHeader` / `Write-SectionFooter` | Magenta `=== Title ===` banner |
| *(none)* | `Write-Progress` | Thin wrapper around the built-in `Write-Progress` cmdlet |

Git/feature-path helpers (`get_repo_root` through `check_dir`) exist only in bash today; PowerShell scripts needing them call `scripts/powershell/get-feature-paths.ps1` instead.

### Variables / constants

| Bash | PowerShell | Purpose |
|------|-----------|---------|
| `PLAESY_VERSION` (readonly) | `$env:PLAESY_VERSION` | Framework version, auto-detected |
| `SCRIPT_START_TIME` (readonly) | `$env:SCRIPT_START_TIME` | Timestamp set on load |
| `RED`/`GREEN`/`YELLOW`/`BLUE`/`PURPLE`/`CYAN`/`WHITE`/`GRAY`/`NC` | `$Colors` hashtable (`Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White`, `Gray`) | ANSI/name colors for output |
| `LOG_LEVEL`, `LOG_FILE`, `DEBUG_MODE` (readonly, from `PLAESY_LOG_LEVEL`/`PLAESY_LOG_FILE`/`PLAESY_DEBUG`) | `$script:LogLevel`, `$script:LogFile`, `$script:DebugMode` (same env vars) | Logging configuration |

Environment variables recognized by both: `PLAESY_DEBUG`, `PLAESY_LOG_LEVEL`, `PLAESY_LOG_FILE`.

## Usage

### Bash
```bash
#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

setup_error_handling
validate_environment
validate_command_exists "git" || exit 1

print_banner "My Script" "Version 1.0.0"
log_info "Starting..."
eval "$(get_feature_paths)"
check_file "$IMPL_PLAN" "Implementation plan"
log_success "Done"
```

### PowerShell
```powershell
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Import-Module "$ScriptDir\common.ps1" -Force

Write-Banner -Title "My Script" -Subtitle "Version 1.0.0"
Log-Info "Starting..."
if (-not (Test-CommandExists "git")) { Log-Error "git required"; exit 1 }
if (-not (Test-FileExists $ConfigFile)) { Log-Error "Missing $ConfigFile"; exit 1 }
Log-Success "Done"
```

## Notes

- Bash's `setup_error_handling` enables `set -euo pipefail` plus `ERR`/`EXIT` traps — call it before other logic, not after. PowerShell has no equivalent single entry point; use `Invoke-ErrorHandling` per failure site or standard `try/catch`.
- `get_feature_paths` output is meant to be `eval`'d in bash; there is no direct PowerShell port in `common.ps1` — use `scripts/powershell/get-feature-paths.ps1`.
- Debug output (`log_debug`) only appears when `PLAESY_DEBUG=true`.
