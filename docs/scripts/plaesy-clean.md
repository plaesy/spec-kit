# plaesy-clean.sh / plaesy-clean.ps1

**Configurable cleanup utility for Plaesy Spec-Kit framework files and directories.**

**Script Version: 1.0.0**

## Purpose

Safely remove Plaesy Spec-Kit framework files while preserving user code, with configurable cleanup levels, platform-specific detection, and automatic backup creation.

## Quick Start

### Bash
```bash
# Safe cleanup with confirmation
./scripts/bash/plaesy-clean.sh

# Preview what would be removed
./scripts/bash/plaesy-clean.sh --dry-run

# Thorough cleanup for specific platform
./scripts/bash/plaesy-clean.sh --level thorough --ai claude

# Auto-confirm cleanup (no prompts)
./scripts/bash/plaesy-clean.sh --yes
```

### Windows PowerShell
```powershell
# Safe cleanup with confirmation
.\scripts\powershell\plaesy-clean.ps1

# Preview what would be removed
.\scripts\powershell\plaesy-clean.ps1 -DryRun

# Thorough cleanup for specific platform
.\scripts\powershell\plaesy-clean.ps1 -Level thorough -AI claude

# Auto-confirm cleanup (no prompts)
.\scripts\powershell\plaesy-clean.ps1 -Yes

# PowerShell parameter examples
.\scripts\powershell\plaesy-clean.ps1 -Help
.\scripts\powershell\plaesy-clean.ps1 -Level complete -NoBackup -Verbose
.\scripts\powershell\plaesy-clean.ps1 -AI claude_code -Target .\my-project
```

## Platform.json Mapping Logic

**IMPORTANT: Script follows platform.json mapping exactly for cleanup**

### What Gets Cleaned

| Component | Source | What Gets Removed |
|-----------|--------|------------------|
| **Plaesy Framework** | Fixed directories | `.plaesy/`, `specs/` directories (docs/ preserved) |
| **Platform-Specific** | `platforms.{platform}.mapping` | Only files/directories specified in mapping |

### Cleanup by Platform Mapping

**Claude Code Example:**
```json
"claude_code": {
  "mapping": {
    "core": "CLAUDE.md",                → Removes: CLAUDE.md
    "instructions": ".claude/instructions", → Removes: .claude/instructions/
    "prompts": ".claude/commands",         → Removes: .claude/commands/
    "chatmodes": ".claude/roles"           → Removes: .claude/roles/
  }
}
```

**GitHub Copilot Example:**
```json
"github_copilot": {
  "mapping": {
    "core": ".github/copilot-instructions.md", → Removes: .github/copilot-instructions.md
    "instructions": ".github/instructions",   → Removes: .github/instructions/
    "prompts": ".github/prompts",              → Removes: .github/prompts/
    "chatmodes": ".github/chatmodes"           → Removes: .github/chatmodes/
  }
}
```

### How It Works

1. **Detect Platform**: Auto-detect or specify `--ai <platform>`
2. **Read Mapping**: Get platform-specific paths from `platforms.{platform}.mapping`
3. **Check Existence**: Only remove files/directories that actually exist
4. **Clean Parent Directories**: Remove empty parent directories (e.g., `.claude/` if empty)
5. **Backup**: Create backup of all files that will be removed

### What Gets Preserved

- **User source code** (src/, lib/, components/, etc.)
- **Git repository** (.git/)
- **Configuration files** (.env, .gitignore, etc.)
- **Files not in Plaesy framework or platform mapping**
- **docs/ directory** - Preserved as it may contain important project documentation

## Cleanup Levels

| Level | What it Removes | Safety |
|-------|----------------|--------|
| **safe** | Framework files only, preserve user code | **Safe** |
| **thorough** | Framework + specs, preserve user code | **Moderate** |
| **complete** | Everything Plaesy-related | **Dangerous** |

## Command Options

### Basic Options
```bash
# Target directory (default: current directory)
./scripts/bash/plaesy-clean.sh /path/to/project
./scripts/powershell/plaesy-clean.ps1 -TargetDir /path/to/project

# Auto-confirm deletion (skip confirmation)
./scripts/bash/plaesy-clean.sh --yes
./scripts/powershell/plaesy-clean.ps1 -Yes

# Show detailed progress
./scripts/bash/plaesy-clean.sh --verbose
./scripts/powershell/plaesy-clean.ps1 -Verbose
```

