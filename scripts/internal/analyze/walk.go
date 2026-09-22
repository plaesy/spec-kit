package analyze

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// fileInfo captures the subset of os/fs data the analyzer needs per file,
// gathered in a single filepath.WalkDir pass so every generator function
// reuses it instead of re-walking the tree (mirrors the bash script's push
// toward one `find` per concern instead of one per check).
type fileInfo struct {
	relPath string // relative to projectPath, forward-slash separated
	absPath string
	size    int64
	modUnix int64
}

// dirInfo captures a directory relative to projectPath plus its direct file
// count, mirroring the bash script's file_counts_map.
type dirInfo struct {
	relPath   string
	fileCount int
}

// walkResult is the single tree walk shared by every generator.
type walkResult struct {
	files      []fileInfo // excludes dotfile dirs, node_modules, .plaesy
	dirs       []dirInfo
	newestUnix int64 // newest mtime among files (for the fingerprint's own use elsewhere if needed)
}

// isExcludedDir mirrors the bash script's `! -path "*/.*/*" ! -path
// "*/node_modules/*"` pruning, applied at directory-descent time so excluded
// subtrees are never visited (matching find's behavior, and much faster).
func isExcludedDirName(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules"
}

// walkProject walks projectPath once, excluding dotfile directories and
// node_modules (matching `! -path "*/.*/*" ! -path "*/node_modules/*"` used
// throughout plaesy-analyze.sh for its general-purpose file/dir listings).
// The project root itself is not excluded even though it may start with a
// dot-named parent elsewhere on disk — only descendants named like that are
// pruned, matching find's path-substring semantics for depth >= 1.
func walkProject(projectPath string) (walkResult, error) {
	var res walkResult
	dirFileCounts := map[string]int{}

	err := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Unreadable entry: skip it, matching find's silent 2>/dev/null.
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == projectPath {
			return nil
		}
		rel, relErr := filepath.Rel(projectPath, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if isExcludedDirName(d.Name()) {
				return filepath.SkipDir
			}
			res.dirs = append(res.dirs, dirInfo{relPath: rel})
			return nil
		}

		// File: skip if any path component is excluded (covers files
		// directly inside an excluded dir that wasn't pruned because the
		// exclusion check only runs at dir-descent; here the parent was
		// never pruned only if it's the root, so this is a defensive check
		// for files at any depth whose ancestor should have been skipped).
		info, statErr := d.Info()
		var size, modUnix int64
		if statErr == nil {
			size = info.Size()
			modUnix = info.ModTime().Unix()
		}
		fi := fileInfo{relPath: rel, absPath: path, size: size, modUnix: modUnix}
		res.files = append(res.files, fi)
		if modUnix > res.newestUnix {
			res.newestUnix = modUnix
		}
		parentDir := filepath.ToSlash(filepath.Dir(path))
		dirFileCounts[parentDir]++
		return nil
	})
	if err != nil {
		return res, err
	}

	for i := range res.dirs {
		abs := filepath.Join(projectPath, filepath.FromSlash(res.dirs[i].relPath))
		res.dirs[i].fileCount = dirFileCounts[filepath.ToSlash(abs)]
	}
	return res, nil
}
