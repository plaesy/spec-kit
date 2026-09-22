// Package scaffold ports scripts/bash/plaesy-init.sh ("plaesy init") to Go.
// It creates the .plaesy/ project structure and populates it with
// instructions, templates, checklists, roles and AI-platform-specific
// prompt/core files copied from the Plaesy repo root.
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindHome locates the Plaesy repo root that holds templates/, instructions/,
// prompts/, chatmodes/ and checklists/ as siblings of scripts/.
//
// Resolution order mirrors install.sh's PLAESY_HOME convention and
// common.sh's VERSION-file search:
//  1. override (e.g. --plaesy-home flag), if non-empty
//  2. PLAESY_HOME environment variable, if set
//  3. walk upward from the executable's directory looking for a directory
//     that contains a templates/ subdirectory
//  4. walk upward from the current working directory, same check
func FindHome(override string) (string, error) {
	if override != "" {
		if isPlaesyHome(override) {
			return filepath.Clean(override), nil
		}
		return "", fmt.Errorf("--plaesy-home %q does not look like a Plaesy repo root (no templates/ subdirectory)", override)
	}

	if envHome := os.Getenv("PLAESY_HOME"); envHome != "" {
		if isPlaesyHome(envHome) {
			return filepath.Clean(envHome), nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		if home, ok := walkUpForHome(filepath.Dir(exe)); ok {
			return home, nil
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		if home, ok := walkUpForHome(cwd); ok {
			return home, nil
		}
	}

	return "", fmt.Errorf("could not locate Plaesy repo root: set PLAESY_HOME or pass --plaesy-home")
}

// walkUpForHome walks upward from start looking for a directory containing
// a templates/ subdirectory (the repo root marker used throughout this
// package, matching how templates/instructions/prompts/chatmodes/checklists
// live as siblings at the repo root).
func walkUpForHome(start string) (string, bool) {
	dir := start
	for {
		if isPlaesyHome(dir) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func isPlaesyHome(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "templates"))
	return err == nil && info.IsDir()
}