### Cleanup Control
```bash
# Cleanup levels
./scripts/bash/plaesy-clean.sh --level safe      # Default
./scripts/bash/plaesy-clean.sh --level thorough
./scripts/bash/plaesy-clean.sh --level complete

./scripts/powershell/plaesy-clean.ps1 -Level safe      # Default
./scripts/powershell/plaesy-clean.ps1 -Level thorough
./scripts/powershell/plaesy-clean.ps1 -Level complete
```

### Platform Specification
```bash
# Clean specific platform only
./scripts/bash/plaesy-clean.sh --ai claude
./scripts/bash/plaesy-clean.sh --ai github_copilot
./scripts/bash/plaesy-clean.sh --ai cursor_ai

./scripts/powershell/plaesy-clean.ps1 -AI claude
./scripts/powershell/plaesy-clean.ps1 -AI github_copilot
./scripts/powershell/plaesy-clean.ps1 -AI cursor_ai
```

### Backup Control
```bash
# Disable backup creation
./scripts/bash/plaesy-clean.sh --no-backup
./scripts/powershell/plaesy-clean.ps1 -NoBackup

# Force backup creation (default)
./scripts/bash/plaesy-clean.sh --backup
./scripts/powershell/plaesy-clean.ps1 -Backup
```

### Dry Run Mode
```bash
# Preview what would be removed without actually removing
./scripts/bash/plaesy-clean.sh --dry-run
./scripts/powershell/plaesy-clean.ps1 -DryRun
```

## Enhanced Features (v2.0)

### Multi-Platform Detection
The script can now detect and clean multiple platforms in a single run:

```bash
# Auto-detect all platforms present in project
./scripts/bash/plaesy-clean.sh

# Output shows all detected platforms:
# ✅ Detected multiple AI platforms: claude_code, github_copilot
# 🧹 Using multi-platform cleanup...
```

### Advanced Platform Detection
```bash
# Bash: Advanced JSON parsing for platform detection
awk '
BEGIN { in_platforms=0; brace_depth=0; in_detection=0; platform=""; found=0 }
/"platforms"[[:space:]]*:[[:space:]]*\{/ { in_platforms=1; brace_depth=1; next }
# ... complex JSON parsing logic
' "$config_file"
```

```powershell
# PowerShell: Platform detection with config manager
$detectedPlatforms = @()
$allPlatforms = & $ConfigManager "list-platforms" 2>$null
foreach ($platform in $platforms) {
    $detectionPatterns = & $ConfigManager "get-detection-patterns" $platform 2>$null
    # Check patterns and detect platforms
}
```

### Enhanced Backup System
```bash
# Bash: Comprehensive backup with platform-specific files
create_backup() {
    local backup_dir="$TARGET_DIR/.plaesy-backup-$(date +%Y%m%d-%H%M%S)"

    # Backup important files
    find "$TARGET_DIR" -maxdepth 1 -name "*.json" -type f -exec cp {} "$backup_dir/" \;

    # Backup platform-specific files
    for platform in "${platforms_to_backup[@]}"; do
        for mapping_type in "${mapping_types[@]}"; do
            target_path=$("$CONFIG_MANAGER" "get-mapping-value" "$platform" "$mapping_type" 2>/dev/null)
            if [[ -e "$TARGET_DIR/$target_path" ]]; then
                cp -r "$TARGET_DIR/$target_path" "$backup_dir/"
            fi
        done
    done
}
```

```powershell
# PowerShell: Enhanced backup with platform detection
function New-Backup {
    $backupDir = Join-Path $script:TARGET_DIRECTORY ".plaesy-backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"

    # Backup JSON and YAML files
    Get-ChildItem $script:TARGET_DIRECTORY -Filter "*.json" -File | ForEach-Object {
        Copy-Item -Path $_.FullName -Destination $backupDir -Force
    }

    # Backup platform-specific files
    foreach ($platform in $platformsToBackup) {
        foreach ($mappingType in $mappingTypes) {
            $targetPath = & $ConfigManager "get-mapping-value" $platform $mappingType 2>$null
            $fullPath = Join-Path $script:TARGET_DIRECTORY $targetPath
            if (Test-Path $fullPath) {
                Copy-Item -Path $fullPath -Destination $backupDir -Recurse -Force
            }
        }
    }
}
```

