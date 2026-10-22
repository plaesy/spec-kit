package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/scaffold"
	"github.com/spf13/cobra"
)

// newReloadCmd is the whole user-facing surface of `plaesy reload`, and the
// parts most worth protecting are its output and its exit code: a refresh that
// half-failed must not report success, and a dry run must not claim to have
// written anything.

// runReload executes the command and returns what the user would have seen.
//
// The logger writes to os.Stdout directly rather than through cobra's writers,
// like every other command in this repo, so capturing the buffer cobra hands
// back would assert nothing. Redirecting the real file descriptor is the only
// way to check the output contract.
func runReload(t *testing.T, args ...string) (string, error) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w

	// Drain the pipe concurrently. A reload of this repository writes far more
	// than the 64 KiB pipe buffer holds, so a reader that only runs after
	// Execute() returns deadlocks the test instead of failing it.
	captured := make(chan string, 1)
	go func() {
		var b bytes.Buffer
		io.Copy(&b, r)
		captured <- b.String()
	}()

	root := &cobra.Command{Use: "plaesy", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(newReloadCmd())
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(append([]string{"reload"}, args...))

	execErr := root.Execute()

	os.Stdout = orig
	w.Close()
	out := <-captured
	r.Close()

	return out + buf.String(), execErr
}

func TestReloadCommandReportsWhatItDid(t *testing.T) {
	// A directory that was never initialised: the command must fail and point
	// at init rather than reporting a clean, empty refresh.
	empty := t.TempDir()
	_, err := runReload(t, empty)
	if err == nil {
		t.Fatal("reload on an uninitialised directory reported success")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error %q does not point the user at `plaesy init`", err)
	}
}

func TestReloadCommandDryRunSaysItWould(t *testing.T) {
	// The CLI is exercised through the real repo's own .plaesy/ so the command
	// has something to report on. Only --dry-run is used, so nothing is
	// written.
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root := filepath.Dir(filepath.Dir(docs))
	if _, err := os.Stat(filepath.Join(root, ".plaesy")); err != nil {
		t.Skip(".plaesy not present")
	}

	out, err := runReload(t, root, "--dry-run", "--plaesy-home", root)
	if err != nil {
		t.Fatalf("reload --dry-run: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Summary:") {
		t.Errorf("output has no summary line:\n%s", out)
	}
	if !strings.Contains(out, "preserved") {
		t.Errorf("output does not report the preserved user files:\n%s", out)
	}
}

func TestReloadCommandRejectsAnUnknownPlatform(t *testing.T) {
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root := filepath.Dir(filepath.Dir(docs))

	_, err := runReload(t, root, "--ai", "not_a_platform", "--plaesy-home", root, "--dry-run")
	if err == nil {
		t.Fatal("reload accepted an unknown platform")
	}
	if !strings.Contains(err.Error(), "not_a_platform") {
		t.Errorf("error %q does not name the offending platform", err)
	}
}

func TestReloadCommandHelpStatesWhatItProtects(t *testing.T) {
	// The protection guarantee is the reason to run this command at all, so it
	// has to be discoverable from --help rather than only from the source.
	out, err := runReload(t, "--help")
	if err != nil {
		t.Fatalf("reload --help: %v", err)
	}
	for _, want := range []string{"memory.md", "context.md", "state.json", "dry-run", "--ai"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help does not mention %q:\n%s", want, out)
		}
	}
}

// --prune is the one part of reload that deletes, and the flags are the only
// thing standing between a stale mirror and a deleted hand-written file. The
// pair is exercised here as the user would type it, against a real project
// built from this repository's sources, because the guarantee worth testing is
// the one the command actually makes.

func repoRootForReload(t *testing.T) string {
	t.Helper()
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	return filepath.Dir(filepath.Dir(docs))
}

// projectWithStaleMirror initialises a throwaway project and leaves behind the
// exact shape of the incident that motivated prune: a prompt renamed in the
// sources, with the old copy still sitting in the mirror.
func projectWithStaleMirror(t *testing.T) (target, home, stale string) {
	t.Helper()
	home = repoRootForReload(t)
	target = tempProject(t)

	if err := scaffold.Init(scaffold.Options{
		TargetDir: target, AIPlatform: "kilo", PlaesyHome: home,
	}); err != nil {
		t.Fatalf("Init: %v", err)
	}

	stale = filepath.Join(target, ".kilo", "commands", "create", "doc", "design.md")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("/create:doc:design command instructions\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return target, home, stale
}

func TestReloadCommandPruneListsStaleFilesWithoutDeletingThem(t *testing.T) {
	target, home, stale := projectWithStaleMirror(t)

	out, err := runReload(t, target, "--ai", "kilo", "--plaesy-home", home, "--prune")
	if err != nil {
		t.Fatalf("reload --prune: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Stale") {
		t.Errorf("--prune did not report the stale section:\n%s", out)
	}
	// Report paths are slash-separated so the output is stable across
	// platforms and greppable.
	if !strings.Contains(out, ".kilo/commands/create/doc/design.md") {
		t.Errorf("--prune did not name the orphaned file:\n%s", out)
	}
	if !strings.Contains(out, "NOT removed") {
		t.Errorf("--prune output does not say the files were kept:\n%s", out)
	}
	if _, err := os.Stat(stale); err != nil {
		t.Error("--prune deleted the file; deletion must be a second, separate opt-in")
	}
}

func TestReloadCommandPruneApplyDeletesStaleFiles(t *testing.T) {
	target, home, stale := projectWithStaleMirror(t)

	out, err := runReload(t, target, "--ai", "kilo", "--plaesy-home", home, "--prune", "--prune-apply")
	if err != nil {
		t.Fatalf("reload --prune --prune-apply: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Pruned") {
		t.Errorf("--prune-apply did not report the pruned section:\n%s", out)
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("the stale mirror survived --prune-apply")
	}
	if !strings.Contains(out, "1 pruned") {
		t.Errorf("summary does not count the pruned file:\n%s", out)
	}
}

func TestReloadCommandPruneApplyWithoutPruneDeletesNothing(t *testing.T) {
	target, home, stale := projectWithStaleMirror(t)

	if _, err := runReload(t, target, "--ai", "kilo", "--plaesy-home", home, "--prune-apply"); err != nil {
		t.Fatalf("reload --prune-apply: %v", err)
	}
	if _, err := os.Stat(stale); err != nil {
		t.Error("--prune-apply deleted a file without --prune; the two must be opt-in separately")
	}
}
