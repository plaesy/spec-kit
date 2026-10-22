package main

// The constitution's product gate says a new command ships with "its registry
// entry, its docs row, and its CHANGELOG line, or it is not done". Two of those
// three had tests: TestEveryCommandIsInTheReferenceIndex and
// TestEverySubcommandIsMentionedInTheReference cover the docs, and
// templates_drift_test.go covers the registry. The CHANGELOG had none, and the
// noun-parent split showed what that costs — `context`, `features`, `images`,
// `platforms` and `stack` shipped with no changelog line at all, while the
// Unreleased section still advertised `inject-ai-headers` (deleted outright),
// `check-task-prerequisites`, `detect-stack`, `task-manage`,
// `update-agent-context`, `get-feature-paths`, `create feature`, `create image`
// and `config detect-platform`. A reader following any of those lines would have
// typed a command that errors.
//
// There is deliberately no companion test for retired spellings. It was written
// and dropped, and the reason is worth not rediscovering: the changelog records
// fixes against the commands as they were spelled at the time, so a bullet in
// Unreleased can name a command that has since been renamed, and that text is
// correct — it is the name the reader would have had to type when the fix
// shipped. Every heuristic for telling that apart from a live instruction to
// type an old name (a nearby "renamed"/"formerly", the enclosing section, the
// bullet rather than the line) resolved to "rewrite thirteen historical
// bullets", which falsifies the log and pads it to satisfy a checker. The
// staleness that actually mattered was the missing entries, which this test
// does catch, and it is caught from the command tree rather than from a list of
// names that no longer resolve anywhere and so cannot be derived.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoChangelog locates CHANGELOG.md by walking up from the package directory,
// the same way repoDocs finds the reference.
func repoChangelog(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "CHANGELOG.md")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("CHANGELOG.md not found: not a repository checkout")
	return ""
}

// unreleasedSection returns the text under "## [Unreleased]" up to the next
// top-level "## " heading.
func unreleasedSection(body string) string {
	const start = "## [Unreleased]"
	i := strings.Index(body, start)
	if i < 0 {
		return ""
	}
	rest := body[i+len(start):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		return rest[:j]
	}
	return rest
}
