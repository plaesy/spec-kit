---
title: "Analyzer Implementation Fixes 2026-09-22"
updatedAt: "2026-09-22T16:06:00Z"
---

# Analyzer Implementation Fixes — 2026-09-22

**Context**: After multi-dimensional assessment found analyzer gaps (65/100 technical score), implemented TDD fixes across bash and PowerShell analyzers.

## Changes Made

### 1. Portable `find -printf` in `plaesy-graph.sh`
- **File**: `scripts/bash/plaesy-graph.sh:611-643`
- **Problem**: `find -printf '%T@'` is GNU-only; breaks on macOS/BSD
- **Fix**: Added `detect_stat_format()` that checks `stat -c %Y` (GNU) vs `stat -f %m` (BSD) once at load time. `source_fingerprint()` uses GNU `-printf` when available, falls back to `find -exec stat -f '%m'` on macOS.

### 2. `--if-changed` Fast Path in `plaesy-analyze.sh`
- **File**: `scripts/bash/plaesy-analyze.sh:1550-1585`
- **Problem**: Analyzer always regenerated all output files even when nothing changed
- **Fix**: Added `analyze_fingerprint()` (file count + newest mtime + framework version, excluding `.plaesy/`) and `should_skip_analysis()` that compares against `.analysis-fingerprint`. When `--if-changed` is passed and fingerprint matches, skips all regeneration.

### 3. `-IfChanged` Fast Path in `plaesy-analyze.ps1`
- **File**: `scripts/powershell/plaesy-analyze.ps1:1189-1233`
- **Problem**: Same as bash — no skip capability
- **Fix**: Added `Get-AnalyzeFingerprint()` and `Test-AnalysisUnchanged()` functions. Added `-IfChanged` parameter. Mirrors bash behavior.

### 4. PowerShell Analyzer Dev Tools/Build Systems
- **File**: `scripts/powershell/plaesy-analyze.ps1:860-970`
- **Problem**: PowerShell analyzer hardcoded `@("Git")` and `@("Manual")` for dev tools and build systems, while bash had full detection
- **Fix**: Added `Get-DevelopmentTools()` and `Get-BuildSystems()` functions mirroring bash `detect_development_tools`/`detect_build_systems`. Now detects package managers, CI/CD tools, testing frameworks, linting, Docker, and build systems.

### 5. Functional Tests
- **New files**: `testing/smoke/smoke-analyze.sh`, `testing/smoke/smoke-analyze.ps1`
- **Coverage**: 4 tests each — basic artifact generation, `--if-changed` skip, `--force` override, source modification invalidation
- **Result**: All 4 tests pass on both bash and PowerShell

### 6. CI Wiring
- **File**: `.github/workflows/ci.yml`
- **Added jobs**:
  - `analyzer-parity`: Runs both analyzers on same fixture, compares framework/language detection
  - Extended `smoke-test` job with `smoke-analyze.sh`
  - Extended `powershell-smoke` job with `smoke-analyze.ps1`

## Test Results
- `smoke-analyze.sh`: 4/4 PASS
- `smoke-e2e.sh`: PASS
- `testing/bash/run.sh`: 7/7 PASS
- PowerShell syntax: OK
- Bash syntax: OK

## Key Bug Found and Fixed
The `find` command's `! -path` exclusions were initially placed inside the `\( \)` OR-group, making them part of the OR chain (always matching). Fixed by placing them outside the group so they AND with the extension filter.