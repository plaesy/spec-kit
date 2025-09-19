# Validate llms.txt file format and content
# Usage: ./validate-llms.ps1

param(
    [switch]$Help
)

if ($Help) {
    Write-Host "Usage: $($MyInvocation.MyCommand.Name)"
    Write-Host "Validates llms.txt file format and checks for excluded internal paths"
    exit 0
}

$ErrorActionPreference = "Stop"

# Source common functions
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
. (Join-Path $scriptDir "common.ps1")

try {
    $rootDir = Split-Path -Parent $scriptDir
    $llmsFile = Join-Path $rootDir "llms.txt"
    
    Write-Banner "Validate llms.txt" "Check links and exclusion rules"
    
    # Validate not running as admin
    Test-NotRoot
    
    if (-not (Test-Path $llmsFile)) {
        Write-Error-Custom "llms.txt not found at repository root."
    }
    
    Write-Step "Checking format sections and last updated line"
    
    $content = Get-Content -Path $llmsFile -Raw -Encoding UTF8
    
    # Check for H1 header
    if ($content -notmatch '^# ') {
        Write-Error-Custom "Missing H1 header."
    }
    
    # Check for blockquote summary (optional)
    if ($content -notmatch '^> ') {
        Write-Warning "Missing blockquote summary (optional)."
    }
    
    # Check for last updated line
    if ($content -notmatch '^Last updated: \d{4}-\d{2}-\d{2}$') {
        Write-Warning "Missing or invalid 'Last updated' line."
    }
    
    Write-Step "Ensuring no links to excluded internal paths"
    
    # Check for excluded internal paths
    $excludedPatterns = @(
        '\]\(\.plaesy',
        '\]\(\.github/copilot-instructions\.md',
        '\]\(\.github/chatmodes/',
        '\]\(\.github/commands/'
    )
    
    $violations = @()
    $lines = $content -split "`n"
    
    for ($i = 0; $i -lt $lines.Count; $i++) {
        $line = $lines[$i]
        foreach ($pattern in $excludedPatterns) {
            if ($line -match $pattern) {
                $violations += @{
                    Line = $i + 1
                    Content = $line.Trim()
                    Pattern = $pattern
                }
            }
        }
    }
    
    if ($violations.Count -gt 0) {
        Write-Host "Found links to excluded internal paths in llms.txt:" -ForegroundColor Red
        foreach ($violation in $violations) {
            Write-Host "  Line $($violation.Line): $($violation.Content)" -ForegroundColor Yellow
        }
        Write-Error-Custom "Found links to excluded internal paths in llms.txt."
    }
    
    Write-Step "Checking file encoding and line endings"
    
    # Check for any obviously binary content or encoding issues
    $bytes = [System.IO.File]::ReadAllBytes($llmsFile)
    $binaryCount = ($bytes | Where-Object { $_ -eq 0 }).Count
    
    if ($binaryCount -gt 0) {
        Write-Warning "File may contain binary content or encoding issues."
    }
    
    Write-Step "Validating markdown structure"
    
    # Count headers and ensure reasonable structure
    $headerCount = ($content | Select-String -Pattern '^#+ ' -AllMatches).Matches.Count
    if ($headerCount -eq 0) {
        Write-Warning "No markdown headers found. Consider adding structure to the document."
    }
    
    # Check for broken markdown links
    $brokenLinks = $content | Select-String -Pattern '\]\([^)]*\s[^)]*\)' -AllMatches
    if ($brokenLinks.Matches.Count -gt 0) {
        Write-Warning "Found potential broken markdown links with spaces in URLs."
        foreach ($match in $brokenLinks.Matches) {
            Write-Host "  Potential issue: $($match.Value)" -ForegroundColor Yellow
        }
    }
    
    Write-Success "llms.txt validation completed successfully"
    Write-Host ""
    Write-Host "Validation summary:"
    Write-Host "- File size: $([math]::Round((Get-Item $llmsFile).Length / 1KB, 2)) KB"
    Write-Host "- Line count: $($lines.Count)"
    Write-Host "- Header count: $headerCount"
    Write-Host "- No excluded paths found"
    Write-Host ""
    Write-Host "✅ llms.txt is valid and follows project guidelines"
}
catch {
    Write-Error-Custom "llms.txt validation failed: $($_.Exception.Message)"
}