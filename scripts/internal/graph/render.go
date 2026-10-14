package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"
)

// SaveJSON writes the graph as project.graph.json (indented, matching the
// bash script's field-by-field pretty-printer closely enough to be a drop-in
// replacement; exact byte-for-byte whitespace is not preserved).
func SaveJSON(g *Graph, path string) error {
	for i := range g.Nodes {
		if g.Nodes[i].Symbols == nil {
			g.Nodes[i].Symbols = []string{}
		}
	}
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// LoadJSON reads project.graph.json back into a Graph, for the standalone
// query commands (--query, --explain, --path-query, --impact-check, ...).
func LoadJSON(path string) (*Graph, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g Graph
	if err := json.Unmarshal(b, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// WriteReport mirrors write_report(): a plain-language reports.md summary.
// rationale, when non-empty, is appended verbatim as an "Explained
// relationships" section (mirrors --apply-semantic's rationale file).
func WriteReport(g *Graph, outFull, rationale string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Plaesy Graph Report\n\n")
	fmt.Fprintf(&b, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "## Overview\n- Nodes: %d\n- Edges: %d\n\n", len(g.Nodes), len(g.Edges))

	fmt.Fprintf(&b, "## Nodes by type\n")
	for _, c := range countDesc(mapValues(g.Nodes, func(n Node) string { return n.Type })) {
		fmt.Fprintf(&b, "- %s: %d\n", c.key, c.count)
	}
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "## Communities (top-level folders)\n")
	groupCounts := countDesc(mapValues(g.Nodes, func(n Node) string { return n.Group }))
	for _, c := range groupCounts {
		fmt.Fprintf(&b, "- %s: %d files\n", c.key, c.count)
	}
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "## Detected communities (label-propagation clustering)\n")
	fmt.Fprintf(&b, "These are computed from actual edge structure, not folder names - they surface cross-folder subsystems.\n")
	writeDetectedCommunities(&b, g)
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "## God nodes (most connected)\n")
	byDegreeDesc := append([]Node(nil), g.Nodes...)
	sort.SliceStable(byDegreeDesc, func(i, j int) bool { return byDegreeDesc[i].Degree > byDegreeDesc[j].Degree })
	for i := 0; i < len(byDegreeDesc) && i < 10; i++ {
		fmt.Fprintf(&b, "- `%s` - degree %d\n", byDegreeDesc[i].ID, byDegreeDesc[i].Degree)
	}
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "## Orphan files (no detected references)\n")
	var orphans []string
	for _, n := range g.Nodes {
		if n.Degree == 0 {
			orphans = append(orphans, n.ID)
		}
	}
	if len(orphans) == 0 {
		fmt.Fprintf(&b, "- None\n")
	} else {
		limit := len(orphans)
		if limit > 30 {
			limit = 30
		}
		for i := 0; i < limit; i++ {
			fmt.Fprintf(&b, "- `%s`\n", orphans[i])
		}
		if len(orphans) > 30 {
			fmt.Fprintf(&b, "- ... and %d more\n", len(orphans)-30)
		}
	}
	fmt.Fprintln(&b)

	if rationale != "" {
		fmt.Fprintf(&b, "## Explained relationships (semantic pass)\n%s\n\n", rationale)
	}

	fmt.Fprintf(&b, "## Suggested questions\n")
	fmt.Fprintf(&b, "- Which files are referenced by the most other files?\n")
	fmt.Fprintf(&b, "- Which files have no callers (potentially dead)?\n")
	if len(groupCounts) >= 2 {
		fmt.Fprintf(&b, "- What connects `%s` to `%s`?\n", groupCounts[0].key, groupCounts[1].key)
	}

	return os.WriteFile(outFull+string(os.PathSeparator)+"reports.md", []byte(b.String()), 0o644)
}

type countEntry struct {
	key   string
	count int
}

func mapValues(nodes []Node, f func(Node) string) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = f(n)
	}
	return out
}

// countDesc counts occurrences and sorts by count desc, then key asc
// (deterministic tie-break; the bash version's `sort -rn` on count alone
// leaves ties in whatever order `uniq -c` produced them).
func countDesc(values []string) []countEntry {
	counts := make(map[string]int)
	for _, v := range values {
		counts[v]++
	}
	out := make([]countEntry, 0, len(counts))
	for k, c := range counts {
		out = append(out, countEntry{k, c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].key < out[j].key
	})
	return out
}

// writeDetectedCommunities mirrors the report's per-community member listing:
// only communities with >1 member, up to 5 members each in degree-desc order.
func writeDetectedCommunities(b *strings.Builder, g *Graph) {
	type member struct {
		id     string
		degree int
	}
	byCommunity := make(map[int][]member)
	var order []int
	seen := make(map[int]bool)
	for _, n := range g.Nodes {
		byCommunity[n.Community] = append(byCommunity[n.Community], member{n.ID, n.Degree})
		if !seen[n.Community] {
			seen[n.Community] = true
			order = append(order, n.Community)
		}
	}
	sort.Ints(order)
	for _, cid := range order {
		members := byCommunity[cid]
		if len(members) < 2 {
			continue
		}
		sort.SliceStable(members, func(i, j int) bool { return members[i].degree > members[j].degree })
		limit := len(members)
		if limit > 5 {
			limit = 5
		}
		names := make([]string, limit)
		for i := 0; i < limit; i++ {
			names[i] = members[i].id
		}
		fmt.Fprintf(b, "- Community %d (%d files): %s\n", cid, len(members), strings.Join(names, ", "))
	}
}

// WriteHTML mirrors write_html(): a self-contained force-directed viz, same
// canvas/physics JS as the bash version, with the graph JSON embedded.
func WriteHTML(g *Graph, outFull string) error {
	jb, err := json.Marshal(g)
	if err != nil {
		return err
	}
	tmpl := template.Must(template.New("html").Parse(htmlTemplate))
	f, err := os.Create(outFull + string(os.PathSeparator) + "project.html")
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, struct{ JSON string }{string(jb)})
}

const htmlTemplate = `<!DOCTYPE html>
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
const data = {{.JSON}};
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
`
