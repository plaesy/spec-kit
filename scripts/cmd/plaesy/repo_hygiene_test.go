package main

// The repository tracked a 14.9 MB Windows PE binary at its root for the whole
// of its life. `plaesy` is a build output — `make build` writes it, and it
// changes on every build, so its blob in history is 14 MB of nothing.
//
// It survived two changes that should have prevented it. `.gitignore` has
// listed `/plaesy` from the start, and the `clean` target's `git add -Af` —
// the thing that committed it — was itself fixed and documented. Neither helped,
// because **an ignore rule does not apply to a file that is already tracked**.
// The rule and the fix were both correct and both inert, which is the shape of
// defect that is easiest to believe is already handled.
//
// So this is not "did we write the ignore rule" but "is the rule doing
// anything": a binary is only ignored if it is also untracked, and the second
// half is the half that actually needed the intervention.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoBuildArtifactIsTracked reports any tracked file that is not text.
//
// The test is by construction — there is no list of forbidden names to drift
// out of date, and a future `plaesy` or `plaesy.exe` or a stray `go build`
// output in a subdirectory is caught the same day it is added.
func TestNoBuildArtifactIsTracked(t *testing.T) {
	root := corpusRoot(t)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}
	out, err := exec.Command("git", "-C", root, "ls-files", "-z").Output()
	if err != nil {
		t.Skipf("git ls-files failed (not a repository checkout?): %v", err)
	}

	paths := strings.Split(string(out), "\x00")
	checked := 0
	for _, rel := range paths {
		if rel == "" {
			continue
		}
		checked++
		raw, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if readErr != nil {
			// Tracked in the index but not in the working tree: still a defect
			// worth naming, and it must not be read as "binary, fine".
			t.Errorf("%s is tracked but not present in the working tree", rel)
			continue
		}
		if !looksBinary(raw) {
			continue
		}
		t.Errorf("%s is tracked and is a binary (%d bytes).\n"+
			"  Every clone downloads it and `git status` stays permanently dirty after each build.\n"+
			"  `git rm --cached %s` to untrack it, and make sure .gitignore covers it — an ignore\n"+
			"  rule does not apply to a file that is already tracked, which is how this survived a\n"+
			"  correct .gitignore and a correct `clean` target.", rel, len(raw), rel)
	}
	if checked == 0 {
		t.Fatal("git ls-files returned nothing; this would pass without checking anything")
	}
	t.Logf("%d tracked file(s) checked, none binary", checked)
}

// TestBuildOutputPathsAreIgnored is the other half. A binary that is untracked
// but not ignored is one `git add -A` away from being back in the index, which
// is precisely how the original got there — `make reset` stages the whole tree.
func TestBuildOutputPathsAreIgnored(t *testing.T) {
	root := corpusRoot(t)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Each is a real build output, and each is named differently by the two
	// ways this repository gets built: `make build` passes `-o ../plaesy` and
	// writes the extensionless name, while a bare `go build` inside the CLI
	// source directory lets Go append `.exe` on Windows. Both spellings need a
	// rule, or the second one is one `git add -A` from being tracked.
	for _, rel := range []string{"plaesy", "plaesy.exe", "plaesy-test"} {
		if out, err := exec.Command("git", "-C", root, "check-ignore", "-q", rel).CombinedOutput(); err != nil {
			t.Errorf("%q is not ignored by .gitignore. It is a build output, so being untracked is "+
				"not enough — the next `git add -A` re-adds it. `make build` writes %q; a bare "+
				"`go build` in scripts/cmd/plaesy writes %q on Windows.",
				rel, "plaesy", "plaesy.exe")
			_ = out
		}
	}
}

// looksBinary uses the same heuristic git itself uses: a NUL byte in the first
// 8000 bytes means the file is not text. A short file with no NUL is text.
func looksBinary(b []byte) bool {
	const sniff = 8000
	if len(b) > sniff {
		b = b[:sniff]
	}
	return bytes.IndexByte(b, 0) >= 0
}

// TestSessionStatePathsAreTrackable closes the same hole from the other side.
//
// `.plaesy/` holds two different kinds of file: generated artefacts
// (`instructions/`, `roles/`, `templates/`, `analysis/`, `scripts/`) and the
// session state a collaborator needs in version control (`context.md`,
// `memory.md`, `state.json`, `tasks/`). The global ignore excludes the whole
// directory, which is right for the first kind and silently wrong for the
// second.
//
// The failure is invisible from the working tree. A task file is tracked when
// it is created, so nothing looks wrong; the first time `plaesy tasks move`
// relocates it, git reports the old path as deleted, `git add -A` does not add
// the new one, and the file has left version control with no error anywhere.
// That is the tracked-binary defect again in a different costume — and it
// needs no fix to the file, only to the rule that was quietly excluding it.
//
// The check is that these paths are *not* ignored. An ignore rule that is
// correct for the generated majority must not classify the hand-maintained
// minority with it, and only git can say which side of the line a path falls
// on — including when the deciding rule lives in ~/.gitignore_global, which no
// test in this repository reads directly.
func TestSessionStatePathsAreTrackable(t *testing.T) {
	root := corpusRoot(t)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// The hand-maintained state, by construction rather than by a list of
	// today's files: a directory's README is tracked (it documents the
	// convention), and every task file in it must be trackable too, whichever
	// status directory `plaesy tasks move` put it in.
	tracked := []string{
		".plaesy/context.md",
		".plaesy/memory.md",
		".plaesy/state.json",
		".plaesy/tasks/README.md",
	}
	taskDirs, err := os.ReadDir(filepath.Join(root, ".plaesy", "tasks"))
	if err != nil {
		t.Fatalf("cannot read .plaesy/tasks: %v", err)
	}
	taskFiles := 0
	for _, d := range taskDirs {
		if !d.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(root, ".plaesy", "tasks", d.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			taskFiles++
			tracked = append(tracked, ".plaesy/tasks/"+d.Name()+"/"+e.Name())
		}
	}
	if taskFiles == 0 {
		t.Fatal("no task files found under .plaesy/tasks/; this would pass without checking anything")
	}

	for _, rel := range tracked {
		// --no-index is load-bearing. `git check-ignore` skips any path that is
		// already in the index, answering "is not ignored" for every tracked
		// file regardless of the rules — which is precisely the set this test
		// is about, so without it the test passes on the very defect it exists
		// to catch. It was found by mutation: deleting the `!.plaesy/`
		// negation left this test green.
		if out, err := exec.Command("git", "-C", root, "check-ignore", "-q", "--no-index", rel).CombinedOutput(); err == nil {
			t.Errorf("%s is ignored, so it cannot be committed.\n"+
				"  It is session state or a task file, not a generated artefact, and the next "+
				"  `git add -A` silently drops it. This is most visible on .plaesy/tasks/: a task "+
				"  is tracked when created, so moving it with `plaesy tasks move` leaves the old "+
				"  path deleted and the new one invisible — the file leaves version control with "+
				"  no error. ~/.gitignore_global excludes the whole `.plaesy` directory; .gitignore "+
				"  must re-include it with `!.plaesy/` (before, not after, the re-ignore rules) and "+
				"  then re-ignore only the generated subdirectories.",
				rel)
			_ = out
		}
	}
	t.Logf("%d session-state path(s) checked, %d of them task files", len(tracked), taskFiles)
}
