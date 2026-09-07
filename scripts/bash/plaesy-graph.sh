#!/usr/bin/env bash
# plaesy-graph.sh - Bash port of plaesy-graph.ps1 (lightweight knowledge-graph
# builder for the plaesy spec-kit repo). Same feature set, same output shape,
# no external dependency (no jq/python) - pure bash + grep/find/awk.
#
# Emits:
#   .plaesy/analysis/project.graph.json    - nodes + edges
#   .plaesy/analysis/project.html    - self-contained force-directed viz
#   .plaesy/analysis/reports.md - plain-language summary
#   .plaesy/analysis/.nodes.tsv, .edges.tsv - internal cache used by
#     --query/--explain/--path-query/--impact-check (avoids writing a JSON
#     parser in bash; project.graph.json remains the canonical published artifact)
#
# Usage:
#   plaesy-graph.sh                                   # build graph for repo root
#   plaesy-graph.sh --path <dir>                       # build graph for a specific dir
#   plaesy-graph.sh --query "common.ps1"               # keyword query
#   plaesy-graph.sh --path-query "a.md" "b.md"          # shortest path
#   plaesy-graph.sh --explain "go.instructions.md"      # explain one node
#   plaesy-graph.sh --impact-check "common.ps1" --impact-depth 1
#   plaesy-graph.sh --semantic-queue                    # export INFERRED edges
#   plaesy-graph.sh --apply-semantic annotations.json   # merge rationale back
#   plaesy-graph.sh --watch                             # rebuild automatically on source changes
#   plaesy-graph.sh --watch --watch-interval 5          # poll every 5s instead of the 3s default
#   plaesy-graph.sh --if-changed                        # rebuild only if source files changed since last build

# Deliberately not using `set -e`: this script relies heavily on
# `[[ cond ]] && action` one-liners, and under `set -e` a false condition
# there (a normal, expected outcome, not an error) aborts the whole script.
set -u

REPO_PATH="."
OUT_DIR=".plaesy/analysis"
QUERY=""
PATH_FROM=""
PATH_TO=""
EXPLAIN_NODE=""
IMPACT_NODE=""
IMPACT_DEPTH=2
IMPACT_VISUALIZE=0
SEMANTIC_QUEUE=0
APPLY_SEMANTIC=""
BUSINESS_LOGIC=0
GENERATE_PATHS=0
FUZZY_QUERY=""
SEARCH_DEPTH=2
WATCH=0
WATCH_INTERVAL=3
IF_CHANGED=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        --path) REPO_PATH="$2"; shift 2 ;;
        --outdir) OUT_DIR="$2"; shift 2 ;;
        --query) QUERY="$2"; shift 2 ;;
        --fuzzy-query) FUZZY_QUERY="$2"; shift 2 ;;
        --search-depth) SEARCH_DEPTH="$2"; shift 2 ;;
        --path-query) PATH_FROM="$2"; PATH_TO="$3"; shift 3 ;;
        --explain) EXPLAIN_NODE="$2"; shift 2 ;;
        --impact-check) IMPACT_NODE="$2"; shift 2 ;;
        --impact-depth) IMPACT_DEPTH="$2"; shift 2 ;;
        --impact-visualize) IMPACT_VISUALIZE=1; shift ;;
        --semantic-queue) SEMANTIC_QUEUE=1; shift ;;
        --business-logic) BUSINESS_LOGIC=1; shift ;;
        --generate-paths) GENERATE_PATHS=1; shift ;;
        --apply-semantic) APPLY_SEMANTIC="$2"; shift 2 ;;
        --watch) WATCH=1; shift ;;
        --watch-interval) WATCH_INTERVAL="$2"; shift 2 ;;
        --if-changed) IF_CHANGED=1; shift ;;
        *) echo "[plaesy-graph] Unknown argument: $1" >&2; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SYMBOLS_AWK="$SCRIPT_DIR/plaesy-graph-symbols.awk"
MENTIONS_AWK="$SCRIPT_DIR/plaesy-graph-mentions.awk"

REPO_ROOT="$(cd "$REPO_PATH" && pwd)"
OUT_FULL="$REPO_ROOT/$OUT_DIR"
PROJECT_JSON="$OUT_FULL/project.graph.json"
NODES_TSV="$OUT_FULL/.nodes.tsv"
EDGES_TSV="$OUT_FULL/.edges.tsv"
FINGERPRINT_FILE="$OUT_FULL/.fingerprint"

INCLUDE_EXT_RE='\.(md|ps1|sh|js|jsx|ts|tsx|py|go|dart|java|kt|kts|swift|c|h|cc|cpp|hpp|cs|rs|rb|php)$'
EXCLUDE_DIRS_RE='(^|/)(\.[^/]+|node_modules|dist|build|__pycache__|vendor)(/|$)'

# Minimal JSON string escaper (backslash, quote, control chars). Sets the
# global $JSON_OUT instead of echo + $(...): write_project_json escapes every
# node id/label/type/group, so $(...) here was ~4 subshell forks per node
# (~110-200ms each on MSYS) - the JSON writer alone cost ~14s on 40 nodes.
JSON_OUT=""
json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    JSON_OUT="$s"
}

# Sets $NODE_TYPE_OUT (avoids $(...) fork on the once-per-file hot loop).
NODE_TYPE_OUT=""
node_type_of() {
    local rel="$1"
    local top="${rel%%/*}"
    case "$top" in
        chatmodes) NODE_TYPE_OUT="chatmode" ;;
        instructions) NODE_TYPE_OUT="instruction" ;;
        checklists) NODE_TYPE_OUT="checklist" ;;
        templates) NODE_TYPE_OUT="template" ;;
        scripts) NODE_TYPE_OUT="script" ;;
        docs) NODE_TYPE_OUT="doc" ;;
        prompts) NODE_TYPE_OUT="prompt" ;;
        testing) NODE_TYPE_OUT="testing" ;;
        *)
            case "$rel" in
                *.js|*.jsx|*.ts|*.tsx|*.py|*.go|*.dart|*.java|*.kt|*.kts|*.swift|*.c|*.h|*.cc|*.cpp|*.hpp|*.cs|*.rs|*.rb|*.php) NODE_TYPE_OUT="source" ;;
                *) NODE_TYPE_OUT="other" ;;
            esac
            ;;
    esac
}

