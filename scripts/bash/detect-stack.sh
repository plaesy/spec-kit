#!/bin/bash

# Detects which instructions/*.instructions.md files are relevant to a target
# project, using instructions/mapping.json as the keyword registry.
#
# Output: one instruction filename per line (includes always_load files and matched category files)
# Usage: detect-stack.sh [TARGET_DIR]

set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PLAESY_ROOT="${PLAESY_ROOT:-$SCRIPT_DIR/../..}"
readonly MAPPING_FILE="$PLAESY_ROOT/instructions/mapping.json"
readonly TARGET_DIR="${1:-.}"

# Output all always_load files from mapping.json (dynamic, no hardcoding)
if [[ -f "$MAPPING_FILE" ]]; then
    # Extract always_load section: from "always_load": [ to closing ]
    sed -n '/"always_load":[[:space:]]*\[/,/\]/p' "$MAPPING_FILE" | \
        grep -o '"[^"]*\.instructions\.md"' | tr -d '"'
fi

# Gather scan text from common manifest files and spec/context files.
# Root-level manifests are read first (fast path), then manifests in
# subdirectories are found recursively so monorepos (e.g. client/pubspec.yaml)
# are detected too. .git and node_modules are pruned for speed.
scan_text=""
for f in package.json go.mod pom.xml build.gradle Cargo.toml pubspec.yaml Gemfile requirements.txt; do
    [[ -f "$TARGET_DIR/$f" ]] && scan_text+=" $(cat "$TARGET_DIR/$f" 2>/dev/null)"
done
if command -v find >/dev/null 2>&1; then
    while IFS= read -r manifest; do
        [[ -n "$manifest" ]] && scan_text+=" $(cat "$manifest" 2>/dev/null)"
    done < <(find "$TARGET_DIR" -maxdepth 3 \
        \( -name .git -o -name node_modules \) -prune -o -type f \
        \( -name package.json -o -name go.mod -o -name pom.xml -o -name build.gradle -o \
           -name Cargo.toml -o -name pubspec.yaml -o -name Gemfile -o -name requirements.txt \) \
        -print 2>/dev/null)
fi
for f in "$TARGET_DIR"/specs/*/context.md "$TARGET_DIR"/specs/*/requirements.md \
         "$TARGET_DIR"/docs/specs/*/context.md "$TARGET_DIR"/docs/specs/*/requirements.md \
         "$TARGET_DIR"/docs/project.json; do
    [[ -f "$f" ]] && scan_text+=" $(cat "$f" 2>/dev/null)"
done
if find "$TARGET_DIR" -maxdepth 3 \( -name '*.tsx' -o -name '*.jsx' \) -print -quit 2>/dev/null | grep -q .; then
    scan_text+=" tsx jsx react"
fi
scan_text_lower=$(echo "$scan_text" | tr '[:upper:]' '[:lower:]')

# mapping.json blocks are always "file" followed by "keywords" in that order,
# so index-aligned extraction is safe for this file's structure.
mapfile -t files < <(grep -oP '"file":\s*"\K[^"]+' "$MAPPING_FILE")
mapfile -t keyword_lines < <(grep -oP '"keywords":\s*\[\K[^\]]+' "$MAPPING_FILE")

for i in "${!files[@]}"; do
    file="${files[$i]}"
    kwline="${keyword_lines[$i]:-}"
    [[ -z "$kwline" ]] && continue

    IFS=',' read -ra kws <<< "$kwline"
    for kw in "${kws[@]}"; do
        # Pure-bash cleanup (no echo/sed/tr forks): strip quotes, trim
        # surrounding whitespace, lowercase. Trimming matters: every keyword
        # after the first carries a leading space from the "kw1", "kw2" split.
        kw="${kw//\"/}"
        kw="${kw#"${kw%%[![:space:]]*}"}"
        kw="${kw%"${kw##*[![:space:]]}"}"
        kw="${kw,,}"
        [[ -z "$kw" ]] && continue
        if [[ "$scan_text_lower" == *"$kw"* ]]; then
            echo "$file"
            break
        fi
    done
done | sort -u
