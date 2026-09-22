package common

import (
	"runtime"
	"strings"
	"testing"
)

func TestToNativePath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("path conversion only applies on Windows")
	}
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"msys drive lowercase", "/c/Users/foo/bar", `C:\Users\foo\bar`},
		{"msys drive uppercase", "/D/Projects/repo", `D:\Projects\repo`},
		{"native windows unchanged", `C:\Users\foo`, `C:\Users\foo`},
		{"unix root preserved", "/home/user/repo", `\home\user\repo`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toNativePath(tt.in)
			if got != tt.want {
				t.Errorf("toNativePath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMsysDriveRE(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("regex behavior only validated on Windows")
	}
	cases := []string{"/c/Users", "/D/Projects", "/z/bar"}
	for _, c := range cases {
		if !msysDriveRE.MatchString(c) {
			t.Errorf("expected %q to match msysDriveRE", c)
		}
	}
	if msysDriveRE.MatchString("C:/Users") {
		t.Error("native Windows path should not match")
	}
	if msysDriveRE.MatchString("/usr/local") {
		t.Error("/usr should not match (no drive letter)")
	}
}

func TestGetRepoRootProducesValidDrivePath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows path validation only")
	}
	root, err := GetRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	// A Windows absolute path must start with a drive letter: C:\ or D:\ etc.
	if len(root) < 3 || root[1] != ':' || root[2] != '\\' {
		t.Errorf("GetRepoRoot returned non-drive path: %q", root)
	}
	if strings.HasPrefix(root, `\c\`) || strings.HasPrefix(root, `\d\`) {
		t.Errorf("GetRepoRoot returned unconverted MSYS2 path: %q", root)
	}
}
