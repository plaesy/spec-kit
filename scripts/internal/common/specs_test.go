package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindSpecName(t *testing.T) {
	tests := []struct {
		name    string
		folders []string
		branch  string
		want    string
		wantErr string
	}{
		{"spec named after the branch", []string{"auth-login"}, "auth-login", "auth-login", ""},
		{"no spec for the branch", []string{"auth-login"}, "billing", "", "no spec for branch"},
		{"a longer name is not the same branch", []string{"x-auth-login"}, "auth-login", "", "no spec for branch"},
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
			if err != nil || got != tt.want {
				t.Fatalf("FindSpecName = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}