### Comprehensive Cleanup Planning
```bash
# Bash: Detailed cleanup plan display
show_cleanup_plan() {
    echo "Cleanup Plan (Level: $CLEANUP_LEVEL, Platforms: $platform_display)"
    echo "═══════════════════════════════════════════════════════════════"

    # Show cleanup level details
    case "$CLEANUP_LEVEL" in
        "safe") echo -e "${YELLOW}Safe cleanup:${NC} Framework files only, preserve user code" ;;
        "thorough") echo -e "${YELLOW}Thorough cleanup:${NC} Framework + specs, preserve user code" ;;
        "complete") echo -e "${RED}Complete cleanup:${NC} Everything Plaesy-related (DANGEROUS)" ;;
    esac

    # Show platform-specific files based on mapping
    for detected_platform in "${detected_platforms[@]}"; do
        echo -e "\n${YELLOW}Platform-Specific Files ($detected_platform):${NC}"
        for mapping_type in "${mapping_types[@]}"; do
            target_path=$("$CONFIG_MANAGER" "get-mapping-value" "$detected_platform" "$mapping_type" 2>/dev/null)
            if [[ -n "$target_path" && "$target_path" != "null" ]]; then
                if [[ -e "$target_path" ]]; then
                    echo "  - $target_path"
                fi
            fi
        done
    done
}
```

## Implementation Details

### Bash Script (plaesy-clean.sh)
- **Lines**: 795 lines
- **Error Handling**: `set -uo pipefail` with trap cleanup
- **Dependencies**: Requires `common.sh` and `config-manager.sh`
- **Features**: Advanced platform detection, comprehensive backup, JSON parsing

### PowerShell Script (plaesy-clean.ps1)
- **Lines**: 928 lines
- **Error Handling**: `$ErrorActionPreference = "Stop"` with trap cleanup
- **Dependencies**: Self-contained, requires `config-manager.ps1`
- **Features**: Enhanced validation, cross-platform compatibility, verbose logging

### Function Equivalents

| Bash Function | PowerShell Function | Purpose |
|---------------|-------------------|---------|
| `parse_arguments()` | `Parse-Arguments()` | Parse command line arguments |
| `detect_all_platforms()` | `Find-AllPlatforms()` | Detect all platforms in project |
| `show_cleanup_plan()` | `Show-CleanupPlan()` | Display cleanup plan |
| `confirm_removal()` | `Confirm-Removal()` | Confirm with user before cleanup |
| `create_backup()` | `New-Backup()` | Create backup of files |
| `remove_plaesy_directories()` | `Remove-PlaesyDirectories()` | Perform cleanup |

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
  - .claude/roles/ (directory)

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

██████╗ ██╗      █████╗ ███████╗███████╗██╗   ██╗
██╔══██╗██║     ██╔══██╗██╔════╝██╔════╝╝██╗ ██╔╝
██████╔╝██║     ███████║█████╗  ███████╗ ╚████╔╝
██╔═══╝ ██║     ██╔══██║██╔══╝  ╚════██║  ╚██╔╝
██║     ███████╗██║  ██║███████╗███████║   ██║
╚═╝     ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝

Plaesy Clean
   Remove plaesy framework directories

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
  ✅ Removed platform claude_code chatmodes: .claude/roles
  ✅ Removed empty parent directory: .claude
[SUCCESS] Successfully removed 6 items

Plaesy clean completed!

Platform-specific cleanup guidance:
For Claude Code: Use /clean command in future

Cleaned directory: /path/to/project

To reinitialize the project:
   plaesy init [--ai=your-choice]
