package graph

import "strings"

// mentionEdges is a Go port of plaesy-graph-mentions.awk.
//
// What the awk did: it read a candidate-path file (one path per line, the
// full node-id list) plus a backslash-form of each candidate, then for every
// input file it read the ENTIRE file content into one string and, for each
// candidate independently, tested `index(content, candidate) > 0 ||
// index(content, candidate_with_backslashes) > 0`. Any candidate found by
// either test was emitted as `file<TAB>candidate`. It deliberately avoided
// `grep -oFf` (multi-pattern fixed-string grep) because GNU grep -F only
// reports the first non-overlapping match per text position among a set of
// patterns - so a candidate that is a substring of another candidate's match
// span (e.g. "install.ps1" inside "scripts/powershell/install.ps1") could be
// silently dropped. Testing each candidate with index() independently gives
// the same exhaustive recall as the original per-candidate loop, without
// spawning a process per file or per candidate.
//
// This function reproduces that exactly: for every file, for every other
// known node id, check whether the file's content contains the id in its
// forward-slash form OR its backslash form. A hit becomes an INFERRED
// "mentions" edge file -> candidate (never file -> itself).
func mentionEdges(repoRootSlash string, files []string, allNodeIDs []string) []Edge {
	type cand struct{ fwd, bs string }
	cands := make([]cand, len(allNodeIDs))
	for i, id := range allNodeIDs {
		cands[i] = cand{fwd: id, bs: strings.ReplaceAll(id, "/", `\`)}
	}

	var edges []Edge
	for _, rel := range files {
		content := readAll(joinRoot(repoRootSlash, rel))
		if content == "" {
			continue
		}
		for _, c := range cands {
			if c.fwd == rel {
				continue
			}
			if strings.Contains(content, c.fwd) || strings.Contains(content, c.bs) {
				edges = append(edges, Edge{Source: rel, Target: c.fwd, Type: "mentions", Confidence: "INFERRED"})
			}
		}
	}
	return edges
}

// dropWeakMentions mirrors the awk pass in build_graph step 3: drop a
// "mentions" edge whenever a non-"mentions" edge already exists for the
// exact same (source,target) pair, then dedupe (the bash pipes through
// `sort -u`).
func dropWeakMentions(edges []Edge) []Edge {
	strong := make(map[[2]string]bool)
	for _, e := range edges {
		if e.Type != "mentions" {
			strong[[2]string{e.Source, e.Target}] = true
		}
	}
	seen := make(map[Edge]bool)
	var out []Edge
	for _, e := range edges {
		if e.Type == "mentions" && strong[[2]string{e.Source, e.Target}] {
			continue
		}
		if seen[e] {
			continue
		}
		seen[e] = true
		out = append(out, e)
	}
	return out
}
