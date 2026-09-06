# Plaesy Scripts Documentation

**Complete guide to Plaesy Spec-Kit automation scripts.**

**Framework Version: 0.0.1**

## Quick Navigation

### **Script Documentation**

| Script | Purpose | AI Priority | Documentation |
|--------|---------|-------------|---------------|
| **[plaesy-init.md](./plaesy-init.md)** | Project initialization & AI platform setup | **HIGH** | Platform detection and setup |
| **[config-manager.md](./config-manager.md)** | AI platform configuration management | **MEDIUM** | Platform configs management |
| **[platform-detector.md](./platform-detector.md)** | PowerShell auto-detection feature | **MEDIUM** | Windows platform detection |
| **[plaesy-clean.md](./plaesy-clean.md)** | Project cleanup and reset | **MEDIUM** | Clean project state |
| **[plaesy-analyze.md](./plaesy-analyze.md)** | Project structure analysis | **CRITICAL** | AI-optimized project analysis |
| **[update-agent-context.md](./update-agent-context.md)** | AI context synchronization | **CRITICAL** | Multi-platform AI context |
| **[get-feature-paths.md](./get-feature-paths.md)** | Feature path resolution | **CRITICAL** | Current feature context |
| **[check-task-prerequisites.md](./check-task-prerequisites.md)** | Development validation | **CRITICAL** | Prerequisites checking |
| **[create-new-feature.md](./create-new-feature.md)** | Feature branch creation | **HIGH** | Feature workflow setup |
| **[inject-ai-headers.md](./inject-ai-headers.md)** | AI header optimization | **MEDIUM** | Platform-specific headers |
| **[install.md](./install.md)** | System installation | **HIGH** | Framework installation |
| **[common.md](./common.md)** | Common functions library | **CRITICAL** | Shared utilities and constants |
| **[plaesy-graph.md](./plaesy-graph.md)** | Repo knowledge-graph builder (nodes/edges, queries, impact check) | **MEDIUM** | Codebase relationship analysis |
| **[plaesy-trim.md](./plaesy-trim.md)** | Token/context compression (command output + memory files) | **MEDIUM** | Reduce token spend per session |

### **Quick Start for AI Assistants**

**MANDATORY SEQUENCE** - Always run this when encountering a Plaesy project:
```bash
# 1. Understand project structure (CRITICAL)
#    Also triggers instruction auto-load to .plaesy/memory/
./scripts/bash/plaesy-analyze.sh

# 2. Get current feature context
./scripts/bash/get-feature-paths.sh

# 3. Validate development setup
./scripts/bash/check-task-prerequisites.sh

# 4. Update AI context understanding
./scripts/bash/update-agent-context.sh
```

**📌 Note**: `plaesy-analyze.sh` detects project technologies and auto-copies relevant instructions from `instructions/mapping.json` to `.plaesy/memory/`. This creates a self-contained project memory.

**Windows PowerShell Alternative:**
```powershell
# 1. Understand project structure (CRITICAL)
.\scripts\powershell\plaesy-analyze.ps1

# 2. Get current feature context
.\scripts\powershell\get-feature-paths.ps1

# 3. Validate development setup
.\scripts\powershell\check-task-prerequisites.ps1

# 4. Update AI context understanding
.\scripts\powershell\update-agent-context.ps1
```

## AI Assistant Essential Information

### **Critical Files for AI Understanding**

1. **`.plaesy/memory/analysis/project.json`** - AI-optimized project summary
2. **`.plaesy/memory/analysis/project.structure.json`** - Complete project structure
3. **`.plaesy/memory/analysis/project.graph.json`** - Dependency graph (nodes + edges; built by `plaesy-analyze` automatically, queryable via `plaesy-graph`)
4. **`specs/[feature-name]/plan.md`** - Current feature implementation plan
5. **Platform-specific AI contexts** (`CLAUDE.md`, `.github/copilot-instructions.md`, etc.)

### **Platform Detection Logic**

Plaesy automatically detects AI platforms and applies appropriate configurations:

- **Claude Code** → `CLAUDE.md`, `.claude/` directory
- **GitHub Copilot** → `.github/copilot-instructions.md`
- **Cursor AI** → `.cursor/rules/specify-rules.mdc`
- **Windsurf AI** → Configuration files
- **Continue.dev** → `.continue/` directory
- **And 10+ other platforms** with optimized configurations

