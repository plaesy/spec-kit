#!/usr/bin/env bash
# plaesy-trim.sh - Token/context compression for the plaesy spec-kit repo
# Version: 1.1.0
# No external dependency (jq, Python, network) required - grep/awk/sed only,
# matching plaesy-graph.sh's zero-dependency approach.
#
# Layer 1 (command output): run <command...>              - per-tool aware dedupe/truncate/filter
# Layer 2 (memory/instruction files): compress --path ...  - heuristic denser style (fast, local)
#   llm-queue / apply-llm                                    - LLM-quality rewrite via the calling
#                                                               assistant, no API key (same pattern as
#                                                               plaesy-graph.sh's --semantic-queue/--apply-semantic)
# Report: report                                            - cumulative savings across layers
#
# Usage:
#   plaesy-trim.sh run git status
#   plaesy-trim.sh run pytest
#   plaesy-trim.sh compress --path .plaesy/memory/plaesy.md
#   plaesy-trim.sh compress --path .plaesy/memory --recurse --dry-run
#   plaesy-trim.sh compress --path README.md --level ultra
#   plaesy-trim.sh llm-queue --path .plaesy/memory/plaesy.md
#   plaesy-trim.sh apply-llm --path .plaesy/memory/plaesy.md --annotations annotations.json
#   plaesy-trim.sh report

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRAMEWORK_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
# REPO_ROOT is the project being worked on (current directory), NOT where this script is
# installed. Framework-shipped rules are read from FRAMEWORK_ROOT as a fallback only, so
# plaesy-trim works correctly when invoked from any other project via its installed path,
# matching plaesy-graph.sh's cwd-relative behavior.
REPO_ROOT="$(pwd)"
# Namespaced under .plaesy/ (per platform.json's "base_directory": ".plaesy" convention), not a
# bare "scripts/" folder - a target project may already have its own unrelated scripts/ dir.
PROJECT_RULES_PATH="$REPO_ROOT/.plaesy/scripts/configs/plaesy-trim-rules.json"
FRAMEWORK_RULES_PATH="$FRAMEWORK_ROOT/scripts/configs/plaesy-trim-rules.json"
if [[ -f "$PROJECT_RULES_PATH" ]]; then RULES_PATH="$PROJECT_RULES_PATH"; else RULES_PATH="$FRAMEWORK_RULES_PATH"; fi
GRAPH_PATH="$REPO_ROOT/.plaesy/memory/analysis/project.graph.json"
STATS_PATH="$REPO_ROOT/.plaesy/memory/token-stats.json"

