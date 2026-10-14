package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractWritesTheEmbeddedTree(t *testing.T) {
	dest := t.TempDir()
	if err := Extract(dest); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	for _, rel := range []string{
		filepath.Join("templates", "README.md"),
		filepath.Join("instructions", "mapping.json"),
		filepath.Join("scripts", "configs", "platform.json"),
	} {
		if _, err := os.Stat(filepath.Join(dest, rel)); err != nil {
			t.Errorf("Extract did not write %s: %v", rel, err)
		}
	}
}

func TestExtractOverwritesExistingFiles(t *testing.T) {
	dest := t.TempDir()
	stale := filepath.Join(dest, "scripts", "configs", "platform.json")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Extract(dest); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	got, err := os.ReadFile(stale)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) == "stale" {
		t.Error("Extract left the stale file in place; it should overwrite")
	}
}
