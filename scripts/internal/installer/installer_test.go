package installer

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBinaryFileName(t *testing.T) {
	expected := "plaesy"
	if runtime.GOOS == "windows" {
		expected = "plaesy.exe"
	}
	if got := binaryFileName(); got != expected {
		t.Errorf("binaryFileName() = %q, want %q", got, expected)
	}
}

func TestOnPath(t *testing.T) {
	sep := string(os.PathListSeparator)
	dirSep := string(os.PathSeparator)

	tests := []struct {
		name     string
		pathEnv  string
		dir      string
		expected bool
	}{
		{
			name:     "empty PATH",
			pathEnv:  "",
			dir:      "some" + dirSep + "dir",
			expected: false,
		},
		{
			name:     "dir not in PATH",
			pathEnv:  "usr" + dirSep + "bin" + sep + "bin",
			dir:      "some" + dirSep + "dir",
			expected: false,
		},
		{
			name:     "dir in PATH",
			pathEnv:  "usr" + dirSep + "bin" + sep + "home" + dirSep + "user" + dirSep + ".local" + dirSep + "bin" + sep + "bin",
			dir:      "home" + dirSep + "user" + dirSep + ".local" + dirSep + "bin",
			expected: true,
		},
		{
			name:     "dir with trailing separator in PATH",
			pathEnv:  "usr" + dirSep + "bin" + sep + "home" + dirSep + "user" + dirSep + ".local" + dirSep + "bin" + dirSep + sep + "bin",
			dir:      "home" + dirSep + "user" + dirSep + ".local" + dirSep + "bin",
			expected: true,
		},
		{
			name:     "empty dir",
			pathEnv:  "usr" + dirSep + "bin" + sep + "bin",
			dir:      "",
			expected: false,
		},
		{
			// A PATH entry keeps whatever whitespace came with it. The
			// directory is on PATH, so reporting otherwise tells the user to
			// edit a PATH that already has the entry.
			name:     "surrounding whitespace in a PATH entry is ignored",
			pathEnv:  "usr" + dirSep + "bin" + sep + " " + "home" + dirSep + "user" + dirSep + ".local" + dirSep + "bin" + " " + sep + "bin",
			dir:      "home" + dirSep + "user" + dirSep + ".local" + dirSep + "bin",
			expected: true,
		},
		{
			name:     "a quoted PATH entry is still that directory",
			pathEnv:  "usr" + dirSep + "bin" + sep + "\"" + "home" + dirSep + "user" + dirSep + ".local" + dirSep + "bin" + "\"",
			dir:      "home" + dirSep + "user" + dirSep + ".local" + dirSep + "bin",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldPath := os.Getenv("PATH")
			os.Setenv("PATH", tt.pathEnv)
			defer os.Setenv("PATH", oldPath)

			if got := OnPath(tt.dir); got != tt.expected {
				t.Errorf("OnPath(%q) = %v, want %v", tt.dir, got, tt.expected)
			}
		})
	}

	if runtime.GOOS == "windows" {
		t.Run("Windows case-insensitive match", func(t *testing.T) {
			oldPath := os.Getenv("PATH")
			os.Setenv("PATH", "C:\\Windows;C:\\Program Files\\Plaesy\\bin")
			defer os.Setenv("PATH", oldPath)

			if got := OnPath(`c:\program files\plaesy\bin`); !got {
				t.Errorf("OnPath should match case-insensitively on Windows")
			}
		})

		t.Run("Windows trailing backslash", func(t *testing.T) {
			oldPath := os.Getenv("PATH")
			os.Setenv("PATH", `C:\Windows;C:\Plaesy\bin\`)
			defer os.Setenv("PATH", oldPath)

			if got := OnPath(`C:\Plaesy\bin`); !got {
				t.Errorf("OnPath should handle trailing backslash on Windows")
			}
		})
	}
}

func TestPathInstructions(t *testing.T) {
	dir := "/home/user/.local/bin"

	if runtime.GOOS == "windows" {
		t.Run("Windows instructions", func(t *testing.T) {
			inst := PathInstructions(dir)
			if !strings.Contains(inst, dir) {
				t.Errorf("instructions should contain dir: %s", inst)
			}
			if !strings.Contains(inst, "SetEnvironmentVariable") {
				t.Errorf("Windows instructions should mention SetEnvironmentVariable: %s", inst)
			}
			if !strings.Contains(inst, "PowerShell") {
				t.Errorf("Windows instructions should mention PowerShell: %s", inst)
			}
		})
	} else {
		t.Run("bash shell", func(t *testing.T) {
			oldShell := os.Getenv("SHELL")
			os.Setenv("SHELL", "/bin/bash")
			defer os.Setenv("SHELL", oldShell)

			inst := PathInstructions(dir)
			if !strings.Contains(inst, dir) {
				t.Errorf("instructions should contain dir: %s", inst)
			}
			if !strings.Contains(inst, ".bashrc") && !strings.Contains(inst, ".bash_profile") {
				t.Errorf("bash instructions should mention .bashrc or .bash_profile: %s", inst)
			}
			if !strings.Contains(inst, "export PATH") {
				t.Errorf("bash instructions should contain export PATH: %s", inst)
			}
		})

		t.Run("zsh shell", func(t *testing.T) {
			oldShell := os.Getenv("SHELL")
			os.Setenv("SHELL", "/bin/zsh")
			defer os.Setenv("SHELL", oldShell)

			inst := PathInstructions(dir)
			if !strings.Contains(inst, ".zshrc") {
				t.Errorf("zsh instructions should mention .zshrc: %s", inst)
			}
		})

		t.Run("unknown shell defaults to ~/.profile", func(t *testing.T) {
			oldShell := os.Getenv("SHELL")
			os.Setenv("SHELL", "/bin/fish")
			defer os.Setenv("SHELL", oldShell)

			inst := PathInstructions(dir)
			if !strings.Contains(inst, ".profile") {
				t.Errorf("unknown shell should default to .profile: %s", inst)
			}
		})

		t.Run("empty SHELL defaults to ~/.profile", func(t *testing.T) {
			oldShell := os.Getenv("SHELL")
			os.Unsetenv("SHELL")
			defer os.Setenv("SHELL", oldShell)

			inst := PathInstructions(dir)
			if !strings.Contains(inst, ".profile") {
				t.Errorf("empty SHELL should default to .profile: %s", inst)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "source.bin")
	dst := filepath.Join(tmpDir, "dest.bin")

	content := []byte("test binary content\n")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	dstContent, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(dstContent) != string(content) {
		t.Errorf("copied content mismatch: got %q, want %q", dstContent, content)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dest: %v", err)
	}
	if runtime.GOOS != "windows" {
		if info.Mode().Perm()&0o111 == 0 {
			t.Errorf("dest should be executable, got %o", info.Mode().Perm())
		}
	}
}

func TestCopyFile_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "source.bin")
	dst := filepath.Join(tmpDir, "dest.bin")

	if err := os.WriteFile(src, []byte("new content"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(dst, []byte("old content"), 0o644); err != nil {
		t.Fatalf("write dest: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	dstContent, _ := os.ReadFile(dst)
	if string(dstContent) != "new content" {
		t.Errorf("should overwrite: got %q", dstContent)
	}
}

func TestCopyFile_SourceNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "nonexistent.bin")
	dst := filepath.Join(tmpDir, "dest.bin")

	err := copyFile(src, dst)
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestCopyFile_DestDirNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "source.bin")
	dst := filepath.Join(tmpDir, "subdir", "dest.bin")

	if err := os.WriteFile(src, []byte("content"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	err := copyFile(src, dst)
	if err == nil {
		t.Error("expected error for nonexistent dest directory")
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"y", "y\n", true},
		{"yes", "yes\n", true},
		{"Y", "Y\n", true},
		{"YES", "YES\n", true},
		{"Yes", "Yes\n", true},
		{"n", "n\n", false},
		{"no", "no\n", false},
		{"N", "N\n", false},
		{"empty", "\n", false},
		{"whitespace", "  \n", false},
		{"random", "maybe\n", false},
		{"y with spaces", " y \n", true},
		{"yes with spaces", " yes \n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got := Confirm(r, "prompt> ")
			if got != tt.expected {
				t.Errorf("Confirm(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}

	t.Run("EOF returns false", func(t *testing.T) {
		r := strings.NewReader("")
		got := Confirm(r, "prompt> ")
		if got {
			t.Error("Confirm at EOF should return false")
		}
	})

	t.Run("prints prompt", func(t *testing.T) {
		var buf strings.Builder
		r := strings.NewReader("y\n")
		// We can't easily capture fmt.Print output, but we can verify
		// the function doesn't panic and returns correctly
		got := Confirm(io.TeeReader(r, &buf), "test prompt> ")
		if !got {
			t.Error("should confirm with y")
		}
	})
}

func TestCollectStatus(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		oldHome = os.Getenv("USERPROFILE")
	}
	defer os.Setenv("HOME", oldHome)
	if runtime.GOOS == "windows" {
		defer os.Setenv("USERPROFILE", oldHome)
	}

	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", tmpDir)
		os.Setenv("LOCALAPPDATA", filepath.Join(tmpDir, "AppData", "Local"))
	} else {
		os.Setenv("HOME", tmpDir)
	}

	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)

	installDir, _ := InstallDir()
	os.Setenv("PATH", installDir+string(os.PathListSeparator)+oldPath)

	installedPath, _ := InstalledPath()
	if err := os.MkdirAll(filepath.Dir(installedPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(installedPath, []byte("fake binary"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	st, err := CollectStatus("1.2.3")
	if err != nil {
		t.Fatalf("CollectStatus failed: %v", err)
	}

	if !st.Installed {
		t.Error("Installed should be true when binary exists")
	}
	if st.InstallPath != installedPath {
		t.Errorf("InstallPath = %q, want %q", st.InstallPath, installedPath)
	}
	if st.InstallDir != installDir {
		t.Errorf("InstallDir = %q, want %q", st.InstallDir, installDir)
	}
	if !st.OnPath {
		t.Error("OnPath should be true when dir is in PATH")
	}
	if st.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", st.Version, "1.2.3")
	}
	if st.RunningFrom == "" {
		t.Error("RunningFrom should be populated")
	}
}

func TestCollectStatus_NotInstalled(t *testing.T) {
	tmpDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		oldHome = os.Getenv("USERPROFILE")
	}
	defer os.Setenv("HOME", oldHome)
	if runtime.GOOS == "windows" {
		defer os.Setenv("USERPROFILE", oldHome)
	}

	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", tmpDir)
		os.Setenv("LOCALAPPDATA", filepath.Join(tmpDir, "AppData", "Local"))
	} else {
		os.Setenv("HOME", tmpDir)
	}

	st, err := CollectStatus("1.2.3")
	if err != nil {
		t.Fatalf("CollectStatus failed: %v", err)
	}

	if st.Installed {
		t.Error("Installed should be false when binary doesn't exist")
	}
	if st.InstallPath == "" {
		t.Error("InstallPath should be set even when not installed")
	}
	if st.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", st.Version, "1.2.3")
	}
}

func TestInstallDir(t *testing.T) {
	dir, err := InstallDir()
	if err != nil {
		t.Fatalf("InstallDir failed: %v", err)
	}
	if dir == "" {
		t.Error("InstallDir should not be empty")
	}

	if runtime.GOOS == "windows" {
		if !strings.Contains(dir, "Plaesy") || !strings.Contains(dir, "bin") {
			t.Errorf("Windows InstallDir should contain Plaesy\\bin: %s", dir)
		}
	} else {
		if !strings.Contains(dir, ".local") || !strings.Contains(dir, "bin") {
			t.Errorf("Unix InstallDir should contain .local/bin: %s", dir)
		}
	}
}

func TestInstalledPath(t *testing.T) {
	path, err := InstalledPath()
	if err != nil {
		t.Fatalf("InstalledPath failed: %v", err)
	}
	if path == "" {
		t.Error("InstalledPath should not be empty")
	}

	expectedName := binaryFileName()
	if !strings.HasSuffix(path, expectedName) {
		t.Errorf("InstalledPath should end with %q: %s", expectedName, path)
	}

	dir, _ := InstallDir()
	expectedPath := filepath.Join(dir, expectedName)
	if filepath.Clean(path) != filepath.Clean(expectedPath) {
		t.Errorf("InstalledPath = %q, want %q", path, expectedPath)
	}
}

func TestInstallDir_FallbackWhenLOCALAPPDATAEmpty(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	tmpDir := t.TempDir()
	oldLocalAppData := os.Getenv("LOCALAPPDATA")
	oldUserProfile := os.Getenv("USERPROFILE")
	defer os.Setenv("LOCALAPPDATA", oldLocalAppData)
	defer os.Setenv("USERPROFILE", oldUserProfile)

	os.Unsetenv("LOCALAPPDATA")
	os.Setenv("USERPROFILE", tmpDir)

	dir, err := InstallDir()
	if err != nil {
		t.Fatalf("InstallDir failed: %v", err)
	}
	expected := filepath.Join(tmpDir, "AppData", "Local", "Plaesy", "bin")
	if filepath.Clean(dir) != filepath.Clean(expected) {
		t.Errorf("InstallDir = %q, want %q", dir, expected)
	}
}

func TestInstallDir_UserHomeDirError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}

	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Unsetenv("HOME")

	_, err := InstallDir()
	if err == nil {
		t.Error("expected error when HOME is not set on Unix")
	}
	if !strings.Contains(err.Error(), "resolve install directory") {
		t.Errorf("error should mention 'resolve install directory': %v", err)
	}
}
