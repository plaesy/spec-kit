package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListContractsFiltersBySpec(t *testing.T) {
	tasksDir := t.TempDir()
	mk := func(spec, status, task string, withContract bool) {
		dir := filepath.Join(tasksDir, status, task)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, TaskFileName), []byte("---\nspec: "+spec+"\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if withContract {
			if err := os.WriteFile(filepath.Join(dir, ContractFileName), []byte(`{"id":"x","checklist":[]}`), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	mk("main", "done", "implement-rename", true)
	mk("main", "doing", "implement-record", true)
	mk("main", "todo", "no-contract-yet", false)
	mk("legacy", "done", "save-2026-10-18", true)

	all, err := ListContracts(tasksDir, "")
	if err != nil || len(all) != 3 {
		t.Fatalf("want 3 contracts across specs, got %d: %+v, %v", len(all), all, err)
	}
	if all[0].Task != "implement-record" || all[0].Status != "doing" {
		t.Errorf("first = %+v, want the doing task first", all[0])
	}
	one, err := ListContracts(tasksDir, "legacy")
	if err != nil || len(one) != 1 || one[0].Task != "save-2026-10-18" {
		t.Fatalf("spec filter: %+v, %v", one, err)
	}
}
