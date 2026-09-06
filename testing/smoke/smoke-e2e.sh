#!/bin/bash

# Plaesy Spec-Kit E2E Smoke Test
# Runs the dependency graph + analyze pipeline on a fixture project and
# validates the generated artifacts. Designed to run on GitHub Actions
# (Linux) where subprocess spawn is cheap; also works under MSYS/Windows
# but with generous timeouts due to ~200ms-per-spawn overhead.

set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK="${TMPDIR:-/tmp}/plaesy-smoke-$$"
FIXTURE="$WORK/project"
FAILURES=0

cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

mkdir -p "$FIXTURE/src" "$FIXTURE/docs"

cat > "$FIXTURE/README.md" <<'EOF'
# Smoke Fixture

See [docs/guide.md](docs/guide.md) and [src/lib.js](src/lib.js).
EOF

cat > "$FIXTURE/docs/guide.md" <<'EOF'
# Guide

Imports [src/lib.js](src/lib.js) and depends on [../README.md](README.md).
EOF

cat > "$FIXTURE/src/lib.js" <<'EOF'
import { helper } from "./helper.js";
export function lib() { return helper(); }
EOF

cat > "$FIXTURE/src/helper.js" <<'EOF'
export function helper() { return "ok"; }
EOF

cat > "$FIXTURE/src/app.js" <<'EOF'
import { lib } from "./lib.js";
console.log(lib());
EOF

# shellcheck disable=SC2016
cat > "$FIXTURE/run.sh" <<'EOF'
#!/bin/bash
# shellcheck source=src/lib.js
source src/helper.sh
EOF

check() {
    local desc="$1"
    shift
    if "$@" >/dev/null 2>&1; then
        echo "  PASS: $desc"
    else
        echo "  FAIL: $desc"
        FAILURES=$((FAILURES + 1))
    fi
}

echo "[1/3] Running plaesy-graph.sh on fixture..."
timeout 120 bash "$ROOT/scripts/bash/plaesy-graph.sh" --path "$FIXTURE" >/dev/null 2>&1 \
    || { echo "  FAIL: plaesy-graph.sh exited non-zero"; FAILURES=$((FAILURES + 1)); }

ANALYSIS="$FIXTURE/.plaesy/memory/analysis"

check "project.graph.json exists" test -f "$ANALYSIS/project.graph.json"
check "reports.md exists" test -f "$ANALYSIS/reports.md"
check "project.html exists" test -f "$ANALYSIS/project.html"

if [[ -f "$ANALYSIS/project.graph.json" ]]; then
    if command -v python3 >/dev/null 2>&1; then
        python3 - "$ANALYSIS/project.graph.json" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
nodes = d.get("nodes", [])
edges = d.get("edges", [])
assert isinstance(nodes, list) and nodes, "no nodes in graph"
assert isinstance(edges, list) and len(edges) > 0, "no edges in graph"
node_ids = {n["id"] for n in nodes}
for e in edges:
    assert e["source"] in node_ids and e["target"] in node_ids, \
        f"edge references unknown node: {e}"
print(f"  PASS: graph JSON valid ({len(nodes)} nodes, {len(edges)} edges)")
PY
        RC=$?
        if [[ $RC -ne 0 ]]; then FAILURES=$((FAILURES + 1)); fi
    else
        echo "  INFO: python3 not found, skipping JSON schema validation"
        check "graph JSON contains nodes" grep -q '"nodes"' "$ANALYSIS/project.graph.json"
        check "graph JSON contains edges" grep -q '"edges"' "$ANALYSIS/project.graph.json"
    fi
fi

echo "[2/3] Running --query smoke..."
if timeout 30 bash "$ROOT/scripts/bash/plaesy-graph.sh" --path "$FIXTURE" --query "lib.js" >/dev/null 2>&1; then
    echo "  PASS: --query executed"
else
    echo "  FAIL: --query exited non-zero"
    FAILURES=$((FAILURES + 1))
fi

echo "[3/3] Running plaesy-analyze.sh on fixture..."
if timeout 180 bash "$ROOT/scripts/bash/plaesy-analyze.sh" "$FIXTURE" >/dev/null 2>&1; then
    echo "  PASS: plaesy-analyze completed"
    check "project.json exists" test -f "$FIXTURE/.plaesy/analysis/project.json"
    check "context.md exists" test -f "$FIXTURE/.plaesy/context.md"
    check "memory.md exists" test -f "$FIXTURE/.plaesy/memory.md"
else
    echo "  FAIL: plaesy-analyze exited non-zero"
    FAILURES=$((FAILURES + 1))
fi

echo ""
if [[ $FAILURES -eq 0 ]]; then
    echo "SMOKE TEST PASSED"
    exit 0
else
    echo "SMOKE TEST FAILED: $FAILURES check(s)"
    exit 1
fi
