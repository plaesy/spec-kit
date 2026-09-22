# plaesy-init.sh / plaesy-init.ps1

**Project initialization and AI platform configuration script.**

**Priority:** HIGH - Required for Plaesy project setup

## Purpose

Initializes Plaesy projects with AI platform detection, configuration setup, and template creation. This script creates the foundation for AI-assisted development workflows.

## Usage

### Interactive Mode
```bash
# Interactive AI platform selection
plaesy init

# Create new project with interactive AI selection
plaesy init my-awesome-app

# Initialize in current directory
plaesy init .
```

### Direct Platform Specification
```bash
# Initialize with specific AI platform
plaesy init . --ai claude_code      # Claude Code
plaesy init . --ai cursor_ai        # Cursor AI
plaesy init . --ai github_copilot   # GitHub Copilot
plaesy init . --ai windsurf_ai      # Windsurf AI

# Create project in specific directory
plaesy init my-app --ai claude_code
plaesy init ./my-project --ai cursor_ai
```

### Windows PowerShell
```powershell
# Interactive mode
plaesy init

# Create new project
plaesy init .\my-awesome-app

# Initialize with specific AI platform
plaesy init . -AI claude_code      # Claude Code
plaesy init . -AI cursor_ai        # Cursor AI
plaesy init . -AI github_copilot   # GitHub Copilot
plaesy init .\my-project -AI windsurf_ai  # Windsurf AI

# PowerShell parameter examples
plaesy init -Help                  # Show help
plaesy init -Version               # Show version
plaesy init --ai claude_code --target .\project
```

## Key Features

### Enhanced AI Platform Detection
- **Automatic Detection**: Scans for existing AI platform configurations
- **Interactive Selection**: Menu-based platform selection with timeout
- **Multi-Platform Support**: 15+ AI platforms supported
- **Fallback Options**: Generic AI configuration for unknown platforms
- **Non-Interactive Mode**: Auto-detection for CI/CD environments

### Configuration-Driven Setup
- **Dynamic Platform Loading**: Loads platforms from `platform.json`
- **Mapping-Based Configuration**: Uses `config-manager.ps1`/`config-manager.sh`
- **Core AI Config**: Creates `CLAUDE.md` at project root from `agents.instructions.md`
- **Instruction File Copy**: Copies selected instructions to `.plaesy/instructions/` (flat structure) for per-project documentation
- **Chatmode Copy**: Copies all chatmodes to `.plaesy/roles/` (flat structure, platform-agnostic)
- **Prompt Copy**: Copies prompts to platform-specific locations (`.claude/commands/`, `.cursor/rules/`, etc.)
- **Exclude Pattern Support**: Respects exclude patterns in mapping
- **Dynamic Structure Creation**: Creates directories based on configuration

### Supported AI Platforms
| Platform | Detection Files | Features | Integration |
|----------|----------------|----------|-------------|
| **claude_code** | `.claude/` directory | Native tool access | Full automation |
| **github_copilot** | `.github/copilot-instructions.md` | VS Code integration | Code completion |
| **cursor_ai** | `.cursor/rules/` directory | IDE integration | Refactoring support |
| **windsurf_ai** | Configuration files | Platform-specific | Custom workflows |
| **cline** | `.cline/` directory | Tool integration | Automated workflows |
| **deepseek** | Configuration detection | API integration | Enhanced responses |
| **kilo_code** | Platform files | Code assistance | Development support |
| **qoder** | Configuration | AI optimization | Quality improvement |
| **trae_ai** | Platform detection | Advanced features | Specialized workflows |
| **continue_dev** | `.continue/` directory | VS Code extension | AI assistance |
| **tabnine** | Configuration files | AI completion | IDE integration |
| **codeium** | Platform detection | IDE tools | AI-powered assistance |
| **codewhisperer** | Configuration | AWS-powered | Code generation |
| **studio_bot** | Platform files | Android development | Google integration |
| **replit_ghostwriter** | `.replit/` directory | Cloud IDE | AI assistance |
| **llama_index** | Configuration files | AI framework | Application development |
| **ollama** | Platform detection | Local AI | Offline capabilities |
| **lm_studio** | Configuration files | Local AI | Model management |
| **generic_ai** | Fallback | Universal compatibility | Basic automation |

## Template-Driven File Creation

When `plaesy init` runs, bootstrap files are created from templates in `templates/` (or fallback to empty file):

### Context File Creation
- **Source**: `templates/context.template.md` (when available)
- **Destination**: `.plaesy/context.md` (project root)
- **Purpose**: Session-specific state tracking (max 100 lines, archive to memory as needed)
- **Fallback**: If template not found, creates empty file (user fills in manually)