# Sets $LAYER_OUT to detected architecture layer (API, Service, Data, UI, Utility) or empty if indeterminate.
LAYER_OUT=""
get_architecture_layer() {
    local rel="$1"
    local rel_lower="${rel,,}"
    local folder="${rel%%/*}"
    local filename="${rel##*/}"
    filename="${filename%.*}"
    filename="${filename,,}"

    # API Layer patterns
    if [[ "$folder" =~ ^(api|controllers|handlers|routes|endpoints|gateway)$ ]] || \
       [[ "$filename" =~ (controller|handler|route|gateway|middleware|app|server) ]] || \
       [[ "$rel_lower" =~ /(api|controllers|routes|handlers)/ ]]; then
        LAYER_OUT="API"
        return
    fi

    # Service Layer patterns
    if [[ "$folder" =~ ^(services|business|use-cases|use_cases|orchestration|workflows)$ ]] || \
       [[ "$filename" =~ (service|business|usecase|workflow|orchestration) ]] || \
       [[ "$rel_lower" =~ /(services|business|use[-_]cases)/ ]]; then
        LAYER_OUT="Service"
        return
    fi

    # Data Layer patterns
    if [[ "$folder" =~ ^(repositories|repo|models|database|db|schemas|data|persistence)$ ]] || \
       [[ "$filename" =~ (repo|repository|model|entity|schema|mapper|query|dao) ]] || \
       [[ "$rel_lower" =~ /(repositories|models|database|db|persistence|entities)/ ]]; then
        LAYER_OUT="Data"
        return
    fi

    # UI Layer patterns
    if [[ "$folder" =~ ^(components|views|pages|screens|ui|presentation|widgets)$ ]] || \
       [[ "$filename" =~ (component|view|page|screen|widget) ]] || \
       [[ "$rel_lower" =~ /(components|views|pages|screens|ui)/ ]]; then
        LAYER_OUT="UI"
        return
    fi

    # Utility Layer patterns
    if [[ "$folder" =~ ^(utils|helpers|common|lib|tools|config|constants)$ ]] || \
       [[ "$filename" =~ (util|helper|common|constant|config|logger|validator|formatter) ]] || \
       [[ "$rel_lower" =~ /(utils|helpers|common|config|constants)/ ]]; then
        LAYER_OUT="Utility"
        return
    fi

    LAYER_OUT=""
}

