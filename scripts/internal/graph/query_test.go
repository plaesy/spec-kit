package graph

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// sampleGraph is the graph the query tests share: two connected pairs, one
// INFERRED mention, and one orphan. Degrees are the ones Build would compute
// from these edges (each edge counts once per endpoint).
func sampleGraph() *Graph {
	return &Graph{
		GeneratedAt: "2026-01-02 03:04:05",
		Root:        "/repo",
		Description: "sample",
		Nodes: []Node{
			{ID: "README.md", Label: "README.md", Type: "other", Group: "README.md", Degree: 2, Community: 0, Symbols: []string{"Title", "Intro"}},
			{ID: "docs/spec.md", Label: "spec.md", Type: "doc", Group: "docs", Degree: 1, Community: 0, Layer: "API"},
			{ID: "app/main.js", Label: "main.js", Type: "source", Group: "app", Degree: 2, Community: 1},
			{ID: "app/util.js", Label: "util.js", Type: "source", Group: "app", Degree: 1, Community: 1},
			{ID: "lonely.md", Label: "lonely.md", Type: "other", Group: "lonely.md", Degree: 0, Community: 2},
		},
		Edges: []Edge{
			{Source: "README.md", Target: "docs/spec.md", Type: "references", Confidence: "EXTRACTED"},
			{Source: "app/main.js", Target: "app/util.js", Type: "imports", Confidence: "EXTRACTED"},
			{Source: "README.md", Target: "app/main.js", Type: "mentions", Confidence: "INFERRED"},
		},
	}
}

func newIndexOf(g *Graph) *index { return newIndex(g) }

func TestNewIndex(t *testing.T) {
	ix := newIndexOf(sampleGraph())

	t.Run("node ids keep their stored order", func(t *testing.T) {
		want := []string{"README.md", "docs/spec.md", "app/main.js", "app/util.js", "lonely.md"}
		if !reflect.DeepEqual(ix.order, want) {
			t.Fatalf("order = %v, want %v", ix.order, want)
		}
	})

	t.Run("byID points at the graph's own nodes", func(t *testing.T) {
		// The pointers must alias g.Nodes, not copies: every query reads
		// Type/Degree/Symbols through them.
		if ix.byID["docs/spec.md"] != &ix.g.Nodes[1] {
			t.Fatalf("byID[docs/spec.md] = %p, want &g.Nodes[1] = %p", ix.byID["docs/spec.md"], &ix.g.Nodes[1])
		}
	})

	t.Run("directed adjacency", func(t *testing.T) {
		if len(ix.adjOut["README.md"]) != 2 {
			t.Errorf("adjOut[README.md] = %v, want 2 edges", ix.adjOut["README.md"])
		}
		if len(ix.adjIn["docs/spec.md"]) != 1 {
			t.Errorf("adjIn[docs/spec.md] = %v, want 1 edge", ix.adjIn["docs/spec.md"])
		}
		if len(ix.adjOut["docs/spec.md"]) != 0 {
			t.Errorf("adjOut[docs/spec.md] = %v, want none", ix.adjOut["docs/spec.md"])
		}
		if len(ix.adjIn["lonely.md"]) != 0 {
			t.Errorf("adjIn[lonely.md] = %v, want none", ix.adjIn["lonely.md"])
		}
	})

	// The neighbor list is the union of both directions, so a query can walk
	// either way; it must be sorted and free of duplicates even when several
	// edges join the same pair.
	t.Run("neighbors are the sorted unique union of both directions", func(t *testing.T) {
		g := &Graph{
			Nodes: []Node{{ID: "a"}, {ID: "b"}, {ID: "c"}},
			Edges: []Edge{
				{Source: "a", Target: "b"},
				{Source: "a", Target: "b"},
				{Source: "c", Target: "a"},
			},
		}
		ix := newIndexOf(g)
		if !reflect.DeepEqual(ix.neighbor["a"], []string{"b", "c"}) {
			t.Errorf("neighbor[a] = %v, want [b c]", ix.neighbor["a"])
		}
		if !reflect.DeepEqual(ix.neighbor["b"], []string{"a"}) {
			t.Errorf("neighbor[b] = %v, want [a]", ix.neighbor["b"])
		}
		if _, ok := ix.neighbor["nonexistent"]; ok {
			t.Error("an id with no edges must have no neighbor entry")
		}
	})
}

func TestFindNode(t *testing.T) {
	ix := newIndexOf(sampleGraph())
	tests := []struct{ name, needle, want string }{
		{name: "exact id", needle: "docs/spec.md", want: "docs/spec.md"},
		{name: "substring", needle: "spec", want: "docs/spec.md"},
		{name: "first in file order wins", needle: ".md", want: "README.md"},
		{name: "case sensitive, like the bash grep -F", needle: "SPEC", want: ""},
		{name: "no match", needle: "zzz", want: ""},
		{name: "empty needle would match everything", needle: "", want: "README.md"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ix.findNode(tc.needle); got != tc.want {
				t.Fatalf("findNode(%q) = %q, want %q", tc.needle, got, tc.want)
			}
		})
	}
}

func TestDoQuery(t *testing.T) {
	g := sampleGraph()

	t.Run("every case-insensitive match, with type, degree and neighbors", func(t *testing.T) {
		var buf bytes.Buffer
		DoQuery(g, "APP", &buf)
		out := buf.String()
		for _, want := range []string{
			"== app/main.js (source, degree 2) ==",
			"== app/util.js (source, degree 1) ==",
			"  -> README.md\n",
			"  -> app/util.js\n",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
		}
		if strings.Contains(out, "docs/spec.md") {
			t.Errorf("a non-matching node was printed:\n%s", out)
		}
	})

	t.Run("a node with no connections says so", func(t *testing.T) {
		var buf bytes.Buffer
		DoQuery(g, "lonely", &buf)
		if !strings.Contains(buf.String(), "(no connections found)") {
			t.Errorf("output = %q, want the no-connections line", buf.String())
		}
	})

	// The "nothing matched" notice goes to stderr, so a caller piping stdout
	// gets no output at all rather than an empty block that reads as a result.
	t.Run("no match writes nothing to the report stream", func(t *testing.T) {
		var buf bytes.Buffer
		DoQuery(g, "nothing-here", &buf)
		if buf.Len() != 0 {
			t.Fatalf("output = %q, want empty", buf.String())
		}
	})
}

func TestDoExplain(t *testing.T) {
	g := sampleGraph()

	t.Run("full node detail", func(t *testing.T) {
		var buf bytes.Buffer
		DoExplain(g, "spec.md", &buf)
		out := buf.String()
		for _, want := range []string{
			"docs/spec.md",
			"type: doc   group: docs   degree: 1",
			"  outgoing:\n    (none)\n",
			"  incoming:\n    <-[references]- README.md\n",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
		}
	})

	t.Run("symbols are listed when present", func(t *testing.T) {
		var buf bytes.Buffer
		DoExplain(g, "README", &buf)
		if !strings.Contains(buf.String(), "symbols (2): Title, Intro") {
			t.Errorf("output = %q, want the symbol list", buf.String())
		}
	})

	t.Run("a node with no outgoing edge", func(t *testing.T) {
		var buf bytes.Buffer
		DoExplain(g, "util.js", &buf)
		out := buf.String()
		if !strings.Contains(out, "  outgoing:\n    (none)\n") {
			t.Errorf("output = %q, want an empty outgoing section", out)
		}
		if !strings.Contains(out, "  incoming:\n    <-[imports]- app/main.js\n") {
			t.Errorf("output = %q, want the incoming edge", out)
		}
	})

	// An unknown name has to say so on stderr and print nothing, so a script
	// piping the output cannot mistake an empty report for a node with no
	// edges.
	t.Run("an unknown name writes nothing to the report stream", func(t *testing.T) {
		var buf bytes.Buffer
		DoExplain(g, "zzz", &buf)
		if buf.Len() != 0 {
			t.Fatalf("output = %q, want empty", buf.String())
		}
	})
}

