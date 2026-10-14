package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSaveJSON(t *testing.T) {
	t.Run("a node with no symbols is written as an empty array, not null", func(t *testing.T) {
		// Reports and the HTML viz both read the JSON back; a null there is a
		// shape difference no consumer asked for, and it makes two builds of
		// the same graph differ in the file.
		out := filepath.Join(t.TempDir(), "project.graph.json")
		g := sampleGraph()
		g.Nodes[3].Symbols = nil
		if err := SaveJSON(g, out); err != nil {
			t.Fatalf("SaveJSON() error = %v", err)
		}
		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if strings.Contains(string(b), `"symbols": null`) {
			t.Errorf("a nil Symbols list was serialized as null:\n%s", b)
		}
		if !strings.Contains(string(b), `"symbols": []`) {
			t.Errorf("want an empty symbols array:\n%s", b)
		}
		// SaveJSON normalizes in place, so the caller's graph is consistent
		// with what was written.
		if g.Nodes[3].Symbols == nil {
			t.Error("SaveJSON left the caller's node with a nil Symbols list")
		}
	})

	t.Run("round trip", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "project.graph.json")
		want := sampleGraph()
		if err := SaveJSON(want, out); err != nil {
			t.Fatalf("SaveJSON() error = %v", err)
		}
		got, err := LoadJSON(out)
		if err != nil {
			t.Fatalf("LoadJSON() error = %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("round trip changed the graph:\n got %+v\nwant %+v", got, want)
		}
	})

	t.Run("the file ends with a newline", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "project.graph.json")
		if err := SaveJSON(sampleGraph(), out); err != nil {
			t.Fatalf("SaveJSON() error = %v", err)
		}
		b, _ := os.ReadFile(out)
		if len(b) == 0 || b[len(b)-1] != '\n' {
			t.Errorf("project.graph.json does not end with a newline")
		}
	})

	t.Run("an unwritable path is an error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir", "project.graph.json")
		if err := SaveJSON(sampleGraph(), missing); err == nil {
			t.Fatal("SaveJSON() = nil error; want a write error")
		}
	})
}

func TestLoadJSON(t *testing.T) {
	t.Run("a missing file is an error", func(t *testing.T) {
		if _, err := LoadJSON(filepath.Join(t.TempDir(), "nope.json")); err == nil {
			t.Fatal("LoadJSON() = nil error; want a read error")
		}
	})

	// A truncated or hand-edited project.graph.json is the normal failure here
	// (the file is the input to every query subcommand), and it has to be an
	// error rather than an empty graph that answers every query with "no
	// nodes matched".
	t.Run("malformed JSON is an error, not an empty graph", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "project.graph.json")
		if err := os.WriteFile(p, []byte(`{"nodes": [`), 0o644); err != nil {
			t.Fatal(err)
		}
		g, err := LoadJSON(p)
		if err == nil {
			t.Fatalf("LoadJSON() = %+v, nil; want a parse error", g)
		}
	})

	t.Run("an empty object loads as an empty graph", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "project.graph.json")
		if err := os.WriteFile(p, []byte(`{}`), 0o644); err != nil {
			t.Fatal(err)
		}
		g, err := LoadJSON(p)
		if err != nil {
			t.Fatalf("LoadJSON() error = %v", err)
		}
		if len(g.Nodes) != 0 || len(g.Edges) != 0 {
			t.Fatalf("LoadJSON() = %+v, want an empty graph", g)
		}
	})
}

func TestMapValues(t *testing.T) {
	nodes := []Node{{Type: "doc"}, {Type: "source"}, {Type: "doc"}}
	got := mapValues(nodes, func(n Node) string { return n.Type })
	if !reflect.DeepEqual(got, []string{"doc", "source", "doc"}) {
		t.Fatalf("mapValues() = %v", got)
	}
	if got := mapValues(nil, func(Node) string { return "" }); len(got) != 0 {
		t.Fatalf("mapValues(nil) = %v, want empty", got)
	}
}

func TestCountDesc(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   []countEntry
	}{
		{name: "empty", values: nil, want: []countEntry{}},
		{name: "count descending", values: []string{"a", "b", "b", "c", "c", "c"}, want: []countEntry{{"c", 3}, {"b", 2}, {"a", 1}}},
		{name: "ties break on key ascending", values: []string{"b", "a"}, want: []countEntry{{"a", 1}, {"b", 1}}},
		{name: "the empty string is a key like any other", values: []string{"", "a", ""}, want: []countEntry{{"", 2}, {"a", 1}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := countDesc(tc.values); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("countDesc(%v) = %v, want %v", tc.values, got, tc.want)
			}
		})
	}
}

