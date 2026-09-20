# XLSX Structural + Render Validator
# Usage: .\validate-xlsx.ps1 -XlsxFile <file.xlsx> [-ExpectedSheets <string[]>]
# 1. Verifies the file is a well-formed OOXML zip (openpyxl can open it)
# 2. Headless-renders the workbook via LibreOffice to catch corruption
#    that the object model alone would not surface

param(
    [Parameter(Mandatory = $true)]
    [string]$XlsxFile,

    [Parameter(Mandatory = $false)]
    [string[]]$ExpectedSheets = @()
)

$ErrorActionPreference = "Stop"

function Log-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Log-Success { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Log-Error { param($msg) Write-Host "[FAIL] $msg" -ForegroundColor Red }

if (-not (Test-Path $XlsxFile)) {
    Log-Error "File not found: $XlsxFile"
    exit 1
}

$python = Get-Command python -ErrorAction SilentlyContinue
if (-not $python) {
    $python = Get-Command python3 -ErrorAction SilentlyContinue
}
if (-not $python) {
    Log-Error "python not found on PATH — required to validate structure via openpyxl"
    exit 1
}

Log-Info "Validating structure of $XlsxFile"

$pyScript = @'
import sys
from openpyxl import load_workbook

path = sys.argv[1]
expected_sheets = sys.argv[2:]

wb = load_workbook(path, read_only=True)
sheet_names = wb.sheetnames
print(f"[INFO] sheets: {sheet_names}")

if not sheet_names:
    raise AssertionError("workbook has no sheets")

for name in expected_sheets:
    assert name in sheet_names, f"expected sheet '{name}' not found in {sheet_names}"

for name in sheet_names:
    ws = wb[name]
    _ = ws.max_row, ws.max_column

print("[OK] structural validation passed")
'@

$tmpScript = Join-Path ([System.IO.Path]::GetTempPath()) "validate_xlsx_$([guid]::NewGuid()).py"
Set-Content -Path $tmpScript -Value $pyScript -NoNewline

try {
    & $python.Source $tmpScript $XlsxFile @ExpectedSheets
    if ($LASTEXITCODE -ne 0) {
        Log-Error "Structural validation failed"
        exit 1
    }
} finally {
    Remove-Item -Path $tmpScript -Force -Confirm:$false -ErrorAction SilentlyContinue
}

$soffice = Get-Command soffice -ErrorAction SilentlyContinue
if ($soffice) {
    Log-Info "Headless-rendering workbook via LibreOffice to catch corruption..."
    $outDir = Join-Path ([System.IO.Path]::GetTempPath()) "xlsx_render_$([guid]::NewGuid())"
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null
    & $soffice.Source --headless --convert-to pdf --outdir $outDir $XlsxFile | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Log-Error "LibreOffice failed to render $XlsxFile — file may be corrupt"
        exit 1
    }
    Log-Success "Rendered without error (output in $outDir)"
} else {
    Log-Info "LibreOffice (soffice) not found on PATH — skipping render check. Install LibreOffice for full CI validation."
}

Log-Success "$XlsxFile passed validation"
