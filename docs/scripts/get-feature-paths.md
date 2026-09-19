# get-feature-paths

**Feature branch and path resolution utility.**

## Purpose

Resolves the current feature branch and prints the paths to its spec/plan/tasks files, without creating anything. Used by other scripts and AI workflows to locate the active feature's documents.

## Quick Start

```bash
# Bash
./scripts/bash/get-feature-paths.sh
```

```powershell
# PowerShell
./scripts/powershell/get-feature-paths.ps1
```

Neither script takes any flags or arguments.

## Output Format

Both scripts print the same six `KEY: value` lines:

```
REPO_ROOT: /home/user/project
BRANCH: 001-user-authentication
FEATURE_DIR: /home/user/project/specs/001-user-authentication
FEATURE_SPEC: /home/user/project/specs/001-user-authentication/spec.md
IMPL_PLAN: /home/user/project/specs/001-user-authentication/plan.md
TASKS: /home/user/project/specs/001-user-authentication/tasks.md
```

Paths are not checked for existence — they are computed from the repo root and current branch name only (`specs/<branch>/...`).

## Behavior When Not on a Feature Branch

This is the one place bash and PowerShell diverge:

- **Bash**: still prints all six lines (using the actual current branch, e.g. `BRANCH: main`), then appends `INFO: Not on a feature branch (format: XXX-feature-name)` and exits `0`.
- **PowerShell**: prints nothing and exits `1`.

If `git rev-parse` fails entirely (e.g. not a git repo), the bash script falls back to `REPO_ROOT: $(pwd)`, `BRANCH: unknown` (or similar), `FEATURE_DIR: Not available`, etc., plus `INFO: Unable to determine feature paths`, and still exits `0`. The PowerShell script has no equivalent fallback — `$ErrorActionPreference = 'Stop'` means a git failure terminates it with an error instead.

## Usage in AI Workflows

```bash
# Bash: load paths into shell variables, then read the spec
eval "$(./scripts/bash/get-feature-paths.sh)"
cat "$FEATURE_SPEC"
```

```powershell
# PowerShell: parse the KEY: value output into a hashtable
$paths = ./scripts/powershell/get-feature-paths.ps1 |
    ForEach-Object { $_ -split ': ', 2 } |
    ForEach-Object -Begin { $h = @{} } -Process { $h[$_[0]] = $_[1] } -End { $h }
Get-Content $paths.FEATURE_SPEC
```

## Known Issue (PowerShell)

`scripts/powershell/get-feature-paths.ps1` calls two helper functions — `Get-FeaturePathsEnv` and `Test-FeatureBranch` — that are not defined or exported by `scripts/powershell/common.ps1` (verified: neither name appears anywhere under `scripts/powershell/`, and `common.ps1`'s `Export-ModuleMember` list does not include them). As written, running the script raises a "command not found" error rather than producing output. `scripts/powershell/check-task-prerequisites.ps1` has the same dependency and is equally affected. This needs a fix in `common.ps1` (or the two scripts) before the PowerShell path works; the bash script (`common.sh`'s `get_feature_paths` / `check_feature_branch`) is unaffected.

## Related Scripts

- **create-new-feature.sh / .ps1** — creates the feature branch and directory this script reads
- **check-task-prerequisites.sh / .ps1** — validates the files this script points to actually exist
- **plaesy-analyze.sh / .ps1** — broader project analysis
