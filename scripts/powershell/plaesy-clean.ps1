# Plaesy Clean Script (PowerShell)
# Remove directories created by plaesy init

param(
  [string]$TargetDir = ".",
  [switch]$Help,
  [Alias("y")]
  [switch]$Yes
)$ErrorActionPreference = "Stop"

# Colors
function Write-Step { 
    param([string]$Message)
    Write-Host "▶ $Message" -ForegroundColor Blue
}

function Write-Success { 
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Warning { 
    param([string]$Message)
    Write-Host "⚠️ $Message" -ForegroundColor Yellow
}

function Write-Error-Custom { 
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
    exit 1
}

function Write-Banner {
    param([string]$Title, [string]$Subtitle)
  
    Write-Host "██████╗ ██╗      █████╗ ███████╗███████╗██╗   ██╗" -ForegroundColor Magenta
    Write-Host "██╔══██╗██║     ██╔══██╗██╔════╝██╔════╝╝██╗ ██╔╝" -ForegroundColor Magenta
    Write-Host "██████╔╝██║     ███████║█████╗  ███████╗ ╚████╔╝ " -ForegroundColor Magenta
    Write-Host "██╔═══╝ ██║     ██╔══██║██╔══╝  ╚════██║  ╚██╔╝  " -ForegroundColor Magenta
    Write-Host "██║     ███████╗██║  ██║███████╗███████║   ██║   " -ForegroundColor Magenta
    Write-Host "╚═╝     ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   " -ForegroundColor Magenta
    Write-Host ""
    if ($Title) {
        Write-Host "🏛️  $Title" -ForegroundColor Magenta
    }
    if ($Subtitle) {
        Write-Host "   $Subtitle" -ForegroundColor Magenta
    }
    Write-Host ""
}

function Show-Help {
    @"
Usage: plaesy-clean.ps1 [OPTIONS]

Remove directories created by plaesy init command.

Parameters:
  -TargetDir <path>    Directory to clean (default: current directory)
  -Yes, -y             Auto-confirm deletion (skip confirmation prompt)
  -Help                Show this help message

Directories that will be removed:
  - .plaesy\                (Core framework files)
  - .github\chatmodes\      (GitHub Copilot chat modes)
  - .github\prompts\        (GitHub Copilot prompts)
  - .cursor\                (Cursor AI configuration)
  - .claude\                (Claude AI configuration)
  - .windsurf\              (Windsurf AI configuration)
  - .trae\                  (Trae.ai configuration)
  - .qwen\                  (Qwen Code configuration)
  - .kilo\                  (Kilo Code configuration)
  - .codex\                 (Codex configuration)
  - .opencode\              (OpenCode configuration)
  - .ai-config\             (Generic AI configuration)

Files that will be removed:
  - .github\copilot-instructions.md  (GitHub Copilot instructions)
  - .cursor\instructions.md           (Cursor AI instructions)
  - .claude\instructions.md           (Claude AI instructions)
  - .windsurf\instructions.md         (Windsurf AI instructions)
  - .ai-config\instructions.md        (Generic AI instructions)

Examples:
  .\plaesy-clean.ps1                           # Clean current directory
  .\plaesy-clean.ps1 -TargetDir "C:\MyProject" # Clean specific project directory
  .\plaesy-clean.ps1 -Yes                      # Clean current directory without confirmation
  .\plaesy-clean.ps1 -y -TargetDir "C:\MyProject" # Clean specific directory without confirmation

Note: This operation cannot be undone. Use with caution.
"@
}

function Test-CleanEnvironment {
    param([string]$Directory)
  
    if (!(Test-Path $Directory)) {
        Write-Error-Custom "Target directory '$Directory' does not exist."
    }
  
    try {
        $testAccess = Get-ChildItem $Directory -ErrorAction Stop | Out-Null
    }
    catch {
        Write-Error-Custom "Target directory '$Directory' is not accessible."
    }
}

function Show-CleanupPlan {
    param([string]$Directory)
  
    Write-Step "Analyzing directories and files to remove..."
  
    $DirectoriesToRemove = @(
        ".plaesy",
        ".github\chatmodes",
        ".github\prompts",
        ".cursor",
        ".claude",
        ".windsurf",
        ".trae",
        ".qwen",
        ".kilo",
        ".codex",
        ".opencode",
        ".ai-config"
    )

    $FilesToRemove = @(
        ".github\copilot-instructions.md",
        ".cursor\instructions.md",
        ".claude\instructions.md",
        ".windsurf\instructions.md",
        ".ai-config\instructions.md"
    )
  
    $FoundDirectories = @()
    $FoundFiles = @()
  
    foreach ($Dir in $DirectoriesToRemove) {
        $FullPath = Join-Path $Directory $Dir
        if (Test-Path $FullPath -PathType Container) {
            $FoundDirectories += $Dir
        }
    }

    foreach ($File in $FilesToRemove) {
        $FullPath = Join-Path $Directory $File
        if (Test-Path $FullPath -PathType Leaf) {
            $FoundFiles += $File
        }
    }
  
    if ($FoundDirectories.Count -eq 0 -and $FoundFiles.Count -eq 0) {
        Write-Success "No plaesy directories or files found to remove."
        exit 0
    }
  
    if ($FoundDirectories.Count -gt 0) {
        Write-Host "The following directories will be removed:" -ForegroundColor Cyan
        foreach ($Dir in $FoundDirectories) {
            Write-Host "  - $Dir"
        }
    }

    if ($FoundFiles.Count -gt 0) {
        Write-Host "The following files will be removed:" -ForegroundColor Cyan
        foreach ($File in $FoundFiles) {
            Write-Host "  - $File"
        }
    }
    Write-Host ""
  
    return @($FoundDirectories, $FoundFiles)
}

function Confirm-Removal {
  if ($Yes) {
    Write-Host "⚠️  Auto-confirm mode: proceeding with deletion..." -ForegroundColor Yellow
    return $true
  }
  
  Write-Host "⚠️  This will permanently delete the directories and files listed above." -ForegroundColor Yellow
  $Confirmation = Read-Host "Are you sure you want to continue? (y/N)"

  if ($Confirmation -match '^[yY]([eE][sS])?$') {
    return $true
  }
  else {
    Write-Host "Clean operation cancelled."
    exit 0
  }
}

function Remove-PleasyDirectories {
    param([string]$Directory)
  
    Write-Step "Removing plaesy directories and files..."
  
    $DirectoriesToRemove = @(
        ".plaesy",
        ".github\chatmodes",
        ".github\prompts",
        ".cursor",
        ".claude",
        ".windsurf",
        ".trae",
        ".qwen",
        ".kilo",
        ".codex",
        ".opencode",
        ".ai-config"
    )

    $FilesToRemove = @(
        ".github\copilot-instructions.md",
        ".cursor\instructions.md",
        ".claude\instructions.md",
        ".windsurf\instructions.md",
        ".ai-config\instructions.md"
    )
  
    $RemovedCount = 0
    $FailedRemovals = @()
  
    # Change to target directory
    try {
        Set-Location $Directory
    }
    catch {
        Write-Error-Custom "Failed to change to target directory: $Directory"
    }
  
    # Remove directories
    foreach ($Dir in $DirectoriesToRemove) {
        if (Test-Path $Dir -PathType Container) {
            try {
                Remove-Item $Dir -Recurse -Force -ErrorAction Stop
                Write-Host "  ✅ Removed directory: $Dir" -ForegroundColor Green
                $RemovedCount++
            }
            catch {
                $FailedRemovals += $Dir
                Write-Host "  ❌ Failed to remove directory: $Dir" -ForegroundColor Red
            }
        }
    }

    # Remove files
    foreach ($File in $FilesToRemove) {
        if (Test-Path $File -PathType Leaf) {
            try {
                Remove-Item $File -Force -ErrorAction Stop
                Write-Host "  ✅ Removed file: $File" -ForegroundColor Green
                $RemovedCount++
            }
            catch {
                $FailedRemovals += $File
                Write-Host "  ❌ Failed to remove file: $File" -ForegroundColor Red
            }
        }
    }
            catch {
                $FailedRemovals += $Dir
                Write-Host "  ❌ Failed to remove: $Dir" -ForegroundColor Red
            }
        }
    }
  
    # Clean up empty .github directory if it exists and is empty
    if ((Test-Path ".github" -PathType Container)) {
        $GitHubContents = Get-ChildItem ".github" -Force -ErrorAction SilentlyContinue
        if ($GitHubContents.Count -eq 0) {
            try {
                Remove-Item ".github" -Force -ErrorAction Stop
                Write-Host "  ✅ Removed empty: .github" -ForegroundColor Green
                $RemovedCount++
            }
            catch {
                # Ignore errors when removing empty .github directory
            }
        }
    }
  
    if ($FailedRemovals.Count -gt 0) {
        Write-Warning "Failed to remove some directories: $($FailedRemovals -join ', ')"
    }
  
    if ($RemovedCount -gt 0) {
        Write-Success "Successfully removed $RemovedCount directories and files"
    }
    else {
        Write-Warning "No directories or files were removed"
    }
}

