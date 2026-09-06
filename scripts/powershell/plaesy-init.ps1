# Plaesy Spec-Kit Initialization Script (PowerShell)
# 100% configuration-driven platform support - Enhanced to match Bash version

[CmdletBinding()]
param(
  [Parameter()]
  [string]$AI = $null,

  [Parameter()]
  [string]$Target = $null,

  [Parameter()]
  [switch]$Help = $false,

  [Parameter()]
  [switch]$Version = $false,

  [Parameter()]
  [switch]$DebugMode = $false,

  [Parameter()]
  [switch]$AllInstructions = $false,

  [Parameter(ValueFromRemainingArguments = $true)]
  [string[]]$RemainingArguments = @()
)

# Enhanced error handling - match Bash strict mode
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Source common functions - match Bash approach
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$CommonFile = Join-Path $ScriptDir "common.ps1"
$ConfigFile = Join-Path $ScriptDir "..\configs\platform.json"
$ConfigManager = Join-Path $ScriptDir "config-manager.ps1"

# Source common functions if available
if (Test-Path $CommonFile) {
    try {
        . $CommonFile
    }
    catch {
        Write-Host "Warning: Could not load common.ps1, using fallback functions" -ForegroundColor Yellow
    }
}

# Fallback logging functions if common.ps1 is not loaded
if (-not (Get-Command "Log-Info" -ErrorAction SilentlyContinue)) {
    function Log-Info {
        param([string]$Message)
        Write-Host "[INFO] $Message" -ForegroundColor Blue
    }
}

if (-not (Get-Command "Log-Error" -ErrorAction SilentlyContinue)) {
    function Log-Error {
        param([string]$Message)
        Write-Host "[ERROR] $Message" -ForegroundColor Red
    }
}

if (-not (Get-Command "Log-Warning" -ErrorAction SilentlyContinue)) {
    function Log-Warning {
        param([string]$Message)
        Write-Host "[WARNING] $Message" -ForegroundColor Yellow
    }
}

if (-not (Get-Command "Write-Success" -ErrorAction SilentlyContinue)) {
    function Write-Success {
        param([string]$Message)
        Write-Host "[SUCCESS] $Message" -ForegroundColor Green
    }
}

if (-not (Get-Command "Write-Banner" -ErrorAction SilentlyContinue)) {
    function Write-Banner {
        param([string]$Title, [string]$Subtitle = "")
        Write-Host ""
        Write-Host "##### PLAESY SPEC-KIT #####" -ForegroundColor Magenta
        Write-Host "$Title" -ForegroundColor Cyan
        if ($Subtitle) {
            Write-Host "$Subtitle" -ForegroundColor Gray
        }
        Write-Host ""
    }
}

# Load available platforms dynamically - match Bash load_platforms()
function Load-Platforms {
    # Try config manager first
    if (Test-Path $ConfigManager) {
        try {
            $platforms = & $ConfigManager "list-platforms" 2>$null
            if ($platforms) {
                return $platforms -split "`n" | Where-Object { $_ -and $_.Trim() }
            }
        }
        catch {
            # Fall through to direct JSON reading
        }
    }

    # Fallback: Read directly from platform.json
    if (Test-Path $ConfigFile) {
        try {
            $content = Get-Content $ConfigFile -Raw | ConvertFrom-Json
            if ($content.platforms) {
                return $content.platforms.PSObject.Properties.Name
            }
        }
        catch {
            Log-Warning "Could not parse platform configuration: $_"
        }
    }

    # Final fallback: hardcoded platforms
    return @(
        "claude_code", "cursor_ai", "github_copilot", "cline", "deepseek",
        "kilo_code", "qoder", "trae_ai", "windsurf_ai", "continue_dev",
        "tabnine", "codeium", "codewhisperer", "studio_bot", "replit_ghostwriter",
        "llama_index", "ollama", "lm_studio", "generic_ai"
    )
}

# Get platform display name dynamically - match Bash get_platform_name()
function Get-PlatformName {
    param([string]$Platform)

    # Try config manager first
    if (Test-Path $ConfigManager) {
        try {
            $name = & $ConfigManager "get-platform-config" $Platform "name" 2>$null
            if ($name -and $name -ne "null") {
                return $name
            }
        }
        catch {
            # Fall through to direct JSON reading
        }
    }

    # Fallback: Read directly from platform.json
    if (Test-Path $ConfigFile) {
        try {
            $content = Get-Content $ConfigFile -Raw | ConvertFrom-Json
            if ($content.platforms -and $content.platforms.PSObject.Properties.Name -contains $Platform) {
                return $content.platforms.$Platform.name
            }
        }
        catch {
            # Fall through to hardcoded names
        }
    }

    # Final fallback: hardcoded display names
    $displayNames = @{
        "claude_code" = "Claude Code"
        "cursor_ai" = "Cursor AI"
        "github_copilot" = "GitHub Copilot"
        "cline" = "Cline"
        "deepseek" = "DeepSeek"
        "kilo_code" = "Kilo Code"
        "qoder" = "Qoder AI"
        "trae_ai" = "Trae AI"
        "windsurf_ai" = "Windsurf AI"
        "continue_dev" = "Continue.dev"
        "tabnine" = "Tabnine"
        "codeium" = "Codeium"
        "codewhisperer" = "Amazon CodeWhisperer"
        "studio_bot" = "Google Studio Bot"
        "replit_ghostwriter" = "Replit Ghostwriter"
        "llama_index" = "LlamaIndex"
        "ollama" = "Ollama"
        "lm_studio" = "LM Studio"
        "generic_ai" = "Generic AI"
    }

    if ($displayNames.ContainsKey($Platform)) {
        return $displayNames[$Platform]
    }

    return $Platform
}

