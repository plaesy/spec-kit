# internal/common

Shared logging, error handling, validation, and git/feature-path helper
package used by every `plaesy` subcommand. This replaces the old
`scripts/bash/common.sh` (sourced by every bash script) and
`scripts/powershell/common.ps1` (imported as a module by every PowerShell
script) with a single Go package imported directly:

```go
import "github.com/plaesy/spec-kit/internal/common"
```

Source: `scripts/internal/common/` (`logging.go`, `validate.go`,
`gitpaths.go`, `version.go`, `banner.go`).

## Purpose

Provides consistent logging, error handling, validation, and git/feature-path
helpers so individual commands don't reimplement them.

## Exported API

| Function | Purpose |
|----------|---------|
| `LogDebug(format, args...)` | Debug message, only when `PLAESY_DEBUG=true` |
| `LogInfo(format, args...)` | Informational message |
| `LogSuccess(format, args...)` | Success message |
| `LogWarning(format, args...)` | Warning message (stderr) |
| `LogError(format, args...)` | Error message (stderr) |
| `ValidateCommandExists(cmd, description)` | Check a command/executable is on PATH |
| `ValidateFileExists(path, description)` | Check a file exists |
| `ValidateDirectoryExists(path, description)` | Check a directory exists |
| `ValidateEnvironment()` | Checks `git` is present, warns if not in a git repo |
| `ValidateNotRoot()` | Errors if running as root/EUID 0 (a no-op on Windows, which has no EUID concept) |
| `PrintBanner(title, subtitle)` | Prints the Plaesy ASCII banner with title/subtitle |
| `Version` | Package variable holding the framework version (set at build time via `-ldflags`, falls back to `0.0.0`) |
| `NormalizeVersion(raw)` | Sanitizes a version string to strict semver, falling back to `0.0.0` |
| `GetRepoRoot()` | `git rev-parse --show-toplevel` |
| `GetCurrentBranch()` | `git rev-parse --abbrev-ref HEAD` |
| `CheckFeatureBranch(branch)` | Validates branch name matches `NNN-*` |
| `GetFeatureDir(repoRoot, branch)` | Builds `<repoRoot>/specs/<branch>` |
| `GetFeaturePaths()` | Returns a `*FeaturePaths` struct with `RepoRoot`, `CurrentBranch`, `FeatureDir`, `FeatureSpec`, `ImplPlan`, `Tasks`, `Research`, `DataModel`, `Quickstart`, `ContractsDir` |
| `(*FeaturePaths).ShellLines()` | Renders the same fields as shell-sourceable `KEY='value'` lines, for anything that still wants to `eval` this output (see [get-feature-paths.md](./get-feature-paths.md)) |
| `CheckFile(path, label)` / `CheckDir(path, label)` | Returns a `  ✓ label` / `  ✗ label` string |

### Environment variables

| Variable | Purpose |
|----------|---------|
| `PLAESY_DEBUG` | `true` enables `LogDebug` output |
| `PLAESY_LOG_FILE` | If set, log lines are additionally appended to this file |

## Usage

```go
package main

import (
    "github.com/plaesy/spec-kit/internal/common"
)

func run() error {
    common.PrintBanner("My Command", "Version 1.0.0")
    common.LogInfo("Starting...")

    if err := common.ValidateCommandExists("git", "command"); err != nil {
        return err
    }

    fp, err := common.GetFeaturePaths()
    if err != nil {
        return err
    }
    fmt.Println(common.CheckFile(fp.ImplPlan, "Implementation plan"))

    common.LogSuccess("Done")
    return nil
}
```

## Notes

- There is one implementation now, not a bash/PowerShell pair — every
  `plaesy` subcommand imports this same package, so behavior is identical
  across platforms by construction (no parity drift to track).
- `GetFeaturePaths().ShellLines()` exists specifically so `plaesy get-feature-paths`
  can still emit `eval`-able output for any external tooling that sources it
  (see [get-feature-paths.md](./get-feature-paths.md)); new Go code should
  use the `*FeaturePaths` struct fields directly instead.
- Debug output (`LogDebug`) only appears when `PLAESY_DEBUG=true`.