func TestDoPathQuery(t *testing.T) {
	g := sampleGraph()

	t.Run("a path across the mention edge", func(t *testing.T) {
		var buf bytes.Buffer
		DoPathQuery(g, "docs/spec", "util.js", &buf)
		// spec -> README -> main -> util: the search is undirected, so the
		// README -> main "mentions" edge carries the path too.
		if !strings.Contains(buf.String(), "Path (4 nodes, 3 hops):\ndocs/spec.md  ->  README.md  ->  app/main.js  ->  app/util.js\n") {
			t.Errorf("output = %q, want the 4-node path", buf.String())
		}
	})

	t.Run("a node to itself is a one-node path", func(t *testing.T) {
		var buf bytes.Buffer
		DoPathQuery(g, "README", "README", &buf)
		if !strings.Contains(buf.String(), "Path (1 nodes, 0 hops):\nREADME.md\n") {
			t.Errorf("output = %q, want a 1-node path", buf.String())
		}
	})

	t.Run("an unresolvable name produces no output", func(t *testing.T) {
		var buf bytes.Buffer
		DoPathQuery(g, "docs/spec", "zzz", &buf)
		if buf.Len() != 0 {
			t.Fatalf("output = %q, want empty", buf.String())
		}
	})

	t.Run("a disconnected node produces no output", func(t *testing.T) {
		var buf bytes.Buffer
		DoPathQuery(g, "README", "lonely", &buf)
		if buf.Len() != 0 {
			t.Fatalf("output = %q, want empty (the no-path notice goes to stderr)", buf.String())
		}
	})
}

func TestDoImpactCheck(t *testing.T) {
	g := sampleGraph()

	t.Run("depth 1 lists the direct neighbors", func(t *testing.T) {
		var buf bytes.Buffer
		DoImpactCheck(g, "README", 1, &buf)
		out := buf.String()
		if !strings.Contains(out, "Impact of changing README.md (depth 1)") {
			t.Errorf("output = %q, want the header", out)
		}
		if !strings.Contains(out, "  Depth 1:\n    - app/main.js  [source]\n    - docs/spec.md  [doc]\n") {
			t.Errorf("output = %q, want both depth-1 nodes sorted", out)
		}
		if strings.Contains(out, "Depth 2") {
			t.Errorf("output = %q, want no depth-2 section", out)
		}
	})

	t.Run("depth 2 reaches the far end", func(t *testing.T) {
		var buf bytes.Buffer
		DoImpactCheck(g, "README", 2, &buf)
		out := buf.String()
		if !strings.Contains(out, "  Depth 2:\n    - app/util.js  [source]\n") {
			t.Errorf("output = %q, want util.js at depth 2", out)
		}
	})

	t.Run("a depth deeper than the graph skips the empty levels", func(t *testing.T) {
		var buf bytes.Buffer
		DoImpactCheck(g, "README", 4, &buf)
		out := buf.String()
		if !strings.Contains(out, "Depth 1:") || !strings.Contains(out, "Depth 2:") {
			t.Errorf("output = %q, want the two levels that exist", out)
		}
		// Levels 3 and 4 have no members and are not printed at all.
		if strings.Contains(out, "Depth 3") || strings.Contains(out, "Depth 4") {
			t.Errorf("output = %q, want no empty level headings", out)
		}
	})

	t.Run("an isolated node is reported as safe", func(t *testing.T) {
		var buf bytes.Buffer
		DoImpactCheck(g, "lonely", 3, &buf)
		if !strings.Contains(buf.String(), "No connected files found - safe to change in isolation (per detected edges).") {
			t.Errorf("output = %q, want the safe-to-change line", buf.String())
		}
	})

	t.Run("a depth of zero explores nothing", func(t *testing.T) {
		var buf bytes.Buffer
		DoImpactCheck(g, "README", 0, &buf)
		if !strings.Contains(buf.String(), "No connected files found") {
			t.Errorf("output = %q, want the safe-to-change line", buf.String())
		}
	})

	t.Run("an unknown name produces no output", func(t *testing.T) {
		var buf bytes.Buffer
		DoImpactCheck(g, "zzz", 2, &buf)
		if buf.Len() != 0 {
			t.Fatalf("output = %q, want empty", buf.String())
		}
	})
}

