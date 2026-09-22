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
		if entry == "" {
			continue
		}
		candidate := filepath.Clean(entry)
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

	shellHint := "~/.profile"
	if shell := os.Getenv("SHELL"); strings.Contains(shell, "zsh") {
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
// the directory if needed. It returns the destination path.
func Install() (string, error) {
	src, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate running binary: %w", err)
	}
	src, err = filepath.EvalSymlinks(src)
	if err != nil {
		return "", fmt.Errorf("resolve running binary path: %w", err)
	}

	dst, err := InstalledPath()
	if err != nil {
		return "", err
	}

	if filepath.Clean(src) == filepath.Clean(dst) {
		return dst, nil // already running from the install location
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
