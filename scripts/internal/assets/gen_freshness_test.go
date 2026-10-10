package assets

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// repoRootForFreshness walks up from the test's working directory looking
// for the repo root marker. It skips (rather than fails) when run outside a
// full checkout, matching scaffold's TestInstalled*MatchSource tests.
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

// TestAssetsDataMatchesSource fails if data/ (the embedded copy) has drifted
// from its real sources — i.e. someone edited templates/, instructions/,
// prompts/, agents/, checklists/ or scripts/configs without re-running
// `go run ./internal/assets/gen` (wired into `make build`/`make assets`).
func TestAssetsDataMatchesSource(t *testing.T) {
	root := repoRootForFreshness(t)

	sources := map[string]string{
		"templates":                         filepath.Join(root, "templates"),
		"instructions":                      filepath.Join(root, "instructions"),
		"prompts":                           filepath.Join(root, "prompts"),
		"agents":                            filepath.Join(root, "agents"),
		"checklists":                        filepath.Join(root, "checklists"),
		filepath.Join("scripts", "configs"): filepath.Join(root, "scripts", "configs"),
	}

	checked := 0
	for destRel, srcDir := range sources {
		entries, err := os.ReadDir(srcDir)
		if err != nil {
			t.Fatalf("read source %s: %v", srcDir, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			srcPath := filepath.Join(srcDir, e.Name())
			want, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatalf("read %s: %v", srcPath, err)
			}
			embeddedPath := filepath.ToSlash(filepath.Join("data", destRel, e.Name()))
			got, err := fs.ReadFile(data, embeddedPath)
			if err != nil {
				t.Errorf("embedded assets missing %s (run `go run ./internal/assets/gen`): %v", embeddedPath, err)
				continue
			}
			if !bytes.Equal(got, want) {
				t.Errorf("embedded %s is stale relative to %s (run `go run ./internal/assets/gen`)", embeddedPath, srcPath)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no files checked; the source directories appear to be empty")
	}
}
