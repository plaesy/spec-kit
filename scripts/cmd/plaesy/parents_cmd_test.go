package main

// Tests for the command wrappers themselves — the parent commands that list
// rather than delegate. internal/featurepath and internal/detectstack are
// covered on their own; what is checked here is that the CLI surface reaches
// them the way the reference documents, and that a bare `plaesy features`
// lists instead of printing help.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// initRepo makes dir a git repository with one commit, so common.GetRepoRoot
// — which shells out to `git rev-parse --show-toplevel` — can resolve it. A
// .gitignore is not enough: there is no repository there until git init runs.
func initRepo(t *testing.T, dir string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "test@example.invalid"},
		{"config", "user.name", "Test"},
		{"commit", "-q", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

// seedSpecsDir lays down <dir>/.plaesy/specs/<name>/<doc> and makes <dir> look
// like a repository root to git, so `plaesy features` has something to list.
func seedSpecsDir(t *testing.T, dir string, features map[string][]string) {
	t.Helper()
	for name, docs := range features {
		for _, doc := range docs {
			path := filepath.Join(dir, ".plaesy", "specs", name, doc)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	initRepo(t, dir)
}

// A parent whose only job is to print a list has to do that with no
// subcommand. Printing help instead — cobra's default for a group — makes
// `plaesy features` look like a group rather than the listing it documents
// itself to be, and it is the one command in this tree a newcomer runs first.
func TestBareFeaturesListsTheProjectFeatures(t *testing.T) {
	dir := tempProject(t)
	seedSpecsDir(t, dir, map[string][]string{
		"001-first":  {"spec.md", "plan.md"},
		"002-second": {"spec.md"},
	})
	withWorkingDirectory(t, dir)

	out, err := captureStdout(t, func() error {
		cmd := newFeaturesCmd()
		cmd.SetArgs(nil)
		return cmd.Execute()
	})
	if err != nil {
		t.Fatalf("bare `plaesy features` failed: %v", err)
	}
	for _, want := range []string{"2 feature(s)", "001-first", "002-second", "spec.md, plan.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("bare `plaesy features` output does not contain %q:\n%s", want, out)
		}
	}
}

// The same command on a project that has created nothing yet must say so, and
// must name the command that fixes it.
func TestBareFeaturesOnAFreshProjectExplainsItself(t *testing.T) {
	dir := tempProject(t)
	seedSpecsDir(t, dir, nil)
	withWorkingDirectory(t, dir)

	out, err := captureStdout(t, func() error {
		cmd := newFeaturesCmd()
		cmd.SetArgs(nil)
		return cmd.Execute()
	})
	if err != nil {
		t.Fatalf("bare `plaesy features` failed: %v", err)
	}
	if !strings.Contains(out, "No features found") {
		t.Errorf("a fresh project must be told it has no features:\n%s", out)
	}
	if !strings.Contains(out, "features create") {
		t.Errorf("the empty listing must name the command that fixes it:\n%s", out)
	}
}

// `stack detect` with an explicit root reads that tree's mapping.json and
// succeeds. With neither an explicit root, $PLAESY_ROOT, nor a repository, it
// must say which of the three it could not resolve.
func TestStackDetectUsesAnExplicitPlaesyRoot(t *testing.T) {
	dir := tempProject(t)
	withWorkingDirectory(t, dir)

	out, err := captureStdout(t, func() error {
		cmd := newStackDetectCmd()
		cmd.SetArgs([]string{"--plaesy-root", dir})
		return cmd.Execute()
	})
	if err != nil {
		t.Fatalf("stack detect with an explicit root failed: %v\n%s", err, out)
	}
}

// A PLAESY_ROOT with no instructions/mapping.json yields no matches and exit 0.
// That is the behaviour both shell twins had, and Detect keeps it deliberately,
// so it is pinned here rather than left to look accidental: the risk is that a
// wrong --plaesy-root looks exactly like a project with no detected stack.
func TestStackDetectWithNoMappingJSONIsSilentAndSucceeds(t *testing.T) {
	dir := tempProject(t)
	withWorkingDirectory(t, dir)
	t.Setenv("PLAESY_ROOT", filepath.Join(dir, "no-such-root"))

	out, err := captureStdout(t, func() error {
		cmd := newStackDetectCmd()
		cmd.SetArgs(nil)
		return cmd.Execute()
	})
	if err != nil {
		t.Fatalf("a missing mapping.json is not fatal by design: %v\n%s", err, out)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("no mapping.json should select no instruction files, got %q", out)
	}
}

// seedPlaesyRoot writes a minimal but real instructions/ tree: a mapping.json
// whose cross_cutting category selects one instruction by a marker filename,
// plus the two instruction files it can therefore reach. detect-stack matches
// on filenames and manifest keyword content, so a Go manifest is what makes
// this tree look like a Go project.
func seedPlaesyRoot(t *testing.T, dir string) {
	t.Helper()
	instrDir := filepath.Join(dir, "instructions")
	if err := os.MkdirAll(instrDir, 0o755); err != nil {
		t.Fatal(err)
	}

	mapping := `{
  "description": "test",
  "version": "1.0.0",
  "mappings": {
    "always_load": [],
    "scope_load": [],
    "frameworks": {},
    "languages": {
      "go": {
        "file": "go.instructions.md",
        "keywords": ["package main"],
        "filenames": ["go.mod"]
      }
    },
    "cross_cutting": {
      "git": {
        "file": "git.instructions.md",
        "keywords": ["version control"]
      }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(instrDir, "mapping.json"), []byte(mapping), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"go.instructions.md":  "# Go\n\npackage main is the entrypoint.\n",
		"git.instructions.md": "# Git\n\nversion control is explicit.\n",
	} {
		if err := os.WriteFile(filepath.Join(instrDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// The whole reason --install exists: detection alone leaves the file off disk,
// so a project that has run init and detected its stack still cannot load the
// instruction. This pins that the flag actually puts the file there.
func TestStackDetectInstallCopiesDetectedInstructions(t *testing.T) {
	root := tempProject(t)
	seedPlaesyRoot(t, root)

	target := tempProject(t)
	if err := os.WriteFile(filepath.Join(target, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, target)

	out, err := captureStdout(t, func() error {
		cmd := newStackDetectCmd()
		cmd.SetArgs([]string{target, "--plaesy-root", root, "--install"})
		return cmd.Execute()
	})
	if err != nil {
		t.Fatalf("stack detect --install failed: %v\n%s", err, out)
	}

	// The destination name must drop ".instructions", matching
	// scaffold.copyInstructions, so the two install paths cannot disagree
	// about what a file is called.
	dst := filepath.Join(target, ".plaesy", "instructions", "go.md")
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("detected instruction was not installed at %s: %v\n%s", dst, err, out)
	}
	if !strings.Contains(string(got), "package main") {
		t.Errorf("installed file does not carry the source content:\n%s", got)
	}

	// git.instructions.md must NOT be installed: nothing in the target
	// mentions version control, so it is not a match. Installing the whole
	// mapping would make --install indistinguishable from init.
	if _, err := os.Stat(filepath.Join(target, ".plaesy", "instructions", "git.md")); err == nil {
		t.Errorf("a non-matching instruction was installed; --install must install only what detected:\n%s", out)
	}
}

// --install is additive and runs repeatedly as a project grows. It must never
// clobber an installed file: a user may have edited the constitution or a bar
// in place, and a silent overwrite on the next run loses that edit.
func TestStackDetectInstallNeverOverwritesAnExistingFile(t *testing.T) {
	root := tempProject(t)
	seedPlaesyRoot(t, root)

	target := tempProject(t)
	if err := os.WriteFile(filepath.Join(target, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	destDir := filepath.Join(target, ".plaesy", "instructions")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	edited := "# Go\n\nHAND-EDITED: do not lose me.\n"
	if err := os.WriteFile(filepath.Join(destDir, "go.md"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, target)

	out, err := captureStdout(t, func() error {
		cmd := newStackDetectCmd()
		cmd.SetArgs([]string{target, "--plaesy-root", root, "--install"})
		return cmd.Execute()
	})
	if err != nil {
		t.Fatalf("stack detect --install failed: %v\n%s", err, out)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "go.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != edited {
		t.Errorf("a hand-edited installed file was overwritten\n got: %s\nwant: %s", got, edited)
	}
	if !strings.Contains(out, "exists, kept") {
		t.Errorf("a skipped file must be reported, not silently dropped:\n%s", out)
	}
}

// Without --install the command must stay exactly as it was: a listing, with
// no .plaesy/ tree created in the target. Detection reporting itself as a
// side effect of printing is the bug this flag was added to fix.
func TestStackDetectWithoutInstallPrintsAndWritesNothing(t *testing.T) {
	root := tempProject(t)
	seedPlaesyRoot(t, root)

	target := tempProject(t)
	if err := os.WriteFile(filepath.Join(target, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, target)

	out, err := captureStdout(t, func() error {
		cmd := newStackDetectCmd()
		cmd.SetArgs([]string{target, "--plaesy-root", root})
		return cmd.Execute()
	})
	if err != nil {
		t.Fatalf("stack detect failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "go.instructions.md") {
		t.Errorf("plain stack detect must still list the match:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(target, ".plaesy")); err == nil {
		t.Errorf("plain `stack detect` created %s; it must only print", filepath.Join(target, ".plaesy"))
	}
}

// Cobra skips ValidateArgs when a command is not Runnable, so a noun parent
// with no Run prints its help for a misspelled subcommand and exits 0. The
// help text then reads as a result: `PLATFORM=$(plaesy config detect)` captured
// the whole help block and the calling script carried on. Every parent has to
// be Runnable for Args: cobra.NoArgs to do anything, which is what
// showHelpWhenBare in registry.go arranges.
func TestNounParentsRejectAMisspelledSubcommand(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"features", newFeaturesCmd},
		{"context", newContextCmd},
		{"platforms", newPlatformsCmd},
		{"config", newConfigCmd},
		{"stack", newStackCmd},
		{"tasks", newTasksCmd},
		{"images", newImagesCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			cmd := tc.cmd()
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{"definitely-not-a-subcommand"})

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("a misspelled subcommand must be an error, not help text (exit 0)\n%s", out.String())
			}
			if !strings.Contains(err.Error(), "unknown command") {
				t.Errorf("error %q does not say the command is unknown", err)
			}
			if !strings.Contains(err.Error(), "definitely-not-a-subcommand") {
				t.Errorf("error %q does not name what was typed", err)
			}
			// cobra pairs the error with the usage block. That is wanted — the
			// user is about to be told which spellings exist. What must not
			// happen is the help text standing in for a result, which is what
			// the err == nil case above is checking.
		})
	}
}

// The pre-rename spellings have to fail the same way. A user upgrading types the
// command they know, and "unknown command" is the only honest answer to it —
// help text at exit 0 is a result they would go on to use.
func TestRemovedSubcommandSpellingsAreRejectedNotEchoed(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  func() *cobra.Command
		args []string
	}{
		{"config detect", newConfigCmd, []string{"detect"}},
		{"config list", newConfigCmd, []string{"list"}},
		{"config show", newConfigCmd, []string{"show", "claude_code"}},
		{"config get-platform", newConfigCmd, []string{"get-platform", "claude_code"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			cmd := tc.cmd()
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tc.args)

			if err := cmd.Execute(); err == nil {
				t.Errorf("`plaesy %s` was removed but still resolves (exit 0):\n%s", tc.name, out.String())
			}
		})
	}
}

// The parents share one shape: run them bare and they list or print help, never
// silence. A group that prints nothing is indistinguishable from one that did
// nothing, which is exactly how a mistyped command used to look.
func TestEveryNounParentIsUsefulWhenRunBare(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  func() *cobra.Command
		want string
	}{
		{"context", newContextCmd, "update"},
		{"platforms", newPlatformsCmd, "detect"},
		{"config", newConfigCmd, "validate"},
		{"stack", newStackCmd, "detect"},
		{"tasks", newTasksCmd, "list"},
		{"images", newImagesCmd, "create"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			cmd := tc.cmd()
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(nil)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("bare `plaesy %s` failed: %v", tc.name, err)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Errorf("bare `plaesy %s` does not mention %q:\n%s", tc.name, tc.want, out.String())
			}
		})
	}
}
