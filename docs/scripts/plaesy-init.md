# plaesy init

**Project initialization and AI platform configuration command.**

**Priority:** HIGH - Required for Plaesy project setup

Source: `scripts/cmd/plaesy/init.go` + `scripts/internal/scaffold/`.

## Purpose

Initializes Plaesy projects with AI platform configuration and template
creation. This command creates the foundation for AI-assisted development
workflows.

## Usage

```bash
# Initialize in current directory, no platform-specific files
plaesy init .

# Initialize with a specific AI platform
plaesy init . --ai claude_code      # Claude Code
plaesy init . --ai cursor_ai        # Cursor AI
plaesy init . --ai github_copilot   # GitHub Copilot

# Create project in a specific (new) directory
plaesy init my-app --ai claude_code

# Override where templates/instructions/prompts/chatmodes are read from
# (defaults to walking up from the current directory to find a `templates/`
# dir, or $PLAESY_HOME)
plaesy init . --ai claude_code --plaesy-home /path/to/spec-kit
```

Run `plaesy init --help` for the exact, current flag list.

> **Changed from the old bash/PowerShell scripts**: the old scripts offered
> an interactive platform-selection menu (with a timeout, for non-interactive
> environments) and an `--all-instructions` flag to force-copy every
> instruction file regardless of detected tech stack. Neither exists in the
> Go port yet — `--ai` is required to configure platform-specific files, and
> instruction selection is always the selective (detected-tech-stack) mode
> described below. If you relied on either of those, say so and they can be
> added back.

## Key Features

### Configuration-Driven Setup
- **Mapping-Based Configuration**: Uses `scripts/configs/platform.json` (see [config-manager.md](./config-manager.md))
- **Core AI Config**: Creates `CLAUDE.md` at project root from `instructions/plaesy.instructions.md`
- **Instruction File Copy**: Copies selected instructions to `.plaesy/instructions/` (flat structure) for per-project documentation
- **Chatmode Copy**: Copies all chatmodes to `.plaesy/roles/` (flat structure, platform-agnostic)
- **Prompt Copy**: Copies prompts to platform-specific locations (`.claude/commands/`, `.cursor/rules/`, etc.)
- **Exclude Pattern Support**: Respects exclude patterns in mapping
- **Dynamic Structure Creation**: Creates directories based on `platform.json`'s `structure` section

### Supported AI Platforms
| Platform | Detection Files | Features |
|----------|----------------|----------|
| **claude_code** | `.claude/` directory | Native tool access |
| **github_copilot** | `.github/copilot-instructions.md` | VS Code integration |
| **cursor_ai** | `.cursor/rules/` directory | IDE integration |
| **windsurf_ai** | Configuration files | Platform-specific |
| *(and more — see `scripts/configs/platform.json` for the full, current list)* | | |

## Template-Driven File Creation

When `plaesy init` runs, bootstrap files are created from templates in
`templates/` (or an empty file as fallback):

### Context File Creation
- **Source**: `templates/context.template.md` (when available)
- **Destination**: `.plaesy/context.md` (project root)
- **Purpose**: Session-specific state tracking (max 100 lines, archive to memory as needed)

### Memory Index File Creation
- **Source**: `templates/memory.template.md` (when available)
- **Destination**: `.plaesy/memory.md` (project root)
- **Purpose**: Index and quick links to knowledge files in `.plaesy/memory/` and instruction copies in `.plaesy/instructions/`

### Loop State File Creation
- **Source**: `templates/state.template.json` (when available)
- **Destination**: `.plaesy/state.json` (project root)
- **Purpose**: Autonomous loop configuration and quality gates
- **Timestamp Handling**: Placeholder `[TIMESTAMP]` replaced with the current UTC time

### Task Management Documentation
When creating task structure:
- Creates task status directories (backlog, todo, doing, done, blocked) under `.plaesy/tasks/`
  — this is the single source of truth for tasks (one file per task, YAML frontmatter,
  folder = status). `/start`, `/continue`, and `/loop` all read/write here.
- Creates `.plaesy/tasks/README.md` with a quick reference
- Task management instructions are auto-loaded from `tasks.instructions.md` during init

---

## Instruction Files Auto-Copy Feature

When `plaesy init` runs, instruction files are **automatically copied to
`.plaesy/instructions/`** (flat structure) for per-project documentation:

### How It Works
1. **Source**: Reads from `instructions/*.instructions.md` (framework reference)
2. **Root Config**: Creates `CLAUDE.md` at project root from `instructions/plaesy.instructions.md` (session startup protocol)
3. **Detection**: Analyzes the project's technology stack (package.json, Cargo.toml, etc. — the same detection `plaesy detect-stack` uses)
4. **Selective Copy**: Only copies instructions matching detected technologies + the always-load core set
5. **Destination**: Copies to `.plaesy/instructions/` (flat, no subfolders) with a renamed extension
   - Always-load (from `instructions/mapping.json`'s `always_load` array): `plaesy.instructions.md`,
     `plaesy-trim.instructions.md`, `plaesy-graph.instructions.md`, `tasks.instructions.md`,
     `quality-gates.instructions.md`, `error-recovery.instructions.md`, `date-system.instructions.md`
     → each copied to `.plaesy/instructions/<name>.md`
   - Detected: `nextjs.instructions.md` → `.plaesy/instructions/nextjs.md`
   - Not detected: `ruby-on-rails.instructions.md` → NOT copied

### File Organization (Example for Next.js + React Project)
```
PROJECT ROOT
├── CLAUDE.md                 # Root AI config (from instructions/plaesy.instructions.md)
├── .plaesy/context.md        # Session state (from template, can be edited)
└── .plaesy/instructions/
    ├── plaesy.md             # Always loaded (universal rules)
    ├── quality-gates.md      # Always loaded
    ├── error-recovery.md     # Always loaded
    ├── tasks.md              # Always loaded
    ├── date-system.md        # Always loaded
    ├── nextjs.md             # Copied (detected Next.js)
    └── reactjs.md            # Copied (detected React)
```

### Why This Feature?
- **Self-Contained**: Project documentation stays with the project
- **Single Source**: One copy per project, not duplicated across platforms
- **Lightweight**: Only relevant instructions copied based on tech stack
- **AI-Friendly**: Stored in the standard `.plaesy/instructions/` location for AI context loading

---

## Platform.json Mapping Logic

`plaesy init` follows `scripts/configs/platform.json`'s mapping for
instructions, prompts, and chatmodes:

```json
{
  "plaesy": {
    "mapping": {
      "core": "instructions/plaesy.instructions.md",    // SOURCE: universal rules
      "instructions": "instructions/*",                 // SOURCE: all instructions
      "prompts": "prompts/*",                           // SOURCE: all workflows
      "chatmodes": "chatmodes/*"                        // SOURCE: all chat modes
    }
  },
  "platforms": {
    "claude_code": {
      "mapping": {
        "prompts": ".claude/commands",
        "chatmodes": ".claude/roles"
      }
    }
  }
}
```

### File Copying Strategy
- **Universal files** (always-load, per `instructions/mapping.json`): Always copied to
  `.plaesy/instructions/` regardless of platform
- **Instruction files**: Copied to `.plaesy/instructions/` based on detected tech stack
- **Prompt/Workflow files**: Copied to platform-specific locations (`.claude/commands/`, etc.)
- **Chat modes**: Always copied to `.plaesy/roles/`, the same path on every platform

### Configuration Manager Integration

`plaesy init` uses the same `scripts/internal/config` package that backs
`plaesy config`, for:

1. Platform lookup: `plaesy config get-platform-config`
2. Mapping values: `plaesy config get-mapping-value`
3. Exclude patterns: `plaesy config get-mapping-excludes`
4. Structure config: `plaesy config get-plaesy-structure`

See [config-manager.md](./config-manager.md) for details.

## File Structure After Init

```
project/
├── CLAUDE.md                    # AI config (from instructions/plaesy.instructions.md)
│
├── .plaesy/
│   ├── context.md               # Current session state
│   ├── memory.md                # Memory index
│   ├── instructions/
│   │   ├── plaesy.md            # ← AI reads THIS at session start
│   │   ├── quality-gates.md     # Always loaded
│   │   ├── error-recovery.md    # Always loaded
│   │   ├── tasks.md             # Always loaded
│   │   ├── date-system.md       # Always loaded
│   │   └── [tech]*.md           # Tech-specific instructions (detected)
│   │
│   └── tasks/
│       ├── backlog/
│       ├── todo/
│       ├── doing/
│       ├── done/
│       └── blocked/
│
├── .claude/
│   ├── commands/                # Prompts copied here
│   └── roles/                   # Chat modes
│
└── ... (platform-specific dirs based on the selected AI platform)
```

## Troubleshooting

1. **Unknown `--ai` value**
   ```bash
   plaesy config list-platforms
   ```

2. **Configuration not found**
   ```bash
   ls -la scripts/configs/platform.json
   plaesy config validate
   ```

3. **Templates/instructions not found**
   Pass `--plaesy-home` explicitly, or set `PLAESY_HOME`, pointing at a
   checkout of this repo (the directory containing `templates/`,
   `instructions/`, `prompts/`, `chatmodes/`).

### Debug Mode

```bash
PLAESY_DEBUG=true plaesy init . --ai claude_code
```

## CI/CD Integration

```yaml
- name: Initialize Plaesy
  run: plaesy init . --ai claude_code
```

## Version Information

- **Current Version**: 0.0.1
- **Configuration Format**: platform.json v1.0
- **Key Feature**: Auto-copy detected-stack instructions to `.plaesy/instructions/` (flat structure)

## Migration Guide

If you need to manually update an existing project's configuration:

```bash
plaesy init . --ai your_current_platform
```

This updates any missing configuration files while preserving your existing setup.
