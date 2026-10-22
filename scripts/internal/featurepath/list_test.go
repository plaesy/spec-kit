package featurepath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These cover `plaesy features` (the bare listing). It is the one command in
// this package that has to be useful with no arguments at all, and the one
// whose output is read by a person rather than by a script, so the report
// format is part of the contract rather than an implementation detail.

// specDir lays down <root>/.plaesy/specs/<name> with the given documents and
// returns the specs directory.
func specDir(t *testing.T, root string, features map[string][]string) string {
	t.Helper()
	dir := filepath.Join(root, ".plaesy", "specs")
	for name, docs := range features {
		for _, d := range docs {
			writeFile(t, filepath.Join(dir, name, d), "content\n")
		}
	}
	return dir
}

func TestListFeaturesOnAProjectWithNoSpecsDirectory(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	chdir(t, root)

	features, err := ListFeatures()
	if err != nil {
		t.Fatalf("a project that has not created a feature yet must not fail: %v", err)
	}
	if len(features) != 0 {
		t.Errorf("features = %v, want empty", features)
	}
}

// Ordering is the whole reason this is a function and not a ReadDir loop: a
// lexicographic sort puts 002 before 010, which reads as if the project went
// backwards.
func TestListFeaturesOrdersByNumberNotName(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	specDir(t, root, map[string][]string{
		"002-second":  {"spec.md"},
		"010-tenth":   {"spec.md"},
		"001-first":   {"spec.md", "plan.md"},
		"3-unpadded":  {"spec.md"},
		"not-a-num":   {"spec.md"},
		"002-second2": {"spec.md"},
	})
	chdir(t, root)

	features, err := ListFeatures()
	if err != nil {
		t.Fatalf("ListFeatures: %v", err)
	}
	var got []string
	for _, f := range features {
		got = append(got, f.Number+"/"+f.Name)
	}
	want := []string{
		"001/001-first",
		"002/002-second",
		"002/002-second2",
		"003/3-unpadded",
		"010/010-tenth",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("order = %v, want %v", got, want)
	}
}

