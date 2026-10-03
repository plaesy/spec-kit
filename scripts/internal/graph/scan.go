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
//
// The excluded-dir and output-dir tests live on the directory branch alone.
// A file's path cannot contain an excluded segment, because the walk skips such
// a directory before it descends, and the same is true of the output dir. The
// checks used to be repeated per file, where they could never fire: untestable
// branches that read as if the file case were handled separately.
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
//
// The root survives the walk. ".." stops at the root rather than climbing out
// of it, and a Windows drive letter stays attached instead of being pushed
// behind a leading slash. Both matter because every caller compares the result
// against `repoRootSlash + "/"`: a path that came back as "/C:/repo/docs/a.md"
// when the root is "C:/repo" matches nothing, so on Windows every markdown
// link, dot-source, JS/Python/Dart import and Java import resolved to "" and
// the graph came out with no structural edges at all — the command whose whole
// job is finding references, silently finding none.
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
			// Climbing above the root is dropped, not applied: the drive letter
			// is the floor.
			if len(parts) > 0 && !isDriveSeg(parts[len(parts)-1]) {
				parts = parts[:len(parts)-1]
			}
		default:
			parts = append(parts, seg)
		}
	}
	root := "/"
	if len(parts) > 0 && isDriveSeg(parts[0]) {
		root = parts[0][:1] + ":/"
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return root
	}
	return root + strings.Join(parts, "/")
}

// isDriveSeg reports whether a path segment is a Windows drive specifier
// ("C:") left behind by splitting an absolute path on "/".
func isDriveSeg(seg string) bool {
	return len(seg) == 2 && seg[1] == ':' &&
		((seg[0] >= 'A' && seg[0] <= 'Z') || (seg[0] >= 'a' && seg[0] <= 'z'))
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
	// The root is not a node. Letting it through the prefix check below meant
	// TrimPrefix could not strip the trailing slash that is not there, and an
	// absolute path — with a drive letter on Windows — came back as a "node id",
	// which matches nothing in `known` and would embed a machine-specific path
	// in project.graph.json.
	prefix := repoRootSlash + "/"
	if resolved == repoRootSlash || !strings.HasPrefix(resolved+"/", prefix) {
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
