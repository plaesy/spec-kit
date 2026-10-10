package main

// E2E smoke test ported from testing/smoke/smoke-e2e.sh: builds a fixture
// project, runs the graph + analyze pipeline in-process (via the cobra
// commands directly, not a subprocess), and validates the generated
// artifacts' shape and internal consistency.

import (
	"encoding/json"
	"io"
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
		"docs/guide.md":      "# Guide\n\nImports [src/lib.js](src/lib.js) and depends on [../README.md](README.md).\n",
		"src/lib.js":         "import { helper } from \"./helper.js\";\nexport function lib() { return helper(); }\n",
		"src/helper.js":      "export function helper() { return \"ok\"; }\n",
		"src/app.js":         "import { lib } from \"./lib.js\";\nconsole.log(lib());\n",
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
		// --no-index: analyze builds the semantic index by default, and that
		// step loads a local embedding model. A smoke test has no business
		// reaching the network — and a test whose result depends on the
		// machine's connectivity is not a test.
		if err := runCmd(t, newAnalyzeCmd(), []string{fixture, "--no-index"}); err != nil {
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
		"project total_files":           project.ProjectSummary.TotalFiles,
		"project total count":           project.Structure.Files.Total,
		"project code count":            project.Structure.Files.Code,
		"project documentation count":   project.Structure.Files.Documentation,
		"project configuration count":   project.Structure.Files.Configuration,
		"structure source_code count":   structure.FileTypes.SourceCode,
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

// A command that needs a required flag must fail when it is missing. Two shapes
// both print something useful and exit 0: `return cmd.Usage()` (Usage returns
// nil) and "print the problem and return nil". A script cannot tell either from
// success, and main.go only exits non-zero on a non-nil error.
//
// These are spelled `create image` rather than the old `generate-image`
// because a test that names a command that no longer exists still passes —
// cobra rejects the unknown name, which is also a non-nil error. The
// assertions below would have held for the wrong reason.
func TestMissingRequiredFlagsExitNonZero(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"create image with no flags", []string{"create", "image"}},
		{"create image with only --prompt", []string{"create", "image", "--prompt", "a cat"}},
		{"create image with only --out", []string{"create", "image", "--out", "cat.png"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := rootForTest()
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs(tc.args)
			if err := root.Execute(); err == nil {
				t.Fatalf("%v exited 0 with a required flag missing", tc.args)
			}
		})
	}
}

// --apply-semantic takes a file of annotations to merge. A missing input file
// used to print "[plaesy-graph] Annotations file not found" and return nil, so
// a typo'd path was indistinguishable from success to a script. (The
// neighbouring --semantic-queue case is different on purpose: "nothing queued"
// is a legitimate outcome of a real run, a missing file is not.)
func TestGraphApplySemanticOnAMissingFileFails(t *testing.T) {
	fixture := filepath.Join(t.TempDir(), "project")
	writeFixture(t, fixture)
	if err := runCmd(t, newGraphCmd(), []string{"--path", fixture}); err != nil {
		t.Fatalf("graph build: %v", err)
	}
	missing := filepath.Join(fixture, "no-such-annotations.json")
	if _, err := os.Stat(missing); err == nil {
		t.Fatal("fixture is wrong: the annotations file exists")
	}
	err := runCmd(t, newGraphCmd(), []string{"--path", fixture, "--apply-semantic", missing})
	if err == nil {
		t.Fatal("--apply-semantic on a missing annotations file exited 0")
	}
	if !strings.Contains(err.Error(), "annotations file not found") {
		t.Errorf("error %q should say what was wrong", err)
	}
}

// The pre-1.0 spellings were removed rather than aliased (see create.go for
// why this is deliberately different from the `validate-*` collapse). An alias
// left behind is invisible: it is deprecated, so it is hidden from help and
// from the documentation bar the drift tests enforce, and nothing would
// notice it still working. These assert the removal.
func TestRemovedCreateSpellingsAreGone(t *testing.T) {
	for _, old := range []string{
		"generate-image", "create-new-feature",
		// `create` itself went too. It briefly owned `feature` and `image` as
		// subcommands, but a verb grouping two unrelated nouns made the two
		// spellings disagree about what a "feature" and an "image" are: they
		// became `features create` and `images create`, each beside its own
		// listing verb.
		"create",
	} {
		root := rootForTest()
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs([]string{old, "--help"})
		err := root.Execute()
		if err == nil {
			t.Errorf("%q still resolves; it was removed outright, not aliased", old)
			continue
		}
		if !strings.Contains(err.Error(), "unknown command") {
			t.Errorf("%q failed with %v, want an unknown-command error", old, err)
		}
	}
}

// `features` and `images` are the nouns; the scaffolding action is `create` under
// each. Both parents must actually own their subcommand, and neither spelling
// may remain at the top level — otherwise the split is cosmetic and the help
// text still carries two entries for one action.
func TestFeaturesAndImagesOwnTheirCreateSubcommand(t *testing.T) {
	root := rootForTest()
	parents := map[string]*cobra.Command{}
	for _, cmd := range root.Commands() {
		parents[cmd.Name()] = cmd
	}

	for _, tc := range []struct {
		parent string
		subs   []string
	}{
		{"features", []string{"create", "paths", "validate"}},
		{"images", []string{"create"}},
	} {
		parent, ok := parents[tc.parent]
		if !ok {
			t.Errorf("no top-level %q command", tc.parent)
			continue
		}
		subs := map[string]bool{}
		for _, sub := range parent.Commands() {
			subs[sub.Name()] = true
		}
		for _, want := range tc.subs {
			if !subs[want] {
				t.Errorf("`plaesy %s %s` is missing", tc.parent, want)
			}
		}
	}

	// Every parent here must also be useful with no subcommand, which is what
	// makes `plaesy features` a listing rather than a help screen.
	for _, name := range []string{"features", "images", "context", "platforms", "config", "stack", "tasks"} {
		if parents[name] == nil {
			t.Errorf("no top-level %q command", name)
		}
	}
}
