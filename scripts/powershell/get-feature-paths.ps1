#!/usr/bin/env pwsh
param()
$ErrorActionPreference = 'Stop'

. "$PSScriptRoot/common.ps1"

$paths = Get-FeaturePathsEnv
if (-not (Test-FeatureBranch -Branch $paths.CURRENT_BRANCH -ErrorAction SilentlyContinue)) {
    Write-Output "REPO_ROOT: $($paths.REPO_ROOT)"
    Write-Output "BRANCH: $($paths.CURRENT_BRANCH)"
    Write-Output "FEATURE_DIR: $($paths.FEATURE_DIR)"
    Write-Output "FEATURE_SPEC: $($paths.FEATURE_SPEC)"
    Write-Output "IMPL_PLAN: $($paths.IMPL_PLAN)"
    Write-Output "TASKS: $($paths.TASKS)"
    Write-Output "INFO: Not on a feature branch (format: XXX-feature-name)"
    exit 0
}

Write-Output "REPO_ROOT: $($paths.REPO_ROOT)"
Write-Output "BRANCH: $($paths.CURRENT_BRANCH)"
Write-Output "FEATURE_DIR: $($paths.FEATURE_DIR)"
Write-Output "FEATURE_SPEC: $($paths.FEATURE_SPEC)"
Write-Output "IMPL_PLAN: $($paths.IMPL_PLAN)"
Write-Output "TASKS: $($paths.TASKS)"
