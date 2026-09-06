# Common functions for plaesy scripts (PowerShell)
# Version: 1.0.0 - Compatible with PowerShell 5.1+
# Constitutional Development Framework

# Script metadata
if (-not $env:PLAESY_VERSION) {
    # Get version from VERSION file in repository root
    function Get-PlaesyVersion {
        $versionFile = ""
        $scriptDir = $PSScriptRoot

        # Try to find VERSION file in multiple locations
        $versionLocations = @(
            ".\VERSION",
            "$scriptDir\..\..\VERSION",
            (Join-Path (Split-Path -Parent $scriptDir) "VERSION"),
            "$env:USERPROFILE\.plaesy\VERSION"
        )

        foreach ($location in $versionLocations) {
            if (Test-Path $location) {
                $versionFile = $location
                break
            }
        }

        if (-not $versionFile) {
            return "0.0.0" # Fallback version
        }

        # Read and sanitize version
        try {
            $version = (Get-Content $versionFile -ErrorAction Stop | Select-Object -First 1).Trim()
            if ($version -match '^\d+\.\d+\.\d+$') {
                return $version
            } else {
                return "0.0.0" # Fallback for invalid format
            }
        }
        catch {
            return "0.0.0" # Fallback for read errors
        }
    }

    $env:PLAESY_VERSION = Get-PlaesyVersion
}

if (-not $env:SCRIPT_START_TIME) {
    $env:SCRIPT_START_TIME = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
}

# Enhanced logging functions with timestamps and levels
function Write-LogMessage {
    param(
        [string]$Message,
        [ValidateSet("INFO", "WARNING", "ERROR", "SUCCESS")]
        [string]$Level = "INFO"
    )

    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $scriptName = Split-Path -Leaf $MyInvocation.PSCommandPath

    # Console output with colors
    switch ($Level) {
        "INFO" { Write-Host "[$timestamp] [$Level] [$scriptName] $Message" -ForegroundColor Cyan }
        "WARNING" { Write-Host "[$timestamp] [$Level] [$scriptName] $Message" -ForegroundColor Yellow }
        "SUCCESS" { Write-Host "[$timestamp] [$Level] [$scriptName] $Message" -ForegroundColor Green }
        "ERROR" { Write-Host "[$timestamp] [$Level] [$scriptName] $Message" -ForegroundColor Red }
    }
}

