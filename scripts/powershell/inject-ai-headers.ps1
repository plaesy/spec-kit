# AI-Specific Header Injection Script for PowerShell
# Injects platform-specific headers into prompt files based on chosen AI platform

param(
    [Parameter(Mandatory=$true, HelpMessage="AI platform to configure headers for")]
    [ValidateSet("copilot", "cursor", "windsurf", "claude", "chatgpt", "gemini", "trae-ai", "qwen-code", "codex-cli", "opencode-cli", "local-ai", "manual")]
    [string]$AiPlatform,

    [Parameter(Mandatory=$true, HelpMessage="Target directory containing prompt files")]
    [string]$TargetDirectory,

    [Parameter(HelpMessage="Show what would be changed without making changes")]
    [switch]$DryRun,

    [Parameter(HelpMessage="Overwrite existing headers without confirmation")]
    [switch]$Force,

    [Parameter(HelpMessage="Create backup of original files")]
    [switch]$Backup,

    [Parameter(HelpMessage="Show help message")]
    [switch]$Help
)

# Script configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$RepoRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$HeadersDir = Join-Path $RepoRoot "templates\ai-headers"
$DefaultHeader = "manual.header.md"

# Color configuration for output
$Colors = @{
    Info = "Cyan"
    Success = "Green"
    Warning = "Yellow"
    Error = "Red"
}

# Logging functions
function Write-LogInfo {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Info
}

function Write-LogSuccess {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Success
}

function Write-LogWarning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Warning
}

function Write-LogError {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Error
}

# Help function
function Show-Help {
    @"
AI Header Injection Script for PowerShell

USAGE:
    .\inject-ai-headers.ps1 -AiPlatform <platform> -TargetDirectory <directory> [OPTIONS]

PARAMETERS:
    -AiPlatform <platform>     AI platform to configure headers for
    -TargetDirectory <dir>     Target directory containing prompt files
    -DryRun                    Show what would be changed without making changes
    -Force                     Overwrite existing headers without confirmation
    -Backup                    Create backup of original files
    -Help                      Show this help message

SUPPORTED AI PLATFORMS:
    copilot              GitHub Copilot
    cursor               Cursor AI
    windsurf             Windsurf AI
    claude               Claude Code
    chatgpt              ChatGPT
    gemini               Google Gemini
    trae-ai              Trae.ai Multi-Agent
    qwen-code            Qwen Code
    codex-cli            Codex CLI
    opencode-cli         OpenCode CLI
    local-ai             Local AI Models (Ollama, LM Studio)
    manual               Manual Development (no AI)

EXAMPLES:
    .\inject-ai-headers.ps1 -AiPlatform cursor -TargetDirectory .\prompts
    .\inject-ai-headers.ps1 -AiPlatform claude -TargetDirectory . -Backup -DryRun
    .\inject-ai-headers.ps1 -AiPlatform copilot -TargetDirectory .\my-project -Force

"@
}

# Get header file path
function Get-HeaderFile {
    param([string]$Platform)

    $HeaderFile = Join-Path $HeadersDir "$Platform.header.md"

    if (Test-Path $HeaderFile) {
        return $HeaderFile
    } else {
        Write-LogWarning "No specific header for $Platform, using default header"
        return Join-Path $HeadersDir $DefaultHeader
    }
}

# Find prompt files
function Find-PromptFiles {
    param([string]$TargetDir)

    if (-not (Test-Path $TargetDir)) {
        Write-LogError "Target directory does not exist: $TargetDir"
        return @()
    }

    return Get-ChildItem -Path $TargetDir -Filter "*.prompt.md" -File -Recurse
}

# Check if file already has header
function Test-HasHeader {
    param([string]$FilePath)

    $FirstLines = Get-Content $FilePath -TotalCount 20 -ErrorAction SilentlyContinue

    if ($FirstLines) {
        $HeaderIndicators = @("^---$", "Constitutional.*Framework", "AI.*Configuration", "# .*Header")

        foreach ($Line in $FirstLines) {
            foreach ($Pattern in $HeaderIndicators) {
                if ($Line -match $Pattern) {
                    return $true
                }
            }
        }
    }

    return $false
}

# Create backup of file
function New-BackupFile {
    param([string]$FilePath)

    $BackupPath = "$FilePath.backup.$(Get-Date -Format 'yyyyMMdd_HHmmss')"
    Copy-Item $FilePath $BackupPath
    Write-LogInfo "Created backup: $(Split-Path $BackupPath -Leaf)"
    return $BackupPath
}