# Normalize platform input - enhanced to match Bash normalize_platform()
function Normalize-Platform {
    # NOTE: do not name this parameter "Input" -- it collides with PowerShell's
    # automatic $input pipeline-enumerator variable, so -Input never actually
    # binds a value and this function silently returns "" for every call.
    param([string]$PlatformInput)

    $platforms = Load-Platforms
    $inputLower = $PlatformInput.ToLower()

    # Check exact match first
    foreach ($platform in $platforms) {
        if ($PlatformInput -eq $platform) {
            return $platform
        }

        $displayName = Get-PlatformName -Platform $platform
        if ($PlatformInput -eq $displayName) {
            return $platform
        }
    }

    # Check partial/common name matches - match Bash patterns
    switch ($inputLower) {
        { $_ -in @("claude", "claude_code", "anthropic", "claude-code") } { return "claude_code" }
        { $_ -in @("cursor", "cursor_ai", "cursor-ai") } { return "cursor_ai" }
        { $_ -in @("github", "copilot", "github_copilot", "github-copilot", "gh-copilot") } { return "github_copilot" }
        { $_ -in @("cline") } { return "cline" }
        { $_ -in @("deepseek", "deepseek_ai", "deepseek-ai") } { return "deepseek" }
        { $_ -in @("kilo", "kilo_code", "kilo-code") } { return "kilo_code" }
        { $_ -in @("qoder", "qoder_ai", "qoder-ai") } { return "qoder" }
        { $_ -in @("trae", "trae_ai", "trae-ai") } { return "trae_ai" }
        { $_ -in @("windsurf", "windsurf_ai", "windsurf-ai") } { return "windsurf_ai" }
        { $_ -in @("continue", "continue_dev", "continue-dev", "continue.dev") } { return "continue_dev" }
        { $_ -in @("tabnine") } { return "tabnine" }
        { $_ -in @("codeium") } { return "codeium" }
        { $_ -in @("codewhisperer", "code_whisperer", "amazon_codewhisperer", "whisperer") } { return "codewhisperer" }
        { $_ -in @("studio_bot", "studio-bot", "google_studio_bot", "studiobot") } { return "studio_bot" }
        { $_ -in @("replit", "ghostwriter", "replit_ghostwriter", "replit-ghostwriter") } { return "replit_ghostwriter" }
        { $_ -in @("llamaindex", "llama_index", "llama-index", "code-llama") } { return "llama_index" }
        { $_ -in @("ollama") } { return "ollama" }
        { $_ -in @("lmstudio", "lm_studio", "lm-studio") } { return "lm_studio" }
        { $_ -in @("generic", "generic_ai", "generic-ai", "none", "manual") } { return "generic_ai" }
        default { return $PlatformInput }
    }
}

# Validate platform exists - match Bash validate_platform()
function Test-PlatformExists {
    param([string]$Platform)

    $platforms = Load-Platforms
    return $platforms -contains $Platform
}

# Get AI choice from user - enhanced to match Bash get_ai_choice()
function Get-AIChoice {
    $platforms = Load-Platforms

    # Check if platforms list is empty
    if ($platforms.Count -eq 0) {
        Log-Error "No platforms available to choose from."
        return "none"
    }

    Write-Host "Available AI Assistants:" -ForegroundColor Cyan
    Write-Host ""

    Write-Host "0. Cancel initialization"
    for ($i = 0; $i -lt $platforms.Count; $i++) {
        $platform = $platforms[$i]
        $displayName = Get-PlatformName -Platform $platform
        Write-Host "$($i + 1). $displayName"
    }

    Write-Host "$($platforms.Count + 1). No AI (Manual development)"
    Write-Host ""

    $maxChoice = $platforms.Count + 1

    # Check if we're in an interactive environment. $Host.UI.RawUI alone is
    # unreliable (often non-null even when input is redirected/no console is
    # attached); also check Console.IsInputRedirected as a real signal that
    # Read-Host has nothing to read.
    $noConsoleInput = $false
    try { $noConsoleInput = [Console]::IsInputRedirected } catch { $noConsoleInput = $false }
    if ((-not $Host.UI.RawUI) -or $noConsoleInput) {
        # Not interactive, use auto-detection or default
        Write-Host "Non-interactive mode detected, auto-selecting Claude Code..." -ForegroundColor Yellow
        # Try to find Claude Code first, then first platform
        if ($platforms -contains "claude_code") {
            return "claude_code"
        } elseif ($platforms.Count -gt 0) {
            return $platforms[0]
        } else {
            return "none"
        }
    }

    # Enhanced input handling with timeout - match Bash approach
    $timeout = 30
    $choice = ""
    $attempts = 0
    $maxAttempts = 3

    do {
        # Counted before Read-Host, not after: if Read-Host itself throws (e.g.
        # no console attached), the increment must still happen or this loop
        # never terminates -- $attempts would stay 0 forever and both the
        # catch block's check and the outer `while` condition stay true.
        $attempts++
        try {
            Write-Host "Choose your AI assistant (0-$maxChoice): " -NoNewline -ForegroundColor Yellow
            $choice = Read-Host

            if ([string]::IsNullOrWhiteSpace($choice)) {
                if ($attempts -lt $maxAttempts) {
                    Write-Host "No input received. Please enter a number between 0 and $maxChoice." -ForegroundColor Yellow
                    continue
                } else {
                    Write-Host "No valid input after $maxAttempts attempts. Auto-selecting Claude Code..." -ForegroundColor Yellow
                    if ($platforms -contains "claude_code") {
                        return "claude_code"
                    } elseif ($platforms.Count -gt 0) {
                        return $platforms[0]
                    } else {
                        return "none"
                    }
                }
            }

            if ($choice -match '^\d+$') {
                $choiceNum = [int]$choice
                if ($choiceNum -lt 0 -or $choiceNum -gt $maxChoice) {
                    Write-Host "Please enter a number between 0 and $maxChoice." -ForegroundColor Red
                    continue
                }

                if ($choiceNum -eq 0) {
                    return "cancel"
                }

                if ($choiceNum -eq $maxChoice) {
                    return "none"
                }

                return $platforms[$choiceNum - 1]
            } else {
                Write-Host "Please enter a valid number." -ForegroundColor Red
            }
        }
        catch {
            # Timeout or error, use default
            if ($attempts -lt $maxAttempts) {
                Write-Host "Input error. Please try again." -ForegroundColor Yellow
                continue
            } else {
                Write-Host "Input failed after $maxAttempts attempts. Auto-selecting Claude Code..." -ForegroundColor Yellow
                if ($platforms -contains "claude_code") {
                    return "claude_code"
                } elseif ($platforms.Count -gt 0) {
                    return $platforms[0]
                } else {
                    return "none"
                }
            }
        }
    } while ($attempts -lt $maxAttempts)
}

