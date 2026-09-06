# plaesy-trim.ps1 - Token/context compression for the plaesy spec-kit repo
# Version: 1.1.0 - Compatible with PowerShell 5.1+
#
# Layer 1 (command output): run <command...>            - per-tool aware dedupe/truncate/filter
# Layer 2 (memory/instruction files): compress -Path ... - heuristic denser style (fast, local)
#   llm-queue / apply-llm                                 - LLM-quality rewrite via the calling
#                                                            assistant, no API key (same pattern as
#                                                            plaesy-graph's -SemanticQueue/-ApplySemantic)
# Report: report                                          - cumulative savings across layers
#
# Usage:
#   plaesy-trim.ps1 run git status
#   plaesy-trim.ps1 run cargo test
#   plaesy-trim.ps1 compress -Path .plaesy\memory\plaesy.md
#   plaesy-trim.ps1 compress -Path .plaesy\memory -Recurse -DryRun
#   plaesy-trim.ps1 compress -Path README.md -Level ultra
#   plaesy-trim.ps1 llm-queue -Path .plaesy\memory\plaesy.md
#   plaesy-trim.ps1 apply-llm -Path .plaesy\memory\plaesy.md -Annotations annotations.json
#   plaesy-trim.ps1 report

param(
    [Parameter(Position = 0)]
    [ValidateSet("run", "compress", "report", "llm-queue", "apply-llm")]
    [string]$Command,

    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Rest,

    [string]$Path,
    [ValidateSet("lite", "full", "ultra")]
    [string]$Level,
    [switch]$Recurse,
    [switch]$DryRun,
    [string]$Annotations
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$FrameworkRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
# $ProjectRoot is the project being worked on (current directory), NOT where this script is
# installed. Framework-shipped rules are read from $FrameworkRoot as a fallback only, so
# plaesy-trim works correctly when invoked from any other project via its installed path
# (e.g. %PLAESY_HOME%\scripts\powershell\plaesy-trim.ps1), matching plaesy-graph.ps1's behavior.
$ProjectRoot = (Get-Location).Path
$RepoRoot = $ProjectRoot
# Namespaced under .plaesy/ (per platform.json's "base_directory": ".plaesy" convention), not a
# bare "scripts/" folder - a target project may already have its own unrelated scripts/ dir.
$ProjectRulesPath = Join-Path $ProjectRoot ".plaesy\scripts\configs\plaesy-trim-rules.json"
$FrameworkRulesPath = Join-Path $FrameworkRoot "scripts\configs\plaesy-trim-rules.json"
$RulesPath = if (Test-Path $ProjectRulesPath) { $ProjectRulesPath } else { $FrameworkRulesPath }
$GraphPath = Join-Path $ProjectRoot ".plaesy\memory\analysis\project.graph.json"
$StatsPath = Join-Path $ProjectRoot ".plaesy\memory\token-stats.json"

function Get-Rules {
    if (-not (Test-Path $RulesPath)) { throw "Rules file not found: $RulesPath" }
    Get-Content $RulesPath -Raw | ConvertFrom-Json
}

function Estimate-Tokens([string]$Text) {
    [Math]::Ceiling($Text.Length / 4.0)
}

function Add-StatRecord([string]$Layer, [string]$TargetPath, [int]$Before, [int]$After) {
    $stats = @()
    if (Test-Path $StatsPath) {
        $raw = Get-Content $StatsPath -Raw
        if ($raw.Trim()) { $stats = @(ConvertFrom-Json $raw) }
    }
    $pct = if ($Before -gt 0) { [Math]::Round((1 - ($After / [double]$Before)) * 100, 1) } else { 0 }
    $record = [PSCustomObject]@{
        timestamp = (Get-Date -Format "o")
        layer     = $Layer
        path      = $TargetPath
        before    = $Before
        after     = $After
        saved_pct = $pct
    }
    $stats = @($stats) + $record
    $parent = Split-Path -Parent $StatsPath
    if (-not (Test-Path $parent)) { New-Item -ItemType Directory -Path $parent -Force | Out-Null }
    ($stats | ConvertTo-Json -Depth 5) | Set-Content -Path $StatsPath -Encoding UTF8
}

# ---------- Layer 1: command output compression ----------

function Compress-CommandOutput([string[]]$Lines, [psobject]$Rule) {
    $maxLines = if ($Rule) { $Rule.max_lines } else { 40 }
    $dedupe = if ($Rule) { $Rule.dedupe_consecutive } else { $true }
    $mode = if ($Rule -and $Rule.mode) { $Rule.mode } else { "generic" }

    if ($mode -eq "test-results") {
        $keepPattern = '(?i)\b(fail|failed|error|✗|✘)\b'
        $summaryPattern = '(?i)\b(passed|failed|tests? run|ok\.|error(s)?:)\b'
        $failing = @($Lines | Where-Object { $_ -match $keepPattern })
        $summary = @($Lines | Select-Object -Last 5 | Where-Object { $_ -match $summaryPattern })
        $Lines = @($failing + $summary | Select-Object -Unique)
        if ($Lines.Count -eq 0) { $Lines = @("(no failures - all tests passed, output suppressed)") }
    }

    $collapsed = New-Object System.Collections.Generic.List[string]
    if ($dedupe) {
        $prev = $null
        $count = 0
        foreach ($line in $Lines) {
            if ($line -eq $prev) {
                $count++
            } else {
                if ($prev -ne $null) {
                    if ($count -gt 1) { $collapsed.Add("$prev (x$count)") } else { $collapsed.Add($prev) }
                }
                $prev = $line
                $count = 1
            }
        }
        if ($prev -ne $null) {
            if ($count -gt 1) { $collapsed.Add("$prev (x$count)") } else { $collapsed.Add($prev) }
        }
    } else {
        $collapsed.AddRange($Lines)
    }

    if ($collapsed.Count -le $maxLines) { return $collapsed }

    $head = [Math]::Ceiling($maxLines / 2)
    $tail = $maxLines - $head
    $omitted = $collapsed.Count - $head - $tail
    $result = New-Object System.Collections.Generic.List[string]
    $result.AddRange($collapsed.GetRange(0, $head))
    $result.Add("... $omitted lines omitted ...")
    $result.AddRange($collapsed.GetRange($collapsed.Count - $tail, $tail))
    return $result
}

function Invoke-Run([string[]]$CommandParts) {
    if (-not $CommandParts -or $CommandParts.Count -eq 0) {
        throw "Usage: plaesy-trim.ps1 run <command...>"
    }
    $rules = Get-Rules
    $cmdString = $CommandParts -join " "

    $matchedKey = $null
    foreach ($key in $rules.layer1_commands.PSObject.Properties.Name) {
        if ($key -eq "_default") { continue }
        if ($cmdString.StartsWith($key)) {
            if (-not $matchedKey -or $key.Length -gt $matchedKey.Length) { $matchedKey = $key }
        }
    }
    $rule = if ($matchedKey) { $rules.layer1_commands.$matchedKey } else { $rules.layer1_commands._default }

    $exe = $CommandParts[0]
    $exeArgs = @($CommandParts | Select-Object -Skip 1)
    $rawOutput = & $exe @exeArgs 2>&1 | Out-String
    $lines = $rawOutput -split "`r?`n"

    $compressedLines = Compress-CommandOutput -Lines $lines -Rule $rule
    $compressedText = $compressedLines -join "`n"

    $before = Estimate-Tokens $rawOutput
    $after = Estimate-Tokens $compressedText
    Add-StatRecord -Layer "layer1" -TargetPath $cmdString -Before $before -After $after

    $compressedText
}

# ---------- Layer 2: prose compression ----------

function Get-NodeDegree([string]$RelativePath) {
    if (-not (Test-Path $GraphPath)) { return $null }
    if (-not $script:GraphCache) {
        $script:GraphCache = Get-Content $GraphPath -Raw | ConvertFrom-Json
    }
    $normalized = $RelativePath -replace "/", "\"
    $node = $script:GraphCache.nodes | Where-Object { $_.id -eq $normalized }
    if ($node) { return $node.degree }
    return $null
}

function Resolve-Level([string]$RelativePath, [string]$Requested, [psobject]$Rules) {
    if ($Requested) { return $Requested }
    $degree = Get-NodeDegree $RelativePath
    if ($null -eq $degree) { return "full" }
    if ($degree -ge $Rules.layer2_degree_thresholds.lite_min_degree) { return "lite" }
    if ($degree -ge $Rules.layer2_degree_thresholds.full_min_degree) { return "full" }
    return "ultra"
}

function Get-FenceSegments([string]$Text) {
    # Line-by-line fence toggle (matches plaesy-trim.sh's compress_prose logic) instead of a
    # regex Split, which mis-pairs adjacent ``` blocks. Each returned segment ends at (and
    # includes) the fence marker line that closes it; isCode=true segments are the code content
    # between an opening and closing fence and are never touched by prose transforms.
    $lines = [regex]::Split($Text, '(?<=\n)')
    $segments = New-Object System.Collections.Generic.List[psobject]
    $inCode = $false
    $buf = New-Object System.Text.StringBuilder
    foreach ($line in $lines) {
        [void]$buf.Append($line)
        if ($line.TrimStart() -match '^```') {
            $segments.Add([PSCustomObject]@{ isCode = $inCode; text = $buf.ToString() })
            $buf = New-Object System.Text.StringBuilder
            $inCode = -not $inCode
        }
    }
    if ($buf.Length -gt 0) {
        $segments.Add([PSCustomObject]@{ isCode = $inCode; text = $buf.ToString() })
    }
    return $segments
}

function Compress-Prose([string]$Text, [string]$LevelToUse, [psobject]$Rules) {
    $segments = Get-FenceSegments $Text
    $out = New-Object System.Text.StringBuilder
    foreach ($seg in $segments) {
        if ($seg.isCode) {
            [void]$out.Append($seg.text)
            continue
        }
        $piece = $seg.text
        if ($LevelToUse -in @("lite", "full", "ultra")) {
            foreach ($p in $Rules.layer2_filler_patterns) {
                $piece = [regex]::Replace($piece, $p.find, $p.replace, "IgnoreCase")
            }
        }
        if ($LevelToUse -in @("full", "ultra")) {
            $piece = [regex]::Replace($piece, " {2,}", " ")
            $piece = [regex]::Replace($piece, "(?m)^\s+", "")
        }
        if ($LevelToUse -eq "ultra") {
            $piece = [regex]::Replace($piece, "\b(that|which|very|really|basically|simply)\b\s*", "", "IgnoreCase")
        }
        $piece = [regex]::Replace($piece, "\n{3,}", "`n`n")
        [void]$out.Append($piece)
    }
    return $out.ToString()
}

function Invoke-CompressFile([string]$FilePath, [string]$RequestedLevel, [psobject]$Rules, [bool]$IsDryRun) {
    $original = Get-Content $FilePath -Raw
    $relative = (Resolve-Path $FilePath).Path.Substring($RepoRoot.Length + 1)
    $level = Resolve-Level -RelativePath $relative -Requested $RequestedLevel -Rules $Rules
    $compressed = Compress-Prose -Text $original -LevelToUse $level -Rules $Rules

    $before = Estimate-Tokens $original
    $after = Estimate-Tokens $compressed
    $pct = if ($before -gt 0) { [Math]::Round((1 - ($after / [double]$before)) * 100, 1) } else { 0 }

    Write-Host "$relative [$level]"
    Write-Host "  $before tokens -> $after tokens  ($pct% saved)"

    if (-not $IsDryRun) {
        Copy-Item -Path $FilePath -Destination "$FilePath.bak" -Force
        Set-Content -Path $FilePath -Value $compressed -Encoding UTF8 -NoNewline
        Add-StatRecord -Layer "layer2" -TargetPath $relative -Before $before -After $after
    }
}

function Invoke-Compress {
    if (-not $Path) { throw "Usage: plaesy-trim.ps1 compress -Path <file|dir> [-Level lite|full|ultra] [-Recurse] [-DryRun]" }
    $rules = Get-Rules
    $target = Join-Path $RepoRoot $Path
    if (-not (Test-Path $target)) { $target = $Path }
    if (-not (Test-Path $target)) { throw "Path not found: $Path" }

    if ((Get-Item $target).PSIsContainer) {
        $files = if ($Recurse) { Get-ChildItem $target -Filter "*.md" -Recurse -File } else { Get-ChildItem $target -Filter "*.md" -File }
        foreach ($f in $files) {
            Invoke-CompressFile -FilePath $f.FullName -RequestedLevel $Level -Rules $rules -IsDryRun $DryRun.IsPresent
        }
    } else {
        Invoke-CompressFile -FilePath $target -RequestedLevel $Level -Rules $rules -IsDryRun $DryRun.IsPresent
    }
}

# ---------- Layer 2 (LLM mode): queue prose for the calling assistant to rewrite ----------

function Get-ProseSegments([string]$Text) {
    # Same fence-toggle split as Compress-Prose: code segments are never queued for rewrite.
    $parts = Get-FenceSegments $Text
    $segments = New-Object System.Collections.Generic.List[psobject]
    for ($i = 0; $i -lt $parts.Count; $i++) {
        if ($parts[$i].isCode) { continue }
        if ($parts[$i].text.Trim().Length -ge 40) {
            $segments.Add([PSCustomObject]@{ index = $i; text = $parts[$i].text })
        }
    }
    return $segments
}

function Invoke-LlmQueue {
    if (-not $Path) { throw "Usage: plaesy-trim.ps1 llm-queue -Path <file>" }
    $target = Join-Path $RepoRoot $Path
    if (-not (Test-Path $target)) { $target = $Path }
    if (-not (Test-Path $target)) { throw "Path not found: $Path" }

    $original = Get-Content $target -Raw
    $segments = Get-ProseSegments $original
    $relative = (Resolve-Path $target).Path.Substring($RepoRoot.Length + 1)

    $outDir = Join-Path $RepoRoot ".plaesy\memory\analysis"
    if (-not (Test-Path $outDir)) { New-Item -ItemType Directory -Path $outDir -Force | Out-Null }
    $queuePath = Join-Path $outDir "trim-queue.json"
    # ConvertTo-Json silently collapses a 1-element array to a bare object (Windows PowerShell 5.1
    # has no -AsArray) - force array syntax explicitly so "segments" is always a JSON array.
    $segmentsJson = if ($segments.Count -eq 0) {
        "[]"
    } elseif ($segments.Count -eq 1) {
        "[" + ($segments[0] | ConvertTo-Json -Depth 5 -Compress) + "]"
    } else {
        $segments | ConvertTo-Json -Depth 5
    }
    $fileJson = $relative | ConvertTo-Json
    "{`n  `"file`": $fileJson,`n  `"segments`": $segmentsJson`n}" | Set-Content -Path $queuePath -Encoding UTF8

    Write-Host "Wrote $($segments.Count) prose segment(s) from $relative to $queuePath"
    Write-Host "Next: have the calling assistant rewrite each segment.text denser (same meaning, no code/commands touched),"
    Write-Host "save as {file, segments:[{index, text}]} JSON, then run:"
    Write-Host "  plaesy-trim.ps1 apply-llm -Path $Path -Annotations <annotations.json>"
}

function Invoke-ApplyLlm {
    if (-not $Path -or -not $Annotations) { throw "Usage: plaesy-trim.ps1 apply-llm -Path <file> -Annotations <annotations.json>" }
    $target = Join-Path $RepoRoot $Path
    if (-not (Test-Path $target)) { $target = $Path }
    if (-not (Test-Path $target)) { throw "Path not found: $Path" }
    if (-not (Test-Path $Annotations)) { throw "Annotations file not found: $Annotations" }

    $original = Get-Content $target -Raw
    $parts = Get-FenceSegments $original
    $rewrites = @{}
    foreach ($s in (Get-Content $Annotations -Raw | ConvertFrom-Json).segments) {
        $rewrites[[int]$s.index] = $s.text
    }

    $out = New-Object System.Text.StringBuilder
    for ($i = 0; $i -lt $parts.Count; $i++) {
        if ($rewrites.ContainsKey($i)) { [void]$out.Append($rewrites[$i]) }
        else { [void]$out.Append($parts[$i].text) }
    }
    $compressed = $out.ToString()
    $relative = (Resolve-Path $target).Path.Substring($RepoRoot.Length + 1)

    $before = Estimate-Tokens $original
    $after = Estimate-Tokens $compressed
    $pct = if ($before -gt 0) { [Math]::Round((1 - ($after / [double]$before)) * 100, 1) } else { 0 }

    Copy-Item -Path $target -Destination "$target.bak" -Force
    Set-Content -Path $target -Value $compressed -Encoding UTF8 -NoNewline
    Add-StatRecord -Layer "layer2-llm" -TargetPath $relative -Before $before -After $after

    Write-Host "$relative [llm]"
    Write-Host "  $before tokens -> $after tokens  ($pct% saved)"
}

# ---------- Report ----------

function Invoke-Report {
    if (-not (Test-Path $StatsPath)) {
        Write-Host "No token stats yet. Run 'compress' or 'run' first."
        return
    }
    $stats = ConvertFrom-Json (Get-Content $StatsPath -Raw)
    $byLayer = $stats | Group-Object layer
    Write-Host "plaesy-trim report"
    Write-Host "===================="
    foreach ($group in $byLayer) {
        $totalBefore = ($group.Group | Measure-Object -Property before -Sum).Sum
        $totalAfter = ($group.Group | Measure-Object -Property after -Sum).Sum
        $pct = if ($totalBefore -gt 0) { [Math]::Round((1 - ($totalAfter / [double]$totalBefore)) * 100, 1) } else { 0 }
        Write-Host "$($group.Name): $($group.Count) runs, $totalBefore -> $totalAfter tokens ($pct% avg saved)"
    }
    $grandBefore = ($stats | Measure-Object -Property before -Sum).Sum
    $grandAfter = ($stats | Measure-Object -Property after -Sum).Sum
    $grandPct = if ($grandBefore -gt 0) { [Math]::Round((1 - ($grandAfter / [double]$grandBefore)) * 100, 1) } else { 0 }
    Write-Host "--------------------"
    Write-Host "total: $grandBefore -> $grandAfter tokens ($grandPct% saved)"
}

# ---------- Dispatch ----------

switch ($Command) {
    "run" { Invoke-Run -CommandParts $Rest }
    "compress" { Invoke-Compress }
    "llm-queue" { Invoke-LlmQueue }
    "apply-llm" { Invoke-ApplyLlm }
    "report" { Invoke-Report }
    default {
        Write-Host "Usage: plaesy-trim.ps1 <run|compress|llm-queue|apply-llm|report> [options]"
        Write-Host "  run <command...>                        Layer 1: per-tool aware compress of command output"
        Write-Host "  compress -Path <file|dir> [-Level lite|full|ultra] [-Recurse] [-DryRun]  Layer 2: fast heuristic prose compression"
        Write-Host "  llm-queue -Path <file>                  Layer 2 (LLM mode): export prose segments to rewrite"
        Write-Host "  apply-llm -Path <file> -Annotations <json>  Layer 2 (LLM mode): merge rewritten segments back"
        Write-Host "  report                                  Cumulative savings across layers"
    }
}
