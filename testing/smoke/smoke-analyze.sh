#!/bin/bash

# Plaesy Spec-Kit Analyzer Functional Smoke Test
# Validates that plaesy-analyze.sh produces expected output artifacts
# and that --if-changed correctly skips regeneration when nothing changed.

set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK="${TMPDIR:-/tmp}/plaesy-analyze-smoke-$$"
FIXTURE="$WORK/project"
FAILURES=0

cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

mkdir -p "$FIXTURE/src" "$FIXTURE/docs"

cat > "$FIXTURE/README.md" <<'EOF'
# Analyzer Smoke Fixture

See [docs/guide.md](docs/guide.md) and [src/lib.js](src/lib.js).
EOF

cat > "$FIXTURE/package.json" <<'EOF'
{
  "name": "plaesy-analyze-smoke-fixture",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0"
  }
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

echo "=== Test 1: Basic analysis produces expected artifacts ==="
bash "$ROOT/scripts/bash/plaesy-analyze.sh" "$FIXTURE" --no-graph

for f in project.json project.structure.json overview.md; do
    if [[ -f "$FIXTURE/.plaesy/analysis/$f" ]]; then
        echo "  PASS: $f exists"
    else
        echo "  FAIL: $f missing"
        FAILURES=$((FAILURES + 1))
    fi
done

# Validate project.json contains expected fields
if grep -q '"express"' "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null; then
    echo "  PASS: project.json detects express framework"
else
    echo "  FAIL: project.json missing express framework"
    FAILURES=$((FAILURES + 1))
fi

if grep -q '"JavaScript"' "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null; then
    echo "  PASS: project.json detects JavaScript language"
else
    echo "  FAIL: project.json missing JavaScript language"
    FAILURES=$((FAILURES + 1))
fi

echo ""
echo "=== Test 2: --if-changed skips regeneration when nothing changed ==="
# First run with --if-changed should regenerate (no prior fingerprint)
bash "$ROOT/scripts/bash/plaesy-analyze.sh" "$FIXTURE" --no-graph --if-changed
if [[ -f "$FIXTURE/.plaesy/analysis/.analysis-fingerprint" ]]; then
    echo "  PASS: fingerprint file created"
else
    echo "  FAIL: fingerprint file not created"
    FAILURES=$((FAILURES + 1))
fi

# Save modification times before second run
MTIME_JSON_BEFORE=$(stat -c %Y "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null || stat -f %m "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null)
MTIME_STRUCT_BEFORE=$(stat -c %Y "$FIXTURE/.plaesy/analysis/project.structure.json" 2>/dev/null || stat -f %m "$FIXTURE/.plaesy/analysis/project.structure.json" 2>/dev/null)

# Second run with --if-changed should skip (fingerprint matches)
sleep 1
bash "$ROOT/scripts/bash/plaesy-analyze.sh" "$FIXTURE" --no-graph --if-changed

MTIME_JSON_AFTER=$(stat -c %Y "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null || stat -f %m "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null)
MTIME_STRUCT_AFTER=$(stat -c %Y "$FIXTURE/.plaesy/analysis/project.structure.json" 2>/dev/null || stat -f %m "$FIXTURE/.plaesy/analysis/project.structure.json" 2>/dev/null)

if [[ "$MTIME_JSON_BEFORE" == "$MTIME_JSON_AFTER" ]]; then
    echo "  PASS: project.json not regenerated (skipped correctly)"
else
    echo "  FAIL: project.json was regenerated despite --if-changed"
    FAILURES=$((FAILURES + 1))
fi

if [[ "$MTIME_STRUCT_BEFORE" == "$MTIME_STRUCT_AFTER" ]]; then
    echo "  PASS: project.structure.json not regenerated (skipped correctly)"
else
    echo "  FAIL: project.structure.json was regenerated despite --if-changed"
    FAILURES=$((FAILURES + 1))
fi

echo ""
echo "=== Test 3: --force overrides --if-changed ==="
bash "$ROOT/scripts/bash/plaesy-analyze.sh" "$FIXTURE" --no-graph --if-changed --force
MTIME_JSON_FORCE=$(stat -c %Y "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null || stat -f %m "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null)
if [[ "$MTIME_JSON_AFTER" != "$MTIME_JSON_FORCE" ]]; then
    echo "  PASS: --force regenerated despite --if-changed"
else
    echo "  FAIL: --force did not override --if-changed"
    FAILURES=$((FAILURES + 1))
fi

echo ""
echo "=== Test 4: Modifying a source file invalidates fingerprint ==="
sleep 1
echo "// modified" >> "$FIXTURE/src/lib.js"
bash "$ROOT/scripts/bash/plaesy-analyze.sh" "$FIXTURE" --no-graph --if-changed
MTIME_JSON_MOD=$(stat -c %Y "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null || stat -f %m "$FIXTURE/.plaesy/analysis/project.json" 2>/dev/null)
if [[ "$MTIME_JSON_FORCE" != "$MTIME_JSON_MOD" ]]; then
    echo "  PASS: source modification triggered regeneration"
else
    echo "  FAIL: source modification did not trigger regeneration"
    FAILURES=$((FAILURES + 1))
fi

echo ""
if [[ $FAILURES -eq 0 ]]; then
    echo "ALL ANALYZER TESTS PASSED"
    exit 0
else
    echo "$FAILURES TEST(S) FAILED"
    exit 1
fi