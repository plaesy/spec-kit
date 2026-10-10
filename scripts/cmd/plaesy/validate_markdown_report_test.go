package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// `--max-files` documents "0 = all", and the report loop already treats a
// non-positive limit as unbounded. It used to clamp <= 0 to 20, so the
// documented value silently produced the default: a full-repo run printed 20
// files and then "… and N more file(s)", while the histogram underneath still
// counted every file. The report and the histogram disagreed, and the
// violations that were counted but not listed could not be located from the
// output at all — which is how three MD031 reports in
// instructions/cli-design-principles.instructions.md went unaccounted for.
func TestMaxFilesZeroDetailsEveryOffendingFile(t *testing.T) {
	dir := t.TempDir()
	// 25 offending files: more than the old hardcoded clamp of 20, so a
	// regression shows up as a missing file rather than a count mismatch.
	const files = 25
	for i := 0; i < files; i++ {
		name := filepath.Join(dir, fmt.Sprintf("f%02d.md", i))
		// A fence with no info string: MD040 fires once per file, so each
		// file contributes exactly one counted violation.
		if err := os.WriteFile(name, []byte("# T\n\n```\nx\n```\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, ".markdownlint.json"),
		[]byte(`{"default":true,"MD013":false,"MD041":false,"MD058":false}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := &cobra.Command{Use: "plaesy", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(registry...)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"validate", "markdown", dir, "--no-baseline", "--max-files", "0"})

	// The run is expected to fail: these files violate MD040. The report is
	// what is under test, so the error is the expected path here.
	report, _ := captureStdout(t, func() error { return root.Execute() })

	var missing []string
	for i := 0; i < files; i++ {
		name := fmt.Sprintf("f%02d.md", i)
		if !strings.Contains(report, name) {
			missing = append(missing, name)
		}
	}
	if len(missing) != 0 {
		t.Errorf("%d of %d offending files absent from a --max-files 0 report: %v\n"+
			"0 must mean all; it must not be clamped to the default of 20",
			len(missing), files, missing)
	}
	if strings.Contains(report, "more file(s)") {
		t.Error("report elided files under --max-files 0; the listing is supposed to be complete")
	}
}
