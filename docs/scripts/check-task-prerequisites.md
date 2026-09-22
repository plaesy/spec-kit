# plaesy check-task-prerequisites

**Development environment validation for Plaesy Spec-Kit projects.**

Source: `scripts/cmd/plaesy/prereq.go` + `scripts/internal/featurepath/prereq.go`.

## Purpose

Validate that all required files and directories exist for the current
feature development task, ensuring the development environment is properly
set up before proceeding with implementation.

## Quick Start

```bash
plaesy check-task-prerequisites
plaesy check-task-prerequisites --json
plaesy check-task-prerequisites --help
```

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

| Flag | Description |
|------|-------------|
| `--json` | Emit a single-line JSON object instead of plain text |
| `--help` / `-h` | Print usage and exit |

The command resolves feature paths via `internal/common.GetFeaturePaths()`,
verifies the current branch matches `XXX-feature-name`, then checks for
`plan.md` (required) and `research.md`, `data-model.md`, `contracts/`,
`quickstart.md` (optional).

## Output Format

### Success (plain text)
```
$ plaesy check-task-prerequisites
FEATURE_DIR:/home/user/project/specs/001-user-auth
AVAILABLE_DOCS:
  ✓ research.md
  ✗ data-model.md
  ✓ contracts/
  ✓ quickstart.md
```

### Success (JSON)
```
$ plaesy check-task-prerequisites --json
{"FEATURE_DIR":"/home/user/project/specs/001-user-auth","AVAILABLE_DOCS":["research.md","contracts/","quickstart.md"]}
```

### Failure (missing required file)
```
$ plaesy check-task-prerequisites
Error: feature directory not found: specs/001-user-auth (run /start first to create the feature structure)
```
Exits non-zero; the equivalent message for a missing `plan.md` is
"plan.md not found in specs/<branch> (create plan.md from templates/plan.template.md first)".

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
if ! plaesy check-task-prerequisites --json; then
    echo "Prerequisites check failed"
    exit 1
fi
```

## Troubleshooting

### Not on a feature branch
```
Error: not on a feature branch (current: main); feature branches should be named like: 001-feature-name
```
Switch to or create a branch matching `XXX-feature-name` (e.g. via `plaesy create-new-feature`).

### Feature directory or plan.md not found
```
Error: feature directory not found: specs/001-feature-auth
Error: plan.md not found in specs/001-feature-auth
```
Run `/start` to create the feature structure, then create `plan.md` from
`templates/plan.template.md`.

## Related Commands

- **`plaesy create-new-feature`** — Creates feature branches and directories (see [create-new-feature.md](./create-new-feature.md))
- **`plaesy get-feature-paths`** — Provides feature path information (see [get-feature-paths.md](./get-feature-paths.md))
- **`plaesy analyze`** — Analyzes overall project structure (see [plaesy-analyze.md](./plaesy-analyze.md))
- **`plaesy update-agent-context`** — Updates AI context files (see [update-agent-context.md](./update-agent-context.md))
