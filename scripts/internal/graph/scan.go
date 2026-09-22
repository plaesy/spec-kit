package graph

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// collectFiles mirrors the bash script's step 1: find every candidate file
// under repoRoot, relative-slash path, matching INCLUDE_EXT_RE, not under an
// excluded dir, and not under outDirNorm (the graph's own output dir).
func collectFiles(repoRoot, outDirNorm string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(repoRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// Best-effort scan: skip unreadable entries rather than aborting.
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, rerr := filepath.Rel(repoRoot, p)
		if rerr != nil {
			return nil
		}
		relSlash := toSlash(rel)
		if relSlash == "." {
			return nil
		}
		if d.IsDir() {
			if excludeDirSegment(d.Name()) {
				return filepath.SkipDir
			}
			if outDirNorm != "" && (relSlash == outDirNorm || strings.HasPrefix(relSlash, outDirNorm+"/")) {
				return filepath.SkipDir
			}
			return nil
		}
		if pathExcluded(relSlash) {
			return nil
		}
		if outDirNorm != "" && strings.HasPrefix(relSlash, outDirNorm+"/") {
			return nil
		}
		if !includeExtRE.MatchString(relSlash) {
			return nil
		}
		out = append(out, relSlash)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// normalizePath is a pure-lexical realpath -m equivalent (no filesystem
// access, symlinks not resolved): mirrors normalize_path(). base is an
// absolute forward-slash directory; p may itself be absolute (Unix "/..." or
// Windows "C:...") in which case base is ignored, matching realpath.
func normalizePath(base, p string) string {
	if strings.HasPrefix(p, "/") || isWindowsAbs(p) {
		base = ""
	}
	combined := base
	if p != "" {
		if combined != "" {
			combined += "/" + p
		} else {
			combined = p
		}
	}
	var parts []string
	for _, seg := range strings.Split(combined, "/") {
		switch seg {
		case "", ".":
			continue
		case "..":
			if len(parts) > 0 {
				parts = parts[:len(parts)-1]
			}
		default:
			parts = append(parts, seg)
		}
	}
	return "/" + strings.Join(parts, "/")
}

func isWindowsAbs(p string) bool {
	return len(p) >= 2 && ((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) && p[1] == ':'
}

// resolveRelLink mirrors resolve_rel_link(): resolve a link found inside
// fromRel against fromRel's directory to a known repo-relative node path.
// Returns "" if the link is external/empty/anchor-only/unresolved.
func resolveRelLink(repoRootSlash, fromRel, link string, known map[string]bool) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return ""
	}
	if strings.HasPrefix(link, "#") {
		return ""
	}
	if i := strings.IndexByte(link, '#'); i >= 0 {
		link = link[:i]
	}
	if link == "" {
		return ""
	}
	fromDir := repoRootSlash
	if i := strings.LastIndexByte(fromRel, '/'); i >= 0 {
		fromDir = repoRootSlash + "/" + fromRel[:i]
	}
	resolved := normalizePath(fromDir, link)
	prefix := repoRootSlash + "/"
	if !strings.HasPrefix(resolved+"/", prefix) && resolved != repoRootSlash {
		return ""
	}
	rel := strings.TrimPrefix(resolved, prefix)
	if known[rel] {
		return rel
	}
	return ""
}

// readAll reads a file's full content; returns "" on error (mirrors the
// bash/awk pipeline's tolerance of unreadable files via 2>/dev/null).
func readAll(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}

// joinRoot joins a repo-root (forward-slash, absolute) with a relative-slash
// path to an OS path usable with os.ReadFile.
func joinRoot(repoRootSlash, rel string) string {
	return filepath.FromSlash(path.Join(repoRootSlash, rel))
}