```

## Platform-Specific Cleanup Guidance

The script provides tailored guidance for each AI platform:

| Platform | Recommended Future Command |
|---------|---------------------------|
| **Claude Code** | Use `/clean` command in future |
| **Cursor AI** | Use `--clean` command in future |
| **Windsurf AI** | Use `clean` command in future |
| **GitHub Copilot** | Continue using `plaesy clean` command |
| **Other platforms** | Use `plaesy clean` command |

## Safety Features

### Pre-Cleanup Validation
1. **Configuration Validation**: Validates `platform.json` before cleanup
2. **Environment Validation**: Checks directory permissions and accessibility
3. **Platform Detection**: Auto-detects platforms to prevent accidental deletion
4. **Dry Run Mode**: Preview changes before execution

### Backup System
1. **Automatic Backup**: Creates timestamped backup before deletion
2. **Platform-Specific Backup**: Backs up platform-specific files based on mapping
3. **Configuration Backup**: Preserves JSON/YAML configuration files
4. **Backup Location**: `.plaesy-backup-YYYYMMDD-HHMMSS` in target directory

### Error Handling
```bash
# Bash: Comprehensive error handling with trap
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}❌ Clean operation failed. Check the error above.${NC}" >&2
    fi
}
trap cleanup EXIT
```

```powershell
# PowerShell: Error handling with trap
function Invoke-Cleanup {
    param([int]$ExitCode = $LASTEXITCODE)
    if ($ExitCode -ne 0) {
        Write-Host "Clean operation failed. Check the error above." -ForegroundColor Red
    }
}
trap {
    Invoke-Cleanup -ExitCode $_.Exception.HResult
}
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   # Check directory permissions
   ls -la /path/to/project

   # Ensure write permissions
   chmod u+w /path/to/project
   ```

2. **Configuration Not Found**
   ```bash
   # Check platform.json exists
   ls -la scripts/configs/platform.json

   # Validate configuration
   ./scripts/bash/config-manager.sh validate
   ```

3. **Config Manager Missing**
   ```bash
   # Ensure config-manager exists
   ls -la scripts/bash/config-manager.sh
   ls -la scripts/powershell/config-manager.ps1
   ```

4. **PowerShell Execution Policy**
   ```powershell
   # Set execution policy
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

### Debug Mode

```bash
# Bash debug mode
bash -x ./scripts/bash/plaesy-clean.sh --dry-run

# PowerShell debug mode
powershell -NoProfile -Command "& { .\scripts\powershell\plaesy-clean.ps1 -DryRun -Verbose }"
```

### Manual Platform Detection

```bash
# Manually detect platforms
./scripts/bash/config-manager.sh list-platforms

# Check detection patterns for a platform
./scripts/bash/config-manager.sh get-detection-patterns claude_code
```

## Integration Examples

### CI/CD Integration
```bash
#!/bin/bash
# Automated cleanup in CI/CD
./scripts/bash/plaesy-clean.sh --yes --level safe --no-backup
```

```powershell
# PowerShell CI/CD cleanup
.\scripts\powershell\plaesy-clean.ps1 -Yes -Level safe -NoBackup
```

### Pre-commit Hook
```bash
#!/bin/sh
# .git/hooks/pre-commit
# Clean before committing
./scripts/bash/plaesy-clean.sh --yes --dry-run
```

### Docker Integration
```dockerfile
# Cleanup in Docker container
COPY scripts/bash/ /opt/plaesy/scripts/
RUN /opt/plaesy/scripts/plaesy-clean.sh --yes --level safe
```

## Version Information

- **Current Version**: 1.0.0
- **Bash Script**: 795 lines, production-ready
- **PowerShell Script**: 928 lines, cross-platform compatible
- **Configuration Format**: platform.json v1.0
- **Supported Platforms**: All platforms defined in platform.json

## Migration Guide

### From v1.0 to v2.0
1. **Enhanced Platform Detection**: Better automatic platform detection
2. **Improved Backup System**: More comprehensive backup creation
3. **Better Validation**: Stricter configuration validation
4. **Enhanced Output**: More detailed cleanup planning and reporting
5. **Cross-Platform Compatibility**: Improved Windows PowerShell support

### Breaking Changes
- **None** - All existing command-line options are preserved
- **Enhanced** - New features added without breaking compatibility
- **Improved** - Better error handling and validation

### Manual Migration
If you need to update an existing cleanup script:

```bash
# Re-run with dry-run to see changes
./scripts/bash/plaesy-clean.sh --dry-run --verbose

# Update to new version if available
./scripts/bash/plaesy-clean.sh --version
```

## Security Considerations

- **Path Validation**: All paths are validated before use
- **Permission Checks**: Validates read/write permissions before operations
- **Backup Safety**: Creates backup before any deletion
- **Confirmation Required**: User confirmation required unless explicitly bypassed
- **Platform Mapping**: Only removes files specified in platform.json mapping
- **Error Handling**: Comprehensive error handling prevents data loss