package graph

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// index bundles the lookup structures every query command needs, built once
// from a loaded Graph (mirrors require_graph + the awk-based neighbors_of /
// find_node helpers, but in memory instead of re-scanning NODES_TSV/EDGES_TSV
// per lookup).
type index struct {
	g        *Graph
	byID     map[string]*Node
	order    []string // node ids in file order (as stored in the graph)
	adjOut   map[string][]Edge
	adjIn    map[string][]Edge
	neighbor map[string][]string // union of in+out, sorted unique
}

func newIndex(g *Graph) *index {
	ix := &index{g: g, byID: map[string]*Node{}, adjOut: map[string][]Edge{}, adjIn: map[string][]Edge{}, neighbor: map[string][]string{}}
	for i := range g.Nodes {
		n := &g.Nodes[i]
		ix.byID[n.ID] = n
		ix.order = append(ix.order, n.ID)
	}
	nbSet := map[string]map[string]bool{}
	for _, e := range g.Edges {
		ix.adjOut[e.Source] = append(ix.adjOut[e.Source], e)
		ix.adjIn[e.Target] = append(ix.adjIn[e.Target], e)
		if nbSet[e.Source] == nil {
			nbSet[e.Source] = map[string]bool{}
		}
		if nbSet[e.Target] == nil {
			nbSet[e.Target] = map[string]bool{}
		}
		nbSet[e.Source][e.Target] = true
		nbSet[e.Target][e.Source] = true
	}
	for id, set := range nbSet {
		var list []string
		for nb := range set {
			list = append(list, nb)
		}
		sort.Strings(list)
		ix.neighbor[id] = list
	}
	return ix
}

// findNode mirrors find_node(): first node whose id contains needle
// (substring, case-sensitive - matches the bash `grep -F`), in file order.
func (ix *index) findNode(needle string) string {
	for _, id := range ix.order {
		if strings.Contains(id, needle) {
			return id
		}
	}
	return ""
}

// DoQuery mirrors do_query(): case-insensitive substring match on node id,
// printing each match's type/degree and neighbors.
func DoQuery(g *Graph, query string, w io.Writer) {
	ix := newIndex(g)
	ql := strings.ToLower(query)
	var matches []string
	for _, id := range ix.order {
		if strings.Contains(strings.ToLower(id), ql) {
			matches = append(matches, id)
		}
	}
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "[plaesy-graph] No nodes matched '%s'.\n", query)
		return
	}
	for _, m := range matches {
		n := ix.byID[m]
		fmt.Fprintf(w, "\n== %s (%s, degree %d) ==\n", m, n.Type, n.Degree)
		nbs := ix.neighbor[m]
		if len(nbs) == 0 {
			fmt.Fprintln(w, "  (no connections found)")
			continue
		}
		for _, nb := range nbs {
			fmt.Fprintf(w, "  -> %s\n", nb)
		}
	}
}

// DoExplain mirrors do_explain().
func DoExplain(g *Graph, name string, w io.Writer) {
	ix := newIndex(g)
	id := ix.findNode(name)
	if id == "" {
		fmt.Fprintf(os.Stderr, "[plaesy-graph] No node matching '%s'.\n", name)
		return
	}
	n := ix.byID[id]
	fmt.Fprintf(w, "\n%s\n", id)
	fmt.Fprintf(w, "  type: %s   group: %s   degree: %d\n", n.Type, n.Group, n.Degree)
	if len(n.Symbols) > 0 {
		fmt.Fprintf(w, "  symbols (%d): %s\n", len(n.Symbols), strings.Join(n.Symbols, ", "))
	}
	fmt.Fprintln(w, "  outgoing:")
	out := ix.adjOut[id]
	if len(out) == 0 {
		fmt.Fprintln(w, "    (none)")
	} else {
		for _, e := range out {
			fmt.Fprintf(w, "    -[%s]-> %s\n", e.Type, e.Target)
		}
	}
	fmt.Fprintln(w, "  incoming:")
	in := ix.adjIn[id]
	if len(in) == 0 {
		fmt.Fprintln(w, "    (none)")
	} else {
		for _, e := range in {
			fmt.Fprintf(w, "    <-[%s]- %s\n", e.Type, e.Source)
		}
	}
}

