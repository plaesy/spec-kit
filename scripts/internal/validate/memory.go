package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MemoryIssue is one file found to contain external, non-self-contained
// references, mirroring one "[✗] External references in: <file>" block from
// plaesy-validate-memory.sh.
type MemoryIssue struct {
	File string
	// Lines maps each matched pattern to the matching "N:content" lines
	// found via grep -n, in the order the patterns were checked.
	Patterns []string
	Lines    map[string][]string
}

// MemoryResult carries the findings of Memory.
type MemoryResult struct {
	FilesChecked int
	Issues       []MemoryIssue
}

// memoryPatterns mirrors the PATTERNS array in plaesy-validate-memory.sh
// exactly, including its two grep-BRE regexes (the "C:\\Users\\.*\\.claude"
// and "/Users/.*/.claude" entries).
var memoryPatternSources = []string{
	`~/.claude`,
	`~/.claude/projects`,
	`~/.claude/plans`,
	`/tmp/`,
	`C:\\Users\\.*\\.claude`,
	`/Users/.*/.claude`,
	`CLAUDE_CODE_DIR`,
}

var memoryPatterns = compileMemoryPatterns()

func compileMemoryPatterns() []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(memoryPatternSources))
	for i, p := range memoryPatternSources {
		out[i] = regexp.MustCompile(p)
	}
	return out
}

// Memory ports plaesy-validate-memory.sh: it scans every *.md file under
// <repoRoot>/.plaesy/memory for a fixed list of external-path patterns
// (paths outside the project, such as ~/.claude or /tmp/), reporting which
// files and lines matched. Unlike the bash script this does not print
// directly; the caller renders MemoryResult (see validate.go for the CLI
// wiring) so the same logic is testable headless.
func Memory(repoRoot string) (*MemoryResult, error) {
	memDir := filepath.Join(repoRoot, ".plaesy", "memory")
	info, err := os.Stat(memDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("memory directory not found: %s", filepath.ToSlash(memDir))
	}

	res := &MemoryResult{}

	var files []string
	err = filepath.WalkDir(memDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", memDir, err)
	}

	for _, file := range files {
		res.FilesChecked++

		data, err := os.ReadFile(file)
		if err != nil {
			continue // matches bash's `2>/dev/null || true` tolerance for unreadable files
		}
		lines := strings.Split(string(data), "\n")

		issue := MemoryIssue{File: file, Lines: map[string][]string{}}
		for i, pat := range memoryPatterns {
			src := memoryPatternSources[i]
			var matched []string
			for lineNo, line := range lines {
				if pat.MatchString(line) {
					matched = append(matched, fmt.Sprintf("%d:%s", lineNo+1, line))
				}
			}
			if len(matched) > 0 {
				issue.Patterns = append(issue.Patterns, src)
				issue.Lines[src] = matched
			}
		}
		if len(issue.Patterns) > 0 {
			res.Issues = append(res.Issues, issue)
		}
	}

	return res, nil
}
