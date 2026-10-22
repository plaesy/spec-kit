package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDropPathEntryRemovesOnlyTheMatchingEntry(t *testing.T) {
	got, changed := dropPathEntry(`C:\Tools;C:\Users\me\AppData\Local\Plaesy\bin;C:\Go\bin`, `C:\Users\me\AppData\Local\Plaesy\bin`, true)
	if !changed {
		t.Fatal("changed = false, want true")
	}
	if want := `C:\Tools;C:\Go\bin`; got != want {
		t.Errorf("dropPathEntry() = %q, want %q", got, want)
	}
}

func TestDropPathEntryFoldsCaseAndTrailingSeparatorOnWindows(t *testing.T) {
	value := `c:\users\me\appdata\local\plaesy\bin\;C:\Go\bin`
	got, changed := dropPathEntry(value, `C:\Users\me\AppData\Local\Plaesy\bin`, true)
	if !changed {
		t.Fatal("changed = false, want true: Windows PATH matching is case-insensitive")
	}
	if want := `C:\Go\bin`; got != want {
		t.Errorf("dropPathEntry() = %q, want %q", got, want)
	}
}

func TestDropPathEntryIsCaseSensitiveOffWindows(t *testing.T) {
	value := `/home/Me/.local/bin:/usr/bin`
	if _, changed := dropPathEntry(value, `/home/me/.local/bin`, false); changed {
		t.Error("changed = true, want false: Unix PATH matching is exact")
	}
}

func TestDropPathEntryWithoutAMatchReturnsTheValueUnchanged(t *testing.T) {
	value := `C:\Tools;C:\Go\bin`
	got, changed := dropPathEntry(value, `C:\Users\me\AppData\Local\Plaesy\bin`, true)
	if changed {
		t.Error("changed = true, want false")
	}
	if got != value {
		t.Errorf("dropPathEntry() = %q, want the input unchanged", got)
	}
}

func TestDropPathEntryKeepsQuotedAndEmptyEntries(t *testing.T) {
	value := `"C:\Program Files\X";;C:\Users\me\AppData\Local\Plaesy\bin`
	got, changed := dropPathEntry(value, `C:\Users\me\AppData\Local\Plaesy\bin`, true)
	if !changed {
		t.Fatal("changed = false, want true")
	}
	if want := `"C:\Program Files\X";`; got != want {
		t.Errorf("dropPathEntry() = %q, want %q", got, want)
	}
}

// installerBlock is exactly what install.sh appends to a profile.
func installerBlock(dir string) string {
	return "\n# Added by the plaesy installer\nexport PATH=\"" + dir + ":$PATH\"\n"
}

func TestRemoveInstallerLinesUndoesWhatTheInstallerWrote(t *testing.T) {
	dir := "/home/me/.local/bin"
	profile := filepath.Join(t.TempDir(), ".bashrc")
	before := "export EDITOR=vim\nalias ll='ls -l'\n"
	if err := os.WriteFile(profile, []byte(before+installerBlock(dir)), 0o600); err != nil {
		t.Fatal(err)
	}

	changed, err := removeInstallerLines(profile, dir)
	if err != nil {
		t.Fatalf("removeInstallerLines: %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	got, _ := os.ReadFile(profile)
	if string(got) != before {
		t.Errorf("profile = %q, want the original content %q", got, before)
	}
}

func TestRemoveInstallerLinesLeavesAHandWrittenExportAlone(t *testing.T) {
	dir := "/home/me/.local/bin"
	profile := filepath.Join(t.TempDir(), ".zshrc")
	// The same export, but without the marker the installer adds: the user wrote it.
	content := "export PATH=\"" + dir + ":$PATH\"\n"
	if err := os.WriteFile(profile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	changed, err := removeInstallerLines(profile, dir)
	if err != nil {
		t.Fatalf("removeInstallerLines: %v", err)
	}
	if changed {
		t.Error("changed = true, want false: the marker is what marks a line as ours")
	}
	got, _ := os.ReadFile(profile)
	if string(got) != content {
		t.Errorf("profile = %q, want it untouched", got)
	}
}

func TestRemoveInstallerLinesOnlyRemovesTheBlockForThatDir(t *testing.T) {
	profile := filepath.Join(t.TempDir(), ".bashrc")
	content := installerBlock("/other/bin") + installerBlock("/home/me/.local/bin")
	if err := os.WriteFile(profile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := removeInstallerLines(profile, "/home/me/.local/bin"); err != nil {
		t.Fatalf("removeInstallerLines: %v", err)
	}
	got, _ := os.ReadFile(profile)
	if !strings.Contains(string(got), `export PATH="/other/bin:$PATH"`) {
		t.Errorf("block for another directory was removed: %q", got)
	}
	if strings.Contains(string(got), "/home/me/.local/bin") {
		t.Errorf("block for the target directory is still present: %q", got)
	}
}

func TestRemoveInstallerLinesMissingProfileIsNotAnError(t *testing.T) {
	changed, err := removeInstallerLines(filepath.Join(t.TempDir(), "absent"), "/x")
	if err != nil {
		t.Fatalf("missing profile returned %v, want nil", err)
	}
	if changed {
		t.Error("changed = true for a missing profile")
	}
}
