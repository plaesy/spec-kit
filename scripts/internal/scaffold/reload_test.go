package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The load-bearing guarantee of `reload` is that it refreshes generated files
// without touching user data. Every test below is written to fail if that
// inverts, because the failure mode is silent: a clobbered memory.md looks
// exactly like a working reload until the user goes looking for a decision
// record that is no longer there.

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

// seedProject runs Init against a synthetic home, then writes recognisable
// content into every file reload is required to preserve.
func seedProject(t *testing.T) (home, target, base string) {
	t.Helper()
	home = fakeHome(t)
	target = newTarget(t)

	if err := Init(Options{TargetDir: target, AIPlatform: "claude_code", PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	cfg, err := LoadPlatformConfig(home)
	if err != nil {
		t.Fatal(err)
	}
	base = filepath.Join(target, cfg.Plaesy.BaseDirectory)

	// Simulate a user who has been working in the project for a while.
	writeFile(t, filepath.Join(base, "memory.md"), "MY DECISION RECORD — do not lose me\n")
	writeFile(t, filepath.Join(base, "context.md"), "MY SESSION STATE — do not lose me\n")
	writeFile(t, filepath.Join(base, "state.json"), `{"iteration":42,"mine":true}`)
	writeFile(t, filepath.Join(base, filepath.Join("specs", "001-thing", "spec.md")), "MY SPEC\n")
	writeFile(t, filepath.Join(base, filepath.Join("tasks", "todo", "T1.md")), "MY TASK\n")
	writeFile(t, filepath.Join(base, filepath.Join("analysis", "graph.md")), "MY ANALYSIS\n")
	writeFile(t, filepath.Join(base, filepath.Join("memory", "note.md")), "MY MEMORY NOTE\n")

	// A hand-written template with no counterpart in the sources: reload must
	// not delete it, because it never deletes anything.
	writeFile(t, filepath.Join(base, "templates", "mine-only.template.md"), "HAND WRITTEN\n")

	return home, target, base
}

func TestReloadPreservesUserData(t *testing.T) {
	homeRoot, target, base := seedProject(t)

	if _, err := Reload(ReloadOptions{TargetDir: target, PlaesyHome: homeRoot}); err != nil {
		t.Fatalf("Reload: %v", err)
	}

	for _, tc := range []struct{ path, want string }{
		{"memory.md", "MY DECISION RECORD"},
		{"context.md", "MY SESSION STATE"},
		{"state.json", `"iteration":42`},
		{filepath.Join("specs", "001-thing", "spec.md"), "MY SPEC"},
		{filepath.Join("tasks", "todo", "T1.md"), "MY TASK"},
		{filepath.Join("analysis", "graph.md"), "MY ANALYSIS"},
		{filepath.Join("memory", "note.md"), "MY MEMORY NOTE"},
		{filepath.Join("templates", "mine-only.template.md"), "HAND WRITTEN"},
	} {
		got := readFile(t, filepath.Join(base, tc.path))
		if !strings.Contains(got, tc.want) {
			t.Errorf("%s was clobbered by reload:\n got: %q\nwant it to contain: %q", tc.path, got, tc.want)
		}
	}
}

func TestReloadOverwritesDriftedGeneratedFiles(t *testing.T) {
	homeRoot, target, base := seedProject(t)

	// Simulate the actual bug this command exists to fix: the installed copy
	// is an OLD version of the source.
	src := filepath.Join(homeRoot, "instructions", "core.instructions.md")
	dst := filepath.Join(base, "instructions", "core.md")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture lacks the instruction this test mutates: %v", err)
	}
	stale := "STALE CONTENT FROM AN OLD VERSION\n"
	writeFile(t, dst, stale)

	report, err := Reload(ReloadOptions{TargetDir: target, PlaesyHome: homeRoot})
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}

	if got := readFile(t, dst); got == stale {
		t.Error("reload left the drifted instruction stale — that is the entire point of the command")
	}
	if !report.Changed() {
		t.Error("report.Changed() = false, but a stale file was repaired")
	}
	found := false
	for _, p := range report.Updated {
		if strings.HasSuffix(p, "core.md") {
			found = true
		}
	}
	if !found {
		t.Errorf("report.Updated does not mention core.md; got %v", report.Updated)
	}
}

