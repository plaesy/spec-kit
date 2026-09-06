[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$common = Join-Path $root 'scripts/powershell/common.ps1'
. $common

$failed = $false

function Assert-Equal {
    param([string]$Desc, [object]$Actual, [object]$Expected)
    if ("$Actual" -ceq "$Expected") {
        Write-Output "  PASS: $Desc"
    } else {
        Write-Output "  FAIL: $Desc (actual='$Actual', expected='$Expected')"
        $script:failed = $true
    }
}

Write-Output "[1/3] ConvertTo-NativePath"
Assert-Equal "Windows absolute path unchanged" (ConvertTo-NativePath 'C:\Users\me\repo') 'C:\Users\me\repo'
Assert-Equal "UNC path unchanged" (ConvertTo-NativePath '\\server\share\dir') '\\server\share\dir'
$posix = ConvertTo-NativePath '/c/Users/me/repo'
Write-Output "  INFO: '/c/Users/me/repo' -> '$posix'"
if ([string]::IsNullOrWhiteSpace($posix) -or $posix -eq '/c/Users/me/repo') {
    Write-Output "  FAIL: posix path not converted (cygpath missing?)"
    $script:failed = $true
} else {
    Write-Output "  PASS: posix path converted to native"
}

Write-Output "[2/3] Get-RepoRoot"
$repoRoot = Get-RepoRoot
if ([string]::IsNullOrWhiteSpace($repoRoot)) {
    Write-Output "  FAIL: Get-RepoRoot returned empty (not a git repo?)"
    $script:failed = $true
} else {
    Write-Output "  INFO: repo root = '$repoRoot'"
    if (-not (Test-Path $repoRoot)) {
        Write-Output "  FAIL: Get-RepoRoot path does not exist: $repoRoot"
        $script:failed = $true
    } else {
        Write-Output "  PASS: Get-RepoRoot returns existing path"
    }
}

Write-Output "[3/3] Get-CurrentBranch"
$branch = Get-CurrentBranch
if ([string]::IsNullOrWhiteSpace($branch)) {
    Write-Output "  FAIL: Get-CurrentBranch returned empty"
    $script:failed = $true
} else {
    Write-Output "  PASS: branch = '$branch'"
}

Write-Output ""
if ($failed) {
    Write-Output "POWERSHELL SMOKE FAILED"
    exit 1
} else {
    Write-Output "POWERSHELL SMOKE PASSED"
    exit 0
}
