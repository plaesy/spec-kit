// Package installer implements the "plaesy install / status / uninstall /
// repair / upgrade" command family.
//
// v1 model: the running Go binary itself is the deliverable. There is no
// generated shell wrapper and no git clone of the whole spec-kit repo into
// a PLAESY_HOME directory (both of which scripts/bash/install.sh did).
// "install" simply copies the current binary to a well-known per-OS bin
// directory; "status" reports where it lives and whether that directory is
// on PATH; "uninstall" removes it after a y/N confirmation. "repair" and
// "upgrade" are stubs for v1 — see Repair/Upgrade below.
package installer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/plaesy/spec-kit/internal/assets"
)

// BinaryName is the installed executable's file name (without extension;
// InstallDir/BinaryName gets ".exe" appended on Windows by binaryFileName).
const BinaryName = "plaesy"

// InstallDir returns the well-known per-OS directory plaesy installs into:
//   - Windows: %LOCALAPPDATA%\Plaesy\bin
//   - macOS/Linux: $HOME/.local/bin
//
// This mirrors install.sh's BIN_DIR choice on Unix, and picks the closest
// Windows equivalent (no POSIX $HOME/.local/bin convention there).
func InstallDir() (string, error) {
	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve install directory: %w", err)
			}
			base = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(base, "Plaesy", "bin"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve install directory: %w", err)
	}
	return filepath.Join(home, ".local", "bin"), nil
}

// AppDataDir returns the per-OS directory Plaesy owns for its own data, as
// opposed to InstallDir which holds only the binary:
//   - Windows: %LOCALAPPDATA%\Plaesy
//   - macOS/Linux: $HOME/.local/share/plaesy
func AppDataDir() (string, error) {
	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve app data directory: %w", err)
			}
			base = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(base, "Plaesy"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve app data directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "plaesy"), nil
}

// PlaesyHomeDir returns the managed Plaesy home: AppDataDir()/.plaesy. This
// is the directory `plaesy install` populates from its embedded assets
// (templates/, instructions/, prompts/, agents/, checklists/,
// scripts/configs), and the one scaffold.FindHome falls back to when no
// project-local checkout or explicit override/PLAESY_HOME is found — so
// `plaesy init` works without the user ever cloning spec-kit themselves.
func PlaesyHomeDir() (string, error) {
	dir, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".plaesy"), nil
}

func binaryFileName() string {
	if runtime.GOOS == "windows" {
		return BinaryName + ".exe"
	}
	return BinaryName
}

// InstalledPath returns the full path the binary is (or would be) installed
// at: InstallDir()/plaesy[.exe].
func InstalledPath() (string, error) {
	dir, err := InstallDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, binaryFileName()), nil
}

// OnPath reports whether dir appears as an entry of the PATH environment
// variable. Comparison is exact-path on Unix and case-insensitive on
// Windows, where PATH entries commonly differ only in case or trailing
// separators.
func OnPath(dir string) bool {
	if dir == "" {
		return false
	}
	pathEnv := os.Getenv("PATH")
	sep := string(os.PathListSeparator)
	target := filepath.Clean(dir)
	for _, entry := range strings.Split(pathEnv, sep) {
		// A PATH entry can carry the whitespace that came with it — "…\bin;
		// C:\Users\me\bin" is a real PATH, and quoting inside a PATH entry is
		// common on Windows. filepath.Clean does not remove surrounding spaces,
		// so without this the directory is on PATH and reported as not on it.
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		candidate := filepath.Clean(entry)
		candidate = strings.Trim(candidate, `"`)
		if runtime.GOOS == "windows" {
			if strings.EqualFold(candidate, target) {
				return true
			}
		} else if candidate == target {
			return true
		}
	}
	return false
}

// PathInstructions returns human-readable, OS-appropriate instructions for
// adding dir to PATH. install.sh never edits shell rc files on the user's
// behalf either — it only prints what to add — so this does the same.
func PathInstructions(dir string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(
			"%q is not on your PATH. Add it, e.g. in PowerShell:\n"+
				"  [Environment]::SetEnvironmentVariable('Path', $env:Path + ';%s', 'User')\n"+
				"then open a new terminal.",
			dir, dir)
	}
	return unixPathInstructions(dir, os.Getenv("SHELL"))
}