func TestReloadIsANoOpOnAnUpToDateTree(t *testing.T) {
	homeRoot, target, _ := seedProject(t)

	if _, err := Reload(ReloadOptions{TargetDir: target, PlaesyHome: homeRoot}); err != nil {
		t.Fatalf("first Reload: %v", err)
	}
	report, err := Reload(ReloadOptions{TargetDir: target, PlaesyHome: homeRoot})
	if err != nil {
		t.Fatalf("second Reload: %v", err)
	}
	if report.Changed() {
		t.Errorf("second reload reported changes on an up-to-date tree: %d created, %d updated",
			len(report.Created), len(report.Updated))
	}
}

func TestReloadDryRunWritesNothing(t *testing.T) {
	homeRoot, target, base := seedProject(t)

	src := filepath.Join(homeRoot, "instructions", "core.instructions.md")
	dst := filepath.Join(base, "instructions", "core.md")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture lacks the instruction this test mutates: %v", err)
	}
	stale := "STALE\n"
	writeFile(t, dst, stale)

	report, err := Reload(ReloadOptions{TargetDir: target, PlaesyHome: homeRoot, DryRun: true})
	if err != nil {
		t.Fatalf("Reload --dry-run: %v", err)
	}
	if got := readFile(t, dst); got != stale {
		t.Errorf("dry run wrote to disk: %q", got)
	}
	if !report.DryRun {
		t.Error("report.DryRun = false on a dry run")
	}
	if !report.Changed() {
		t.Error("dry run did not report the pending update, so it cannot be used to preview one")
	}
	if !strings.Contains(report.Summary(), "Would ") {
		t.Errorf("Summary() = %q, want it to read as hypothetical", report.Summary())
	}
}

func TestReloadRefusesWithoutAnInitialisedProject(t *testing.T) {
	homeRoot := fakeHome(t)
	target := newTarget(t)

	_, err := Reload(ReloadOptions{TargetDir: target, PlaesyHome: homeRoot})
	if err == nil {
		t.Fatal("Reload on an uninitialised directory succeeded; it should tell the user to run init")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error %q does not point the user at `plaesy init`", err)
	}
}

func TestReloadRefreshesPlatformFilesWhenAsked(t *testing.T) {
	homeRoot, target, base := seedProject(t)

	// Drift an agent role, which lives in .plaesy/roles/ and is only reachable
	// via the platform path, then reload with --ai.
	role := filepath.Join(base, "roles", "architect.md")
	stale := "STALE ROLE\n"
	writeFile(t, role, stale)

	report, err := Reload(ReloadOptions{TargetDir: target, AIPlatform: "claude_code", PlaesyHome: homeRoot})
	if err != nil {
		t.Fatalf("Reload --ai: %v", err)
	}
	if got := readFile(t, role); got == stale {
		t.Error("reload --ai left the agent role stale")
	}
	found := false
	for _, p := range append(append([]string{}, report.Updated...), report.Created...) {
		if strings.HasSuffix(filepath.ToSlash(p), "roles/architect.md") {
			found = true
		}
	}
	if !found {
		t.Errorf("report does not mention roles/architect.md; updated=%v created=%v",
			report.Updated, report.Created)
	}
}

func TestReloadSkipsPlatformFilesWithoutAI(t *testing.T) {
	homeRoot, target, base := seedProject(t)

	role := filepath.Join(base, "roles", "architect.md")
	stale := "STALE ROLE\n"
	writeFile(t, role, stale)

	report, err := Reload(ReloadOptions{TargetDir: target, PlaesyHome: homeRoot})
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if got := readFile(t, role); got != stale {
		t.Error("reload without --ai refreshed platform files; that must require --ai")
	}
	if report.Changed() {
		t.Error("report.Changed() = true although the only drift was a platform file and --ai was not passed")
	}
}

func TestReloadRejectsUnknownPlatform(t *testing.T) {
	homeRoot, target, _ := seedProject(t)

	_, err := Reload(ReloadOptions{TargetDir: target, AIPlatform: "not_a_platform", PlaesyHome: homeRoot})
	if err == nil {
		t.Fatal("Reload accepted an unknown platform")
	}
	if !strings.Contains(err.Error(), "not_a_platform") {
		t.Errorf("error %q does not name the offending platform", err)
	}
}

// The remaining uncovered branches in Reload are `if err != nil` on stat and
// mkdir calls that fail for a reason other than "not there". Those need a
// filesystem that refuses access, which a test account does not get on
// Windows, and the coverage baseline already documents that class as debt.
// The three below are genuinely reachable and were untested.

