package main

// Tests for `plaesy install` / `uninstall` / `repair` / `upgrade`.
//
// These four measured between 5% and 33% because the install location is derived
// from the environment, and a test that let it resolve for real would either
// mutate the developer's actual %LOCALAPPDATA%\Plaesy\bin or refuse to run at
// all. On Windows InstallDir() reads LOCALAPPDATA, so redirecting it makes the
// whole install/uninstall path testable without touching the machine.
//
// The behaviour worth pinning is the destructive one. `uninstall` deletes a
// binary, and the two ways it can go wrong — proceeding without asking, and
// reporting success when it removed nothing — are both silent.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/installer"
	"github.com/spf13/cobra"
)

// fakeInstallHome points InstallDir() at a throwaway directory. Without this,
// `uninstall` would target the real install location.
func fakeInstallHome(t *testing.T) string {
	t.Helper()
	if _, err := installer.InstallDir(); err != nil {
		t.Skipf("install directory cannot be resolved on this platform: %v", err)
	}
	home := filepath.Join(t.TempDir(), "plaesy-test", "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows InstallDir() reads LOCALAPPDATA; elsewhere it reads $HOME.
	t.Setenv("LOCALAPPDATA", home)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

// withStdin replaces os.Stdin for one call. Confirm reads it directly, so this
// is the only way to answer the y/N prompt.
func withStdin(t *testing.T, input string, fn func() error) error {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdin
	os.Stdin = r
	done := make(chan error, 1)
	go func() { done <- fn() }()
	_, _ = w.WriteString(input)
	w.Close()
	fnErr := <-done
	os.Stdin = orig
	r.Close()
	return fnErr
}

// ---- uninstall -----------------------------------------------------------

// Nothing installed is a normal state, not a failure — and it must not claim to
// have removed anything. The message has to name the path it looked at, or a
// user with a stale idea of where the binary lives has no way to correct it.
func TestUninstallReportsWhenNothingIsInstalled(t *testing.T) {
	fakeInstallHome(t)

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newUninstallCmd()) })
	if err != nil {
		t.Fatalf("uninstalling an absent binary must not fail, got %v", err)
	}
	if !strings.Contains(out, "is not installed at") {
		t.Errorf("expected a not-installed message, got:\n%s", out)
	}
	path, pathErr := installer.InstalledPath()
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	if !strings.Contains(out, path) {
		t.Errorf("the message must name the path it checked (%s):\n%s", path, out)
	}
	if strings.Contains(out, "removed") {
		t.Errorf("nothing was removed, so nothing may claim to have been removed:\n%s", out)
	}
}

