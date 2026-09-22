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

mkdir -p "$FIXTURE/src" "$FIXTURE/docs" "$FIXTURE/.plaesy"

cat > "$FIXTURE/README.md" <<'EOF'
# Smoke Fixture

See [docs/guide.md](docs/guide.md) and [src/lib.js](src/lib.js).
EOF

cat > "$FIXTURE/package.json" <<'EOF'
{
  "name": "plaesy-smoke-fixture",
  "version": "1.0.0"
}
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

printf '%s\n' 'context sentinel: do not modify' > "$FIXTURE/.plaesy/context.md"
printf '%s\n' 'memory sentinel: do not modify' > "$FIXTURE/.plaesy/memory.md"
cp "$FIXTURE/.plaesy/context.md" "$WORK/context.md.before"
cp "$FIXTURE/.plaesy/memory.md" "$WORK/memory.md.before"

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

validate_analysis_contract() {
    local project_json="$1"
    local structure_json="$2"
    local overview_md="$3"

    if command -v python3 >/dev/null 2>&1; then
        if python3 - "$project_json" "$structure_json" "$overview_md" <<'PY'
import json
import sys
from pathlib import Path

project = json.load(open(sys.argv[1], encoding="utf-8"))
structure = json.load(open(sys.argv[2], encoding="utf-8"))
overview = Path(sys.argv[3]).read_text(encoding="utf-8")

project_files = project["structure"]["files"]
structure_types = structure["file_types"]
counts = {
    "project total_files": project["project_summary"]["total_files"],
    "project total count": project_files["total"],
    "project code count": project_files["code"],
    "project documentation count": project_files["documentation"],
    "project configuration count": project_files["configuration"],
    "structure source_code count": structure_types["source_code"],
    "structure documentation count": structure_types["documentation"],
    "structure configuration count": structure_types["configuration"],
}
for label, value in counts.items():
    if type(value) is not int or value <= 0:
        raise AssertionError(f"{label} must be a positive integer")

if project_files["total"] != project["project_summary"]["total_files"]:
    raise AssertionError("project file totals do not match")
if project_files["code"] != structure_types["source_code"]:
    raise AssertionError("source code counts do not match")
if project_files["documentation"] != structure_types["documentation"]:
    raise AssertionError("documentation counts do not match")
if project_files["configuration"] != structure_types["configuration"]:
    raise AssertionError("configuration counts do not match")
if not overview.startswith("# Project Analysis Overview"):
    raise AssertionError("overview.md has an unexpected format")
PY
        then
            echo "  PASS: analyzer JSON contract and nonzero counts valid"
        else
            echo "  FAIL: analyzer JSON contract or nonzero counts"
            FAILURES=$((FAILURES + 1))
        fi
    elif command -v jq >/dev/null 2>&1; then
        if jq -e -s '
            (.[0] | (has("project_summary") and has("structure") and has("analysis_files"))) and
            (.[1] | (has("directories") and has("key_files") and has("file_types") and has("related_files"))) and
            (.[0].project_summary.total_files > 0) and
            (.[0].structure.files.total > 0) and
            (.[0].structure.files.code > 0) and
            (.[0].structure.files.documentation > 0) and
            (.[0].structure.files.configuration > 0) and
            (.[1].file_types.source_code > 0) and
            (.[1].file_types.documentation > 0) and
            (.[1].file_types.configuration > 0) and
            (.[0].project_summary.total_files == .[0].structure.files.total) and
            (.[0].structure.files.code == .[1].file_types.source_code) and
            (.[0].structure.files.documentation == .[1].file_types.documentation) and
            (.[0].structure.files.configuration == .[1].file_types.configuration)
        ' "$project_json" "$structure_json" >/dev/null 2>&1 &&
            grep -q '^# Project Analysis Overview$' "$overview_md"; then
            echo "  PASS: analyzer JSON contract and nonzero counts valid"
        else
            echo "  FAIL: analyzer JSON contract or nonzero counts"
            FAILURES=$((FAILURES + 1))
        fi
    elif command -v node >/dev/null 2>&1; then
        if node - "$project_json" "$structure_json" "$overview_md" <<'JS'
const fs = require("fs");
const assert = require("assert");

const project = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const structure = JSON.parse(fs.readFileSync(process.argv[3], "utf8"));
const overview = fs.readFileSync(process.argv[4], "utf8");
const projectFiles = project.structure.files;
const structureTypes = structure.file_types;
const counts = [
    project.project_summary.total_files,
    projectFiles.total,
    projectFiles.code,
    projectFiles.documentation,
    projectFiles.configuration,
    structureTypes.source_code,
    structureTypes.documentation,
    structureTypes.configuration,
];
for (const value of counts) {
    assert(Number.isInteger(value) && value > 0, "analyzer counts must be positive integers");
}
assert.strictEqual(projectFiles.total, project.project_summary.total_files);
assert.strictEqual(projectFiles.code, structureTypes.source_code);
assert.strictEqual(projectFiles.documentation, structureTypes.documentation);
assert.strictEqual(projectFiles.configuration, structureTypes.configuration);
assert(overview.startsWith("# Project Analysis Overview"));
JS
        then
            echo "  PASS: analyzer JSON contract and nonzero counts valid"
        else
            echo "  FAIL: analyzer JSON contract or nonzero counts"
            FAILURES=$((FAILURES + 1))
        fi
    else
        echo "  INFO: no JSON parser found, skipping analyzer JSON contract validation"
    fi
}

echo "[1/3] Running plaesy-graph.sh on fixture..."
timeout 120 bash "$ROOT/scripts/bash/plaesy-graph.sh" --path "$FIXTURE" >/dev/null 2>&1 \
    || { echo "  FAIL: plaesy-graph.sh exited non-zero"; FAILURES=$((FAILURES + 1)); }

ANALYSIS="$FIXTURE/.plaesy/analysis"

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
    check "project.json exists" test -f "$ANALYSIS/project.json"
    check "project.structure.json exists" test -f "$ANALYSIS/project.structure.json"
    check "overview.md exists" test -f "$ANALYSIS/overview.md"
    check "test-runner.sh is not generated" test ! -e "$FIXTURE/scripts/test-runner.sh"
    check "context.md is not generated or modified" cmp -s "$FIXTURE/.plaesy/context.md" "$WORK/context.md.before"
    check "memory.md is not generated or modified" cmp -s "$FIXTURE/.plaesy/memory.md" "$WORK/memory.md.before"

    if [[ -f "$ANALYSIS/project.json" && -f "$ANALYSIS/project.structure.json" && -f "$ANALYSIS/overview.md" ]]; then
        validate_analysis_contract \
            "$ANALYSIS/project.json" \
            "$ANALYSIS/project.structure.json" \
            "$ANALYSIS/overview.md"
    fi
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
