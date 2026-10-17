package featurepath

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// worktreeHomeEnv overrides where feature worktrees are kept. It is for
// users who want them on another drive, and for tests.
const worktreeHomeEnv = "PLAESY_WORKTREE_HOME"

// WorktreeHome returns the directory that holds every feature worktree, for all
// projects. It lives outside any repository so a worktree never nests inside
// the tree it was created from:
//
//   - PLAESY_WORKTREE_HOME, when set
//   - Windows: %LOCALAPPDATA%\Plaesy\worktrees
//   - elsewhere: $XDG_DATA_HOME/plaesy/worktrees, or ~/.local/share/plaesy/worktrees
func WorktreeHome() (string, error) {
	if v := os.Getenv(worktreeHomeEnv); v != "" {
		return v, nil
	}
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "Plaesy", "worktrees"), nil
		}
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" && runtime.GOOS != "windows" {
		return filepath.Join(xdg, "plaesy", "worktrees"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locating worktree home: %w", err)
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Local", "Plaesy", "worktrees"), nil
	}
	return filepath.Join(home, ".local", "share", "plaesy", "worktrees"), nil
}

// ProjectWorktreeDir returns the folder that holds one project's worktrees,
// {WorktreeHome}/{project}-{hash8}. The hash comes from the main checkout's
// path, so two repositories with the same folder name do not share worktrees.
func ProjectWorktreeDir(mainRoot string) (string, error) {
	home, err := WorktreeHome()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(filepath.Clean(mainRoot)))
	return filepath.Join(home, filepath.Base(mainRoot)+"-"+hex.EncodeToString(sum[:])[:8]), nil
}

// isInside reports whether path is dir or lies beneath it. Both are cleaned,
// and the comparison is case-insensitive on Windows, where paths differ only
// in case.
func isInside(path, dir string) bool {
	p, d := filepath.Clean(path), filepath.Clean(dir)
	if runtime.GOOS == "windows" {
		p, d = strings.ToLower(p), strings.ToLower(d)
	}
	rel, err := filepath.Rel(d, p)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
