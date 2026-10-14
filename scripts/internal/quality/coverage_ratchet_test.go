package quality

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// baseline mirrors scripts/coverage-baseline.json. The $comment key is carried
// as a field so the explanation stays in the data file instead of a comment
// nobody reads, and so the JSON stays self-describing when copied.
type baseline struct {
	Comment    []string           `json:"$comment"`
	MeasuredAt string             `json:"measured_at"`
	Packages   map[string]float64 `json:"packages"`
}

// loadBaseline reads the recorded per-package coverage.
func loadBaseline(t *testing.T) baseline {
	t.Helper()
	path := filepath.Join(moduleDir(t), BaselineName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v\n\nWithout a recorded baseline there is no ratchet, and "+
			"a bar that only exists in prose is a wish. Measure with "+
			"'go test ./... -cover -count=1' in scripts/ and record every package.", path, err)
	}
	var b baseline
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return b
}

// moduleDir walks up from the package directory to the go.mod that declares
// this module. Unlike the Markdown ratchet this cannot skip when the file is
// absent: without a module the coverage check has nothing to measure, and a
// check that cannot run must not report as passing.
func moduleDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
			if strings.Contains(string(data), "module "+ModulePath) {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("no go.mod for this module above %s", t.Name())
	return ""
}

// TestCoverageBaselineIsWellFormed guards the guard: a typo in the JSON, a
// negative percentage or a value over 100 must fail as a malformed baseline
// rather than as a coverage regression nobody can act on.
func TestCoverageBaselineIsWellFormed(t *testing.T) {
	b := loadBaseline(t)
	if len(b.Packages) == 0 {
		t.Fatalf("%s lists no packages: the ratchet would pass on an empty map", BaselineName)
	}
	if b.MeasuredAt == "" {
		t.Errorf("%s has no measured_at date: a number with no date cannot be re-measured on purpose", BaselineName)
	}
	for pkg, pct := range b.Packages {
		if strings.TrimSpace(pkg) == "" {
			t.Errorf("%s has an empty package name", BaselineName)
		}
		if math.IsNaN(pct) || pct < 0 || pct > 100 {
			t.Errorf("%s: %s is %v%%, not a coverage percentage", BaselineName, pkg, pct)
		}
	}
}

// updateBaseline re-measures every package and rewrites the baseline file, so
// the fix for an outgrown entry is one command instead of a shell loop copied
// out of a comment. It is a flag rather than a second program because the
// measurement already exists here, and a second implementation of it would be a
// second thing to keep honest.
var updateBaseline = flag.Bool("update-coverage-baseline", false,
	"re-measure every package and rewrite "+BaselineName+" with the result")

