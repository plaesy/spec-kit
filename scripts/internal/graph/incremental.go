package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
)

// buildCacheVersion is written into the cache and checked on load. Bump it
// whenever a file's extracted results change for the same content (a new
// extractor, a changed regex or resolution rule), so an old cache is discarded
// instead of silently reused.
const buildCacheVersion = 1

const buildCacheName = "graph.cache.json"

// cachedFile is one file's results from the last build. Size and mtime let a
// build skip reading a file it has seen; the hash decides whether a file whose
// mtime moved actually changed.
type cachedFile struct {
	Size    int64    `json:"size"`
	MtimeNS int64    `json:"mtime_ns"`
	Hash    string   `json:"hash"`
	Symbols []string `json:"symbols,omitempty"`
	Edges   []Edge   `json:"edges,omitempty"`
}

// buildCache is what the last build computed, and what it was computed against.
type buildCache struct {
	Version  int                   `json:"version"`
	Files    []string              `json:"files"`
	Manifest string                `json:"manifest"`
	Entries  map[string]cachedFile `json:"entries"`
}

// fileWork counts the files whose results were recomputed in the current run,
// so a test can check that an unchanged file was not reread. It is not part of
// the output.
var fileWork int

// fileResults returns the structural and mention edges and the symbols of every
// file. A file's results depend on its own content, on the node set, and on the
// go.mod and pubspec.yaml manifests. So when the node set or a manifest changed,
// every file is recomputed. Otherwise only files whose content changed are.
//
// The result matches a full build exactly. Only the work differs.
func fileResults(opts Options, p Paths, files []string, known map[string]bool) ([]Edge, map[string][]string, error) {
	cachePath := filepath.Join(p.OutFull, buildCacheName)
	manifest := manifestDigest(p.RepoRootSlash)

	var prev *buildCache
	if !opts.Full {
		prev = loadBuildCache(cachePath)
	}
	full := prev == nil || !sameStrings(prev.Files, files) || prev.Manifest != manifest

	ctx := newExtractContext(p.RepoRootSlash, files, known)
	cands := mentionCandidates(files)
	next := buildCache{Version: buildCacheVersion, Files: files, Manifest: manifest, Entries: make(map[string]cachedFile, len(files))}

	var edges []Edge
	symbols := make(map[string][]string, len(files))
	for _, rel := range files {
		abs := joinRoot(p.RepoRootSlash, rel)
		var size, mtime int64
		if info, err := os.Stat(abs); err == nil {
			size, mtime = info.Size(), info.ModTime().UnixNano()
		}
		old, seen := cachedFile{}, false
		if !full {
			old, seen = prev.Entries[rel]
		}

		// Unchanged size and mtime: the cached results stand, and the file is not read.
		if seen && old.Size == size && old.MtimeNS == mtime {
			next.Entries[rel] = old
			edges = append(edges, old.Edges...)
			if len(old.Symbols) > 0 {
				symbols[rel] = old.Symbols
			}
			continue
		}

		content := readAll(abs)
		hash := hashContent(content)
		// Touched but not edited: the same content gives the same results.
		if seen && old.Hash == hash {
			old.Size, old.MtimeNS = size, mtime
			next.Entries[rel] = old
			edges = append(edges, old.Edges...)
			if len(old.Symbols) > 0 {
				symbols[rel] = old.Symbols
			}
			continue
		}

		fileWork++
		entry := cachedFile{Size: size, MtimeNS: mtime, Hash: hash}
		if content != "" {
			entry.Symbols = symbolsOf(rel, content)
			entry.Edges = append(ctx.fileEdges(rel, content), mentionEdgesFor(rel, content, cands)...)
		}
		next.Entries[rel] = entry
		edges = append(edges, entry.Edges...)
		if len(entry.Symbols) > 0 {
			symbols[rel] = entry.Symbols
		}
	}

	// The cache only saves work, so a failed write is reported and the build goes on.
	if err := saveBuildCache(cachePath, next); err != nil {
		fmt.Fprintf(os.Stderr, "[plaesy-graph] could not save build cache (next build will be full): %v\n", err)
	}
	return edges, symbols, nil
}

func hashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// manifestDigest covers the go.mod and pubspec.yaml files, which the import
// resolvers read by content. Changing one changes how other files resolve.
func manifestDigest(repoRootSlash string) string {
	h := sha256.New()
	for _, name := range []string{"go.mod", "pubspec.yaml"} {
		for _, rel := range findFilesNamed(repoRootSlash, name) {
			fmt.Fprintf(h, "%s\x00%s\x00", rel, readAll(joinRoot(repoRootSlash, rel)))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// loadBuildCache returns the cache at path, or nil when there is none or it
// cannot be used. A nil cache means a full build.
func loadBuildCache(path string) *buildCache {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var c buildCache
	if err := json.Unmarshal(b, &c); err != nil || c.Version != buildCacheVersion || c.Entries == nil {
		return nil
	}
	return &c
}

// saveBuildCache writes the cache to a temporary file and renames it into place,
// so an interrupted write never leaves a partial cache behind.
func saveBuildCache(path string, c buildCache) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sa := append([]string(nil), a...)
	sb := append([]string(nil), b...)
	sort.Strings(sa)
	sort.Strings(sb)
	return reflect.DeepEqual(sa, sb)
}
