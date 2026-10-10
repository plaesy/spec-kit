package installer

// Which rc file a user is told to edit is the part of the PATH instructions
// that can be wrong, and a wrong answer is the worst kind: adding a line to
// ~/.profile as a zsh user produces no error and no effect. The decision
// depends on $SHELL, so it was untested on the Windows CI runs this package's
// other tests all target — the branch was only ever exercised on a developer's
// machine, by whichever shell happened to be theirs.

import (
	"strings"
	"testing"
)

func TestUnixPathInstructionsPicksTheRCTheUsersShellReads(t *testing.T) {
	const dir = "/home/me/.local/bin"
	cases := []struct {
		name  string
		shell string
		want  string
	}{
		// zsh does not read ~/.profile for interactive shells, so a user told
		// to edit it gets a PATH change that never takes effect.
		{"zsh", "/bin/zsh", "~/.zshrc"},
		{"bash", "/usr/bin/bash", "~/.bashrc (or ~/.bash_profile)"},
		// fish and csh read neither, and ~/.profile is the documented fallback
		// rather than a claim that it will definitely work.
		{"fish", "/usr/bin/fish", "~/.profile"},
		{"unknown shell", "", "~/.profile"},
		// SHELL is a path, so the match has to be on the name, not the whole
		// value — a test asserting the substring keeps that honest.
		{"zsh with a versioned path", "/opt/homebrew/bin/zsh-5.9", "~/.zshrc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := unixPathInstructions(dir, tc.shell)
			if !strings.Contains(got, tc.want) {
				t.Errorf("unixPathInstructions(%q) did not point at %s:\n%s",
					tc.shell, tc.want, got)
			}
			// Whatever the shell, the line the user copies has to contain the
			// directory, or the instruction is not actionable.
			if !strings.Contains(got, "export PATH=\""+dir+":$PATH\"") {
				t.Errorf("instruction does not contain an export line for %s:\n%s", dir, got)
			}
		})
	}
}

func TestPathInstructionsAlwaysNamesTheDirectory(t *testing.T) {
	// The OS branch is chosen by runtime.GOOS, so only one is reachable per
	// platform. Asserting the directory appears is what holds on both.
	const dir = "some-dir-that-must-appear"
	if got := PathInstructions(dir); !strings.Contains(got, dir) {
		t.Errorf("PathInstructions(%q) does not mention the directory:\n%s", dir, got)
	}
}
