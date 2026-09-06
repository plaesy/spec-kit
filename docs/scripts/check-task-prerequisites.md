# Check Task Prerequisites Script

**Development environment validation for Plaesy Spec-Kit projects.**

## Purpose

Validate that all required files and directories exist for the current feature development task, ensuring the development environment is properly set up before proceeding with implementation.

## Quick Start

```bash
# Bash
./scripts/bash/check-task-prerequisites.sh
./scripts/bash/check-task-prerequisites.sh --json
./scripts/bash/check-task-prerequisites.sh --help
```

```powershell
# PowerShell
./scripts/powershell/check-task-prerequisites.ps1
./scripts/powershell/check-task-prerequisites.ps1 -Json
Get-Help ./scripts/powershell/check-task-prerequisites.ps1
```

Note: the PowerShell script has no dedicated `--help` handler; use `Get-Help` (from `[CmdletBinding()]`) for parameter info.

## Validation Checks

### Core Requirements
| Check | Description | Required File/Directory |
|-------|-------------|-------------------------|
| **Feature Branch** | Must be on a feature branch | Branch format: `XXX-feature-name` |
| **Feature Directory** | Feature specification directory | `specs/XXX-feature-name/` |
| **Implementation Plan** | Detailed implementation plan | `specs/XXX-feature-name/plan.md` |

### Optional Documentation
| Check | Description | File/Directory |
|-------|-------------|----------------|
| **Research Documentation** | Background research | `specs/XXX-feature-name/research.md` |
| **Data Model** | Data structure specifications | `specs/XXX-feature-name/data-model.md` |
| **API Contracts** | API specifications | `specs/XXX-feature-name/contracts/` |
| **Quick Start Guide** | Implementation quick start | `specs/XXX-feature-name/quickstart.md` |

## Options

| Flag (bash) | Flag (PowerShell) | Description |
|-------------|--------------------|--------------|
| `--json` | `-Json` | Emit a single-line JSON object instead of plain text |
| `--help` / `-h` | — (use `Get-Help`) | Print usage and exit |

Both scripts resolve feature paths via their shared helper (`common.sh` / `common.ps1`), verify the current branch matches `XXX-feature-name`, then check for `plan.md` (required) and `research.md`, `data-model.md`, `contracts/`, `quickstart.md` (optional).

## Output Format

### Success (plain text)
```
$ ./scripts/bash/check-task-prerequisites.sh
FEATURE_DIR:/home/user/project/specs/001-user-auth
AVAILABLE_DOCS:
  ✓ research.md
  ✗ data-model.md
  ✓ contracts/
  ✓ quickstart.md
```

### Success (JSON — identical shape for both scripts)
```
$ ./scripts/bash/check-task-prerequisites.sh --json
{"FEATURE_DIR":"/home/user/project/specs/001-user-auth","AVAILABLE_DOCS":["research.md","contracts/","quickstart.md"]}
```

### Failure (missing required file)
```
$ ./scripts/bash/check-task-prerequisites.sh
ERROR: Feature directory not found: specs/001-user-auth
Run /specify first.
```
The PowerShell script prints the equivalent message ("Run /specify first to create the feature structure.") and exits with code 1; same for a missing `plan.md`.

## File Structure

```
specs/XXX-feature-name/
├── plan.md                       # Implementation plan (REQUIRED)
├── research.md                   # Background research (optional)
├── data-model.md                 # Data specifications (optional)
├── quickstart.md                 # Quick start guide (optional)
└── contracts/                    # API contracts (optional)
```

## Integration Example — CI/CD

```bash
# Bash (e.g. CI pipeline step)
if ! ./scripts/bash/check-task-prerequisites.sh --json; then
    echo "Prerequisites check failed"
    exit 1
fi
```

```powershell
# PowerShell equivalent
if (-not (./scripts/powershell/check-task-prerequisites.ps1 -Json)) {
    Write-Error "Prerequisites check failed"
    exit 1
}
```

## Troubleshooting

### Not on a feature branch
```
ERROR: Not on a feature branch. Current branch: main
```
Switch to or create a branch matching `XXX-feature-name` (e.g. via `create-new-feature.sh` / `.ps1`).

### Feature directory or plan.md not found
```
ERROR: Feature directory not found: specs/001-feature-auth
ERROR: plan.md not found in specs/001-feature-auth
```
Run `/specify` to create the feature structure, then `/plan` to generate `plan.md`.

### Bash script not executable (Linux/macOS)
```
chmod +x scripts/bash/check-task-prerequisites.sh
```
Not applicable on Windows/PowerShell — no execute-bit is required, but the execution policy must allow local scripts (`Set-ExecutionPolicy -Scope Process RemoteSigned` if blocked).

## Related Scripts

- **create-new-feature.sh / .ps1** - Creates feature branches and directories
- **get-feature-paths.sh / .ps1** - Provides feature path information
- **plaesy-analyze.sh / .ps1** - Analyzes overall project structure
- **update-agent-context.sh / .ps1** - Updates AI context files
