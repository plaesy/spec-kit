# Simple common functions for plaesy scripts (PowerShell)
# Constitutional Development Framework

# Error handling
$ErrorActionPreference = "Stop"

# Colors for output
$Colors = @{
    Red    = "Red"
    Green  = "Green"
    Yellow = "Yellow"
    Blue   = "Blue"
    Purple = "Magenta"
    Cyan   = "Cyan"
    White  = "White"
}

# Print functions for user feedback
function Write-Step {
    param([string]$Message)
    Write-Host "[STEP] $Message" -ForegroundColor $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Red
    exit 1
}

function Write-Banner {
    param(
        [string]$Title,
        [string]$Subtitle
    )
    
    Write-Host ""
    Write-Host "██████╗ ██╗      █████╗ ███████╗███████╗██╗   ██╗" -ForegroundColor $Colors.Purple
    Write-Host "██╔══██╗██║     ██╔══██╗██╔════╝██╔════╝╝██╗ ██╔╝" -ForegroundColor $Colors.Purple
    Write-Host "██████╔╝██║     ███████║█████╗  ███████╗ ╚████╔╝ " -ForegroundColor $Colors.Purple
    Write-Host "██╔═══╝ ██║     ██╔══██║██╔══╝  ╚════██║  ╚██╔╝  " -ForegroundColor $Colors.Purple
    Write-Host "██║     ███████╗██║  ██║███████╗███████║   ██║   " -ForegroundColor $Colors.Purple
    Write-Host "╚═╝     ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   " -ForegroundColor $Colors.Purple
    Write-Host ""
    
    if ($Title) {
        Write-Host "$Title" -ForegroundColor $Colors.Purple
    }
    if ($Subtitle) {
        Write-Host "   $Subtitle" -ForegroundColor $Colors.Purple
    }
    Write-Host ""
}

# Validation functions
function Test-NotRoot {
    # In PowerShell, check if running as administrator
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    $isAdmin = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    
    if ($isAdmin) {
        Write-Error-Custom "This script should not be run as administrator for security reasons."
    }
}

function Test-DiskSpace {
    param(
        [int]$RequiredMB,
        [string]$Path = $env:USERPROFILE
    )
    
    try {
        $drive = (Get-Item $Path).PSDrive
        $availableSpaceGB = [math]::Round($drive.Free / 1GB, 2)
        $availableSpaceMB = [math]::Round($drive.Free / 1MB, 2)
        $requiredGB = [math]::Round($RequiredMB / 1024, 2)
        
        if ($availableSpaceMB -lt $RequiredMB) {
            Write-Error-Custom "Insufficient disk space. At least ${RequiredMB}MB (${requiredGB}GB) required, but only ${availableSpaceMB}MB (${availableSpaceGB}GB) available."
        }
    }
    catch {
        Write-Warning "Could not check disk space for path: $Path"
    }
}

function Get-RepoRoot {
    git rev-parse --show-toplevel
}

function Get-CurrentBranch {
    git rev-parse --abbrev-ref HEAD
}

function Test-FeatureBranch {
    param([string]$Branch)
    if ($Branch -notmatch '^[0-9]{3}-') {
        Write-Output "ERROR: Not on a feature branch. Current branch: $Branch"
        Write-Output "Feature branches should be named like: 001-feature-name"
        return $false
    }
    return $true
}

function Get-FeatureDir {
    param([string]$RepoRoot, [string]$Branch)
    Join-Path $RepoRoot "specs/$Branch"
}

function Get-FeaturePathsEnv {
    $repoRoot = Get-RepoRoot
    $currentBranch = Get-CurrentBranch
    $featureDir = Get-FeatureDir -RepoRoot $repoRoot -Branch $currentBranch
    [PSCustomObject]@{
        REPO_ROOT      = $repoRoot
        CURRENT_BRANCH = $currentBranch
        FEATURE_DIR    = $featureDir
        FEATURE_SPEC   = Join-Path $featureDir 'spec.md'
        IMPL_PLAN      = Join-Path $featureDir 'plan.md'
        TASKS          = Join-Path $featureDir 'tasks.md'
        RESEARCH       = Join-Path $featureDir 'research.md'
        DATA_MODEL     = Join-Path $featureDir 'data-model.md'
        QUICKSTART     = Join-Path $featureDir 'quickstart.md'
        CONTRACTS_DIR  = Join-Path $featureDir 'contracts'
    }
}

function Test-FileExists {
    param([string]$Path, [string]$Description)
    if (Test-Path -Path $Path -PathType Leaf) {
        Write-Output "  ✓ $Description"
        return $true
    }
    else {
        Write-Output "  ✗ $Description"
        return $false
    }
}

function Test-DirHasFiles {
    param([string]$Path, [string]$Description)
    if ((Test-Path -Path $Path -PathType Container) -and (Get-ChildItem -Path $Path -ErrorAction SilentlyContinue | Where-Object { -not $_.PSIsContainer } | Select-Object -First 1)) {
        Write-Output "  ✓ $Description"
        return $true
    }
    else {
        Write-Output "  ✗ $Description"
        return $false
    }
}

# Export functions for module usage
Export-ModuleMember -Function @(
    'Write-Step',
    'Write-Success', 
    'Write-Warning',
    'Write-Error-Custom',
    'Write-Banner',
    'Test-NotRoot',
    'Test-DiskSpace',
    'Get-FeaturePaths',
    'Test-FeatureBranch'
)