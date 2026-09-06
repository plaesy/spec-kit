# AI-Specific Header Injection Script for PowerShell
# Injects platform-specific headers into prompt files based on chosen AI platform

param(
    [Parameter(Mandatory = $true, HelpMessage = "AI platform to configure headers for")]
    [ValidateSet("copilot", "github_copilot", "cursor", "cursor_ai", "windsurf", "windsurf_ai", "claude", "claude_code", "chatgpt", "gemini", "trae-ai", "trae_ai", "qwen-code", "qwen_code", "codex-cli", "opencode-cli", "local-ai", "manual")]
    [string]$AiPlatform,

    [Parameter(Mandatory = $true, HelpMessage = "Target directory containing prompt files")]
    [string]$TargetDirectory,

    [Parameter(HelpMessage = "Show what would be changed without making changes")]
    [switch]$DryRun,

    [Parameter(HelpMessage = "Overwrite existing headers without confirmation")]
    [switch]$Force,

    [Parameter(HelpMessage = "Create backup of original files")]
    [switch]$Backup,

    [Parameter(HelpMessage = "Merge header keys into existing front-matter (add missing keys)")]
    [switch]$Merge,

    [Parameter(HelpMessage = "Only list files and chosen headers (no changes)")]
    [switch]$ListOnly,

    [Parameter(HelpMessage = "Additional include patterns (can be repeated)")]
    [string[]]$Pattern = @(),

    [Parameter(HelpMessage = "Show help message")]
    [switch]$Help
)

# Script configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$RepoRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$HeadersDir = Join-Path $RepoRoot "templates\ai-headers"
$DefaultHeader = "manual.header.yaml"

# Color configuration for output
$Colors = @{
    Info    = "Cyan"
    Success = "Green"
    Warning = "Yellow"
    Error   = "Red"
}

# Logging functions
function Write-LogInfo {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Info
}

function Write-LogSuccess {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Success
}

function Write-LogWarning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Warning
}

function Write-LogError {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Error
}

# Help function
function Show-Help {
    @"
AI Header Injection Script for PowerShell

USAGE:
    .\inject-ai-headers.ps1 -AiPlatform <platform> -TargetDirectory <directory> [OPTIONS]

PARAMETERS:
    -AiPlatform <platform>     AI platform to configure headers for
    -TargetDirectory <dir>     Target directory containing prompt files
    -DryRun                    Show what would be changed without making changes
    -Force                     Overwrite existing headers without confirmation
    -Backup                    Create backup of original files
    -Merge                     Merge header keys into existing front-matter (add missing keys)
    -Help                      Show this help message

SUPPORTED AI PLATFORMS:
    copilot              GitHub Copilot
    github_copilot       GitHub Copilot (Framework)
    cursor               Cursor AI
    cursor_ai            Cursor AI (Framework)
    windsurf             Windsurf AI
    windsurf_ai          Windsurf AI (Framework)
    claude               Claude Code
    claude_code          Claude Code (Framework)
    chatgpt              ChatGPT
    gemini               Google Gemini
    trae-ai              Trae.ai Multi-Agent
    trae_ai              Trae.ai Multi-Agent (Framework)
    qwen-code            Qwen Code
    qwen_code            Qwen Code (Framework)
    codex-cli            Codex CLI
    opencode-cli         OpenCode CLI
    local-ai             Local AI Models (Ollama, LM Studio)
    manual               Manual Development (no AI)

EXAMPLES:
    .\inject-ai-headers.ps1 -AiPlatform cursor -TargetDirectory .\prompts
    .\inject-ai-headers.ps1 -AiPlatform claude -TargetDirectory . -Backup -DryRun
    .\inject-ai-headers.ps1 -AiPlatform copilot -TargetDirectory .\my-project -Force

"@
}

