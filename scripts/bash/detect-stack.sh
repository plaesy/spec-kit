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
    # mindepth 1 (not 2) so a root-level *.csproj (single-project .NET repos) is
    # caught by this same pass — wildcard join means no separate find is needed.
    while IFS= read -r manifest; do
        [[ -n "$manifest" ]] && scan_text+=" $(cat "$manifest" 2>/dev/null)"
    done < <(find "$TARGET_DIR" -mindepth 1 -maxdepth 3 \
        \( -name .git -o -name node_modules \) -prune -o -type f \
        \( -name package.json -o -name go.mod -o -name pom.xml -o -name build.gradle -o \
           -name Cargo.toml -o -name pubspec.yaml -o -name Gemfile -o -name requirements.txt -o \
           -name '*.csproj' \) \
        -print 2>/dev/null)
fi
# Presence of a matching document (e.g. a .docx file, not just code importing
# python-docx) is itself a signal that an instructions file should load. Driven
# generically by each mapping.json entry's "extensions" array — no hardcoded
# per-framework extension list here.
direct_files=""
if command -v find >/dev/null 2>&1; then
    # Each entry with an "extensions" array looks like:
    #   "file": "word.instructions.md", ... "extensions": [".doc", ".docx"]
    # -Pzo (null-data) lets the pattern span the keywords/patterns lines in
    # between; [^}]*? never crosses an object boundary since none of those
    # in-between arrays contain a literal "}".
    declare -A ext_to_files=()
    # .tsx/.jsx ride the same single find pass as the mapping.json extensions
    # below (existence-only check either way); ext_to_files has no entry for
    # them, so they only ever flip react_detected, never add to direct_files.
    all_ext_args=(-o -iname '*.tsx' -o -iname '*.jsx')
    while IFS= read -r -d '' block; do
        file=$(grep -oP '"file":\s*"\K[^"]+' <<<"$block" | head -1)
        exts=$(grep -oP '"extensions":\s*\[\K[^\]]+' <<<"$block" | grep -oP '"\K\.[^"]+')
        [[ -z "$file" || -z "$exts" ]] && continue
        for ext in $exts; do
            [[ -z "${ext_to_files[$ext]:-}" ]] && all_ext_args+=(-o -iname "*$ext")
            ext_to_files[$ext]+="$file"$'\n'
        done
    done < <(grep -Pzo '"file":\s*"[^"]+"[^}]*?"extensions":\s*\[[^\]]*\]' "$MAPPING_FILE")
    # Same idea as extensions above, but matching an exact filename (e.g.
    # "next.config.js") rather than a suffix — for frameworks like Next.js
    # whose manifest content carries no distinctive, prose-safe keyword (a
    # bare "next" dependency key is too generic: it would match ordinary
    # English "next" in spec/context prose) but whose config filename is a
    # reliable, unambiguous signal on its own.
    declare -A name_to_files=()
    while IFS= read -r -d '' block; do
        file=$(grep -oP '"file":\s*"\K[^"]+' <<<"$block" | head -1)
        names=$(grep -oP '"filenames":\s*\[\K[^\]]+' <<<"$block" | grep -oP '"\K[^"]+')
        [[ -z "$file" || -z "$names" ]] && continue
        while IFS= read -r nm; do
            [[ -z "$nm" ]] && continue
            [[ -z "${name_to_files[$nm]:-}" ]] && all_ext_args+=(-o -iname "$nm")
            name_to_files[$nm]+="$file"$'\n'
        done <<< "$names"
    done < <(grep -Pzo '"file":\s*"[^"]+"[^}]*?"filenames":\s*\[[^\]]*\]' "$MAPPING_FILE")
    # One find call for every extension/filename across every entry (plus
    # tsx/jsx), instead of one find per extension: the tree walk (the
    # expensive part) happens exactly once.
    react_detected=0
    mapfile -t found_names < <(find "$TARGET_DIR" -maxdepth 3 \( -name .git -o -name node_modules \) -prune -o -type f \( "${all_ext_args[@]:1}" \) -print 2>/dev/null)
    for name in "${found_names[@]}"; do
        case "${name,,}" in
            *.tsx|*.jsx) react_detected=1 ;;
        esac
        for ext in "${!ext_to_files[@]}"; do
            if [[ "${name,,}" == *"${ext,,}" ]]; then
                direct_files+="${ext_to_files[$ext]}"
            fi
        done
        base="${name##*/}"
        for nm in "${!name_to_files[@]}"; do
            if [[ "${base,,}" == "${nm,,}" ]]; then
                direct_files+="${name_to_files[$nm]}"
            fi
        done
    done
fi
for f in "$TARGET_DIR"/specs/*/context.md "$TARGET_DIR"/specs/*/requirements.md \
         "$TARGET_DIR"/docs/specs/*/context.md "$TARGET_DIR"/docs/specs/*/requirements.md \
         "$TARGET_DIR"/docs/project.json \
         "$TARGET_DIR"/.plaesy/analysis/project.json; do
    [[ -f "$f" ]] && scan_text+=" $(cat "$f" 2>/dev/null)"
done
[[ "${react_detected:-0}" -eq 1 ]] && scan_text+=" tsx jsx react"
scan_text_lower=$(echo "$scan_text" | tr '[:upper:]' '[:lower:]')

# mapping.json blocks are always "file" followed by "keywords" in that order,
# so index-aligned extraction is safe for this file's structure.
mapfile -t files < <(grep -oP '"file":\s*"\K[^"]+' "$MAPPING_FILE")
mapfile -t keyword_lines < <(grep -oP '"keywords":\s*\[\K[^\]]+' "$MAPPING_FILE")

{
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
        # Word-boundary match, not plain substring: a bare substring test would
        # match short/common keywords ("java" inside "javascript", "pr" inside
        # "prepare"/"private") against any manifest, defeating selective
        # install. Non-alnum/underscore on both sides (or string start/end)
        # counts as a boundary; multi-word keywords ("pull request") still
        # match as one literal unit since only their outer edges are checked.
        if [[ "$scan_text_lower" == *[![:alnum:]_]"$kw"[![:alnum:]_]* ]] || \
           [[ "$scan_text_lower" == "$kw"[![:alnum:]_]* ]] || \
           [[ "$scan_text_lower" == *[![:alnum:]_]"$kw" ]] || \
           [[ "$scan_text_lower" == "$kw" ]]; then
            echo "$file"
            break
        fi
    done
done
printf '%s' "$direct_files"
} | sort -u
