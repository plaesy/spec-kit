package analyze

import (
	"context"
	"os"
	"testing"
)

// TestMain replaces the document-index seam for the whole package. The real
// build loads the embedding model, and no analyze test may reach it through Run.
// TestRunDocIndexSeam restores the seam to observe the calls it receives.
func TestMain(m *testing.M) {
	ensureDocs = func(context.Context, string, string, bool) (bool, error) { return false, nil }
	os.Exit(m.Run())
}

// The document index follows the same rules as the semantic index: it is built
// by default, forced by --force, and skipped by --no-index.
func TestRunDocIndexSeam(t *testing.T) {
	type call struct{ force bool }
	var calls []call
	orig := ensureDocs
	ensureDocs = func(_ context.Context, _, _ string, force bool) (bool, error) {
		calls = append(calls, call{force: force})
		return false, nil
	}
	t.Cleanup(func() { ensureDocs = orig })
	captureLogs(t)
	stubSemantic(t, false, nil)

	project := analysisFixture(t)
	if err := Run(Options{ProjectPath: project, Force: true}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(calls) == 0 || !calls[len(calls)-1].force {
		t.Errorf("doc index calls = %+v, want a final call with force=true", calls)
	}

	calls = nil
	if err := Run(Options{ProjectPath: project, NoIndex: true}); err != nil {
		t.Fatalf("Run NoIndex: %v", err)
	}
	if len(calls) != 0 {
		t.Errorf("--no-index still reached the doc index %d time(s)", len(calls))
	}
}