func TestReloadDefaultsToTheCurrentDirectory(t *testing.T) {
	homeRoot, target, _ := seedProject(t)

	// Run from inside the project with no directory argument, which is what
	// `plaesy reload` with no args does.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(target); err != nil {
		t.Skipf("cannot chdir into the target: %v", err)
	}
	t.Cleanup(func() { os.Chdir(wd) })

	report, err := Reload(ReloadOptions{PlaesyHome: homeRoot, DryRun: true})
	if err != nil {
		t.Fatalf("Reload with no target directory: %v", err)
	}
	if !report.DryRun {
		t.Error("dry run flag was lost")
	}
}

func TestReloadFailsWhenTheHomeCannotBeFound(t *testing.T) {
	// A directory that is not a Plaesy root has no templates/ marker, so
	// FindHome rejects it. This is the "you ran plaesy from the wrong place"
	// path, and it must not be reported as a successful no-op refresh.
	notAHome := t.TempDir()
	target := newTarget(t)
	if err := Init(Options{TargetDir: target, AIPlatform: "claude_code", PlaesyHome: mustFakeHome(t)}); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if _, err := Reload(ReloadOptions{TargetDir: target, PlaesyHome: notAHome}); err == nil {
		t.Error("Reload accepted a PlaesyHome that is not a Plaesy root")
	}
}

func TestReloadSummaryReportsFailures(t *testing.T) {
	r := &ReloadReport{
		Created:   []string{"a"},
		Updated:   []string{"b", "c"},
		Unchanged: []string{"d"},
		Protected: []string{"memory.md"},
		Failed:    []string{"e"},
	}
	got := r.Summary()
	for _, want := range []string{"1 created", "2 updated", "1 unchanged", "1 preserved", "1 FAILED"} {
		if !strings.Contains(got, want) {
			t.Errorf("Summary() = %q, missing %q", got, want)
		}
	}
	if r.Changed() != true {
		t.Error("Changed() = false on a report with a creation")
	}
	if (&ReloadReport{}).Changed() {
		t.Error("Changed() = true on an empty report")
	}
	dry := (&ReloadReport{DryRun: true, Updated: []string{"x"}}).Summary()
	if !strings.Contains(dry, "Would") {
		t.Errorf("dry-run Summary() = %q, want it to read as hypothetical", dry)
	}
}

func mustFakeHome(t *testing.T) string {
	t.Helper()
	return fakeHome(t)
}

// ---------------------------------------------------------------- prune ----
//
// The bug these tests exist for: reload copies but never deletes, so a prompt
// renamed in prompts/ left a complete, still-invokable command in the mirror.
// The correction is not "delete files with no source" — that would eat the
// hand-written template reload is built to protect — but "delete files with no
// source, in the trees reload owns".

// staleMirrorFile reproduces the shape of the real incident: a renamed command
// leaving a nested orphan behind in the platform prompt mirror.
func staleMirrorFile(t *testing.T, target string) string {
	t.Helper()
	p := filepath.Join(target, ".claude", "commands", "create", "doc", "design.md")
	writeFile(t, p, "/create:doc:design command instructions\n")
	return p
}

func TestReloadPruneReportsStaleMirrorWithoutDeletingIt(t *testing.T) {
	homeRoot, target, _ := seedProject(t)
	stale := staleMirrorFile(t, target)

	report, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot,
		AIPlatform: "claude_code", Prune: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Stale) != 1 {
		t.Fatalf("Stale = %v, want exactly the renamed prompt", report.Stale)
	}
	// Report paths are slash-separated so they are stable across platforms.
	if report.Stale[0] != ".claude/commands/create/doc/design.md" {
		t.Errorf("Stale[0] = %q, want .claude/commands/create/doc/design.md", report.Stale[0])
	}
	if len(report.Pruned) != 0 {
		t.Errorf("Pruned = %v, want nothing deleted by --prune alone", report.Pruned)
	}
	if !exists(t, stale) {
		t.Error("--prune deleted the file; pruning must be a second, separate opt-in")
	}
}

