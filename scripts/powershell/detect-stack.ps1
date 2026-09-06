# Detects which instructions/*.instructions.md files are relevant to a target
# project, using instructions/mapping.json as the keyword registry.
#
# Output: one instruction filename per line (includes always_load files and matched category files)
# Usage: detect-stack.ps1 [-TargetDir <path>] [-PlaesyRoot <path>]

param(
    [string]$TargetDir = ".",
    [string]$PlaesyRoot = ""
)

$ErrorActionPreference = "SilentlyContinue"

# Resolve paths
if ($TargetDir -ne ".") {
    $TargetDir = Resolve-Path -Path $TargetDir -ErrorAction SilentlyContinue
}
if (-not $TargetDir) { $TargetDir = "." }

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $PlaesyRoot) {
    # Navigate up from scripts/powershell to repo root
    $PlaesyRoot = Resolve-Path (Join-Path $ScriptDir "../..") -ErrorAction SilentlyContinue
}
if (-not (Test-Path $PlaesyRoot)) { exit 0 }

$MappingFile = Join-Path $PlaesyRoot "instructions/mapping.json"

# Output all always_load files from mapping.json (dynamic, no hardcoding)
try {
    if (Test-Path $MappingFile) {
        $mapping = Get-Content $MappingFile -Raw | ConvertFrom-Json -ErrorAction Stop
        if ($mapping.mappings.always_load) {
            foreach ($file in $mapping.mappings.always_load) {
                Write-Output $file
            }
        }
    }
}
catch {
    # If mapping fails, silently continue with empty output
    exit 0
}

# Gather scan text from common manifest files and spec/context files
$manifestFiles = @(
    "package.json", "go.mod", "pom.xml", "build.gradle", "Cargo.toml",
    "pubspec.yaml", "Gemfile", "requirements.txt", "*.csproj"
)
# Gather scan text from manifest files. Manifests are found recursively (depth
# limited) so monorepos such as client/pubspec.yaml are detected; .git and
# node_modules are skipped to keep the walk fast.
$scanText = ""
Get-ChildItem -Path $TargetDir -Recurse -Depth 3 -File -Include $manifestFiles -ErrorAction SilentlyContinue |
    Where-Object { $_.FullName -notmatch '\\.git\\|\\node_modules\\' } |
    ForEach-Object {
        $scanText += (Get-Content $_.FullName -Raw -ErrorAction SilentlyContinue) + " "
    }
foreach ($specFile in @(
        "specs/*/context.md",
        "specs/*/requirements.md",
        "docs/specs/*/context.md",
        "docs/specs/*/requirements.md",
        "docs/project.json"
    )) {
    Get-ChildItem -Path (Join-Path $TargetDir $specFile) -File -ErrorAction SilentlyContinue | ForEach-Object {
        $scanText += (Get-Content $_.FullName -Raw -ErrorAction SilentlyContinue) + " "
    }
}
if (Get-ChildItem -Path $TargetDir -Recurse -Depth 3 -Include "*.tsx", "*.jsx" -File -ErrorAction SilentlyContinue | Select-Object -First 1) {
    $scanText += " tsx jsx react "
}
$scanTextLower = $scanText.ToLower()

$matched = New-Object System.Collections.Generic.HashSet[string]

function Test-Category($categoryObj) {
    if (-not $categoryObj) { return }
    foreach ($prop in $categoryObj.PSObject.Properties) {
        $entry = $prop.Value
        if (-not $entry.file) { continue }
        foreach ($kw in $entry.keywords) {
            if ($scanTextLower.Contains($kw.ToLower())) {
                [void]$matched.Add($entry.file)
                break
            }
        }
    }
}

Test-Category $mapping.mappings.frameworks
Test-Category $mapping.mappings.languages
Test-Category $mapping.mappings.cross_cutting
Test-Category $mapping.mappings.methodologies

$matched | Sort-Object | ForEach-Object { Write-Output $_ }