// DoPathQuery mirrors do_path_query(): BFS shortest path between two
// (substring-resolved) nodes.
func DoPathQuery(g *Graph, from, to string, w io.Writer) {
	ix := newIndex(g)
	f, t := ix.findNode(from), ix.findNode(to)
	if f == "" || t == "" {
		fmt.Fprintln(os.Stderr, "[plaesy-graph] Could not resolve one or both node names.")
		return
	}
	prev := map[string]string{}
	visited := map[string]bool{f: true}
	queue := []string{f}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == t {
			break
		}
		for _, nb := range ix.neighbor[cur] {
			if !visited[nb] {
				visited[nb] = true
				prev[nb] = cur
				queue = append(queue, nb)
			}
		}
	}
	if !visited[t] {
		fmt.Fprintf(os.Stderr, "[plaesy-graph] No path found between %s and %s.\n", f, t)
		return
	}
	var path []string
	for cur := t; cur != f; cur = prev[cur] {
		path = append([]string{cur}, path...)
	}
	path = append([]string{f}, path...)
	fmt.Fprintf(w, "[plaesy-graph] Path (%d nodes, %d hops):\n", len(path), len(path)-1)
	fmt.Fprintln(w, strings.Join(path, "  ->  "))
}

// DoImpactCheck mirrors do_impact_check(): BFS levels up to depth.
func DoImpactCheck(g *Graph, name string, depth int, w io.Writer) {
	ix := newIndex(g)
	id := ix.findNode(name)
	if id == "" {
		fmt.Fprintf(os.Stderr, "[plaesy-graph] No node matching '%s'.\n", name)
		return
	}
	depthOf := map[string]int{id: 0}
	queue := []string{id}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		d := depthOf[cur]
		if d >= depth {
			continue
		}
		for _, nb := range ix.neighbor[cur] {
			if _, ok := depthOf[nb]; !ok {
				depthOf[nb] = d + 1
				queue = append(queue, nb)
			}
		}
	}
	delete(depthOf, id)
	fmt.Fprintf(w, "\n[plaesy-graph] Impact of changing %s (depth %d):\n", id, depth)
	if len(depthOf) == 0 {
		fmt.Fprintln(w, "  No connected files found - safe to change in isolation (per detected edges).")
		return
	}
	for d := 1; d <= depth; d++ {
		var atDepth []string
		for k, dd := range depthOf {
			if dd == d {
				atDepth = append(atDepth, k)
			}
		}
		if len(atDepth) == 0 {
			continue
		}
		sort.Strings(atDepth)
		fmt.Fprintf(w, "  Depth %d:\n", d)
		for _, k := range atDepth {
			fmt.Fprintf(w, "    - %s  [%s]\n", k, ix.byID[k].Type)
		}
	}
}

// DoFuzzyQuery mirrors do_fuzzy_query(): case-insensitive substring match,
// top 10, printing type/layer/degree.
func DoFuzzyQuery(g *Graph, query string, w io.Writer) {
	ql := strings.ToLower(query)
	count := 0
	for _, n := range g.Nodes {
		if !strings.Contains(strings.ToLower(n.ID), ql) {
			continue
		}
		layer := n.Layer
		if layer == "" {
			layer = "n/a"
		}
		fmt.Fprintf(w, "  - %s (type: %s, layer: %s, connections: %d)\n", n.ID, n.Type, layer, n.Degree)
		count++
		if count >= 10 {
			break
		}
	}
}

// DoSemanticQueue mirrors do_semantic_queue(): dumps every INFERRED edge to
// semantic-queue.json.
func DoSemanticQueue(g *Graph, outFull string) (int, error) {
	type item struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Type   string `json:"type"`
	}
	var items []item
	for _, e := range g.Edges {
		if e.Confidence == "INFERRED" {
			items = append(items, item{e.Source, e.Target, e.Type})
		}
	}
	if len(items) == 0 {
		return 0, nil
	}
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return 0, err
	}
	path := outFull + string(os.PathSeparator) + "semantic-queue.json"
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		return 0, err
	}
	return len(items), nil
}

