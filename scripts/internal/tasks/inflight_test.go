package tasks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderInFlight(t *testing.T) {
	refs := []TaskRef{
		{Spec: "s1", Status: "backlog", Name: "low_a", Title: "A"},
		{Spec: "s1", Status: "todo", Name: "high_b", Title: "B"},
		{Spec: "s2", Status: "doing", Name: "high_c", Title: "C"},
		{Spec: "s2", Status: "done", Name: "high_d", Title: "D"},
	}
	got := RenderInFlight(refs)
	want := "- [s1] high_b (todo): B\n- [s2] high_c (doing): C\n"
	if got != want {
		t.Fatalf("RenderInFlight =\n%q\nwant\n%q", got, want)
	}
	if RenderInFlight(nil) != "- none\n" {
		t.Errorf("empty in-flight should say none, got %q", RenderInFlight(nil))
	}
}

func TestReplaceSection(t *testing.T) {
	content := "# Session\n\n## Context\n\nkeep me\n\n## In-Flight Tasks\n\n- old line\n- old line 2\n\n## Next\n\n1. keep next\n"
	got := ReplaceSection(content, InFlightHeading, "- new line\n")
	want := "# Session\n\n## Context\n\nkeep me\n\n## In-Flight Tasks\n- new line\n\n## Next\n\n1. keep next\n"
	if got != want {
		t.Fatalf("ReplaceSection =\n%q\nwant\n%q", got, want)
	}

	crlf := strings.ReplaceAll(content, "\n", "\r\n")
	gotCRLF := ReplaceSection(crlf, InFlightHeading, "- new line\n")
	if strings.Count(gotCRLF, "\r\n") != strings.Count(gotCRLF, "\n") {
		t.Errorf("CRLF content must keep CRLF endings: %q", gotCRLF)
	}
	if !strings.Contains(gotCRLF, "- new line") || strings.Contains(gotCRLF, "old line") {
		t.Errorf("CRLF section not replaced: %q", gotCRLF)
	}

	appended := ReplaceSection("# Session\n", InFlightHeading, "- x\n")
	if appended != "# Session\n\n## In-Flight Tasks\n- x\n" {
		t.Errorf("missing heading must be appended, got %q", appended)
	}
}

func TestSyncInFlight(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".plaesy"), 0o755); err != nil {
		t.Fatal(err)
	}
	original := "# Session\n\n## In-Flight Tasks\n- stale\n\n## Next\n\n1. keep\n"
	if err := os.WriteFile(filepath.Join(root, ".plaesy", "context.md"), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	refs := []TaskRef{{Spec: "s1", Status: "doing", Name: "high_a", Title: "A"}}

	changed, err := SyncInFlight(root, refs)
	if err != nil || !changed {
		t.Fatalf("first sync: changed=%v err=%v", changed, err)
	}
	data, _ := os.ReadFile(filepath.Join(root, ".plaesy", "context.md"))
	if !strings.Contains(string(data), "- [s1] high_a (doing): A") || strings.Contains(string(data), "stale") {
		t.Fatalf("section not replaced:\n%s", data)
	}
	if !strings.Contains(string(data), "## Next\n\n1. keep") {
		t.Fatalf("rest of file changed:\n%s", data)
	}

	changed, err = SyncInFlight(root, refs)
	if err != nil || changed {
		t.Fatalf("second sync must be a no-op: changed=%v err=%v", changed, err)
	}
}
