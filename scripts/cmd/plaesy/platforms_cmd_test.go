package main

// Tests for the `plaesy platforms` subcommands. These four used to live in
// config_cmd_test.go against `plaesy config detect|list|get-platform|show`,
// which is why these and the tests left behind in config_cmd_test.go read as
// near-duplicates: they exercise the same internal/config reads through two
// different shells. They stay separate because the two groups answer different
// questions — these are about the PLATFORM a project is on, the others about
// the config FILE that declares the platforms.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runPlatformsSub executes one `plaesy platforms <sub>` in-process against
// cfgPath and returns what it printed.
func runPlatformsSub(t *testing.T, cfgPath string, args ...string) (string, error) {
	t.Helper()
	full := append(args, "--config", cfgPath)
	return captureStdout(t, func() error {
		cmd := newPlatformsCmd()
		cmd.SetArgs(full)
		return cmd.Execute()
	})
}

// The flag decides the file, and it decides it for every subcommand. An explicit
// --config pointing somewhere else must be honoured, not merged with a default
// found from the repo root.
func TestPlatformsSubcommandsHonourAnExplicitConfigPath(t *testing.T) {
	dir := tempProject(t)
	other := filepath.Join(dir, "other-platform.json")
	body := `{"platforms": {"only_here": {"name": "Only Here", "provider": "Nobody",
		"category": "custom", "detection": ["only.marker"], "mapping": {"core": "ONLY.md"}}}}`
	if err := os.WriteFile(other, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runPlatformsSub(t, other, "list")
	if err != nil {
		t.Fatalf("platforms list failed: %v\n%s", err, out)
	}
	// The repo's real config has 20 platforms; this one has exactly one. If the
	// flag were ignored, the list would be the repo's.
	if strings.TrimSpace(out) != "only_here" {
		t.Errorf("platforms list read the wrong config: %q", strings.TrimSpace(out))
	}
}

// The output is consumed by scripts that index platform.json, so the order has
// to be the file's declaration order rather than Go map order.
func TestPlatformsListKeepsDeclarationOrder(t *testing.T) {
	dir := tempProject(t)
	path := filepath.Join(dir, "platform.json")
	body := `{"platforms": {"zulu": {"name": "Zulu", "detection": [], "mapping": {}},
		"alpha": {"name": "Alpha", "detection": [], "mapping": {}},
		"mike": {"name": "Mike", "detection": [], "mapping": {}}}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runPlatformsSub(t, path, "list")
	if err != nil {
		t.Fatalf("platforms list failed: %v\n%s", err, out)
	}
	if got := strings.Fields(strings.TrimSpace(out)); strings.Join(got, ",") != "zulu,alpha,mike" {
		t.Errorf("platforms list = %v, want the file's declaration order", got)
	}
}

func TestPlatformsGet(t *testing.T) {
	path := repoConfigPath()
	cases := []struct {
		args []string
		want string
	}{
		// A top-level field.
		{[]string{"get", "claude", "provider"}, "Anthropic"},
		// A mapping value, addressed through the "mapping." prefix.
		{[]string{"get", "claude", "mapping.core"}, "CLAUDE.md"},
	}
	for _, tc := range cases {
		out, err := runPlatformsSub(t, path, tc.args...)
		if err != nil {
			t.Fatalf("%v failed: %v\n%s", tc.args, err, out)
		}
		if got := strings.TrimSpace(out); got != tc.want {
			t.Errorf("%v = %q, want %q", tc.args, got, tc.want)
		}
	}
}

// An unknown platform must be an error. An empty line reads as "this platform
// has an empty provider", which is a different claim.
func TestPlatformsGetRejectsAnUnknownPlatform(t *testing.T) {
	_, err := runPlatformsSub(t, repoConfigPath(), "get", "mystery", "provider")
	if err == nil {
		t.Fatal("an unknown platform must be an error, not an empty line")
	}
	if !strings.Contains(err.Error(), "mystery") {
		t.Errorf("error %q does not name the platform the user typed", err)
	}
}

func TestPlatformsShow(t *testing.T) {
	out, err := runPlatformsSub(t, repoConfigPath(), "show", "claude")
	if err != nil {
		t.Fatalf("platforms show failed: %v\n%s", err, out)
	}
	for _, want := range []string{"claude", "Claude Code", "Anthropic"} {
		if !strings.Contains(out, want) {
			t.Errorf("platforms show does not print %q:\n%s", want, out)
		}
	}
}

// An unknown platform is reported, not crashed on, and the fields it cannot know
// say so rather than printing empty values that read as real data.
func TestPlatformsShowForAnUnknownPlatform(t *testing.T) {
	out, err := runPlatformsSub(t, repoConfigPath(), "show", "mystery")
	if err != nil {
		t.Fatalf("platforms show must not fail for an unknown platform: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Unknown") {
		t.Errorf("unknown fields are blank rather than marked unknown:\n%s", out)
	}
	if strings.Contains(out, "Anthropic") {
		t.Errorf("an unknown platform printed another platform's data:\n%s", out)
	}
}

// A missing config must fail the same way for every subcommand. One that prints
// help or an empty list instead looks like "this project uses no platforms".
func TestPlatformsSubcommandsAllFailTheSameWayOnAMissingFile(t *testing.T) {
	missing := filepath.Join(tempProject(t), "absent.json")
	subcommands := [][]string{
		{"detect"},
		{"list"},
		{"get", "claude", "provider"},
		{"show", "claude"},
	}
	for _, args := range subcommands {
		t.Run(args[0], func(t *testing.T) {
			out, err := runPlatformsSub(t, missing, args...)
			if err == nil {
				t.Fatalf("a missing config must be an error, not empty output\n%s", out)
			}
			if !strings.Contains(err.Error(), "absent.json") {
				t.Errorf("error %q does not name the config it could not read", err)
			}
			// A failed load must not also print a value: an empty line on stdout
			// is what a script consuming this output would treat as the answer.
			if strings.TrimSpace(out) != "" {
				t.Errorf("%s printed output despite failing: %q", args[0], out)
			}
		})
	}
}

// Detection resolves its patterns against the working directory, not against the
// directory the config was loaded from, and that is deliberate: the question
// `plaesy platforms detect` answers is "which AI platform is set up in the
// project I am standing in", and --config only says where to read the list of
// markers from. Pinning it here because it looks like a bug otherwise, and
// because changing it silently would break every caller that runs it from a
// project directory.
func TestPlatformsDetectReadsTheWorkingDirectory(t *testing.T) {
	// Resolve the config before moving: the path is relative to the cwd, and
	// this test's whole point is to move the cwd.
	cfg := absoluteRepoConfigPath(t)
	dir := tempProject(t)
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, dir)

	out, err := runPlatformsSub(t, cfg, "detect")
	if err != nil {
		t.Fatalf("platforms detect failed: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(out); got != "claude" {
		t.Errorf("platforms detect = %q, want claude (the cwd has .claude/commands)", got)
	}
}

func TestPlatformsDetectWithNoMarkersIsAnError(t *testing.T) {
	cfg := absoluteRepoConfigPath(t)
	withWorkingDirectory(t, tempProject(t))
	out, err := runPlatformsSub(t, cfg, "detect")
	if err == nil {
		t.Fatalf("no platform in the working directory must not report a platform\n%s", out)
	}
	if !strings.Contains(err.Error(), "no platform detected") {
		t.Errorf("error %q does not say that nothing was detected", err)
	}
	// A bare newline would be read as "the platform is the empty string".
	if strings.TrimSpace(out) != "" {
		t.Errorf("platforms detect printed a value alongside its error: %q", out)
	}
}

// `plaesy platforms` with no subcommand prints help rather than doing nothing,
// so the group is not a silent no-op for someone who guessed the wrong name.
func TestPlatformsGroupWithNoSubcommandShowsHelp(t *testing.T) {
	var out strings.Builder
	cmd := newPlatformsCmd()
	cmd.SetOut(&out)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("bare `plaesy platforms` failed: %v", err)
	}
	for _, sub := range []string{"detect", "list", "get", "show"} {
		if !strings.Contains(out.String(), sub) {
			t.Errorf("help does not list %q:\n%s", sub, out.String())
		}
	}
}

// Every subcommand must accept the same --config flag; one that silently ignores
// it would read the repo's config while the user believes they pointed at
// another.
func TestEveryPlatformsSubcommandAcceptsTheConfigFlag(t *testing.T) {
	for _, sub := range newPlatformsCmd().Commands() {
		if sub.Flags().Lookup("config") == nil && sub.InheritedFlags().Lookup("config") == nil {
			t.Errorf("`platforms %s` does not accept --config", sub.Name())
		}
		if sub.Short == "" && sub.Long == "" {
			t.Errorf("`platforms %s` has no description", sub.Name())
		}
	}
}

// `claude` is the canonical id platform.json declares, but `claude_code` is
// the legacy long form still kept as an alias. Every subcommand that takes a
// platform name has to bridge that, or a documented or muscle-memory
// invocation is the one that breaks. These are the spellings from
// docs/scripts/platforms.md.
func TestPlatformsSubcommandsAcceptTheDocumentedShorthand(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"show", "claude"}, "Claude Code"},
		{[]string{"get", "claude", "provider"}, "Anthropic"},
		{[]string{"get", "copilot", "mapping.core"}, "copilot-instructions.md"},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			out, err := runPlatformsSub(t, repoConfigPath(), tc.args...)
			if err != nil {
				t.Fatalf("the documented invocation failed: %v\n%s", err, out)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("output does not contain %q:\n%s", tc.want, out)
			}
		})
	}
}
