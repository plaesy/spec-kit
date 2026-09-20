# Test script for platform detection
# This script tests the platform-detector.ps1 functionality

param(
    [string]$TestPattern = ""
)

# Import the platform detector
$detectorPath = Join-Path $PSScriptRoot "platform-detector.ps1"

if (-not (Test-Path $detectorPath)) {
    Write-Error "Platform detector not found: $detectorPath"
    exit 1
}

# Source the detector script
. $detectorPath

Write-Host "=== Platform Detection Test ===" -ForegroundColor Cyan
Write-Host ""

if ($TestPattern) {
    Write-Host "Testing with specific pattern: $TestPattern" -ForegroundColor Yellow
    Write-Host ""

    # Create test file
    $testFile = $TestPattern
    $testDir = Split-Path $testFile -Parent

    if ($testDir -and -not (Test-Path $testDir)) {
        New-Item -ItemType Directory -Path $testDir -Force | Out-Null
        Write-Host "Created directory: $testDir" -ForegroundColor Green
    }

    if (-not (Test-Path $testFile)) {
        New-Item -ItemType File -Path $testFile -Force | Out-Null
        Write-Host "Created test file: $testFile" -ForegroundColor Green
    }

    # Test detection
    $detected = Get-DetectedAiPlatform
    if ($detected) {
        Write-Host "✅ Detected platform: $detected" -ForegroundColor Green
    } else {
        Write-Host "❌ No platform detected" -ForegroundColor Red
    }

    # Cleanup
    if (Test-Path $testFile) {
        Remove-Item $testFile -Force
        Write-Host "Cleaned up test file: $testFile" -ForegroundColor Gray
    }
    if ($testDir -and (Test-Path $testDir)) {
        try {
            Remove-Item $testDir -Force -Recurse
            Write-Host "Cleaned up test directory: $testDir" -ForegroundColor Gray
        } catch {
            Write-Warning "Could not clean up directory: $testDir"
        }
    }
} else {
    Write-Host "Testing current directory for existing platforms..." -ForegroundColor Yellow
    Write-Host ""

    # Test current directory
    $detected = Get-DetectedAiPlatform
    if ($detected) {
        Write-Host "✅ Detected platform: $detected" -ForegroundColor Green
    } else {
        Write-Host "ℹ️  No platform detected in current directory" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "To test with a specific pattern:" -ForegroundColor Yellow
        Write-Host ".\test-detection.ps1 -TestPattern '.claude/settings.local.json'" -ForegroundColor Gray
        Write-Host ".\test-detection.ps1 -TestPattern 'CLAUDE.md'" -ForegroundColor Gray
        Write-Host ".\test-detection.ps1 -TestPattern '.cursorrules'" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan