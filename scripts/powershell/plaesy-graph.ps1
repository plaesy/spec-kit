# plaesy-graph.ps1 - Lightweight knowledge-graph builder for the plaesy spec-kit repo
# Version: 1.0.0 - Compatible with PowerShell 5.1+
# Scans chatmodes/instructions/checklists/templates/scripts/docs/prompts for
# markdown links and script source/call references, then emits:
#   .plaesy/analysis/project.graph.json    - nodes + edges
#   .plaesy/analysis/project.html    - self-contained force-directed viz (no CDN)
#   .plaesy/analysis/reports.md - plain-language summary
#
# Usage:
#   plaesy-graph.ps1                              # build graph for repo root
#   plaesy-graph.ps1 -Path <dir>                  # build graph for a specific dir
#   plaesy-graph.ps1 -Query "task prerequisites"  # keyword query over existing graph
#   plaesy-graph.ps1 -PathFrom "a.md" -PathTo "b.md"  # shortest path between two nodes
#   plaesy-graph.ps1 -Explain "instructions\go.instructions.md"  # explain one node
#   plaesy-graph.ps1 -SemanticQueue                              # export INFERRED edges for an LLM/human to explain
#   plaesy-graph.ps1 -ApplySemantic annotations.json              # merge rationale back into the graph
#   plaesy-graph.ps1 -Watch                                       # rebuild automatically whenever a source file changes
#   plaesy-graph.ps1 -Watch -WatchInterval 5                      # poll every 5s instead of the 3s default
#   plaesy-graph.ps1 -IfChanged                                   # rebuild only if source files changed since last build
#
# Semantic pass (no external API key): -SemanticQueue writes
# .plaesy/analysis/semantic-queue.json, a plain list of {source,target,type}
# for every INFERRED edge. Hand that file to your AI assistant (e.g. this
# session's Claude) with the prompt "explain why each of these files is
# related, one sentence each" - it should return a JSON array of
# {source,target,rationale}. Save that as annotations.json and run
# -ApplySemantic to merge rationale into project.graph.json / reports.md.

param(
    [string]$Path = ".",
    [string]$OutDir = ".plaesy/analysis",
    [string]$Query,
    [string]$PathFrom,
    [string]$PathTo,
    [string]$Explain,
    [string]$ImpactCheck,
    [int]$ImpactDepth = 2,
    [switch]$SemanticQueue,
    [string]$ApplySemantic,
    [switch]$BusinessLogic,
    [switch]$GenerateLearningPaths,
    [string]$FuzzyQuery,
    [int]$SearchDepth = 2,
    [switch]$ImpactVisualize,
    [switch]$Watch,
    [int]$WatchInterval = 3,
    [switch]$IfChanged
)

# Confidence tag per edge type - EXTRACTED = directly parsed from a concrete
# syntactic construct (link, source, call). INFERRED = derived from a weaker
# text-mention signal.
$EdgeConfidence = @{
    references = "EXTRACTED"
    sources    = "EXTRACTED"
    calls      = "EXTRACTED"
    imports    = "EXTRACTED"
    mentions   = "INFERRED"
    mirrors    = "EXTRACTED"
}

$ErrorActionPreference = "Stop"
$RepoRoot = (Resolve-Path $Path).Path
$OutFull = Join-Path $RepoRoot $OutDir
$GraphJsonPath = Join-Path $OutFull "project.graph.json"
$FingerprintPath = Join-Path $OutFull ".fingerprint"

# Extensions we index as nodes. Covers both this repo (.md/.ps1/.sh) and
# arbitrary application codebases -Path may point at.
$SourceExt = @(".ps1", ".sh", ".js", ".jsx", ".ts", ".tsx", ".py", ".go", ".dart", ".java", ".kt", ".kts", ".swift", ".c", ".h", ".cc", ".cpp", ".hpp", ".cs", ".rs", ".rb", ".php")
$IncludeExt = @(".md") + $SourceExt
$ExcludeDirs = @("node_modules", "dist", "build", "__pycache__", "vendor")

# True if any directory segment (excluding the filename itself) starts with a dot
# (.git, .plaesy, .vscode, .idea, .venv, ...) - such paths are always skipped, at any depth.
function Test-UnderDotDir([string]$RelPath) {
    $segs = $RelPath -split '[\\/]'
    for ($i = 0; $i -lt $segs.Length - 1; $i++) {
        if ($segs[$i].StartsWith(".")) { return $true }
    }
    return $false
}

# Per-language import/call extraction. Regex-based (no tree-sitter dependency),
# but scoped per-language so each pattern is precise rather than one generic
# "path looks like a file" guess. Group 1 must capture the raw import spec.
$LanguagePatterns = @{
    ".js"  = @{ type = "imports"; patterns = @('import\s+.*?from\s+[''"]([^''"]+)[''"]', 'require\([''"]([^''"]+)[''"]\)') }
    ".jsx" = @{ type = "imports"; patterns = @('import\s+.*?from\s+[''"]([^''"]+)[''"]', 'require\([''"]([^''"]+)[''"]\)') }
    ".ts"  = @{ type = "imports"; patterns = @('import\s+.*?from\s+[''"]([^''"]+)[''"]', 'require\([''"]([^''"]+)[''"]\)') }
    ".tsx" = @{ type = "imports"; patterns = @('import\s+.*?from\s+[''"]([^''"]+)[''"]', 'require\([''"]([^''"]+)[''"]\)') }
    ".py"  = @{ type = "imports"; patterns = @('^\s*from\s+([\w\.]+)\s+import', '^\s*import\s+([\w\.]+)') }
    ".go"  = @{ type = "imports"; patterns = @('^\s*"([^"]+)"\s*$') }
    ".dart" = @{ type = "imports"; patterns = @('import\s+[''"]([^''"]+)[''"]') }
    ".java" = @{ type = "imports"; patterns = @('(?m)^\s*import\s+(?:static\s+)?([\w.]+)\s*;') }
}

# Populated by Build-Graph before the import-resolution pass: Java dotted
# class name -> file, and Dart package name (from pubspec.yaml) -> lib dir.
$script:JavaIndex = @{}
$script:DartPkgLib = @{}
# Go: module dir (repo-relative, "" for root) -> declared module path, and
# package dir -> list of .go files in it (Go imports name a package/directory,
# not a single file, so one import spec can fan out to several edges).
$script:GoModules = @{}
$script:GoDirFiles = @{}

# Symbol extraction: function/class names (or markdown headings) per file, so
# -Explain and project.graph.json tell an AI assistant what's actually INSIDE a
# file, not just that the file exists and what it links to. Regex-based, one
# capture group per pattern (the symbol name).
$SymbolPatterns = @{
    ".ps1"  = @('(?m)^\s*function\s+([\w-]+)')
    ".sh"   = @('(?m)^\s*function\s+([\w_]+)', '(?m)^\s*([\w_]+)\s*\(\)\s*\{')
    ".js"   = @('(?m)^\s*(?:export\s+)?function\s+([\w$]+)\s*\(', '(?m)^\s*(?:export\s+)?class\s+([\w$]+)')
    ".jsx"  = @('(?m)^\s*(?:export\s+)?function\s+([\w$]+)\s*\(', '(?m)^\s*(?:export\s+)?class\s+([\w$]+)')
    ".ts"   = @('(?m)^\s*(?:export\s+)?function\s+([\w$]+)\s*\(', '(?m)^\s*(?:export\s+)?class\s+([\w$]+)')
    ".tsx"  = @('(?m)^\s*(?:export\s+)?function\s+([\w$]+)\s*\(', '(?m)^\s*(?:export\s+)?class\s+([\w$]+)')
    ".py"   = @('(?m)^\s*def\s+([\w_]+)\s*\(', '(?m)^\s*class\s+([\w_]+)')
    ".go"   = @('(?m)^func\s+(?:\([^)]*\)\s*)?([\w]+)\s*\(')
    ".md"   = @('(?m)^#{1,3}\s+(.+?)\s*$')
}

function Get-Symbols {
    param([string]$Content, [string]$Ext)
    if (-not $SymbolPatterns.ContainsKey($Ext)) { return @() }
    $found = New-Object System.Collections.Generic.List[string]
    foreach ($pat in $SymbolPatterns[$Ext]) {
        $matches = [regex]::Matches($Content, $pat)
        foreach ($m in $matches) {
            $sym = $m.Groups[1].Value.Trim()
            if ($sym -and -not $found.Contains($sym)) { $found.Add($sym) }
        }
    }
    return @($found)
}

function Get-NodeType {
    param([string]$RelPath)
    $top = ($RelPath -split '[\\/]')[0]
    switch ($top) {
        "chatmodes" { return "chatmode" }
        "instructions" { return "instruction" }
        "checklists" { return "checklist" }
        "templates" { return "template" }
        "scripts" { return "script" }
        "docs" { return "doc" }
        "prompts" { return "prompt" }
        "testing" { return "testing" }
        default {
            $ext = [System.IO.Path]::GetExtension($RelPath).ToLower()
            if ($SourceExt -contains $ext) { return "source" }
            return "other"
        }
    }
}

function Get-ArchitectureLayer {
    # Detect semantic architecture layer based on folder path and filename patterns
    # Returns: API, Service, Data, UI, Utility, or null if indeterminate
    param([string]$RelPath)

    $parts = $RelPath -split '[\\/]'
    $folder = if ($parts.Count -gt 1) { $parts[1] } else { "" }
    $filename = [System.IO.Path]::GetFileNameWithoutExtension($RelPath).ToLower()
    $fullpath = $RelPath.ToLower()

    # API Layer patterns
    if ($folder -match '^(api|controllers|handlers|routes|endpoints|gateway)$' -or
        $filename -match '(controller|handler|route|gateway|middleware|app|server)' -or
        $fullpath -match '(^|/)api/' -or $fullpath -match '(^|/)controllers/' -or
        $fullpath -match '(^|/)routes/' -or $fullpath -match '(^|/)handlers/') {
        return "API"
    }

    # Service Layer patterns
    if ($folder -match '^(services|business|use-?cases|orchestration|workflows)$' -or
        $filename -match '(service|business|usecase|workflow|orchestration)' -or
        $fullpath -match '(^|/)services/' -or $fullpath -match '(^|/)business/' -or
        $fullpath -match '(^|/)use-?cases/') {
        return "Service"
    }

    # Data Layer patterns (repositories, models, database)
    if ($folder -match '^(repositories|repo|models|database|db|schemas|data|persistence)$' -or
        $filename -match '(repo|repository|model|entity|schema|mapper|query|dao)' -or
        $fullpath -match '(^|/)repositories/' -or $fullpath -match '(^|/)models/' -or
        $fullpath -match '(^|/)database/' -or $fullpath -match '(^|/)db/' -or
        $fullpath -match '(^|/)persistence/' -or $fullpath -match '(^|/)entities/') {
        return "Data"
    }

    # UI Layer patterns (components, views, pages)
    if ($folder -match '^(components|views|pages|screens|ui|presentation|widgets)$' -or
        $filename -match '(component|view|page|screen|widget)' -or
        $fullpath -match '(^|/)components/' -or $fullpath -match '(^|/)views/' -or
        $fullpath -match '(^|/)pages/' -or $fullpath -match '(^|/)screens/' -or
        $fullpath -match '(^|/)ui/') {
        return "UI"
    }

    # Utility Layer patterns
    if ($folder -match '^(utils|helpers|common|lib|tools|config|constants)$' -or
        $filename -match '(util|helper|common|constant|config|logger|validator|formatter)' -or
        $fullpath -match '(^|/)utils/' -or $fullpath -match '(^|/)helpers/' -or
        $fullpath -match '(^|/)common/' -or $fullpath -match '(^|/)config/' -or
        $fullpath -match '(^|/)constants/') {
        return "Utility"
    }

    return $null
}

function Resolve-ImportTarget {
    # Best-effort resolution of an import spec to a file already in $nodes.
    # Tries relative-path + common extensions, then falls back to a suffix
    # match against known nodes (handles bare module specs like "./utils").
    param([string]$FromRel, [string]$Spec, [string]$Ext, [hashtable]$KnownNodes)

    if ($Ext -eq ".java") {
        $spec = $Spec.Trim()
        if ($spec -match '\*$') { return $null } # wildcard import, no single target
        if ($script:JavaIndex.ContainsKey($spec)) { return $script:JavaIndex[$spec] }
        return $null
    }

    if ($Ext -eq ".dart") {
        if ($Spec -match '^dart:') { return $null }
        if ($Spec -match '^package:([^/]+)/(.+)$') {
            $pkgName = $Matches[1]; $subPath = $Matches[2]
            if (-not $script:DartPkgLib.ContainsKey($pkgName)) { return $null }
            try {
                $resolved = [System.IO.Path]::GetFullPath((Join-Path $script:DartPkgLib[$pkgName] $subPath))
            } catch { return $null }
        } else {
            $fromDir = Split-Path -Parent (Join-Path $RepoRoot $FromRel)
            try {
                $resolved = [System.IO.Path]::GetFullPath((Join-Path $fromDir $Spec))
            } catch { return $null }
        }
        if ($resolved.StartsWith($RepoRoot, [System.StringComparison]::OrdinalIgnoreCase) -and (Test-Path $resolved -PathType Leaf)) {
            return ($resolved.Substring($RepoRoot.Length).TrimStart('\','/')) -replace '/', '\'
        }
        return $null
    }

    if ($Spec -match '^[a-zA-Z0-9_-]+$' -and $Ext -eq ".py") {
        # bare top-level python module import (e.g. "os", "requests") - skip, not repo-local
        if (-not ($KnownNodes.Keys | Where-Object { $_ -like "*\$Spec.py" -or $_ -like "*\$Spec\__init__.py" })) { return $null }
    }
    if ($Spec -notmatch '^\.' -and $Ext -ne ".py" -and $Ext -ne ".go") { return $null } # skip bare npm package imports

    if ($Ext -eq ".py" -and $Spec -match '^\.+$') { return $null } # relative-package marker only ("." / ".."), no concrete module named

    $fromDir = Split-Path -Parent (Join-Path $RepoRoot $FromRel)
    $candidates = if ($Ext -eq ".py") {
        @("$Spec.py", "$($Spec -replace '\.', '\')  \__init__.py" -replace '\s','')
    } else {
        @($Spec, "$Spec.js", "$Spec.jsx", "$Spec.ts", "$Spec.tsx", "$Spec/index.js", "$Spec/index.ts")
    }
    foreach ($c in $candidates) {
        try {
            $resolved = [System.IO.Path]::GetFullPath((Join-Path $fromDir $c))
        } catch { continue }
        if ($resolved.StartsWith($RepoRoot, [System.StringComparison]::OrdinalIgnoreCase) -and (Test-Path $resolved -PathType Leaf)) {
            return ($resolved.Substring($RepoRoot.Length).TrimStart('\','/')) -replace '/', '\'
        }
    }
    return $null
}

function Resolve-RelLink {
    param([string]$FromRel, [string]$LinkTarget)
    if ($LinkTarget -match '^(https?:)?//' -or $LinkTarget -match '^#') { return $null }
    $LinkTarget = ($LinkTarget -split '#')[0].Trim()
    if (-not $LinkTarget) { return $null }
    $fromDir = Split-Path -Parent (Join-Path $RepoRoot $FromRel)
    try {
        $resolved = Join-Path $fromDir $LinkTarget
        $resolved = [System.IO.Path]::GetFullPath($resolved)
    } catch { return $null }
    if (-not $resolved.StartsWith($RepoRoot, [System.StringComparison]::OrdinalIgnoreCase)) { return $null }
    if (-not (Test-Path $resolved -PathType Leaf)) { return $null }
    return ($resolved.Substring($RepoRoot.Length).TrimStart('\','/')) -replace '/', '\'
}

function Get-Communities {
    param($NodeIds, $Edges)

    $adj = @{}
    foreach ($n in $NodeIds) { $adj[$n] = New-Object System.Collections.Generic.List[string] }
    foreach ($e in $Edges) {
        if ($adj.ContainsKey($e.source)) { $adj[$e.source].Add($e.target) }
        if ($adj.ContainsKey($e.target)) { $adj[$e.target].Add($e.source) }
    }

    $label = @{}
    foreach ($n in $NodeIds) { $label[$n] = $n }

    $nodeList = @($NodeIds)
    $maxIterations = 15
    for ($iter = 0; $iter -lt $maxIterations; $iter++) {
        $changed = $false
        foreach ($n in $nodeList) {
            $neighbors = $adj[$n]
            if ($neighbors.Count -eq 0) { continue }
            $counts = @{}
            foreach ($nb in $neighbors) {
                $l = $label[$nb]
                if (-not $counts.ContainsKey($l)) { $counts[$l] = 0 }
                $counts[$l]++
            }
            $best = $null
            $bestCount = -1
            foreach ($k in ($counts.Keys | Sort-Object)) {
                if ($counts[$k] -gt $bestCount) { $bestCount = $counts[$k]; $best = $k }
            }
            if ($best -and $best -ne $label[$n]) {
                $label[$n] = $best
                $changed = $true
            }
        }
        if (-not $changed) { break }
    }

    # Normalize labels into stable sequential community ids ordered by size
    $groups = $label.GetEnumerator() | Group-Object Value | Sort-Object Count -Descending
    $communityId = @{}
    $idx = 0
    foreach ($g in $groups) {
        $communityId[$g.Name] = $idx
        $idx++
    }
    $result = @{}
    foreach ($n in $NodeIds) { $result[$n] = $communityId[$label[$n]] }
    return $result
}

function Build-Graph {
    Write-Host "[plaesy-graph] Scanning $RepoRoot ..." -ForegroundColor Cyan

    $outDirNorm = ($OutDir -replace '/', '\').Trim('\')
    $allFiles = Get-ChildItem -Path $RepoRoot -Recurse -File | Where-Object {
        $rel = ($_.FullName.Substring($RepoRoot.Length).TrimStart('\','/')) -replace '/', '\'
        $top = ($rel -split '\\')[0]
        $underOutDir = $rel -eq $outDirNorm -or $rel.StartsWith("$outDirNorm\", [System.StringComparison]::OrdinalIgnoreCase)
        ($IncludeExt -contains $_.Extension.ToLower()) -and ($ExcludeDirs -notcontains $top) -and (-not (Test-UnderDotDir $rel)) -and (-not $underOutDir)
    }

    $nodes = @{}
    $edges = New-Object System.Collections.Generic.List[object]

    foreach ($f in $allFiles) {
        $rel = $f.FullName.Substring($RepoRoot.Length).TrimStart('\','/') -replace '/', '\'
        $layer = Get-ArchitectureLayer $rel
        $nodes[$rel] = @{
            id    = $rel
            label = $f.Name
            type  = Get-NodeType $rel
            group = ($rel -split '\\')[0]
        }
        if ($layer) { $nodes[$rel].layer = $layer }
        if ($f.Extension -eq ".java") {
            $parts = $rel -split '\\'
            if ($parts.Count -gt 1) {
                $dotted = ($parts[1..($parts.Count - 1)] -join '.') -replace '\.java$', ''
                $script:JavaIndex[$dotted] = $rel
            }
        }
        if ($f.Extension -eq ".go" -and $f.Name -notlike "*_test.go") {
            $dir = Split-Path -Parent $rel
            if (-not $script:GoDirFiles.ContainsKey($dir)) { $script:GoDirFiles[$dir] = New-Object System.Collections.Generic.List[string] }
            $script:GoDirFiles[$dir].Add($rel)
        }
    }

    Get-ChildItem -Path $RepoRoot -Recurse -Filter "pubspec.yaml" -File -ErrorAction SilentlyContinue | Where-Object {
        $rel = $_.FullName.Substring($RepoRoot.Length).TrimStart('\','/')
        $top = ($rel -split '[\\/]')[0]
        ($ExcludeDirs -notcontains $top) -and (-not (Test-UnderDotDir $rel))
    } | ForEach-Object {
        $nameLine = Get-Content $_.FullName -ErrorAction SilentlyContinue | Where-Object { $_ -match '^name:\s*(\S+)' } | Select-Object -First 1
        if ($nameLine -match '^name:\s*(\S+)') {
            $script:DartPkgLib[$Matches[1]] = Join-Path (Split-Path -Parent $_.FullName) "lib"
        }
    }

    Get-ChildItem -Path $RepoRoot -Recurse -Filter "go.mod" -File -ErrorAction SilentlyContinue | Where-Object {
        $rel = $_.FullName.Substring($RepoRoot.Length).TrimStart('\','/')
        $top = ($rel -split '[\\/]')[0]
        ($ExcludeDirs -notcontains $top) -and (-not (Test-UnderDotDir $rel))
    } | ForEach-Object {
        $modLine = Get-Content $_.FullName -ErrorAction SilentlyContinue | Where-Object { $_ -match '^module\s+(\S+)' } | Select-Object -First 1
        if ($modLine -match '^module\s+(\S+)') {
            $modDir = Split-Path -Parent $_.FullName
            $modRel = ($modDir.Substring($RepoRoot.Length).TrimStart('\','/')) -replace '/', '\'
            $script:GoModules[$modRel] = $Matches[1]
        }
    }

    foreach ($f in $allFiles) {
        $rel = $f.FullName.Substring($RepoRoot.Length).TrimStart('\','/') -replace '/', '\'
        $content = Get-Content -Path $f.FullName -Raw -ErrorAction SilentlyContinue
        if (-not $content) { continue }

        $nodes[$rel].symbols = Get-Symbols -Content $content -Ext $f.Extension

        if ($f.Extension -eq ".md") {
            $mdLinkMatches = [regex]::Matches($content, '\[[^\]]*\]\(([^)\s]+)\)')
            foreach ($m in $mdLinkMatches) {
                $target = Resolve-RelLink -FromRel $rel -LinkTarget $m.Groups[1].Value
                if ($target -and $nodes.ContainsKey($target) -and $target -ne $rel) {
                    $edges.Add(@{ source = $rel; target = $target; type = "references"; confidence = $EdgeConfidence["references"] })
                }
            }
        }

        if ($f.Extension -eq ".ps1") {
            $sourceMatches = [regex]::Matches($content, '(?:\.\s+["'']?)([^"''\r\n]+\.ps1)')
            foreach ($m in $sourceMatches) {
                $target = Resolve-RelLink -FromRel $rel -LinkTarget ($m.Groups[1].Value -replace '\$PSScriptRoot[\\/]?', '')
                if ($target -and $nodes.ContainsKey($target) -and $target -ne $rel) {
                    $edges.Add(@{ source = $rel; target = $target; type = "sources"; confidence = $EdgeConfidence["sources"] })
                }
            }
            $callMatches = [regex]::Matches($content, '&\s+["'']([^"'']+\.ps1)["'']')
            foreach ($m in $callMatches) {
                $target = Resolve-RelLink -FromRel $rel -LinkTarget ($m.Groups[1].Value -replace '\$PSScriptRoot[\\/]?', '')
                if ($target -and $nodes.ContainsKey($target) -and $target -ne $rel) {
                    $edges.Add(@{ source = $rel; target = $target; type = "calls"; confidence = $EdgeConfidence["calls"] })
                }
            }
        }

        if ($f.Extension -eq ".sh") {
            $sourceMatches = [regex]::Matches($content, '(?:^|\s)(?:source|\.)\s+["'']?([^"''\r\n]+\.sh)')
            foreach ($m in $sourceMatches) {
                $target = Resolve-RelLink -FromRel $rel -LinkTarget $m.Groups[1].Value
                if ($target -and $nodes.ContainsKey($target) -and $target -ne $rel) {
                    $edges.Add(@{ source = $rel; target = $target; type = "sources"; confidence = $EdgeConfidence["sources"] })
                }
            }
        }

        if ($f.Extension -eq ".go" -and $script:GoModules.Count -gt 0) {
            # Go imports name a package (directory), not a single file, so one
            # import spec fans out to every .go file in the target directory.
            $relDir = Split-Path -Parent $rel
            $bestModDir = $null; $bestLen = -1
            foreach ($md in $script:GoModules.Keys) {
                if ($md -eq "" -or $relDir -eq $md -or $relDir.StartsWith("$md\", [System.StringComparison]::OrdinalIgnoreCase)) {
                    if ($md.Length -gt $bestLen) { $bestLen = $md.Length; $bestModDir = $md }
                }
            }
            if ($null -ne $bestModDir) {
                $modPath = $script:GoModules[$bestModDir]
                $importMatches = [regex]::Matches($content, $LanguagePatterns[".go"].patterns[0], [System.Text.RegularExpressions.RegexOptions]::Multiline)
                foreach ($m in $importMatches) {
                    $spec = $m.Groups[1].Value
                    if ($spec -eq $modPath -or $spec.StartsWith("$modPath/")) {
                        $subPath = $spec.Substring($modPath.Length).TrimStart('/') -replace '/', '\'
                        $targetDir = if ($subPath) { if ($bestModDir) { Join-Path $bestModDir $subPath } else { $subPath } } else { $bestModDir }
                        if ($script:GoDirFiles.ContainsKey($targetDir)) {
                            foreach ($tfile in $script:GoDirFiles[$targetDir]) {
                                if ($tfile -ne $rel) {
                                    $edges.Add(@{ source = $rel; target = $tfile; type = "imports"; confidence = $EdgeConfidence["imports"] })
                                }
                            }
                        }
                    }
                }
            }
        } elseif ($LanguagePatterns.ContainsKey($f.Extension)) {
            foreach ($pat in $LanguagePatterns[$f.Extension].patterns) {
                $importMatches = [regex]::Matches($content, $pat, [System.Text.RegularExpressions.RegexOptions]::Multiline)
                foreach ($m in $importMatches) {
                    $target = Resolve-ImportTarget -FromRel $rel -Spec $m.Groups[1].Value -Ext $f.Extension -KnownNodes $nodes
                    if ($target -and $nodes.ContainsKey($target) -and $target -ne $rel) {
                        $edges.Add(@{ source = $rel; target = $target; type = "imports"; confidence = $EdgeConfidence["imports"] })
                    }
                }
            }
        }

        # Plain path mentions to other known files (e.g. "instructions/go.instructions.md")
        # Only meaningful for our own doc/script corpus - skip for source files in
        # arbitrary target codebases where import edges already give precise signal
        # and mention-scanning every file against every other file doesn't scale.
        if ($LanguagePatterns.ContainsKey($f.Extension)) { continue }
        foreach ($cand in $nodes.Keys) {
            if ($cand -eq $rel) { continue }
            $asForward = $cand -replace '\\', '/'
            if ($content -match [regex]::Escape($asForward) -or $content -match [regex]::Escape($cand)) {
                $already = $edges | Where-Object { $_.source -eq $rel -and $_.target -eq $cand }
                if (-not $already) {
                    $edges.Add(@{ source = $rel; target = $cand; type = "mentions"; confidence = $EdgeConfidence["mentions"] })
                }
            }
        }
    }

    # Mirror-install detection: this repo (per scripts/configs/platform.json)
    # installs the same chatmode/instruction/prompt content into multiple
    # per-editor target directories (e.g. chatmodes/*.chatmode.md is mirrored
    # to .claude/roles/*.chatmode.md, .github/chatmodes/*, etc). A pure
    # link/import scanner is blind to this - it's a filename-identity fact
    # about the install layout, not a textual reference. Without it, every
    # mirrored copy looks like an orphan even though it is a 1:1 install
    # target of a source file.
    $byBasename = @{}
    foreach ($n in $nodes.Keys) {
        $base = Split-Path -Leaf $n
        if (-not $byBasename.ContainsKey($base)) { $byBasename[$base] = New-Object System.Collections.Generic.List[string] }
        $byBasename[$base].Add($n)
    }
    foreach ($base in $byBasename.Keys) {
        $group = $byBasename[$base]
        if ($group.Count -lt 2) { continue }
        for ($i = 0; $i -lt $group.Count; $i++) {
            for ($j = $i + 1; $j -lt $group.Count; $j++) {
                $edges.Add(@{ source = $group[$i]; target = $group[$j]; type = "mirrors"; confidence = $EdgeConfidence["mirrors"] })
            }
        }
    }

    # Drop a "mentions" (INFERRED) edge whenever a stronger EXTRACTED edge
    # already exists for the exact same (source,target) pair. Done as a final
    # pass over the complete edge list (not during the earlier per-file loop)
    # because mirrors edges are only added above, after mentions - an
    # in-loop check would miss mentions/mirrors collisions like
    # "docs/chatmodes/README.md -> README.md" (both a mirrors and a mentions
    # edge for the same pair) since mirrors didn't exist yet when that file
    # was scanned.
    $strongPairs = New-Object 'System.Collections.Generic.HashSet[string]'
    foreach ($e in $edges) {
        if ($e.type -ne "mentions") { [void]$strongPairs.Add("$($e.source)|$($e.target)") }
    }
    $edges = [System.Collections.Generic.List[object]](
        $edges | Where-Object { $_.type -ne "mentions" -or -not $strongPairs.Contains("$($_.source)|$($_.target)") }
    )

    # Degree count for god-node detection
    $degree = @{}
    foreach ($n in $nodes.Keys) { $degree[$n] = 0 }
    foreach ($e in $edges) {
        $degree[$e.source]++
        $degree[$e.target]++
    }
    foreach ($n in $nodes.Keys) { $nodes[$n].degree = $degree[$n] }

    # Real community detection (label propagation) - distinct from the
    # folder-based 'group' field. Converges fast and needs no external deps.
    $communities = Get-Communities -NodeIds $nodes.Keys -Edges $edges
    foreach ($n in $nodes.Keys) { $nodes[$n].community = $communities[$n] }

    $nodeArray = New-Object System.Collections.Generic.List[object]
    foreach ($v in $nodes.Values) { $nodeArray.Add($v) }
    $edgeArray = New-Object System.Collections.Generic.List[object]
    foreach ($e in $edges) { $edgeArray.Add($e) }

    $description = "Knowledge graph of $(Split-Path -Leaf $RepoRoot) ($($nodeArray.Count) files): source, docs, config and their references, calls, imports, and text mentions."

    $graph = @{
        generated_at = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
        root         = $RepoRoot
        description  = $description
        nodes        = $nodeArray
        edges        = $edgeArray
    }

    New-Item -ItemType Directory -Force -Path $OutFull | Out-Null
    $graph | ConvertTo-Json -Depth 6 | Set-Content -Path $GraphJsonPath -Encoding UTF8

    Write-GraphHtml -Graph $graph
    Write-GraphReport -Graph $graph

    Write-Host "[plaesy-graph] Done. Nodes: $($nodes.Count)  Edges: $($edges.Count)" -ForegroundColor Green
    Write-Host "[plaesy-graph] Output: $OutFull" -ForegroundColor Green

    try { Get-SourceFingerprint | Set-Content -Path $FingerprintPath -Encoding UTF8 } catch {}
}

function Get-SourceFingerprint {
    # Cheap change-detection signature for -Watch: file count + max
    # LastWriteTime across the same file set Build-Graph scans. Avoids
    # hashing file contents (slow on a repo this size) - a full rescan is
    # already fast, so we only need to know *whether* to rescan, not *what*
    # changed.
    $outDirNorm = ($OutDir -replace '/', '\').Trim('\')
    $files = Get-ChildItem -Path $RepoRoot -Recurse -File | Where-Object {
        $rel = ($_.FullName.Substring($RepoRoot.Length).TrimStart('\','/')) -replace '/', '\'
        $top = ($rel -split '\\')[0]
        $underOutDir = $rel -eq $outDirNorm -or $rel.StartsWith("$outDirNorm\", [System.StringComparison]::OrdinalIgnoreCase)
        ($IncludeExt -contains $_.Extension.ToLower()) -and ($ExcludeDirs -notcontains $top) -and (-not (Test-UnderDotDir $rel)) -and (-not $underOutDir)
    }
    if ($files.Count -eq 0) { return "0|0" }
    $maxWrite = ($files | Measure-Object -Property LastWriteTimeUtc -Maximum).Maximum.Ticks
    return "$($files.Count)|$maxWrite"
}

function Invoke-Watch {
    Write-Host "[plaesy-graph] Watch mode: rebuilding now, then polling every ${WatchInterval}s. Ctrl+C to stop." -ForegroundColor Cyan
    Build-Graph
    $lastFingerprint = Get-SourceFingerprint
    while ($true) {
        Start-Sleep -Seconds $WatchInterval
        $fp = Get-SourceFingerprint
        if ($fp -ne $lastFingerprint) {
            Write-Host "[plaesy-graph] Change detected, rebuilding..." -ForegroundColor Cyan
            Build-Graph
            $lastFingerprint = Get-SourceFingerprint
        }
    }
}

function Write-GraphHtml {
    param($Graph)
    $jsonForHtml = ($Graph | ConvertTo-Json -Depth 6 -Compress)
    $html = @"
<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Plaesy Graph</title>
<style>
  body{margin:0;font-family:Segoe UI,Arial,sans-serif;background:#111;color:#eee;}
  #graph{width:100vw;height:100vh;display:block;}
  #panel{position:fixed;top:10px;left:10px;background:#1e1e1eee;padding:10px 14px;border-radius:8px;max-width:320px;font-size:13px;}
  #panel input{width:100%;box-sizing:border-box;margin-top:6px;}
  .legend span{display:inline-block;width:10px;height:10px;border-radius:50%;margin-right:4px;}
</style></head>
<body>
<div id="panel">
  <b>Plaesy Knowledge Graph</b><br/>
  <input id="search" placeholder="filter nodes..." />
  <div id="info" style="margin-top:8px;color:#aaa;"></div>
</div>
<canvas id="graph"></canvas>
<script>
const data = $jsonForHtml;
const canvas = document.getElementById('graph');
const ctx = canvas.getContext('2d');
function resize(){canvas.width=window.innerWidth;canvas.height=window.innerHeight;}
window.addEventListener('resize', resize); resize();

const groups = [...new Set(data.nodes.map(n=>n.group))];
const colors = ['#4fc3f7','#81c784','#ffb74d','#e57373','#ba68c8','#4db6ac','#f06292','#a1887f','#90a4ae'];
const colorOf = g => colors[groups.indexOf(g) % colors.length];

const nodes = data.nodes.map(n => ({...n, x: Math.random()*canvas.width, y: Math.random()*canvas.height, vx:0, vy:0}));
const idx = {}; nodes.forEach((n,i)=>idx[n.id]=i);
const edges = data.edges.filter(e => idx[e.source]!==undefined && idx[e.target]!==undefined);

let filter = '';
document.getElementById('search').addEventListener('input', e => { filter = e.target.value.toLowerCase(); });

function tick(){
  const k = 0.002, rep = 800, damp = 0.85, center = 0.0005;
  for(let i=0;i<nodes.length;i++){
    for(let j=i+1;j<nodes.length;j++){
      const a=nodes[i], b=nodes[j];
      let dx=a.x-b.x, dy=a.y-b.y;
      let dist = Math.sqrt(dx*dx+dy*dy)||1;
      const f = rep/(dist*dist);
      dx/=dist; dy/=dist;
      a.vx += dx*f; a.vy += dy*f;
      b.vx -= dx*f; b.vy -= dy*f;
    }
  }
  edges.forEach(e=>{
    const a=nodes[idx[e.source]], b=nodes[idx[e.target]];
    let dx=b.x-a.x, dy=b.y-a.y;
    a.vx += dx*k; a.vy += dy*k;
    b.vx -= dx*k; b.vy -= dy*k;
  });
  nodes.forEach(n=>{
    n.vx += (canvas.width/2 - n.x)*center;
    n.vy += (canvas.height/2 - n.y)*center;
    n.vx*=damp; n.vy*=damp;
    n.x += n.vx; n.y += n.vy;
  });
}

function draw(){
  ctx.clearRect(0,0,canvas.width,canvas.height);
  ctx.strokeStyle = '#444';
  edges.forEach(e=>{
    const a=nodes[idx[e.source]], b=nodes[idx[e.target]];
    ctx.beginPath(); ctx.moveTo(a.x,a.y); ctx.lineTo(b.x,b.y); ctx.stroke();
  });
  nodes.forEach(n=>{
    const match = filter && n.id.toLowerCase().includes(filter);
    const r = 3 + Math.min(10, n.degree);
    ctx.beginPath();
    ctx.fillStyle = filter ? (match ? colorOf(n.group) : '#333') : colorOf(n.group);
    ctx.arc(n.x, n.y, r, 0, Math.PI*2); ctx.fill();
    if (match || n.degree > 4) {
      ctx.fillStyle = '#eee'; ctx.font = '10px sans-serif';
      ctx.fillText(n.label, n.x+r+2, n.y+3);
    }
  });
}

function loop(){ tick(); draw(); requestAnimationFrame(loop); }
loop();

canvas.addEventListener('mousemove', e=>{
  const mx=e.clientX, my=e.clientY;
  let closest=null, best=20;
  nodes.forEach(n=>{ const d=Math.hypot(n.x-mx,n.y-my); if(d<best){best=d;closest=n;} });
  document.getElementById('info').textContent = closest ? closest.id + ' (' + closest.type + ', degree ' + closest.degree + ')' : '';
});
</script>
</body></html>
"@
    Set-Content -Path (Join-Path $OutFull "project.html") -Value $html -Encoding UTF8
}

function Write-GraphReport {
    param($Graph)
    $nodes = $Graph.nodes
    $edges = $Graph.edges
    # Nodes may be Hashtable (fresh build) or PSCustomObject (round-tripped
    # through Load-Graph/JSON) - dot access works for both, unlike the ['key']
    # indexer which only Hashtable supports.
    $byType = $nodes | Group-Object { $_.type } | Sort-Object Count -Descending
    $byGroup = $nodes | Group-Object { $_.group } | Sort-Object Count -Descending
    $godNodes = $nodes | Sort-Object { $_.degree } -Descending | Select-Object -First 10
    $orphans = $nodes | Where-Object { $_.degree -eq 0 }

    $sb = New-Object System.Text.StringBuilder
    [void]$sb.AppendLine("# Plaesy Graph Report")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("Generated: $($Graph.generated_at)")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## Overview")
    [void]$sb.AppendLine("- Nodes: $($nodes.Count)")
    [void]$sb.AppendLine("- Edges: $($edges.Count)")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## Nodes by type")
    foreach ($g in $byType) { [void]$sb.AppendLine("- $($g.Name): $($g.Count)") }
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## Communities (top-level folders)")
    foreach ($g in $byGroup) { [void]$sb.AppendLine("- $($g.Name): $($g.Count) files") }
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## Detected communities (label-propagation clustering)")
    [void]$sb.AppendLine("These are computed from actual edge structure, not folder names - they surface cross-folder subsystems.")
    $byCommunity = $nodes | Group-Object { $_.community } | Where-Object { $_.Count -gt 1 } | Sort-Object Count -Descending | Select-Object -First 10
    foreach ($c in $byCommunity) {
        $members = $c.Group | Sort-Object { $_.degree } -Descending | Select-Object -First 5 | ForEach-Object { $_.id }
        [void]$sb.AppendLine("- Community $($c.Name) ($($c.Count) files): $($members -join ', ')")
    }
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## God nodes (most connected)")
    foreach ($n in $godNodes) { [void]$sb.AppendLine("- ``$($n.id)`` - degree $($n.degree)") }
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("## Orphan files (no detected references)")
    if ($orphans.Count -eq 0) {
        [void]$sb.AppendLine("- None")
    } else {
        foreach ($o in ($orphans | Select-Object -First 30)) { [void]$sb.AppendLine("- ``$($o.id)``") }
        if ($orphans.Count -gt 30) { [void]$sb.AppendLine("- ... and $($orphans.Count - 30) more") }
    }
    [void]$sb.AppendLine("")
    $annotated = $edges | Where-Object { $_.rationale }
    if ($annotated) {
        [void]$sb.AppendLine("## Explained relationships (semantic pass)")
        foreach ($e in $annotated) { [void]$sb.AppendLine("- ``$($e.source)`` -> ``$($e.target)``: $($e.rationale)") }
        [void]$sb.AppendLine("")
    }
    [void]$sb.AppendLine("## Suggested questions")
    [void]$sb.AppendLine("- Which files are referenced by the most other files?")
    [void]$sb.AppendLine("- Which files have no callers (potentially dead)?")
    if ($byGroup.Count -ge 2) {
        [void]$sb.AppendLine("- What connects ``$($byGroup[0].Name)`` to ``$($byGroup[1].Name)``?")
    }

    Set-Content -Path (Join-Path $OutFull "reports.md") -Value $sb.ToString() -Encoding UTF8
}

function Load-Graph {
    if (-not (Test-Path $GraphJsonPath)) {
        Write-Host "[plaesy-graph] No graph found at $GraphJsonPath. Run without -Query/-PathFrom/-Explain first to build it." -ForegroundColor Yellow
        exit 1
    }
    return Get-Content $GraphJsonPath -Raw | ConvertFrom-Json
}

function Get-Neighbors {
    param($Graph, [string]$NodeId)
    $out = New-Object System.Collections.Generic.List[string]
    foreach ($e in $Graph.edges) {
        if ($e.source -eq $NodeId) { $out.Add($e.target) }
        if ($e.target -eq $NodeId) { $out.Add($e.source) }
    }
    return $out | Select-Object -Unique
}

function Invoke-Query {
    param([string]$Q)
    $graph = Load-Graph
    $ql = $Q.ToLower()
    $matches = $graph.nodes | Where-Object { $_.id.ToLower().Contains($ql) -or $_.label.ToLower().Contains($ql) }
    if (-not $matches) {
        Write-Host "[plaesy-graph] No nodes matched '$Q'." -ForegroundColor Yellow
        return
    }
    foreach ($m in $matches) {
        Write-Host ""
        Write-Host "== $($m.id) ($($m.type), degree $($m.degree)) ==" -ForegroundColor Cyan
        $neighbors = Get-Neighbors -Graph $graph -NodeId $m.id
        if ($neighbors.Count -eq 0) {
            Write-Host "  (no connections found)" -ForegroundColor DarkGray
        } else {
            foreach ($n in $neighbors) { Write-Host "  -> $n" }
        }
    }
}

function Invoke-PathQuery {
    param([string]$From, [string]$To)
    $graph = Load-Graph
    $adj = @{}
    foreach ($n in $graph.nodes) { $adj[$n.id] = New-Object System.Collections.Generic.List[string] }
    foreach ($e in $graph.edges) {
        if ($adj.ContainsKey($e.source)) { $adj[$e.source].Add($e.target) }
        if ($adj.ContainsKey($e.target)) { $adj[$e.target].Add($e.source) }
    }
    $fromNode = ($graph.nodes | Where-Object { $_.id -like "*$From*" } | Select-Object -First 1).id
    $toNode = ($graph.nodes | Where-Object { $_.id -like "*$To*" } | Select-Object -First 1).id
    if (-not $fromNode -or -not $toNode) {
        Write-Host "[plaesy-graph] Could not resolve one or both node names." -ForegroundColor Yellow
        return
    }
    $visited = @{ $fromNode = $null }
    $queue = New-Object System.Collections.Generic.Queue[string]
    $queue.Enqueue($fromNode)
    while ($queue.Count -gt 0) {
        $cur = $queue.Dequeue()
        if ($cur -eq $toNode) { break }
        foreach ($nb in $adj[$cur]) {
            if (-not $visited.ContainsKey($nb)) {
                $visited[$nb] = $cur
                $queue.Enqueue($nb)
            }
        }
    }
    if (-not $visited.ContainsKey($toNode)) {
        Write-Host "[plaesy-graph] No path found between $fromNode and $toNode." -ForegroundColor Yellow
        return
    }
    $path = New-Object System.Collections.Generic.List[string]
    $cursor = $toNode
    while ($null -ne $cursor) {
        $path.Insert(0, $cursor)
        $cursor = $visited[$cursor]
    }
    Write-Host ("[plaesy-graph] Path (" + ($path.Count - 1) + " hops):") -ForegroundColor Cyan
    Write-Host ($path -join "  ->  ")
}

function Invoke-Explain {
    param([string]$NodeName)
    $graph = Load-Graph
    $node = $graph.nodes | Where-Object { $_.id -like "*$NodeName*" } | Select-Object -First 1
    if (-not $node) {
        Write-Host "[plaesy-graph] No node matching '$NodeName'." -ForegroundColor Yellow
        return
    }
    Write-Host ""
    Write-Host "$($node.id)" -ForegroundColor Cyan
    Write-Host "  type: $($node.type)   group: $($node.group)   degree: $($node.degree)"
    if ($node.symbols -and $node.symbols.Count -gt 0) {
        Write-Host "  symbols ($($node.symbols.Count)): $($node.symbols -join ', ')"
    }
    $outEdges = $graph.edges | Where-Object { $_.source -eq $node.id }
    $inEdges = $graph.edges | Where-Object { $_.target -eq $node.id }
    Write-Host "  outgoing:"
    if ($outEdges) { $outEdges | ForEach-Object { Write-Host "    -[$($_.type)]-> $($_.target)" } } else { Write-Host "    (none)" }
    Write-Host "  incoming:"
    if ($inEdges) { $inEdges | ForEach-Object { Write-Host "    <-[$($_.type)]- $($_.source)" } } else { Write-Host "    (none)" }
}

function Invoke-ImpactCheck {
    param([string]$NodeName, [int]$Depth)
    $graph = Load-Graph
    $start = ($graph.nodes | Where-Object { $_.id -like "*$NodeName*" } | Select-Object -First 1).id
    if (-not $start) {
        Write-Host "[plaesy-graph] No node matching '$NodeName'." -ForegroundColor Yellow
        return
    }
    $adj = @{}
    foreach ($n in $graph.nodes) { $adj[$n.id] = New-Object System.Collections.Generic.List[string] }
    foreach ($e in $graph.edges) {
        if ($adj.ContainsKey($e.source)) { $adj[$e.source].Add($e.target) }
        if ($adj.ContainsKey($e.target)) { $adj[$e.target].Add($e.source) }
    }

    $visited = @{ $start = 0 }
    $queue = New-Object System.Collections.Generic.Queue[string]
    $queue.Enqueue($start)
    while ($queue.Count -gt 0) {
        $cur = $queue.Dequeue()
        $d = $visited[$cur]
        if ($d -ge $Depth) { continue }
        foreach ($nb in $adj[$cur]) {
            if (-not $visited.ContainsKey($nb)) {
                $visited[$nb] = $d + 1
                $queue.Enqueue($nb)
            }
        }
    }
    $visited.Remove($start)

    Write-Host ""
    Write-Host "[plaesy-graph] Impact of changing $start (depth $Depth):" -ForegroundColor Cyan
    if ($visited.Count -eq 0) {
        Write-Host "  No connected files found - safe to change in isolation (per detected edges)." -ForegroundColor DarkGray
        return
    }
    $byDepth = $visited.GetEnumerator() | Group-Object Value | Sort-Object Name
    foreach ($g in $byDepth) {
        Write-Host "  Depth $($g.Name):" -ForegroundColor Yellow
        foreach ($item in $g.Group) {
            $n = $graph.nodes | Where-Object { $_.id -eq $item.Name } | Select-Object -First 1
            Write-Host "    - $($item.Name)  [$($n.type)]"
        }
    }
}

function Invoke-SemanticQueue {
    $graph = Load-Graph
    $inferred = $graph.edges | Where-Object { $_.confidence -eq "INFERRED" }
    $queuePath = Join-Path $OutFull "semantic-queue.json"
    if ($inferred.Count -eq 0) {
        Write-Host "[plaesy-graph] No INFERRED edges to review - nothing queued." -ForegroundColor Yellow
        return
    }
    $queue = $inferred | ForEach-Object { @{ source = $_.source; target = $_.target; type = $_.type } }
    $queue | ConvertTo-Json -Depth 4 | Set-Content -Path $queuePath -Encoding UTF8
    Write-Host "[plaesy-graph] Wrote $($queue.Count) INFERRED edges to $queuePath" -ForegroundColor Cyan
    Write-Host "[plaesy-graph] Ask your AI assistant to explain each pair, save the result as" -ForegroundColor Cyan
    Write-Host "  a JSON array of {source,target,rationale}, then run -ApplySemantic <file>." -ForegroundColor Cyan
}

# Feature #2: Business Logic Mapper
function Invoke-BusinessLogicQueue {
    $graph = Load-Graph
    $inferred = $graph.edges | Where-Object { $_.confidence -eq "INFERRED" }
    $queuePath = Join-Path $OutFull "business-logic-queue.json"
    if ($inferred.Count -eq 0) {
        Write-Host "[plaesy-graph] No INFERRED edges - nothing queued." -ForegroundColor Yellow
        return
    }
    $nodeIndex = @{}
    foreach ($n in $graph.nodes) { $nodeIndex[$n.id] = $n }
    $queue = $inferred | ForEach-Object {
        $src = $nodeIndex[$_.source]
        $tgt = $nodeIndex[$_.target]
        @{
            source = $_.source
            source_layer = $src.layer ?? "untyped"
            source_type = $src.type
            target = $_.target
            target_layer = $tgt.layer ?? "untyped"
            target_type = $tgt.type
            type = $_.type
            question = "What is the business logic connecting $($_.source) → $($_.target)?"
        }
    }
    $queue | ConvertTo-Json -Depth 4 | Set-Content -Path $queuePath -Encoding UTF8
    Write-Host "[plaesy-graph] Wrote $($queue.Count) edges with business context to $queuePath" -ForegroundColor Cyan
    Write-Host "[plaesy-graph] For each pair, explain the business logic in the 'answer' field, then save as business-logic-annotations.json" -ForegroundColor Cyan
}

# Feature #3: Learning Path Generator
function Invoke-LearningPathGenerator {
    $graph = Load-Graph
    $pathsPath = Join-Path $OutFull "learning-paths.json"

    # Topological sort via Kahn's algorithm
    $inDegree = @{}
    $adj = @{}
    foreach ($n in $graph.nodes) {
        $inDegree[$n.id] = 0
        $adj[$n.id] = @()
    }
    foreach ($e in $graph.edges) {
        if ($e.type -in @("imports", "sources", "calls")) {
            $inDegree[$e.target]++
            $adj[$e.source] += $e.target
        }
    }

    $queueList = New-Object System.Collections.Generic.List[object]
    @($graph.nodes | Where-Object { $inDegree[$_.id] -eq 0 } | Sort-Object { $_.degree } -Descending) | ForEach-Object { $queueList.Add($_) }
    $sorted = @()
    while ($queueList.Count -gt 0) {
        $node = $queueList[0]
        $queueList.RemoveAt(0)
        $sorted += $node.id
        foreach ($neighbor in $adj[$node.id]) {
            $inDegree[$neighbor]--
            if ($inDegree[$neighbor] -eq 0) {
                $nnode = $graph.nodes | Where-Object { $_.id -eq $neighbor }
                $queueList.Add($nnode)
            }
        }
    }

    $paths = @{
        entry_points = @($sorted | Select-Object -First 5)
        full_order = $sorted
        count = $sorted.Count
        generated_at = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
    }
    $paths | ConvertTo-Json -Depth 4 | Set-Content -Path $pathsPath -Encoding UTF8
    Write-Host "[plaesy-graph] Generated learning path with $($sorted.Count) ordered nodes to $pathsPath" -ForegroundColor Green
}

# Feature #4: Fuzzy/Semantic Search
function Invoke-FuzzyQuery {
    param([string]$Q, [int]$Depth = 2)
    $graph = Load-Graph
    $nodeMap = @{}
    foreach ($n in $graph.nodes) { $nodeMap[$n.id] = $n }

    # Fuzzy match
    $matches = @($graph.nodes | Where-Object { $_.id -ilike "*$Q*" -or $_.label -ilike "*$Q*" })
    if ($matches.Count -eq 0) { Write-Host "No matches for '$Q'" -ForegroundColor Yellow; return }

    Write-Host "[plaesy-graph] Fuzzy matches for '$Q':" -ForegroundColor Cyan
    foreach ($m in ($matches | Select-Object -First 10)) {
        Write-Host "  $($m.id) (layer: $($m.layer ?? 'n/a'), degree: $($m.degree))"
    }

    # Multi-hop BFS from first match
    $root = $matches[0]
    $visited = @{}
    $queueList = New-Object System.Collections.Generic.List[string]
    $queueList.Add($root.id)
    $depth = 0
    Write-Host "`nNeighbors (up to $Depth hops):" -ForegroundColor Cyan
    while ($queueList.Count -gt 0 -and $depth -lt $Depth) {
        $nextQueue = New-Object System.Collections.Generic.List[string]
        foreach ($n in $queueList) {
            $visited[$n] = 1
            $edges = $graph.edges | Where-Object { $_.source -eq $n -or $_.target -eq $n }
            foreach ($e in $edges) {
                $other = if ($e.source -eq $n) { $e.target } else { $e.source }
                if (-not $visited[$other]) {
                    if (-not $nextQueue.Contains($other)) { $nextQueue.Add($other) }
                    Write-Host "  [$($e.type)] $($e.source) → $($e.target)"
                }
            }
        }
        $queueList = $nextQueue
        $depth++
    }
}

function Invoke-ApplySemantic {
    param([string]$AnnotationsPath)
    if (-not (Test-Path $AnnotationsPath)) {
        Write-Host "[plaesy-graph] Annotations file not found: $AnnotationsPath" -ForegroundColor Yellow
        return
    }
    $graph = Load-Graph
    $annotations = Get-Content $AnnotationsPath -Raw | ConvertFrom-Json
    $applied = 0
    foreach ($a in $annotations) {
        $edge = $graph.edges | Where-Object { $_.source -eq $a.source -and $_.target -eq $a.target } | Select-Object -First 1
        if ($edge) {
            $edge | Add-Member -NotePropertyName rationale -NotePropertyValue $a.rationale -Force
            $applied++
        }
    }
    $graph | ConvertTo-Json -Depth 6 | Set-Content -Path $GraphJsonPath -Encoding UTF8
    Write-GraphReport -Graph $graph
    Write-Host "[plaesy-graph] Applied rationale to $applied edge(s). reports.md and project.graph.json updated." -ForegroundColor Green
}

# Feature #5: Impact Check Visualization
function Invoke-ImpactCheckVisualize {
    param([string]$NodeName, [int]$Depth = 2)
    $graph = Load-Graph
    $root = $graph.nodes | Where-Object { $_.id -ilike "*$NodeName*" } | Select-Object -First 1
    if (-not $root) { Write-Host "Node '$NodeName' not found" -ForegroundColor Yellow; return }

    $visited = @{}
    $queueList = New-Object System.Collections.Generic.List[object]
    $queueList.Add(@{ node = $root.id; depth = 0 })
    $impact = @()
    while ($queueList.Count -gt 0) {
        $current = $queueList[0]
        $queueList.RemoveAt(0)
        if ($visited[$current.node] -or $current.depth -gt $Depth) { continue }
        $visited[$current.node] = 1
        $node = $graph.nodes | Where-Object { $_.id -eq $current.node }
        $impact += @{ id = $current.node; depth = $current.depth; type = $node.type; layer = $node.layer ?? "n/a" }
        $edges = $graph.edges | Where-Object { $_.source -eq $current.node -or $_.target -eq $current.node }
        foreach ($e in $edges) {
            $next = if ($e.source -eq $current.node) { $e.target } else { $e.source }
            if (-not $visited[$next]) { $queueList.Add(@{ node = $next; depth = $current.depth + 1 }) }
        }
    }

    # Generate HTML visualization
    $htmlPath = Join-Path $OutFull "impact-visualization.html"
    $colors = @{ API = "#4CAF50"; Service = "#2196F3"; Data = "#FF9800"; UI = "#E91E63"; Utility = "#9C27B0"; "n/a" = "#999" }
    $html = @"
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Impact Analysis: $($root.id)</title></head>
<style>
body { font-family: Arial; margin: 20px; background: #f5f5f5 }
.impact-tree { background: white; padding: 20px; border-radius: 8px }
.depth-0 { margin-left: 0; background: #e3f2fd; padding: 10px; border-left: 4px solid #2196F3 }
.depth-1 { margin-left: 20px; background: #f3e5f5; padding: 8px; border-left: 2px solid #666 }
.depth-2 { margin-left: 40px; background: #fce4ec; padding: 6px; border-left: 2px dashed #999 }
.layer { display: inline-block; padding: 2px 8px; border-radius: 3px; color: white; font-size: 12px; margin-left: 10px }
</style>
<body>
<h1>Impact Analysis: $($root.label)</h1>
<p>Root node: <code>$($root.id)</code> | Depth: $Depth | Affected: $($impact.Count) nodes</p>
<div class="impact-tree">
"@
    foreach ($item in ($impact | Sort-Object depth)) {
        $layer = $item.layer
        $color = $colors[$layer]
        $html += "<div class='depth-$($item.depth)'><code>$($item.id)</code><span class='layer' style='background:$color'>$layer</span></div>"
    }
    $html += "</div></body></html>"
    $html | Set-Content -Path $htmlPath -Encoding UTF8
    Write-Host "[plaesy-graph] Impact visualization saved to impact-visualization.html" -ForegroundColor Green
    Write-Host "[plaesy-graph] Affected nodes: $($impact.Count)" -ForegroundColor Cyan
}

if ($Watch) { Invoke-Watch }
elseif ($Explain) { Invoke-Explain -NodeName $Explain }
elseif ($ImpactCheck) {
    if ($ImpactVisualize) {
        Invoke-ImpactCheckVisualize -NodeName $ImpactCheck -Depth $ImpactDepth
    } else {
        Invoke-ImpactCheck -NodeName $ImpactCheck -Depth $ImpactDepth
    }
}
elseif ($PathFrom -and $PathTo) { Invoke-PathQuery -From $PathFrom -To $PathTo }
elseif ($BusinessLogic) { Invoke-BusinessLogicQueue }
elseif ($GenerateLearningPaths) { Invoke-LearningPathGenerator }
elseif ($FuzzyQuery) { Invoke-FuzzyQuery -Q $FuzzyQuery -Depth $SearchDepth }
elseif ($SemanticQueue) { Invoke-SemanticQueue }
elseif ($ApplySemantic) { Invoke-ApplySemantic -AnnotationsPath $ApplySemantic }
elseif ($Query) { Invoke-Query -Q $Query }
elseif ($IfChanged) {
    $currentFp = Get-SourceFingerprint
    $storedFp = if (Test-Path $FingerprintPath) { (Get-Content -Path $FingerprintPath -Raw).Trim() } else { "" }
    if ((Test-Path $GraphJsonPath) -and $storedFp -and ($currentFp -eq $storedFp)) {
        Write-Host "[plaesy-graph] No source changes detected since last build - skipping rebuild." -ForegroundColor Yellow
        Write-Host "[plaesy-graph] Output: $OutFull" -ForegroundColor Yellow
    } else {
        Build-Graph
    }
}
else { Build-Graph }