func readReport(t *testing.T, outFull string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(outFull, "reports.md"))
	if err != nil {
		t.Fatalf("ReadFile(reports.md): %v", err)
	}
	return string(b)
}

// reportSection returns the body of one "## <heading>" section. Several
// sections list node ids in the same "- `id`" shape, so counting them in the
// whole document would mix the god-node list into the orphan count.
func reportSection(report, heading string) string {
	_, after, ok := strings.Cut(report, "## "+heading+"\n")
	if !ok {
		return ""
	}
	body, _, _ := strings.Cut(after, "\n## ")
	return body
}

func TestWriteReport(t *testing.T) {
	t.Run("every section, in order", func(t *testing.T) {
		out := t.TempDir()
		if err := WriteReport(sampleGraph(), out, ""); err != nil {
			t.Fatalf("WriteReport() error = %v", err)
		}
		report := readReport(t, out)
		for _, want := range []string{
			"# Plaesy Graph Report",
			"## Overview\n- Nodes: 5\n- Edges: 3\n",
			"## Nodes by type\n- other: 2\n- source: 2\n- doc: 1\n",
			"## Communities (top-level folders)\n- app: 2 files\n- README.md: 1 files\n- docs: 1 files\n- lonely.md: 1 files\n",
			"## Detected communities (label-propagation clustering)",
			"- Community 0 (2 files): README.md, docs/spec.md\n",
			"- Community 1 (2 files): app/main.js, app/util.js\n",
			"## God nodes (most connected)\n- `README.md` - degree 2\n- `app/main.js` - degree 2\n",
			"## Orphan files (no detected references)\n- `lonely.md`\n",
			"## Suggested questions",
		} {
			if !strings.Contains(report, want) {
				t.Errorf("reports.md missing %q:\n%s", want, report)
			}
		}
		// A community of one is not a community; it is listed under folders.
		if strings.Contains(report, "Community 2") {
			t.Errorf("a single-member community was reported:\n%s", report)
		}
		// The "what connects" question is only meaningful with two groups.
		if !strings.Contains(report, "- What connects `app` to `README.md`?\n") {
			t.Errorf("the cross-group question is missing:\n%s", report)
		}
		if strings.Contains(report, "Explained relationships") {
			t.Errorf("an empty rationale section was written:\n%s", report)
		}
	})

	t.Run("a rationale is appended verbatim", func(t *testing.T) {
		out := t.TempDir()
		rationale := "- `a` -> `b`: because\n"
		if err := WriteReport(sampleGraph(), out, rationale); err != nil {
			t.Fatalf("WriteReport() error = %v", err)
		}
		report := readReport(t, out)
		if !strings.Contains(report, "## Explained relationships (semantic pass)\n"+rationale) {
			t.Fatalf("reports.md is missing the rationale section:\n%s", report)
		}
	})

	t.Run("a graph with no orphans says None", func(t *testing.T) {
		out := t.TempDir()
		g := sampleGraph()
		for i := range g.Nodes {
			g.Nodes[i].Degree = 1
		}
		if err := WriteReport(g, out, ""); err != nil {
			t.Fatalf("WriteReport() error = %v", err)
		}
		if !strings.Contains(readReport(t, out), "## Orphan files (no detected references)\n- None\n") {
			t.Errorf("want the explicit None line:\n%s", readReport(t, out))
		}
	})

	// The orphan list is a report, not a data dump: past 30 entries it has to
	// say how many it dropped, or a reader cannot tell a complete list from a
	// truncated one.
	t.Run("more than thirty orphans are truncated with a count", func(t *testing.T) {
		out := t.TempDir()
		g := &Graph{}
		for i := 0; i < 35; i++ {
			g.Nodes = append(g.Nodes, Node{ID: "orphan" + string(rune('a'+i)) + ".md", Type: "other", Group: "g"})
		}
		if err := WriteReport(g, out, ""); err != nil {
			t.Fatalf("WriteReport() error = %v", err)
		}
		report := readReport(t, out)
		if !strings.Contains(report, "- ... and 5 more\n") {
			t.Errorf("want the overflow count:\n%s", report)
		}
		if n := strings.Count(reportSection(report, "Orphan files (no detected references)"), "- `orphan"); n != 30 {
			t.Errorf("printed %d orphans, want 30:\n%s", n, report)
		}
	})

	// Same shape for the community listing: a big subsystem is truncated so one
	// community cannot own the report.
	t.Run("a community lists at most five members", func(t *testing.T) {
		out := t.TempDir()
		g := &Graph{}
		for i := 0; i < 7; i++ {
			g.Nodes = append(g.Nodes, Node{
				ID: "c" + string(rune('a'+i)) + ".md", Type: "other", Group: "g",
				Community: 0, Degree: i,
			})
		}
		if err := WriteReport(g, out, ""); err != nil {
			t.Fatalf("WriteReport() error = %v", err)
		}
		report := readReport(t, out)
		// degree-descending order, so cg.md (degree 6) leads.
		if !strings.Contains(report, "- Community 0 (7 files): cg.md, cf.md, ce.md, cd.md, cc.md\n") {
			t.Errorf("want the five highest-degree members:\n%s", report)
		}
	})

	t.Run("god nodes stop at ten", func(t *testing.T) {
		out := t.TempDir()
		g := &Graph{}
		for i := 0; i < 15; i++ {
			g.Nodes = append(g.Nodes, Node{ID: "n" + string(rune('a'+i)) + ".md", Type: "other", Group: "g", Degree: i})
		}
		if err := WriteReport(g, out, ""); err != nil {
			t.Fatalf("WriteReport() error = %v", err)
		}
		report := readReport(t, out)
		if n := strings.Count(reportSection(report, "God nodes (most connected)"), "- `n"); n != 10 {
			t.Errorf("printed %d god nodes, want 10:\n%s", n, report)
		}
	})

	t.Run("a single group gets no cross-group question", func(t *testing.T) {
		out := t.TempDir()
		g := &Graph{Nodes: []Node{{ID: "a.md", Type: "other", Group: "g"}, {ID: "b.md", Type: "other", Group: "g"}}}
		if err := WriteReport(g, out, ""); err != nil {
			t.Fatalf("WriteReport() error = %v", err)
		}
		if strings.Contains(readReport(t, out), "What connects") {
			t.Errorf("one group cannot be connected to anything:\n%s", readReport(t, out))
		}
	})

	t.Run("an unwritable output dir is an error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir")
		if err := WriteReport(sampleGraph(), missing, ""); err == nil {
			t.Fatal("WriteReport() = nil error; want a write error")
		}
	})
}

