# plaesy get-feature-paths

**Feature branch and path resolution command.**

Source: `scripts/cmd/plaesy/feature_paths.go` + `scripts/internal/featurepath/prereq.go` (`GetPathsReport`).

## Purpose

Resolves the current feature branch and prints the paths to its spec/plan/tasks
files, without creating anything. Used by other commands and AI workflows to
locate the active feature's documents.

## Quick Start

```bash
plaesy get-feature-paths
```

Takes no flags or arguments.

## Output Format

```
REPO_ROOT: /home/user/project
BRANCH: 001-user-authentication
FEATURE_DIR: /home/user/project/.plaesy/specs/001-user-authentication
FEATURE_SPEC: /home/user/project/.plaesy/specs/001-user-authentication/spec.md
IMPL_PLAN: /home/user/project/.plaesy/specs/001-user-authentication/plan.md
TASKS: /home/user/project/.plaesy/specs/001-user-authentication/tasks.md
```

Paths are not checked for existence — they are computed from the repo root
and current branch name only (`.plaesy/specs/<branch>/...`).

## Behavior When Not on a Feature Branch

This command always follows the old bash script's forgiving behavior (the
Go port standardized on it, rather than porting the PowerShell script's
stricter one — see "Migration note" below): it still prints all six lines
(using the actual current branch, e.g. `BRANCH: main`), then appends
`INFO: Not on a feature branch (format: XXX-feature-name)` and exits `0`.

If `git rev-parse` fails entirely (e.g. not a git repo), it falls back to
`REPO_ROOT: <cwd>`, `BRANCH: unknown`, `FEATURE_DIR: Not available`, etc.,
plus `INFO: Unable to determine feature paths`, and still exits `0`.

## Usage in AI Workflows

```bash
# Load paths into shell variables (for a script that wants to source them)
eval "$(plaesy get-feature-paths | sed -n 's/^\([A-Z_]*\): \(.*\)$/\1=\2/p')"
```

For programmatic use from other Go code in this repo, call
`common.GetFeaturePaths()` directly instead of shelling out — see
[common.md](./common.md).

## Migration note

The old PowerShell script (`get-feature-paths.ps1`) diverged from bash here:
it printed nothing and exited `1` when not on a feature branch, and had a
known bug (it called two helper functions, `Get-FeaturePathsEnv` and
`Test-FeatureBranch`, that were never defined anywhere in
`scripts/powershell/`, so the script actually errored out with "command not
found" rather than running at all). The Go port fixes that bug by
construction — there is one implementation, not two to keep in sync — and
adopts the bash script's always-exit-0, always-print-something behavior as
the single correct behavior going forward.

## Related Commands

- **`plaesy create-new-feature`** — creates the feature branch and directory this command reads (see [create-new-feature.md](./create-new-feature.md))
- **`plaesy check-task-prerequisites`** — validates the files this command points to actually exist (see [check-task-prerequisites.md](./check-task-prerequisites.md))
- **`plaesy analyze`** — broader project analysis (see [plaesy-analyze.md](./plaesy-analyze.md))