# Merge header front-matter keys into existing file front-matter.
function Merge-FrontMatter {
    param(
        [string]$FilePath,
        [string]$HeaderContent,
        [bool]$BackupFlag,
        [bool]$DryRunFlag
    )

    $content = Get-Content -Raw -Path $FilePath -ErrorAction Stop

    $fmRegex = [regex]::new('(?ms)\A---\r?\n(.*?)\r?\n---')
    $existingMatch = $fmRegex.Match($content)
    if (-not $existingMatch.Success) {
        Write-LogWarning "No existing front-matter found for $(Split-Path $FilePath -Leaf)"
        return $false
    }

    $existingFm = $existingMatch.Groups[1].Value.TrimEnd("`r", "`n")

    $headerMatch = $fmRegex.Match($HeaderContent)
    if (-not $headerMatch.Success) {
        Write-LogWarning "No front-matter found in header content for merging"
        return $false
    }
    $headerFm = $headerMatch.Groups[1].Value.TrimEnd("`r", "`n")

    # Parse simple key: value pairs (flat YAML only)
    $existingLines = $existingFm -split "\r?\n" | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne '' }
    $headerLines = $headerFm -split "\r?\n" | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne '' }

    $existingKeys = @{}
    $existingOrder = @()
    foreach ($line in $existingLines) {
        if ($line -match '^([^:]+):\s*(.*)$') {
            $k = $matches[1].Trim()
            $v = $matches[2].Trim()
            $existingKeys[$k] = $v
            $existingOrder += $k
        }
    }

    $headerKeys = @{}
    $headerOrder = @()
    foreach ($line in $headerLines) {
        if ($line -match '^([^:]+):\s*(.*)$') {
            $k = $matches[1].Trim()
            $v = $matches[2].Trim()
            $headerKeys[$k] = $v
            $headerOrder += $k
        }
    }

    # Build merged fm: preserve existing order, then append header-only keys
    $merged = New-Object System.Collections.Generic.List[System.String]
    foreach ($k in $existingOrder) { $merged.Add($k + ': ' + $existingKeys[$k]) }
    foreach ($k in $headerOrder) {
        if (-not $existingKeys.ContainsKey($k)) { $merged.Add($k + ': ' + $headerKeys[$k]) }
    }

    $mergedText = $merged -join "`n"

    if ($DryRunFlag) {
        Write-LogInfo "[DRY RUN] Would merge front-matter into $(Split-Path $FilePath -Leaf):`n$mergedText"
        return $true
    }

    if ($BackupFlag) { New-BackupFile $FilePath | Out-Null }

    $restStart = $existingMatch.Index + $existingMatch.Length
    $rest = $content.Substring($restStart)

    $newContent = "---`n$mergedText`n---" + $rest
    Set-Content -Path $FilePath -Value $newContent
    Write-LogSuccess "Merged front-matter into $(Split-Path $FilePath -Leaf)"
    return $true
}

# Get header file path (supports per-type headers)
function Get-HeaderFile {
    param(
        [string]$Platform,
        [string]$Type
    )

    if ($Type) {
        $candidate = Join-Path $HeadersDir "$($Platform).$($Type).yaml"
        if (Test-Path $candidate) { return $candidate }
    }

    # Special handling for framework platforms -> fallback to base platforms
    $fallbackMappings = @{
        "claude_code" = "claude"
        "github_copilot" = "copilot"
        "cursor_ai" = "cursor"
        "windsurf_ai" = "windsurf"
        "trae_ai" = "trae-ai"
        "qwen_code" = "qwen-code"
    }

    if ($fallbackMappings.ContainsKey($Platform)) {
        $fallbackHeader = Join-Path $HeadersDir "$($fallbackMappings[$Platform]).header.yaml"
        if (Test-Path $fallbackHeader) {
            Write-LogInfo "Using $($fallbackMappings[$Platform]) header for $Platform platform"
            return $fallbackHeader
        }
    }

    # Try .yaml extension first
    $platformHeader = Join-Path $HeadersDir "$($Platform).header.yaml"
    if (Test-Path $platformHeader) { return $platformHeader }

    # Fallback to .md extension
    $platformHeaderMd = Join-Path $HeadersDir "$($Platform).header.md"
    if (Test-Path $platformHeaderMd) { return $platformHeaderMd }

    if ([string]::IsNullOrEmpty($Type)) {
        Write-LogWarning "No specific header for $Platform, using default header"
    }
    else {
        Write-LogWarning "No specific header for $Platform (type: $Type), using default header"
    }
    return Join-Path $HeadersDir $DefaultHeader
}