func TestReloadPruneApplyRemovesRenamedMirrorAndItsEmptyDirs(t *testing.T) {
	homeRoot, target, _ := seedProject(t)
	stale := staleMirrorFile(t, target)

	report, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot,
		AIPlatform: "claude_code", Prune: true, PruneApply: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Pruned) != 1 {
		t.Fatalf("Pruned = %v, want the renamed prompt removed", report.Pruned)
	}
	if exists(t, stale) {
		t.Error("the stale mirror survived --prune-apply")
	}
	for _, dir := range []string{"doc", "create"} {
		p := filepath.Join(target, ".claude", "commands", dir)
		if _, err := os.Stat(p); err == nil {
			t.Errorf("%s/ survived; an empty directory tree is still a tree the runtime walks", dir)
		}
	}
}

func TestReloadPruneKeepsEveryLivePrompt(t *testing.T) {
	homeRoot, target, _ := seedProject(t)
	staleMirrorFile(t, target)

	if _, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot,
		AIPlatform: "claude_code", Prune: true, PruneApply: true,
	}); err != nil {
		t.Fatal(err)
	}

	mirror := filepath.Join(target, ".claude", "commands")
	for _, rel := range []string{"start.md", filepath.Join("nested", "deep.md")} {
		if !exists(t, filepath.Join(mirror, rel)) {
			t.Errorf("%s was pruned even though prompts/ still has it", rel)
		}
	}
	if !exists(t, filepath.Join(mirror, "nested")) {
		t.Error("the nested/ directory was removed even though it still holds a live prompt")
	}
}

func TestReloadPruneNeverTouchesUserData(t *testing.T) {
	homeRoot, target, base := seedProject(t)
	handWritten := filepath.Join(base, "templates", "mine-only.template.md")
	staleMirrorFile(t, target)

	if _, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot,
		AIPlatform: "claude_code", Prune: true, PruneApply: true,
	}); err != nil {
		t.Fatal(err)
	}

	if !exists(t, handWritten) {
		t.Error("a hand-written template with no counterpart in the sources was deleted; " +
			"that guarantee is the reason prune is scoped to owned trees")
	}
	for name, want := range map[string]string{
		"memory.md":  "MY DECISION RECORD",
		"context.md": "MY SESSION STATE",
	} {
		if got := readFile(t, filepath.Join(base, name)); !strings.HasPrefix(got, want) {
			t.Errorf("%s was clobbered by a pruning reload", name)
		}
	}
	for _, rel := range []string{
		filepath.Join("specs", "001-thing", "spec.md"),
		filepath.Join("tasks", "todo", "T1.md"),
		filepath.Join("analysis", "graph.md"),
		filepath.Join("memory", "note.md"),
	} {
		if !exists(t, filepath.Join(base, rel)) {
			t.Errorf("%s was removed by a pruning reload", rel)
		}
	}
}

func TestReloadPruneRemovesRenamedAgentRole(t *testing.T) {
	homeRoot, target, base := seedProject(t)
	ghost := filepath.Join(base, "roles", "retired.agents.md")
	writeFile(t, ghost, "an agent that no longer exists\n")

	report, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot, Prune: true, PruneApply: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Pruned) != 1 || !strings.HasSuffix(report.Pruned[0], "retired.agents.md") {
		t.Fatalf("Pruned = %v, want the retired agent role removed", report.Pruned)
	}
	for _, name := range []string{"architect.md", "reviewer.md"} {
		if !exists(t, filepath.Join(base, "roles", name)) {
			t.Errorf("%s was pruned even though agents/ still has it", name)
		}
	}
}

// The prompt mirror is pruned only for the platform the user named. Reload
// refuses to guess the platform when copying, and guessing it when deleting
// would prune a directory belonging to some other tool.
func TestReloadPruneIgnoresPromptMirrorWithoutAnExplicitPlatform(t *testing.T) {
	homeRoot, target, _ := seedProject(t)
	stale := staleMirrorFile(t, target)

	report, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot, Prune: true, PruneApply: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Pruned) != 0 {
		t.Errorf("Pruned = %v, want nothing: no platform was named", report.Pruned)
	}
	if !exists(t, stale) {
		t.Error("a prompt mirror was pruned without the platform being named")
	}
}

