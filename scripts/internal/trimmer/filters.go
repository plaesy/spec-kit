package trimmer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Filter restructures one command's output before the generic Layer 1
// compression runs. It returns an error when the input is not what the filter
// understands, and the caller then falls back to the raw lines.
type Filter func(lines []string) ([]string, error)

// warnWriter receives fail-open warnings. Tests swap it; production uses stderr.
var warnWriter io.Writer = os.Stderr

// detectFilter picks a filter from the command name, using the first word and,
// for git, the subcommand. An unknown command returns ("", nil).
func detectFilter(args []string) (string, Filter) {
	if len(args) == 0 {
		return "", nil
	}
	switch filepath.Base(strings.TrimSuffix(args[0], ".exe")) {
	case "git":
		if len(args) < 2 {
			return "", nil
		}
		switch args[1] {
		case "diff":
			return "git-diff", gitDiffFilter
		case "status":
			return "git-status", gitStatusFilter
		case "log":
			return "git-log", gitLogFilter
		}
	case "grep", "rg":
		return "grep", grepFilter
	case "ls", "find", "tree":
		return "listing", listingFilter
	}
	return "", nil
}

// applyFilter runs f on lines and fails open: a returned error or a panic keeps
// the raw lines and writes one warning to warnWriter. An empty result is never
// returned in place of a non-empty input.
func applyFilter(name string, f Filter, lines []string) (out []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(warnWriter, "plaesy trim: filter %s panicked (%v); passing raw output through\n", name, r)
			out = lines
		}
	}()
	filtered, err := f(lines)
	if err != nil {
		fmt.Fprintf(warnWriter, "plaesy trim: filter %s failed (%v); passing raw output through\n", name, err)
		return lines
	}
	return filtered
}

// prepareLines applies the filter that matches args, if any, and returns the
// lines to compress along with the stats layer to record them under. Filtered
// commands record under "layer1:<filter>", which is what `trim report` groups
// by. Unknown commands keep the plain "layer1" layer.
func prepareLines(args []string, raw []string) ([]string, string) {
	name, f := detectFilter(args)
	if f == nil {
		return raw, "layer1"
	}
	return applyFilter(name, f, raw), "layer1:" + name
}

// errNotRecognised is returned when a filter gets input with none of the
// structure it parses, so the raw output is kept instead.
var errNotRecognised = errors.New("input not recognised")

// gitDiffFilter keeps file headers, hunk headers and every +/- change line, and
// drops the index lines and the unchanged context lines. A non-empty input with
// no "diff --git" header is not a diff, so it fails open.
func gitDiffFilter(lines []string) ([]string, error) {
	if len(lines) > 0 && !containsPrefix(lines, "diff --git") {
		return nil, errNotRecognised
	}
	var out []string
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "index "):
			continue
		case strings.HasPrefix(l, " "):
			continue
		}
		out = append(out, l)
	}
	return out, nil
}

// gitStatusFilter drops the "(use ...)" hint lines and blank lines. The state
// lines are kept as they are, so the branch and file lists read the same.
func gitStatusFilter(lines []string) ([]string, error) {
	var out []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "(use ") {
			continue
		}
		out = append(out, l)
	}
	return out, nil
}

// gitLogFilter keeps one line per commit: the 7-character short sha with any
// decoration, followed by the first non-empty message line. Author, Date, and
// Merge lines are dropped. A non-empty input with no "commit " line fails open.
func gitLogFilter(lines []string) ([]string, error) {
	if len(lines) > 0 && !containsPrefix(lines, "commit ") {
		return nil, errNotRecognised
	}
	var out []string
	awaitingSubject := false
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "commit "):
			fields := strings.Fields(l)
			sha := fields[1]
			if len(sha) > 7 {
				sha = sha[:7]
			}
			rest := ""
			if len(fields) > 2 {
				rest = " " + strings.Join(fields[2:], " ")
			}
			out = append(out, "commit "+sha+rest)
			awaitingSubject = true
		case awaitingSubject && strings.TrimSpace(l) != "":
			out = append(out, "    "+strings.TrimSpace(l))
			awaitingSubject = false
		}
	}
	return out, nil
}

// grepFilter drops blank lines and collapses runs of "--" group separators into
// one. The match lines themselves are untouched, so file:line references stay
// exact.
func grepFilter(lines []string) ([]string, error) {
	var out []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if l == "--" && len(out) > 0 && out[len(out)-1] == "--" {
			continue
		}
		out = append(out, l)
	}
	return out, nil
}

// listingFilter drops the "total N" line from ls -l and blank lines.
func listingFilter(lines []string) ([]string, error) {
	var out []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "total ") {
			continue
		}
		out = append(out, l)
	}
	return out, nil
}

func containsPrefix(lines []string, prefix string) bool {
	for _, l := range lines {
		if strings.HasPrefix(l, prefix) {
			return true
		}
	}
	return false
}
