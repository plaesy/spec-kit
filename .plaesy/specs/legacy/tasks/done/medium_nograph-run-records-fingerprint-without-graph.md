---
title: plaesy analyze --no-graph records the analysis fingerprint although it writes no graph
phase: fix
status: backlog
createdAt: "2026-10-04T10:47:00.000Z"
updatedAt: "2026-10-10T02:45:00.000Z"
---

## Description

`analyze.Run` records `.analysis-fingerprint` whenever `graphFailed` is false —
and `graphFailed` is never set on the `--no-graph` path, because with no graph
there is nothing to fail. So:

```bash
plaesy analyze --no-graph        # writes project.json, project.structure.json, overview.md
plaesy analyze                   # fingerprint matches → skip path → prints
                                  #   - project.graph.json - Dependency graph (cached)
                                  #   - project.symbols.md - Function/class index (cached)
```

Those two files were never written by the first run, and the second run reports
them as cached without building them. `--force` is the only way out.

This is the same failure class the code comments two functions above already
describe for the graph-failure path: "a run whose graph build failed must not
leave a fingerprint behind, or the fast path would report a graph that was never
built as cached forever." `--no-graph` reaches the same state through a
different door.

## Design Rationale (ADR)
**Fingerprint Must Reflect What Was Built:** The fingerprint is a claim about on-disk artifacts. Current format stores only `{count, maxMtime, version}`. It needs a `components` field listing which sub-systems were built:

```json
{
  "count": 1234,
  "maxMtime": "2026-10-10T02:00:00Z",
  "version": 2,
  "components": ["structure", "graph", "symbols", "embeddings"]
}
```

**Skip Logic:** A run is skippable only if:
1. Fingerprint matches (count + maxMtime + version)
2. All components in current run's plan are present in fingerprint's `components`

**Backward Compatibility:** Old fingerprints (no `components` field) are treated as:
- If version < 2: treat as "built everything" → requires `--force` to re-run any component
- OR: treat as "rebuild required" (safer default)

Decision: **Treat old fingerprints as "rebuild required"** — safer, avoids silent cache hits on missing artifacts. Users get one `--force` run, then new format takes over.

**--no-graph Behavior:** 
- Does not record fingerprint at all (trades fast path for correctness)
- OR: records fingerprint with `components: ["structure", "symbols", "embeddings"]` (no "graph")
- Decision: **Do not record fingerprint on `--no-graph`** — simpler, documented as "full analysis each time"

## Deliverables
1. `internal/analyze/fingerprint.go` - Extend `Fingerprint` struct with `Components []string`
2. `internal/analyze/analyze.go` - Update skip logic to check components; don't write fingerprint on `--no-graph`
3. `internal/analyze/fingerprint_test.go` - Tests for component-aware skip logic, backward compat
4. Decision record: `.plaesy/decisions/fingerprint-component-tracking.md`
5. Migration: old fingerprints auto-upgraded on next full run

## Acceptance Criteria
- `plaesy analyze --no-graph` does NOT write `.analysis-fingerprint` (verified by file absence)
- `plaesy analyze` after `--no-graph` runs full graph + symbols (not skipped)
- Fingerprint v2 includes `components` array
- Old fingerprint (v1, no components) triggers full rebuild on next run
- `plaesy analyze --force` always works and writes v2 fingerprint
- Skip path correctly reports which components are cached vs rebuilt
- No regression: `go test ./internal/analyze/...` passes

## Test Plan
```bash
# Unit tests
go test ./internal/analyze/... -v -run TestFingerprint

# Integration tests
# 1. Fresh repo - full analyze
rm -rf .plaesy/analysis
plaesy analyze
# Verify: .analysis-fingerprint exists with components: [structure, graph, symbols, embeddings]

# 2. --no-graph run
plaesy analyze --no-graph
# Verify: .analysis-fingerprint NOT updated (same mtime or absent)
# Verify: project.graph.json NOT created

# 3. Full analyze after --no-graph
plaesy analyze
# Verify: graph + symbols rebuilt (not skipped)
# Verify: output shows "Building dependency graph..." not "cached"

# 4. Old fingerprint compat
# Create v1 fingerprint (no components)
cat > .plaesy/analysis/.analysis-fingerprint <<'EOF'
{"count":100,"maxMtime":"2026-01-01T00:00:00Z","version":1}
EOF
plaesy analyze
# Verify: full rebuild (not skipped)
# Verify: new fingerprint has version=2 and components

# 5. Normal incremental (no changes)
plaesy analyze
# Verify: all components skipped (cached)

# Regression
go test ./internal/analyze/... -count=1
```

## Dependencies
- None (self-contained in `internal/analyze`)

## Config Migration
- Fingerprint format version bump: v1 → v2
- Automatic upgrade: v1 fingerprints trigger rebuild, then v2 written
- No user action required

## Operations
- One-time cost: all existing projects require one `--force` or full analyze after upgrade
- Fingerprint file size: ~200 bytes (negligible)
- No performance impact

## Estimated Effort
Small (1 day)

## Definition of Done
- All deliverables implemented and tested
- Decision record created
- Backward compatibility verified with v1 fingerprint
- Documentation: update `docs/analyze-fingerprint.md`