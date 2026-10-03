package common

import (
	"os"
	"runtime"
	"testing"
)

// The repo-root repair must never invent a path: a candidate is only accepted
// once it exists.
func TestRootViaCwdRejectsOutsideARepository(t *testing.T) {
	if _, err := os.Stat(`\`); err == nil && os.PathSeparator == '\\' {
		// Nothing to assert about a volume-less path off Windows.
	}
	plain := t.TempDir()
	chdir(t, plain)
	if got, ok := rootViaCwd(`\definitely\not\here`); ok {
		t.Errorf("rootViaCwd invented %q outside a repository", got)
	}
}

func TestRootViaCwdAcceptsTheDirectoryItIsGiven(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("rootViaCwd only repairs Git-Bash-on-Windows MSYS paths; it is a deliberate no-op everywhere else")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	got, ok := rootViaCwd(wd)
	if !ok || got != wd {
		t.Errorf("rootViaCwd(%q) = %q, %v; want the same path accepted", wd, got, ok)
	}
}
