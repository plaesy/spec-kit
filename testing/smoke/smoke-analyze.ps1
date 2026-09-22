# Plaesy Constitution Kit Analyzer Functional Smoke Test (PowerShell)
# Validates that plaesy-analyze.ps1 produces expected output artifacts
# and that regeneration is skipped by default when nothing changed (fingerprint
# fast path is default behavior; -Force bypasses it).

$ErrorActionPreference = "Stop"

$ROOT = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$WORK = Join-Path ([System.IO.Path]::GetTempPath()) ("plaesy-analyze-smoke-" + [System.IO.Path]::GetRandomFileName())
$FIXTURE = Join-Path $WORK "project"
$FAILURES = 0

function Cleanup { Remove-Item -Recurse -Force $WORK -ErrorAction SilentlyContinue }
Cleanup
New-Item -ItemType Directory -Force -Path "$FIXTURE/src" | Out-Null
New-Item -ItemType Directory -Force -Path "$FIXTURE/docs" | Out-Null

# Create fixture files
@'
# Analyzer Smoke Fixture

See [docs/guide.md](docs/guide.md) and [src/lib.js](src/lib.js).
'@ | Out-File -FilePath "$FIXTURE/README.md" -Encoding UTF8

@'
{
  "name": "plaesy-analyze-smoke-fixture",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0"
  }
}
'@ | Out-File -FilePath "$FIXTURE/package.json" -Encoding UTF8

@'
# Guide

Imports [src/lib.js](src/lib.js) and depends on [../README.md](README.md).
'@ | Out-File -FilePath "$FIXTURE/docs/guide.md" -Encoding UTF8

@'
import { helper } from "./helper.js";
export function lib() { return helper(); }
'@ | Out-File -FilePath "$FIXTURE/src/lib.js" -Encoding UTF8

@'
export function helper() { return "ok"; }
'@ | Out-File -FilePath "$FIXTURE/src/helper.js" -Encoding UTF8

Write-Host "=== Test 1: Basic analysis produces expected artifacts ==="
& "$ROOT/scripts/powershell/plaesy-analyze.ps1" -ProjectPath $FIXTURE -NoGraph

foreach ($f in @("project.json", "project.structure.json", "overview.md")) {
    $path = Join-Path "$FIXTURE/.plaesy/analysis" $f
    if (Test-Path $path) {
        Write-Host "  PASS: $f exists"
    } else {
        Write-Host "  FAIL: $f missing"
        $FAILURES++
    }
}

# Validate project.json contains expected fields
$projectJson = Get-Content "$FIXTURE/.plaesy/analysis/project.json" -Raw
if ($projectJson -match "express") {
    Write-Host "  PASS: project.json detects express framework"
} else {
    Write-Host "  FAIL: project.json missing express framework"
    $FAILURES++
}

if ($projectJson -match "JavaScript") {
    Write-Host "  PASS: project.json detects JavaScript language"
} else {
    Write-Host "  FAIL: project.json missing JavaScript language"
    $FAILURES++
}

Write-Host ""
Write-Host "=== Test 2: default run skips regeneration when nothing changed ==="
& "$ROOT/scripts/powershell/plaesy-analyze.ps1" -ProjectPath $FIXTURE -NoGraph

$fpPath = Join-Path "$FIXTURE/.plaesy/analysis" ".analysis-fingerprint"
if (Test-Path $fpPath) {
    Write-Host "  PASS: fingerprint file created"
} else {
    Write-Host "  FAIL: fingerprint file not created"
    $FAILURES++
}

$mtimeJsonBefore = (Get-Item "$FIXTURE/.plaesy/analysis/project.json").LastWriteTime
$mtimeStructBefore = (Get-Item "$FIXTURE/.plaesy/analysis/project.structure.json").LastWriteTime

Start-Sleep -Seconds 1

# Second run (default, no switch) should skip (fingerprint matches)
& "$ROOT/scripts/powershell/plaesy-analyze.ps1" -ProjectPath $FIXTURE -NoGraph

$mtimeJsonAfter = (Get-Item "$FIXTURE/.plaesy/analysis/project.json").LastWriteTime
$mtimeStructAfter = (Get-Item "$FIXTURE/.plaesy/analysis/project.structure.json").LastWriteTime

if ($mtimeJsonBefore -eq $mtimeJsonAfter) {
    Write-Host "  PASS: project.json not regenerated (skipped correctly)"
} else {
    Write-Host "  FAIL: project.json was regenerated despite unchanged fingerprint"
    $FAILURES++
}

if ($mtimeStructBefore -eq $mtimeStructAfter) {
    Write-Host "  PASS: project.structure.json not regenerated (skipped correctly)"
} else {
    Write-Host "  FAIL: project.structure.json was regenerated despite unchanged fingerprint"
    $FAILURES++
}

Write-Host ""
Write-Host "=== Test 3: -Force overrides the default skip ==="
& "$ROOT/scripts/powershell/plaesy-analyze.ps1" -ProjectPath $FIXTURE -NoGraph -Force
$mtimeJsonForce = (Get-Item "$FIXTURE/.plaesy/analysis/project.json").LastWriteTime
if ($mtimeJsonAfter -ne $mtimeJsonForce) {
    Write-Host "  PASS: -Force regenerated despite unchanged fingerprint"
} else {
    Write-Host "  FAIL: -Force did not override the default skip"
    $FAILURES++
}

Write-Host ""
Write-Host "=== Test 4: Modifying a source file invalidates fingerprint ==="
Start-Sleep -Seconds 1
Add-Content -Path "$FIXTURE/src/lib.js" -Value "// modified"
& "$ROOT/scripts/powershell/plaesy-analyze.ps1" -ProjectPath $FIXTURE -NoGraph
$mtimeJsonMod = (Get-Item "$FIXTURE/.plaesy/analysis/project.json").LastWriteTime
if ($mtimeJsonForce -ne $mtimeJsonMod) {
    Write-Host "  PASS: source modification triggered regeneration"
} else {
    Write-Host "  FAIL: source modification did not trigger regeneration"
    $FAILURES++
}

Write-Host ""
if ($FAILURES -eq 0) {
    Write-Host "ALL ANALYZER TESTS PASSED"
    Cleanup
    exit 0
} else {
    Write-Host "$FAILURES TEST(S) FAILED"
    Cleanup
    exit 1
}