# Extract the `header_content: |` block scalar from a header YAML file
# (see templates/ai-headers/README.md for the format). Falls back to the
# raw file content if it doesn't look like that format.
function Get-YamlHeaderTemplate {
    param([string]$FilePath)

    $lines = Get-Content -Path $FilePath
    $startIndex = -1
    for ($i = 0; $i -lt $lines.Length; $i++) {
        if ($lines[$i] -match '^header_content:\s*\|') { $startIndex = $i + 1; break }
    }
    if ($startIndex -lt 0) {
        return (Get-Content -Path $FilePath -Raw)
    }

    $bodyLines = $lines[$startIndex..($lines.Length - 1)]
    $indent = $null
    $result = New-Object System.Collections.Generic.List[string]
    foreach ($line in $bodyLines) {
        if ($null -eq $indent) {
            if ($line.Trim() -eq '') { continue }
            $indent = ($line.Length - $line.TrimStart(' ').Length)
        }
        if ($line.Length -ge $indent) { $result.Add($line.Substring($indent)) }
        else { $result.Add('') }
    }
    # Trim trailing blank lines from the block scalar
    while ($result.Count -gt 0 -and $result[$result.Count - 1].Trim() -eq '') { $result.RemoveAt($result.Count - 1) }
    return ($result -join "`n")
}

# Extract {{DESCRIPTION}} for a target file: its own front-matter `description:`
# if present, otherwise a filename-based default matching the bash script's
# convention ("security.chatmode.md" -> "Chat mode: security").
function Get-FileDescription {
    param([string]$FilePath)

    $content = Get-Content -Path $FilePath -Raw -ErrorAction SilentlyContinue
    if ($content -and $content -match '(?ms)\A---\r?\n.*?^description:\s*"?([^"\r\n]*)"?\s*$') {
        $desc = $matches[1].Trim()
        if ($desc) { return $desc }
    }

    $name = [System.IO.Path]::GetFileName($FilePath)
    if ($name -match '^(?<base>.+)\.prompt\.md$') { return "Prompt: $($matches['base'])" }
    if ($name -match '^(?<base>.+)\.chatmode\.md$') { return "Chat mode: $($matches['base'])" }
    if ($name -match '^(?<base>.+)\.instructions?\.md$') { return "Instructions: $($matches['base'])" }
    return [System.IO.Path]::GetFileNameWithoutExtension($name)
}

# Find prompt files
function Find-PromptFiles {
    param([string]$TargetDir)

    if (-not (Test-Path $TargetDir)) {
        Write-LogError "Target directory does not exist: $TargetDir"
        return @()
    }

    $patterns = if ($Pattern.Count -gt 0) { $Pattern } else { @('\.prompt\.md$', '\.chatmode\.md$', '\.instruction\.md$') }

    return Get-ChildItem -Path $TargetDir -File -Recurse | Where-Object {
        foreach ($pat in $patterns) { if ($_.Name -match $pat) { return $true } }
        return $false
    }
}

# Check if file already has header
function Test-HasHeader {
    param([string]$FilePath)

    $FirstLines = Get-Content $FilePath -TotalCount 20 -ErrorAction SilentlyContinue

    if ($FirstLines) {
        $HeaderIndicators = @("^---$", "Constitutional.*Framework", "AI.*Configuration", "# .*Header")

        foreach ($Line in $FirstLines) {
            foreach ($Pattern in $HeaderIndicators) {
                if ($Line -match $Pattern) {
                    return $true
                }
            }
        }
    }

    return $false
}