func TestDoFuzzyQuery(t *testing.T) {
	t.Run("type, layer and degree, with n/a for an unset layer", func(t *testing.T) {
		var buf bytes.Buffer
		DoFuzzyQuery(sampleGraph(), "app", &buf)
		want := "  - app/main.js (type: source, layer: n/a, connections: 2)\n" +
			"  - app/util.js (type: source, layer: n/a, connections: 1)\n"
		if buf.String() != want {
			t.Fatalf("output = %q, want %q", buf.String(), want)
		}
	})

	t.Run("a node with a layer reports it", func(t *testing.T) {
		var buf bytes.Buffer
		DoFuzzyQuery(sampleGraph(), "spec", &buf)
		if !strings.Contains(buf.String(), "layer: API") {
			t.Fatalf("output = %q, want the API layer", buf.String())
		}
	})

	// The cap exists so a one-character query cannot flood the terminal; it is
	// a hard stop, not a "best ten".
	t.Run("at most ten results", func(t *testing.T) {
		g := &Graph{}
		for i := 0; i < 25; i++ {
			g.Nodes = append(g.Nodes, Node{ID: "match" + string(rune('a'+i)) + ".md", Type: "other", Degree: i})
		}
		var buf bytes.Buffer
		DoFuzzyQuery(g, "match", &buf)
		if n := strings.Count(buf.String(), "\n"); n != 10 {
			t.Fatalf("printed %d lines, want 10:\n%s", n, buf.String())
		}
	})

	t.Run("no match prints nothing", func(t *testing.T) {
		var buf bytes.Buffer
		DoFuzzyQuery(sampleGraph(), "zzz", &buf)
		if buf.Len() != 0 {
			t.Fatalf("output = %q, want empty", buf.String())
		}
	})
}

func readJSONFile(t *testing.T, path string) []map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	var out []map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("%s is not a JSON array: %v\n%s", path, err, b)
	}
	return out
}

func TestDoSemanticQueue(t *testing.T) {
	t.Run("only INFERRED edges are queued", func(t *testing.T) {
		out := t.TempDir()
		n, err := DoSemanticQueue(sampleGraph(), out)
		if err != nil {
			t.Fatalf("DoSemanticQueue() error = %v", err)
		}
		if n != 1 {
			t.Fatalf("count = %d, want 1", n)
		}
		items := readJSONFile(t, filepath.Join(out, "semantic-queue.json"))
		want := []map[string]any{{"source": "README.md", "target": "app/main.js", "type": "mentions"}}
		if !reflect.DeepEqual(items, want) {
			t.Fatalf("semantic-queue.json = %v, want %v", items, want)
		}
	})

	// No INFERRED edges means nothing to review, and the caller says so from
	// the count. Writing an empty file here would leave a queue artifact that
	// a later "is the queue empty?" check cannot tell from a real one.
	t.Run("no inferred edges writes no file", func(t *testing.T) {
		out := t.TempDir()
		g := sampleGraph()
		g.Edges = nil
		n, err := DoSemanticQueue(g, out)
		if err != nil {
			t.Fatalf("DoSemanticQueue() error = %v", err)
		}
		if n != 0 {
			t.Fatalf("count = %d, want 0", n)
		}
		if _, err := os.Stat(filepath.Join(out, "semantic-queue.json")); !os.IsNotExist(err) {
			t.Fatalf("semantic-queue.json exists (err=%v), want no file", err)
		}
	})

	t.Run("an unwritable output dir is an error, not a silent zero", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir")
		n, err := DoSemanticQueue(sampleGraph(), missing)
		if err == nil {
			t.Fatalf("DoSemanticQueue() = %d, nil; want a write error", n)
		}
	})
}