// "007" and "7" name the same feature, so both report the zero-padded number
// CreateNewFeature would have written. A directory whose name has no digits is
// not ours and is not listed at all.
func TestListFeaturesNormalisesTheNumberAndSkipsNonFeatures(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	dir := specDir(t, root, map[string][]string{"7-seventh": {"spec.md"}})
	if err := os.MkdirAll(filepath.Join(dir, "scratch"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "stray.md"), "not a feature\n")
	chdir(t, root)

	features, err := ListFeatures()
	if err != nil {
		t.Fatalf("ListFeatures: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("features = %v, want only 7-seventh (a plain file and a non-numeric directory are skipped)", features)
	}
	if features[0].Number != "007" {
		t.Errorf("Number = %q, want 007", features[0].Number)
	}
}

// Only the documents a feature actually has are listed, in the order /implement
// writes them — not alphabetically, which would put contracts.md before
// data-model.md.
func TestListFeaturesListsOnlyPresentDocumentsInWriteOrder(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	specDir(t, root, map[string][]string{
		"001-first": {"tasks.md", "spec.md", "plan.md", "notes.md"},
	})
	chdir(t, root)

	features, err := ListFeatures()
	if err != nil {
		t.Fatalf("ListFeatures: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("features = %v", features)
	}
	want := "spec.md, plan.md, tasks.md"
	if got := strings.Join(features[0].Documents, ", "); got != want {
		t.Errorf("Documents = %q, want %q", got, want)
	}
}

// The branch with no commits is the state `plaesy features create` leaves
// behind, and it is the first thing anyone lists afterwards.
func TestListFeaturesMarksTheCurrentBranchOnAnUnbornHead(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "001-first")
	t.Setenv(shimFailEnv, "no-head")
	specDir(t, root, map[string][]string{
		"001-first":  {"spec.md"},
		"002-second": {"spec.md"},
	})
	chdir(t, root)

	features, err := ListFeatures()
	if err != nil {
		t.Fatalf("ListFeatures: %v", err)
	}
	for _, f := range features {
		want := f.Name == "001-first"
		if f.Current != want {
			t.Errorf("%s Current = %v, want %v", f.Name, f.Current, want)
		}
	}
}

// A listing is still correct without the branch — only the `*` marker is lost —
// so an undetectable branch must not fail the command.
func TestListFeaturesSurvivesAnUndeterminableBranch(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "001-first")
	t.Setenv(shimFailEnv, "branch")
	specDir(t, root, map[string][]string{"001-first": {"spec.md"}})
	chdir(t, root)

	features, err := ListFeatures()
	if err != nil {
		t.Fatalf("ListFeatures must not fail when the branch is unknown: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("features = %v, want the one feature", features)
	}
	if features[0].Current {
		t.Error("Current = true with no determinable branch")
	}
}

func TestListFeaturesFailsWithoutARepository(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	t.Setenv(shimFailEnv, "root")
	chdir(t, t.TempDir())

	if _, err := ListFeatures(); err == nil {
		t.Error("ListFeatures outside a repository must be an error, not an empty list")
	}
}

// A missing .plaesy/specs/ is an empty project, but a .plaesy/specs that exists
// and cannot be read is a real failure. Collapsing both into "no features" is
// how a permissions problem or a file where a directory belongs turns into a
// project that silently lost its work.
func TestListFeaturesFailsWhenSpecsIsNotADirectory(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	if err := os.MkdirAll(filepath.Join(root, ".plaesy"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, ".plaesy", "specs"), "not a directory\n")
	chdir(t, root)

	features, err := ListFeatures()
	if err == nil {
		t.Fatalf("an unreadable .plaesy/specs must be an error, got %v", features)
	}
	if !strings.Contains(err.Error(), "specs") {
		t.Errorf("error %q does not name the path it could not read", err)
	}
}

// The number is only used to order and to print, so a directory whose leading
// digits overflow int — which no CreateNewFeature call can produce, but a
// hand-made or restored directory can — must still be listed rather than
// aborting the whole command.
func TestListFeaturesListsAFeatureWhoseNumberOverflows(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	specDir(t, root, map[string][]string{
		"99999999999999999999-huge": {"spec.md"},
		"001-first":                 {"spec.md"},
	})
	chdir(t, root)

	features, err := ListFeatures()
	if err != nil {
		t.Fatalf("an out-of-range feature number must not fail the listing: %v", err)
	}
	if len(features) != 2 {
		t.Fatalf("features = %v, want both", features)
	}
	for _, f := range features {
		if f.Name == "99999999999999999999-huge" && f.Number != "99999999999999999999" {
			t.Errorf("Number = %q, want the directory name used verbatim", f.Number)
		}
	}
}

func TestListReportOnAnEmptyProject(t *testing.T) {
	got := ListReport(nil)
	if !strings.Contains(got, "No features found") {
		t.Errorf("an empty listing must say why it is empty, got %q", got)
	}
	if !strings.Contains(got, `plaesy features create "<description>"`) {
		t.Errorf("an empty listing must name the command that fixes it, got %q", got)
	}
}

func TestListReportMarksTheCurrentFeature(t *testing.T) {
	report := ListReport([]Feature{
		{Name: "001-first", Number: "001", Current: true, Documents: []string{"spec.md"}},
		{Name: "002-second", Number: "002", Documents: []string{"spec.md", "plan.md"}},
	})
	for _, want := range []string{"2 feature(s)", "* 001", "  002", "spec.md, plan.md"} {
		if !strings.Contains(report, want) {
			t.Errorf("report does not contain %q:\n%s", want, report)
		}
	}
	// A current feature exists, so the "none is checked out" footnote is noise.
	if strings.Contains(report, "none is checked out") {
		t.Errorf("report claims nothing is checked out while one is:\n%s", report)
	}
}

// An empty feature directory is a real state, and "nothing here yet" is the
// thing a reader needs to see — a blank column reads as a rendering bug.
func TestListReportSaysWhenAFeatureHasNoDocumentsYet(t *testing.T) {
	report := ListReport([]Feature{{Name: "001-first", Number: "001", Current: true}})
	if !strings.Contains(report, "no design documents yet") {
		t.Errorf("report does not flag the empty feature:\n%s", report)
	}
}

func TestListReportExplainsAMissingCurrentMarker(t *testing.T) {
	report := ListReport([]Feature{{Name: "001-first", Number: "001", Documents: []string{"spec.md"}}})
	if !strings.Contains(report, "none is checked out") {
		t.Errorf("report must explain the absent `*` marker:\n%s", report)
	}
}
