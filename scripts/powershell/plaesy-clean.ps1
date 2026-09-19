# Plaesy Clean Script (PowerShell)
# Enhanced to match Bash version - Configurable cleanup using centralized configuration

[CmdletBinding()]
param(
    [Parameter()]
    [string]$TargetDir = ".",

    [Parameter()]
    [switch]$Help,

    [Parameter()]
    [Alias("y")]
    [switch]$Yes,

    [Parameter()]
    [switch]$DryRun,

    [Parameter()]
    [switch]$Backup = $true,

    [Parameter()]
    [string]$Level = "safe",

    # NOTE: no custom -Verbose switch here -- [CmdletBinding()] already adds
    # -Verbose as a common parameter; declaring it again throws "A parameter
    # with the name 'Verbose' was defined multiple times" on every invocation.

    [Parameter()]
    [string]$AI = $null,

    [Parameter()]
    [switch]$NoBackup,

    [Parameter()]
    [string]$CleanupLevel = "safe"
)

# Enhanced error handling - match Bash strict mode
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Global variables - match Bash approach
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ConfigManager = Join-Path $ScriptDir "config-manager.ps1"

# Script-scope variables must be initialized here: the rest of the script reads
# them with the $script: qualifier, and an unset script-scope variable evaluates
# to $null (no auto-vivification) -- which left CLEANUP_LEVEL empty and made
# every -Level/-AI/-Yes/-DryRun switch a silent no-op until assigned.
$script:AUTO_CONFIRM = $false
$script:TARGET_DIRECTORY = "."
$script:AI_CHOICE = $null
$script:CLEANUP_LEVEL = "safe"
$script:DRY_RUN_MODE = $false
$script:BACKUP_ENABLED = $true
$script:VERBOSE_MODE = $false

# $PSBoundParameters must be captured at script scope (this spot, right after
# the param block): Main is a function, so inside Main $PSBoundParameters is
# Main's own empty dictionary -- every ContainsKey('Level'/'Backup'/etc.) check
# would return $false and -Level/-TargetDir/-Backup would be silent no-ops.
$script:BOUND_PARAMS = $PSBoundParameters