func TestDoBusinessLogic(t *testing.T) {
	t.Run("every inferred edge becomes a question", func(t *testing.T) {
		out := t.TempDir()
		n, err := DoBusinessLogic(sampleGraph(), out)
		if err != nil {
			t.Fatalf("DoBusinessLogic() error = %v", err)
		}
		if n != 1 {
			t.Fatalf("count = %d, want 1", n)
		}
		items := readJSONFile(t, filepath.Join(out, "business-logic-queue.json"))
		want := []map[string]any{{
			"source":   "README.md",
			"target":   "app/main.js",
			"question": "Business logic connecting README.md to app/main.js?",
		}}
		if !reflect.DeepEqual(items, want) {
			t.Fatalf("business-logic-queue.json = %v, want %v", items, want)
		}
	})

	// The counterpart to DoSemanticQueue, and deliberately different: the
	// caller always reports where the file went, so the file has to exist even
	// when there is nothing in it. Pinned because the two disagree.
	t.Run("no inferred edges writes an empty array", func(t *testing.T) {
		out := t.TempDir()
		g := sampleGraph()
		g.Edges = nil
		n, err := DoBusinessLogic(g, out)
		if err != nil {
			t.Fatalf("DoBusinessLogic() error = %v", err)
		}
		if n != 0 {
			t.Fatalf("count = %d, want 0", n)
		}
		b, err := os.ReadFile(filepath.Join(out, "business-logic-queue.json"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if strings.TrimSpace(string(b)) != "[]" {
			t.Fatalf("business-logic-queue.json = %q, want %q", b, "[]")
		}
	})

	t.Run("an unwritable output dir is an error, not a silent zero", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir")
		if _, err := DoBusinessLogic(sampleGraph(), missing); err == nil {
			t.Fatal("DoBusinessLogic() = nil error; want a write error")
		}
	})
}

func TestDoGeneratePaths(t *testing.T) {
	t.Run("the first five nodes are the entry points", func(t *testing.T) {
		out := t.TempDir()
		if err := DoGeneratePaths(sampleGraph(), out); err != nil {
			t.Fatalf("DoGeneratePaths() error = %v", err)
		}
		b, err := os.ReadFile(filepath.Join(out, "learning-paths.json"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		var doc struct {
			GeneratedAt string   `json:"generated_at"`
			TotalNodes  int      `json:"total_nodes"`
			Strategy    string   `json:"strategy"`
			EntryPoints []string `json:"entry_points"`
		}
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatalf("learning-paths.json: %v\n%s", err, b)
		}
		if doc.TotalNodes != 5 {
			t.Errorf("total_nodes = %d, want 5", doc.TotalNodes)
		}
		if len(doc.EntryPoints) != 5 {
			t.Errorf("entry_points = %v, want five ids", doc.EntryPoints)
		}
		if doc.Strategy == "" || doc.GeneratedAt == "" {
			t.Errorf("learning-paths.json is missing its metadata: %s", b)
		}
	})

	t.Run("fewer than five nodes yields fewer entry points", func(t *testing.T) {
		out := t.TempDir()
		g := sampleGraph()
		g.Nodes = g.Nodes[:2]
		if err := DoGeneratePaths(g, out); err != nil {
			t.Fatalf("DoGeneratePaths() error = %v", err)
		}
		b, _ := os.ReadFile(filepath.Join(out, "learning-paths.json"))
		if !strings.Contains(string(b), `"entry_points": [`) {
			t.Fatalf("learning-paths.json = %s", b)
		}
		if !strings.Contains(string(b), `"total_nodes": 2`) {
			t.Errorf("total_nodes should follow the node count: %s", b)
		}
	})

	t.Run("an unwritable output dir is an error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir")
		if err := DoGeneratePaths(sampleGraph(), missing); err == nil {
			t.Fatal("DoGeneratePaths() = nil error; want a write error")
		}
	})
}

