package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewSpecNameIsUTCTimestampThenBranch(t *testing.T) {
	// 14:30:22 in UTC+7 is 07:30:22 UTC: the name must use UTC.
	loc := time.FixedZone("UTC+7", 7*3600)
	now := time.Date(2026, 10, 15, 14, 30, 22, 0, loc)
	if got := NewSpecName(now, "auth-login"); got != "20261015-073022-auth-login" {
		t.Errorf("NewSpecName = %q, want the UTC timestamp", got)
	}
}

func TestSpecBranch(t *testing.T) {
	tests := map[string]string{
		"20261015-143022-auth-login": "auth-login",
		"001-add-login":              "001-add-login",
		"20261015-143022-x":          "x",
		"2026-auth":                  "2026-auth",
	}
	for folder, want := range tests {
		if got := SpecBranch(folder); got != want {
			t.Errorf("SpecBranch(%q) = %q, want %q", folder, got, want)
		}
	}
}

func TestFindSpecName(t *testing.T) {
	tests := []struct {
		name    string
		folders []string
		branch  string
		want    string
		wantErr string
	}{
		{"timestamped spec", []string{"20261015-143022-auth-login"}, "auth-login", "20261015-143022-auth-login", ""},
		{"legacy NNN spec", []string{"001-add-login"}, "001-add-login", "001-add-login", ""},
		{"no spec", []string{"20261015-143022-auth-login"}, "billing", "", "no spec for branch"},
		{"two specs share a branch", []string{"20261015-143022-auth-login", "20261016-090000-auth-login"}, "auth-login", "", "several specs"},
		{"a branch that merely ends the name is not a match", []string{"20261015-143022-x-auth-login"}, "auth-login", "", "no spec for branch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, f := range tt.folders {
				if err := os.MkdirAll(filepath.Join(root, ".plaesy", "specs", f), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			got, err := FindSpecName(root, tt.branch)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("FindSpecName = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetFeatureDirFallsBackToTheBranchPath(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".plaesy", "specs", "new-feature")
	if got := GetFeatureDir(root, "new-feature"); got != want {
		t.Errorf("GetFeatureDir with no spec = %q, want %q", got, want)
	}
}