# Setup platform configuration dynamically - enhanced to match Bash setup_platform_config()
function Set-AISpecificConfig {
    param([string]$Platform)

    Log-Info "Setting up configuration for: $(Get-PlatformName -Platform $Platform) based on platform.json mapping..."

    # Get expanded root path (same as create_structure function)
    $expandedRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
    Write-Host "[DEBUG] expandedRoot=$expandedRoot" -ForegroundColor Gray

    # Check for PLEASY_ROOT override
    if ($env:PLEASY_ROOT) {
        $expandedRoot = $env:PLEASY_ROOT
        Write-Host "[DEBUG] Using PLEASY_ROOT override: $expandedRoot" -ForegroundColor Gray
    }

    # Read platform configuration directly from JSON
    try {
        Write-Host "[DEBUG] Reading config from: $ConfigFile" -ForegroundColor Gray
        $configContent = Get-Content $ConfigFile -Raw | ConvertFrom-Json
        Write-Host "[DEBUG] Config loaded successfully" -ForegroundColor Gray

        $platformConfig = $configContent.platforms.$Platform
        Write-Host "[DEBUG] platformConfig retrieved for platform: $Platform" -ForegroundColor Gray

        $plaesyConfig = $configContent.plaesy
        Write-Host "[DEBUG] plaesyConfig retrieved" -ForegroundColor Gray

        # Note: Core file (agents.instructions.md → CLAUDE.md/AGENTS.md) is handled by platform-specific loop below
        Write-Host "[DEBUG] About to process platform mappings" -ForegroundColor Gray

        # Determine which instruction files are relevant to this project (selective
        # install, like a skills marketplace) unless the user asked for everything.
        $selectedInstructions = $null
        $detectScript = Join-Path $ScriptDir "detect-stack.ps1"
        if (-not $AllInstructions -and (Test-Path $detectScript)) {
            try {
                $selectedInstructions = & $detectScript -TargetDir "." -PlaesyRoot $expandedRoot
                if ($selectedInstructions.Count -gt 0) {
                    Log-Success "Selective install: $($selectedInstructions.Count) instruction file(s) matched to this project (use -AllInstructions to install all)"
                }
            }
            catch {
                Write-Warning "Selective instruction detection failed, installing all: $_"
                $AllInstructions = $true
            }
        }

        Write-Host "[DEBUG] platformConfig exists: $($platformConfig -ne $null), has mapping: $($platformConfig.mapping -ne $null)" -ForegroundColor Gray

        if ($platformConfig -and $platformConfig.mapping) {
            Write-Host "[DEBUG] Processing platform mappings..." -ForegroundColor Gray

            # Copy chatmodes to .plaesy/roles/ (platform-agnostic, not platform-specific)
            # This is universal like instructions → .plaesy/memory/
            Log-Info "Copying chatmodes to .plaesy/roles/..."
            $sourceMapping = $plaesyConfig.mapping.chatmodes.value

            if ($sourceMapping) {
                $sourceMappingClean = $sourceMapping -replace '/\*$'
                $sourceDir = Join-Path $expandedRoot $sourceMappingClean
                $rolesDir = Join-Path ".plaesy" "roles"

                if (-not (Test-Path $rolesDir)) {
                    New-Item -ItemType Directory -Path $rolesDir -Force | Out-Null
                }

                if (Test-Path $sourceDir) {
                    Get-ChildItem -Path $sourceDir -Filter "*.chatmode.md" -File | ForEach-Object {
                        $filename = $_.BaseName
                        $destFile = Join-Path $rolesDir "$filename.md"

                        if (Test-Path $destFile) {
                            Write-Warning "  ⊘ $filename.md (already exists, skipping)"
                        } else {
                            Copy-Item -Path $_.FullName -Destination $destFile
                            Log-Info "  ✓ $filename.md"
                        }
                    }
                } else {
                    Write-Warning "Source chatmodes directory not found: $sourceDir"
                }
            }

            # Setup platform-specific files based on platform.json mapping
            # Note: instructions are already copied to .plaesy/memory/ by Copy-Instructions (universal, not platform-specific)
            # Note: chatmodes are already copied to .plaesy/roles/ above (universal, not platform-specific)
            $platformMappings = @("prompts")

            # Also add core mapping if platform has one (for platform-specific core files like Cursor, Copilot)
            if ($platformConfig.mapping.core) {
                Write-Host "[DEBUG] Platform has core mapping: $($platformConfig.mapping.core)" -ForegroundColor Gray
                $platformMappings = @("core") + $platformMappings
            }

            Write-Host "[DEBUG] platformMappings: $($platformMappings -join ', ')" -ForegroundColor Gray

            foreach ($mappingType in $platformMappings) {
                $targetPath = $platformConfig.mapping.$mappingType
                Write-Host "[DEBUG] mappingType=$mappingType, targetPath=$targetPath" -ForegroundColor Gray

                if ($targetPath) {
                    # Create target directory if needed
                    $targetDir = Split-Path -Path $targetPath -Parent
                    Write-Host "[DEBUG] targetDir=$targetDir" -ForegroundColor Gray
                    if ($targetDir -ne "." -and $targetDir) {
                        if (-not (Test-Path $targetDir)) {
                            New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
                            Write-Success "Created directory: $targetDir"
                        }
                    }

                    # Copy source files based on mapping type
                    switch ($mappingType) {
                        "core" {
                            # Copy platform-specific core file (Cursor, Copilot, etc.)
                            $sourceMapping = $plaesyConfig.mapping.core.value

                            if ($sourceMapping) {
                                $sourceFile = Join-Path $expandedRoot $sourceMapping
                                if (Test-Path $sourceFile) {
                                    if (Test-Path $targetPath) {
                                        Write-Info "Platform-specific core exists, skipping: $targetPath"
                                    } else {
                                        Copy-Item -Path $sourceFile -Destination $targetPath
                                        Write-Success "Created platform-specific core file: $targetPath"
                                    }
                                } else {
                                    Write-Warning "Source core file not found: $sourceFile"
                                }
                            }
                        }

                        "prompts" {
                            # Copy prompt files: source from plaesy.prompts, destination from platform.prompts
                            $sourceMapping = $plaesyConfig.mapping.prompts.value

                            if ($sourceMapping) {
                                # Remove wildcard /* from mapping for directory path
                                $sourceMappingClean = $sourceMapping -replace '/\*$'
                                $sourceDir = Join-Path $expandedRoot $sourceMappingClean

                                # Ensure targetPath exists as a directory
                                if (-not (Test-Path $targetPath)) {
                                    New-Item -ItemType Directory -Path $targetPath -Force | Out-Null
                                    Write-Success "Created directory: $targetPath"
                                }

                                if (Test-Path $sourceDir) {
                                    $resolvedSource = (Resolve-Path -Path $sourceDir -ErrorAction SilentlyContinue).Path.TrimEnd('\', '/')
                                    if (-not $resolvedSource) {
                                        Write-Warning "Could not resolve source directory: $sourceDir"
                                        continue
                                    }
                                    Get-ChildItem -Path $resolvedSource -Filter "*.md" -File -Recurse | ForEach-Object {
                                        $relativePath = $_.FullName.Substring($resolvedSource.Length).TrimStart('\', '/')
                                        $relativeBase = $relativePath -replace '\.md$'
                                        $extension = switch ($Platform) {
                                            "claude_code" { ".md" }
                                            "cursor_ai" { ".md" }
                                            "github_copilot" { ".prompt.md" }
                                            default { ".md" }
                                        }

                                        $destFile = Join-Path $targetPath ($relativeBase + $extension)
                                        $destParent = Split-Path $destFile -Parent
                                        if (-not (Test-Path $destParent)) {
                                            New-Item -ItemType Directory -Path $destParent -Force | Out-Null
                                        }
                                        Copy-Item -Path $_.FullName -Destination $destFile -Force
                                        Write-Success "Copied prompt: $destFile"
                                    }
                                } else {
                                    Write-Warning "Source prompts directory not found: $sourceDir"
                                }
                            }
                        }

                    }
                }
            }
        }
    }
    catch {
        Write-Host "[ERROR] Could not setup platform configuration: $_" -ForegroundColor Red
        Write-Host "[ERROR] Exception Details:`n$($_.Exception | Format-List * -Force | Out-String)" -ForegroundColor Red
    }

    Write-Success "AI-specific configuration completed based on platform.json mapping"
}

# Create basic project structure - enhanced to match Bash create_structure()
function New-TaskSystem {
    param([string]$BaseDir)

    Log-Info "Setting up task management system..."

    # Create task status directories
    $taskStatuses = @("backlog", "todo", "doing", "done", "blocked")
    $tasksDir = Join-Path $BaseDir "tasks"

    foreach ($status in $taskStatuses) {
        $statusDir = Join-Path $tasksDir $status
        if (-not (Test-Path $statusDir)) {
            New-Item -ItemType Directory -Path $statusDir -Force | Out-Null
            Write-Success "Created task directory: $statusDir"
        }
    }

    # Create README for task structure
    $readmePath = Join-Path $tasksDir "README.md"
    if (-not (Test-Path $readmePath)) {
        $readmeContent = @"
# Task Management

Tasks are organized by status:

- **backlog/** — Ideas, features, bugs (unscheduled)
- **todo/** — Ready to start (next in queue)
- **doing/** — In active work
- **done/** — Completed and validated
- **blocked/** — Waiting on dependency

## Task Lifecycle

\`\`\`
backlog → todo → doing → [quality-gates] → done / blocked
\`\`\`

Move tasks between directories as status changes. Use \`/evolve\` autonomous loop for automated workflow.

**See \`.plaesy/memory/tasks.md\` for detailed task management instructions.**
"@
        Set-Content -Path $readmePath -Value $readmeContent -Encoding UTF8
        Log-Info "  ✓ tasks/README.md"
    }
}

function New-ProjectStructure {
    param([string]$AIPlatform)

    Log-Info "Creating Plaesy structure..."

    # Get structure configuration from platform.json
    $baseDir = ".plaesy"
    $coreDirs = @("memory")
    $projectDirs = @("docs", "specs")

    # Read structure directly from JSON
    if (Test-Path $ConfigFile) {
        try {
            # Read JSON directly to avoid config manager issues
            $content = Get-Content $ConfigFile -Raw | ConvertFrom-Json
            if ($content.plaesy) {
                $plaesy = $content.plaesy

                if ($plaesy.base_directory) {
                    $baseDir = $plaesy.base_directory
                }

                if ($plaesy.core_directories) {
                    $coreDirs = $plaesy.core_directories
                }

                if ($plaesy.project_directories) {
                    $projectDirs = $plaesy.project_directories
                }
            }
        }
        catch {
            Write-Warning "Could not parse platform configuration for structure, using defaults: $_"
        }
    }

    # Structure configuration loaded successfully

    # Create base directory
    if (-not (Test-Path $baseDir)) {
        New-Item -ItemType Directory -Path $baseDir -Force | Out-Null
    }

    # Create core directories
    foreach ($dir in $coreDirs) {
        if ($dir) {
            $targetPath = Join-Path $baseDir $dir
            Log-Info "Creating core directory: '$targetPath'"
            if (-not (Test-Path $targetPath)) {
                New-Item -ItemType Directory -Path $targetPath -Force | Out-Null
            }
        }
    }

    # Create memory subdirectories
    $memorySubDirs = @()
    try {
        if (Test-Path $ConfigFile) {
            $content = Get-Content $ConfigFile -Raw | ConvertFrom-Json
            if ($content.plaesy.memory_subdirectories) {
                $memorySubDirs = $content.plaesy.memory_subdirectories
            }
        }
    }
    catch {
        # Fallback to default
        $memorySubDirs = @("knowledge")
    }

    foreach ($subdir in $memorySubDirs) {
        if ($subdir) {
            $targetPath = Join-Path $baseDir "memory" $subdir
            Log-Info "Creating memory subdirectory: '$targetPath'"
            if (-not (Test-Path $targetPath)) {
                New-Item -ItemType Directory -Path $targetPath -Force | Out-Null
            }
        }
    }

    # Create project directories
    foreach ($dir in $projectDirs) {
        if ($dir) {
            Log-Info "Creating project directory: '$dir'"
            if (-not (Test-Path $dir)) {
                New-Item -ItemType Directory -Path $dir -Force | Out-Null
            }
        }
    }

    # Create task management system directories
    New-TaskSystem -BaseDir $baseDir

    Write-Success "Plaesy structure created with dynamic configuration"
}

# Bootstrap autonomous loop template files
function Initialize-AutonomousLoop {
    param([string]$ScriptDir)

    $templatesDir = Join-Path $ScriptDir "..\..\templates"
    $memoryDir = ".plaesy\memory"

    # Only copy if files don't already exist
    if (-not (Test-Path (Join-Path $memoryDir "backlog.md")) -and (Test-Path (Join-Path $templatesDir "backlog.template.md"))) {
        Log-Info "Bootstrapping autonomous loop templates..."
        Copy-Item -Path (Join-Path $templatesDir "backlog.template.md") -Destination (Join-Path $memoryDir "backlog.md") -Force
        Write-Success "Created: $memoryDir\backlog.md"
    }

    if (-not (Test-Path (Join-Path $memoryDir "state.json")) -and (Test-Path (Join-Path $templatesDir "loop-state.template.md"))) {
        Copy-Item -Path (Join-Path $templatesDir "loop-state.template.md") -Destination (Join-Path $memoryDir "state.json") -Force
        Write-Success "Created: $memoryDir\state.json"
    }
}

# Show help - enhanced to match Bash
function Show-Help {
    Write-Host "Plaesy Spec-Kit Initialization Script" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "USAGE:" -ForegroundColor Yellow
    Write-Host "    plaesy-init [DIRECTORY] [OPTIONS]" -ForegroundColor White
    Write-Host ""
    Write-Host "ARGUMENTS:" -ForegroundColor Yellow
    Write-Host "    DIRECTORY       Target directory (default: current directory)" -ForegroundColor White
    Write-Host ""
    Write-Host "OPTIONS:" -ForegroundColor Yellow
    Write-Host '    -h, -help      Show this help message' -ForegroundColor White
    Write-Host '    -v, -version   Show version information' -ForegroundColor White
    Write-Host "    --ai [platform] Specify AI platform (will prompt if not provided)" -ForegroundColor White
    Write-Host "    --target [dir]  Specify target directory (alternative to positional argument)" -ForegroundColor White
    Write-Host "    -AllInstructions  Install every instruction file instead of only the ones" -ForegroundColor White
    Write-Host "                      detected as relevant to this project's stack" -ForegroundColor White
    Write-Host ""
    Write-Host "AVAILABLE AI PLATFORMS:" -ForegroundColor Yellow
    Write-Host "    (Dynamically loaded from platform.json)" -ForegroundColor Gray
    Write-Host ""
    Write-Host "EXAMPLES:" -ForegroundColor Yellow
    Write-Host "    plaesy init                                    # Interactive mode in current directory" -ForegroundColor White
    Write-Host "    plaesy init ./my-project                        # Interactive mode in specified directory" -ForegroundColor White
    Write-Host "    plaesy init . --ai claude_code                   # Use Claude Code in current directory" -ForegroundColor White
    Write-Host "    plaesy init ./my-project --ai claude_code       # Use Claude Code in specified directory" -ForegroundColor White
    Write-Host "    plaesy init --ai claude_code --target ./project # Long form options" -ForegroundColor White
    Write-Host "    plaesy init --ai=claude_code --target=./project # Combined format" -ForegroundColor White
}

# Show version - enhanced to match Bash
function Show-Version {
    $versionFile = Join-Path $ScriptDir "..\..\VERSION"
    if (Test-Path $versionFile) {
        $version = (Get-Content $versionFile -Raw).Trim()
        if ($version -match '^\d+\.\d+\.\d+$') {
            Write-Host "Plaesy Spec-Kit v$version"
        } else {
            Write-Host "Plaesy Spec-Kit v0.0.0"
        }
    } else {
        Write-Host "Plaesy Spec-Kit v0.0.0"
    }
}

# Parse arguments - enhanced to match Bash parse_args()
function Parse-Arguments {
    param([string[]]$Args)

    $aiChoice = ""
    $targetDir = "."

    for ($i = 0; $i -lt $Args.Count; $i++) {
        $arg = $Args[$i]

        switch ($arg) {
            { $_ -in @("-h", "--help") } {
                Show-Help
                exit 0
            }
            { $_ -in @("-v", "--version") } {
                Show-Version
                exit 0
            }
            "--ai" {
                if ($i + 1 -lt $Args.Count) {
                    $aiChoice = $Args[$i + 1]
                    $i++
                } else {
                    Log-Error "--ai requires a value"
                    exit 1
                }
            }
            "--target" {
                if ($i + 1 -lt $Args.Count) {
                    $targetDir = $Args[$i + 1]
                    $i++
                } else {
                    Log-Error "--target requires a value"
                    exit 1
                }
            }
            default {
                if ($arg.StartsWith("--ai=")) {
                    $aiChoice = $arg.Substring(6)
                } elseif ($arg.StartsWith("--target=")) {
                    $targetDir = $arg.Substring(10)
                } elseif ($arg.StartsWith("-")) {
                    Log-Error "Unknown option: $arg"
                    Write-Host "Use --help for usage information"
                    exit 1
                } else {
                    # First positional argument is target directory
                    if ($targetDir -eq "." -or [string]::IsNullOrEmpty($targetDir)) {
                        $targetDir = $arg
                    } else {
                        Log-Error "Too many arguments. Use: plaesy-init [directory] [--ai [platform]]"
                        exit 1
                    }
                }
            }
        }
    }

    return @{
        AIChoice = $aiChoice
        TargetDir = $targetDir
    }
}

# Validate target directory - enhanced to match Bash
function Test-TargetDirectory {
    param([string]$TargetDir)

    if (-not (Test-Path $TargetDir)) {
        Log-Error "Target directory '$TargetDir' does not exist."
        return $false
    }

    # Check writable
    try {
        $testFile = Join-Path $TargetDir "test-write-$(Get-Random).tmp"
        "test" | Out-File -FilePath $testFile -ErrorAction Stop
        Remove-Item $testFile -ErrorAction SilentlyContinue
    }
    catch {
        Log-Error "Target directory '$TargetDir' is not writable."
        return $false
    }

    return $true
}

# Detect platform automatically - enhanced to match Bash
function Find-AIPlatform {
    if (Test-Path $ConfigManager) {
        try {
            return & $ConfigManager "detect-platform" 2>$null
        }
        catch {
            return $null
        }
    }
    return $null
}

# ============================================================================
# COPY FUNCTIONS - Copy instructions, scripts, and templates to project
# ============================================================================

function Copy-Instructions {
    [CmdletBinding()]
    param([Parameter(Mandatory=$true)][string]$TargetDir)

    $sourceDir = Join-Path $ScriptDir "..\.." "instructions"
    if (-not (Test-Path $sourceDir)) {
        Write-Warning "Instructions source directory not found"
        return $false
    }

    Log-Info "Copying instructions to .plaesy/memory/..."
    $memoryDir = Join-Path $TargetDir "memory"
    $null = New-Item -ItemType Directory -Force -Path $memoryDir

    $mappingFile = Join-Path $sourceDir "mapping.json"
    $alwaysLoad = $null

    if (Test-Path $mappingFile) {
        $config = Get-Content $mappingFile -Raw | ConvertFrom-Json -ErrorAction SilentlyContinue
        $alwaysLoad = $config.mappings.always_load
    }

    if ($alwaysLoad) {
        foreach ($file in $alwaysLoad) {
            $src = Join-Path $sourceDir $file
            if (Test-Path $src) {
                $name = $file -replace '\.instructions\.md$', '.md'
                $dst = Join-Path $memoryDir $name
                if (-not (Test-Path $dst)) {
                    Copy-Item -Path $src -Destination $dst -Force
                    Log-Info "  ✓ $name"
                } else {
                    Write-Host "  ⊘ $name (already exists)" -ForegroundColor Yellow
                }
            }
        }
    }

    Write-Success "Instructions copied"
}

function Copy-Scripts {
    [CmdletBinding()]
    param([Parameter(Mandatory=$true)][string]$TargetDir)

    $sourceDir = Join-Path $ScriptDir "..\.." "scripts"
    if (-not (Test-Path $sourceDir)) {
        Write-Warning "Scripts source directory not found: $sourceDir"
        return $false
    }

    Log-Info "Copying scripts to .plaesy/scripts/..."
    $null = New-Item -ItemType Directory -Force -Path (Join-Path $TargetDir "scripts" "powershell")
    $null = New-Item -ItemType Directory -Force -Path (Join-Path $TargetDir "scripts" "bash")
    $null = New-Item -ItemType Directory -Force -Path (Join-Path $TargetDir "scripts" "configs")

    $psScriptDir = Join-Path $sourceDir "powershell"
    if (Test-Path $psScriptDir) {
        Get-ChildItem -Path $psScriptDir -Filter "*.ps1" -ErrorAction SilentlyContinue | ForEach-Object {
            $targetPath = Join-Path $TargetDir "scripts" "powershell" $_.Name
            if (-not (Test-Path $targetPath)) {
                Copy-Item -Path $_.FullName -Destination $targetPath -Force
            }
        }
        Log-Info "  ✓ PowerShell scripts copied"
    }

    $bashScriptDir = Join-Path $sourceDir "bash"
    if (Test-Path $bashScriptDir) {
        Get-ChildItem -Path $bashScriptDir -Filter "*.sh" -ErrorAction SilentlyContinue | ForEach-Object {
            $targetPath = Join-Path $TargetDir "scripts" "bash" $_.Name
            if (-not (Test-Path $targetPath)) {
                Copy-Item -Path $_.FullName -Destination $targetPath -Force
            }
        }
        Log-Info "  ✓ Bash scripts copied"
    }

    $configDir = Join-Path $sourceDir "configs"
    if (Test-Path $configDir) {
        Get-ChildItem -Path $configDir -ErrorAction SilentlyContinue | ForEach-Object {
            $targetPath = Join-Path $TargetDir "scripts" "configs" $_.Name
            if (-not (Test-Path $targetPath)) {
                Copy-Item -Path $_.FullName -Destination $targetPath -Force
            }
        }
        Log-Info "  ✓ Configuration files copied"
    }

    Write-Success "Scripts copied"
}

function New-MemoryStructure {
    [CmdletBinding()]
    param([Parameter(Mandatory=$true)][string]$TargetDir)

    Log-Info "Creating memory and analysis directories..."
    $null = New-Item -ItemType Directory -Force -Path (Join-Path $TargetDir "memory")
    $null = New-Item -ItemType Directory -Force -Path (Join-Path $TargetDir "analysis")

    $contextPath = Join-Path $TargetDir "context.md"
    if (-not (Test-Path $contextPath)) {
        # Try to copy from template first
        $templatesDir = Join-Path $ScriptDir "..\..\templates"
        $templatePath = Join-Path $templatesDir "context.template.md"

        if (Test-Path $templatePath) {
            Copy-Item -Path $templatePath -Destination $contextPath -Force
            Log-Info "  ✓ context.md"
        } else {
            # Fallback: create empty file if template not found
            Log-Warning "  ! context.template.md not found, creating empty file"
            "" | Out-File -FilePath $contextPath -Encoding UTF8
            Log-Info "  ✓ context.md (empty)"
        }
    } else {
        Write-Host "  ⊘ context.md (already exists, skipping)" -ForegroundColor Yellow
    }

    $memoryMd = Join-Path $TargetDir "memory.md"
    if (-not (Test-Path $memoryMd)) {
        # Try to copy from template first
        $templatesDir = Join-Path $ScriptDir "..\..\templates"
        $templatePath = Join-Path $templatesDir "memory.template.md"

        if (Test-Path $templatePath) {
            Copy-Item -Path $templatePath -Destination $memoryMd -Force
            Log-Info "  ✓ memory.md"
        } else {
            # Fallback: create empty file if template not found
            Log-Warning "  ! memory.template.md not found, creating empty file"
            "" | Out-File -FilePath $memoryMd -Encoding UTF8
            Log-Info "  ✓ memory.md (empty)"
        }
    } else {
        Write-Host "  ⊘ memory.md (already exists)" -ForegroundColor Yellow
    }

    Write-Success "Memory and analysis structure created"
}

# Create loop state for autonomous workflows (matching bash create_loop_state)
function Create-LoopState {
    param([string]$TargetDir)

    Write-Host "[DEBUG] Create-LoopState called with TargetDir: $TargetDir" -ForegroundColor Gray
    Log-Info "Creating autonomous loop configuration..."

    Write-Host "[DEBUG] Building loop state path..." -ForegroundColor Gray
    $loopStatePath = Join-Path $TargetDir "state.json"
    Write-Host "[DEBUG] loopStatePath: $loopStatePath" -ForegroundColor Gray

    # Only create if not exists (preserve existing configurations on re-run)
    if (Test-Path $loopStatePath) {
        Write-Host "  ⊘ state.json (already exists, skipping)" -ForegroundColor Yellow
        return
    }

    # Try to copy from template first
    $templatesDir = Join-Path $ScriptDir "..\..\templates"
    $templatePath = Join-Path $templatesDir "state.template.json"

    if (Test-Path $templatePath) {
        # Copy from template and replace timestamp placeholder
        $timestamp = [System.DateTime]::UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ")
        $content = (Get-Content -Path $templatePath -Raw) -replace '\[TIMESTAMP\]', $timestamp
        Set-Content -Path $loopStatePath -Value $content
        Log-Info "  ✓ state.json"
    } else {
        # Fallback: create empty file if template not found
        Log-Warning "  ! state.template.json not found, creating empty file"
        "" | Out-File -FilePath $loopStatePath -Encoding UTF8
        Log-Info "  ✓ state.json (empty)"
    }

    Write-Success "Loop state created"
}

# Main function - enhanced to match Bash main()
function Main {
    # Check for help and version flags first
    if ($Help) {
        Show-Help
        exit 0
    }

    if ($Version) {
        Show-Version
        exit 0
    }

    # Parse arguments
    $parsed = Parse-Arguments -Args $RemainingArguments
    $aiChoice = $parsed.AIChoice
    $targetDir = $parsed.TargetDir

    # Override with explicit parameters if provided
    if ($AI) { $aiChoice = $AI }
    if ($Target) { $targetDir = $Target }

    # Convert to absolute path
    $targetDir = Resolve-Path -Path $targetDir -ErrorAction SilentlyContinue
    if (-not $targetDir) {
        $targetDir = $parsed.TargetDir
    }

    # Show banner
    $version = Show-Version
    Write-Banner "Plaesy Spec-Kit Initialization" "Constitutional Development Framework"

    # Validate target directory
    if (-not (Test-TargetDirectory -TargetDir $targetDir)) {
        exit 1
    }

    Log-Info "Target directory: $targetDir"

    # Change to target directory
    try {
        Set-Location $targetDir
    }
    catch {
        Log-Error "Cannot change to directory: $targetDir"
        exit 1
    }

    # Get AI choice - enhanced to match Bash logic
    if ([string]::IsNullOrEmpty($aiChoice)) {
        Write-Host "Detecting AI platform..." -ForegroundColor Yellow
        $detected = Find-AIPlatform

        if ($detected) {
            Write-Host "Detected: $(Get-PlatformName -Platform $detected)" -ForegroundColor Green

            # Check if we're in interactive mode
            if ($Host.UI.RawUI) {
                $useDetected = Read-Host "Use detected platform? (Y/n)"
                if ($useDetected -notmatch '^[nN]') {
                    $aiChoice = $detected
                }
            } else {
                # Non-interactive mode, use detected platform automatically
                Write-Host "Auto-using detected platform (non-interactive mode)" -ForegroundColor Yellow
                $aiChoice = $detected
            }
        }

        if ([string]::IsNullOrEmpty($aiChoice)) {
            # Always show platform selection when no AI selected
            Write-Host "Please choose from available platforms:" -ForegroundColor Cyan
            Write-Host ""

            $aiChoice = Get-AIChoice
        }
    } else {
        # Normalize provided AI choice
        $aiChoice = Normalize-Platform -PlatformInput $aiChoice
        # Debug: Show what we got
        Log-Info "DEBUG: aiChoice='$aiChoice', IsNullOrEmpty=$([string]::IsNullOrEmpty($aiChoice))"

        # Handle empty/whitespace input by forcing interactive mode
        if ([string]::IsNullOrWhiteSpace($aiChoice)) {
            Write-Host "Empty AI platform detected, showing selection menu..." -ForegroundColor Yellow
            $aiChoice = Get-AIChoice
        } elseif (-not (Test-PlatformExists -Platform $aiChoice) -and $aiChoice -ne "none") {
            Log-Error "Invalid AI platform: '$aiChoice'"
            Write-Host "Available platforms:" -ForegroundColor Cyan
            $platforms = Load-Platforms
            foreach ($platform in $platforms) {
                $displayName = Get-PlatformName -Platform $platform
                Write-Host "  - $displayName"
            }
            exit 1
        }
    }

    # Check if user cancelled
    if ($aiChoice -eq "cancel") {
        Write-Host ""
        Write-Host "Initialization cancelled by user." -ForegroundColor Yellow
        exit 0
    }

    Write-Success "Selected: $(Get-PlatformName -Platform $aiChoice)"

    # Create structure
    New-ProjectStructure -AIPlatform $aiChoice

    # Copy files from spec-kit to project
    Copy-Instructions -TargetDir ".plaesy"
    Copy-Scripts -TargetDir ".plaesy"
    New-MemoryStructure -TargetDir ".plaesy"

    # Create autonomous loop configuration
    Create-LoopState -TargetDir ".plaesy"

    # Setup platform configuration
    if ($aiChoice -ne "none") {
        Write-Host "[DEBUG] About to call Set-AISpecificConfig with platform: $aiChoice" -ForegroundColor Gray
        try {
            Set-AISpecificConfig -Platform $aiChoice
            Write-Host "[DEBUG] Set-AISpecificConfig completed successfully" -ForegroundColor Gray
        } catch {
            Write-Host "[ERROR] Set-AISpecificConfig failed with error:" -ForegroundColor Red
            Write-Host "[ERROR] Message: $($_.Exception.Message)" -ForegroundColor Red
            Write-Host "[ERROR] StackTrace: $($_.Exception.StackTrace)" -ForegroundColor Red
            throw $_
        }
    }

    # Success message
    Write-Host ""
    Write-Host "Plaesy Spec-Kit initialization completed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Project: $(Get-Location)" -ForegroundColor Cyan
    Write-Host "Platform: $(Get-PlatformName -Platform $aiChoice)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Next Steps:" -ForegroundColor Yellow
    Write-Host "1. Start your project with /start command"
    Write-Host "2. Use /continue for automated workflow"
    Write-Host "3. Use /implement for implementation phase"
    Write-Host ""
    Write-Host "Constitutional Development Framework Active" -ForegroundColor Magenta
    Write-Host "   Quality through discipline. Excellence through automation."
}

# Run main function
Main