func TestDoImpactVisualize(t *testing.T) {
	t.Run("the page names the node and points back at the real report", func(t *testing.T) {
		out := t.TempDir()
		if err := DoImpactVisualize("app/main.js", out); err != nil {
			t.Fatalf("DoImpactVisualize() error = %v", err)
		}
		b, err := os.ReadFile(filepath.Join(out, "impact-visualization.html"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		html := string(b)
		for _, want := range []string{
			"<!DOCTYPE html>",
			"Depth analysis configured for: app/main.js",
			"plaesy graph --impact-check NODE --impact-depth N",
		} {
			if !strings.Contains(html, want) {
				t.Errorf("page missing %q", want)
			}
		}
	})

	// CURRENT BEHAVIOUR — suspected bug: the node name is concatenated into the
	// HTML template with no escaping, so a file name containing markup is
	// written into the page as live markup. It is a locally generated report,
	// so the impact is a corrupted page rather than a remote attack, but the
	// fix is one call (html/template instead of string concatenation, or
	// template.JSEscapeString/HTMLEscapeString on the name).
	t.Run("a node name is escaped, not interpolated as markup", func(t *testing.T) {
		out := t.TempDir()
		name := `<img src=x onerror=alert(1)>`
		if err := DoImpactVisualize(name, out); err != nil {
			t.Fatalf("DoImpactVisualize() error = %v", err)
		}
		b, _ := os.ReadFile(filepath.Join(out, "impact-visualization.html"))
		page := string(b)
		// The name is a repository-relative path, so a file or directory
		// literally named like a tag must not become a tag.
		if strings.Contains(page, name) {
			t.Errorf("the raw name reached the page as live markup:\n%s", page)
		}
		if !strings.Contains(page, "&lt;img") {
			t.Errorf("the name should appear escaped instead:\n%s", page)
		}
	})

	t.Run("an unwritable output dir is an error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir")
		if err := DoImpactVisualize("a", missing); err == nil {
			t.Fatal("DoImpactVisualize() = nil error; want a write error")
		}
	})
}