// .cursor/rules and .qoder/rules hold hand-written rules alongside copied
// prompts, so a file there is indistinguishable from a mirror of a prompt that
// has since been renamed. The config marks the platform prunable; the deny
// list still wins.
func TestReloadPruneRefusesSharedPromptDirectories(t *testing.T) {
	homeRoot, target, _ := seedProject(t)

	// A platform whose prompt destination is a rules directory the user also
	// writes to, marked prunable so the deny list is what stops the deletion.
	shared := strings.Replace(platformJSON,
		`  }
}`,
		`    ,
    "x_rules": {
      "name": "Shared Rules Tool",
      "mapping": {"core": ".x/rules/plaesy.mdc", "prompts": ".cursor/rules", "prune_prompts": "true"}
    }
  }
}`, 1)
	writeFile(t, filepath.Join(homeRoot, "scripts", "configs", "platform.json"), shared)

	handWritten := filepath.Join(target, ".cursor", "rules", "my-team-rules.md")
	writeFile(t, handWritten, "rules the team wrote by hand\n")

	report, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot,
		AIPlatform: "x_rules", Prune: true, PruneApply: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Pruned) != 0 {
		t.Errorf("Pruned = %v, want nothing: .cursor/rules is shared with hand-written rules", report.Pruned)
	}
	if !exists(t, handWritten) {
		t.Error("a hand-written file in a shared prompt directory was deleted")
	}
}

// The safety property that makes the rest of prune defensible: a source that
// cannot be read produces an empty expected set, and an empty expected set
// makes every file look like an orphan. Prune must skip, not delete.
func TestReloadPruneDeletesNothingWhenTheSourceIsGone(t *testing.T) {
	homeRoot, target, base := seedProject(t)
	ghost := filepath.Join(base, "roles", "retired.agents.md")
	writeFile(t, ghost, "an agent that no longer exists\n")

	if err := os.RemoveAll(filepath.Join(homeRoot, "agents")); err != nil {
		t.Fatal(err)
	}

	report, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot, Prune: true, PruneApply: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Pruned) != 0 {
		t.Errorf("Pruned = %v, want nothing: the source tree is unreadable", report.Pruned)
	}
	if !exists(t, ghost) {
		t.Fatal("a missing source directory caused a deletion; this is the failure mode that makes " +
			"path-based pruning unsafe without a source check")
	}
}

// An empty source directory is the same hazard as a missing one: the expected
// set is empty either way, so every file in the tree would look like an orphan.
func TestReloadPruneDeletesNothingWhenTheSourceIsEmpty(t *testing.T) {
	homeRoot, target, base := seedProject(t)
	ghost := filepath.Join(base, "roles", "retired.agents.md")
	writeFile(t, ghost, "an agent that no longer exists\n")

	if err := os.RemoveAll(filepath.Join(homeRoot, "agents")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(homeRoot, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}

	report, err := Reload(ReloadOptions{
		TargetDir: target, PlaesyHome: homeRoot, Prune: true, PruneApply: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Pruned) != 0 || !exists(t, ghost) {
		t.Fatalf("an empty source directory caused %v to be deleted", report.Pruned)
	}
}

// Every platform must be explicitly classified. A new platform whose prompts
// directory is exclusive is prunable; one whose directory is shared is not.
// Falling through to a default either way is how a wrong answer becomes an
// unrecoverable deletion, so the config is not allowed to leave the question
// open.
func TestEveryPlatformClassifiesItsPromptDirectory(t *testing.T) {
	home := repoRoot(t)
	cfg, err := LoadPlatformConfig(home)
	if err != nil {
		t.Fatalf("loading the real platform.json: %v", err)
	}
	if len(cfg.Platforms) == 0 {
		t.Fatal("platform.json declared no platforms; the test is not looking at the real file")
	}

	for id, entry := range cfg.Platforms {
		dest, ok := entry.Mapping["prompts"]
		if !ok || dest == "" {
			continue
		}
		flag, flagged := entry.Mapping["prune_prompts"]
		_, denied := sharedPromptDirs[dest]
		if flagged == denied {
			t.Errorf("platform %q maps prompts to %q and is %s; every prompt directory must be "+
				"either shared (and never pruned) or exclusive with \"prune_prompts\": \"true\"",
				id, dest, pruneClassification(flagged, denied))
		}
		if flagged && flag != "true" {
			t.Errorf("platform %q has prune_prompts = %q, want \"true\"", id, flag)
		}
	}
}

func pruneClassification(flagged, denied bool) string {
	switch {
	case flagged:
		return "marked prunable, but that directory is on the shared deny list"
	case denied:
		return "on the shared deny list without saying so"
	default:
		return "classified as neither"
	}
}
