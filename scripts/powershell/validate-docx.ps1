# DOCX Structural + Render Validator
# Usage: .\validate-docx.ps1 -DocxFile <file.docx>
# 1. Verifies the file is a well-formed OOXML zip (python-docx can open it)
# 2. Checks for leftover unfilled Jinja2 placeholder tags ({{ }})
# 3. Headless-renders the document via LibreOffice to catch corruption
#    that the object model alone would not surface

param(
    [Parameter(Mandatory = $true)]
    [string]$DocxFile
)

$ErrorActionPreference = "Stop"

function Log-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Log-Success { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Log-Error { param($msg) Write-Host "[FAIL] $msg" -ForegroundColor Red }

if (-not (Test-Path $DocxFile)) {
    Log-Error "File not found: $DocxFile"
    exit 1
}

$python = Get-Command python -ErrorAction SilentlyContinue
if (-not $python) {
    $python = Get-Command python3 -ErrorAction SilentlyContinue
}
if (-not $python) {
    Log-Error "python not found on PATH — required to validate structure via python-docx"
    exit 1
}

Log-Info "Validating structure of $DocxFile"

$pyScript = @'
import re
import sys
from docx import Document

path = sys.argv[1]
doc = Document(path)

paragraph_count = len(doc.paragraphs)
table_count = len(doc.tables)
print(f"[INFO] paragraphs: {paragraph_count}, tables: {table_count}")

if paragraph_count == 0 and table_count == 0:
    raise AssertionError("document has no paragraphs or tables")

leftover = [p.text for p in doc.paragraphs if re.search(r"\{\{.*?\}\}", p.text)]
if leftover:
    raise AssertionError(f"leftover unfilled template tags found: {leftover}")

print("[OK] structural validation passed")
'@

$tmpScript = Join-Path ([System.IO.Path]::GetTempPath()) "validate_docx_$([guid]::NewGuid()).py"
Set-Content -Path $tmpScript -Value $pyScript -NoNewline

try {
    & $python.Source $tmpScript $DocxFile
    if ($LASTEXITCODE -ne 0) {
        Log-Error "Structural validation failed"
        exit 1
    }
} finally {
    Remove-Item -Path $tmpScript -Force -Confirm:$false -ErrorAction SilentlyContinue
}

$soffice = Get-Command soffice -ErrorAction SilentlyContinue
if ($soffice) {
    Log-Info "Headless-rendering document via LibreOffice to catch corruption..."
    $outDir = Join-Path ([System.IO.Path]::GetTempPath()) "docx_render_$([guid]::NewGuid())"
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null
    & $soffice.Source --headless --convert-to pdf --outdir $outDir $DocxFile | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Log-Error "LibreOffice failed to render $DocxFile — file may be corrupt"
        exit 1
    }
    Log-Success "Rendered without error (output in $outDir)"
} else {
    Log-Info "LibreOffice (soffice) not found on PATH — skipping render check. Install LibreOffice for full CI validation."
}

Log-Success "$DocxFile passed validation"