# Enhanced logging functions
function Write-Step {
    param([string]$Message)
    Write-Host "[STEP] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning-Custom {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

function Log-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

function Log-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Log-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Log-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

# Banner function with ASCII characters (PowerShell compatible)
function Write-Banner {
    param([string]$Title, [string]$Subtitle)
    Write-Host "##### PLAESY SPEC-KIT #####" -ForegroundColor Magenta
    Write-Host "Constitutional Development Framework" -ForegroundColor Magenta
    Write-Host "Version: $env:PLAESY_VERSION" -ForegroundColor Gray
    Write-Host ""
    if ($Title) {
        Write-Host "$Title" -ForegroundColor Magenta
    }
    if ($Subtitle) {
        Write-Host "   $Subtitle" -ForegroundColor Magenta
    }
    Write-Host ""
}

# Validation functions
function Test-NotRoot {
    # In PowerShell, check if running as administrator
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return -not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-CommandExists {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

function Test-FileExists {
    param([string]$Path)
    return Test-Path $Path -PathType Leaf
}

function Test-DirectoryExists {
    param([string]$Path)
    return Test-Path $Path -PathType Container
}

# Git/feature-path helpers (PowerShell equivalents of common.sh's
# get_repo_root/get_current_branch/get_feature_dir/get_feature_paths/check_feature_branch)

# Normalize a POSIX-style path (from git-bash/MSYS git) to a native Windows
# path. Without this, `git rev-parse --show-toplevel` under MSYS returns
# `/c/Users/...` or `/tmp/...`, and Join-Path turns that into `\tmp\...` -
# an invalid Windows path with no drive letter that Test-Path rejects.
function ConvertTo-NativePath {
    param([string]$Path)
    if ([string]::IsNullOrWhiteSpace($Path)) { return $Path }
    if ($env:OS -eq 'Windows_NT' -and $Path -notmatch '^[A-Za-z]:[\\/]') {
        $cygpath = Get-Command cygpath -ErrorAction SilentlyContinue
        if ($cygpath) {
            $converted = & $cygpath.Source -w $Path 2>$null
            if ($converted) { return $converted.Trim() }
        }
        # Fallback: resolve against the current location's root drive.
        return [System.IO.Path]::GetFullPath($Path)
    }
    return $Path
}

function Get-RepoRoot {
    $root = git rev-parse --show-toplevel
    if ($root) { return (ConvertTo-NativePath ($root.Trim())) }
    return $root
}

function Get-CurrentBranch {
    $branch = git rev-parse --abbrev-ref HEAD
    if ($branch) { return $branch.Trim() }
    return $branch
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
        FEATURE_SPEC   = Join-Path $featureDir "spec.md"
        IMPL_PLAN      = Join-Path $featureDir "plan.md"
        TASKS          = Join-Path $featureDir "tasks.md"
        RESEARCH       = Join-Path $featureDir "research.md"
        DATA_MODEL     = Join-Path $featureDir "data-model.md"
        QUICKSTART     = Join-Path $featureDir "quickstart.md"
        CONTRACTS_DIR  = Join-Path $featureDir "contracts"
    }
}

function Test-FeatureBranch {
    param([string]$Branch)
    if ($Branch -notmatch '^\d{3}-') {
        Write-Error "Not on a feature branch. Current branch: $Branch`nFeature branches should be named like: 001-feature-name"
        return $false
    }
    return $true
}

function Write-FileCheck {
    param([string]$Path, [string]$Description)
    if (Test-Path $Path -PathType Leaf) { Write-Output "  [OK] $Description" }
    else { Write-Output "  [MISSING] $Description" }
}

function Write-DirCheck {
    param([string]$Path, [string]$Description)
    if ((Test-Path $Path -PathType Container) -and (Get-ChildItem -Path $Path -ErrorAction SilentlyContinue | Select-Object -First 1)) {
        Write-Output "  [OK] $Description"
    }
    else { Write-Output "  [MISSING] $Description" }
}

# Utility functions
function Get-Timestamp {
    return Get-Date -Format "yyyy-MM-dd HH:mm:ss"
}

function Get-ScriptDirectory {
    return Split-Path -Parent $MyInvocation.PSCommandPath
}

function Write-SectionHeader {
    param([string]$Title)
    Write-Host ""
    Write-Host "=== $Title ===" -ForegroundColor Magenta
    Write-Host ""
}

function Write-SectionFooter {
    param([string]$Title = "Done")
    Write-Host ""
    Write-Host "=== $Title ===" -ForegroundColor Magenta
    Write-Host ""
}

# Error handling
function Invoke-ErrorHandling {
    param(
        [string]$ErrorMessage,
        [int]$ExitCode = 1,
        [switch]$ShowStackTrace
    )

    Write-LogError "Error: $ErrorMessage"
    Write-LogError "Exit code: $ExitCode"

    if ($ShowStackTrace) {
        Write-LogError "Stack trace:"
        $caller = Get-PSCallStack | Select-Object -Skip 1 -First 1
        if ($caller) {
            Write-LogError "  Command: $($caller.Command)"
            Write-LogError "  Location: $($caller.Location)"
            Write-LogError "  Line: $($caller.ScriptLineNumber)"
        }
    }

    exit $ExitCode
}

# Progress reporting
function Write-Progress {
    param(
        [string]$Activity,
        [string]$Status,
        [int]$PercentComplete = -1,
        [string]$CurrentOperation = ""
    )

    $params = @{
        Activity = $Activity
        Status = $Status
    }

    if ($PercentComplete -ge 0) {
        $params.PercentComplete = $PercentComplete
    }

    if ($CurrentOperation) {
        $params.CurrentOperation = $CurrentOperation
    }

    Write-Progress @params
}

# Color scheme constants
$Colors = @{
    Red     = "Red"
    Green   = "Green"
    Yellow  = "Yellow"
    Blue    = "Blue"
    Magenta = "Magenta"
    Cyan    = "Cyan"
    White   = "White"
    Gray    = "Gray"
}

# Logging configuration
$script:LogLevel = if ($env:PLAESY_LOG_LEVEL) { $env:PLAESY_LOG_LEVEL } else { "INFO" }
$script:LogFile = if ($env:PLAESY_LOG_FILE) { $env:PLAESY_LOG_FILE } else { "" }
$script:PlaesyDebugMode = if ($env:PLAESY_DEBUG) { $env:PLAESY_DEBUG } else { "false" }

# NOTE: this file is dot-sourced (". $PSScriptRoot/common.ps1"), not imported
# as a module, so every function above is already available in the caller's
# scope. Export-ModuleMember only works inside a .psm1 module -- calling it
# here throws ("can only be called from inside a module"), and since callers
# set $ErrorActionPreference = 'Stop' before dot-sourcing this file, that
# error was terminating every script that loads common.ps1. Do not add it back.