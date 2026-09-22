package main

// E2E smoke test ported from testing/smoke/smoke-e2e.sh: builds a fixture
// project, runs the graph + analyze pipeline in-process (via the cobra
// commands directly, not a subprocess), and validates the generated
// artifacts' shape and internal consistency.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func writeFixture(t *testing.T, dir string) {
	t.Helper()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(dir, "src"), 0o755))
	must(os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	must(os.MkdirAll(filepath.Join(dir, ".plaesy"), 0o755))

	files := map[string]string{
		"README.md": "# Smoke Fixture\n\nSee [docs/guide.md](docs/guide.md) and [src/lib.js](src/lib.js).\n",
		"package.json": `{
  "name": "plaesy-smoke-fixture",
  "version": "1.0.0"
}
`,
		"docs/guide.md":     "# Guide\n\nImports [src/lib.js](src/lib.js) and depends on [../README.md](README.md).\n",
		"src/lib.js":        "import { helper } from \"./helper.js\";\nexport function lib() { return helper(); }\n",
		"src/helper.js":     "export function helper() { return \"ok\"; }\n",
		"src/app.js":        "import { lib } from \"./lib.js\";\nconsole.log(lib());\n",
		".plaesy/context.md": "context sentinel: do not modify\n",
		".plaesy/memory.md":  "memory sentinel: do not modify\n",
	}
	for rel, content := range files {
		must(os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644))
	}
}

