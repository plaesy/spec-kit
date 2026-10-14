package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The drift this guards against is silent and self-reinforcing: `init` never
// overwrites, so a project initialised before an instruction was edited keeps
// the old copy, and nothing in the build ever said so. `plaesy reload` fixes it
// when someone remembers to run it; this fails the build when they don't.
//
// The three user-owned files are excluded by design. They are created from
// templates and then edited, so "installed" does not mean "generated" and
// comparing them against source would report every real memory as drift.

var userOwnedFiles = map[string]bool{
	"memory.md":              true,
	"context.md":             true,
	"state.json":             true,
	"instructions.md":        true, // generated index, maintained by hand
	"template-registry.json": true,
}

// repoRootForFreshness walks up from the package directory to the repository
// root. The tests live in scripts/internal/scaffold, so the repo root is four
// levels up.
func repoRootForFreshness(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "instructions", "mapping.json")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("not a repository checkout")
	return ""
}

// mappingLists returns the always_load and scope_load filenames from
// instructions/mapping.json.
func mappingLists(t *testing.T, root string) (always, scope []string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "instructions", "mapping.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Mappings struct {
			AlwaysLoad []string `json:"always_load"`
			ScopeLoad  []string `json:"scope_load"`
		} `json:"mappings"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m.Mappings.AlwaysLoad, m.Mappings.ScopeLoad
}

func TestInstalledInstructionsMatchSource(t *testing.T) {
	root := repoRootForFreshness(t)
	always, scope := mappingLists(t, root)

	installedDir := filepath.Join(root, ".plaesy", "instructions")
	entries, err := os.ReadDir(installedDir)
	if err != nil {
		t.Skipf(".plaesy/instructions not present: %v", err)
	}

	installed := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			installed[e.Name()] = true
		}
	}

	checked := 0
	for _, name := range append(append([]string{}, always...), scope...) {
		dstName := strings.TrimSuffix(filepath.Base(name), ".instructions.md") + ".md"
		if !installed[dstName] {
			// Not every mapping entry is necessarily installed in this checkout
			// (scope_load is installed, but a stale tree may predate a rename).
			// Absence is reported by its own concern; this test is about drift.
			continue
		}
		src := filepath.Join(root, "instructions", name)
		dst := filepath.Join(installedDir, dstName)
		srcData, err := os.ReadFile(src)
		if err != nil {
			t.Errorf("mapping.json lists %s but it does not exist: %v", name, err)
			continue
		}
		dstData, err := os.ReadFile(dst)
		if err != nil {
			t.Errorf("%s: %v", dstName, err)
			continue
		}
		checked++
		if string(srcData) != string(dstData) {
			t.Errorf("%s has drifted from %s — run `plaesy reload`", dstName, name)
		}
	}
	if checked == 0 {
		t.Skip("no installed instructions to compare")
	}
}

func TestInstalledTemplatesMatchSource(t *testing.T) {
	root := repoRootForFreshness(t)
	installedDir := filepath.Join(root, ".plaesy", "templates")

	entries, err := os.ReadDir(installedDir)
	if err != nil {
		t.Skipf(".plaesy/templates not present: %v", err)
	}

	checked := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Not everything in .plaesy/templates came from templates/: the
		// registry is a hand-maintained index, and README.md is the human one.
		if userOwnedFiles[name] {
			continue
		}
		src := filepath.Join(root, "templates", name)
		if _, err := os.Stat(src); err != nil {
			// A template with no source is a user file or a leftover. reload
			// never deletes, so this is legal; it is not drift.
			continue
		}
		srcData, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		dstData, err := os.ReadFile(filepath.Join(installedDir, name))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		checked++
		if string(srcData) != string(dstData) {
			t.Errorf(".plaesy/templates/%s has drifted from templates/%s — run `plaesy reload`", name, name)
		}
	}
	if checked == 0 {
		t.Skip("no installed templates to compare")
	}
}

// TestReloadWouldFindNothingToDo is the closure test: if the tree is fresh and
// reload is a no-op, then drift is genuinely absent rather than merely
// untested. It also pins the idempotence the other reload tests rely on.
func TestReloadOnThisRepoIsIdempotent(t *testing.T) {
	root := repoRootForFreshness(t)
	if _, err := os.Stat(filepath.Join(root, ".plaesy")); err != nil {
		t.Skip(".plaesy not present")
	}

	report, err := Reload(ReloadOptions{TargetDir: root, PlaesyHome: root, DryRun: true})
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if report.Changed() {
		t.Errorf("`plaesy reload` reports pending changes on the repository's own .plaesy/:\n  updated: %v\n  created: %v\nRun `plaesy reload` to bring the tree forward.",
			report.Updated, report.Created)
	}
}
