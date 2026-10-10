# plaesy features validate

Checks that a feature's development environment is ready: the branch is
current, the tools a spec depends on are present, and nothing about the
prerequisites has drifted.

Source: `scripts/cmd/plaesy/features.go` + `scripts/cmd/plaesy/prereq.go` +
`scripts/internal/featurepath/prereq.go`.

## Purpose

Validate that all required files and directories exist for the current
feature development task, ensuring the development environment is properly
set up before proceeding with implementation.

## Quick Start

```bash

plaesy features validate
plaesy features validate --json
plaesy features validate --help
```

## Validation Checks

### Core Requirements

| Check | Description | Required File/Directory |
|-------|-------------|-------------------------|
| **Feature Branch** | Must be on a feature branch | Branch format: `XXX-feature-name` |
| **Feature Directory** | Feature specification directory | `.plaesy/specs/XXX-feature-name/` |
| **Implementation Plan** | Detailed implementation plan | `.plaesy/specs/XXX-feature-name/plan.md` |

### Optional Documentation

| Check | Description | File/Directory |
|-------|-------------|----------------|
| **Research Documentation** | Background research | `.plaesy/specs/XXX-feature-name/research.md` |
| **Data Model** | Data structure specifications | `.plaesy/specs/XXX-feature-name/data-model.md` |
| **API Contracts** | API specifications | `.plaesy/specs/XXX-feature-name/contracts/` |
| **Quick Start Guide** | Implementation quick start | `.plaesy/specs/XXX-feature-name/quickstart.md` |

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

```bash

$ plaesy features validate
FEATURE_DIR:/home/user/project/.plaesy/specs/001-user-auth
AVAILABLE_DOCS:
  ✓ research.md
  ✗ data-model.md
  ✓ contracts/
  ✓ quickstart.md
```

### Success (JSON)

```bash

$ plaesy features validate --json
{"FEATURE_DIR":"/home/user/project/.plaesy/specs/001-user-auth","AVAILABLE_DOCS":["research.md","contracts/","quickstart.md"]}
```

### Failure (missing required file)

```bash

$ plaesy features validate
Error: feature directory not found: .plaesy/specs/001-user-auth (run /start first to create the feature structure)
```

Exits non-zero; the equivalent message for a missing `plan.md` is
"plan.md not found in .plaesy/specs/<branch> (create plan.md from templates/plan.template.md first)".

## File Structure

```text

.plaesy/specs/XXX-feature-name/
├── plan.md                       # Implementation plan (REQUIRED)
├── research.md                   # Background research (optional)
├── data-model.md                 # Data specifications (optional)
├── quickstart.md                 # Quick start guide (optional)
└── contracts/                    # API contracts (optional)
```

## Integration Example — CI/CD

```bash

if ! plaesy features validate --json; then
    echo "Prerequisites check failed"
    exit 1
fi
```

## Troubleshooting

### Not on a feature branch

```text

Error: not on a feature branch (current: main); feature branches should be named like: 001-feature-name
```

Switch to or create a branch matching `XXX-feature-name` (e.g. via
`plaesy features create`).

### Feature directory or plan.md not found

```text

Error: feature directory not found: .plaesy/specs/001-feature-auth
Error: plan.md not found in .plaesy/specs/001-feature-auth
```

Run `/start` to create the feature structure, then create `plan.md` from
`templates/plan.template.md`.

## Related Commands

- **`plaesy features create`** — Creates feature branches and directories (see [features.md](./features.md))
- **`plaesy features paths`** — Provides feature path information (see [features-paths.md](./features-paths.md))
- **`plaesy analyze`** — Analyzes overall project structure (see [plaesy-analyze.md](./plaesy-analyze.md))
- **`plaesy context update`** — Updates AI context files (see [context-update.md](./context-update.md))