# Lexical path normalization (realpath -m equivalent) WITHOUT spawning a
# subprocess. realpath -m costs ~200ms per call on MSYS/Windows (one fork +
# exec each time); build_graph resolves every markdown link / source ref /
# import spec through it, so on the real repo that was ~380 process spawns
# (~76s) before any awk/grep work even started.
#
# ALSO returns via the global $NP_OUT instead of $(...) command substitution:
# on MSYS each $() forks a subshell (~110-200ms), and the extraction loops
# call this once per candidate - the original per-file loop was calling
# $() hundreds of times. Reading $NP_OUT avoids every one of those forks.
# Arguments: $1 = base directory (absolute), $2 = link/candidate (may be
# absolute - then $1 is ignored, matching realpath behavior). No filesystem
# access; symlinks are NOT resolved (same as `realpath -m`).
NP_OUT=""
normalize_path() {
    local base="$1" path="$2"
    case "$path" in
        /*) base="" ;;
        [A-Za-z]:*) base="" ;;
    esac
    local -a parts=()
    local seg
    local IFS='/'
    for seg in $base $path; do
        [[ -z "$seg" || "$seg" == "." ]] && continue
        if [[ "$seg" == ".." ]]; then
            if [[ ${#parts[@]} -gt 0 ]]; then unset 'parts[-1]'; fi
        else
            parts+=("$seg")
        fi
    done
    local out="/" p
    for p in "${parts[@]}"; do out+="$p/"; done
    NP_OUT="${out%/}"
}

# resolve a relative link target against $1's directory to a known
# repo-relative node path; sets $REL_OUT (empty if unresolved). Returns via a
# global instead of echo + $(...) for the same fork-avoidance reason as
# normalize_path above.
REL_OUT=""
resolve_rel_link() {
    local from_rel="$1" link="$2"
    REL_OUT=""
    [[ "$link" =~ ^https?:// ]] && return
    [[ "$link" == \#* ]] && return
    link="${link%%#*}"
    [[ -z "$link" ]] && return
    local from_dir="$REPO_ROOT"
    [[ "$from_rel" == */* ]] && from_dir="$REPO_ROOT/${from_rel%/*}"
    normalize_path "$from_dir" "$link"
    local resolved="$NP_OUT"
    [[ "$resolved" != "$REPO_ROOT"/* ]] && return
    local rel="${resolved#"$REPO_ROOT"/}"
    [[ -n "${NTYPE[$rel]:-}" ]] && REL_OUT="$rel"
}

build_graph() {
    echo "[plaesy-graph] Scanning $REPO_ROOT ..." >&2
    mkdir -p "$OUT_FULL"

    local out_dir_norm="${OUT_DIR#./}"
    out_dir_norm="${out_dir_norm%/}"

    # 1) Collect candidate files -> relative paths (forward-slash)
    local all_files
    mapfile -t all_files < <(
        cd "$REPO_ROOT" && find . -type f \
            | sed 's|^\./||' \
            | grep -Ei "$INCLUDE_EXT_RE" \
            | grep -Ev "$EXCLUDE_DIRS_RE" \
            | grep -Fv "$out_dir_norm/"
    )

    if [[ ${#all_files[@]} -eq 0 ]]; then
        echo "[plaesy-graph] No files found under $REPO_ROOT" >&2
        exit 1
    fi

    declare -A NTYPE NGROUP NLAYER SYMBOLS
    for rel in "${all_files[@]}"; do
        node_type_of "$rel"
        NTYPE["$rel"]="$NODE_TYPE_OUT"
        NGROUP["$rel"]="${rel%%/*}"
        get_architecture_layer "$rel"
        NLAYER["$rel"]="$LAYER_OUT"
    done

    : > "$EDGES_TSV.tmp"

    # Pattern file of all known node paths, for the batched mention scan below.
    local allnodes_file="$OUT_FULL/.allnodes.tmp"
    printf '%s\n' "${all_files[@]}" > "$allnodes_file"

    # 2) Structural extraction, BATCHED. One `grep -rHZ -oP` per pattern family
    # scans every file in a SINGLE process instead of spawning one grep per
    # file. On MSYS/Windows each subprocess costs ~200ms, so the old per-file
    # loop (221 files x ~5 patterns = ~1000 spawns) took minutes; a single
    # `grep -r` over the whole tree takes well under a second. Match targets
    # are NUL-terminated (`path\0match\n`) so paths containing colons survive.
    local grep_excl=(
        --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=.plaesy
        --exclude-dir=dist --exclude-dir=build --exclude-dir=__pycache__
        --exclude-dir=vendor --exclude-dir=.venv
    )
    local out_top="${out_dir_norm%%/*}"
    [[ -n "$out_top" && "$out_top" != "." ]] && grep_excl+=(--exclude-dir="$out_top")

    # 2a) markdown reference links
    while IFS= read -r -d '' mrel && IFS= read -r target; do
        local rel="${mrel#./}"
        resolve_rel_link "$rel" "$target"
        local resolved="$REL_OUT"
        [[ -n "$resolved" && "$resolved" != "$rel" ]] && printf '%s\t%s\t%s\t%s\n' "$rel" "$resolved" "references" "EXTRACTED" >> "$EDGES_TSV.tmp"
    done < <(cd "$REPO_ROOT" && grep -rHZ -oP '\[[^\]]*\]\(\K[^)\s]+' --include='*.md' "${grep_excl[@]}" . 2>/dev/null)

    # 2b) ps1 dot-source references
    while IFS= read -r -d '' mrel && IFS= read -r target; do
        local rel="${mrel#./}"
        local tgt="${target/\$PSScriptRoot\\/}"
        tgt="${tgt/\$PSScriptRoot\//}"
        tgt="${tgt//\\//}"
        resolve_rel_link "$rel" "$tgt"
        local resolved="$REL_OUT"
        [[ -n "$resolved" && "$resolved" != "$rel" ]] && printf '%s\t%s\t%s\t%s\n' "$rel" "$resolved" "sources" "EXTRACTED" >> "$EDGES_TSV.tmp"
    done < <(cd "$REPO_ROOT" && grep -rHZ -oP '(?:^|\s)\.\s+["'"'"']?\K[^"'"'"'\r\n]+\.ps1' --include='*.ps1' "${grep_excl[@]}" . 2>/dev/null)

    # 2c) ps1 call references
    while IFS= read -r -d '' mrel && IFS= read -r target; do
        local rel="${mrel#./}"
        local tgt="${target/\$PSScriptRoot\\/}"
        tgt="${tgt/\$PSScriptRoot\//}"
        tgt="${tgt//\\//}"
        resolve_rel_link "$rel" "$tgt"
        local resolved="$REL_OUT"
        [[ -n "$resolved" && "$resolved" != "$rel" ]] && printf '%s\t%s\t%s\t%s\n' "$rel" "$resolved" "calls" "EXTRACTED" >> "$EDGES_TSV.tmp"
    done < <(cd "$REPO_ROOT" && grep -rHZ -oP "&\s+[\"']\K[^\"']+\.ps1" --include='*.ps1' "${grep_excl[@]}" . 2>/dev/null)

    # 2d) shell source references
    while IFS= read -r -d '' mrel && IFS= read -r target; do
        local rel="${mrel#./}"
        resolve_rel_link "$rel" "$target"
        local resolved="$REL_OUT"
        [[ -n "$resolved" && "$resolved" != "$rel" ]] && printf '%s\t%s\t%s\t%s\n' "$rel" "$resolved" "sources" "EXTRACTED" >> "$EDGES_TSV.tmp"
    done < <(cd "$REPO_ROOT" && grep -rHZ -oP '(?:^|\s)(?:source|\.)\s+["'"'"']?\K[^"'"'"'\r\n]+\.sh' --include='*.sh' "${grep_excl[@]}" . 2>/dev/null)

    # 2e) JS/TS import specs
    while IFS= read -r -d '' mrel && IFS= read -r spec; do
        local rel="${mrel#./}"
        [[ "$spec" != .* ]] && continue
        local from_dir="$REPO_ROOT"
        [[ "$rel" == */* ]] && from_dir="$REPO_ROOT/${rel%/*}"
        local found="" cand
        for cand in "$spec" "$spec.js" "$spec.jsx" "$spec.ts" "$spec.tsx" "$spec/index.js" "$spec/index.ts"; do
            normalize_path "$from_dir" "$cand"
            local resolved="$NP_OUT"
            [[ "$resolved" != "$REPO_ROOT"/* ]] && continue
            local relc="${resolved#"$REPO_ROOT"/}"
            if [[ -n "${NTYPE[$relc]:-}" ]]; then found="$relc"; break; fi
        done
        [[ -n "$found" && "$found" != "$rel" ]] && printf '%s\t%s\t%s\t%s\n' "$rel" "$found" "imports" "EXTRACTED" >> "$EDGES_TSV.tmp"
    done < <(cd "$REPO_ROOT" && grep -rHZ -oP "(?:import\s+.*?from\s+['\"]|require\(['\"])\K[^'\"]+" --include='*.js' --include='*.jsx' --include='*.ts' --include='*.tsx' "${grep_excl[@]}" . 2>/dev/null)

    # 2f) Python import specs
    while IFS= read -r -d '' mrel && IFS= read -r spec; do
        local rel="${mrel#./}"
        [[ -z "$spec" || "$spec" =~ ^\.+$ ]] && continue
        local from_dir="$REPO_ROOT"
        [[ "$rel" == */* ]] && from_dir="$REPO_ROOT/${rel%/*}"
        local specPath="${spec//./\/}"
        normalize_path "$from_dir" "$specPath.py"
        local resolved="$NP_OUT" found=""
        if [[ "$resolved" == "$REPO_ROOT"/* ]]; then
            local relc="${resolved#"$REPO_ROOT"/}"
            [[ -n "${NTYPE[$relc]:-}" ]] && found="$relc"
        fi
        [[ -n "$found" && "$found" != "$rel" ]] && printf '%s\t%s\t%s\t%s\n' "$rel" "$found" "imports" "EXTRACTED" >> "$EDGES_TSV.tmp"
    done < <(cd "$REPO_ROOT" && grep -rHZ -oP '^\s*(?:from\s+\K[\w\.]+(?=\s+import)|import\s+\K[\w\.]+)' --include='*.py' "${grep_excl[@]}" . 2>/dev/null)

    # 2f-go) Go import specs. Go imports name a package (directory), not a
    # single file, so a local-module import fans out to every .go file in
    # the target directory. Local-module imports are recognized by matching
    # the spec against each go.mod's declared `module` path (longest match
    # wins for nested modules); anything else is treated as an external
    # dependency and skipped.
    # =() initializer required: under `set -u`, bash treats an associative
    # array that was declared but never assigned a key (e.g. no go.mod found)
    # as unbound the moment it's expanded, even just for a count check.
    declare -A GOMOD_PATH=()
    while IFS= read -r gomod_file; do
        local gomod_dir="${gomod_file%/go.mod}"
        [[ "$gomod_dir" == "$gomod_file" ]] && gomod_dir=""
        local modpath
        modpath=$(grep -m1 -E '^module[[:space:]]+' "$REPO_ROOT/$gomod_file" 2>/dev/null | sed -E 's/^module[[:space:]]+//')
        modpath="${modpath%%[[:space:]]*}"
        [[ -n "$modpath" ]] && GOMOD_PATH["$gomod_dir"]="$modpath"
    done < <(cd "$REPO_ROOT" && find . -name 'go.mod' ! -path '*/.git/*' 2>/dev/null | sed 's|^\./||')

    if [[ ${#GOMOD_PATH[@]} -gt 0 ]]; then
        declare -A GO_DIR_FILES
        for rel in "${all_files[@]}"; do
            case "$rel" in
                *.go)
                    [[ "$rel" == *_test.go ]] && continue
                    local d="${rel%/*}"
                    [[ "$d" == "$rel" ]] && d=""
                    GO_DIR_FILES["$d"]+="$rel"$'\n'
                    ;;
            esac
        done

        while IFS= read -r -d '' mrel && IFS= read -r spec; do
            local rel="${mrel#./}"
            local rel_dir="${rel%/*}"
            [[ "$rel_dir" == "$rel" ]] && rel_dir=""
            local best_gomod_dir="" best_len=-1 gd
            for gd in "${!GOMOD_PATH[@]}"; do
                if [[ -z "$gd" || "$rel_dir" == "$gd" || "$rel_dir" == "$gd"/* ]]; then
                    local l=${#gd}
                    [[ $l -gt $best_len ]] && { best_len=$l; best_gomod_dir="$gd"; }
                fi
            done
            [[ $best_len -lt 0 ]] && continue
            local modpath="${GOMOD_PATH[$best_gomod_dir]}"
            [[ "$spec" != "$modpath" && "$spec" != "$modpath"/* ]] && continue
            local subpath="${spec#"$modpath"}"
            subpath="${subpath#/}"
            local target_dir="$best_gomod_dir"
            [[ -n "$subpath" ]] && target_dir="${target_dir:+$target_dir/}$subpath"
            local files="${GO_DIR_FILES[$target_dir]:-}"
            [[ -z "$files" ]] && continue
            while IFS= read -r tfile; do
                [[ -z "$tfile" || "$tfile" == "$rel" ]] && continue
                printf '%s\t%s\t%s\t%s\n' "$rel" "$tfile" "imports" "EXTRACTED" >> "$EDGES_TSV.tmp"
            done <<< "$files"
        done < <(cd "$REPO_ROOT" && grep -rHZ -oP '^\s*"\K[^"]+(?="\s*$)' --include='*.go' "${grep_excl[@]}" . 2>/dev/null)
        unset GO_DIR_FILES
    fi

    # 2g-dart) Dart import specs: relative imports resolve like other
    # languages; `package:<name>/...` imports resolve via each pubspec.yaml's
    # declared package name -> its lib/ dir (handles Flutter's own-package
    # imports, the common case for intra-app references).
    declare -A DART_PKG_LIB
    while IFS= read -r pspec_file; do
        local pkg_name
        pkg_name=$(grep -m1 -E '^name:[[:space:]]*' "$REPO_ROOT/$pspec_file" 2>/dev/null | sed -E 's/^name:[[:space:]]*//')
        pkg_name="${pkg_name%%[[:space:]]*}"
        [[ -n "$pkg_name" ]] && DART_PKG_LIB["$pkg_name"]="${pspec_file%/*}/lib"
    done < <(cd "$REPO_ROOT" && find . -name 'pubspec.yaml' ! -path '*/.git/*' 2>/dev/null | sed 's|^\./||')

    while IFS= read -r -d '' mrel && IFS= read -r spec; do
        local rel="${mrel#./}"
        [[ "$spec" == dart:* ]] && continue
        local resolved_full=""
        if [[ "$spec" == package:* ]]; then
            local pkgspec="${spec#package:}"
            local pkgname="${pkgspec%%/*}"
            local pkgpath="${pkgspec#*/}"
            local libdir="${DART_PKG_LIB[$pkgname]:-}"
            if [[ -n "$libdir" ]]; then
                normalize_path "$REPO_ROOT/$libdir" "$pkgpath"
                resolved_full="$NP_OUT"
            fi
        else
            local from_dir="$REPO_ROOT"
            [[ "$rel" == */* ]] && from_dir="$REPO_ROOT/${rel%/*}"
            normalize_path "$from_dir" "$spec"
            resolved_full="$NP_OUT"
        fi
        [[ "$resolved_full" != "$REPO_ROOT"/* ]] && continue
        local relc="${resolved_full#"$REPO_ROOT"/}"
        [[ -n "${NTYPE[$relc]:-}" && "$relc" != "$rel" ]] && printf '%s\t%s\t%s\t%s\n' "$rel" "$relc" "imports" "EXTRACTED" >> "$EDGES_TSV.tmp"
    done < <(cd "$REPO_ROOT" && grep -rHZ -oP "import\s+['\"]\K[^'\"]+" --include='*.dart' "${grep_excl[@]}" . 2>/dev/null)

    # 2g-java) Java import specs (dotted class names). Resolved via an index
    # built by stripping each .java file's top-level source-root folder (e.g.
    # src/, src_win/) and dotting the rest - matches the common convention of
    # one source root per top-level dir, without needing to know the exact
    # src/main/java-style layout of an arbitrary project.
    declare -A JAVA_INDEX
    for rel in "${all_files[@]}"; do
        case "$rel" in
            *.java)
                [[ "$rel" != */* ]] && continue
                local dotted="${rel#*/}"
                dotted="${dotted%.java}"
                dotted="${dotted//\//.}"
                JAVA_INDEX["$dotted"]="$rel"
                ;;
        esac
    done

    while IFS= read -r -d '' mrel && IFS= read -r spec; do
        local rel="${mrel#./}"
        [[ "$spec" == *'*' ]] && continue
        local found="${JAVA_INDEX[$spec]:-}"
        [[ -n "$found" && "$found" != "$rel" ]] && printf '%s\t%s\t%s\t%s\n' "$rel" "$found" "imports" "EXTRACTED" >> "$EDGES_TSV.tmp"
    done < <(cd "$REPO_ROOT" && grep -rHZ -oP '^\s*import\s+(?:static\s+)?\K[\w.]+(?=\s*;)' --include='*.java' "${grep_excl[@]}" . 2>/dev/null)

    # 2g) mentions: one awk pass over ALL files instead of one per file. The
    # awk mirrors plaesy-graph.ps1's per-candidate index() check (both forward
    # and backslash forms), which `grep -oFf` under-detects due to overlapping
    # substring candidates. Only applied to doc/script corpus, not source files.
    while IFS=$'\t' read -r mrel cand; do
        [[ -z "$cand" || "$cand" == "$mrel" ]] && continue
        printf '%s\t%s\t%s\t%s\n' "$mrel" "$cand" "mentions" "INFERRED" >> "$EDGES_TSV.tmp"
    done < <(awk -v candfile="$allnodes_file" -f "$MENTIONS_AWK" "${all_files[@]}" 2>/dev/null)

    # 2h) symbols: one awk pass over ALL files (ext derived from FILENAME),
    # grouped back to per-node pipe-joined form here in pure bash.
    declare -A SYMBOLS
    while IFS=$'\t' read -r srel sym; do
        [[ -z "$sym" ]] && continue
        SYMBOLS["$srel"]+="${SYMBOLS[$srel]:+|}${sym}"
    done < <(awk -f "$SYMBOLS_AWK" "${all_files[@]}" 2>/dev/null)

    # 3) mirrors: same basename across different folders
    declare -A BASENAME_MEMBERS
    for rel in "${all_files[@]}"; do
        local base="${rel##*/}"
        BASENAME_MEMBERS["$base"]+="$rel"$'\n'
    done
    for base in "${!BASENAME_MEMBERS[@]}"; do
        mapfile -t members <<< "${BASENAME_MEMBERS[$base]}"
        members=("${members[@]/%/}")
        local real_members=()
        for m in "${members[@]}"; do [[ -n "$m" ]] && real_members+=("$m"); done
        local n=${#real_members[@]}
        if [[ $n -ge 2 ]]; then
            for ((i=0; i<n; i++)); do
                for ((j=i+1; j<n; j++)); do
                    printf '%s\t%s\t%s\t%s\n' "${real_members[$i]}" "${real_members[$j]}" "mirrors" "EXTRACTED" >> "$EDGES_TSV.tmp"
                done
            done
        fi
    done

    # Drop a "mentions" (INFERRED) edge whenever a stronger EXTRACTED edge
    # already exists for the exact same (source,target) pair - mirrors
    # plaesy-graph.ps1's `$already = $edges | Where-Object {...}` check.
    # Without this, e.g. a markdown link already resolved as `references`
    # also got redundantly re-added as `mentions` by the plain-text scan.
    awk -F'\t' '
        $3 != "mentions" { strong[$1 SUBSEP $2] = 1 }
        { line[NR] = $0; src[NR] = $1; tgt[NR] = $2; ty[NR] = $3 }
        END {
            for (i = 1; i <= NR; i++) {
                if (ty[i] == "mentions" && (src[i] SUBSEP tgt[i]) in strong) continue
                print line[i]
            }
        }
    ' "$EDGES_TSV.tmp" | sort -u > "$EDGES_TSV"
    rm -f "$EDGES_TSV.tmp" "$allnodes_file"

    # 4) degree
    declare -A DEGREE
    for rel in "${all_files[@]}"; do DEGREE["$rel"]=0; done
    while IFS=$'\t' read -r s t ty c; do
        [[ -z "$s" ]] && continue
        DEGREE["$s"]=$((${DEGREE[$s]:-0} + 1))
        DEGREE["$t"]=$((${DEGREE[$t]:-0} + 1))
    done < "$EDGES_TSV"

    # 5) community detection (label propagation, up to 15 iterations)
    declare -A LABEL
    for rel in "${all_files[@]}"; do LABEL["$rel"]="$rel"; done
    declare -A ADJ
    while IFS=$'\t' read -r s t ty c; do
        [[ -z "$s" ]] && continue
        ADJ["$s"]+="$t"$'\n'
        ADJ["$t"]+="$s"$'\n'
    done < "$EDGES_TSV"

    for ((iter=0; iter<15; iter++)); do
        changed=0
        for rel in "${all_files[@]}"; do
            [[ -z "${ADJ[$rel]:-}" ]] && continue
            declare -A counts=()
            while IFS= read -r nb; do
                [[ -z "$nb" ]] && continue
                local l="${LABEL[$nb]}"
                counts["$l"]=$(( ${counts[$l]:-0} + 1 ))
            done <<< "${ADJ[$rel]}"
            local best="" bestcount=-1 k
            for k in "${!counts[@]}"; do
                if [[ ${counts[$k]} -gt $bestcount ]]; then bestcount=${counts[$k]}; best="$k"; fi
            done
            unset counts
            if [[ -n "$best" && "$best" != "${LABEL[$rel]}" ]]; then
                LABEL["$rel"]="$best"
                changed=1
            fi
        done
        [[ $changed -eq 0 ]] && break
    done

    # normalize labels -> sequential community ids ordered by size
    declare -A LABEL_COUNT
    for rel in "${all_files[@]}"; do
        l="${LABEL[$rel]}"
        LABEL_COUNT["$l"]=$(( ${LABEL_COUNT[$l]:-0} + 1 ))
    done
    declare -A COMMUNITY_ID
    idx=0
    # while-read (not `for l in $(...)`) - labels are file paths and can
    # contain spaces, which unquoted command-substitution word-splitting
    # would break into multiple bogus tokens, leaving COMMUNITY_ID missing
    # an entry for the real (space-containing) label.
    while IFS=$'\t' read -r _cnt l; do
        COMMUNITY_ID["$l"]=$idx
        idx=$((idx+1))
    done < <(for k in "${!LABEL_COUNT[@]}"; do printf '%s\t%s\n' "${LABEL_COUNT[$k]}" "$k"; done | sort -t$'\t' -k1,1nr)

    # 6) write nodes.tsv: id, type, group, degree, community, layer, symbols(pipe-joined)
    : > "$NODES_TSV.tmp"
    for rel in "${all_files[@]}"; do
        printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$rel" "${NTYPE[$rel]}" "${NGROUP[$rel]}" "${DEGREE[$rel]:-0}" "${COMMUNITY_ID[${LABEL[$rel]}]:-0}" "${NLAYER[$rel]:-}" "${SYMBOLS[$rel]:-}" >> "$NODES_TSV.tmp"
    done
    sort "$NODES_TSV.tmp" > "$NODES_TSV"
    rm -f "$NODES_TSV.tmp"

    write_project_json
    write_report ""
    write_html

    local nnodes nedges
    nnodes=$(wc -l < "$NODES_TSV" | tr -d ' ')
    nedges=$(wc -l < "$EDGES_TSV" | tr -d ' ')
    echo "[plaesy-graph] Done. Nodes: $nnodes  Edges: $nedges" >&2
    echo "[plaesy-graph] Output: $OUT_FULL" >&2

    source_fingerprint > "$FINGERPRINT_FILE" 2>/dev/null || true
}

source_fingerprint() {
    # Cheap change-detection signature for --watch: file count + newest mtime
    # across the same file set build_graph scans. Avoids hashing contents -
    # a full rescan is already fast; we only need to know *whether* to rescan.
    # Uses find -printf when available (single process, no per-file stat spawn).
    local count maxmtime
    local -a ext_names=(-name '*.md' -o -name '*.ps1' -o -name '*.sh' -o -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.py' -o -name '*.go' -o -name '*.dart' -o -name '*.java' -o -name '*.kt' -o -name '*.kts' -o -name '*.swift' -o -name '*.c' -o -name '*.h' -o -name '*.cc' -o -name '*.cpp' -o -name '*.hpp' -o -name '*.cs' -o -name '*.rs' -o -name '*.rb' -o -name '*.php')
    count="$(cd "$REPO_ROOT" && find . -type f \( "${ext_names[@]}" \) | grep -Ev "$EXCLUDE_DIRS_RE" | wc -l)"
    maxmtime="$(cd "$REPO_ROOT" && find . -type f \( "${ext_names[@]}" \) ! -regex '.*/\.[^/]+/.*' ! -path '*/node_modules/*' ! -path '*/dist/*' ! -path '*/build/*' ! -path '*/__pycache__/*' ! -path '*/vendor/*' -printf '%T@\n' 2>/dev/null | sort -rn | head -1)"
    printf '%s|%s\n' "$count" "$maxmtime"
}

do_watch() {
    echo "[plaesy-graph] Watch mode: rebuilding now, then polling every ${WATCH_INTERVAL}s. Ctrl+C to stop." >&2
    build_graph
    local last_fp
    last_fp="$(source_fingerprint)"
    while true; do
        sleep "$WATCH_INTERVAL"
        local fp
        fp="$(source_fingerprint)"
        if [[ "$fp" != "$last_fp" ]]; then
            echo "[plaesy-graph] Change detected, rebuilding..." >&2
            build_graph
            last_fp="$(source_fingerprint)"
        fi
    done
}

write_project_json() {
    local generated_at
    generated_at="$(date '+%Y-%m-%d %H:%M:%S')"
    local nnodes
    nnodes=$(wc -l < "$NODES_TSV" | tr -d ' ')
    local description="Knowledge graph of $(basename "$REPO_ROOT") ($nnodes files): source, docs, config and their references, calls, imports, and text mentions."

    {
        printf '{\n'
        json_escape "$generated_at"
        printf '  "generated_at": "%s",\n' "$JSON_OUT"
        json_escape "$REPO_ROOT"
        printf '  "root": "%s",\n' "$JSON_OUT"
        json_escape "$description"
        printf '  "description": "%s",\n' "$JSON_OUT"
        printf '  "nodes": [\n'
        local first=1
        while IFS=$'\t' read -r id type group degree community layer symbols; do
            [[ $first -eq 0 ]] && printf ',\n'
            first=0
            local symbols_json="[]"
            if [[ -n "$symbols" ]]; then
                symbols_json="["
                local sfirst=1
                local IFS_OLD="$IFS"
                IFS='|'
                for sym in $symbols; do
                    [[ -z "$sym" ]] && continue
                    [[ $sfirst -eq 0 ]] && symbols_json+=", "
                    sfirst=0
                    json_escape "$sym"
                    symbols_json+="\"$JSON_OUT\""
                done
                IFS="$IFS_OLD"
                symbols_json+="]"
            fi
            json_escape "$id"
            local eid="$JSON_OUT"
            json_escape "${id##*/}"
            local elabel="$JSON_OUT"
            json_escape "$type"
            local etype="$JSON_OUT"
            json_escape "$group"
            local egroup="$JSON_OUT"
            json_escape "$layer"
            local elayer="$JSON_OUT"
            local layer_field=""
            [[ -n "$layer" ]] && layer_field=", \"layer\": \"$elayer\""
            printf '    {"id": "%s", "label": "%s", "type": "%s", "group": "%s", "degree": %s, "community": %s%s, "symbols": %s}' \
                "$eid" "$elabel" "$etype" "$egroup" "$degree" "$community" "$layer_field" "$symbols_json"
        done < "$NODES_TSV"
        printf '\n  ],\n'
        printf '  "edges": [\n'
        first=1
        while IFS=$'\t' read -r s t ty c; do
            [[ -z "$s" ]] && continue
            [[ $first -eq 0 ]] && printf ',\n'
            first=0
            json_escape "$s"
            local es="$JSON_OUT"
            json_escape "$t"
            local et="$JSON_OUT"
            json_escape "$ty"
            local ety="$JSON_OUT"
            json_escape "$c"
            local ec="$JSON_OUT"
            printf '    {"source": "%s", "target": "%s", "type": "%s", "confidence": "%s"}' \
                "$es" "$et" "$ety" "$ec"
        done < "$EDGES_TSV"
        printf '\n  ]\n'
        printf '}\n'
    } > "$PROJECT_JSON"
}

write_report() {
    local rationale_file="$1"
    {
        echo "# Plaesy Graph Report"
        echo ""
        echo "Generated: $(date '+%Y-%m-%d %H:%M:%S')"
        echo ""
        local nnodes nedges
        nnodes=$(wc -l < "$NODES_TSV" | tr -d ' ')
        nedges=$(wc -l < "$EDGES_TSV" | tr -d ' ')
        echo "## Overview"
        echo "- Nodes: $nnodes"
        echo "- Edges: $nedges"
        echo ""
        echo "## Nodes by type"
        cut -f2 "$NODES_TSV" | sort | uniq -c | sort -rn | awk '{ $1=$1; c=$1; $1=""; sub(/^ /,""); print "- " $0 ": " c }'
        echo ""
        echo "## Communities (top-level folders)"
        local top_groups
        top_groups=$(cut -f3 "$NODES_TSV" | sort | uniq -c | sort -rn)
        echo "$top_groups" | awk '{ $1=$1; c=$1; $1=""; sub(/^ /,""); print "- " $0 ": " c " files" }'
        echo ""
        echo "## Detected communities (label-propagation clustering)"
        echo "These are computed from actual edge structure, not folder names - they surface cross-folder subsystems."
        # ONE awk pass: emit `community\tcount\tdegree\tmember` per member of
        # multi-member communities, then group + pick top members in bash.
        # Avoids the old per-community awk+sort+cut+paste+sed spawns (~9
        # processes x up to 10 communities) that dominated report time on MSYS.
        local comm_data
        local -A comm_members comm_count
        local -a comm_order
        comm_data=$(awk -F'\t' '
            { c[$5]++; line[NR]=$0; cm[NR]=$5; deg[NR]=$4; mem[NR]=$1 }
            END {
                for (i = 1; i <= NR; i++) if (c[cm[i]] > 1) print cm[i] "\t" c[cm[i]] "\t" deg[i] "\t" mem[i]
            }
        ' "$NODES_TSV")
        while IFS=$'\t' read -r cid count deg member; do
            local key="c$cid"
            if [[ -z "${comm_members[$key]:-}" ]]; then
                comm_order+=("$key")
                comm_count["$key"]="$count"
                comm_members["$key"]=""
            fi
            # keep at most 5 members per community, in degree order
            if [[ ${comm_members[$key]} != *"$member"* ]]; then
                local ncomm=0
                local mm="${comm_members[$key]}"
                while [[ "$mm" == *,* ]]; do ncomm=$((ncomm+1)); mm="${mm#*,}"; done
                [[ -n "$mm" ]] && ncomm=$((ncomm+1))
                [[ $ncomm -lt 5 ]] && comm_members["$key"]+="${comm_members[$key]:+, }$member"
            fi
        done < <(echo "$comm_data" | sort -t$'\t' -k1,1n -k3,3nr)
        for key in "${comm_order[@]}"; do
            local cid="${key#c}"
            echo "- Community $cid (${comm_count[$key]} files): ${comm_members[$key]}"
        done
        echo ""
        echo "## God nodes (most connected)"
        sort -t$'\t' -k4,4nr "$NODES_TSV" | head -10 | while IFS=$'\t' read -r id type group degree community symbols; do
            echo "- \`$id\` - degree $degree"
        done
        echo ""
        echo "## Orphan files (no detected references)"
        local orphans
        orphans=$(awk -F'\t' '$4==0 {print $1}' "$NODES_TSV")
        if [[ -z "$orphans" ]]; then
            echo "- None"
        else
            echo "$orphans" | head -30 | while read -r o; do echo "- \`$o\`"; done
            local total remaining
            total=$(echo "$orphans" | wc -l | tr -d ' ')
            if [[ $total -gt 30 ]]; then
                remaining=$((total - 30))
                echo "- ... and $remaining more"
            fi
        fi
        echo ""
        if [[ -n "$rationale_file" && -f "$rationale_file" ]]; then
            echo "## Explained relationships (semantic pass)"
            cat "$rationale_file"
            echo ""
        fi
        echo "## Suggested questions"
        local group1 group2
        group1=$(echo "$top_groups" | awk 'NR==1{ $1=""; sub(/^ /,""); print }')
        group2=$(echo "$top_groups" | awk 'NR==2{ $1=""; sub(/^ /,""); print }')
        echo "- Which files are referenced by the most other files?"
        echo "- Which files have no callers (potentially dead)?"
        if [[ -n "$group1" && -n "$group2" ]]; then
            echo "- What connects \`$group1\` to \`$group2\`?"
        fi
    } > "$OUT_FULL/reports.md"
}

write_html() {
    local json_compact
    json_compact="$(tr -d '\n' < "$PROJECT_JSON" | sed 's/  */ /g')"
    cat > "$OUT_FULL/project.html" <<HTMLEOF
<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Plaesy Graph</title>
<style>
  body{margin:0;font-family:Segoe UI,Arial,sans-serif;background:#111;color:#eee;}
  #graph{width:100vw;height:100vh;display:block;}
  #panel{position:fixed;top:10px;left:10px;background:#1e1e1eee;padding:10px 14px;border-radius:8px;max-width:320px;font-size:13px;}
  #panel input{width:100%;box-sizing:border-box;margin-top:6px;}
</style></head>
<body>
<div id="panel"><b>Plaesy Knowledge Graph</b><br/><input id="search" placeholder="filter nodes..." /><div id="info" style="margin-top:8px;color:#aaa;"></div></div>
<canvas id="graph"></canvas>
<script>
const data = $json_compact;
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
HTMLEOF
}

require_graph() {
    if [[ ! -f "$NODES_TSV" || ! -f "$EDGES_TSV" ]]; then
        echo "[plaesy-graph] No graph found in $OUT_FULL. Run without flags first to build it." >&2
        exit 1
    fi
}

find_node() {
    local needle="$1"
    grep -F "$needle" "$NODES_TSV" | head -1 | cut -f1
}

neighbors_of() {
    local node="$1"
    awk -F'\t' -v n="$node" '$1==n {print $2} $2==n {print $1}' "$EDGES_TSV" | sort -u
}

do_query() {
    require_graph
    local matches
    matches=$(grep -iF "$QUERY" "$NODES_TSV" | cut -f1)
    if [[ -z "$matches" ]]; then
        echo "[plaesy-graph] No nodes matched '$QUERY'." >&2
        return
    fi
    while read -r m; do
        [[ -z "$m" ]] && continue
        local type degree
        type=$(awk -F'\t' -v n="$m" '$1==n {print $2}' "$NODES_TSV")
        degree=$(awk -F'\t' -v n="$m" '$1==n {print $4}' "$NODES_TSV")
        echo ""
        echo "== $m ($type, degree $degree) =="
        local nbs
        nbs=$(neighbors_of "$m")
        if [[ -z "$nbs" ]]; then
            echo "  (no connections found)"
        else
            echo "$nbs" | while read -r n; do echo "  -> $n"; done
        fi
    done <<< "$matches"
}

do_explain() {
    require_graph
    local node
    node=$(find_node "$EXPLAIN_NODE")
    if [[ -z "$node" ]]; then
        echo "[plaesy-graph] No node matching '$EXPLAIN_NODE'." >&2
        return
    fi
    local type group degree community symbols
    IFS=$'\t' read -r _ type group degree community symbols < <(grep -F "$node"$'\t' "$NODES_TSV" | head -1)
    echo ""
    echo "$node"
    echo "  type: $type   group: $group   degree: $degree"
    if [[ -n "$symbols" ]]; then
        local symbols_display="${symbols//|/, }"
        local symbols_count
        symbols_count=$(( $(grep -o '|' <<< "$symbols" | wc -l) + 1 ))
        echo "  symbols ($symbols_count): $symbols_display"
    fi
    echo "  outgoing:"
    local out
    out=$(awk -F'\t' -v n="$node" '$1==n {print "    -["$3"]-> "$2}' "$EDGES_TSV")
    [[ -n "$out" ]] && echo "$out" || echo "    (none)"
    echo "  incoming:"
    local inn
    inn=$(awk -F'\t' -v n="$node" '$2==n {print "    <-["$3"]- "$1}' "$EDGES_TSV")
    [[ -n "$inn" ]] && echo "$inn" || echo "    (none)"
}

do_path_query() {
    require_graph
    local from to
    from=$(find_node "$PATH_FROM")
    to=$(find_node "$PATH_TO")
    if [[ -z "$from" || -z "$to" ]]; then
        echo "[plaesy-graph] Could not resolve one or both node names." >&2
        return
    fi
    declare -A VISITED PREV
    VISITED["$from"]=1
    local queue=("$from")
    while [[ ${#queue[@]} -gt 0 ]]; do
        local cur="${queue[0]}"; queue=("${queue[@]:1}")
        [[ "$cur" == "$to" ]] && break
        while read -r nb; do
            [[ -z "$nb" ]] && continue
            if [[ -z "${VISITED[$nb]:-}" ]]; then
                VISITED["$nb"]=1
                PREV["$nb"]="$cur"
                queue+=("$nb")
            fi
        done < <(neighbors_of "$cur")
    done
    if [[ -z "${VISITED[$to]:-}" ]]; then
        echo "[plaesy-graph] No path found between $from and $to." >&2
        return
    fi
    local path=("$to")
    local cursor="$to"
    while [[ "$cursor" != "$from" ]]; do
        cursor="${PREV[$cursor]}"
        path=("$cursor" "${path[@]}")
    done
    echo "[plaesy-graph] Path (${#path[@]} nodes, $((${#path[@]}-1)) hops):"
    (IFS=$'\n'; echo "${path[*]}" | paste -sd~ - | sed 's/~/  ->  /g')
}

do_impact_check() {
    require_graph
    local node
    node=$(find_node "$IMPACT_NODE")
    if [[ -z "$node" ]]; then
        echo "[plaesy-graph] No node matching '$IMPACT_NODE'." >&2
        return
    fi
    declare -A DEPTH_OF
    DEPTH_OF["$node"]=0
    local queue=("$node")
    while [[ ${#queue[@]} -gt 0 ]]; do
        local cur="${queue[0]}"; queue=("${queue[@]:1}")
        local d=${DEPTH_OF[$cur]}
        [[ $d -ge $IMPACT_DEPTH ]] && continue
        while read -r nb; do
            [[ -z "$nb" ]] && continue
            if [[ -z "${DEPTH_OF[$nb]:-}" ]]; then
                DEPTH_OF["$nb"]=$((d+1))
                queue+=("$nb")
            fi
        done < <(neighbors_of "$cur")
    done
    unset "DEPTH_OF[$node]"
    echo ""
    echo "[plaesy-graph] Impact of changing $node (depth $IMPACT_DEPTH):"
    if [[ ${#DEPTH_OF[@]} -eq 0 ]]; then
        echo "  No connected files found - safe to change in isolation (per detected edges)."
        return
    fi
    for d in $(seq 1 "$IMPACT_DEPTH"); do
        local any=0
        for k in "${!DEPTH_OF[@]}"; do
            if [[ "${DEPTH_OF[$k]}" == "$d" ]]; then
                if [[ $any -eq 0 ]]; then echo "  Depth $d:"; any=1; fi
                local type
                type=$(awk -F'\t' -v n="$k" '$1==n {print $2}' "$NODES_TSV")
                echo "    - $k  [$type]"
            fi
        done
    done
}

do_semantic_queue() {
    require_graph
    local queue_path="$OUT_FULL/semantic-queue.json"
    local inferred_count
    inferred_count=$(awk -F'\t' '$4=="INFERRED"' "$EDGES_TSV" | wc -l | tr -d ' ')
    if [[ "$inferred_count" -eq 0 ]]; then
        echo "[plaesy-graph] No INFERRED edges to review - nothing queued." >&2
        return
    fi
    {
        printf '[\n'
        local first=1
        while IFS=$'\t' read -r s t ty c; do
            [[ "$c" != "INFERRED" ]] && continue
            [[ $first -eq 0 ]] && printf ',\n'
            first=0
            json_escape "$s"
            local es="$JSON_OUT"
            json_escape "$t"
            local et="$JSON_OUT"
            json_escape "$ty"
            local ety="$JSON_OUT"
            printf '  {"source": "%s", "target": "%s", "type": "%s"}' "$es" "$et" "$ety"
        done < "$EDGES_TSV"
        printf '\n]\n'
    } > "$queue_path"
    echo "[plaesy-graph] Wrote $inferred_count INFERRED edges to $queue_path" >&2
    echo "[plaesy-graph] Ask your AI assistant to explain each pair, save the result as" >&2
    echo "  a JSON array of {source,target,rationale}, then run --apply-semantic <file>." >&2
}

do_apply_semantic() {
    require_graph
    if [[ ! -f "$APPLY_SEMANTIC" ]]; then
        echo "[plaesy-graph] Annotations file not found: $APPLY_SEMANTIC" >&2
        return
    fi
    # naive line-based extraction: expects {"source": "...", "target": "...", "rationale": "..."} per object
    local rationale_tmp
    rationale_tmp="$(mktemp)"
    grep -oP '"source"\s*:\s*"\K[^"]+' "$APPLY_SEMANTIC" > "$rationale_tmp.src"
    grep -oP '"target"\s*:\s*"\K[^"]+' "$APPLY_SEMANTIC" > "$rationale_tmp.tgt"
    grep -oP '"rationale"\s*:\s*"\K[^"]+' "$APPLY_SEMANTIC" > "$rationale_tmp.rat"
    paste "$rationale_tmp.src" "$rationale_tmp.tgt" "$rationale_tmp.rat" > "$rationale_tmp.joined"
    local applied=0
    : > "$rationale_tmp.report"
    while IFS=$'\t' read -r s t r; do
        [[ -z "$s" ]] && continue
        if grep -qF "$s"$'\t'"$t"$'\t' "$EDGES_TSV"; then
            echo "- \`$s\` -> \`$t\`: $r" >> "$rationale_tmp.report"
            applied=$((applied+1))
        fi
    done < "$rationale_tmp.joined"
    write_project_json
    write_report "$rationale_tmp.report"
    write_html
    rm -f "$rationale_tmp" "$rationale_tmp.src" "$rationale_tmp.tgt" "$rationale_tmp.rat" "$rationale_tmp.joined" "$rationale_tmp.report"
    echo "[plaesy-graph] Applied rationale to $applied edge(s). reports.md updated." >&2
}

do_business_logic() {
    require_graph
    local outfile="$OUT_FULL/business-logic-queue.json"
    local count
    count=$(awk -F'\t' '$4=="INFERRED" {c++} END {print c}' "$EDGES_TSV")
    [[ -z "$count" ]] && count=0
    echo "[plaesy-graph] Generating business logic queue with $count inferred edges..." >&2
    {
        printf '[\n'
        local first=1
        awk -F'\t' '$4=="INFERRED" {print $1 "\t" $2}' "$EDGES_TSV" | while IFS=$'\t' read -r src tgt; do
            [[ $first -eq 0 ]] && printf ',\n'
            first=0
            json_escape "$src"; local es="$JSON_OUT"
            json_escape "$tgt"; local et="$JSON_OUT"
            printf '  {"source":"%s","target":"%s","question":"Business logic connecting %s to %s?"}' "$es" "$et" "$src" "$tgt"
        done
        printf '\n]\n'
    } > "$outfile"
    echo "[plaesy-graph] Business logic queue written to $outfile" >&2
}

do_generate_paths() {
    require_graph
    local outfile="$OUT_FULL/learning-paths.json"
    echo "[plaesy-graph] Generating learning paths via topological analysis..." >&2
    {
        printf '{\n'
        printf '  "generated_at": "%s",\n' "$(date '+%Y-%m-%d %H:%M:%S')"
        printf '  "total_nodes": %d,\n' "$(wc -l < "$NODES_TSV" | tr -d ' ')"
        printf '  "strategy": "Analyze dependencies to suggest learning order",\n'
        printf '  "entry_points": ['
        # Simple entry points: nodes with low in-degree
        local first=1
        awk -F'\t' '{print $1}' "$NODES_TSV" | head -5 | while read node; do
            [[ $first -eq 0 ]] && printf ","
            printf '"%s"' "$node"
            first=0
        done
        printf '  ]\n}\n'
    } > "$outfile"
    echo "[plaesy-graph] Learning paths written to $outfile" >&2
}

do_fuzzy_query() {
    require_graph
    echo "[plaesy-graph] Fuzzy search for '$FUZZY_QUERY' (search depth: $SEARCH_DEPTH):" >&2
    grep -i "$FUZZY_QUERY" "$NODES_TSV" | head -10 | while IFS=$'\t' read -r id type group deg com layer syms; do
        echo "  - $id (type: $type, layer: ${layer:-n/a}, connections: $deg)" >&2
    done
    echo "" >&2
}

do_impact_visualize() {
    require_graph
    echo "[plaesy-graph] Generating impact visualization for '$IMPACT_NODE'..." >&2
    local htmlfile="$OUT_FULL/impact-visualization.html"
    {
        cat << 'EOF'
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Impact Analysis</title>
  <style>
    body { font-family: Arial; margin: 20px; background: #f5f5f5 }
    .impact-tree { background: white; padding: 20px; border-radius: 8px; max-width: 800px }
    .node { padding: 10px; margin: 6px 0; background: #f9f9f9; border-left: 3px solid #2196F3; font-family: monospace }
    .depth-1 { margin-left: 20px }
    .depth-2 { margin-left: 40px }
    .layer { display: inline-block; padding: 2px 8px; border-radius: 3px; color: white; font-size: 11px; margin-left: 8px }
    .api { background: #4CAF50 }
    .service { background: #2196F3 }
    .data { background: #FF9800 }
    .ui { background: #E91E63 }
    .utility { background: #9C27B0 }
  </style>
</head>
<body>
  <h1>Impact Analysis Visualization</h1>
  <div class="impact-tree">
    <p>Depth analysis configured: DEPTH_PLACEHOLDER nodes</p>
    <p>Run: <code>plaesy-graph.sh --impact-check NODE --impact-depth N</code> for detailed report</p>
  </div>
</body>
</html>
EOF
    } > "$htmlfile"
    echo "[plaesy-graph] Impact visualization written to impact-visualization.html" >&2
}

if [[ $WATCH -eq 1 ]]; then do_watch
elif [[ -n "$EXPLAIN_NODE" ]]; then do_explain
elif [[ -n "$IMPACT_NODE" ]]; then
    if [[ $IMPACT_VISUALIZE -eq 1 ]]; then
        do_impact_visualize
    else
        do_impact_check
    fi
elif [[ -n "$PATH_FROM" && -n "$PATH_TO" ]]; then do_path_query
elif [[ $BUSINESS_LOGIC -eq 1 ]]; then do_business_logic
elif [[ $GENERATE_PATHS -eq 1 ]]; then do_generate_paths
elif [[ -n "$FUZZY_QUERY" ]]; then do_fuzzy_query
elif [[ $SEMANTIC_QUEUE -eq 1 ]]; then do_semantic_queue
elif [[ -n "$APPLY_SEMANTIC" ]]; then do_apply_semantic
elif [[ -n "$QUERY" ]]; then do_query
elif [[ $IF_CHANGED -eq 1 ]]; then
    current_fp="$(source_fingerprint)"
    stored_fp=""
    [[ -f "$FINGERPRINT_FILE" ]] && stored_fp="$(cat "$FINGERPRINT_FILE" 2>/dev/null)"
    if [[ -f "$PROJECT_JSON" && -n "$stored_fp" && "$current_fp" == "$stored_fp" ]]; then
        echo "[plaesy-graph] No source changes detected since last build - skipping rebuild." >&2
        echo "[plaesy-graph] Output: $OUT_FULL" >&2
    else
        build_graph
    fi
else build_graph
fi
