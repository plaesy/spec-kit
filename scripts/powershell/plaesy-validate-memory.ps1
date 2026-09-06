# Plaesy Memory Self-Containment Validator
# Scans .plaesy/memory/ for external paths and validates self-containment rule

param(
    [switch]$Help = $false
)

if ($Help) {
    Write-Host "Plaesy Memory Self-Containment Validator" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "Usage: .\plaesy-validate-memory.ps1" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Scans .plaesy/memory/ for external paths like:" -ForegroundColor Yellow
    Write-Host "  - ~/.claude/projects (global auto-memory)" -ForegroundColor Gray
    Write-Host "  - ~/.claude/plans (plan-mode files)" -ForegroundColor Gray
    Write-Host "  - /tmp/ or temp directories" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Exit codes:" -ForegroundColor Yellow
    Write-Host "  0 = No external references (✓ safe to commit)" -ForegroundColor Green
    Write-Host "  1 = External references found (✗ needs fixing)" -ForegroundColor Red
    exit 0
}

# Colors
function Write-Success {
    param([string]$Message)
    Write-Host "[✓] $Message" -ForegroundColor Green
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[✗] $Message" -ForegroundColor Red
}

function Write-Warning-Custom {
    param([string]$Message)
    Write-Host "[!] $Message" -ForegroundColor Yellow
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

# Get memory directory
$MemoryDir = ".plaesy\memory"
if (-not (Test-Path $MemoryDir)) {
    Write-Error-Custom "Memory directory not found: $MemoryDir"
    exit 1
}

Write-Info "Scanning $MemoryDir for external references..."
Write-Host ""

# Track stats
$FilesChecked = 0
$ErrorsFound = 0

# Patterns to check for (external paths)
$Patterns = @(
    "~\.\\claude",
    "~/.\.claude",
    "~/.claude/projects",
    "~/.claude/plans",
    "/tmp/",
    "C:\\Users\\.*\.claude",
    "/Users/.*/.claude",
    "CLAUDE_CODE_DIR"
)

# Scan all markdown files in .plaesy/memory/
$MemoryFiles = Get-ChildItem -Path $MemoryDir -Filter "*.md" -Recurse -ErrorAction SilentlyContinue

foreach ($file in $MemoryFiles) {
    $FilesChecked++
    $Issues = @()

    # Read file content
    $Content = Get-Content -Path $file.FullName -Raw -ErrorAction SilentlyContinue
    if (-not $Content) { continue }

    # Check each pattern
    foreach ($pattern in $Patterns) {
        if ($Content -match $pattern) {
            $Issues += $pattern
        }
    }

    # Report if issues found
    if ($Issues.Count -gt 0) {
        $ErrorsFound++
        Write-Error-Custom "External references in: $($file.FullName)"

        # Show specific lines with issues
        foreach ($pattern in $Issues) {
            Write-Host "  Lines with '$pattern':" -ForegroundColor Yellow
            $LineNumber = 0
            foreach ($line in ($Content -split "`n")) {
                $LineNumber++
                if ($line -match $pattern) {
                    Write-Host "    $($LineNumber): $line" -ForegroundColor Gray
                }
            }
        }
        Write-Host ""
    }
}

Write-Host ""
Write-Info "Scan complete:"
Write-Host "  Files checked: $FilesChecked"
Write-Host "  External refs found: $ErrorsFound"
Write-Host ""

if ($ErrorsFound -eq 0) {
    Write-Success "Self-containment validated ✓"
    Write-Success "Memory is project-local and git-safe"
    exit 0
} else {
    Write-Warning-Custom "Self-containment issues detected"
    Write-Host ""
    Write-Host "To fix:" -ForegroundColor Yellow
    Write-Host "1. For content in ~/.claude/projects: Copy it into .plaesy/memory/" -ForegroundColor Gray
    Write-Host "2. For external links: Replace with internal .plaesy/memory/ references (flat structure, no subfolders)" -ForegroundColor Gray
    Write-Host "3. Re-run this script after fixes" -ForegroundColor Gray
    exit 1
}
