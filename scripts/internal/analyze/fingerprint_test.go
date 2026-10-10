package analyze

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// setMtimes pins the mtime of every file in root whose relative path appears
// in the times map, so fingerprint strings can be asserted exactly.
func setMtimes(t *testing.T, root string, times map[string]time.Time) {
	t.Helper()
	for rel, ts := range times {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.Chtimes(p, ts, ts); err != nil {
			t.Fatalf("chtimes %s: %v", rel, err)
		}
	}
}

func TestIsExcludedFingerprintDirName(t *testing.T) {
	cases := []struct {
		name string
		dir  string
		want bool
	}{
		{"dot directory", ".git", true},
		{"plaesy output dir is excluded as a dot dir", ".plaesy", true},
		{"node_modules", "node_modules", true},
		{"ordinary directory", "lib", false},
		{"build is not excluded by the fingerprint walk", "build", false},
		{"empty name", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isExcludedFingerprintDirName(tc.dir); got != tc.want {
				t.Errorf("isExcludedFingerprintDirName(%q) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}
}

func TestComputeFingerprint(t *testing.T) {
	base := time.Unix(1_700_000_000, 0).UTC()
	older := base.Add(-2 * time.Hour)
	newest := base.Add(time.Hour)

	root := newFixture(t, map[string]string{
		"main.go":                     "package main",
		"notes.md":                    "# notes",
		"sub/impl.py":                 "print(1)",
		"config/settings.yaml":        "a: 1",
		"skipme.txt":                  "documentation only",
		"skipme.lock":                 "lockfile",
		"skipme.png":                  "binary",
		".git/objects/blob.go":        "ignored",
		".plaesy/analysis/project.js": "generated",
		"node_modules/dep/index.js":   "vendored",
	})
	setMtimes(t, root, map[string]time.Time{
		"main.go":              older,
		"notes.md":             base,
		"sub/impl.py":          newest,
		"config/settings.yaml": base,
		"skipme.txt":           newest,
		"skipme.lock":          newest,
		"skipme.png":           newest,
	})

	// Counted: main.go, notes.md, sub/impl.py, config/settings.yaml, plus
	// skipme.txt and skipme.lock — .txt is documentation the analysis counts and
	// .lock is a build input, so both are tracked. Newest counted mtime is the
	// one on sub/impl.py. Excluded: the binary, the dotfile dir, .plaesy's own
	// output and node_modules.
	want := "6|" + strconv.FormatInt(newest.Unix(), 10) + "|9.9.9"
	if got := computeFingerprint(root, "9.9.9"); got != want {
		t.Errorf("computeFingerprint() = %q, want %q", got, want)
	}
}

func TestComputeFingerprintCountsDanglingSymlink(t *testing.T) {
	root := newFixture(t, map[string]string{"main.go": "package main"})

	// A dangling symlink is a directory entry with a counted extension, and
	// filepath.WalkDir reports the link's own metadata without following it.
	// That matches `find`, which does not follow symlinks either, so the
	// fingerprint must count the link rather than its missing target.
	if err := os.Symlink(filepath.Join(root, "does-not-exist"), filepath.Join(root, "broken.go")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	got := computeFingerprint(root, "1.0.0")
	if !strings.HasPrefix(got, "2|") {
		t.Errorf("computeFingerprint() = %q, want the dangling link counted (2 files)", got)
	}
	if !strings.HasSuffix(got, "|1.0.0") {
		t.Errorf("computeFingerprint() = %q, want the version appended", got)
	}
}

func TestComputeFingerprintFormat(t *testing.T) {
	cases := []struct {
		name            string
		files           map[string]string
		frameworkVerion string
		want            string
	}{
		{"empty project", map[string]string{}, "1.2.3", "0|0|1.2.3"},
		{"files no bucket claims", map[string]string{"image.png": "x", "logo.svg": "x", "data.bin": "x"}, "0.0.1", "0|0|0.0.1"},
		{"version is appended verbatim including non-semver", map[string]string{}, "dev", "0|0|dev"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newFixture(t, tc.files)
			if got := computeFingerprint(root, tc.frameworkVerion); got != tc.want {
				t.Errorf("computeFingerprint() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestComputeFingerprintCountsCaseInsensitiveExtensions(t *testing.T) {
	root := newFixture(t, map[string]string{"README.MD": "# hi"})
	got := computeFingerprint(root, "1.0.0")
	if got[:1] != "1" {
		t.Errorf("expected the uppercase .MD extension to be counted, got %q", got)
	}
}

func TestComputeFingerprintMissingRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	if got := computeFingerprint(missing, "1.0.0"); got != "0|0|1.0.0" {
		t.Errorf("computeFingerprint(missing) = %q, want %q", got, "0|0|1.0.0")
	}
}

// TestComputeFingerprintTracksDocumentationEdits is the regression guard for the
// drift between the file-type taxonomy and the fingerprint. countFileTypes
// classifies .txt/.rst/.adoc as documentation, but the fingerprint used to
// ignore them, so a documentation-only edit left overview.md and project.json
// describing the previous state of the docs and reported nothing.
func TestComputeFingerprintTracksDocumentationEdits(t *testing.T) {
	// The fingerprint is count|maxmtime|version, so the mtimes are pinned:
	// a rewrite inside the same second would leave it unchanged and the test
	// would pass for the wrong reason.
	base := time.Unix(1_700_000_000, 0).UTC()
	root := newFixture(t, map[string]string{"main.go": "package main", "notes.txt": "v1"})
	setMtimes(t, root, map[string]time.Time{"main.go": base, "notes.txt": base})
	before := computeFingerprint(root, "1.0.0")
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("completely rewritten"), 0o644); err != nil {
		t.Fatalf("rewrite notes.txt: %v", err)
	}
	setMtimes(t, root, map[string]time.Time{"notes.txt": base.Add(time.Hour)})
	if after := computeFingerprint(root, "1.0.0"); after == before {
		t.Errorf("a .txt-only edit must change the fingerprint, got %q both times", before)
	}
	after := computeFingerprint(root, "1.0.0")
	if err := os.WriteFile(filepath.Join(root, "extra.txt"), []byte("new"), 0o644); err != nil {
		t.Fatalf("write extra.txt: %v", err)
	}
	setMtimes(t, root, map[string]time.Time{"extra.txt": base.Add(2 * time.Hour)})
	if got := computeFingerprint(root, "1.0.0"); got == after {
		t.Errorf("adding a .txt file must change the fingerprint, got %q both times", after)
	}
}

// TestFingerprintTracksEveryExtensionTheAnalysisCounts ties the two lists
// together structurally: if a future bucket gains an extension, this fails until
// the fingerprint knows about it, which is the whole point of building both from
// one list per bucket.
func TestFingerprintTracksEveryExtensionTheAnalysisCounts(t *testing.T) {
	for _, group := range [][]string{sourceCodeExts, docExts, cfgExts} {
		for _, ext := range group {
			if !fingerprintExts[ext] {
				t.Errorf("extension %q is classified by the analysis but missing from the fingerprint; "+
					"a change to that file would skip regeneration", ext)
			}
		}
	}
	// The buckets are keyed by bare extension ("go", not ".go"), so the set the
	// walk looks up matches filepath.Ext minus its dot.
	for _, tc := range []struct {
		ext  string
		want bool
	}{
		{"go", true}, {"rst", true}, {"toml", true},
		{"sql", true}, {"scss", true},
		{"png", false}, {"zip", false}, {"bin", false},
	} {
		tracked := fingerprintExts[tc.ext]
		if tracked != tc.want {
			t.Errorf("fingerprintExts[%q] = %v, want %v", tc.ext, tracked, tc.want)
		}
	}
}

func TestShouldSkipAnalysis(t *testing.T) {
	project := newFixture(t, map[string]string{"main.go": "package main"})
	analysisDir := filepath.Join(project, ".plaesy", "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("mkdir analysis dir: %v", err)
	}
	fpFile := filepath.Join(analysisDir, ".analysis-fingerprint")

	if shouldSkipAnalysis(project, analysisDir, "1.0.0") {
		t.Fatal("first run must not skip")
	}
	if _, err := os.Stat(fpFile); !os.IsNotExist(err) {
		t.Errorf("shouldSkipAnalysis must not record anything, found %s (err=%v)", fpFile, err)
	}
	if err := recordFingerprint(project, analysisDir, "1.0.0"); err != nil {
		t.Fatalf("recordFingerprint: %v", err)
	}
	stored := readFileString(t, fpFile)
	if stored != computeFingerprint(project, "1.0.0") {
		t.Errorf("stored fingerprint = %q, want %q", stored, computeFingerprint(project, "1.0.0"))
	}
	if !shouldSkipAnalysis(project, analysisDir, "1.0.0") {
		t.Error("second run with an unchanged project must skip")
	}

	// A framework version bump alone invalidates the cache.
	if shouldSkipAnalysis(project, analysisDir, "2.0.0") {
		t.Error("a framework version change must invalidate the cache")
	}
	if err := recordFingerprint(project, analysisDir, "2.0.0"); err != nil {
		t.Fatalf("recordFingerprint after version bump: %v", err)
	}
	if !shouldSkipAnalysis(project, analysisDir, "2.0.0") {
		t.Error("the run after a version bump should skip again")
	}
	if got, want := readFileString(t, fpFile), computeFingerprint(project, "2.0.0"); got != want {
		t.Errorf("fingerprint file not refreshed with the new version: %q, want %q", got, want)
	}
}

func TestShouldSkipAnalysisDetectsNewAndRemovedFiles(t *testing.T) {
	project := newFixture(t, map[string]string{"main.go": "package main"})
	analysisDir := filepath.Join(project, ".plaesy", "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("mkdir analysis dir: %v", err)
	}
	if shouldSkipAnalysis(project, analysisDir, "1.0.0") {
		t.Fatal("first run must not skip")
	}
	if err := recordFingerprint(project, analysisDir, "1.0.0"); err != nil {
		t.Fatalf("recordFingerprint: %v", err)
	}
	if !shouldSkipAnalysis(project, analysisDir, "1.0.0") {
		t.Fatal("unchanged project must skip")
	}

	if err := os.WriteFile(filepath.Join(project, "added.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write added.go: %v", err)
	}
	if shouldSkipAnalysis(project, analysisDir, "1.0.0") {
		t.Error("adding a tracked file must invalidate the cache")
	}
	if err := recordFingerprint(project, analysisDir, "1.0.0"); err != nil {
		t.Fatalf("recordFingerprint after add: %v", err)
	}
	if !shouldSkipAnalysis(project, analysisDir, "1.0.0") {
		t.Error("cache should be valid again after the regeneration run")
	}

	if err := os.Remove(filepath.Join(project, "main.go")); err != nil {
		t.Fatalf("remove main.go: %v", err)
	}
	if shouldSkipAnalysis(project, analysisDir, "1.0.0") {
		t.Error("removing a tracked file must invalidate the cache")
	}
}

func TestShouldSkipAnalysisTreatsCorruptFingerprintAsChanged(t *testing.T) {
	project := newFixture(t, map[string]string{"main.go": "package main"})
	analysisDir := filepath.Join(project, ".plaesy", "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("mkdir analysis dir: %v", err)
	}
	fpFile := filepath.Join(analysisDir, ".analysis-fingerprint")
	if err := os.WriteFile(fpFile, []byte("garbage"), 0o644); err != nil {
		t.Fatalf("seed fingerprint: %v", err)
	}
	if shouldSkipAnalysis(project, analysisDir, "1.0.0") {
		t.Error("a corrupt fingerprint must not cause a skip")
	}
	// The corrupt file is left as it is until a run actually finishes and
	// records the new value; a read-only check has no business rewriting it.
	if got := readFileString(t, fpFile); got != "garbage" {
		t.Errorf("shouldSkipAnalysis rewrote the fingerprint file to %q", got)
	}
	if err := recordFingerprint(project, analysisDir, "1.0.0"); err != nil {
		t.Fatalf("recordFingerprint: %v", err)
	}
	if got := readFileString(t, fpFile); got != computeFingerprint(project, "1.0.0") {
		t.Errorf("fingerprint not repaired, got %q", got)
	}
}
