package main

// Tests for the `plaesy config` subcommands. internal/config is at 99.3%, so
// these are not about the reads — they are about the shell around them: which
// config file a flag resolves to, what a subcommand prints on the success path,
// and, most of all, what happens when the file is not there. A subcommand that
// prints an empty line because a lookup missed is indistinguishable from one
// that found an empty value.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// absoluteRepoConfigPath resolves the repo's platform.json before any test
// changes the working directory: a relative path is relative to the cwd, and
// these tests deliberately move it.
func absoluteRepoConfigPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(repoConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

// runConfigSub executes one `plaesy config <sub>` in-process against cfgPath and
// returns what it printed.
func runConfigSub(t *testing.T, cfgPath string, args ...string) (string, error) {
	t.Helper()
	full := append(args, "--config", cfgPath)
	return captureStdout(t, func() error {
		cmd := newConfigCmd()
		cmd.SetArgs(full)
		return cmd.Execute()
	})
}
func TestConfigGetMappingValueAndExcludes(t *testing.T) {
	path := repoConfigPath()

	out, err := runConfigSub(t, path, "get-mapping", "core", "core")
	if err != nil {
		t.Fatalf("get-mapping-value failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("get-mapping-value printed an empty line for a configured section")
	}

	// The excludes list of a section that has them, and of one that has none.
	// Both print successfully; the second prints nothing, which is correct and
	// must not be confused with an error.
	out, err = runConfigSub(t, path, "get-excludes", "instructions", "instructions")
	if err != nil {
		t.Fatalf("get-mapping-excludes failed: %v\n%s", err, out)
	}
	first := strings.TrimSpace(strings.SplitN(out, "\n", 2)[0])
	if first == "" {
		t.Errorf("a section with excludes printed nothing:\n%s", out)
	}
	out, err = runConfigSub(t, path, "get-excludes", "core", "core")
	if err != nil {
		t.Fatalf("get-mapping-excludes on an exclude-less section failed: %v\n%s", err, out)
	}
}

func TestConfigGetPlaesyStructure(t *testing.T) {
	path := repoConfigPath()
	for _, component := range []string{"base_directory", "core_directories", "project_directories"} {
		out, err := runConfigSub(t, path, "get-structure", component)
		if err != nil {
			t.Fatalf("get-plaesy-structure %s failed: %v\n%s", component, err, out)
		}
		if strings.TrimSpace(out) == "" {
			t.Errorf("get-plaesy-structure %s printed an empty line", component)
		}
	}
	if _, err := runConfigSub(t, path, "get-structure", "not_a_component"); err == nil {
		t.Error("an unknown structure component must be an error, not an empty line")
	}
}

func TestConfigGetCleanFilesAndDirs(t *testing.T) {
	path := repoConfigPath()
	out, err := runConfigSub(t, path, "get-clean-files")
	if err != nil {
		t.Fatalf("get-clean-files failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("get-clean-files printed nothing; the list is what the cleaner works from")
	}

	// With an explicit platform the answer must not depend on what is in the
	// working directory, which is the whole point of naming one.
	named, err := runConfigSub(t, path, "get-clean-dirs", "claude")
	if err != nil {
		t.Fatalf("get-clean-dirs failed: %v\n%s", err, named)
	}
	if strings.TrimSpace(named) == "" {
		t.Error("get-clean-dirs claude printed nothing")
	}
	if _, err := runConfigSub(t, path, "get-clean-dirs", "mystery"); err == nil {
		t.Error("an unknown platform must be an error, not an empty directory list")
	}
}
func TestConfigValidateAcceptsTheShippedConfig(t *testing.T) {
	out, err := runConfigSub(t, repoConfigPath(), "validate")
	if err != nil {
		t.Fatalf("the shipped platform.json must validate: %v\n%s", err, out)
	}
	if !strings.Contains(out, "valid") {
		t.Errorf("validate did not report success:\n%s", out)
	}
}

func TestConfigValidateRejectsAMalformedConfig(t *testing.T) {
	dir := tempProject(t)
	path := filepath.Join(dir, "platform.json")
	if err := os.WriteFile(path, []byte(`{"platforms": {`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runConfigSub(t, path, "validate")
	if err == nil {
		t.Fatalf("a truncated config must not validate\n%s", out)
	}
	if !strings.Contains(err.Error(), path) && !strings.Contains(err.Error(), "JSON") {
		t.Errorf("error %q does not say what was wrong or where", err)
	}
}

// Every subcommand shares one config-loading path, so a missing file has to
// surface the same way from all of them — and has to name the file it looked
// for, or the user cannot tell which of several --config values is wrong.
func TestConfigSubcommandsAllFailTheSameWayOnAMissingFile(t *testing.T) {
	missing := filepath.Join(tempProject(t), "absent.json")
	subcommands := [][]string{
		{"get-mapping", "core", "core"},
		{"get-excludes", "core", "core"},
		{"get-clean-dirs", "claude"},
		{"get-structure", "base_directory"},
		{"validate"},
	}
	for _, args := range subcommands {
		t.Run(args[0], func(t *testing.T) {
			out, err := runConfigSub(t, missing, args...)
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

// `plaesy config` with no subcommand prints help rather than doing nothing, so
// the group is not a silent no-op for someone who guessed the wrong name.
func TestConfigGroupWithNoSubcommandShowsHelp(t *testing.T) {
	var out strings.Builder
	cmd := newConfigCmd()
	cmd.SetOut(&out)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("bare `plaesy config` failed: %v", err)
	}
	for _, sub := range []string{"get-mapping", "get-excludes", "get-clean-dirs", "get-clean-files", "get-structure", "validate"} {
		if !strings.Contains(out.String(), sub) {
			t.Errorf("help does not list %q:\n%s", sub, out.String())
		}
	}
}

// Every subcommand must accept the same --config flag; one that silently ignores
// it would read the repo's config while the user believes they pointed at
// another.
func TestEveryConfigSubcommandAcceptsTheConfigFlag(t *testing.T) {
	for _, sub := range newConfigCmd().Commands() {
		if sub.Flags().Lookup("config") == nil && sub.InheritedFlags().Lookup("config") == nil {
			t.Errorf("`config %s` does not accept --config", sub.Name())
		}
		if !strings.Contains(sub.Long+sub.Short, sub.Name()) && sub.Short == "" {
			t.Errorf("`config %s` has no description", sub.Name())
		}
	}
}

// A shorthand must resolve to the same answer as the canonical id, or the two
// spellings describe different platforms.
func TestShorthandAndCanonicalIdAgree(t *testing.T) {
	short, err := runConfigSub(t, repoConfigPath(), "get-clean-dirs", "cursor")
	if err != nil {
		t.Fatalf("get-clean-dirs cursor failed: %v\n%s", err, short)
	}
	full, err := runConfigSub(t, repoConfigPath(), "get-clean-dirs", "cursor_ai")
	if err != nil {
		t.Fatalf("get-clean-dirs cursor_ai failed: %v\n%s", err, full)
	}
	if short != full {
		t.Errorf("cursor = %q but cursor_ai = %q; the alias resolved somewhere else", short, full)
	}
}