# Source common functions (Log-Info/Log-Error/Log-Warning/Write-Banner, etc.)
# This script calls those directly but never loaded the file that defines
# them, so every one of those calls failed with "not recognized" once
# reached. Fall back silently if unavailable, matching plaesy-init.ps1.
$CommonFile = Join-Path $ScriptDir "common.ps1"
if (Test-Path $CommonFile) {
    try {
        . $CommonFile
    }
    catch {
        Write-Host "Warning: Could not load common.ps1, using fallback functions" -ForegroundColor Yellow
    }
}
if (-not (Get-Command Log-Info -ErrorAction SilentlyContinue)) {
    function Log-Info { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Cyan }
    function Log-Success { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
    function Log-Warning { param([string]$Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
    function Log-Error { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }
    function Write-Banner { param([string]$Title, [string]$Subtitle) Write-Host "=== $Title ===" -ForegroundColor Magenta; if ($Subtitle) { Write-Host $Subtitle } }
}
# Show help - enhanced to match Bash
function Show-Help {
    Write-Host "Usage: plaesy clean [OPTIONS] [TARGET_DIR]" -ForegroundColor White
    Write-Host ""
    Write-Host "Remove Plaesy Spec-Kit framework files and directories with safety checks." -ForegroundColor Gray
    Write-Host ""
    Write-Host "ARGUMENTS:" -ForegroundColor Yellow
    Write-Host "  TARGET_DIR    Directory to clean (default: current directory)" -ForegroundColor White
    Write-Host ""
    Write-Host "CLEANUP LEVELS:" -ForegroundColor Yellow
    Write-Host "  safe        Remove framework files only, preserve user code (default)" -ForegroundColor White
    Write-Host "  thorough    Remove framework + specs, preserve user code" -ForegroundColor White
    Write-Host "  complete    Remove everything Plaesy-related (DANGEROUS)" -ForegroundColor White
    Write-Host ""
    Write-Host "OPTIONS:" -ForegroundColor Yellow
    Write-Host "  -y, -yes                 Auto-confirm deletion (skip confirmation prompt)" -ForegroundColor White
    Write-Host "  -h, -help                Show this help message" -ForegroundColor White
    Write-Host "  -dry-run                 Show what would be removed without actually removing" -ForegroundColor White
    Write-Host "  -backup                  Create backup before removal (enabled by default)" -ForegroundColor White
    Write-Host "  -no-backup               Skip backup creation" -ForegroundColor White
    Write-Host "  -level [LEVEL]           Specify cleanup level (safe|thorough|complete)" -ForegroundColor White
    Write-Host "  -ai [PLATFORM]           Specify AI platform to clean (auto-detected if not specified)" -ForegroundColor White
    Write-Host "  -verbose                 Show detailed progress" -ForegroundColor White
    Write-Host ""
    Write-Host "AVAILABLE AI PLATFORMS:" -ForegroundColor Yellow
    Write-Host "  claude      Claude Code configuration" -ForegroundColor White
    Write-Host "  copilot     GitHub Copilot configuration" -ForegroundColor White
    Write-Host "  cursor      Cursor AI configuration" -ForegroundColor White
    Write-Host "  windsurf    Windsurf AI configuration" -ForegroundColor White
    Write-Host "  generic     Generic AI configuration" -ForegroundColor White
    Write-Host ""
    Write-Host "CONFIGURATION:" -ForegroundColor Yellow
    Write-Host "  This script uses centralized configuration from:" -ForegroundColor Gray
    Write-Host "  - scripts/configs/framework-config.json" -ForegroundColor Gray
    Write-Host "  - scripts/configs/ai-platforms.json" -ForegroundColor Gray
    Write-Host "  - scripts/configs/clean-config.json" -ForegroundColor Gray
    Write-Host ""
    Write-Host "EXAMPLES:" -ForegroundColor Yellow
    Write-Host "  plaesy clean                         # Safe cleanup with backup" -ForegroundColor White
    Write-Host "  plaesy clean -dry-run               # Preview what would be removed" -ForegroundColor White
    Write-Host "  plaesy clean thorough -y        # Thorough cleanup without confirmation" -ForegroundColor White
    Write-Host "  plaesy clean -level complete -no-backup  # Complete cleanup without backup" -ForegroundColor White
    Write-Host "  plaesy clean -ai claude             # Clean Claude Code configuration only" -ForegroundColor White
    Write-Host "  plaesy clean -dry-run -level thorough  # Preview thorough cleanup" -ForegroundColor White
    Write-Host ""
    Write-Host "SAFETY FEATURES:" -ForegroundColor Yellow
    Write-Host "  - Configuration validation before cleanup" -ForegroundColor White
    Write-Host "  - Dry-run mode to preview changes" -ForegroundColor White
    Write-Host "  - Automatic backup creation (default)" -ForegroundColor White
    Write-Host "  - Platform-specific cleanup detection" -ForegroundColor White
    Write-Host "  - Clear confirmation prompts" -ForegroundColor White
    Write-Host ""
    Write-Host "Note: This operation removes Plaesy framework files but preserves your source code." -ForegroundColor Gray
    Write-Host "Always use -dry-run first to preview what will be removed." -ForegroundColor Gray
}

# Validate cleanup level - match Bash
function Test-CleanupLevel {
    param([string]$Level)

    $validLevels = @("safe", "thorough", "complete")
    if ($Level -notin $validLevels) {
        Log-Error "Invalid cleanup level: $Level"
        Write-Host "Available levels: $($validLevels -join ', ')" -ForegroundColor Cyan
        exit 1
    }
}

# Validate AI platform - enhanced to match Bash
function Test-AIPlatform {
    param([string]$Platform)

    if (-not (Test-Path $ConfigManager)) {
        # Fallback validation
        $validPlatforms = @("claude", "cursor", "windsurf", "copilot", "generic")
        if ($Platform -notin $validPlatforms) {
            Log-Error "Invalid AI assistant: $Platform"
            Write-Host "Available assistants: $($validPlatforms -join ', ')" -ForegroundColor Cyan
            exit 1
        }
        return
    }

    try {
        $availablePlatforms = & $ConfigManager "detect-platform" 2>$null
        if ($availablePlatforms -and $availablePlatforms -notmatch "^$Platform$") {
            Log-Error "Invalid AI assistant: $Platform"
            Write-Host "Available assistants: claude, copilot, cursor, windsurf, generic" -ForegroundColor Cyan
            exit 1
        }
    }
    catch {
        # Continue with validation
    }
}

# Parse arguments - enhanced to match Bash parse_arguments()
function Parse-Arguments {
    param([string[]]$Args)

    for ($i = 0; $i -lt $Args.Count; $i++) {
        $arg = $Args[$i]

        switch ($arg) {
            { $_ -in @("-h", "--help", "-help") } {
                Show-Help
                exit 0
            }
            { $_ -in @("-y", "--yes", "-yes") } {
                $script:AUTO_CONFIRM = $true
            }
            { $_ -in @("--dry-run", "-dry-run") } {
                $script:DRY_RUN_MODE = $true
            }
            { $_ -in @("--backup", "-backup") } {
                $script:BACKUP_ENABLED = $true
            }
            { $_ -in @("--no-backup", "-no-backup") } {
                $script:BACKUP_ENABLED = $false
            }
            "--level" {
                if ($i + 1 -lt $Args.Count) {
                    $script:CLEANUP_LEVEL = $Args[$i + 1]
                    Test-CleanupLevel -Level $script:CLEANUP_LEVEL
                    $i++
                } else {
                    Log-Error "--level requires a value"
                    exit 1
                }
            }
            "--verbose" {
                $script:VERBOSE_MODE = $true
            }
            "--ai" {
                if ($i + 1 -lt $Args.Count) {
                    $script:AI_CHOICE = $Args[$i + 1]
                    Test-AIPlatform -Platform $script:AI_CHOICE
                    $i++
                } else {
                    Log-Error "--ai requires a value"
                    exit 1
                }
            }
            default {
                if ($arg.StartsWith("--level=")) {
                    $script:CLEANUP_LEVEL = $arg.Substring(8)
                    Test-CleanupLevel -Level $script:CLEANUP_LEVEL
                } elseif ($arg.StartsWith("--ai=")) {
                    $script:AI_CHOICE = $arg.Substring(5)
                    Test-AIPlatform -Platform $script:AI_CHOICE
                } elseif ($arg.StartsWith("-")) {
                    Log-Error "Unknown option: $arg"
                    Write-Host "Run 'plaesy clean -help' for usage information"
                    exit 1
                } else {
                    # Positional argument - target directory
                    if ([string]::IsNullOrEmpty($script:TARGET_DIRECTORY) -or $script:TARGET_DIRECTORY -eq ".") {
                        $script:TARGET_DIRECTORY = $arg
                    } else {
                        Log-Error "Multiple directories specified. Only one directory is allowed."
                        exit 1
                    }
                }
            }
        }
    }

    # Set default target directory if not specified
    if ([string]::IsNullOrEmpty($script:TARGET_DIRECTORY)) {
        $script:TARGET_DIRECTORY = "."
    }
}

# Cleanup function for trap - match Bash
function Invoke-Cleanup {
    param([int]$ExitCode = $LASTEXITCODE)

    if ($ExitCode -ne 0) {
        Write-Host "Clean operation failed. Check the error above." -ForegroundColor Red
    }
}

# Set up error handling
trap {
    Invoke-Cleanup -ExitCode $_.Exception.HResult
}

# Validate environment - enhanced to match Bash validate_clean_environment()
function Test-CleanEnvironment {
    param([string]$Directory)

    if (-not (Test-Path $Directory)) {
        Log-Error "Target directory '$Directory' does not exist."
        exit 1
    }

    if (-not (Test-Path $Directory -PathType Container)) {
        Log-Error "Target '$Directory' is not a directory."
        exit 1
    }

    # Test read access
    try {
        Get-ChildItem $Directory -ErrorAction Stop | Out-Null
    }
    catch {
        Log-Error "Target directory '$Directory' is not readable."
        exit 1
    }

    # Test write access
    try {
        $testFile = Join-Path $Directory "test-write-$(Get-Random).tmp"
        "test" | Out-File -FilePath $testFile -ErrorAction Stop
        Remove-Item $testFile -ErrorAction SilentlyContinue
    }
    catch {
        Log-Error "Target directory '$Directory' is not writable."
        exit 1
    }
}

# Detect all platforms present in the project using platform.json detection patterns - enhanced to match Bash
function Find-AllPlatforms {
    $configFile = Join-Path $ScriptDir "..\configs\platform.json"

    if (-not (Test-Path $configFile)) {
        return ,@("generic_ai")
    }

    if (-not (Test-Path $ConfigManager)) {
        return ,@("generic_ai")
    }

    try {
        $detectedPlatforms = @()

        # Get all available platforms
        $allPlatforms = & $ConfigManager "list-platforms" 2>$null
        if ($allPlatforms) {
            $platforms = $allPlatforms -split "`n" | Where-Object { $_ -and $_.Trim() }

            foreach ($platform in $platforms) {
                # Get detection patterns for this platform
                $detectionPatterns = & $ConfigManager "get-detection-patterns" $platform 2>$null

                if ($detectionPatterns) {
                    $patterns = $detectionPatterns -split "`n" | Where-Object { $_ -and $_.Trim() }

                    foreach ($pattern in $patterns) {
                        $pattern = $pattern.Trim() -replace '^"', '' -replace '"$', ''

                        # Check if pattern exists as file or directory
                        if (Test-Path $pattern) {
                            $detectedPlatforms += $platform
                            break  # Found this platform, move to next
                        }
                    }
                }
            }
        }

        if ($detectedPlatforms.Count -gt 0) {
            # The @() wrap is NOT enough: a single-element array is unwrapped to
            # a bare scalar at the function boundary, so callers indexing [0]
            # get the first CHARACTER of the platform name ('c' for "claude").
            # The comma operator emits the array as a single pipeline element,
            # which preserves it across the return regardless of element count.
            return ,@($detectedPlatforms | Sort-Object -Unique)
        } else {
            return ,@("generic_ai")
        }
    }
    catch {
        return ,@("generic_ai")
    }
}

# Show what will be removed using platform.json mapping - enhanced to match Bash
function Show-CleanupPlan {
    param([string]$Directory)

    Log-Info "Analyzing directories and files to remove based on platform.json mapping..."

    # Validate config manager exists
    if (-not (Test-Path $ConfigManager)) {
        Log-Error "Configuration manager not found: $ConfigManager"
        exit 1
    }

    # Change to target directory for file checks
    try {
        Set-Location $Directory
    }
    catch {
        Log-Error "Cannot change to target directory: $Directory"
        exit 1
    }

    # Determine platform(s) to clean
    $detectedPlatforms = @()
    if ($script:AI_CHOICE) {
        $detectedPlatforms = @($script:AI_CHOICE)
    } else {
        # Auto-detect all platforms present in the project
        $detectedPlatforms = Find-AllPlatforms

        if ($detectedPlatforms.Count -eq 0) {
            $detectedPlatforms = @("generic")
        }
    }

    if ($script:DRY_RUN_MODE) {
        Write-Host "[DRY RUN MODE] No files will be actually removed." -ForegroundColor Blue
        Write-Host ""
    }

    # Build platform display string
    $platformDisplay = if ($detectedPlatforms.Count -eq 1) {
        $detectedPlatforms[0]
    } else {
        $detectedPlatforms -join ", "
    }

    Write-Host "Cleanup Plan (Level: $script:CLEANUP_LEVEL, Platforms: $platformDisplay)" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan

    # Show cleanup level details
    switch ($script:CLEANUP_LEVEL) {
        "safe" {
            Write-Host "Safe cleanup: Framework files only, preserve user code" -ForegroundColor Yellow
        }
        "thorough" {
            Write-Host "Thorough cleanup: Framework + specs, preserve user code" -ForegroundColor Yellow
        }
        "complete" {
            Write-Host "Complete cleanup: Everything Plaesy-related (DANGEROUS)" -ForegroundColor Red
        }
    }

    # Show Plaesy framework directories
    Write-Host ""
    Write-Host "Plaesy Framework Directories:" -ForegroundColor Yellow
    # Note: docs/ is not removed as it may contain important project documentation
    $plaesyDirs = @(".plaesy", "specs")
    if ($script:CLEANUP_LEVEL -eq "complete") {
        $plaesyDirs += @("docs")
    }
    $foundPlaesyDirs = @()

    foreach ($dir in $plaesyDirs) {
        if (Test-Path $dir -PathType Container) {
            Write-Host "  - $dir/" -ForegroundColor White
            $foundPlaesyDirs += $dir
        }
    }

    # Show platform-specific files based on mapping for each detected platform
    $foundAnyPlatform = $false
    foreach ($detectedPlatform in $detectedPlatforms) {
        if ($detectedPlatform -ne "generic") {
            Write-Host ""
            Write-Host "Platform-Specific Files ($detectedPlatform):" -ForegroundColor Yellow
            Write-Host "  Based on platform.json mapping for $detectedPlatform" -ForegroundColor Gray

            $mappingTypes = @("core", "instructions", "prompts", "chatmodes")
            $foundAny = $false

            foreach ($mappingType in $mappingTypes) {
                try {
                    $targetPath = & $ConfigManager "get-mapping-value" $detectedPlatform $mappingType 2>$null

                    if ($targetPath -and $targetPath -ne "null") {
                        if (Test-Path $targetPath) {
                            if (Test-Path $targetPath -PathType Container) {
                                Write-Host "  - $targetPath/ (directory)" -ForegroundColor White
                            } else {
                                Write-Host "  - $targetPath (file)" -ForegroundColor White
                            }
                            $foundAny = $true
                            $foundAnyPlatform = $true
                        }
                    }
                }
                catch {
                    # Continue with next mapping type
                }
            }

            if (-not $foundAny) {
                Write-Host "  - No platform-specific files found for $detectedPlatform" -ForegroundColor Gray
            }
        }
    }

    if (-not $foundAnyPlatform) {
        Write-Host ""
        Write-Host "Platform-Specific Files:" -ForegroundColor Yellow
        Write-Host "  - Generic platform: No specific files to remove" -ForegroundColor Gray
    }

    # Show what will be preserved
    Write-Host ""
    Write-Host "What will be preserved:" -ForegroundColor Green
    Write-Host "  - Your source code (src/, lib/, components/, etc.)" -ForegroundColor Gray
    Write-Host "  - Git repository (.git/)" -ForegroundColor Gray
    Write-Host "  - User-generated files not in Plaesy directories" -ForegroundColor Gray
    Write-Host "  - Configuration files (.env, .gitignore, etc.)" -ForegroundColor Gray

    if ($script:BACKUP_ENABLED -and -not $script:DRY_RUN_MODE) {
        Write-Host "  - Backup will be created before removal" -ForegroundColor Gray
    }

    # Check if anything found to remove
    if ($foundPlaesyDirs.Count -eq 0 -and -not $foundAnyPlatform) {
        Write-Success "No Plaesy directories or files found to remove."
        exit 0
    }

    Write-Host ""

    return @{
        FoundDirectories = $foundPlaesyDirs
        FoundPlatforms = $detectedPlatforms
        FoundAnyPlatform = $foundAnyPlatform
    }
}

# Confirm removal with user - match Bash
function Confirm-Removal {
    if ($script:AUTO_CONFIRM) {
        Write-Host "Auto-confirm mode: proceeding with deletion..." -ForegroundColor Yellow
        return
    }

    Write-Host "This will permanently delete the directories and files listed above." -ForegroundColor Yellow
    $confirmation = Read-Host "Are you sure you want to continue? (y/N)"

    if ($confirmation -match '^[yY]([eE][sS])?$') {
        return
    } else {
        Write-Host "Clean operation cancelled."
        exit 0
    }
}

# Create backup function - enhanced to match Bash
function New-Backup {
    if (-not $script:BACKUP_ENABLED) {
        return
    }

    $backupDir = Join-Path $script:TARGET_DIRECTORY ".plaesy-backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"

    Log-Info "Creating backup: $backupDir"

    try {
        New-Item -ItemType Directory -Path $backupDir -Force | Out-Null

        # Backup important files
        $filesToBackup = @("CLAUDE.md", "AI-INSTRUCTIONS.md", "README.md", ".gitignore")
        foreach ($filePattern in $filesToBackup) {
            Get-ChildItem $script:TARGET_DIRECTORY -Filter $filePattern -File -ErrorAction SilentlyContinue | ForEach-Object {
                Copy-Item -Path $_.FullName -Destination $backupDir -Force
            }
        }

        # Backup JSON and YAML files
        Get-ChildItem $script:TARGET_DIRECTORY -Filter "*.json" -File -ErrorAction SilentlyContinue | ForEach-Object {
            Copy-Item -Path $_.FullName -Destination $backupDir -Force
        }
        Get-ChildItem $script:TARGET_DIRECTORY -Filter "*.yaml" -File -ErrorAction SilentlyContinue | ForEach-Object {
            Copy-Item -Path $_.FullName -Destination $backupDir -Force
        }
        Get-ChildItem $script:TARGET_DIRECTORY -Filter "*.yml" -File -ErrorAction SilentlyContinue | ForEach-Object {
            Copy-Item -Path $_.FullName -Destination $backupDir -Force
        }

        # Backup Plaesy directory if it exists
        $plaesyDir = Join-Path $script:TARGET_DIRECTORY ".plaesy"
        if (Test-Path $plaesyDir) {
            Copy-Item -Path $plaesyDir -Destination $backupDir -Recurse -Force
        }

        # Determine platforms to backup
        $platformsToBackup = @()
        if ($script:AI_CHOICE) {
            $platformsToBackup = @($script:AI_CHOICE)
        } else {
            $platformsToBackup = Find-AllPlatforms
            if ($platformsToBackup.Count -eq 0) {
                $platformsToBackup = @("generic")
            }
        }

        # Backup platform-specific files based on mapping for each detected platform
        foreach ($platform in $platformsToBackup) {
            if ($platform -ne "generic") {
                $mappingTypes = @("core", "instructions", "prompts", "chatmodes")

                foreach ($mappingType in $mappingTypes) {
                    try {
                        $targetPath = & $ConfigManager "get-mapping-value" $platform $mappingType 2>$null

                        if ($targetPath -and $targetPath -ne "null") {
                            $fullPath = Join-Path $script:TARGET_DIRECTORY $targetPath
                            if (Test-Path $fullPath) {
                                if (Test-Path $fullPath -PathType Container) {
                                    Copy-Item -Path $fullPath -Destination $backupDir -Recurse -Force
                                    Log-Info "Backed up platform $platform directory: $targetPath"
                                } else {
                                    Copy-Item -Path $fullPath -Destination $backupDir -Force
                                    Log-Info "Backed up platform $platform file: $targetPath"
                                }
                            }
                        }
                    }
                    catch {
                        # Continue with next mapping type
                    }
                }
            }
        }

        Write-Success "Backup created: $backupDir"
    }
    catch {
        Log-Error "Failed to create backup: $($_.Exception.Message)"
    }
}

# Remove plaesy directories based on platform.json mapping - enhanced to match Bash
function Remove-PlaesyDirectories {
    Log-Info "Removing plaesy directories based on platform.json mapping..."

    $removedCount = 0
    $failedRemovals = @()

    # Change to target directory
    try {
        Set-Location $script:TARGET_DIRECTORY
    }
    catch {
        Log-Error "Failed to change to target directory: $script:TARGET_DIRECTORY"
        return
    }

    # Determine platforms to clean
    $platformsToClean = @()
    if ($script:AI_CHOICE) {
        $platformsToClean = @($script:AI_CHOICE)
    } else {
        # Auto-detect all platforms present in the project
        $platformsToClean = Find-AllPlatforms

        if ($platformsToClean.Count -eq 0) {
            $platformsToClean = @("generic")
        }
    }

    # Always remove Plaesy framework directories
    # Note: docs/ is not removed as it may contain important project documentation
    $plaesyDirs = @(".plaesy", "specs")
    if ($script:CLEANUP_LEVEL -eq "complete") {
        $plaesyDirs += @("docs")
    }

    Log-Info "Removing Plaesy framework directories..."
    foreach ($dir in $plaesyDirs) {
        if (Test-Path $dir -PathType Container) {
            try {
                Remove-Item $dir -Recurse -Force -ErrorAction Stop
                Write-Host "  ✅ Removed Plaesy framework: $dir" -ForegroundColor Green
                $removedCount++
            }
            catch {
                $failedRemovals += $dir
                Write-Host "  ❌ Failed to remove: $dir" -ForegroundColor Red
            }
        }
    }

    # Remove platform-specific files for each detected platform
    foreach ($platformToClean in $platformsToClean) {
        if ($platformToClean -ne "generic") {
            Log-Info "Removing platform-specific files for: $platformToClean"

            # Get platform mappings from config manager
            $mappingTypes = @("core", "instructions", "prompts", "chatmodes")

            foreach ($mappingType in $mappingTypes) {
                try {
                    $targetPath = & $ConfigManager "get-mapping-value" $platformToClean $mappingType 2>$null

                    if ($targetPath -and $targetPath -ne "null") {
                        # Check if target is a directory or file
                        if ($targetPath.EndsWith("/")) {
                            # It's a directory
                            $dirPath = $targetPath.TrimEnd("/")
                            if (Test-Path $dirPath -PathType Container) {
                                try {
                                    Remove-Item $dirPath -Recurse -Force -ErrorAction Stop
                                    Write-Host "  ✅ Removed platform $platformToClean $mappingType`: $dirPath" -ForegroundColor Green
                                    $removedCount++
                                }
                                catch {
                                    $failedRemovals += $dirPath
                                    Write-Host "  ❌ Failed to remove: $dirPath" -ForegroundColor Red
                                }
                            }
                        } else {
                            # It's a file or directory (need to check)
                            if (Test-Path $targetPath -PathType Leaf) {
                                try {
                                    Remove-Item $targetPath -Force -ErrorAction Stop
                                    Write-Host "  ✅ Removed platform $platformToClean $mappingType`: $targetPath" -ForegroundColor Green
                                    $removedCount++
                                }
                                catch {
                                    $failedRemovals += $targetPath
                                    Write-Host "  ❌ Failed to remove: $targetPath" -ForegroundColor Red
                                }
                            } elseif (Test-Path $targetPath -PathType Container) {
                                try {
                                    Remove-Item $targetPath -Recurse -Force -ErrorAction Stop
                                    Write-Host "  ✅ Removed platform $platformToClean $mappingType`: $targetPath" -ForegroundColor Green
                                    $removedCount++
                                }
                                catch {
                                    $failedRemovals += $targetPath
                                    Write-Host "  ❌ Failed to remove: $targetPath" -ForegroundColor Red
                                }
                            }
                        }
                    }
                }
                catch {
                    # Continue with next mapping type
                }
            }

            # Clean up empty parent directories
            $parentDirs = @()
            foreach ($mappingType in $mappingTypes) {
                try {
                    $targetPath = & $ConfigManager "get-mapping-value" $platformToClean $mappingType 2>$null

                    if ($targetPath -and $targetPath -ne "null") {
                        $parentDir = Split-Path $targetPath -Parent
                        if ($parentDir -and $parentDir -ne ".") {
                            $parentDirs += $parentDir
                        }
                    }
                }
                catch {
                    # Continue with next mapping type
                }
            }

            # Remove duplicate parent directories
            $uniqueParents = $parentDirs | Sort-Object -Unique

            foreach ($parentDir in $uniqueParents) {
                if (Test-Path $parentDir -PathType Container) {
                    try {
                        $isEmpty = @(Get-ChildItem $parentDir -ErrorAction SilentlyContinue).Count -eq 0
                        if ($isEmpty) {
                            Remove-Item $parentDir -Force -ErrorAction Stop
                            Write-Host "  ✅ Removed empty parent directory: $parentDir" -ForegroundColor Green
                            $removedCount++
                        }
                    }
                    catch {
                        # Continue if directory removal fails
                    }
                }
            }
        }
    }

    # Clean up empty .github directory if it exists and is empty
    if (Test-Path ".github" -PathType Container) {
        try {
            $githubContents = Get-ChildItem ".github" -Force -ErrorAction SilentlyContinue
            if ($githubContents.Count -eq 0) {
                Remove-Item ".github" -Force -ErrorAction Stop
                Write-Host "  ✅ Removed empty: .github" -ForegroundColor Green
                $removedCount++
            }
        }
        catch {
            # Ignore errors when removing empty .github directory
        }
    }

    if ($failedRemovals.Count -gt 0) {
        Log-Warning "Failed to remove some items: $($failedRemovals -join ', ')"
    }

    if ($removedCount -gt 0) {
        Write-Success "Successfully removed $removedCount items"
    } else {
        Log-Warning "No items were removed"
    }
}

# Main function - enhanced to match Bash
function Main {
    if ($Help) { Show-Help; exit 0 }

    # Parse command line arguments (legacy bash-style long options, e.g. if
    # invoked via a wrapper that forwards raw args instead of native params)
    Parse-Arguments -Args $args

    # Override with the script's own native PowerShell parameters. These are
    # bound by PowerShell directly (not through $args, which is why
    # Parse-Arguments above sees nothing when callers use -Yes/-Level/etc.
    # the normal way) -- without this block none of -Yes/-DryRun/-Backup/
    # -NoBackup/-Level/-CleanupLevel/-AI/-TargetDir ever took effect.
    if ($Yes) { $script:AUTO_CONFIRM = $true }
    if ($DryRun) { $script:DRY_RUN_MODE = $true }
    if ($NoBackup) { $script:BACKUP_ENABLED = $false }
    elseif ($script:BOUND_PARAMS.ContainsKey('Backup')) { $script:BACKUP_ENABLED = $Backup }
    elseif (-not $script:BACKUP_ENABLED) { $script:BACKUP_ENABLED = $true }
    if ($script:BOUND_PARAMS.ContainsKey('CleanupLevel') -and $CleanupLevel) {
        $script:CLEANUP_LEVEL = $CleanupLevel
        Test-CleanupLevel -Level $script:CLEANUP_LEVEL
    } elseif ($script:BOUND_PARAMS.ContainsKey('Level') -and $Level) {
        $script:CLEANUP_LEVEL = $Level
        Test-CleanupLevel -Level $script:CLEANUP_LEVEL
    }
    if ($AI) {
        $script:AI_CHOICE = $AI
        Test-AIPlatform -Platform $script:AI_CHOICE
    }
    if ($script:BOUND_PARAMS.ContainsKey('TargetDir') -and $TargetDir -and $TargetDir -ne ".") {
        $script:TARGET_DIRECTORY = $TargetDir
    }

    # Show configuration if verbose
    if ($script:VERBOSE_MODE) {
        Write-Host "Configuration:" -ForegroundColor Blue
        Write-Host "  Cleanup Level: $script:CLEANUP_LEVEL"
        Write-Host "  Dry Run: $script:DRY_RUN_MODE"
        Write-Host "  Backup: $script:BACKUP_ENABLED"
        Write-Host "  Platform: $(if ($script:AI_CHOICE) { $script:AI_CHOICE } else { 'auto-detect' })"
        Write-Host "  Target Directory: $script:TARGET_DIRECTORY"
        Write-Host ""
    }

    # AI platform detection if not specified
    if ([string]::IsNullOrEmpty($script:AI_CHOICE)) {
        Write-Host "🔍 Detecting AI platforms for cleanup..." -ForegroundColor Yellow

        # Change to target directory for detection
        try {
            Set-Location $script:TARGET_DIRECTORY
        }
        catch {
            Log-Error "Cannot change to target directory: $script:TARGET_DIRECTORY"
            exit 1
        }

        # Detect all platforms present in the project
        $detectedPlatforms = Find-AllPlatforms

        if ($detectedPlatforms.Count -eq 1) {
            Write-Host "✅ Detected AI platform: $($detectedPlatforms[0])" -ForegroundColor Green
            Write-Host "🧹 Using platform-adapted cleanup..." -ForegroundColor Yellow
            $script:AI_CHOICE = $detectedPlatforms[0]
        } elseif ($detectedPlatforms.Count -gt 1) {
            Write-Host "✅ Detected multiple AI platforms: $($detectedPlatforms -join ', ')" -ForegroundColor Green
            Write-Host "🧹 Using multi-platform cleanup..." -ForegroundColor Yellow
            $script:AI_CHOICE = $detectedPlatforms -join ", "
        } else {
            Write-Host "🤖 No specific AI platform detected. Using generic cleanup..." -ForegroundColor Yellow
            $script:AI_CHOICE = "generic"
        }
    } else {
        Write-Host "✅ Using specified AI: $script:AI_CHOICE" -ForegroundColor Green
    }

    # Validate configuration files
    if (Test-Path $ConfigManager) {
        try {
            & $ConfigManager "validate" | Out-Null
        }
        catch {
            Log-Error "Configuration validation failed"
            exit 1
        }
    }

    Write-Banner "Plaesy Clean" "Remove plaesy framework directories"

    # Validate environment
    Test-CleanEnvironment -Directory $script:TARGET_DIRECTORY

    # Show what will be cleaned
    $cleanupPlan = Show-CleanupPlan -Directory $script:TARGET_DIRECTORY

    # Confirm with user (or auto-confirm)
    Confirm-Removal

    # Create backup if needed
    if ($script:BACKUP_ENABLED -and -not $script:DRY_RUN_MODE) {
        New-Backup
    }

    # Perform cleanup
    if (-not $script:DRY_RUN_MODE) {
        Remove-PlaesyDirectories
    }

    Write-Host ""
    Write-Success "🧹 Plaesy clean completed!"

    # Show platform-specific guidance
    if ($script:AI_CHOICE) {
        Write-Host ""
        Write-Host "🎯 Platform-specific cleanup guidance:" -ForegroundColor Cyan
        switch ($script:AI_CHOICE.ToLower()) {
            "claude" {
                Write-Host "💡 For Claude Code: Use /clean command in future" -ForegroundColor Gray
            }
            "cursor" {
                Write-Host "💡 For Cursor AI: Use --clean command in future" -ForegroundColor Gray
            }
            "windsurf" {
                Write-Host "💡 For Windsurf AI: Use clean command in future" -ForegroundColor Gray
            }
            "copilot" {
                Write-Host "💡 For GitHub Copilot: Continue using plaesy clean command" -ForegroundColor Gray
            }
            default {
                Write-Host "For other platforms: Use plaesy clean command" -ForegroundColor Gray
            }
        }
    }

    Write-Host ""
    $location = Get-Location
    Write-Host "Cleaned directory: $location" -ForegroundColor Cyan
    Write-Host ""
    Write-Host 'To reinitialize the project:' -ForegroundColor Yellow
    Write-Host '   plaesy init [-ai=your-choice]' -ForegroundColor Gray
    Write-Host ""
}

# Run main function
Main