# Create backup of file
function New-BackupFile {
    param([string]$FilePath)

    $BackupPath = "$FilePath.backup.$(Get-Date -Format 'yyyyMMdd_HHmmss')"
    Copy-Item $FilePath $BackupPath
    Write-LogInfo "Created backup: $(Split-Path $BackupPath -Leaf)"
    return $BackupPath
}

# Remove existing header from file
function Remove-ExistingHeader {
    param([string]$FilePath)

    $Content = Get-Content $FilePath
    $HeaderEndLine = -1
    $HeaderCount = 0

    for ($i = 0; $i -lt $Content.Length; $i++) {
        if ($Content[$i] -eq "---") {
            $HeaderCount++
            if ($HeaderCount -eq 2) {
                $HeaderEndLine = $i
                break
            }
        }
    }

    if ($HeaderEndLine -ge 0) {
        $NewContent = $Content[($HeaderEndLine + 2)..($Content.Length - 1)]
        Set-Content -Path $FilePath -Value $NewContent
        Write-LogInfo "Removed existing header from $(Split-Path $FilePath -Leaf)"
    }
}

# Inject header into file
function Add-HeaderToFile {
    param(
        [string]$FilePath,
        [string]$HeaderContent
    )

    $OriginalContent = Get-Content $FilePath -Raw
    $NewContent = $HeaderContent + "`n`n" + $OriginalContent
    Set-Content -Path $FilePath -Value $NewContent
}

# Process single file
function Invoke-ProcessFile {
    param(
        [System.IO.FileInfo]$File,
        [string]$HeaderContent,
        [bool]$Force,
        [bool]$Backup,
        [bool]$DryRun,
        [bool]$Merge
    )

    $FileName = $File.Name
    $FilePath = $File.FullName

    if ($Merge) {
        return Merge-FrontMatter -FilePath $FilePath -HeaderContent $HeaderContent -BackupFlag $Backup -DryRunFlag $DryRun
    }

    if ($DryRun) {
        if ((Test-HasHeader $FilePath) -and -not $Force) {
            Write-LogInfo "[DRY RUN] Would skip $FileName (already has header, use -Force to overwrite)"
        }
        else {
            Write-LogInfo "[DRY RUN] Would inject header into $FileName"
        }
        return $true
    }

    # Check if file already has header
    if ((Test-HasHeader $FilePath) -and -not $Force) {
        Write-LogWarning "Skipping $FileName (already has header, use -Force to overwrite)"
        return $false
    }

    # Create backup if requested
    if ($Backup) {
        New-BackupFile $FilePath | Out-Null
    }

    # Remove existing header if forcing update
    if ($Force -and (Test-HasHeader $FilePath)) {
        Remove-ExistingHeader $FilePath
    }

    # Inject new header
    Add-HeaderToFile $FilePath $HeaderContent
    Write-LogSuccess "Injected header into $FileName"
    return $true
}