### **Enhanced PowerShell Support (v2.0)**

All scripts now have full PowerShell equivalents with enhanced features:

| Feature | Bash | PowerShell |
|---------|------|------------|
| **Platform Detection** | Automatic | Automatic |
| **Configuration Management** | config-manager.sh | config-manager.ps1 |
| **Error Handling** | `set -euo pipefail` | `$ErrorActionPreference = "Stop"` |
| **Parameter Binding** | Manual parsing | PowerShell parameters |
| **Cross-Platform** | Linux/macOS | Windows/Linux/macOS |
| **Output Formatting** | Color + Emojis | Color Only (Professional) |

## PowerShell Script Development Guidelines

### **IMPORTANT: No Emojis in PowerShell Scripts**

**CRITICAL**: PowerShell scripts have encoding issues with emojis that cause runtime errors. When creating or modifying PowerShell scripts:

- **NEVER** use emojis (🎉📁🔍✅🚀❌⏰📚🏛️) in PowerShell scripts
- **ALWAYS** use text-based indicators instead: `[INFO]`, `[SUCCESS]`, `[WARNING]`, `[ERROR]`
- **USE** color formatting with `-ForegroundColor` for visual distinction
- **REASON**: Emojis cause encoding and execution errors in Windows PowerShell environments

### **PowerShell vs Bash Script Differences**

- **Bash scripts** can safely use emojis (UTF-8 supported)
- **PowerShell scripts** must avoid emojis due to Windows encoding limitations
- **Both** should use consistent color schemes and text formatting
- **Both** follow platform.json mapping for configuration

## Enhanced Script Features (v2.0)

### **Configuration-Driven Architecture**