### Memory Index File Creation
- **Source**: `templates/memory.template.md` (when available)
- **Destination**: `.plaesy/memory.md` (project root)
- **Purpose**: Index and quick links to knowledge files in `.plaesy/memory/` and instruction copies in `.plaesy/instructions/`
- **Fallback**: If template not found, creates empty file (user fills in manually)

This replaces the previous broken inline generation (context.md used literal `\n` escape sequences, memory.md was minimal).

### Loop State File Creation
- **Source**: `templates/state.template.json` (when available)
- **Destination**: `.plaesy/state.json` (project root)
- **Purpose**: Autonomous loop configuration and quality gates
- **Timestamp Handling**: Placeholder `[TIMESTAMP]` replaced with current UTC time
- **Fallback**: If template not found, creates empty file (user fills in manually)

### Task Management Documentation
When creating task structure:
- Creates task status directories (backlog, todo, doing, done, blocked) under `.plaesy/tasks/`
  — this is the single source of truth for tasks (one file per task, YAML frontmatter,
  folder = status). `/start`, `/continue`, and `/loop` all read/write here; there is no
  separate `backlog.md` file — a flat `backlog.md` was previously bootstrapped from a
  template that never existed (`templates/backlog.template.md`) and was never read by
  any prompt, so the dead code path was removed.
- Creates `.plaesy/tasks/README.md` with quick reference
- Refers users to `.plaesy/instructions/tasks.md` for detailed task management instructions
- Task management instructions auto-loaded from `tasks.instructions.md` during init

---

## Instruction Files Auto-Copy Feature

When `plaesy init` runs, instruction files are **automatically copied to `.plaesy/instructions/`** (flat structure) for per-project documentation:

### How It Works
1. **Source**: Reads from `instructions/*.instructions.md` (framework reference)
2. **Root Config**: Creates `CLAUDE.md` at project root from `agents.instructions.md` (session startup protocol)
3. **Detection**: Analyzes project's technology stack (package.json, Cargo.toml, etc.)
4. **Selective Copy**: Only copies instructions matching detected technologies + always-load core
5. **Destination**: Copies to `.plaesy/instructions/` (flat, no subfolders) with renamed extension
   - Always-load (from `instructions/mapping.json`'s `always_load` array): `plaesy.instructions.md`,
     `plaesy-trim.instructions.md`, `plaesy-graph.instructions.md`, `tasks.instructions.md`,
     `quality-gates.instructions.md`, `error-recovery.instructions.md`, `date-system.instructions.md`
     → each copied to `.plaesy/instructions/<name>.md`
   - Detected: `nextjs.instructions.md` → `.plaesy/instructions/nextjs.md`
   - Detected: `reactjs.instructions.md` → `.plaesy/instructions/reactjs.md`
   - Not detected: `ruby-on-rails.instructions.md` → NOT copied (unless `--all-instructions`)

### File Organization (Example for Next.js + React Project)
```
PROJECT ROOT
├── CLAUDE.md                 # Root AI config (from agents.instructions.md)
├── context.md                      # Session state (from template, can be edited)
└── .plaesy/instructions/
    ├── plaesy.md             # Always loaded (universal rules)
    ├── quality-gates.md      # Always loaded
    ├── error-recovery.md     # Always loaded
    ├── tasks.md              # Always loaded
    ├── date-system.md        # Always loaded
    ├── nextjs.md             # Copied (detected Next.js)
    ├── reactjs.md            # Copied (detected React)
    └── sql.md                # Copied (if DB detected)
```

### Selective vs Full Install
- **Selective (Default)**: Only copies instructions matching project's detected technologies
  ```bash
  plaesy init . --ai claude_code    # Only matched instructions
  ```
- **Full**: Use `--all-instructions` flag to copy ALL available instructions (23 total)
  ```bash
  plaesy init . --ai claude_code --all-instructions  # All 23 instructions
  ```

### Why This Feature?
- ✅ **Self-Contained**: Project documentation stays with project (CLAUDE.md compliant)
- ✅ **Single Source**: One copy per project, not duplicated across platforms
- ✅ **Lightweight (Default)**: Only relevant instructions copied based on tech stack (selective install)
  - React project: ~5 relevant files instead of all 23
  - Spring Boot project: Different 5 relevant files
  - Optimized for each project's actual needs
- ✅ **Flexible**: Can copy all 23 with `--all-instructions` if needed
- ✅ **AI-Friendly**: Stored in standard `.plaesy/instructions/` location for AI context loading

---

## Platform.json Mapping Logic

**Script follows platform.json mapping for instructions, prompts, and chatmodes**

### Source vs Destination Mapping
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
        // NOTE: No "core" mapping - uses universal .plaesy/instructions/plaesy.md instead
        // NOTE: No "instructions" mapping - all go to .plaesy/instructions/ flat structure
        "prompts": ".claude/commands",                  // DESTINATION: platform-specific (actually used)
        "chatmodes": ".claude/roles"                    // DEFINED but NOT read by plaesy-init.sh — see note below
      }
    }
  }
}
```

> **Known discrepancy (verified against `scripts/bash/plaesy-init.sh` 2026-09-16)**: the
> per-platform `chatmodes` mapping above is defined in `platform.json` but
> `setup_platform_config()`'s copy loop only iterates `"prompts"` (plus optional
> `"core"`) — it never reads the `chatmodes` key for any platform. Chatmodes are
> actually copied by a separate, platform-agnostic function straight to
> `.plaesy/roles/` (suffix `.chatmode.md` stripped to `.md`), regardless of which
> AI platform was selected. Don't trust the `chatmodes` destination in
> `platform.json` as evidence of real behavior for any platform — confirm against
> the actual copy function in `plaesy-init.sh`/`plaesy-init.ps1` instead.

### File Copying Strategy
- **Universal files** (always-load, per `instructions/mapping.json`): Always copied to
  `.plaesy/instructions/` regardless of platform
  - `plaesy.instructions.md` → `.plaesy/instructions/plaesy.md`
  - `agents.instructions.md` → `.plaesy/instructions/agents.md`
  - `quality-gates.instructions.md`, `error-recovery.instructions.md`, `tasks.instructions.md`,
    `date-system.instructions.md`, `plaesy-trim.instructions.md`, `plaesy-graph.instructions.md`
- **Instruction files**: Copied to `.plaesy/instructions/` based on detected tech stack
- **Prompt/Workflow files**: Copied to platform-specific locations (`.claude/commands/`, etc)
- **Chat modes**: Always copied to `.plaesy/roles/`, the same path on every platform
  (see discrepancy note above — this is not what `platform.json` implies)

### Configuration Manager Integration

**The script uses `config-manager.ps1` (PowerShell) or `config-manager.sh` (Bash) for:**

1. **Platform Detection**: `config-manager detect-platform`
2. **Platform Loading**: `config-manager list-platforms`
3. **Platform Names**: `config-manager get-platform-config {platform} name`
4. **Mapping Values**: `config-manager get-mapping-value {platform} {type}`
5. **Exclude Patterns**: `config-manager get-mapping-excludes {platform} {type}`
6. **Structure Config**: `config-manager get-plaesy-structure {field}`

## Implementation Details

### Bash Script (plaesy-init.sh)
- **Status**: Production-ready with comprehensive features
- **Error Handling**: `set -euo pipefail`
- **Dependencies**: Requires `common.sh`
- **Features**: Platform detection, timeout handling, selective instruction copying, comprehensive validation

### PowerShell Script (plaesy-init.ps1)
- **Status**: Production-ready with cross-platform compatibility
- **Error Handling**: `$ErrorActionPreference = "Stop"`
- **Dependencies**: Self-contained, requires `config-manager.ps1`
- **Features**: Parameter binding, selective instruction copying, enhanced validation, cross-platform compatibility

### Function Equivalents

| Bash Function | PowerShell Function | Purpose |
|---------------|-------------------|---------|
| `load_platforms()` | `Load-Platforms()` | Load available platforms |
| `get_platform_name()` | `Get-PlatformName()` | Get platform display name |
| `normalize_platform()` | `Normalize-Platform()` | Normalize platform input |
| `validate_platform()` | `Test-PlatformExists()` | Validate platform exists |
| `get_ai_choice()` | `Get-AIChoice()` | Get AI platform choice |
| `setup_platform_config()` | `Set-AISpecificConfig()` | Setup platform configuration |
| `create_structure()` | `New-ProjectStructure()` | Create project structure |

## Enhanced Features (v2.0)

### Dynamic Configuration Loading
```bash
# Bash loads platforms dynamically
platforms=$(load_platforms)

# PowerShell loads platforms dynamically
$platforms = Load-Platforms
```

### Advanced Platform Detection
```bash
# Bash: Auto-detect with user confirmation
detected=$(detect_platform)
read -p "Use detected platform? (Y/n): " use_detected