// runCmd invokes a freshly constructed cobra command in-process (mirrors
// running the compiled binary, but without spawning a subprocess).
func runCmd(t *testing.T, cmd *cobra.Command, args []string) error {
	t.Helper()
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestE2ESmoke(t *testing.T) {
	work := t.TempDir()
	fixture := filepath.Join(work, "project")
	writeFixture(t, fixture)

	beforeContext, err := os.ReadFile(filepath.Join(fixture, ".plaesy", "context.md"))
	if err != nil {
		t.Fatal(err)
	}
	beforeMemory, err := os.ReadFile(filepath.Join(fixture, ".plaesy", "memory.md"))
	if err != nil {
		t.Fatal(err)
	}

	analysis := filepath.Join(fixture, ".plaesy", "analysis")

	t.Run("graph", func(t *testing.T) {
		if err := runCmd(t, newGraphCmd(), []string{"--path", fixture}); err != nil {
			t.Fatalf("graph command failed: %v", err)
		}

		for _, name := range []string{"project.graph.json", "reports.md", "project.html"} {
			if _, err := os.Stat(filepath.Join(analysis, name)); err != nil {
				t.Errorf("expected %s to exist: %v", name, err)
			}
		}

		data, err := os.ReadFile(filepath.Join(analysis, "project.graph.json"))
		if err != nil {
			t.Fatalf("reading project.graph.json: %v", err)
		}
		var g struct {
			Nodes []struct {
				ID string `json:"id"`
			} `json:"nodes"`
			Edges []struct {
				Source string `json:"source"`
				Target string `json:"target"`
			} `json:"edges"`
		}
		if err := json.Unmarshal(data, &g); err != nil {
			t.Fatalf("project.graph.json is not valid JSON: %v", err)
		}
		if len(g.Nodes) == 0 {
			t.Error("expected at least one node in project.graph.json")
		}
		nodeIDs := map[string]bool{}
		for _, n := range g.Nodes {
			nodeIDs[n.ID] = true
		}
		for _, e := range g.Edges {
			if !nodeIDs[e.Source] || !nodeIDs[e.Target] {
				t.Errorf("edge references unknown node: %+v", e)
			}
		}
	})

	t.Run("graph_query", func(t *testing.T) {
		if err := runCmd(t, newGraphCmd(), []string{"--path", fixture, "--query", "lib.js"}); err != nil {
			t.Fatalf("graph --query failed: %v", err)
		}
	})

	t.Run("analyze", func(t *testing.T) {
		if err := runCmd(t, newAnalyzeCmd(), []string{fixture}); err != nil {
			t.Fatalf("analyze command failed: %v", err)
		}

		for _, name := range []string{"project.json", "project.structure.json", "overview.md"} {
			if _, err := os.Stat(filepath.Join(analysis, name)); err != nil {
				t.Errorf("expected %s to exist: %v", name, err)
			}
		}

		if _, err := os.Stat(filepath.Join(fixture, "scripts", "test-runner.sh")); err == nil {
			t.Error("test-runner.sh should not be generated")
		}

		afterContext, err := os.ReadFile(filepath.Join(fixture, ".plaesy", "context.md"))
		if err != nil {
			t.Fatal(err)
		}
		if string(afterContext) != string(beforeContext) {
			t.Error("context.md was modified by analyze")
		}
		afterMemory, err := os.ReadFile(filepath.Join(fixture, ".plaesy", "memory.md"))
		if err != nil {
			t.Fatal(err)
		}
		if string(afterMemory) != string(beforeMemory) {
			t.Error("memory.md was modified by analyze")
		}

		validateAnalysisContract(t, analysis)
	})
}

// validateAnalysisContract mirrors validate_analysis_contract() in
// smoke-e2e.sh: cross-checks project.json / project.structure.json file
// counts for internal consistency and asserts overview.md's expected header.
func validateAnalysisContract(t *testing.T, analysisDir string) {
	t.Helper()

	projectRaw, err := os.ReadFile(filepath.Join(analysisDir, "project.json"))
	if err != nil {
		t.Fatalf("reading project.json: %v", err)
	}
	structureRaw, err := os.ReadFile(filepath.Join(analysisDir, "project.structure.json"))
	if err != nil {
		t.Fatalf("reading project.structure.json: %v", err)
	}
	overviewRaw, err := os.ReadFile(filepath.Join(analysisDir, "overview.md"))
	if err != nil {
		t.Fatalf("reading overview.md: %v", err)
	}

	var project struct {
		ProjectSummary struct {
			TotalFiles int `json:"total_files"`
		} `json:"project_summary"`
		Structure struct {
			Files struct {
				Total         int `json:"total"`
				Code          int `json:"code"`
				Documentation int `json:"documentation"`
				Configuration int `json:"configuration"`
			} `json:"files"`
		} `json:"structure"`
	}
	if err := json.Unmarshal(projectRaw, &project); err != nil {
		t.Fatalf("project.json is not valid JSON: %v", err)
	}

	var structure struct {
		FileTypes struct {
			SourceCode    int `json:"source_code"`
			Documentation int `json:"documentation"`
			Configuration int `json:"configuration"`
		} `json:"file_types"`
	}
	if err := json.Unmarshal(structureRaw, &structure); err != nil {
		t.Fatalf("project.structure.json is not valid JSON: %v", err)
	}

	counts := map[string]int{
		"project total_files":          project.ProjectSummary.TotalFiles,
		"project total count":          project.Structure.Files.Total,
		"project code count":           project.Structure.Files.Code,
		"project documentation count":  project.Structure.Files.Documentation,
		"project configuration count":  project.Structure.Files.Configuration,
		"structure source_code count":  structure.FileTypes.SourceCode,
		"structure documentation count": structure.FileTypes.Documentation,
		"structure configuration count": structure.FileTypes.Configuration,
	}
	for label, v := range counts {
		if v <= 0 {
			t.Errorf("%s must be a positive integer, got %d", label, v)
		}
	}

	if project.Structure.Files.Total != project.ProjectSummary.TotalFiles {
		t.Error("project file totals do not match")
	}
	if project.Structure.Files.Code != structure.FileTypes.SourceCode {
		t.Error("source code counts do not match")
	}
	if project.Structure.Files.Documentation != structure.FileTypes.Documentation {
		t.Error("documentation counts do not match")
	}
	if project.Structure.Files.Configuration != structure.FileTypes.Configuration {
		t.Error("configuration counts do not match")
	}
	if !strings.HasPrefix(string(overviewRaw), "# Project Analysis Overview") {
		t.Error("overview.md has an unexpected format")
	}
}