function Main {
    # Handle help
    if ($Help) {
        Show-Help
        exit 0
    }
  
    Write-Banner "Plaesy Clean" "Remove plaesy framework directories"
  
    # Resolve target directory to absolute path
    $AbsoluteTargetDir = Resolve-Path $TargetDir -ErrorAction SilentlyContinue
    if (!$AbsoluteTargetDir) {
        $AbsoluteTargetDir = $TargetDir
    }
  
    # Validate environment
    Test-CleanEnvironment $AbsoluteTargetDir
  
    # Show what will be cleaned
    $CleanupResults = Show-CleanupPlan $AbsoluteTargetDir
    $FoundDirectories = $CleanupResults[0]
    $FoundFiles = $CleanupResults[1]
  
    # Confirm with user
    Confirm-Removal
  
    # Perform cleanup
    Remove-PleasyDirectories $AbsoluteTargetDir
  
    Write-Host ""
    Write-Success "🧹 Plaesy clean completed!"
    Write-Host ""
    Write-Host "📁 Cleaned directory: " -NoNewline -ForegroundColor Cyan
    Write-Host (Get-Location).Path
    Write-Host ""
    Write-Host "💡 To reinitialize the project:" -ForegroundColor Yellow
    Write-Host "   .\plaesy-init.ps1 -AI your-choice"
    Write-Host ""
}

# Run main function
Main