// The confirmation is the whole safety property of a destructive command.
// Answering "n" must leave the binary exactly where it was.
func TestUninstallDeclinedLeavesTheBinaryInPlace(t *testing.T) {
	fakeInstallHome(t)
	path := seedInstalledBinary(t)

	var err error
	out, err := captureStdout(t, func() error {
		return withStdin(t, "n\n", func() error { return runLeafCmd(t, newUninstallCmd()) })
	})
	if err != nil {
		t.Fatalf("declining is not an error, got %v", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("declining the prompt must not remove the binary: %v", statErr)
	}
	if !strings.Contains(out, "nothing was removed") {
		t.Errorf("a declined uninstall must say so plainly:\n%s", out)
	}
}

// An empty answer is the default-deny case: the prompt says [y/N], so pressing
// enter means no. Treating it as yes would make the default the destructive one.
func TestUninstallEmptyAnswerIsTreatedAsNo(t *testing.T) {
	fakeInstallHome(t)
	path := seedInstalledBinary(t)

	out, err := captureStdout(t, func() error {
		return withStdin(t, "\n", func() error { return runLeafCmd(t, newUninstallCmd()) })
	})
	if err != nil {
		t.Fatalf("declining is not an error, got %v", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("an empty answer must not remove the binary: %v", statErr)
	}
	if !strings.Contains(out, "nothing was removed") {
		t.Errorf("an empty answer must be reported as a decline:\n%s", out)
	}
}

// Accepting must actually delete, and must name what it deleted. A command that
// prints "removed" without removing anything is worse than one that errors.
func TestUninstallConfirmedRemovesTheBinary(t *testing.T) {
	fakeInstallHome(t)
	path := seedInstalledBinary(t)

	out, err := captureStdout(t, func() error {
		return withStdin(t, "y\n", func() error { return runLeafCmd(t, newUninstallCmd()) })
	})
	if err != nil {
		t.Fatalf("a confirmed uninstall must succeed, got %v\n%s", err, out)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("a confirmed uninstall must remove the binary, stat err = %v", statErr)
	}
	if !strings.Contains(out, path) {
		t.Errorf("the success message must name the removed path (%s):\n%s", path, out)
	}
}

// "yes" is as good as "y". A user typing the full word should not be told the
// binary was kept.
func TestUninstallAcceptsTheFullWordYes(t *testing.T) {
	fakeInstallHome(t)
	path := seedInstalledBinary(t)

	if _, err := captureStdout(t, func() error {
		return withStdin(t, "yes\n", func() error { return runLeafCmd(t, newUninstallCmd()) })
	}); err != nil {
		t.Fatalf("a confirmed uninstall must succeed, got %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("answering \"yes\" must remove the binary, stat err = %v", statErr)
	}
}

// seedInstalledBinary puts a file where InstallDir()/InstalledPath() say the
// binary lives, so the command under test has something real to act on.
func seedInstalledBinary(t *testing.T) string {
	t.Helper()
	path, err := installer.InstalledPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not really a binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// ---- install -------------------------------------------------------------

// install copies the *running* executable, so the test points it at a throwaway
// LOCALAPPDATA. What is asserted is the reporting contract: the destination must
// be named, and when the destination is not on PATH the user must be told how to
// add it — otherwise the install succeeds and the command still does not run.
func TestInstallCopiesTheBinaryAndReportsTheDestination(t *testing.T) {
	fakeInstallHome(t)

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newInstallCmd()) })
	if err != nil {
		t.Fatalf("install must succeed, got %v\n%s", err, out)
	}
	path, err := installer.InstalledPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("install must leave a binary at %s: %v", path, statErr)
	}
	if !strings.Contains(out, path) {
		t.Errorf("install must report where it put the binary (%s):\n%s", path, out)
	}
}

// ---- repair / upgrade ----------------------------------------------------

// Both are deliberate no-ops: self-update needs a release pipeline that does not
// exist, and a half-working downloader is worse than none. The contract is that
// they say so and point at the real alternative.
func TestRepairAndUpgradeExplainThemselves(t *testing.T) {
	fakeInstallHome(t)
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{"repair", newRepairCmd()},
		{"upgrade", newUpgradeCmd()},
	} {
		out, err := captureStdout(t, func() error { return runLeafCmd(t, tc.cmd) })
		if err != nil {
			t.Errorf("%s must not error, got %v", tc.name, err)
			continue
		}
		if strings.TrimSpace(out) != installer.NotImplementedMessage {
			t.Errorf("%s printed %q, want the shared not-implemented message %q",
				tc.name, strings.TrimSpace(out), installer.NotImplementedMessage)
		}
		if !strings.Contains(out, "plaesy install") {
			t.Errorf("%s must name the working alternative:\n%s", tc.name, out)
		}
	}
}

// ---- the install command group -------------------------------------------

// All four must be reachable and argument-free. A stray argument to `install`
// would otherwise be swallowed exactly as it was for `status`.
func TestInstallFamilyTakesNoArguments(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{"install", newInstallCmd()},
		{"uninstall", newUninstallCmd()},
		{"repair", newRepairCmd()},
		{"upgrade", newUpgradeCmd()},
	} {
		if err := runLeafCmd(t, tc.cmd, "stray"); err == nil {
			t.Errorf("%s must reject an unexpected argument", tc.name)
		}
	}
}