# PowerShell: Auto-detect with user confirmation
$detected = Find-AIPlatform
$useDetected = Read-Host "Use detected platform? (Y/n)"
```

### Instruction File Copy Logic
```bash
# Bash: Copy instructions to .plaesy/instructions/ with rename
find "$source_dir" -maxdepth 1 -name "*.instructions.md" -type f | while read -r file; do
    filename="${file##*/}"
    
    # Skip plaesy.instructions.md (handled by core mapping)
    if [[ "$filename" != "plaesy.instructions.md" ]]; then
        # Check exclude patterns and selective install
        knowledge_filename="${filename%.instructions.md}.md"
        cp "$file" ".plaesy/memory/$knowledge_filename"
    fi
done

# PowerShell: Similar logic with rename
Get-ChildItem $sourceDir -Filter "*.instructions.md" -File | ForEach-Object {
    $filename = $_.Name
    
    # Skip plaesy.instructions.md (handled by core mapping)
    if ($filename -ne "plaesy.instructions.md") {
        $knowledgeFilename = $filename -replace '\.instructions\.md$', '.md'
        Copy-Item -Path $_.FullName -Destination ".plaesy\memory\knowledge\$knowledgeFilename"
    }
}
```

### Prompts & Chatmodes (Platform-Specific)
Prompts and chatmodes are still copied to platform-specific locations:
```bash
# Prompts: .claude/commands/, .cursor/rules/, etc.
# Chatmodes: .claude/roles/, .cursor/chatmodes/, etc.
```

## Error Handling and Validation

### Bash Script Validation
```bash
# Strict error handling
set -euo pipefail

# Directory validation
validate_directory_exists "$target_dir" "target directory"

# Platform validation
validate_platform "$ai_choice"

# Writable test
test_file=$(mktemp)
echo "test" > "$test_file"
rm "$test_file"
```

### PowerShell Script Validation
```powershell
# Strict error handling
$ErrorActionPreference = "Stop"

# Directory validation
if (-not (Test-Path $TargetDir)) {
    Write-Error-Custom "Target directory '$TargetDir' does not exist."
}

# Platform validation
if (-not (Test-PlatformExists -Platform $aiChoice)) {
    Write-Error-Custom "Invalid AI platform: $aiChoice"
}

# Writable test
$testFile = Join-Path $TargetDir "test-write-$(Get-Random).tmp"
"test" | Out-File -FilePath $testFile -ErrorAction Stop
Remove-Item $testFile -ErrorAction SilentlyContinue
```

## Output Examples

### Successful Initialization
```
██████╗ ██╗      █████╗ ███████╗███████╗██╗   ██╗
██╔══██╗██║     ██╔══██╗██╔════╝██╔════╝╝██╗ ██╔╝
██████╔╝██║     ███████║█████╗  ███████╗ ╚████╔╝
██╔═══╝ ██║     ██╔══██║██╔══╝  ╚════██║  ╚██╔╝
██║     ███████╗██║  ██║███████╗███████║   ██║
╚═╝     ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝

Plaesy Spec-Kit Initialization
   Constitutional Development Framework

Target directory: /path/to/project
Detecting AI platform...
Detected: Claude Code
[SUCCESS] Selected: Claude Code
[STEP] Creating Plaesy structure...
[INFO] Base dir: '.plaesy'
[INFO] Core dirs: 'memory'
[INFO] Project dirs: 'docs, specs'
[SUCCESS] Plaesy structure created with dynamic configuration
[STEP] Setting up configuration for: Claude Code based on platform.json mapping...
[SUCCESS] Created directory: .claude
[SUCCESS] Created core file: CLAUDE.md (from instructions/plaesy.instructions.md)
[SUCCESS] Copied instruction to knowledge: .plaesy/instructions/angular.md
[SUCCESS] Copied instruction to knowledge: .plaesy/instructions/nextjs.md
[SUCCESS] Copied instruction to knowledge: .plaesy/instructions/reactjs.md
[SUCCESS] Created directory: .claude/commands
[SUCCESS] Copied prompt: .claude/commands/start.md
[SUCCESS] Copied prompt: .claude/commands/implement.md
[INFO] Copying chatmodes to .plaesy/roles/...
[SUCCESS] software-developer.md
[SUCCESS] AI-specific configuration completed based on platform.json mapping

Plaesy Spec-Kit initialization completed!

Project: /path/to/project
Platform: Claude Code

Next Steps:
1. Start your project with /start command
2. Use /continue for automated workflow
3. Use /implement for implementation phase

Constitutional Development Framework Active
   Quality through discipline. Excellence through automation.
