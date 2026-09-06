# PowerShell Auto-Detection Feature

## Overview

PowerShell Plaesy Spec-Kit now supports automatic AI platform detection, just like the bash version. This enhancement allows users to simply run `plaesy init .` and have the framework automatically detect their AI platform.

## What's New

### Auto-Detection Support

The PowerShell version now includes:

1. **Platform Detection Script** (`platform-detector.ps1`)
   - Automatically detects AI platforms based on project files
   - Supports all major AI platforms: Claude Code, Cursor AI, GitHub Copilot, Windsurf AI, etc.
   - Environment variable detection for enhanced accuracy

2. **Enhanced plaesy-init.ps1**
   - Built-in auto-detection logic
   - Interactive confirmation when platform is detected
   - Fallback to manual selection if no platform detected
   - Non-interactive mode support

3. **Updated Batch CLI (plaesy.bat)**
   - Smart argument parsing
   - Default auto-detection behavior like bash
   - Backward compatibility with explicit AI selection

## Usage Examples

### Auto-Detection (Recommended)

```powershell
# Initialize in current directory with auto-detection
plaesy init .

# Initialize in specific directory with auto-detection
plaesy init my-project

# Interactive mode - shows platform selection menu
plaesy init
```

### Explicit AI Selection (Still Supported)

```powershell
# PowerShell format
plaesy init . -AI claude_code
plaesy init . -AI cursor_ai
plaesy init . -AI github_copilot

# Bash format (also supported)
plaesy init . --ai claude_code
```

## Detection Logic

The platform detector uses **platform.json configuration** for dynamic detection patterns:

1. **Configuration-Driven Detection**
   - Reads detection patterns from `scripts/configs/platform.json`
   - Supports all platforms defined in the configuration
   - Handles wildcards and complex file patterns
   - Maintains priority order from JSON configuration

2. **Dynamic Pattern Matching**
   - Each platform defines its own detection patterns in `platform.json`
   - Supports file paths, directory patterns, and wildcards
   - Example patterns from configuration:
     - **Claude Code**: `.claude/settings.local.json`, `CLAUDE.md`, `.claude/commands`
     - **Cursor AI**: `.cursorrules`, `.cursor/rules`, `.cursor`
     - **GitHub Copilot**: `.github/copilot-instructions.md`, `.github/prompts`
     - **Windsurf AI**: `.windsurf/workflow.yaml`, `.windsurfrules`

3. **Consistent with Bash Version**
   - Uses the same `platform.json` file as bash version
   - Maintains identical detection logic and priority
   - Centralized configuration management

## Behavior Changes

### Before (Old PowerShell)

```powershell
# ❌ This would fail
plaesy init .
# ERROR: Invalid AI platform

# ❌ This would fail
plaesy init . -DebugMode
# ERROR: Invalid AI platform

# ✅ Only worked with explicit AI selection
plaesy init . -AI claude_code
```

### After (New PowerShell)

```powershell
# ✅ Now works with auto-detection
plaesy init .
# [SUCCESS] Detected AI platform: claude_code
# 🚀 Using platform-adapted initialization...

# ✅ Simple auto-detection
plaesy init my-project
# 🤖 No AI platform detected. Starting interactive platform selection...

# ✅ Still supports explicit selection
plaesy init . -AI claude_code
```

## Installation

To get the auto-detection feature:

1. **New Installation:**
   ```powershell
   iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 | iex
   ```

2. **Upgrade Existing:**
   ```powershell
   plaesy upgrade
   ```

## Files Modified/Created

- ✅ `scripts/powershell/platform-detector.ps1` (NEW)
- ✅ `scripts/powershell/plaesy-init.ps1` (ENHANCED)
- ✅ `scripts/powershell/install.ps1` (UPDATED)
- ✅ Batch CLI auto-detection logic (UPDATED)
- ✅ `scripts/powershell/test-detection.ps1` (NEW - Testing script)

## Backward Compatibility

All existing commands continue to work exactly as before. The auto-detection is an enhancement, not a breaking change.

## Technical Details

### Auto-Detection Flow

1. User runs `plaesy init .`
2. Batch file calls `plaesy-init.ps1` directly
3. `plaesy-init.ps1` calls `Find-AIPlatform` function
4. Platform detector scans for AI platform indicators
5. If detected: shows platform and asks for confirmation
6. If not detected: shows interactive platform selection menu
7. Initialization continues with selected/detected platform

### Error Handling

- Graceful fallback when detection fails
- Interactive mode always available as backup
- Clear error messages for invalid platforms
- Timeout handling for non-interactive environments

## Testing

To test the auto-detection:

```powershell
# Test auto-detection
cd your-project
plaesy init .

# Test with specific platforms
echo "# Claude Code test" > CLAUDE.md
plaesy init .

# Test interactive mode
plaesy init

# Test explicit selection
plaesy init . -AI claude_code
```

## Troubleshooting

**Auto-detection not working?**
- Run `plaesy repair` to fix missing scripts
- Check that `platform-detector.ps1` exists in `.plaesy/scripts/powershell/`
- Ensure PowerShell execution policy allows script execution

**Wrong platform detected?**
- Use explicit AI selection: `plaesy init . -AI correct_platform`
- Remove conflicting platform files
- Check environment variables

**Still getting "Invalid AI platform" error?**
- Ensure you're using the updated version: `plaesy version`
- Run `plaesy upgrade` to get latest features
- Check that all scripts are present: `plaesy status`

---

**Result:** PowerShell Plaesy Spec-Kit now has the same auto-detection convenience as bash! 🎉