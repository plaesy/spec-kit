# Script: create-new-idea.ps1
# Purpose: Create idea capture document from raw idea description
# Usage: ./create-new-idea.ps1 [-Json] "idea description" [branch-name]
# Part of Plaesy Spec-Kit

param(
    [switch]$Json,
    [switch]$Help,
    [Parameter(Position = 0)]
    [string]$IdeaDescription,
    [Parameter(Position = 1)]
    [string]$BranchName
)

$ErrorActionPreference = "Stop"

# Source common functions
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
. (Join-Path $scriptDir "common.ps1")

function Show-Usage {
    Write-Host "Usage: $($MyInvocation.MyCommand.Name) [-Json] `"idea description`" [branch-name]"
    Write-Host ""
    Write-Host "Create an idea capture document from a raw idea description."
    Write-Host ""
    Write-Host "Arguments:"
    Write-Host "  -Json             Output results in JSON format"
    Write-Host "  idea description  The raw idea or problem description"
    Write-Host "  branch-name       Optional: custom branch name (auto-generated if not provided)"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  $($MyInvocation.MyCommand.Name) `"Build a task management system for teams`""
    Write-Host "  $($MyInvocation.MyCommand.Name) `"Add user authentication`" user-auth"
}

if ($Help) {
    Show-Usage
    exit 0
}

if ([string]::IsNullOrWhiteSpace($IdeaDescription)) {
    Write-Host "Error: Idea description is required" -ForegroundColor Red
    Write-Host ""
    Show-Usage
    exit 1
}

try {
    $repoRoot = git rev-parse --show-toplevel 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Error-Custom "Not in a git repository"
    }
    
    # Configuration
    $ideaTemplate = Join-Path $repoRoot "templates" "idea.template.md"
    $ideasDir = Join-Path $repoRoot "ideas"
    
    # Create ideas directory if it doesn't exist
    if (-not (Test-Path $ideasDir)) {
        New-Item -ItemType Directory -Path $ideasDir -Force | Out-Null
    }
    
    # Generate branch name if not provided
    if ([string]::IsNullOrWhiteSpace($BranchName)) {
        $sanitized = $IdeaDescription -replace '[^\w\s-]', '' -replace '\s+', '-' -replace '--+', '-'
        $sanitized = $sanitized.Trim('-').ToLowerInvariant()
        if ($sanitized.Length -gt 40) {
            $sanitized = $sanitized.Substring(0, 40).TrimEnd('-')
        }
        
        # Add timestamp to make it unique
        $timestamp = Get-Date -Format "MMdd"
        $BranchName = "idea-$timestamp-$sanitized"
    }
    
    # Ensure branch name doesn't conflict
    $originalBranchName = $BranchName
    $counter = 1
    while (git branch --list $BranchName 2>$null) {
        $BranchName = "$originalBranchName-$counter"
        $counter++
    }
    
    # Create idea file name
    $ideaFileName = "$BranchName.md"
    $ideaFilePath = Join-Path $ideasDir $ideaFileName
    
    # Check if idea file already exists
    if (Test-Path $ideaFilePath) {
        Write-Error-Custom "Idea file already exists: $ideaFilePath"
    }
    
    # Create idea document content
    $currentDate = Get-Date -Format "yyyy-MM-dd"
    $currentTime = Get-Date -Format "HH:mm:ss"
    
    $ideaContent = @"
# Idea: $IdeaDescription

**Status**: Raw Idea
**Created**: $currentDate at $currentTime
**Proposed Branch**: $BranchName

## Raw Idea Description

$IdeaDescription

## Initial Thoughts

*Use this section to capture any immediate thoughts, considerations, or questions about this idea.*

## Potential Impact

- **Business Value**: TODO - What business problem does this solve?
- **Users Affected**: TODO - Who will benefit from this?
- **Technical Complexity**: TODO - High/Medium/Low initial assessment

## Next Steps

- [ ] Refine the idea description
- [ ] Identify key stakeholders
- [ ] Assess technical feasibility
- [ ] Create feature specification using /specify
- [ ] Develop implementation plan using /plan

## Related Ideas/Features

*Links to related ideas, existing features, or dependencies*

## Research Notes

*Capture any research, links, or references related to this idea*

---

*This document was generated using the Plaesy Constitutional Framework*
*To convert this idea to a feature specification, use: /specify*
"@
    
    # Write the idea document
    Set-Content -Path $ideaFilePath -Value $ideaContent -Encoding UTF8
    
    # Stage the new file if we're in a git repository
    git add $ideaFilePath 2>$null
    
    if ($Json) {
        $result = @{
            status          = "created"
            idea_file       = $ideaFilePath
            proposed_branch = $BranchName
            description     = $IdeaDescription
            created_date    = $currentDate
            created_time    = $currentTime
        }
        $result | ConvertTo-Json -Depth 2
    }
    else {
        Write-Success "Idea document created successfully"
        Write-Host ""
        Write-Host "Idea File: $ideaFilePath"
        Write-Host "Proposed Branch: $BranchName"
        Write-Host "Description: $IdeaDescription"
        Write-Host ""
        Write-Host "Next steps:"
        Write-Host "1. Review and refine the idea in: $ideaFilePath"
        Write-Host "2. When ready, create a feature branch: git checkout -b $BranchName"
        Write-Host "3. Use /specify to create detailed requirements"
        Write-Host "4. Use /plan to create implementation plan"
        Write-Host ""
        Write-Host "The idea document has been staged for commit."
    }
}
catch {
    if ($Json) {
        $error = @{
            status  = "error"
            message = $_.Exception.Message
        }
        $error | ConvertTo-Json
    }
    else {
        Write-Error-Custom "Idea creation failed: $($_.Exception.Message)"
    }
    exit 1
}