package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// docs/metadata.json is a hand-maintained snapshot of the repository's shape.
// Two of its fourteen counts were already wrong within hours of it being
// "regenerated from the tree", which is the failure mode: a number in a
// document drifts silently, and the document is what people trust.
//
// Deleting the file is not an option — prompts/doc.md requires the agent to
// write into its gaps/openQuestions/inferences sections. Generating it just
// moves the same drift behind one more generator. So the counts are guarded
// instead: this test recomputes every one of them and fails on a mismatch, and
// the message says how to fix it.

type metadataFile struct {
	Counts            map[string]int    `json:"counts"`
	CountsHowToVerify map[string]string `json:"countsHowToVerify"`
}

func loadMetadata(t *testing.T) (root string, meta metadataFile) {
	t.Helper()
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root = filepath.Dir(filepath.Dir(docs))
	data, err := os.ReadFile(filepath.Join(root, "docs", "metadata.json"))
	if err != nil {
		t.Fatalf("reading docs/metadata.json: %v", err)
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parsing docs/metadata.json: %v", err)
	}
	return root, meta
}

// countGoFiles counts .go files under dir, recursive, splitting test from
// non-test. Recursive matters for scripts/internal, where every package is a
// subdirectory: a non-recursive ReadDir there reports zero and looks like a
// legitimate answer rather than a broken measurement.
func countGoFiles(t *testing.T, root, dir string, tests bool) int {
	t.Helper()
	n := 0
	filepath.WalkDir(filepath.Join(root, filepath.FromSlash(dir)), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") == tests {
			n++
		}
		return nil
	})
	return n
}

func countMatching(t *testing.T, root, pattern string) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
	if err != nil {
		t.Fatalf("glob %s: %v", pattern, err)
	}
	return len(matches)
}

// countRegex sums matches of re across a set of files, which is how the
// Go-source counts (constructors, register call sites, test functions) are
// defined.
//
// Counting FILES that call register() undercounts: install.go registers four
// commands, one per call site, and is one file. The count has to be of call
// sites.
func countRegex(t *testing.T, root, dir string, re *regexp.Regexp) int {
	t.Helper()
	var n int
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(dir), e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		n += len(re.FindAll(data, -1))
	}
	return n
}

var (
	reNewCmd      = regexp.MustCompile(`func new\w*Cmd\(`)
	reRegister    = regexp.MustCompile(`\bregister\(new`)
	reTestFunc    = regexp.MustCompile(`(?m)^func Test\w+\(`)
	reTestAllDirs = regexp.MustCompile(`(?m)^func Test\w+\(`)
)

func TestMetadataCountsMatchTheTree(t *testing.T) {
	root, meta := loadMetadata(t)

	// Counters that need more than a glob.
	platforms := func() int {
		data, err := os.ReadFile(filepath.Join(root, "scripts", "configs", "platform.json"))
		if err != nil {
			t.Fatal(err)
		}
		var p struct {
			Platforms map[string]json.RawMessage `json:"platforms"`
		}
		if err := json.Unmarshal(data, &p); err != nil {
			t.Fatal(err)
		}
		return len(p.Platforms)
	}()

	// All _test.go files in the whole scripts tree, and the test funcs in them.
	goTestFiles := 0
	goTestFunctions := 0
	filepath.WalkDir(filepath.Join(root, "scripts"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		goTestFiles++
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		goTestFunctions += len(reTestFunc.FindAll(data, -1))
		return nil
	})

	internalPackages := 0
	if entries, err := os.ReadDir(filepath.Join(root, "scripts", "internal")); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				internalPackages++
			}
		}
	}

	actual := map[string]int{
		"goFilesCmd":                countGoFiles(t, root, "scripts/cmd/plaesy", false),
		"goFilesInternal":           countGoFiles(t, root, "scripts/internal", false),
		"internalPackages":          internalPackages,
		"cobraCommandConstructors":  countRegex(t, root, "scripts/cmd/plaesy", reNewCmd),
		"topLevelRegisteredParents": countRegex(t, root, "scripts/cmd/plaesy", reRegister),
		"platforms":                 platforms,
		"instructionFiles":          countMatching(t, root, "instructions/*.instructions.md"),
		"promptFiles":               countMatching(t, root, "prompts/*.md"),
		"dimensionSubPrompts":       countMatching(t, root, "prompts/*/*.md"),
		// Deeper nesting is a separate shape, not a bigger `dimensionSubPrompts`.
		// `prompts/*/*.md` matches one level only, so when the first three-segment
		// selector (`/create:doc:design` → prompts/create/doc/design.md) shipped, the
		// file was invisible to every count in this map — the count stayed at 60
		// and nothing reported the new prompt. Its own measurement is what makes a
		// future `doc/adr.md` visible.
		"nestedSubPrompts": countMatching(t, root, "prompts/*/*/*.md"),
		"agentFiles":       countMatching(t, root, "agents/*.agents.md"),
		"checklistFiles":   countMatching(t, root, "checklists/*.md"),
		"templateFiles":    countMatching(t, root, "templates/*.template.*"),
		"goTestFiles":      goTestFiles,
		"goTestFunctions":  goTestFunctions,
	}

	for key, want := range meta.Counts {
		got, ok := actual[key]
		if !ok {
			t.Errorf("docs/metadata.json counts has key %q, which this test does not know how to measure; "+
				"either add a measurement or drop the key", key)
			continue
		}
		if got != want {
			t.Errorf("docs/metadata.json counts.%s = %s, actual %s — update the number, or delete the count if it is no longer meaningful",
				key, strconv.Itoa(want), strconv.Itoa(got))
		}
	}
	for key := range actual {
		if _, ok := meta.Counts[key]; !ok {
			t.Errorf("docs/metadata.json counts is missing %q, which this test measures as %d", key, actual[key])
		}
	}
}

// TestMetadataFilesReadExist catches a subtler rot: the file lists the sources
// it was derived from, and a source that has been renamed leaves a dangling
// entry that still looks like provenance.
func TestMetadataFilesReadExist(t *testing.T) {
	root := repoDocs(t)
	if root == "" {
		t.Skip("not a repository checkout")
	}
	repoRoot := filepath.Dir(filepath.Dir(root))

	data, err := os.ReadFile(filepath.Join(repoRoot, "docs", "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		FilesRead []string `json:"filesRead"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	for _, f := range m.FilesRead {
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(f))); err != nil {
			t.Errorf("docs/metadata.json filesRead lists %q, which does not exist", f)
		}
	}
}