estimate_tokens() {
    local text="$1"
    local len=${#text}
    echo $(( (len + 3) / 4 ))
}

pct_saved() {
    local before="$1" after="$2"
    if [[ "$before" -gt 0 ]]; then
        awk -v b="$before" -v a="$after" 'BEGIN{printf "%.1f", (1-(a/b))*100}'
    else
        echo "0.0"
    fi
}

add_stat_record() {
    local layer="$1" target="$2" before="$3" after="$4"
    mkdir -p "$(dirname "$STATS_PATH")"
    if [[ ! -f "$STATS_PATH" ]]; then printf '[]\n' > "$STATS_PATH"; fi
    local pct ts esc_target
    pct=$(pct_saved "$before" "$after")
    ts=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    esc_target=$(printf '%s' "$target" | sed 's/\\/\\\\/g; s/"/\\"/g')
    local record="{\"timestamp\":\"$ts\",\"layer\":\"$layer\",\"path\":\"$esc_target\",\"before\":$before,\"after\":$after,\"saved_pct\":$pct}"
    local tmp
    tmp=$(mktemp)
    if grep -q '^\[\]$' "$STATS_PATH" || [[ ! -s "$STATS_PATH" ]]; then
        printf '[\n  %s\n]\n' "$record" > "$tmp"
    else
        # replace trailing "]" with ",\n  <record>\n]"
        sed '$ s/^\]$/,\n  '"$(printf '%s' "$record" | sed 's/[&/\]/\\&/g')"'\n]/' "$STATS_PATH" > "$tmp"
    fi
    mv "$tmp" "$STATS_PATH"
}

# ---------- Layer 1: command output compression ----------

compress_command_output() {
    local max_lines="$1" dedupe="$2" mode="${3:-generic}"
    local -a collapsed=()
    local prev="" count=0 first=1 line

    if [[ "$mode" == "test-results" ]]; then
        local -a kept=()
        local -a all_lines=()
        while IFS= read -r line || [[ -n "$line" ]]; do all_lines+=("$line"); done
        for line in "${all_lines[@]}"; do
            if [[ "$line" =~ [Ff][Aa][Ii][Ll]|[Ee][Rr][Rr][Oo][Rr]|✗|✘ ]]; then kept+=("$line"); fi
        done
        local total_all=${#all_lines[@]}
        local tail_start=$(( total_all > 5 ? total_all - 5 : 0 ))
        local i
        for (( i = tail_start; i < total_all; i++ )); do
            if [[ "${all_lines[$i]}" =~ [Pp]assed|[Ff]ailed|[Tt]ests?\ run|ok\. ]]; then kept+=("${all_lines[$i]}"); fi
        done
        if [[ ${#kept[@]} -eq 0 ]]; then
            printf '%s\n' "(no failures - all tests passed, output suppressed)"
            return
        fi
        local -A seen=()
        local -a dedup_kept=()
        for line in "${kept[@]}"; do
            [[ -n "${seen[$line]:-}" ]] && continue
            seen[$line]=1
            dedup_kept+=("$line")
        done
        printf '%s\n' "${dedup_kept[@]}"
        return
    fi

    while IFS= read -r line || [[ -n "$line" ]]; do
        if [[ "$dedupe" != "true" ]]; then
            collapsed+=("$line")
            continue
        fi
        if [[ "$first" -eq 1 ]]; then
            prev="$line"; count=1; first=0
            continue
        fi
        if [[ "$line" == "$prev" ]]; then
            count=$((count + 1))
        else
            if [[ "$count" -gt 1 ]]; then collapsed+=("$prev (x$count)"); else collapsed+=("$prev"); fi
            prev="$line"; count=1
        fi
    done
    if [[ "$dedupe" == "true" && "$first" -eq 0 ]]; then
        if [[ "$count" -gt 1 ]]; then collapsed+=("$prev (x$count)"); else collapsed+=("$prev"); fi
    fi

    local total=${#collapsed[@]}
    if [[ "$total" -le "$max_lines" ]]; then
        printf '%s\n' "${collapsed[@]}"
        return
    fi

    local head=$(( (max_lines + 1) / 2 ))
    local tail=$(( max_lines - head ))
    local omitted=$(( total - head - tail ))
    printf '%s\n' "${collapsed[@]:0:$head}"
    echo "... $omitted lines omitted ..."
    printf '%s\n' "${collapsed[@]:$((total - tail)):$tail}"
}

get_layer1_keys() {
    grep -oP '^\s*"\K[^"]+(?="\s*:\s*\{\s*"max_lines")' "$RULES_PATH"
}

get_layer1_rule() {
    local key="$1"
    local line
    line=$(grep -P "^\s*\"$(printf '%s' "$key" | sed 's/[.[\*^$]/\\&/g')\"\s*:\s*\{\s*\"max_lines\"" "$RULES_PATH")
    echo "$line" | grep -oP '"max_lines":\s*\K[0-9]+"?'
    echo "$line" | grep -oP '"dedupe_consecutive":\s*\K(true|false)'
    echo "$line" | grep -oP '"mode":\s*"\K[^"]+' || echo "generic"
}

cmd_run() {
    if [[ $# -eq 0 ]]; then echo "Usage: plaesy-trim.sh run <command...>" >&2; exit 1; fi
    local cmd_string="$*"

    local matched_key="" matched_len=0
    while IFS= read -r key; do
        [[ "$key" == "_default" ]] && continue
        if [[ "$cmd_string" == "$key"* ]]; then
            if [[ ${#key} -gt "$matched_len" ]]; then matched_key="$key"; matched_len=${#key}; fi
        fi
    done < <(get_layer1_keys)

    local max_lines dedupe mode
    if [[ -n "$matched_key" ]]; then
        read -r max_lines dedupe mode < <(get_layer1_rule "$matched_key" | tr '\n' ' ')
    else
        read -r max_lines dedupe mode < <(get_layer1_rule "_default" | tr '\n' ' ')
    fi
    max_lines=${max_lines:-40}
    dedupe=${dedupe:-true}
    mode=${mode:-generic}

    local raw_output
    raw_output="$("$@" 2>&1)"
    local compressed
    compressed="$(printf '%s\n' "$raw_output" | compress_command_output "$max_lines" "$dedupe" "$mode")"

    local before after
    before=$(estimate_tokens "$raw_output")
    after=$(estimate_tokens "$compressed")
    add_stat_record "layer1" "$cmd_string" "$before" "$after"

    printf '%s\n' "$compressed"
}

# ---------- Layer 2: prose compression ----------

get_node_degree() {
    local relative="$1"
    [[ -f "$GRAPH_PATH" ]] || return 0
    local win_path="${relative//\//\\\\}"
    awk -v target="$win_path" '
        index($0, "\"id\"") && index($0, target) { found=1 }
        found && index($0, "\"degree\"") {
            line=$0
            gsub(/[^0-9]/, "", line)
            print line
            exit
        }
    ' "$GRAPH_PATH"
}

resolve_level() {
    local relative="$1" requested="$2"
    if [[ -n "$requested" ]]; then echo "$requested"; return; fi
    local degree
    degree=$(get_node_degree "$relative")
    if [[ -z "$degree" ]]; then echo "full"; return; fi
    local lite_min full_min
    lite_min=$(grep -oP '"lite_min_degree":\s*\K[0-9]+' "$RULES_PATH")
    full_min=$(grep -oP '"full_min_degree":\s*\K[0-9]+' "$RULES_PATH")
    if [[ "$degree" -ge "$lite_min" ]]; then echo "lite";
    elif [[ "$degree" -ge "$full_min" ]]; then echo "full";
    else echo "ultra"; fi
}

compress_prose() {
    local text="$1" level="$2"
    local in_code=0 out="" buf="" line

    while IFS= read -r line || [[ -n "$line" ]]; do
        if [[ "$line" =~ ^\`\`\` ]]; then
            if [[ "$in_code" -eq 1 ]]; then
                buf+="$line"$'\n'
                out+="$buf"
                buf=""
                in_code=0
            else
                out+="$(apply_level "$buf" "$level")"
                buf="$line"$'\n'
                in_code=1
            fi
            continue
        fi
        buf+="$line"$'\n'
    done <<< "$text"

    if [[ "$in_code" -eq 1 ]]; then
        out+="$buf"
    else
        out+="$(apply_level "$buf" "$level")"
    fi
    printf '%s' "$out"
}

apply_level() {
    local piece="$1" level="$2"
    if [[ "$level" == "lite" || "$level" == "full" || "$level" == "ultra" ]]; then
        while IFS=$'\t' read -r find replace; do
            [[ -z "$find" ]] && continue
            piece="$(printf '%s' "$piece" | sed -E "s/$find/$replace/gI" 2>/dev/null || printf '%s' "$piece")"
        done < <(grep -oP '"find":\s*"\K[^"]*(?=")' "$RULES_PATH" | paste -d '\t' - <(grep -oP '"replace":\s*"\K[^"]*(?=")' "$RULES_PATH"))
    fi
    if [[ "$level" == "full" || "$level" == "ultra" ]]; then
        piece="$(printf '%s' "$piece" | sed -E 's/ {2,}/ /g')"
    fi
    if [[ "$level" == "ultra" ]]; then
        piece="$(printf '%s' "$piece" | sed -E 's/\b(that|which|very|really|basically|simply)\b[[:space:]]*//gI' 2>/dev/null || printf '%s' "$piece")"
    fi
    piece="$(printf '%s' "$piece" | sed -E ':a;N;$!ba; s/\n{3,}/\n\n/g')"
    printf '%s' "$piece"
}

compress_one_file() {
    local file="$1" requested_level="$2" dry_run="$3"
    local original
    original="$(cat "$file")"
    local relative="${file#"$REPO_ROOT"/}"
    local level
    level=$(resolve_level "$relative" "$requested_level")
    local compressed
    compressed="$(compress_prose "$original" "$level")"

    local before after pct
    before=$(estimate_tokens "$original")
    after=$(estimate_tokens "$compressed")
    pct=$(pct_saved "$before" "$after")

    echo "$relative [$level]"
    echo "  $before tokens -> $after tokens  ($pct% saved)"

    if [[ "$dry_run" != "true" ]]; then
        cp "$file" "$file.bak"
        printf '%s' "$compressed" > "$file"
        add_stat_record "layer2" "$relative" "$before" "$after"
    fi
}

cmd_compress() {
    local path="" level="" recurse="false" dry_run="false"
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path) path="$2"; shift 2 ;;
            --level) level="$2"; shift 2 ;;
            --recurse) recurse="true"; shift ;;
            --dry-run) dry_run="true"; shift ;;
            *) shift ;;
        esac
    done
    if [[ -z "$path" ]]; then echo "Usage: plaesy-trim.sh compress --path <file|dir> [--level lite|full|ultra] [--recurse] [--dry-run]" >&2; exit 1; fi

    local target="$path"
    [[ -e "$target" ]] || target="$REPO_ROOT/$path"
    [[ -e "$target" ]] || { echo "Path not found: $path" >&2; exit 1; }

    if [[ -d "$target" ]]; then
        if [[ "$recurse" == "true" ]]; then
            while IFS= read -r f; do compress_one_file "$f" "$level" "$dry_run"; done < <(find "$target" -type f -name "*.md")
        else
            while IFS= read -r f; do compress_one_file "$f" "$level" "$dry_run"; done < <(find "$target" -maxdepth 1 -type f -name "*.md")
        fi
    else
        compress_one_file "$target" "$level" "$dry_run"
    fi
}

# ---------- Layer 2 (LLM mode): queue prose for the calling assistant to rewrite ----------

json_escape() {
    printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g' | awk '{printf "%s\\n", $0}' | sed '$ s/\\n$//'
}

cmd_llm_queue() {
    local path="" i=0
    while [[ $# -gt 0 ]]; do
        case "$1" in --path) path="$2"; shift 2 ;; *) shift ;; esac
    done
    [[ -z "$path" ]] && { echo "Usage: plaesy-trim.sh llm-queue --path <file>" >&2; exit 1; }
    local target="$path"
    [[ -e "$target" ]] || target="$REPO_ROOT/$path"
    [[ -e "$target" ]] || { echo "Path not found: $path" >&2; exit 1; }
    local relative="${target#"$REPO_ROOT"/}"

    local out_dir="$REPO_ROOT/.plaesy/memory/analysis"
    mkdir -p "$out_dir"
    local queue_path="$out_dir/trim-queue.json"

    local in_code=0 buf="" line idx=0
    local entries=()
    flush_prose() {
        if [[ "$in_code" -eq 0 && ${#buf} -ge 40 ]]; then
            entries+=("{\"index\":$idx,\"text\":\"$(json_escape "$buf")\"}")
        fi
        idx=$((idx + 1))
    }
    while IFS= read -r line || [[ -n "$line" ]]; do
        if [[ "$line" =~ ^\`\`\` ]]; then
            buf+="$line"$'\n'
            flush_prose
            buf=""
            if [[ "$in_code" -eq 1 ]]; then in_code=0; else in_code=1; fi
            continue
        fi
        buf+="$line"$'\n'
    done < "$target"
    flush_prose

    {
        printf '{"file":"%s","segments":[' "$(json_escape "$relative")"
        local first=1
        for e in "${entries[@]}"; do
            [[ "$first" -eq 0 ]] && printf ','
            printf '%s' "$e"
            first=0
        done
        printf ']}\n'
    } > "$queue_path"

    echo "Wrote ${#entries[@]} prose segment(s) from $relative to $queue_path"
    echo "Next: have the calling assistant rewrite each segment.text denser (same meaning, no code/commands touched),"
    echo "save as {file, segments:[{index, text}]} JSON, then run:"
    echo "  plaesy-trim.sh apply-llm --path $path --annotations <annotations.json>"
}

cmd_apply_llm() {
    local path="" annotations=""
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path) path="$2"; shift 2 ;;
            --annotations) annotations="$2"; shift 2 ;;
            *) shift ;;
        esac
    done
    [[ -z "$path" || -z "$annotations" ]] && { echo "Usage: plaesy-trim.sh apply-llm --path <file> --annotations <annotations.json>" >&2; exit 1; }
    local target="$path"
    [[ -e "$target" ]] || target="$REPO_ROOT/$path"
    [[ -e "$target" ]] || { echo "Path not found: $path" >&2; exit 1; }
    [[ -f "$annotations" ]] || { echo "Annotations file not found: $annotations" >&2; exit 1; }

    local original
    original="$(cat "$target")"
    local relative="${target#"$REPO_ROOT"/}"

    local in_code=0 buf="" line idx=0 out=""
    while IFS= read -r line || [[ -n "$line" ]]; do
        buf+="$line"$'\n'
        if [[ "$line" =~ ^\`\`\` ]]; then
            local rewrite
            rewrite=$(grep -oP "\"index\":$idx,\"text\":\"\K[^\"]*(?=\")" "$annotations" | head -1 || true)
            if [[ "$in_code" -eq 0 && -n "$rewrite" ]]; then
                out+="$(printf '%b' "$rewrite")"
            else
                out+="$buf"
            fi
            buf=""
            idx=$((idx + 1))
            if [[ "$in_code" -eq 1 ]]; then in_code=0; else in_code=1; fi
        fi
    done <<< "$original"
    local rewrite
    rewrite=$(grep -oP "\"index\":$idx,\"text\":\"\K[^\"]*(?=\")" "$annotations" | head -1 || true)
    if [[ "$in_code" -eq 0 && -n "$rewrite" ]]; then out+="$(printf '%b' "$rewrite")"; else out+="$buf"; fi

    local before after pct
    before=$(estimate_tokens "$original")
    after=$(estimate_tokens "$out")
    pct=$(pct_saved "$before" "$after")

    cp "$target" "$target.bak"
    printf '%s' "$out" > "$target"
    add_stat_record "layer2-llm" "$relative" "$before" "$after"

    echo "$relative [llm]"
    echo "  $before tokens -> $after tokens  ($pct% saved)"
}

# ---------- Report ----------

cmd_report() {
    if [[ ! -f "$STATS_PATH" ]]; then
        echo "No token stats yet. Run 'compress' or 'run' first."
        return
    fi
    echo "plaesy-trim report"
    echo "===================="

    local layers
    layers=$(grep -oP '"layer":"\K[^"]+' "$STATS_PATH" | sort -u)
    for layer in $layers; do
        local count=0 total_before=0 total_after=0
        while IFS= read -r rec; do
            [[ "$rec" == *"\"layer\":\"$layer\""* ]] || continue
            local b a
            b=$(printf '%s' "$rec" | grep -oP '"before":\K[0-9]+')
            a=$(printf '%s' "$rec" | grep -oP '"after":\K[0-9]+')
            [[ -z "$b" || -z "$a" ]] && continue
            count=$((count + 1))
            total_before=$((total_before + b))
            total_after=$((total_after + a))
        done < <(grep -oP '\{[^}]*\}' "$STATS_PATH")
        [[ "$count" -eq 0 ]] && continue
        local pct
        pct=$(pct_saved "$total_before" "$total_after")
        echo "$layer: $count runs, $total_before -> $total_after tokens ($pct% avg saved)"
    done

    echo "--------------------"
    local grand_before=0 grand_after=0
    while IFS= read -r rec; do
        local b a
        b=$(printf '%s' "$rec" | grep -oP '"before":\K[0-9]+')
        a=$(printf '%s' "$rec" | grep -oP '"after":\K[0-9]+')
        [[ -z "$b" || -z "$a" ]] && continue
        grand_before=$((grand_before + b))
        grand_after=$((grand_after + a))
    done < <(grep -oP '\{[^}]*\}' "$STATS_PATH")
    local grand_pct
    grand_pct=$(pct_saved "$grand_before" "$grand_after")
    echo "total: $grand_before -> $grand_after tokens ($grand_pct% saved)"
}

# ---------- Dispatch ----------

main() {
    local command="${1:-}"
    shift || true
    case "$command" in
        run) cmd_run "$@" ;;
        compress) cmd_compress "$@" ;;
        llm-queue) cmd_llm_queue "$@" ;;
        apply-llm) cmd_apply_llm "$@" ;;
        report) cmd_report "$@" ;;
        *)
            echo "Usage: plaesy-trim.sh <run|compress|llm-queue|apply-llm|report> [options]"
            echo "  run <command...>                                                                Layer 1: per-tool aware compress of command output"
            echo "  compress --path <file|dir> [--level lite|full|ultra] [--recurse] [--dry-run]     Layer 2: fast heuristic prose compression"
            echo "  llm-queue --path <file>                                                          Layer 2 (LLM mode): export prose segments to rewrite"
            echo "  apply-llm --path <file> --annotations <json>                                     Layer 2 (LLM mode): merge rewritten segments back"
            echo "  report                                                                           Cumulative savings across layers"
            ;;
    esac
}

main "$@"
