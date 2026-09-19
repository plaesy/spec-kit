# Exhaustive plain-text mention detection for plaesy-graph.sh, mirroring
# plaesy-graph.ps1's per-candidate check (tests both forward-slash and
# backslash forms of every candidate path). Deliberately NOT implemented as
# `grep -oFf candidates content` - GNU grep -F with many overlapping/
# substring patterns (e.g. "install.ps1" is a substring of
# "scripts/powershell/install.ps1") only reports the first non-overlapping
# match per text position and silently drops the others, which under-detects
# relative to the per-candidate PS1 loop. Reading the whole file once and
# testing each candidate independently via index() gives identical recall
# without spawning one grep process per candidate (which was the original,
# too-slow approach).
#
# BATCHED: accepts ALL files as arguments in a single invocation (one awk
# process total instead of one per file - each spawn costs ~200ms on
# MSYS/Windows). Emits one matched candidate per line as `file\tcandidate`
# using the file path exactly as passed on the command line.
#
# Usage: awk -v candfile=path/to/allnodes.txt -f plaesy-graph-mentions.awk file1 file2 ...
BEGIN {
    while ((getline line < candfile) > 0) {
        n++
        cand[n] = line
        bs = line
        gsub(/\//, "\\", bs)
        candbs[n] = bs
    }
    close(candfile)
}
FNR == 1 {
    if (NR > 1) emit(curfile)
    curfile = FILENAME
    content = ""
}
{ content = content $0 "\n" }
END {
    if (n > 0) emit(curfile)
}
function emit(file) {
    for (i = 1; i <= n; i++) {
        if (index(content, cand[i]) > 0 || index(content, candbs[i]) > 0) print file "\t" cand[i]
    }
}
