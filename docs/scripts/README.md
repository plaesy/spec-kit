# Plaesy Scripts Documentation

**Complete guide to the `plaesy` CLI.**

**Framework Version:** 0.0.1

The framework's automation is a single cross-platform Go binary, `plaesy`
(source: `scripts/cmd/plaesy` + `scripts/internal/`). There is no longer a
separate bash and PowerShell implementation to keep in sync — one binary,
one set of behavior, on Linux/macOS/Windows. Build it with:

```bash

cd scripts && go build -o plaesy ./cmd/plaesy
```

## Quick Navigation

### **Command Documentation**

| Doc | Command(s) documented | Purpose | AI Priority |
|-----|------------------------|---------|-------------|
| **[plaesy-init.md](./plaesy-init.md)** | `plaesy init` | Project initialization & AI platform setup | **HIGH** |
| **[features.md](./features.md)** | `plaesy features`, `plaesy features create` | Feature branch creation + listing | **HIGH** |
| **[features-paths.md](./features-paths.md)** | `plaesy features paths` | Feature path resolution | **CRITICAL** |
| **[features-validate.md](./features-validate.md)** | `plaesy features validate` | Development validation | **CRITICAL** |
| **[context-update.md](./context-update.md)** | `plaesy context update` | AI context synchronization | **CRITICAL** |
| **[images-create.md](./images-create.md)** | `plaesy images create` | Image asset generation | **HIGH** |
| **[platforms.md](./platforms.md)** | `plaesy platforms <subcommand>` | AI platform detection, listing, inspection | **MEDIUM** |
| **[config-manager.md](./config-manager.md)** | `plaesy config <subcommand>` | Platform configuration file management | **MEDIUM** |
| **[plaesy-clean.md](./plaesy-clean.md)** | `plaesy clean` | Project cleanup and reset | **MEDIUM** |
| **[plaesy-analyze.md](./plaesy-analyze.md)** | `plaesy analyze` | Project structure analysis | **CRITICAL** |
| **[install.md](./install.md)** | `plaesy install` / `uninstall` / `repair` / `upgrade` / `status` | Binary installation & self-management | **HIGH** |
| **[common.md](./common.md)** | `internal/common` package | Shared logging/validation/git helpers used by every command | **CRITICAL** |
| **[plaesy-graph.md](./plaesy-graph.md)** | `plaesy graph` | Repo knowledge-graph builder (nodes/edges, queries, impact check) | **MEDIUM** |
| **[plaesy-search.md](./plaesy-search.md)** | `plaesy search` | Semantic (meaning-based) search over extracted symbols — anti-duplication aid | **MEDIUM** |
| **[plaesy-trim.md](./plaesy-trim.md)** | `plaesy trim <subcommand>` | Token/context compression (command output + memory files) | **MEDIUM** |

`platform-detector.md` was removed: it documented a Windows-only PowerShell
auto-detection helper (`platform-detector.ps1`) that has no Go equivalent —
`plaesy platforms detect` now covers platform detection on every OS.

### **Quick Start for AI Assistants**

**MANDATORY SEQUENCE** — Always run this when encountering a Plaesy project:

```bash

# 1. Understand project structure (CRITICAL)
#    Analyze does not copy instructions; plaesy init performs instruction auto-load.
plaesy analyze

# 2. Get current feature context
plaesy features paths

# 3. Validate development setup
plaesy features validate

# 4. Update AI context understanding
plaesy context update
```

**📌 Note**: `plaesy init` (not `plaesy analyze`) detects project technologies
and auto-copies relevant instructions from `instructions/mapping.json` to
`.plaesy/instructions/`. Run `plaesy init` first to set up the instruction
set, then `plaesy analyze` for project structure analysis.

The same four commands run identically on Linux, macOS, and Windows — no
separate PowerShell invocation is needed anymore.

## AI Assistant Essential Information

### **Critical Files for AI Understanding**

1. **`.plaesy/analysis/project.json`** - AI-optimized project summary
2. **`.plaesy/analysis/project.structure.json`** - Complete project structure
3. **`.plaesy/analysis/project.graph.json`** - Dependency graph (nodes + edges; built by `plaesy analyze` automatically, queryable via `plaesy graph`)
4. **`.plaesy/specs/[feature-name]/plan.md`** - Current feature implementation plan
5. **Platform-specific AI contexts** (`CLAUDE.md`, `.github/copilot-instructions.md`, etc.)

### **Platform Detection Logic**

Plaesy automatically detects AI platforms and applies appropriate configurations:

- **Claude Code** → `CLAUDE.md`, `.claude/` directory
- **GitHub Copilot** → `.github/copilot-instructions.md`
- **Cursor AI** → `.cursor/rules/specify-rules.mdc`
- **Windsurf AI** → Configuration files
- **Continue.dev** → `.continue/` directory
- **And 10+ other platforms** with optimized configurations

## Command Architecture

### **Single Binary, Registry-based Commands**

`plaesy` is a [cobra](https://github.com/spf13/cobra) CLI. Every subcommand
lives in its own file under `scripts/cmd/plaesy/` and registers itself via
`init() { register(new<Name>Cmd()) }` (see `scripts/cmd/plaesy/registry.go`)
— so `main.go` never needs editing when a command is added or changed. Each
command's actual logic lives in a matching package under
`scripts/internal/<name>/`, reusing `scripts/internal/common` for logging,
git-repo resolution, and validation.

List every available command:

```bash

plaesy --help
```

Get flags/usage for one command:

```bash

plaesy <command> --help
```

### **Configuration-Driven Architecture**

Platform behavior (which files go where for Claude Code, Copilot, Cursor,
etc.) is centralized in `scripts/configs/platform.json` and read via
`plaesy config <subcommand>` — used internally by `plaesy init` and
`plaesy clean`:

```json

{
  "plaesy": {
    "structure": {
      "base_directory": ".plaesy",
      "core_directories": ["memory"],
      "project_directories": ["docs", "specs"]
    },
    "mapping": {
      "core": "instructions/plaesy.instructions.md",
      "instructions": "instructions/*",
      "prompts": "prompts/*",
      "agents": "agents/*"
    }
  },
  "platforms": {
    "claude": {
      "name": "Claude Code",
      "mapping": {
        "core": "CLAUDE.md",
        "instructions": ".claude/instructions",
        "prompts": ".claude/commands",
        "agents": ".claude/roles"
      },
      "detection_patterns": [".claude/", "CLAUDE.md"]
    }
  }
}
```

### **Multi-Platform Support**

Enhanced platform support with multiple AI platforms:

1. **Claude Code** - Native tool access, full automation
2. **GitHub Copilot** - VS Code integration, code completion
3. **Cursor AI** - IDE integration, refactoring support
4. **Windsurf AI** - Platform-specific, custom workflows
5. **Cline** - Tool integration, automated workflows
6. **Deepseek** - API integration, enhanced responses
7. **Kilo Code** - Code assistance, development support
8. **Qoder** - AI optimization, quality improvement
9. **Trae AI** - Advanced features, specialized workflows
10. **Continue.dev** - VS Code extension, AI assistance
11. **Tabnine** - AI completion, IDE integration
12. **Codeium** - IDE tools, AI-powered assistance
13. **CodeWhisperer** - AWS-powered, code generation
14. **Studio Bot** - Android development, Google integration
15. **Replit Ghostwriter** - Cloud IDE, AI assistance
16. **LlamaIndex** - AI framework, application development
17. **Ollama** - Local AI, offline capabilities
18. **LM Studio** - Local AI, model management
19. **Generic AI** - Universal compatibility, basic automation

## Common Workflows

### **New Feature Development**

```bash

# 1. Create feature
plaesy features create "Feature description"

# 2. Run critical sequence
plaesy analyze
plaesy features paths
plaesy features validate
plaesy context update
```

### **AI Platform Setup**

```bash

# Initialize with specific AI platform
plaesy init . --ai claude
```

### **Project Cleanup**

```bash

# Safe cleanup with preview
plaesy clean --dry-run

# Complete cleanup for specific platform
plaesy clean --level thorough --ai claude --yes
```

### **Installing the CLI**

There is no published release yet — build from source, then let the binary
install itself to a well-known bin directory:

```bash

cd scripts && go build -o plaesy ./cmd/plaesy
./plaesy install    # copies itself into ~/.local/bin (or the Windows equivalent)
plaesy status        # verify install location, PATH, and version
```

## Integration Examples

### **CI/CD Integration**

```yaml
- name: Setup Plaesy
  run: |
    plaesy init --ai claude
    plaesy context update

- name: Analyze Project
  run: plaesy analyze

- name: Cleanup
  run: plaesy clean --yes --level safe
```

### **Docker Integration**

```dockerfile

# Build Plaesy from source and install it
COPY scripts/ /opt/plaesy/scripts/
RUN cd /opt/plaesy/scripts && go build -o /usr/local/bin/plaesy ./cmd/plaesy

# Initialize with Claude Code
RUN plaesy init --ai claude --target /app

WORKDIR /app
CMD ["plaesy", "analyze"]
```

```yaml

version: '3.8'
services:
  app:
    build: .
    volumes:
      - .:/app
      - plaesy-cache:/root/.cache/claude
    environment:
      - AI_PLATFORM=claude
    command: plaesy context update

```

### **VS Code Integration**

#### .vscode/tasks.json

```json

{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Plaesy: Analyze Project",
            "type": "shell",
            "command": "plaesy analyze",
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "Plaesy: Update Context",
            "type": "shell",
            "command": "plaesy context update",
            "group": "build"
        }
    ]
}
```

## Troubleshooting

### **Common Issues**

1. **`plaesy: command not found`**

   ```bash
   # Confirm the binary landed on PATH after `plaesy install`
   plaesy status
   ```

1. **Configuration Not Found**

   ```bash
   # Check platform.json exists
   ls -la scripts/configs/platform.json

   # Validate configuration
   plaesy config validate
   ```

### **Debug Mode**

Set `PLAESY_DEBUG=true` to enable debug-level log lines, and `PLAESY_LOG_FILE=<path>`
to additionally write logs to a file (see `scripts/internal/common/logging.go`):

```bash

PLAESY_DEBUG=true plaesy analyze
```

## Version Information

- **Framework Version:** 0.0.1
- **Implementation**: single Go binary (`plaesy`), Go standard library only — no bash/PowerShell scripts remain
- **Configuration**: Centralized, mapping-based (`scripts/configs/platform.json`)
- **Documentation**: Complete with examples and troubleshooting

## Security Considerations

- **Path Validation**: All paths validated before use
- **Permission Checks**: Read/write permissions verified
- **Input Sanitization**: User inputs validated and sanitized
- **Backup Safety**: `plaesy clean` creates backups before destructive operations by default
- **Error Handling**: Comprehensive error handling prevents data loss

## Migration Guide

### **From the bash/PowerShell scripts to the Go CLI**

The `scripts/bash/*.sh` and `scripts/powershell/*.ps1` scripts have been
removed entirely and replaced by the single `plaesy` binary. All prior
command-line options and behavior were preserved 1:1 where the Go port
found a direct equivalent; a few deliberate deviations exist (documented
in each command's own doc page) — most notably `plaesy repair` and
`plaesy upgrade` are not yet implemented (the old bash self-update flow
depended on a `git clone`-based install model that no longer applies) and
print a clear "not yet implemented" message instead.

If you have an existing project that was set up with the old scripts,
re-run the equivalent Go commands — the on-disk `.plaesy/` structure and
platform files are unchanged:

```bash

# Re-initialize to update configuration
plaesy init . --ai your_current_platform

# Update context
plaesy context update

# Verify setup
plaesy analyze
```
