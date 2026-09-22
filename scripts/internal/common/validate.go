package common

import (
	"fmt"
	"os"
	"os/exec"
)

// ValidateCommandExists mirrors validate_command_exists from common.sh.
func ValidateCommandExists(cmd, description string) error {
	if description == "" {
		description = "command"
	}
	path, err := exec.LookPath(cmd)
	if err != nil {
		LogError("Required %s not found: %s", description, cmd)
		return fmt.Errorf("required %s not found: %s", description, cmd)
	}
	LogDebug("Found %s: %s", cmd, path)
	return nil
}

// ValidateFileExists mirrors validate_file_exists.
func ValidateFileExists(path, description string) error {
	if description == "" {
		description = "file"
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		LogError("Required %s not found: %s", description, path)
		return fmt.Errorf("required %s not found: %s", description, path)
	}
	LogDebug("Found %s: %s", description, path)
	return nil
}

// ValidateDirectoryExists mirrors validate_directory_exists.
func ValidateDirectoryExists(path, description string) error {
	if description == "" {
		description = "directory"
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		LogError("Required %s not found: %s", description, path)
		return fmt.Errorf("required %s not found: %s", description, path)
	}
	LogDebug("Found %s: %s", description, path)
	return nil
}

// ValidateNotRoot mirrors validate_not_root (POSIX EUID check; always passes
// on Windows since there is no equivalent root concept).
func ValidateNotRoot() error {
	if os.Geteuid() == 0 {
		LogError("This script should not be run as root for security reasons.")
		return fmt.Errorf("running as root is not permitted")
	}
	return nil
}

// ValidateEnvironment mirrors validate_environment: checks required external
// tools and warns (does not fail) if not inside a git repository.
func ValidateEnvironment() error {
	LogDebug("Validating execution environment...")

	for _, cmd := range []string{"git"} {
		if err := ValidateCommandExists(cmd, "command"); err != nil {
			return err
		}
	}

	if _, err := GetRepoRoot(); err != nil {
		LogWarning("Not in a git repository")
	}

	LogSuccess("Environment validation passed")
	return nil
}

// CheckFile mirrors check_file: prints a ✓/✗ status line for a status report.
func CheckFile(path, label string) string {
	if _, err := os.Stat(path); err == nil {
		return fmt.Sprintf("  ✓ %s", label)
	}
	return fmt.Sprintf("  ✗ %s", label)
}

// CheckDir mirrors check_dir: ✓ only if the directory exists and is non-empty.
func CheckDir(path, label string) string {
	entries, err := os.ReadDir(path)
	if err == nil && len(entries) > 0 {
		return fmt.Sprintf("  ✓ %s", label)
	}
	return fmt.Sprintf("  ✗ %s", label)
}
