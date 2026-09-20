# PPTX Structural + Render Validator
# Usage: .\validate-pptx.ps1 -PptxFile <file.pptx> [-ExpectedSlides <int>]
# 1. Verifies the file is a well-formed OOXML zip (python-pptx can open it)
# 2. Headless-renders each slide to PNG via LibreOffice to catch corruption
#    that the object model alone would not surface

param(
    [Parameter(Mandatory = $true)]
    [string]$PptxFile,

    [Parameter(Mandatory = $false)]
    [int]$ExpectedSlides = 0
)

$ErrorActionPreference = "Stop"

function Log-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Log-Success { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Log-Error { param($msg) Write-Host "[FAIL] $msg" -ForegroundColor Red }

if (-not (Test-Path $PptxFile)) {
    Log-Error "File not found: $PptxFile"
    exit 1
}

$python = Get-Command python -ErrorAction SilentlyContinue
if (-not $python) {
    $python = Get-Command python3 -ErrorAction SilentlyContinue
}
if (-not $python) {
    Log-Error "python not found on PATH — required to validate structure via python-pptx"
    exit 1
}

Log-Info "Validating structure of $PptxFile"

$pyScript = @'
import sys
from pptx import Presentation

path, expected = sys.argv[1], sys.argv[2] if len(sys.argv) > 2 else ""
prs = Presentation(path)
slide_count = len(prs.slides)
print(f"[INFO] slide count: {slide_count}")

for i, slide in enumerate(prs.slides, start=1):
    if slide.shapes.title is None:
        print(f"[WARN] slide {i} has no title placeholder")

if expected:
    assert slide_count == int(expected), f"expected {expected} slides, got {slide_count}"
print("[OK] structural validation passed")
'@

$tmpScript = Join-Path ([System.IO.Path]::GetTempPath()) "validate_pptx_$([guid]::NewGuid()).py"
Set-Content -Path $tmpScript -Value $pyScript -NoNewline

try {
    $expectedArg = if ($ExpectedSlides -gt 0) { "$ExpectedSlides" } else { "" }
    & $python.Source $tmpScript $PptxFile $expectedArg
    if ($LASTEXITCODE -ne 0) {
        Log-Error "Structural validation failed"
        exit 1
    }
} finally {
    Remove-Item -Path $tmpScript -Force -Confirm:$false -ErrorAction SilentlyContinue
}

$soffice = Get-Command soffice -ErrorAction SilentlyContinue
if ($soffice) {
    Log-Info "Headless-rendering slides via LibreOffice to catch corruption..."
    $outDir = Join-Path ([System.IO.Path]::GetTempPath()) "pptx_render_$([guid]::NewGuid())"
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null
    & $soffice.Source --headless --convert-to png --outdir $outDir $PptxFile | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Log-Error "LibreOffice failed to render $PptxFile — file may be corrupt"
        exit 1
    }
    Log-Success "Rendered without error (output in $outDir)"
} else {
    Log-Info "LibreOffice (soffice) not found on PATH — skipping render check. Install LibreOffice for full CI validation."
}

Log-Success "$PptxFile passed validation"
