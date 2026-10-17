---
title: internal/graph has three out-of-sync extension lists; .hh and .cxx are never collected
phase: implement
status: backlog
createdAt: "2026-10-03T14:38:00.000Z"
updatedAt: "2026-10-10T02:45:00.000Z"
---

## Description

`scripts/internal/graph` keeps the set of "interesting" file extensions in
**three separate places**, and they no longer agree:

1. `graph.go` `includeExtRE` — which files are collected into the graph:
   `md|ps1|sh|js|jsx|ts|tsx|py|go|dart|java|kt|kts|swift|c|h|cc|cpp|hpp|cs|rs|rb|php`
2. `graph.go` `classify()` → `"source"` — which files are labelled source:
   `js|jsx|ts|tsx|py|go|dart|java|kt|kts|swift|c|h|cc|cpp|hpp|cs|rs|rb|php`
3. `symbols.go` `symbolsOf()` — which extensions get symbol extraction,
   including `.c .h .cpp .hpp .cc .hh .cxx` for C-family.

**Concrete bug:** `.hh` and `.cxx` are handled by `symbolsOf` but are absent
from `includeExtRE`, so those files are never collected. No graph node can
exist for them, so that branch of `symbolsOf` is unreachable dead code, and
`semantic.docsByExt` can never fire for them either.

Secondary: `md`, `sh`, and `ps1` are collected and *do* get symbol extraction
(`mdSymbols`, `shSymbols`, and the `.ps1` case) but `classify()` labels them
`"other"` rather than `"source"`. That may well be intentional — a Markdown or
shell file arguably isn't "source" — but it is currently an undocumented
divergence rather than a decision anyone recorded.

## Design Rationale (ADR)
**Single Source of Truth:** Consolidate three lists into one exported variable:
```go
// SourceExtensions is the canonical list of extensions considered "source code"
var SourceExtensions = []string{
    ".js", ".jsx", ".ts", ".tsx",
    ".py", ".go", ".dart", ".java", ".kt", ".kts", ".swift",
    ".c", ".h", ".cc", ".cpp", ".hpp", ".cxx", ".hh",
    ".cs", ".rs", ".rb", ".php",
}
```

**Derived Lists:**
- `includeExtRE` = `SourceExtensions` + `.md` + `.sh` + `.ps1` (docs/scripts collected but not "source")
- `classify()` returns `"source"` iff extension in `SourceExtensions`
- `symbolsOf()` handles `SourceExtensions` + `.md` + `.sh` + `.ps1` (each has extractor)

**Classification Decision:**
- `"source"`: compilable/executable code with symbol extractors
- `"doc"`: Markdown, documentation
- `"script"`: shell, PowerShell, other scripts
- `"other"`: everything else collected but not classified

This makes the divergence explicit and documented.

## Deliverables
1. `internal/graph/extensions.go` - New file: `SourceExtensions` + derived helpers
2. `internal/graph/graph.go` - Replace `includeExtRE` and `classify()` with derived versions
3. `internal/graph/symbols.go` - Verify `symbolsOf()` covers all `SourceExtensions`
4. Tests: graph rebuild includes `.hh`/`.cxx` files, classification correct
5. Decision record: `.plaesy/decisions/graph-extension-consolidation.md`

## Acceptance Criteria
- `.hh` and `.cxx` files are collected into graph (verified by `plaesy analyze --force` on test repo)
- Graph nodes for `.hh`/`.cxx` have symbols extracted (verified by `project.symbols.md`)
- Three extension lists consolidated to single source (`extensions.go`)
- Classification documented: `"source"` vs `"doc"` vs `"script"` vs `"other"`
- No regression: `go test ./internal/graph/...` passes
- Graph rebuild on this repo shows no new errors/warnings

## Test Plan
```bash
# Unit tests
go test ./internal/graph/... -v -run TestExtensions
go test ./internal/graph/... -v -run TestClassify

# Integration test
# 1. Create test repo with .hh/.cxx files
mkdir -p /tmp/test-graph/{src,include}
echo 'class Foo { void bar(); };' > /tmp/test-graph/include/foo.hh
echo '#include "foo.hh"' > /tmp/test-graph/src/main.cxx
echo 'void Foo::bar() {}' >> /tmp/test-graph/src/main.cxx

# 2. Run analyze
cd /tmp/test-graph
plaesy analyze --force

# 3. Verify
grep -q "foo.hh" .plaesy/analysis/project.symbols.md
grep -q "main.cxx" .plaesy/analysis/project.symbols.md
grep -q "Foo" .plaesy/analysis/project.symbols.md
grep -q "bar" .plaesy/analysis/project.symbols.md

# 4. Verify classification
# project.graph.json should have nodes with type "source" for .hh/.cxx
jq '.nodes[] | select(.path | endswith(".hh") or endswith(".cxx")) | .type' .plaesy/analysis/project.graph.json
# Should output "source"

# Regression
go test ./internal/graph/... -count=1
```

## Dependencies
- None (self-contained in `internal/graph`)

## Config Migration
- No config changes required
- Behavior change: `.hh`/`.cxx` now collected (was silently omitted)
- Existing projects: next `plaesy analyze --force` picks them up automatically

## Operations
- Blast radius: widening `includeExtRE` changes what `plaesy analyze` collects for **every** project
- Verify with graph rebuild on multiple test repos before merge
- No performance impact (regex alternation, same complexity)

## Estimated Effort
Small (1 day)

## Definition of Done
- All deliverables implemented and tested
- Decision record created
- Test repo with `.hh`/`.cxx` verified end-to-end
- Documentation: update `docs/graph-extensions.md` (or create)