All scripts now use centralized configuration through `platform.json`:

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
      "chatmodes": "chatmodes/*"
    }
  },
  "platforms": {
    "claude_code": {
      "name": "Claude Code",
      "mapping": {
        "core": "CLAUDE.md",
        "instructions": ".claude/instructions",
        "prompts": ".claude/commands",
        "chatmodes": ".claude/roles"
      },
      "detection_patterns": [
        ".claude/",
        "CLAUDE.md"
      ]
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

### **Advanced Error Handling**

#### Bash Error Handling
```bash
# Strict error handling
set -euo pipefail

# Trap-based cleanup
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo "[ERROR] Operation failed with exit code $exit_code" >&2
    fi
}
trap cleanup EXIT
```

#### PowerShell Error Handling
```powershell
# Strict error handling
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Trap-based cleanup
function Invoke-Cleanup {
    param([int]$ExitCode = $LASTEXITCODE)
    if ($ExitCode -ne 0) {
        Write-Host "[ERROR] Operation failed with exit code $ExitCode" -ForegroundColor Red
    }
}
trap {
    Invoke-Cleanup -ExitCode $_.Exception.HResult
}
```

## Common Workflows

### **New Feature Development**

#### Bash Workflow
```bash
# 1. Create feature
./scripts/bash/create-new-feature.sh "Feature description"

# 2. Run critical sequence
./scripts/bash/plaesy-analyze.sh
./scripts/bash/get-feature-paths.sh
./scripts/bash/check-task-prerequisites.sh
./scripts/bash/update-agent-context.sh
```

#### PowerShell Workflow
```powershell
# 1. Create feature
.\scripts\powershell\create-new-feature.ps1 "Feature description"

# 2. Run critical sequence
.\scripts\powershell\plaesy-analyze.ps1
.\scripts\powershell\get-feature-paths.ps1
.\scripts\powershell\check-task-prerequisites.ps1
.\scripts\powershell\update-agent-context.ps1
```

### **AI Platform Setup**

#### Bash Setup
```bash
# Initialize with specific AI platform
./scripts/bash/plaesy-init.sh . --ai claude_code

# Optimize AI headers
./scripts/bash/inject-ai-headers.sh --ai claude --target . --backup
```

#### PowerShell Setup
```powershell
# Initialize with specific AI platform
.\scripts\powershell\plaesy-init.ps1 -Target . -AI claude_code

# Optimize AI headers
.\scripts\powershell\inject-ai-headers.ps1 -AI claude -Target . -Backup
```

### **Project Cleanup**

#### Bash Cleanup
```bash
# Safe cleanup with preview
./scripts/bash/plaesy-clean.sh --dry-run

# Complete cleanup for specific platform
./scripts/bash/plaesy-clean.sh --level thorough --ai claude_code --yes
```

#### PowerShell Cleanup
```powershell
# Safe cleanup with preview
.\scripts\powershell\plaesy-clean.ps1 -DryRun

# Complete cleanup for specific platform
.\scripts\powershell\plaesy-clean.ps1 -Level thorough -AI claude_code -Yes
```

### **Project Installation**

#### Bash Installation
```bash
# Linux/macOS installation
curl -fsSL https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/install.sh | bash

# Local installation
./scripts/bash/install.sh
```

#### PowerShell Installation
```powershell
# Windows installation
powershell -NoProfile -ExecutionPolicy Bypass -Command "& { iwr https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/install.ps1 | iex }"

# Local installation
.\scripts\powershell\install.ps1
```

## Script Architecture

### **Configuration Management**

All scripts use centralized configuration management:

```bash
# Bash: Use config-manager.sh
CONFIG_MANAGER="$(dirname "$0")/config-manager.sh"
platforms=$("$CONFIG_MANAGER" "list-platforms")
platform_name=$("$CONFIG_MANAGER" "get-platform-config" "$platform" "name")
```

```powershell
# PowerShell: Use config-manager.ps1
$ConfigManager = Join-Path $ScriptDir "config-manager.ps1"
$platforms = & $ConfigManager "list-platforms"
$platformName = & $ConfigManager "get-platform-config" $platform "name"
```

### **Platform Detection**

Advanced platform detection using multiple methods:

1. **File-based detection** - Check for platform-specific files
2. **Directory detection** - Look for platform directories
3. **Configuration detection** - Parse configuration files
4. **Environment detection** - Check environment variables
5. **User preference** - Respect user-specified platforms

### **Dynamic File Mapping**

Platform-specific file mapping based on configuration:

```bash
# Bash: Dynamic file copying
for mapping_type in "${mapping_types[@]}"; do
    target_path=$("$CONFIG_MANAGER" "get-mapping-value" "$platform" "$mapping_type")
    if [[ -n "$target_path" && "$target_path" != "null" ]]; then
        # Copy files with appropriate extensions
    fi
done
```

```powershell
# PowerShell: Dynamic file copying
foreach ($mappingType in $mappingTypes) {
    $targetPath = & $ConfigManager "get-mapping-value" $platform $mappingType
    if ($targetPath -and $targetPath -ne "null") {
        # Copy files with appropriate extensions
    }
}
```

## Integration Examples

### **CI/CD Integration**

#### GitHub Actions (Bash)
```yaml
- name: Setup Plaesy
  run: |
    ./scripts/bash/plaesy-init.sh --ai claude_code
    ./scripts/bash/update-agent-context.sh

- name: Analyze Project
  run: ./scripts/bash/plaesy-analyze.sh

- name: Cleanup
  run: ./scripts/bash/plaesy-clean.sh --yes --level safe
```

#### GitHub Actions (PowerShell)
```yaml
- name: Setup Plaesy
  shell: pwsh
  run: |
    .\scripts\powershell\plaesy-init.ps1 -AI claude_code
    .\scripts\powershell\update-agent-context.ps1

- name: Analyze Project
  shell: pwsh
  run: .\scripts\powershell\plaesy-analyze.ps1

- name: Cleanup
  shell: pwsh
  run: .\scripts\powershell\plaesy-clean.ps1 -Yes -Level safe
```

### **Docker Integration**

#### Dockerfile (Multi-platform)
```dockerfile
# Install Plaesy Spec-Kit
COPY scripts/ /opt/plaesy/scripts/
RUN chmod +x /opt/plaesy/scripts/bash/*.sh

# Initialize with Claude Code
RUN /opt/plaesy/scripts/bash/plaesy-init.sh --ai claude_code --target /app

# Set working directory
WORKDIR /app

# Run analysis
CMD ["/opt/plaesy/scripts/bash/plaesy-analyze.sh"]
```

#### Docker Compose
```yaml
version: '3.8'
services:
  app:
    build: .
    volumes:
      - .:/app
      - plaesy-cache:/root/.cache/claude
    environment:
      - AI_PLATFORM=claude_code
    command: /opt/plaesy/scripts/bash/update-agent-context.sh
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
            "command": "./scripts/bash/plaesy-analyze.sh",
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
            "command": "./scripts/bash/update-agent-context.sh",
            "group": "build"
        }
    ]
}
```

#### PowerShell (Windows)
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Plaesy: Analyze Project",
            "type": "shell",
            "command": ".\\scripts\\powershell\\plaesy-analyze.ps1",
            "group": "build"
        },
        {
            "label": "Plaesy: Update Context",
            "type": "shell",
            "command": ".\\scripts\\powershell\\update-agent-context.ps1",
            "group": "build"
        }
    ]
}
```

## Troubleshooting

### **Common Issues**

1. **Permission Denied (Bash)**
   ```bash
   # Make scripts executable
   chmod +x scripts/bash/*.sh
   ```

2. **PowerShell Execution Policy**
   ```powershell
   # Set execution policy
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

3. **Configuration Not Found**
   ```bash
   # Check platform.json exists
   ls -la scripts/configs/platform.json

   # Validate configuration
   ./scripts/bash/config-manager.sh validate
   ```

4. **Config Manager Missing**
   ```bash
   # Ensure config-manager exists
   ls -la scripts/bash/config-manager.sh
   ls -la scripts/powershell/config-manager.ps1
   ```

### **Debug Mode**

#### Bash Debug
```bash
# Enable debug mode
bash -x ./scripts/bash/plaesy-init.sh

# Verbose output
./scripts/bash/plaesy-analyze.sh --verbose
```

#### PowerShell Debug
```powershell
# Enable debug mode
powershell -NoProfile -Command "& { .\scripts\powershell\plaesy-init.ps1 -DebugMode }"

# Verbose output
.\scripts\powershell\plaesy-analyze.ps1 -Verbose
```

## Version Information

- **Framework Version**: 0.0.1
- **Configuration Format**: platform.json v1.0
- **Bash Scripts**: Production-ready scripts
- **PowerShell Scripts**: Cross-platform compatible scripts
- **Supported Platforms**: Multiple AI platforms
- **Configuration**: Centralized, mapping-based
- **Error Handling**: Comprehensive with trap support
- **Documentation**: Complete with examples and troubleshooting

## Security Considerations

- **Path Validation**: All paths validated before use
- **Permission Checks**: Read/write permissions verified
- **Input Sanitization**: User inputs validated and sanitized
- **Backup Safety**: Automatic backups before destructive operations
- **Error Handling**: Comprehensive error handling prevents data loss
- **Configuration Security**: Platform configurations validated
- **Cross-Platform Safety**: Scripts work safely across platforms

## Migration Guide

### **From v1.0 to v2.0**

1. **Enhanced PowerShell Support**: Full PowerShell equivalents for all scripts
2. **Configuration-Driven Architecture**: All scripts use platform.json
3. **Improved Platform Detection**: Better automatic platform detection
4. **Enhanced Error Handling**: Trap-based error handling
5. **Cross-Platform Compatibility**: Better Windows support
6. **Professional Output**: Removed emojis from PowerShell scripts
7. **Comprehensive Documentation**: Updated documentation for all scripts

### **Breaking Changes**

- **None** - All existing command-line options preserved
- **Enhanced** - New features added without breaking compatibility
- **Improved** - Better error handling and validation

### **Manual Migration**

If upgrading an existing project:

```bash
# Re-initialize to update configuration
./scripts/bash/plaesy-init.sh . --ai your_current_platform

# Update context
./scripts/bash/update-agent-context.sh

# Verify setup
./scripts/bash/plaesy-analyze.sh
```

```powershell
# Re-initialize to update configuration
.\scripts\powershell\plaesy-init.ps1 -Target . -AI your_current_platform

# Update context
.\scripts\powershell\update-agent-context.ps1

# Verify setup
.\scripts\powershell\plaesy-analyze.ps1
```