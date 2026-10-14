// Package scaffold ports scripts/bash/plaesy-init.sh ("plaesy init") to Go.
// It creates the .plaesy/ project structure and populates it with
// instructions, templates, checklists, roles and AI-platform-specific
// prompt/core files copied from the Plaesy repo root.
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/assets"
	"github.com/plaesy/spec-kit/internal/installer"
)

// FindHome locates the Plaesy repo root that holds templates/, instructions/,
// prompts/, agents/ and checklists/ as siblings of scripts/.
//
// Resolution order mirrors install.sh's PLAESY_HOME convention and
// common.sh's VERSION-file search:
//  1. override (e.g. --plaesy-home flag), if non-empty
//  2. PLAESY_HOME environment variable, if set
//  3. walk upward from the executable's directory looking for a directory
//     that contains a templates/ subdirectory (a spec-kit checkout the
//     binary happens to be running from)
//  4. walk upward from the current working directory, same check
//  5. installer.PlaesyHomeDir() — the managed home, self-provisioned from
//     the embedded assets on first need (see step 5's comment below) if it
//     is not already there
func FindHome(override string) (string, error) {
	if override != "" {
		if isPlaesyHome(override) {
			return filepath.Clean(override), nil
		}
		return "", fmt.Errorf("--plaesy-home %q does not look like a Plaesy repo root (missing templates/ or scripts/configs/platform.json)", override)
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

	// The managed home is provisioned here, lazily, rather than only by
	// `plaesy install`. The binary can reach a machine by many paths that
	// never run that subcommand -- install.sh / install.ps1 just place the
	// downloaded binary, a package manager just unpacks it -- and every one
	// of them must still produce a working `plaesy init`. Extracting here
	// means the binary provisions its own home the first time anything
	// needs it, regardless of how it got onto the machine.
	if managedHome, err := installer.PlaesyHomeDir(); err == nil {
		if !isPlaesyHome(managedHome) {
			_ = assets.Extract(managedHome) // best-effort; isPlaesyHome below is the real check
		}
		if isPlaesyHome(managedHome) {
			return managedHome, nil
		}
	}

	return "", fmt.Errorf("could not locate or provision a Plaesy home: set PLAESY_HOME or pass --plaesy-home")
}

// walkUpForHome walks upward from start looking for a directory containing
// a templates/ subdirectory (the repo root marker used throughout this
// package, matching how templates/instructions/prompts/agents/checklists
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

// isPlaesyHome requires both a templates/ subdirectory and a readable
// scripts/configs/platform.json. templates/ alone is not enough: a stale or
// partial install (e.g. a leftover templates/ directory from an older
// installer) can satisfy that check while missing platform.json, which then
// makes LoadPlatformConfig silently fall back to a config with zero
// platforms — every --ai choice is then rejected with a confusing "available:
// []" instead of pointing at the real problem (a broken PLAESY_HOME).
func isPlaesyHome(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "templates"))
	if err != nil || !info.IsDir() {
		return false
	}
	platformInfo, err := os.Stat(filepath.Join(dir, "scripts", "configs", "platform.json"))
	return err == nil && !platformInfo.IsDir()
}