func TestDoApplySemantic(t *testing.T) {
	annotate := func(t *testing.T, content string) string {
		t.Helper()
		p := filepath.Join(t.TempDir(), "annotations.json")
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		return p
	}

	t.Run("a rationale for an existing edge reaches reports.md", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `[{"source":"app/main.js","target":"app/util.js","rationale":"it calls the helper"}]`)
		applied, err := DoApplySemantic(sampleGraph(), ann, out)
		if err != nil {
			t.Fatalf("DoApplySemantic() error = %v", err)
		}
		if applied != 1 {
			t.Fatalf("applied = %d, want 1", applied)
		}
		report, err := os.ReadFile(filepath.Join(out, "reports.md"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if !strings.Contains(string(report), "## Explained relationships (semantic pass)\n- `app/main.js` -> `app/util.js`: it calls the helper\n") {
			t.Fatalf("reports.md is missing the rationale:\n%s", report)
		}
		// The graph artifacts are rewritten so the next reader sees the same run.
		for _, name := range []string{"reports.md", "project.graph.json", "project.html"} {
			if _, err := os.Stat(filepath.Join(out, name)); err != nil {
				t.Errorf("%s was not written: %v", name, err)
			}
		}
	})

	// An annotation for a pair that is not an edge of this graph is not
	// something the report can point at, so it is dropped rather than listed.
	t.Run("an annotation for an unknown pair is dropped", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `[{"source":"a.md","target":"b.md","rationale":"invented"}]`)
		applied, err := DoApplySemantic(sampleGraph(), ann, out)
		if err != nil {
			t.Fatalf("DoApplySemantic() error = %v", err)
		}
		if applied != 0 {
			t.Fatalf("applied = %d, want 0", applied)
		}
		report, _ := os.ReadFile(filepath.Join(out, "reports.md"))
		if strings.Contains(string(report), "invented") {
			t.Errorf("an unapplied rationale reached the report:\n%s", report)
		}
		if strings.Contains(string(report), "Explained relationships") {
			t.Errorf("an empty rationale section was written:\n%s", report)
		}
	})

	t.Run("an empty annotations file is a no-op, not an error", func(t *testing.T) {
		out := t.TempDir()
		applied, err := DoApplySemantic(sampleGraph(), annotate(t, "[]\n"), out)
		if err != nil {
			t.Fatalf("DoApplySemantic() error = %v", err)
		}
		if applied != 0 {
			t.Fatalf("applied = %d, want 0", applied)
		}
	})

	// A missing target must be reported, not absorbed. The old scrape paired
	// the three key lists by index, so the targetless object consumed the NEXT
	// object's target and rationale — reports.md then explained an edge with a
	// rationale nobody wrote for it, and the command exited 0.
	t.Run("an annotation with no target is an error, not a silent mispairing", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `[
  {"source":"app/main.js","target":"app/util.js","rationale":"one"},
  {"source":"app/util.js"},
  {"source":"lonely.md","target":"docs/spec.md","rationale":"two"}
]`)
		applied, err := DoApplySemantic(sampleGraph(), ann, out)
		if err == nil {
			t.Fatalf("DoApplySemantic() = nil error for a malformed annotation (applied %d)", applied)
		}
		if !strings.Contains(err.Error(), "annotation 1") {
			t.Errorf("the error must identify which annotation is wrong, got %v", err)
		}
	})

	// The same for an empty source: the old `"([^"]+)"` pattern could not match
	// an empty string, so such an object was invisible to the scraper while its
	// target and rationale were still collected, shifting every later object
	// one position left.
	t.Run("an empty source is an error, not a shift for every later object", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `[
  {"source":"","target":"app/util.js","rationale":"belongs to the object with no source"},
  {"source":"app/main.js","target":"app/util.js","rationale":"the real one"}
]`)
		if _, err := DoApplySemantic(sampleGraph(), ann, out); err == nil {
			t.Fatal("an annotation with an empty source must be an error")
		}
		// The decisive check: the later object's own rationale must not be
		// attached to an earlier edge.
		report, _ := os.ReadFile(filepath.Join(out, "reports.md"))
		if strings.Contains(string(report), "belongs to the object with no source") {
			t.Errorf("a rationale was attached to an edge its author never wrote: %s", report)
		}
	})

	// The same rule for a missing rationale: an annotation with nothing to say
	// has no business in the report, and borrowing the next object's text is
	// exactly the false statement this parsing change exists to prevent.
	t.Run("an annotation with no rationale is an error", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `[
  {"source":"app/main.js","target":"app/util.js"},
  {"source":"app/util.js","target":"docs/spec.md","rationale":"two"}
]`)
		applied, err := DoApplySemantic(sampleGraph(), ann, out)
		if err == nil {
			t.Fatalf("DoApplySemantic() = nil error for an annotation with no rationale (applied %d)", applied)
		}
		if !strings.Contains(err.Error(), "rationale") {
			t.Errorf("the error must name the missing field, got %v", err)
		}
	})

	// Well-formed annotations are applied per object, each keeping its OWN
	// rationale — the property the positional scrape could not guarantee.
	t.Run("each annotation keeps its own rationale", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `[
  {"source":"app/main.js","target":"app/util.js","rationale":"first-pair"},
  {"source":"README.md","target":"docs/spec.md","rationale":"second-pair"}
]`)
		applied, err := DoApplySemantic(sampleGraph(), ann, out)
		if err != nil {
			t.Fatalf("DoApplySemantic() error = %v", err)
		}
		if applied != 2 {
			t.Fatalf("applied = %d, want 2", applied)
		}
		report, _ := os.ReadFile(filepath.Join(out, "reports.md"))
		got := string(report)
		if !strings.Contains(got, "`app/main.js` -> `app/util.js`: first-pair") {
			t.Errorf("the first pair lost its own rationale:\n%s", got)
		}
		if !strings.Contains(got, "`README.md` -> `docs/spec.md`: second-pair") {
			t.Errorf("the second pair lost its own rationale:\n%s", got)
		}
	})

	// An annotation naming a pair that is not in the graph is skipped, and said
	// out loud on stderr: a silently-ignored file is indistinguishable from an
	// applied one.
	t.Run("an unmatched pair is skipped and reported", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `[{"source":"a.md","target":"b.md","rationale":"nope"}]`)
		applied, err := DoApplySemantic(sampleGraph(), ann, out)
		if err != nil {
			t.Fatalf("DoApplySemantic() error = %v", err)
		}
		if applied != 0 {
			t.Fatalf("applied = %d, want 0", applied)
		}
		report, _ := os.ReadFile(filepath.Join(out, "reports.md"))
		if strings.Contains(string(report), "nope") {
			t.Errorf("an unmatched pair must not reach the report:\n%s", report)
		}
	})

	// Malformed input is an error. The old scrape never noticed: it pulled
	// whatever strings it could find out of any text at all, so a notes file
	// could half-apply.
	t.Run("a file that is not a JSON array is an error", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `these are my notes, not annotations`)
		if _, err := DoApplySemantic(sampleGraph(), ann, out); err == nil {
			t.Fatal("a non-JSON annotations file must be an error")
		}
	})

	// Writing the report and writing the graph are separate steps, and the
	// second one can fail on its own: a directory sitting where project.graph.json
	// belongs makes the report succeed and the save fail. That has to surface
	// as an error, not as a successful run with a stale graph on disk.
	t.Run("a failure to rewrite the graph is reported", func(t *testing.T) {
		out := t.TempDir()
		if err := os.Mkdir(filepath.Join(out, "project.graph.json"), 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		ann := annotate(t, `[{"source":"app/main.js","target":"app/util.js","rationale":"ok"}]`)
		applied, err := DoApplySemantic(sampleGraph(), ann, out)
		if err == nil {
			t.Fatalf("DoApplySemantic() = %d, nil; want the save error", applied)
		}
		if applied != 1 {
			t.Errorf("applied = %d, want the count of rationales accepted before the failure", applied)
		}
	})

	t.Run("a missing annotations file is an error", func(t *testing.T) {
		out := t.TempDir()
		missing := filepath.Join(t.TempDir(), "nope.json")
		if _, err := DoApplySemantic(sampleGraph(), missing, out); err == nil {
			t.Fatal("DoApplySemantic() = nil error; want a read error")
		}
	})

	t.Run("an unwritable output dir is reported as an error", func(t *testing.T) {
		ann := annotate(t, `[]`)
		missing := filepath.Join(t.TempDir(), "no-such-dir")
		applied, err := DoApplySemantic(sampleGraph(), ann, missing)
		if err == nil {
			t.Fatalf("DoApplySemantic() = %d, nil; want a write error", applied)
		}
	})

	// CURRENT BEHAVIOUR — suspected bug: the three keys are collected
	// independently and paired by position, not per object. An annotations
	// file whose first object has no "rationale" therefore gets that later
	// object's rationale attached to the FIRST edge, and the report asserts a
	// A missing rationale is reported, not filled in from the next object. The
	// old scrape pulled all "rationale" values into one list and zipped it
	// against the sources by index, so this object consumed the SECOND object's
	// rationale and reports.md explained a real edge with text its author never
	// wrote for it — and the command exited 0.
	t.Run("a missing rationale is an error, not a shift for every later one", func(t *testing.T) {
		out := t.TempDir()
		ann := annotate(t, `[
  {"source":"app/main.js","target":"app/util.js"},
  {"source":"README.md","target":"docs/spec.md","rationale":"belongs to the second pair"}
]`)
		applied, err := DoApplySemantic(sampleGraph(), ann, out)
		if err == nil {
			t.Fatalf("DoApplySemantic() = nil error for an annotation with no rationale (applied %d)", applied)
		}
		if !strings.Contains(err.Error(), "rationale") {
			t.Errorf("the error must name the missing field, got %v", err)
		}
		// The decisive check: the second object's rationale must not have been
		// attached to the first object's edge.
		report, _ := os.ReadFile(filepath.Join(out, "reports.md"))
		if strings.Contains(string(report), "belongs to the second pair") {
			t.Errorf("a rationale was attached to an edge its author never wrote:\n%s", report)
		}
	})
}
