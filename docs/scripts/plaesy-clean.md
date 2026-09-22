# plaesy clean

**Configurable cleanup command for Plaesy Spec-Kit framework files and directories.**

Source: `scripts/cmd/plaesy/clean.go` + `scripts/internal/cleaner/cleaner.go`.

## Purpose

Safely removes Plaesy Spec-Kit framework files while preserving user code,
with configurable cleanup levels, platform-specific detection, and automatic
backup creation.

## Quick Start

```bash
# Safe cleanup with confirmation
plaesy clean

# Preview what would be removed
plaesy clean --dry-run

# Thorough cleanup for a specific platform
plaesy clean --level thorough --ai claude

# Auto-confirm cleanup (no prompts)
plaesy clean --yes

# Target a specific directory
plaesy clean /path/to/project

# Disable backup creation
plaesy clean --no-backup

# Show detailed progress
plaesy clean --verbose
```

Run `plaesy clean --help` for the exact, current flag list.

## Platform.json Mapping Logic

**Cleanup follows `scripts/configs/platform.json` mapping exactly.**

### What Gets Cleaned

| Component | Source | What Gets Removed |
|-----------|--------|------------------|
| **Plaesy Framework** | Fixed directories | `.plaesy/`, `specs/` directories (`docs/` preserved) |
| **Platform-Specific** | `platforms.{platform}.mapping` | Only files/directories specified in mapping |

### Cleanup by Platform Mapping

**Claude Code Example:**
```json
"claude_code": {
  "mapping": {
    "core": "CLAUDE.md",                    → Removes: CLAUDE.md
    "instructions": ".claude/instructions", → Removes: .claude/instructions/
    "prompts": ".claude/commands"           → Removes: .claude/commands/
  }
}
```

**GitHub Copilot Example:**
```json
"github_copilot": {
  "mapping": {
    "core": ".github/copilot-instructions.md", → Removes: .github/copilot-instructions.md
    "instructions": ".github/instructions",     → Removes: .github/instructions/
    "prompts": ".github/prompts",                → Removes: .github/prompts/
    "chatmodes": ".github/chatmodes"             → Removes: .github/chatmodes/
  }
}
```

### How It Works

1. **Detect Platform**: Auto-detect or specify `--ai <platform>`
2. **Read Mapping**: Get platform-specific paths from `platforms.{platform}.mapping`
3. **Check Existence**: Only remove files/directories that actually exist
4. **Clean Parent Directories**: Remove empty parent directories (e.g., `.claude/` if empty)
5. **Backup**: Create a backup of all files that will be removed (unless `--no-backup`)

### What Gets Preserved

- **User source code** (src/, lib/, components/, etc.)
- **Git repository** (`.git/`)
- **Configuration files** (`.env`, `.gitignore`, etc.)
- **Files not in Plaesy framework or platform mapping**
- **`docs/` directory** — preserved, may contain important project documentation

## Cleanup Levels

| Level | What it Removes | Safety |
|-------|----------------|--------|
| **safe** (default) | Framework files only, preserve user code | Safe |
| **thorough** | Framework + specs, preserve user code | Moderate |
| **complete** | Everything Plaesy-related | Dangerous |

## Backup System

- **Automatic backup**: Creates a timestamped backup before deletion (`--no-backup` to skip)
- **Platform-specific backup**: Backs up platform-specific files based on mapping
- **Configuration backup**: Preserves JSON configuration files
- **Backup location**: `.plaesy-backup-YYYYMMDD-HHMMSS` in target directory

## Output Examples

### Dry Run Output
```
Cleanup Plan (Level: safe, Platforms: claude_code)
═══════════════════════════════════════════════════════════════
Safe cleanup: Framework files only, preserve user code

Plaesy Framework Directories:
  - .plaesy/

Platform-Specific Files (claude_code):
  Based on platform.json mapping for claude_code:
  - CLAUDE.md (file)
  - .claude/instructions/ (directory)
  - .claude/commands/ (directory)

What will be preserved:
  - Your source code (src/, lib/, components/, etc.)
  - Git repository (.git/)
  - User-generated files not in Plaesy directories
  - Configuration files (.env, .gitignore, etc.)
  - Backup will be created before removal

This will permanently delete the directories and files listed above.
Are you sure you want to continue? (y/N):
```

### Successful Cleanup Output
```
Detecting AI platforms for cleanup...
Detected AI platform: claude_code
Using platform-adapted cleanup...

[STEP] Analyzing directories and files to remove based on platform.json mapping...
[INFO] Creating backup: ./.plaesy-backup-20240318-143022
[SUCCESS] Backup created: ./.plaesy-backup-20240318-143022
[STEP] Removing plaesy directories based on platform.json mapping...
[INFO] Removing Plaesy framework directories...
  ✅ Removed Plaesy framework: .plaesy
[INFO] Removing platform-specific files for: claude_code
  ✅ Removed platform claude_code core: CLAUDE.md
  ✅ Removed platform claude_code instructions: .claude/instructions
  ✅ Removed platform claude_code prompts: .claude/commands
  ✅ Removed empty parent directory: .claude
[SUCCESS] Successfully removed 5 items

Plaesy clean completed!

Cleaned directory: /path/to/project

To reinitialize the project:
   plaesy init [--ai=your-choice]
```

## Safety Features

### Pre-Cleanup Validation
1. **Configuration Validation**: Validates `platform.json` before cleanup
2. **Environment Validation**: Checks directory permissions and accessibility
3. **Platform Detection**: Auto-detects platforms to prevent accidental deletion
4. **Dry Run Mode**: Preview changes before execution

## Troubleshooting

1. **Permission Denied**
   ```bash
   ls -la /path/to/project
   chmod u+w /path/to/project    # Linux/macOS only
   ```

2. **Configuration Not Found**
   ```bash
   ls -la scripts/configs/platform.json
   plaesy config validate
   ```

3. **Config package missing / not built**
   ```bash
   cd scripts && go build ./...
   ```

### Debug Mode

```bash
PLAESY_DEBUG=true plaesy clean --dry-run
```

### Manual Platform Detection

```bash
plaesy config list-platforms
plaesy config get-platform-config claude_code detection_patterns
```

## Integration Examples

### CI/CD Integration
```bash
plaesy clean --yes --level safe --no-backup
```

### Pre-commit Hook
```bash
#!/bin/sh
# .git/hooks/pre-commit
plaesy clean --yes --dry-run
```

## Version Information

- **Current Version**: 0.0.1
- **Implementation**: `scripts/internal/cleaner/cleaner.go` (single Go implementation)
- **Configuration Format**: platform.json v1.0
- **Supported Platforms**: All platforms defined in `platform.json`

## Security Considerations

- **Path Validation**: All paths are validated before use
- **Permission Checks**: Validates read/write permissions before operations
- **Backup Safety**: Creates backup before any deletion (unless `--no-backup`)
- **Confirmation Required**: User confirmation required unless `--yes` is passed
- **Platform Mapping**: Only removes files specified in `platform.json` mapping
