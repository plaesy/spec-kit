package analyze

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/plaesy/spec-kit/internal/graph"
	"github.com/plaesy/spec-kit/internal/search/docindex"
)

// Watch runs analyze once, then polls the watch fingerprint and runs analyze
// again whenever it changes. The fingerprint covers the graph's source files and
// the document index's documents (see watchFingerprint), so an edit to either
// re-runs analyze and keeps both indexes current. Each re-run goes through Run,
// which is fingerprint-gated, so a run with nothing relevant changed does no
// extra work.
func Watch(opts Options, interval time.Duration) error {
	if err := Run(opts); err != nil {
		return err
	}

	repo := opts.ProjectPath
	if repo == "" {
		repo = "."
	}
	gopts := graph.Options{RepoPath: repo}
	paths, err := graph.ResolvePaths(gopts)
	if err != nil {
		return err
	}
	last, err := watchFingerprint(gopts, paths, repo)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "[plaesy analyze] Watch mode: polling every %s. Ctrl+C to stop.\n", interval)
	for {
		time.Sleep(interval)
		fp, err := watchFingerprint(gopts, paths, repo)
		if err != nil {
			return err
		}
		if fp == last {
			continue
		}
		fmt.Fprintln(os.Stderr, "[plaesy analyze] Change detected, re-running analyze...")
		if err := Run(opts); err != nil {
			return err
		}
		if last, err = watchFingerprint(gopts, paths, repo); err != nil {
			return err
		}
	}
}

// watchFingerprint identifies the state of everything a watch re-run would
// rebuild: the graph's source files (graph.SourceFingerprint) and the document
// index's documents (docindex.Collect, the same set `plaesy search` indexes).
// The document half is mtime-based, like the graph half, so it costs one stat
// per document on each poll rather than a content read.
func watchFingerprint(gopts graph.Options, paths graph.Paths, repo string) (string, error) {
	src, err := graph.SourceFingerprint(gopts, paths)
	if err != nil {
		return "", err
	}
	docs, err := docindex.Collect(repo)
	if err != nil {
		return "", err
	}
	var newest int64
	for _, rel := range docs {
		fi, err := os.Stat(filepath.Join(repo, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		if m := fi.ModTime().UnixNano(); m > newest {
			newest = m
		}
	}
	return fmt.Sprintf("%s|%d|%d", src, len(docs), newest), nil
}
