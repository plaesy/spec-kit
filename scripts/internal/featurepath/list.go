package featurepath

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// Feature is one entry under .plaesy/specs/.
type Feature struct {
	// Name is the directory name, which is also the branch name: NNN-slug.
	Name string
	// Number is the leading NNN as a string, zero-padded, so it can be
	// printed without re-formatting.
	Number string
	// Dir is the absolute path to the feature's spec directory.
	Dir string
	// Current is true when this is the checked-out branch.
	Current bool
	// Documents names the design documents present in the directory, sorted.
	// A feature with none still lists: an empty branch is a real state, and
	// "nothing here yet" is the thing a reader needs to see.
	Documents []string
}

// designDocuments are the files a feature directory accumulates, in the order
// they are written by /implement. Only the ones present are listed.
var designDocuments = []string{
	"spec.md", "plan.md", "research.md", "data-model.md",
	"quickstart.md", "contracts", "tasks.md",
}

// ListFeatures reports every feature under .plaesy/specs/, ordered by number
// and then by name so the output is stable across runs and across platforms
// (os.ReadDir already sorts by filename, but sorting on the numeric prefix
// first is what makes 002 come after 010 rather than before it).
//
// A missing .plaesy/specs/ is an empty list, not an error: a project that has
// not run `plaesy init` has no features, and that is a fact to report, not a
// failure. Anything else that goes wrong reading the directory is returned.
func ListFeatures() ([]Feature, error) {
	repoRoot, err := common.GetRepoRoot()
	if err != nil {
		return nil, err
	}
	specsDir := filepath.Join(repoRoot, ".plaesy", "specs")

	current := currentBranch()

	// Absent means "no features yet"; present-but-unreadable means the project
	// is broken. Those must not collapse into each other, and on Windows they
	// would: os.ReadDir on a path that exists as a regular file fails with
	// ERROR_PATH_NOT_FOUND, so os.IsNotExist reports true for a corrupted
	// .plaesy/specs just as it does for a missing one. Trusting that check
	// alone reported a project whose spec directory had been replaced by a file
	// as an empty project, and the report then told the user to go create a
	// feature. Stat first so the two cases stay apart.
	if info, statErr := os.Stat(specsDir); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", specsDir, statErr)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("%s exists but is not a directory", specsDir)
	}

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", specsDir, err)
	}

	features := make([]Feature, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		match := leadingNumberRE.FindString(name)
		if match == "" {
			continue // not a NNN-slug directory; not ours to report
		}
		// "007" and "7" name the same feature; the zero-padded form is what
		// CreateNewFeature writes, so that is what is reported. A prefix that
		// is all digits but not a number cannot happen (the regex is [0-9]+),
		// so a failed Atoi here would be a bug, not input — but falling back
		// to the name keeps a listing working rather than aborting over it.
		number := match
		if n, atoiErr := strconv.Atoi(match); atoiErr == nil {
			number = fmt.Sprintf("%03d", n)
		}
		dir := filepath.Join(specsDir, name)
		features = append(features, Feature{
			Name:      name,
			Number:    number,
			Dir:       dir,
			Current:   name == current,
			Documents: presentDocuments(dir),
		})
	}

	sort.Slice(features, func(i, j int) bool {
		if features[i].Number != features[j].Number {
			return features[i].Number < features[j].Number
		}
		return features[i].Name < features[j].Name
	})
	return features, nil
}

func presentDocuments(dir string) []string {
	var found []string
	for _, doc := range designDocuments {
		if _, err := os.Stat(filepath.Join(dir, doc)); err == nil {
			found = append(found, doc)
		}
	}
	return found
}

// currentBranch returns the checked-out branch, or "" when it cannot be
// determined. A feature listing is still correct without it — only the
// `*` marker is lost — so the error is not propagated.
func currentBranch() string {
	branch, err := common.GetCurrentBranch()
	if err != nil {
		return ""
	}
	return branch
}

// ListReport renders the feature list for the terminal, and the empty case as
// a line of its own rather than a blank block, so `plaesy features` on a fresh
// project says why it printed nothing.
func ListReport(features []Feature) string {
	if len(features) == 0 {
		return "No features found. Create one with: plaesy features create \"<description>\"\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d feature(s) in .plaesy/specs:\n\n", len(features))
	for _, f := range features {
		marker := " "
		if f.Current {
			marker = "*"
		}
		if len(f.Documents) == 0 {
			fmt.Fprintf(&b, "%s %-6s %-40s (no design documents yet)\n", marker, f.Number, f.Name)
			continue
		}
		fmt.Fprintf(&b, "%s %-6s %-40s %s\n", marker, f.Number, f.Name, strings.Join(f.Documents, ", "))
	}
	if !anyCurrent(features) {
		fmt.Fprintf(&b, "\n(* marks the checked-out branch; none is checked out, or the current one has no feature directory)\n")
	}
	return b.String()
}

func anyCurrent(features []Feature) bool {
	for _, f := range features {
		if f.Current {
			return true
		}
	}
	return false
}