# Main execution
function Invoke-Main {
    # Show help if requested
    if ($Help) {
        Show-Help
        return
    }

    Write-LogInfo "Starting header injection for AI platform: $AiPlatform"

    # Get header file
    $HeaderFile = Get-HeaderFile $AiPlatform

    if (-not (Test-Path $HeaderFile)) {
        Write-LogError "Header file not found: $HeaderFile"
        return
    }

    # Load header content
    $HeaderContent = Get-Content $HeaderFile -Raw
    Write-LogInfo "Using header file: $(Split-Path $HeaderFile -Leaf)"

    # Find prompt files
    $PromptFiles = Find-PromptFiles $TargetDirectory

    if ($PromptFiles.Count -eq 0) {
        Write-LogWarning "No prompt files found in $TargetDirectory"
        return
    }

    Write-LogInfo "Found $($PromptFiles.Count) prompt file(s)"

    # If ListOnly, print mapping and exit
    if ($ListOnly) {
        foreach ($File in $PromptFiles) {
            $relpath = $File.FullName.Substring($TargetDirectory.Length).TrimStart('\', '/')
            if ($relpath -match '(^|/)prompts?(/|$)' -or $File.Name -match '\.prompt\.md$') { $ftype = 'prompts' }
            elseif ($relpath -match '(^|/)chatmodes?(/|$)' -or $File.Name -match '\.chatmode\.md$' -or $File.Name -match 'chatmode') { $ftype = 'chatmodes' }
            elseif ($relpath -match '(^|/)instructions?(/|$)' -or $File.Name -match '\.instruction\.md$' -or $File.Name -match 'instructions') { $ftype = 'instructions' }
            else { $ftype = 'generic' }
            $headerFileForType = Get-HeaderFile $AiPlatform $ftype
            Write-LogInfo "$($File.FullName) -> type=$ftype -> header=$(Split-Path $headerFileForType -Leaf)"
        }
        return
    }

    # Process each file
    $Processed = 0
    $Skipped = 0

    foreach ($File in $PromptFiles) {
        $relpath = $File.FullName.Substring($TargetDirectory.Length).TrimStart('\', '/')
        $ftype = ''

        if ($relpath -match '(^|/)prompts?(/|$)' -or $File.Name -match '\.prompt\.md$') {
            $ftype = 'prompts'
        }
        elseif ($relpath -match '(^|/)chatmodes?(/|$)' -or $File.Name -match '\.chatmode\.md$' -or $File.Name -match 'chatmode') {
            $ftype = 'chatmodes'
        }
        elseif ($relpath -match '(^|/)instructions?(/|$)' -or $File.Name -match '\.instruction\.md$' -or $File.Name -match 'instructions') {
            $ftype = 'instructions'
        }
        else {
            $ftype = 'generic'
        }

        $headerFileForType = Get-HeaderFile $AiPlatform $ftype
        $headerTemplate = Get-YamlHeaderTemplate $headerFileForType
        $description = Get-FileDescription $File.FullName
        $headerContentForFile = $headerTemplate.Replace('{{DESCRIPTION}}', $description)

        if (Invoke-ProcessFile $File $headerContentForFile $Force $Backup $DryRun $Merge) {
            $Processed++
        }
        else {
            $Skipped++
        }
    }

    # Summary
    Write-Host ""
    if ($DryRun) {
        Write-LogInfo "DRY RUN COMPLETE"
        Write-LogInfo "Would process: $Processed files"
        Write-LogInfo "Would skip: $Skipped files"
    }
    else {
        Write-LogSuccess "HEADER INJECTION COMPLETE"
        Write-LogSuccess "Processed: $Processed files"
        if ($Skipped -gt 0) {
            Write-LogInfo "Skipped: $Skipped files"
        }
    }

    # Create configuration file - DISABLED
    # if (-not $DryRun -and $Processed -gt 0) {
    #     $ConfigFile = Join-Path $TargetDirectory ".plaesy-headers.json"
    #     $Config = @{
    #         ai_platform                      = $AiPlatform
    #         injection_timestamp              = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
    #         processed_files                  = $Processed
    #         header_file                      = Split-Path $HeaderFile -Leaf
    #         constitutional_framework_version = "3.0.0"
    #     } | ConvertTo-Json -Depth 3

    #     Set-Content -Path $ConfigFile -Value $Config
    #     Write-LogInfo "Created configuration: $(Split-Path $ConfigFile -Leaf)"
    # }
}

# Execute main function
try {
    Invoke-Main
}
catch {
    Write-LogError "An error occurred: $($_.Exception.Message)"
    exit 1
}