func TestWriteDetectedCommunities(t *testing.T) {
	t.Run("community ids are reported in ascending order", func(t *testing.T) {
		g := &Graph{Nodes: []Node{
			{ID: "b.md", Community: 3, Degree: 1},
			{ID: "a.md", Community: 1, Degree: 1},
			{ID: "c.md", Community: 1, Degree: 5},
			{ID: "solo.md", Community: 2, Degree: 9},
		}}
		var b strings.Builder
		writeDetectedCommunities(&b, g)
		want := "- Community 1 (2 files): c.md, a.md\n"
		if b.String() != want {
			t.Fatalf("writeDetectedCommunities() = %q, want %q", b.String(), want)
		}
	})

	t.Run("a graph of singletons writes nothing", func(t *testing.T) {
		// Distinct community ids: two nodes that both carry the zero value are
		// one community, not two.
		g := &Graph{Nodes: []Node{{ID: "a.md", Community: 0}, {ID: "b.md", Community: 1}}}
		var b strings.Builder
		writeDetectedCommunities(&b, g)
		if b.Len() != 0 {
			t.Fatalf("writeDetectedCommunities() = %q, want empty", b.String())
		}
	})
}

func TestWriteHTML(t *testing.T) {
	t.Run("the graph JSON is embedded and the template is filled in", func(t *testing.T) {
		out := t.TempDir()
		if err := WriteHTML(sampleGraph(), out); err != nil {
			t.Fatalf("WriteHTML() error = %v", err)
		}
		b, err := os.ReadFile(filepath.Join(out, "project.html"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		html := string(b)
		if !strings.Contains(html, "<canvas id=\"graph\"></canvas>") {
			t.Errorf("the page is missing its canvas:\n%s", html[:200])
		}
		// The viz reads `data.nodes` / `data.edges` out of the page, so the
		// embedded JSON has to be a real, complete graph object.
		i := strings.Index(html, "const data = ")
		if i < 0 {
			t.Fatalf("the page does not embed the graph JSON")
		}
		rest := html[i+len("const data = "):]
		j := strings.Index(rest, ";")
		if j < 0 {
			t.Fatalf("the embedded JSON is not terminated")
		}
		var embedded Graph
		if err := json.Unmarshal([]byte(rest[:j]), &embedded); err != nil {
			t.Fatalf("the embedded JSON does not parse: %v\n%s", err, rest[:j])
		}
		if len(embedded.Nodes) != 5 || len(embedded.Edges) != 3 {
			t.Errorf("embedded graph = %d nodes / %d edges, want 5/3", len(embedded.Nodes), len(embedded.Edges))
		}
	})

	t.Run("an unwritable output dir is an error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-dir")
		if err := WriteHTML(sampleGraph(), missing); err == nil {
			t.Fatal("WriteHTML() = nil error; want a create error")
		}
	})
}
