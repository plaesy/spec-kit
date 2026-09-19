# Symbol extraction for plaesy-graph.sh. BATCHED: accepts ALL files as
# arguments in a single awk process (each spawn costs ~200ms on MSYS/Windows,
# so one pass over every file instead of one per file). The extension is
# derived from FILENAME so no `-v ext=` is needed. Emits one symbol per line
# as `file\tsymbol`; plaesy-graph.sh regroups them per node.
# Mirrors plaesy-graph.ps1's $SymbolPatterns table so both scripts surface the
# same "what's actually inside this file" info.
{
    line = $0
    if (ext == ".ps1") {
        if (match(line, /^[ \t]*function[ \t]+[A-Za-z0-9_-]+/)) {
            s = substr(line, RSTART, RLENGTH); sub(/^[ \t]*function[ \t]+/, "", s); print FILENAME "\t" s
        }
    } else if (ext == ".sh") {
        if (match(line, /^[ \t]*function[ \t]+[A-Za-z0-9_]+/)) {
            s = substr(line, RSTART, RLENGTH); sub(/^[ \t]*function[ \t]+/, "", s); print FILENAME "\t" s
        } else if (match(line, /^[ \t]*[A-Za-z0-9_]+[ \t]*\(\)[ \t]*\{/)) {
            s = substr(line, RSTART, RLENGTH); sub(/^[ \t]*/, "", s); sub(/[ \t]*\(\).*/, "", s); print FILENAME "\t" s
        }
    } else if (ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx") {
        if (match(line, /^[ \t]*(export[ \t]+)?function[ \t]+[A-Za-z0-9_$]+[ \t]*\(/)) {
            s = substr(line, RSTART, RLENGTH); sub(/\($/, "", s); sub(/.*function[ \t]+/, "", s); print FILENAME "\t" s
        } else if (match(line, /^[ \t]*(export[ \t]+)?class[ \t]+[A-Za-z0-9_$]+/)) {
            s = substr(line, RSTART, RLENGTH); sub(/.*class[ \t]+/, "", s); print FILENAME "\t" s
        }
    } else if (ext == ".py") {
        if (match(line, /^[ \t]*def[ \t]+[A-Za-z0-9_]+[ \t]*\(/)) {
            s = substr(line, RSTART, RLENGTH); sub(/\($/, "", s); sub(/.*def[ \t]+/, "", s); print FILENAME "\t" s
        } else if (match(line, /^[ \t]*class[ \t]+[A-Za-z0-9_]+/)) {
            s = substr(line, RSTART, RLENGTH); sub(/.*class[ \t]+/, "", s); print FILENAME "\t" s
        }
    } else if (ext == ".go") {
        if (match(line, /^func[ \t]+(\([^)]*\)[ \t]*)?[A-Za-z0-9_]+[ \t]*\(/)) {
            s = substr(line, RSTART, RLENGTH); sub(/\($/, "", s)
            sub(/^func[ \t]+/, "", s); sub(/^\([^)]*\)[ \t]*/, "", s)
            print FILENAME "\t" s
        }
    } else if (ext == ".md") {
        if (match(line, /^#{1,3}[ \t]+.+$/)) {
            s = substr(line, RSTART, RLENGTH); sub(/^#{1,3}[ \t]+/, "", s); sub(/[ \t]+$/, "", s); print FILENAME "\t" s
        }
    }
}
FNR == 1 {
    ext = ""
    if (match(FILENAME, /\.[^.]+$/)) ext = substr(FILENAME, RSTART, RLENGTH)
}
