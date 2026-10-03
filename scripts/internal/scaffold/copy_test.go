package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// repoRoot walks up from this test file's directory to the spec-kit repository
// root (the folder that contains instructions/mapping.json).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "instructions", "mapping.json")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repo root (instructions/mapping.json) from %s", filepath.Dir(file))
		}
		dir = parent
	}
}

// TestLoadAlwaysLoadAssertsMappingIsSourceOfTruth locks in the contract that
// `plaesy init`'s always-load copy list is DRIVEN BY mapping.json (not a
// hard-coded duplicate). It asserts the framework anchors are present and that
// files consolidated away by the 2026-09 optimization are absent.
func TestLoadAlwaysLoadAssertsMappingIsSourceOfTruth(t *testing.T) {
	home := repoRoot(t)
	names, err := loadAlwaysLoad(home)
	if err != nil {
		t.Fatalf("loadAlwaysLoad: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("expected non-empty always_load list from mapping.json")
	}

	got := map[string]bool{}
	for _, n := range names {
		got[n] = true
	}

	// Anchors the framework depends on must be present.
	for _, need := range []string{
		"plaesy.instructions.md",
		"context-engineering.instructions.md",
		"dimension-mapping.instructions.md",
		"error-recovery.instructions.md",
	} {
		if !got[need] {
			t.Errorf("always_load missing %q (must be driven by mapping.json)", need)
		}
	}

	// Files consolidated away (2026-09) must NOT be re-introduced by a stale
	// hard-coded list — i.e. there is no duplicate source of truth.
	for _, gone := range []string{
		"plaesy-trim.instructions.md",
		"plaesy-graph.instructions.md",
		"error-recovery-predictive.instructions.md",
	} {
		if got[gone] {
			t.Errorf("always_load should NOT contain %q (consolidated — proves no hard-coded duplicate)", gone)
		}
	}

	// Cross-check against mapping.json directly so the test fails loudly if the
	// two ever drift (the exact bug this change prevents).
	raw, err := os.ReadFile(filepath.Join(home, "instructions", "mapping.json"))
	if err != nil {
		t.Fatalf("re-read mapping.json: %v", err)
	}
	var m struct {
		Mappings struct {
			AlwaysLoad []string `json:"always_load"`
		} `json:"mappings"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse mapping.json: %v", err)
	}
	want := map[string]bool{}
	for _, n := range m.Mappings.AlwaysLoad {
		want[n] = true
	}
	for n := range want {
		if !got[n] {
			t.Errorf("loadAlwaysLoad diverged from mapping.json: %q in json but not returned", n)
		}
	}
	for n := range got {
		if !want[n] {
			t.Errorf("loadAlwaysLoad diverged from mapping.json: %q returned but not in json", n)
		}
	}
}

// TestTemplateRegistryPathsExist locks in that every path advertised by
// templates/template-registry.json is a real file. The registry is described as
// machine-readable and used to resolve {template} references, so an entry
// pointing at a missing file is a latent runtime failure: that is exactly how
// templates/agent-file-template.md survived in the registry, in
// templates/README.md, and in docs/scripts/update-agent-context.md while being
// absent from the repo — the file internal/agentcontext read at runtime. (It was
// restored, then deliberately removed on 2026-09-25 along with its only reader;
// this test is what keeps that class of drift from coming back.)
func TestTemplateRegistryPathsExist(t *testing.T) {
	home := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(home, "templates", "template-registry.json"))
	if err != nil {
		t.Fatalf("read template-registry.json: %v", err)
	}
	var reg struct {
		TemplateRegistry struct {
			Categories map[string]struct {
				Templates []struct {
					ID   string `json:"id"`
					Path string `json:"path"`
				} `json:"templates"`
			} `json:"categories"`
		} `json:"templateRegistry"`
	}
	if err := json.Unmarshal(raw, &reg); err != nil {
		t.Fatalf("parse template-registry.json: %v", err)
	}
	if len(reg.TemplateRegistry.Categories) == 0 {
		t.Fatal("template-registry.json declares no categories")
	}
	checked := 0
	for category, c := range reg.TemplateRegistry.Categories {
		for _, tmpl := range c.Templates {
			checked++
			if tmpl.Path == "" {
				t.Errorf("registry category %q has entry %q with an empty path", category, tmpl.ID)
				continue
			}
			if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(tmpl.Path))); err != nil {
				t.Errorf("registry entry %q (%s) points at a missing file: %s", tmpl.ID, category, tmpl.Path)
			}
		}
	}
	if checked < 50 {
		t.Errorf("only %d registry entries parsed — the JSON shape changed?", checked)
	}
}

// The memory.md fallback branch was unreachable, and a copy failure said
// nothing at all. The chain was `else if _, err := copyFile(...); err == nil
// { if stat(memoryTarget) == nil {ok} else {not found, create empty} }`, and
// copyFile returns a nil error only when the destination already exists —
// which the Stat immediately above had just ruled out. So a missing
// templates/memory.template.md (a partial PLAESY_HOME) or an unwritable target
// fell through the whole chain with no line printed, and the function went on
// to report "Instructions copied". That is the constitution's NN-08: a command
// never reports success when it changed nothing. The observable consequence is
// that memory.md was simply not there afterwards.
func TestCopyInstructionsAlwaysProducesMemoryMd(t *testing.T) {
	home := repoRoot(t)
	for _, tc := range []struct {
		name       string
		noTemplate bool
		wantEmpty  bool
	}{
		{"template present", false, false},
		{"template absent", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			work := t.TempDir()
			// %TEMP% is C:\msys64\tmp on this machine and a nested fixture
			// under a long test name overflows MAX_PATH; one level keeps it
			// inside the limit.
			root := filepath.Join(work, "plaesy-test", "project")
			srcHome := home
			if tc.noTemplate {
				srcHome = filepath.Join(work, "plaesy-test", "partial-home")
				copyTree(t, home, srcHome)
				if err := os.Remove(filepath.Join(srcHome, "templates", "memory.template.md")); err != nil {
					t.Fatalf("remove template: %v", err)
				}
			}
			if err := os.MkdirAll(root, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := copyInstructions(srcHome, root, copyPolicy{}); err != nil {
				t.Fatalf("copyInstructions: %v", err)
			}
			got, err := os.ReadFile(filepath.Join(root, "memory.md"))
			if err != nil {
				t.Fatalf("memory.md was not produced at all: %v", err)
			}
			if tc.wantEmpty && len(got) != 0 {
				t.Errorf("with no template the documented fallback is an empty memory.md, got %d bytes", len(got))
			}
			if !tc.wantEmpty && len(got) == 0 {
				t.Error("memory.md was created empty even though the template was present")
			}
		})
	}
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dst, err)
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyTree(t, s, d)
			continue
		}
		data, err := os.ReadFile(s)
		if err != nil {
			t.Fatalf("read %s: %v", s, err)
		}
		if err := os.WriteFile(d, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", d, err)
		}
	}
}