// DoBusinessLogic mirrors do_business_logic(): every INFERRED edge, as a
// {source,target,question} prompt for a human/AI to answer.
func DoBusinessLogic(g *Graph, outFull string) (int, error) {
	type item struct {
		Source   string `json:"source"`
		Target   string `json:"target"`
		Question string `json:"question"`
	}
	var items []item
	for _, e := range g.Edges {
		if e.Confidence == "INFERRED" {
			items = append(items, item{e.Source, e.Target, fmt.Sprintf("Business logic connecting %s to %s?", e.Source, e.Target)})
		}
	}
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return 0, err
	}
	if items == nil {
		b = []byte("[]")
	}
	path := outFull + string(os.PathSeparator) + "business-logic-queue.json"
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		return 0, err
	}
	return len(items), nil
}

// DoGeneratePaths mirrors do_generate_paths(): a stub learning-paths.json
// with the first 5 nodes (in graph/id order) as "entry points".
func DoGeneratePaths(g *Graph, outFull string) error {
	entry := make([]string, 0, 5)
	for i := 0; i < len(g.Nodes) && i < 5; i++ {
		entry = append(entry, g.Nodes[i].ID)
	}
	doc := struct {
		GeneratedAt string   `json:"generated_at"`
		TotalNodes  int      `json:"total_nodes"`
		Strategy    string   `json:"strategy"`
		EntryPoints []string `json:"entry_points"`
	}{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		TotalNodes:  len(g.Nodes),
		Strategy:    "Analyze dependencies to suggest learning order",
		EntryPoints: entry,
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	path := outFull + string(os.PathSeparator) + "learning-paths.json"
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// DoImpactVisualize mirrors do_impact_visualize(): a static placeholder HTML
// page pointing back at --impact-check for the real report (matches the
// bash version, which is itself a stub - DEPTH_PLACEHOLDER is never filled
// in there either).
func DoImpactVisualize(node, outFull string) error {
	tmpl := `<!DOCTYPE html>
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
    <p>Depth analysis configured for: ` + node + `</p>
    <p>Run: <code>plaesy graph --impact-check NODE --impact-depth N</code> for detailed report</p>
  </div>
</body>
</html>
`
	path := outFull + string(os.PathSeparator) + "impact-visualization.html"
	return os.WriteFile(path, []byte(tmpl), 0o644)
}

var (
	semSrcRE = regexp.MustCompile(`"source"\s*:\s*"([^"]+)"`)
	semTgtRE = regexp.MustCompile(`"target"\s*:\s*"([^"]+)"`)
	semRatRE = regexp.MustCompile(`"rationale"\s*:\s*"([^"]+)"`)
)

// DoApplySemantic mirrors do_apply_semantic(): a naive line-based scrape of
// {"source","target","rationale"} objects from an annotations file (same
// non-JSON-parser approach the bash script uses), matched against existing
// edges, then folded into an "Explained relationships" section of
// reports.md.
func DoApplySemantic(g *Graph, annotationsPath, outFull string) (int, error) {
	b, err := os.ReadFile(annotationsPath)
	if err != nil {
		return 0, err
	}
	content := string(b)
	srcs := semSrcRE.FindAllStringSubmatch(content, -1)
	tgts := semTgtRE.FindAllStringSubmatch(content, -1)
	rats := semRatRE.FindAllStringSubmatch(content, -1)

	existing := map[[2]string]bool{}
	for _, e := range g.Edges {
		existing[[2]string{e.Source, e.Target}] = true
	}

	var rationale strings.Builder
	applied := 0
	n := len(srcs)
	if len(tgts) < n {
		n = len(tgts)
	}
	if len(rats) < n {
		n = len(rats)
	}
	for i := 0; i < n; i++ {
		s, t, r := srcs[i][1], tgts[i][1], rats[i][1]
		if s == "" {
			continue
		}
		if existing[[2]string{s, t}] {
			fmt.Fprintf(&rationale, "- `%s` -> `%s`: %s\n", s, t, r)
			applied++
		}
	}
	if err := WriteReport(g, outFull, rationale.String()); err != nil {
		return applied, err
	}
	if err := SaveJSON(g, outFull+string(os.PathSeparator)+"project.graph.json"); err != nil {
		return applied, err
	}
	return applied, WriteHTML(g, outFull)
}