# Remove existing header from file
function Remove-ExistingHeader {
    param([string]$FilePath)

    $Content = Get-Content $FilePath
    $HeaderEndLine = -1
    $HeaderCount = 0

    for ($i = 0; $i -lt $Content.Length; $i++) {
        if ($Content[$i] -eq "---") {
            $HeaderCount++
            if ($HeaderCount -eq 2) {
                $HeaderEndLine = $i
                break
            }
        }
    }

    if ($HeaderEndLine -ge 0) {
        $NewContent = $Content[($HeaderEndLine + 2)..($Content.Length - 1)]
        Set-Content -Path $FilePath -Value $NewContent
        Write-LogInfo "Removed existing header from $(Split-Path $FilePath -Leaf)"
    }
}

# Inject header into file
function Add-HeaderToFile {
    param(
        [string]$FilePath,
        [string]$HeaderContent
    )

    $OriginalContent = Get-Content $FilePath -Raw
    $NewContent = $HeaderContent + "`n`n" + $OriginalContent
    Set-Content -Path $FilePath -Value $NewContent
}

# Process single file
function Invoke-ProcessFile {
    param(
        [System.IO.FileInfo]$File,
        [string]$HeaderContent,
        [bool]$Force,
        [bool]$Backup,
        [bool]$DryRun
    )

    $FileName = $File.Name
    $FilePath = $File.FullName

    if ($DryRun) {
        if ((Test-HasHeader $FilePath) -and -not $Force) {
            Write-LogInfo "[DRY RUN] Would skip $FileName (already has header, use -Force to overwrite)"
        } else {
            Write-LogInfo "[DRY RUN] Would inject header into $FileName"
        }
        return $true
    }

    # Check if file already has header
    if ((Test-HasHeader $FilePath) -and -not $Force) {
        Write-LogWarning "Skipping $FileName (already has header, use -Force to overwrite)"
        return $false
    }

    # Create backup if requested
    if ($Backup) {
        New-BackupFile $FilePath | Out-Null
    }

    # Remove existing header if forcing update
    if ($Force -and (Test-HasHeader $FilePath)) {
        Remove-ExistingHeader $FilePath
    }

    # Inject new header
    Add-HeaderToFile $FilePath $HeaderContent
    Write-LogSuccess "Injected header into $FileName"
    return $true
}

# Main execution
function Invoke-Main {
    # Show help if requested
    if ($Help) {
        Show-Help
        return
    }

    Write-LogInfo "Starting header injection for AI platform: $AiPlatform"

    # Get header file
    $HeaderFile = Get-HeaderFile $AiPlatform

    if (-not (Test-Path $HeaderFile)) {
        Write-LogError "Header file not found: $HeaderFile"
        return
    }

    # Load header content
    $HeaderContent = Get-Content $HeaderFile -Raw
    Write-LogInfo "Using header file: $(Split-Path $HeaderFile -Leaf)"

    # Find prompt files
    $PromptFiles = Find-PromptFiles $TargetDirectory

    if ($PromptFiles.Count -eq 0) {
        Write-LogWarning "No prompt files found in $TargetDirectory"
        return
    }

    Write-LogInfo "Found $($PromptFiles.Count) prompt file(s)"

    # Process each file
    $Processed = 0
    $Skipped = 0

    foreach ($File in $PromptFiles) {
        if (Invoke-ProcessFile $File $HeaderContent $Force $Backup $DryRun) {
            $Processed++
        } else {
            $Skipped++
        }
    }

    # Summary
    Write-Host ""
    if ($DryRun) {
        Write-LogInfo "DRY RUN COMPLETE"
        Write-LogInfo "Would process: $Processed files"
        Write-LogInfo "Would skip: $Skipped files"
    } else {
        Write-LogSuccess "HEADER INJECTION COMPLETE"
        Write-LogSuccess "Processed: $Processed files"
        if ($Skipped -gt 0) {
            Write-LogInfo "Skipped: $Skipped files"
        }
    }

    # Create configuration file
    if (-not $DryRun -and $Processed -gt 0) {
        $ConfigFile = Join-Path $TargetDirectory ".plaesy-headers.json"
        $Config = @{
            ai_platform = $AiPlatform
            injection_timestamp = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
            processed_files = $Processed
            header_file = Split-Path $HeaderFile -Leaf
            constitutional_framework_version = "3.0.0"
        } | ConvertTo-Json -Depth 3

        Set-Content -Path $ConfigFile -Value $Config
        Write-LogInfo "Created configuration: $(Split-Path $ConfigFile -Leaf)"
    }
}

# Execute main function
try {
    Invoke-Main
} catch {
    Write-LogError "An error occurred: $($_.Exception.Message)"
    exit 1
}