// TestCoverageDoesNotRegress is the CI gate. It measures every package in the
// module on its own (one `go test -cover -count=1` per package) and compares
// the result with the recorded baseline. It fails on four distinct conditions,
// each with its own fix:
//
//   - a package below its recorded number — add tests, or accept a lower bar
//   - a package with no entry — a new package is an unexamined default
//   - an entry for a package that no longer exists — a claim about gone code
//   - an entry the tree has outgrown by more than outgrowthTolerance — the
//     file is stale, and a too-low bar is the one error nothing else catches
//
// Raising a number is the point, so an improvement within the tolerance is
// logged and never fails. Pass -update-coverage-baseline to re-record every
// number after a deliberate improvement.
func TestCoverageDoesNotRegress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping the coverage ratchet in -short mode: it runs the suite again")
	}
	// The baseline is one number per package, with no OS dimension, but
	// several packages (internal/cleaner, internal/common, internal/installer,
	// internal/graph, internal/mdlint, at least) have a genuine
	// `if runtime.GOOS == "windows"` fork, so the statement actually reachable
	// differs by platform -- not by test quality. A baseline tight enough to
	// catch a real regression on one OS is, by construction, either a
	// regression or "recorded far too low" on another: this test's own
	// decayed-baseline guard (coverage_ratchet_test.go below) makes a single
	// number that satisfies every OS at once impossible for those packages.
	// ubuntu-latest is canonical here for the same reason the "corpus" CI job
	// already runs on ubuntu-latest only rather than across the matrix: one
	// designated platform, not a race to the lowest common denominator.
	if runtime.GOOS != "linux" {
		t.Skip("the coverage ratchet is measured on linux (ubuntu-latest in CI) only: several packages fork on runtime.GOOS, so a Windows or macOS run would compare against numbers that platform cannot reach")
	}
	b := loadBaseline(t)
	dir := moduleDir(t)

	// The baseline is keyed by module-relative path because that is what a
	// human reads in a diff; `go list` hands back import paths. Compare in
	// import-path space, report in path space.
	want := make(map[string]float64, len(b.Packages))
	for path, pct := range b.Packages {
		want[importPath(path)] = pct
	}

	pkgs := allPackages(t, dir)
	measured := measureCoverage(t, dir, pkgs)

	if len(measured) == 0 {
		t.Fatalf("measured zero packages in %s: the runner is broken, not the tree untested", dir)
	}

	if *updateBaseline {
		writeBaseline(t, filepath.Join(dir, BaselineName), b, measured)
		t.Skipf("rewrote %s from a fresh measurement of %d package(s)", BaselineName, len(measured))
	}

	// A tree whose package names share nothing with the baseline is a copy of
	// the module, not this repository. Enforcing our numbers on someone else's
	// packages would be noise, so say what happened instead of guessing. Both
	// key styles are quoted because a mismatch between them looks exactly like
	// this and must not hide behind the skip.
	var overlap int
	for pkg := range measured {
		if _, ok := want[pkg]; ok {
			overlap++
		}
	}
	if overlap == 0 {
		t.Skipf("%s describes a different tree: %d package(s) measured, none listed "+
			"(measured e.g. %q, baseline e.g. %q); delete it if this copy will never "+
			"grow these packages", BaselineName, len(measured), firstKey(measured), firstKey(want))
	}

	drift := compareCoverage(measured, want)

	for _, line := range drift.improved {
		t.Logf("coverage improved: %s — record it", line)
	}

	if len(drift.regressed) > 0 {
		t.Errorf("coverage ratchet broken in %d package(s):\n%s\n\n"+
			"Add tests, or accept the lower number explicitly by editing %s. "+
			"Do not raise the entry to silence this: the bar moves when the tests do.",
			len(drift.regressed), strings.Join(drift.regressed, "\n"), BaselineName)
	}
	if len(drift.unlisted) > 0 {
		t.Errorf("package(s) with no coverage entry in %s:\n%s\n\n"+
			"A new package is an unexamined default. Record the measured number, then decide "+
			"whether the 0.0%% of packages exercised only through the CLI should stay that way.",
			BaselineName, strings.Join(drift.unlisted, "\n"))
	}
	if len(drift.stale) > 0 {
		t.Errorf("stale coverage entries in %s — these packages have no tests or do not exist:\n%s\n\n"+
			"Remove the entries, or add the tests the entry is claiming to measure.",
			BaselineName, strings.Join(drift.stale, "\n"))
	}
	if len(drift.outgrown) > 0 {
		t.Errorf("recorded coverage in %s is more than %.1f points behind the tree:\n%s\n\n"+
			"The bar is still \"never lower\" — nothing here asks for less coverage. This fires because the "+
			"file is supposed to be a measurement, and a number the tree has outgrown is a claim about the code "+
			"as it was: a too-low entry is indistinguishable from a correct one, so nothing else would ever "+
			"notice it. Re-measure with:\n\n"+
			"    go test ./internal/quality -run TestCoverageDoesNotRegress -update-coverage-baseline\n\n"+
			"Review the diff before committing it: the new numbers are the record, not a target.",
			BaselineName, outgrowthTolerance, strings.Join(drift.outgrown, "\n"))
	}

	t.Logf("measured %d package(s) against %d baseline entries", len(measured), len(b.Packages))
}

