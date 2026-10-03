package validate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// assumptionTagPattern matches the `ASSUMED — <default>` label this framework's
// instructions ask an autonomous run to write whenever it applies a
// best-practice default instead of stopping to ask (see
// instructions/plaesy.instructions.md rule 2). Both the em-dash and a plain
// hyphen are accepted, since a hand-typed note may not reach for the em-dash.
var assumptionTagPattern = regexp.MustCompile(`ASSUMED\s*[—-]\s*(.+)`)

// skippedDirs mirrors the directories every other repo-wide scan in this
// package excludes — version control internals and build output are never a
// source of a genuine ASSUMED tag.
var assumptionsSkippedDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
}

// AssumptionTag is one `ASSUMED —` label found in the corpus.
type AssumptionTag struct {
	File string
	Line int
	Text string
}

// AssumptionsResult carries every ASSUMED tag found, and — only when the
// caller opted into a review window — which ones are stale.
type AssumptionsResult struct {
	FilesScanned int
	Tags         []AssumptionTag
	// ReviewDays is the configured window; 0 means the check is purely
	// informational (see Assumptions doc comment) and Stale is always empty.
	ReviewDays int
	Stale      []AssumptionTag
}

// Assumptions scans every Markdown file in repoRoot for `ASSUMED —` tags and
// reports them. This is deliberately an audit, not a gate: on its own it never
// fails a run, because an autonomous agent recording its own assumptions is
// the thing rule 2 asks for, not a defect.
//
// It becomes a gate only when reviewDays > 0 (opt-in, default off — set via
// `.plaesy/state.json`'s `assumptions_review_days`, the same convention as
// `checkpoint_interval`). In that mode, a tag is "stale" when the file that
// carries it has had no commit in over reviewDays — a file-level proxy for
// the tag's age, not a per-line one: git blame on an inline label gives the
// line's age only if nothing else in the file changed around it since, and a
// simple, honestly-described proxy beats a precise-looking one that silently
// breaks on the common case of an unrelated edit to the same file.
func Assumptions(repoRoot string, reviewDays int) (*AssumptionsResult, error) {
	res := &AssumptionsResult{ReviewDays: reviewDays}

	var files []string
	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if assumptionsSkippedDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", repoRoot, err)
	}

	now := time.Now()
	for _, file := range files {
		res.FilesScanned++

		data, err := os.ReadFile(file)
		if err != nil {
			continue // unreadable file: skip rather than fail the whole scan
		}

		var fileTags []AssumptionTag
		for lineNo, line := range strings.Split(string(data), "\n") {
			if m := assumptionTagPattern.FindStringSubmatch(line); m != nil {
				fileTags = append(fileTags, AssumptionTag{
					File: file,
					Line: lineNo + 1,
					Text: strings.TrimSpace(m[1]),
				})
			}
		}
		if len(fileTags) == 0 {
			continue
		}
		res.Tags = append(res.Tags, fileTags...)

		if reviewDays <= 0 {
			continue
		}
		age, ok := fileAgeDays(file, now)
		if ok && age > float64(reviewDays) {
			res.Stale = append(res.Stale, fileTags...)
		}
	}

	return res, nil
}

// fileAgeDays returns how many days have passed since the file's last commit,
// via `git log -1 --format=%ct`. The second return is false when the file is
// not tracked (not yet committed, or git is unavailable) — callers treat that
// as "cannot judge staleness" rather than guessing an age.
func fileAgeDays(file string, now time.Time) (float64, bool) {
	// git needs to run inside the repository to resolve "--", file; the
	// process's own working directory need not be it (e.g. under `go test`,
	// it is the package directory). -C with the file's own directory lets
	// git discover the repo root upward from there.
	cmd := exec.Command("git", "-C", filepath.Dir(file), "log", "-1", "--format=%ct", "--", filepath.Base(file))
	out, err := cmd.Output()
	if err != nil {
		return 0, false
	}
	ts := strings.TrimSpace(string(out))
	if ts == "" {
		return 0, false
	}
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return 0, false
	}
	committed := time.Unix(sec, 0)
	return now.Sub(committed).Hours() / 24, true
}
