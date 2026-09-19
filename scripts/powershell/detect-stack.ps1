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
        "docs/project.json",
        ".plaesy/analysis/project.json"
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

# Collect every entry's "extensions" up front so presence-checking is one
# recursive directory walk total, not one walk per extension per entry (the
# previous version called Get-ChildItem -Recurse inside the category loop,
# which re-walked the whole tree for every single extension).
$allExtensions = New-Object System.Collections.Generic.HashSet[string]
foreach ($categoryProp in $mapping.mappings.PSObject.Properties) {
    if ($categoryProp.Name -eq "always_load") { continue }
    foreach ($prop in $categoryProp.Value.PSObject.Properties) {
        foreach ($ext in $prop.Value.extensions) { [void]$allExtensions.Add($ext.ToLower()) }
    }
}
$foundExtensions = New-Object System.Collections.Generic.HashSet[string]
if ($allExtensions.Count -gt 0) {
    $includePatterns = $allExtensions | ForEach-Object { "*$_" }
    Get-ChildItem -Path $TargetDir -Recurse -Depth 3 -Include $includePatterns -File -ErrorAction SilentlyContinue |
        ForEach-Object { [void]$foundExtensions.Add($_.Extension.ToLower()) }
}

# Same one-walk-total approach as extensions above, but matching by exact
# filename (e.g. "next.config.js") instead of extension — for frameworks like
# Next.js whose manifest content carries no distinctive, prose-safe keyword
# (a bare "next" dependency key is too generic to add as a keyword: it would
# match ordinary English "next" in spec/context prose) but whose config
# filename is a reliable, unambiguous signal on its own.
$allFilenames = New-Object System.Collections.Generic.HashSet[string]
foreach ($categoryProp in $mapping.mappings.PSObject.Properties) {
    if ($categoryProp.Name -eq "always_load") { continue }
    foreach ($prop in $categoryProp.Value.PSObject.Properties) {
        foreach ($fn in $prop.Value.filenames) { [void]$allFilenames.Add($fn.ToLower()) }
    }
}
$foundFilenames = New-Object System.Collections.Generic.HashSet[string]
if ($allFilenames.Count -gt 0) {
    $filenamePatterns = $allFilenames | ForEach-Object { $_ }
    Get-ChildItem -Path $TargetDir -Recurse -Depth 3 -Include $filenamePatterns -File -ErrorAction SilentlyContinue |
        ForEach-Object { [void]$foundFilenames.Add($_.Name.ToLower()) }
}

function Test-Category($categoryObj) {
    if (-not $categoryObj) { return }
    foreach ($prop in $categoryObj.PSObject.Properties) {
        $entry = $prop.Value
        if (-not $entry.file) { continue }
        foreach ($kw in $entry.keywords) {
            # Word-boundary match, not plain substring: .Contains() would match
            # short/common keywords ("java" inside "javascript", "pr" inside
            # "prepare"/"private") against any manifest, defeating selective
            # install (see the bash twin's identical fix in detect-stack.sh).
            $pattern = "\b" + [regex]::Escape($kw.ToLower()) + "\b"
            if ($scanTextLower -match $pattern) {
                [void]$matched.Add($entry.file)
                break
            }
        }
        # Presence of a matching document (e.g. a .docx file, not just code
        # importing python-docx) is itself a signal the instructions file
        # should load. Driven generically by each entry's "extensions" array,
        # checked against the single up-front directory walk above.
        foreach ($ext in $entry.extensions) {
            if ($foundExtensions.Contains($ext.ToLower())) {
                [void]$matched.Add($entry.file)
                break
            }
        }
        # Presence of a specific config filename (e.g. next.config.js) is its
        # own strong signal, driven generically by each entry's "filenames"
        # array, checked against the single up-front directory walk above.
        foreach ($fn in $entry.filenames) {
            if ($foundFilenames.Contains($fn.ToLower())) {
                [void]$matched.Add($entry.file)
                break
            }
        }
    }
}

# Iterate every category generically (like the bash twin) instead of a
# hardcoded name list — a new top-level mapping.json category (besides
# always_load) is picked up automatically without editing this script.
foreach ($categoryProp in $mapping.mappings.PSObject.Properties) {
    if ($categoryProp.Name -eq "always_load") { continue }
    Test-Category $categoryProp.Value
}

$matched | Sort-Object | ForEach-Object { Write-Output $_ }
