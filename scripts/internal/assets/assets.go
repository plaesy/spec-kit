// Package assets embeds the Plaesy source tree (templates/, instructions/,
// prompts/, agents/, checklists/, scripts/configs) that `plaesy init` copies
// from a "Plaesy home" into a target project.
//
// Historically that source tree only existed if a user had a full clone of
// the spec-kit repo and pointed PLAESY_HOME (or --plaesy-home) at it — the
// Go installer only ever copied the plaesy binary itself. Embedding a synced
// copy here lets `plaesy install` materialize a working home on its own, at
// installer.PlaesyHomeDir(), so `plaesy init` works out of the box on a
// machine that never cloned spec-kit.
//
// The data/ directory is a generated mirror of the repo root, kept in sync
// by `go run ./internal/assets/gen` (wired into `make build`/`make assets`).
// TestAssetsDataMatchesSource in this package fails the build if it drifts.
package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:data
var data embed.FS

// Extract materializes the embedded tree into destDir, overwriting any file
// already there. destDir is expected to be a directory Plaesy owns
// exclusively (see installer.PlaesyHomeDir) — nothing user-authored belongs
// there, so overwriting on every `plaesy install` is what keeps it current
// after an upgrade rather than something that needs reconciling.
func Extract(destDir string) error {
	return fs.WalkDir(data, "data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("data", path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(destDir, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		content, err := data.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o644)
	})
}
