#!/usr/bin/env pwsh
[CmdletBinding()]
param([switch]$Json)
$ErrorActionPreference = 'Stop'
. "$PSScriptRoot/common.ps1"

$paths = Get-FeaturePathsEnv
if (-not (Test-FeatureBranch -Branch $paths.CURRENT_BRANCH)) { exit 1 }

if (-not (Test-Path $paths.FEATURE_DIR -PathType Container)) {
    [Console]::Error.WriteLine("ERROR: Feature directory not found: $($paths.FEATURE_DIR)")
    [Console]::Error.WriteLine("Run /specify first to create the feature structure.")
    exit 1
}
if (-not (Test-Path $paths.IMPL_PLAN -PathType Leaf)) {
    [Console]::Error.WriteLine("ERROR: plan.md not found in $($paths.FEATURE_DIR)")
    [Console]::Error.WriteLine("Run /plan first to create the plan.")
    exit 1
}

if ($Json) {
    $docs = @()
    if (Test-Path $paths.RESEARCH) { $docs += 'research.md' }
    if (Test-Path $paths.DATA_MODEL) { $docs += 'data-model.md' }
    if ((Test-Path $paths.CONTRACTS_DIR) -and (Get-ChildItem -Path $paths.CONTRACTS_DIR -ErrorAction SilentlyContinue | Select-Object -First 1)) { $docs += 'contracts/' }
    if (Test-Path $paths.QUICKSTART) { $docs += 'quickstart.md' }
    [PSCustomObject]@{ FEATURE_DIR = $paths.FEATURE_DIR; AVAILABLE_DOCS = $docs } | ConvertTo-Json -Compress
}
else {
    Write-Output "FEATURE_DIR:$($paths.FEATURE_DIR)"
    Write-Output "AVAILABLE_DOCS:"
    Write-FileCheck -Path $paths.RESEARCH -Description 'research.md'
    Write-FileCheck -Path $paths.DATA_MODEL -Description 'data-model.md'
    Write-DirCheck -Path $paths.CONTRACTS_DIR -Description 'contracts/'
    Write-FileCheck -Path $paths.QUICKSTART -Description 'quickstart.md'
}