// unixPathInstructions is the Unix half of PathInstructions, with $SHELL passed
// in rather than read here. Which rc file a user is told to edit is the part
// that can be wrong — telling a zsh user to add a line to ~/.profile is
// advice that silently does nothing — and reading the environment inside the
// function made that untestable on a Windows CI run, which is where this
// package's other tests all run.
func unixPathInstructions(dir, shell string) string {
	shellHint := "~/.profile"
	if strings.Contains(shell, "zsh") {
		shellHint = "~/.zshrc"
	} else if strings.Contains(shell, "bash") {
		shellHint = "~/.bashrc (or ~/.bash_profile)"
	}
	return fmt.Sprintf(
		"%q is not on your PATH. Add this line to %s, then restart your shell:\n"+
			"  export PATH=\"%s:$PATH\"",
		dir, shellHint, dir)
}

// Install copies the currently running executable to InstallDir, creating
// the directory if needed, then extracts the embedded Plaesy home (see
// PlaesyHomeDir) so `plaesy init` has a source tree to copy from without the
// user needing a spec-kit checkout of their own. It returns the binary's
// destination path.
func Install() (string, error) {
	src, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate running binary: %w", err)
	}
	src, err = filepath.EvalSymlinks(src)
	if err != nil {
		return "", fmt.Errorf("resolve running binary path: %w", err)
	}
	dst, err := installFrom(src)
	if err != nil {
		return "", err
	}

	homeDir, err := PlaesyHomeDir()
	if err != nil {
		return dst, fmt.Errorf("binary installed, but resolve Plaesy home: %w", err)
	}
	if err := assets.Extract(homeDir); err != nil {
		return dst, fmt.Errorf("binary installed, but extract bundled assets to %s: %w", homeDir, err)
	}

	return dst, nil
}

// installFrom is Install with the source named rather than discovered. The
// guard that makes Install safe to run from the installed copy — the one that
// stops copyFile truncating the binary it is executing — can only be reached if
// the source is a value the test chooses, since os.Executable() points
// wherever `go test` happened to build the binary. Naming the source also
// reads as what it is: a copy, with a self-copy case to avoid.
func installFrom(src string) (string, error) {
	dst, err := InstalledPath()
	if err != nil {
		return "", err
	}

	if filepath.Clean(src) == filepath.Clean(dst) {
		// Already running from the install location. copyFile truncates its
		// destination before writing, so without this an `plaesy install` run
		// from the installed copy would empty the binary it is executing.
		return dst, nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", fmt.Errorf("create install directory: %w", err)
	}

	if err := copyFile(src, dst); err != nil {
		return "", fmt.Errorf("copy binary to %s: %w", dst, err)
	}

	return dst, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	// Remove any existing file first: on Windows, Rename fails if dst
	// exists (and the running binary may hold dst open).
	_ = os.Remove(dst)
	if err := os.Rename(tmp, dst); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// Status describes the current installation state, for "plaesy status".
type Status struct {
	InstallDir  string
	InstallPath string
	Installed   bool
	OnPath      bool
	RunningFrom string
	Version     string
}

// CollectStatus gathers current installation status. version should be
// common.Version.
func CollectStatus(version string) (Status, error) {
	dir, err := InstallDir()
	if err != nil {
		return Status{}, err
	}
	path, err := InstalledPath()
	if err != nil {
		return Status{}, err
	}

	st := Status{
		InstallDir:  dir,
		InstallPath: path,
		OnPath:      OnPath(dir),
		Version:     version,
	}

	if _, err := os.Stat(path); err == nil {
		st.Installed = true
	}

	if running, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(running); err == nil {
			st.RunningFrom = resolved
		} else {
			st.RunningFrom = running
		}
	}

	return st, nil
}

// Confirm reads a y/N confirmation line from r, matching install.sh's
// `read -p "... [Y/n]: "` prompts elsewhere in the script but defaulting to
// N (as required for a destructive uninstall). Only "y" or "yes"
// (case-insensitive) counts as confirmation; anything else, including an
// empty line, is a decline.
func Confirm(r io.Reader, prompt string) bool {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}

// Uninstall removes the installed binary at InstalledPath, if present.
func Uninstall() (string, error) {
	path, err := InstalledPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path, nil
	}
	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("remove %s: %w", path, err)
	}
	return path, nil
}

// NotImplementedMessage is printed by "plaesy repair" and "plaesy upgrade"
// in v1: self-update requires a release/distribution pipeline (signed
// binaries published somewhere fetchable) that does not exist yet, so this
// deliberately does not attempt a half-working curl-based downloader.
const NotImplementedMessage = "not yet implemented — re-run 'plaesy install' with the latest release binary"