```

## Troubleshooting

### Common Issues

1. **Platform not found**
   ```bash
   # Check available platforms
   ./scripts/bash/plaesy-init.sh --help

   # List platforms from config
   ./scripts/bash/config-manager.sh list-platforms
   ```

2. **Config manager missing**
   ```bash
   # Ensure config-manager exists
   ls -la scripts/bash/config-manager.sh
   ls -la scripts/powershell/config-manager.ps1
   ```

3. **Permission denied**
   ```bash
   # Make script executable
   chmod +x scripts/bash/plaesy-init.sh
   ```

4. **PowerShell execution policy**
   ```powershell
   # Set execution policy
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

### Debug Mode

```bash
# Bash debug mode
bash -x ./scripts/bash/plaesy-init.sh

# PowerShell debug mode
powershell -NoProfile -Command "& { .\scripts\powershell\plaesy-init.ps1 -DebugMode }"
```

## Integration Examples

### CI/CD Integration
```bash
#!/bin/bash
# Non-interactive initialization for CI/CD
plaesy init . --ai claude_code --target ./build
```

```powershell
# PowerShell for CI/CD
.\scripts\powershell\plaesy-init.ps1 -AI claude_code -Target .\build
```

### Docker Integration
```dockerfile
# Dockerfile
COPY scripts/bash/ /opt/plaesy/scripts/
RUN chmod +x /opt/plaesy/scripts/plaesy-init.sh
RUN /opt/plaesy/scripts/plaesy-init.sh --ai claude_code --target /app
```

## CLAUDE.md / AGENTS.md Creation

After `plaesy init`, projects are created with self-contained AI configuration:

### How It Works

**Option 1: Auto-Create (Coming Soon)**
- Script automatically creates `.plaesy/instructions/agents.md` from `agents.instructions.md`
- Project can optionally create `CLAUDE.md` or `AGENTS.md` at root if needed

**Option 2: Manual (Current)**
- Copy `docs/CLAUDE.md.template` → `CLAUDE.md` in project root
- Template shows how to load `.plaesy/instructions/plaesy.md` at session start
- Template includes session initialization protocol

### File Structure After Init

```
project/
├── CLAUDE.md                    # Optional: AI config (copy from template)
│   └── Reads: .plaesy/instructions/plaesy.md (once per session)
│
├── .plaesy/
│   ├── memory/
│   │   ├── plaesy.md           # ← AI reads THIS at session start
│   │   ├── agents.md           # Session init protocol
│   │   ├── context.md          # Current session state
│   │   ├── quality-gates.md    # Always loaded
│   │   ├── error-recovery.md   # Always loaded
│   │   ├── tasks.md            # Always loaded
│   │   ├── date-system.md      # Always loaded
│   │   └── [tech]*.md          # Tech-specific instructions (detected)
│   │
│   ├── scripts/
│   │   ├── bash/
│   │   └── powershell/
│   │
│   └── tasks/
│       ├── backlog/
│       ├── todo/
│       ├── doing/
│       ├── done/
│       └── blocked/
│
├── .claude/
│   ├── commands/               # Workflows copied here (thin references)
│   └── roles/                  # Chat modes
│
└── ... (platform-specific dirs based on detected AI)
```

### Session Initialization Flow

```
AI Session Starts
    ↓
Load CLAUDE.md (or AGENTS.md)
    ↓
Read .plaesy/instructions/plaesy.md (~150 tokens, cached)
    ↓
Check .plaesy/context.md (if resuming)
    ↓
✓ Context loaded, ready to work
```

**Token Impact**:
- Before init: No context available
- After init: Full project context in 150-200 tokens (~57% vs typical full load)

## Version Information

- **Current Version**: 0.0.1
- **Configuration Format**: platform.json v1.0
- **Supported Platforms**: 18+ AI platforms
- **Status**: Production-ready for all platforms
- **Key Feature**: Auto-copy all instructions to `.plaesy/instructions/` (flat structure)

## Security Considerations

- **File Permissions**: Scripts respect existing file permissions
- **Path Validation**: All paths are validated before use
- **Input Sanitization**: User inputs are validated and sanitized
- **Backup Safety**: Scripts create backups before making changes
- **Error Handling**: Comprehensive error handling prevents data loss

## Migration Guide

### From v1.0 to v2.0
1. **No breaking changes** - existing projects continue to work
2. **Enhanced detection** - better platform detection
3. **Improved validation** - stricter input validation
4. **Better error messages** - more helpful error reporting
5. **Cross-platform compatibility** - improved Windows support

### Manual Migration
If you need to manually update an existing project:

```bash
# Re-run initialization to update configuration
plaesy init . --ai your_current_platform
```

This will update any missing configuration files while preserving your existing setup.