// baselineFile is the on-disk shape used when writing. The numbers are
// json.Number so they keep the one decimal place the file has always used:
// encoding a float64 0 produces "0" and 98 produces "98", which turns the
// "0.0%" entries — the visible marker of the packages with no tests at all —
// into "0" and makes a diff of that file harder to read than it needs to be.
type baselineFile struct {
	Comment    []string               `json:"$comment"`
	MeasuredAt string                 `json:"measured_at"`
	Packages   map[string]json.Number `json:"packages"`
}

// writeBaseline rewrites the baseline file from a fresh measurement, keeping the
// $comment block: the reasoning for measuring one package at a time is the part
// a future editor needs most, and a rewrite that dropped it would lose the
// reason the file looks the way it does.
func writeBaseline(t *testing.T, path string, old baseline, measured map[string]float64) {
	t.Helper()
	next := baselineFile{
		Comment:    old.Comment,
		MeasuredAt: time.Now().Format("2006-01-02"),
		Packages:   make(map[string]json.Number, len(measured)),
	}
	for pkg, pct := range measured {
		next.Packages[relPath(pkg)] = json.Number(strconv.FormatFloat(math.Round(pct*10)/10, 'f', 1, 64))
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		t.Fatalf("encode %s: %v", BaselineName, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	t.Logf("recorded %d package(s) in %s", len(next.Packages), BaselineName)
}

// drift is the difference between what the tree measures and what the baseline
// claims, split by cause. The causes are separate because the fix is different
// for each: a falling number needs tests or a lower bar, an unlisted package
// needs a decision, a stale entry needs a deletion, and a number the tree has
// outgrown needs re-measuring.
type drift struct {
	regressed []string // measured below the recorded number
	unlisted  []string // measured, but absent from the baseline
	stale     []string // recorded, but not measured
	improved  []string // measured above the recorded number, within the tolerance
	outgrown  []string // measured far above the recorded number: the file is stale
}

// epsilon absorbs the one-decimal rounding `go test` prints. A difference
// smaller than that is not a change anyone can act on, and a gate that trips on
// it gets switched off.
const epsilon = 0.05

// outgrowthTolerance bounds how far the tree may run ahead of the recorded
// number before the file counts as stale rather than merely behind.
//
// Failing on every improvement would punish the thing the ratchet exists to
// encourage, and the file would be rewritten by every test that lands. But
// logging improvement only is what let "cmd/plaesy 22.9%" sit in the baseline
// while the tree measured 26.6% — reproducible at HEAD, three days after the
// number was written down. Nothing failed, because a bar that is too low is
// indistinguishable from a bar that is right, and nothing in the build ever
// re-read the number. Two points of drift is where "the file is a measurement"
// stops being true; past that the entry is a claim about code as it was, and
// the only way to tell is to measure again.
const outgrowthTolerance = 2.0

func compareCoverage(measured, want map[string]float64) drift {
	var d drift
	for _, pkg := range sortedKeys(measured) {
		got := measured[pkg]
		recorded, ok := want[pkg]
		if !ok {
			d.unlisted = append(d.unlisted, fmt.Sprintf("  %-26s %5.1f%%  (no entry in %s)", relPath(pkg), got, BaselineName))
			continue
		}
		switch {
		case got+epsilon < recorded:
			d.regressed = append(d.regressed, fmt.Sprintf("  %-26s %5.1f%%  (was %5.1f%%)", relPath(pkg), got, recorded))
		case got-recorded > outgrowthTolerance:
			d.outgrown = append(d.outgrown, fmt.Sprintf("  %-26s %5.1f%%  (recorded %5.1f%%)", relPath(pkg), got, recorded))
		case got > recorded+epsilon:
			d.improved = append(d.improved, fmt.Sprintf("%s %5.1f%% (was %5.1f%%)", relPath(pkg), got, recorded))
		}
	}
	for _, pkg := range sortedKeys(want) {
		if _, ok := measured[pkg]; !ok {
			d.stale = append(d.stale, "  "+relPath(pkg))
		}
	}
	return d
}

// TestCompareCoverageNamesEveryWayTheBarCanBeBroken pins the rules of the gate
// without measuring the tree, so each failure class is proven on every run
// instead of being discovered the day a package quietly loses its tests.
func TestCompareCoverageNamesEveryWayTheBarCanBeBroken(t *testing.T) {
	const pkg = "github.com/plaesy/spec-kit/internal/mdlint"

	t.Run("unchanged is not a finding", func(t *testing.T) {
		d := compareCoverage(map[string]float64{pkg: 86.9}, map[string]float64{pkg: 86.9})
		if d.regressed != nil || d.unlisted != nil || d.stale != nil || d.improved != nil {
			t.Fatalf("equal coverage produced %+v", d)
		}
	})

	t.Run("rounding noise is not a regression", func(t *testing.T) {
		d := compareCoverage(map[string]float64{pkg: 86.94}, map[string]float64{pkg: 86.9})
		if len(d.regressed) != 0 {
			t.Fatalf("0.04pp of rounding tripped the gate: %v", d.regressed)
		}
	})

	t.Run("a fall is a regression", func(t *testing.T) {
		d := compareCoverage(map[string]float64{pkg: 80.1}, map[string]float64{pkg: 86.9})
		if len(d.regressed) != 1 {
			t.Fatalf("want 1 regression, got %+v", d)
		}
		if !strings.Contains(d.regressed[0], "80.1") || !strings.Contains(d.regressed[0], "86.9") {
			t.Errorf("the finding must name both numbers, got %q", d.regressed[0])
		}
		if !strings.Contains(d.regressed[0], "internal/mdlint") {
			t.Errorf("the finding must name the package, got %q", d.regressed[0])
		}
	})

	t.Run("a small rise is reported but never fails", func(t *testing.T) {
		d := compareCoverage(map[string]float64{pkg: 88.0}, map[string]float64{pkg: 86.9})
		if len(d.regressed) != 0 || len(d.outgrown) != 0 {
			t.Fatalf("improvement must not fail the gate: %+v", d)
		}
		if len(d.improved) != 1 {
			t.Fatalf("improvement must be surfaced so it gets recorded, got %+v", d)
		}
	})

	// The class of failure the log-only rule could not see: a number the tree
	// has outgrown. "cmd/plaesy 22.9%" sat in the baseline while the tree
	// measured 26.6% — a fact reproducible at HEAD, days after the number was
	// written — and nothing failed, because improvement was only ever logged.
	t.Run("a large rise means the recorded number is stale, not merely behind", func(t *testing.T) {
		d := compareCoverage(map[string]float64{pkg: 26.6}, map[string]float64{pkg: 22.9})
		if len(d.regressed) != 0 {
			t.Fatalf("a rise must never be reported as a regression: %v", d.regressed)
		}
		if len(d.outgrown) != 1 {
			t.Fatalf("3.7 points of drift must be reported as an outgrown entry, got %+v", d)
		}
		if !strings.Contains(d.outgrown[0], "26.6") || !strings.Contains(d.outgrown[0], "22.9") {
			t.Errorf("the finding must name both numbers, got %q", d.outgrown[0])
		}
		if !strings.Contains(d.outgrown[0], "cmd/plaesy") && !strings.Contains(d.outgrown[0], "mdlint") {
			t.Errorf("the finding must name the package, got %q", d.outgrown[0])
		}
		if d.improved != nil {
			t.Errorf("an outgrown entry is not also a plain improvement: %v", d.improved)
		}
	})

	t.Run("a rise exactly at the tolerance is still an improvement", func(t *testing.T) {
		// The boundary is inclusive: outgrowth is "more than" the tolerance, so
		// a package sitting exactly on it is not a stale file.
		d := compareCoverage(map[string]float64{pkg: 88.9}, map[string]float64{pkg: 86.9})
		if len(d.outgrown) != 0 {
			t.Fatalf("%.1f points of drift is not more than the tolerance: %v", 88.9-86.9, d.outgrown)
		}
		if len(d.improved) != 1 {
			t.Fatalf("it is still reported as an improvement to record, got %+v", d)
		}
	})

	t.Run("an unlisted package is unexamined debt", func(t *testing.T) {
		fresh := ModulePath + "/internal/brandnew"
		d := compareCoverage(map[string]float64{fresh: 0}, map[string]float64{pkg: 86.9})
		if len(d.unlisted) != 1 {
			t.Fatalf("want 1 unlisted package, got %+v", d)
		}
		if !strings.Contains(d.unlisted[0], "internal/brandnew") {
			t.Errorf("the finding must name the new package, got %q", d.unlisted[0])
		}
	})

	t.Run("an entry for a package that is gone is stale", func(t *testing.T) {
		gone := ModulePath + "/internal/removed"
		d := compareCoverage(map[string]float64{pkg: 86.9}, map[string]float64{pkg: 86.9, gone: 12.3})
		if len(d.stale) != 1 {
			t.Fatalf("want 1 stale entry, got %+v", d)
		}
		if !strings.Contains(d.stale[0], "internal/removed") {
			t.Errorf("the finding must name the stale package, got %q", d.stale[0])
		}
	})

	t.Run("findings are sorted so a diff of two runs reads the same", func(t *testing.T) {
		measured := map[string]float64{
			ModulePath + "/internal/zeta":  1,
			ModulePath + "/internal/alpha": 2,
			ModulePath + "/internal/gamma": 3,
		}
		d := compareCoverage(measured, map[string]float64{
			ModulePath + "/internal/alpha": 2, // unchanged, so it is not a finding at all
			ModulePath + "/internal/beta":  50,
		})
		if len(d.unlisted) != 2 || len(d.stale) != 1 {
			t.Fatalf("want 2 unlisted and 1 stale, got %+v", d)
		}
		if !strings.Contains(d.unlisted[0], "gamma") || !strings.Contains(d.unlisted[1], "zeta") {
			t.Errorf("unlisted findings must be sorted, got %v", d.unlisted)
		}
	})
}

// TestBaselineKeyStylesAreEquivalent keeps the forgiving key handling honest:
// the baseline may be written either way, and both must reach the same verdict.
func TestBaselineKeyStylesAreEquivalent(t *testing.T) {
	rel := "internal/mdlint"
	full := ModulePath + "/" + rel
	if importPath(rel) != full || importPath(full) != full {
		t.Fatalf("importPath: %q -> %q, %q -> %q", rel, importPath(rel), full, importPath(full))
	}
	if relPath(full) != rel {
		t.Fatalf("relPath(%q) = %q", full, relPath(full))
	}
	// The round trip is what a hand-edit of the JSON depends on.
	if importPath(relPath(full)) != full {
		t.Fatalf("round trip lost the package path")
	}
}

// allPackages lists the module's packages, minus this one. Packages without
// test files are included deliberately: they measure 0.0%, and that 0.0 is the
// debt this repository wants to see. Excluding them would make a large new
// untested package invisible to the gate, which is the silent-debt failure it
// exists to prevent.
func allPackages(t *testing.T, dir string) []string {
	t.Helper()
	out, err := runGo(t, dir, "list", "./...")
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	var pkgs []string
	for _, line := range strings.Split(out, "\n") {
		pkg := strings.TrimSpace(line)
		if pkg == "" || pkg == SelfPackage {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	return pkgs
}

// measureCoverage runs each package's own tests with -cover and returns the
// statement coverage go reported for it.
//
// It shells out rather than inspecting a merged profile because a merged
// profile attributes a package's coverage to whichever test binary touched it,
// which is how a package that nothing tests ends up credited with whatever
// else happened to run.
func measureCoverage(t *testing.T, dir string, pkgs []string) map[string]float64 {
	t.Helper()
	if len(pkgs) == 0 {
		return map[string]float64{}
	}
	results := make([]float64, len(pkgs))
	failures := make([]string, len(pkgs))

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > len(pkgs) {
		workers = len(pkgs)
	}
	var wg sync.WaitGroup
	next := make(chan int)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				out, err := runGo(t, dir, "test", "-cover", "-count=1", pkgs[i])
				if err != nil {
					// A package whose tests fail has no coverage to compare. The
					// failure is the finding; the ratchet must not add a
					// misleading second one.
					failures[i] = fmt.Sprintf("  %s: go test failed: %v\n%s", pkgs[i], err, indent(out))
					continue
				}
				pct, ok := parseCoverage(out)
				if !ok {
					failures[i] = fmt.Sprintf("  %s: no coverage figure in output:\n%s", pkgs[i], indent(out))
				}
				results[i] = pct
			}
		}()
	}
	for i := range pkgs {
		next <- i
	}
	close(next)
	wg.Wait()

	if problems := nonEmpty(failures); len(problems) > 0 {
		t.Fatalf("coverage could not be measured for %d of %d package(s):\n%s\n\n"+
			"A ratchet that skips what it cannot see is the silent-debt failure this check "+
			"exists to prevent.", len(problems), len(pkgs), strings.Join(problems, "\n"))
	}

	measured := make(map[string]float64, len(pkgs))
	for i, pkg := range pkgs {
		measured[pkg] = results[i]
	}
	return measured
}

// parseCoverage pulls "coverage: 12.3% of statements" out of `go test` output.
// A package with no test files is reported as 0.0, which is what the toolchain
// means and what the baseline records.
func parseCoverage(out string) (float64, bool) {
	for _, line := range strings.Split(out, "\n") {
		i := strings.Index(line, "coverage: ")
		if i < 0 {
			continue
		}
		rest := line[i+len("coverage: "):]
		end := strings.Index(rest, "%")
		if end < 0 {
			continue
		}
		var pct float64
		if _, err := fmt.Sscanf(rest[:end], "%g", &pct); err != nil {
			continue
		}
		return pct, true
	}
	if strings.Contains(out, "[no test files]") {
		return 0, true
	}
	return 0, false
}

func runGo(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), measureTimeout*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, goTool(), args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("`go %s` exceeded %d minutes", strings.Join(args, " "), measureTimeout)
	}
	return string(out), err
}

func goTool() string {
	if goroot := os.Getenv("GOROOT"); goroot != "" {
		candidate := filepath.Join(goroot, "bin", "go")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "go"
}

func indent(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, line := range lines {
		lines[i] = "    " + line
	}
	return strings.Join(lines, "\n")
}

// importPath accepts either spelling a baseline entry might use — the
// module-relative path a human writes, or a full import path pasted from `go
// list` — and returns the import path.
func importPath(path string) string {
	if strings.HasPrefix(path, ModulePath+"/") {
		return path
	}
	return ModulePath + "/" + path
}

// relPath is the inverse, for messages that a human reads in a diff.
func relPath(pkg string) string {
	return strings.TrimPrefix(pkg, ModulePath+"/")
}

func firstKey(m map[string]float64) string {
	keys := sortedKeys(m)
	if len(keys) == 0 {
		return ""
	}
	return keys[0]
}

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// nonEmpty drops the unused slots of a per-package result slice, so a run with
// no failures joins to nothing instead of to a screen of blank lines.
func nonEmpty(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
