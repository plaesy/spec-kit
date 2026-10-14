// Package exclude is the single definition of which paths the repo-wide scans
// (the dependency graph, analyze's walks, the doc and semantic indexes, and the
// validators) leave out. Every walker asks this package instead of keeping its
// own list, so a folder excluded in one output is excluded in all of them.
package exclude

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// dirNames are directory base names that hold dependencies, build output or
// tool caches rather than project source. Dot-directories (.git, .plaesy,
// .dart_tool, .venv, ...) are excluded by DirName's prefix rule, not listed here.
var dirNames = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"__pycache__":  true,
	"venv":         true,
	"Pods":         true,
	"DerivedData":  true,
	"coverage":     true,
}

// DirName reports whether a directory with this base name is excluded: any
// dot-directory, or one of the well-known dependency and build directories.
func DirName(name string) bool {
	if name == "" {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	return dirNames[name]
}

// RelPath reports whether a slash-separated path relative to the project root
// is excluded by its location rather than by its name. Flutter generates
// platform host projects with an "ephemeral" folder under each platform
// (ios/Flutter/ephemeral, linux/flutter/ephemeral, ...); it is tooling output.
func RelPath(rel string) bool {
	return strings.HasSuffix(strings.ToLower(rel), "/flutter/ephemeral") ||
		strings.ToLower(rel) == "flutter/ephemeral"
}

// Ignore holds the rules read from a project's root .gitignore. The zero value
// and a nil *Ignore both match nothing.
type Ignore struct {
	rules []rule
}

type rule struct {
	re      *regexp.Regexp
	negate  bool
	dirOnly bool
}

// LoadGitignore reads root/.gitignore. A missing or unreadable file yields an
// empty Ignore, so a project without one is walked exactly as before. Only the
// root file is read: nested .gitignore files are not supported.
func LoadGitignore(root string) *Ignore {
	f, err := os.Open(filepath.Join(root, ".gitignore"))
	if err != nil {
		return &Ignore{}
	}
	defer f.Close()
	ig := &Ignore{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if r, ok := parseRule(sc.Text()); ok {
			ig.rules = append(ig.rules, r)
		}
	}
	return ig
}

// Match reports whether a slash-separated path relative to the root is ignored
// by the .gitignore rules. Rules apply in order and the last one that matches
// decides, so a later "!" pattern re-includes a file.
func (ig *Ignore) Match(rel string, isDir bool) bool {
	if ig == nil {
		return false
	}
	ignored := false
	for _, r := range ig.rules {
		if r.dirOnly && !isDir {
			continue
		}
		if r.re.MatchString(rel) {
			ignored = !r.negate
		}
	}
	return ignored
}

// SkipDir reports whether a directory must not be entered. rel is its
// slash-separated path relative to the root and name its base name.
func (ig *Ignore) SkipDir(rel, name string) bool {
	return DirName(name) || RelPath(rel) || ig.Match(rel, true)
}

// SkipFile reports whether a file must be left out of a scan.
func (ig *Ignore) SkipFile(rel string) bool {
	return ig.Match(rel, false)
}

// parseRule turns one .gitignore line into a rule. Blank lines and comments
// yield no rule. The supported syntax is the common subset: "!" negation, a
// trailing "/" for directories only, a leading "/" or any inner "/" to anchor
// the pattern to the root, "*", "?", "[...]", and "**" as a whole path segment.
func parseRule(line string) (rule, bool) {
	line = strings.TrimRight(line, " \t\r")
	if line == "" || strings.HasPrefix(line, "#") {
		return rule{}, false
	}
	var r rule
	if strings.HasPrefix(line, "!") {
		r.negate = true
		line = line[1:]
	} else if strings.HasPrefix(line, `\`) {
		line = line[1:]
	}
	if strings.HasSuffix(line, "/") {
		r.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}
	anchored := strings.Contains(line, "/")
	line = strings.TrimPrefix(line, "/")
	if line == "" {
		return rule{}, false
	}
	pat := "(?:^|/)" + globToRegexp(line) + "$"
	if anchored {
		pat = "^" + globToRegexp(line) + "$"
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return rule{}, false
	}
	r.re = re
	return r, true
}

// globToRegexp converts a gitignore glob into an unanchored regexp fragment.
func globToRegexp(glob string) string {
	var b strings.Builder
	for i := 0; i < len(glob); {
		switch {
		case strings.HasPrefix(glob[i:], "**/"):
			b.WriteString(`(?:.*/)?`)
			i += 3
		case strings.HasPrefix(glob[i:], "/**") && i+3 == len(glob):
			b.WriteString(`/.*`)
			i += 3
		case strings.HasPrefix(glob[i:], "**"):
			b.WriteString(`.*`)
			i += 2
		case glob[i] == '*':
			b.WriteString(`[^/]*`)
			i++
		case glob[i] == '?':
			b.WriteString(`[^/]`)
			i++
		case glob[i] == '[':
			end := strings.IndexByte(glob[i+1:], ']')
			if end < 0 {
				b.WriteString(regexp.QuoteMeta("["))
				i++
				continue
			}
			class := glob[i+1 : i+1+end]
			if strings.HasPrefix(class, "!") {
				class = "^" + class[1:]
			}
			b.WriteString("[" + class + "]")
			i += end + 2
		case glob[i] == '\\' && i+1 < len(glob):
			b.WriteString(regexp.QuoteMeta(string(glob[i+1])))
			i += 2
		default:
			b.WriteString(regexp.QuoteMeta(string(glob[i])))
			i++
		}
	}
	return b.String()
}
