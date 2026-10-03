package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// fakeHome builds a synthetic Plaesy repo root: the directories and files
// FindHome looks for, plus a platform.json and mapping.json. Tests then run
// entirely inside t.TempDir(), which is the only way to be sure `plaesy init`
// never writes into the real repository.
func fakeHome(t *testing.T) string {
	t.Helper()
	home := longTempDir(t, "scaffold-home")

	write := func(rel, content string) {
		path := filepath.Join(home, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	mkdir := func(rel string) {
		if err := os.MkdirAll(filepath.Join(home, filepath.FromSlash(rel)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
	}

	// The marker FindHome looks for.
	mkdir("templates")
	write("templates/state.template.json", `{"initialized_at": "[TIMESTAMP]"}`)
	write("templates/context.template.md", "# Context\n")
	write("templates/README.md", "templates readme\n")
	write("templates/spec.template.md", "spec template\n")
	write("templates/sdd.template.md", "sdd\n")
	write("templates/decisions.template.md", "decisions index\n")
	// An extension the template copy must ignore.
	write("templates/notes.txt", "not a template\n")

	write("instructions/mapping.json", `{
  "mappings": {
    "always_load": ["core.instructions.md", "scoped.instructions.md"],
    "scope_load": ["scoped.instructions.md"]
  }
}`)
	write("instructions/core.instructions.md", "core instructions\n")
	write("instructions/scoped.instructions.md", "scoped instructions\n")
	write("instructions/agents.instructions.md", "agents core file\n")
	write("instructions/unlisted.instructions.md", "not in always_load\n")

	write("agents/architect.agents.md", "agent architect\n")
	write("agents/reviewer.agents.md", "agent reviewer\n")
	write("agents/notes.md", "not an agent file\n")

	write("prompts/start.md", "prompt start\n")
	write("prompts/nested/deep.md", "prompt deep\n")
	write("prompts/ignored.txt", "not a prompt\n")

	write("checklists/requirements.md", "checklist\n")
	write("checklists/notes.txt", "not a checklist\n")

	write("scripts/bash/tool.sh", "#!/bin/sh\n")
	write("scripts/configs/platform.json", platformJSON)
	write("scripts/powershell/tool.ps1", "# tool\n")

	return home
}

const platformJSON = `{
  "plaesy": {
    "base_directory": ".plaesy",
    "core_directories": ["memory", "decisions", "instructions", "tasks"],
    "project_directories": [".plaesy/specs"],
    "mapping": {
      "core": {"value": "instructions/agents.instructions.md"},
      "prompts": {"value": "prompts/*"}
    }
  },
  "platforms": {
    "claude": {
      "name": "Claude Code",
      "mapping": {"core": "CLAUDE.md", "prompts": ".claude/commands", "prune_prompts": "true"}
    },
    "github_copilot": {
      "name": "GitHub Copilot",
      "mapping": {"core": ".github/copilot-instructions.md", "prompts": ".github/prompts", "prune_prompts": "true"}
    }
  }
}`

// longTempDir returns a fresh temp directory under a deliberately long path.
// t.TempDir() can hand back a short path (TMP is under C:\msys64\tmp here),
// which overflows the legacy MAX_PATH limit once a scaffold nests
// .plaesy/instructions/scripts/configs under it.
func longTempDir(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "plaesy-test", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	return dir
}

func newTarget(t *testing.T) string {
	t.Helper()
	return longTempDir(t, "project")
}

func exists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	return err == nil
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read %s: %v", path, err)
		return
	}
	if string(data) != want {
		t.Errorf("%s = %q, want %q", path, data, want)
	}
}

// listing returns every path under root, relative and slash-separated, so a
// test can assert on the whole shape of a scaffold at once.
func listing(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if info.IsDir() {
			rel += "/"
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

func hasEntry(list []string, want string) bool {
	for _, e := range list {
		if e == want {
			return true
		}
	}
	return false
}

func TestInitBuildsTheStructure(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)

	if err := Init(Options{TargetDir: target, AIPlatform: "claude", PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}

	got := listing(t, target)
	want := []string{
		// Core directories from platform.json.
		".plaesy/memory/",
		".plaesy/decisions/",
		".plaesy/instructions/",
		".plaesy/tasks/",
		// Project directory from platform.json.
		".plaesy/specs/",
		// always_load, with ".instructions" stripped from the name.
		".plaesy/instructions/core.md",
		".plaesy/instructions/scoped.md",
		// Task management.
		".plaesy/tasks/backlog/",
		".plaesy/tasks/todo/",
		".plaesy/tasks/doing/",
		".plaesy/tasks/done/",
		".plaesy/tasks/blocked/",
		".plaesy/tasks/README.md",
		// Memory and analysis.
		".plaesy/context.md",
		".plaesy/analysis/",
		// Decisions index.
		".plaesy/decisions.md",
		// Loop state, with the placeholder replaced.
		".plaesy/state.json",
		// Templates, checklists, scripts.
		".plaesy/templates/state.template.json",
		".plaesy/templates/spec.template.md",
		".plaesy/templates/decisions.template.md",
		".plaesy/checklists/requirements.md",
		".plaesy/scripts/bash/tool.sh",
		".plaesy/scripts/configs/platform.json",
		".plaesy/scripts/powershell/tool.ps1",
		// Agents land in roles/ with ".agents" stripped.
		".plaesy/roles/architect.md",
		".plaesy/roles/reviewer.md",
		// Platform mapping: core file and prompts.
		"CLAUDE.md",
		".claude/commands/start.md",
		".claude/commands/nested/deep.md",
	}
	for _, w := range want {
		if !hasEntry(got, w) {
			t.Errorf("missing %s\ngot:\n  %s", w, strings.Join(got, "\n  "))
		}
	}

	// Files the source does not list must not be copied: the template copy is
	// extension-filtered, prompts are .md only, and the instruction list comes
	// from mapping.json rather than from whatever is in the directory.
	for _, w := range []string{
		".plaesy/templates/notes.txt",
		".plaesy/checklists/notes.txt",
		".plaesy/roles/notes.md",
		".plaesy/instructions/unlisted.md",
		".claude/commands/ignored.txt",
	} {
		if hasEntry(got, w) {
			t.Errorf("unexpected %s", w)
		}
	}

	assertContent(t, filepath.Join(target, ".plaesy", "instructions", "core.md"), "core instructions\n")
	// The platform core file is a copy of plaesy.mapping.core, not of the
	// always-load list: the mapping points at agents.instructions.md.
	assertContent(t, filepath.Join(target, "CLAUDE.md"), "agents core file\n")
	assertContent(t, filepath.Join(target, ".plaesy", "roles", "architect.md"), "agent architect\n")
	assertContent(t, filepath.Join(target, ".claude", "commands", "nested", "deep.md"), "prompt deep\n")
}

func TestInitReplacesTheLoopStatePlaceholder(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)
	if err := Init(Options{TargetDir: target, PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(target, ".plaesy", "state.json"))
	if err != nil {
		t.Fatalf("read state.json: %v", err)
	}
	if strings.Contains(string(data), "[TIMESTAMP]") {
		t.Errorf("state.json still holds the placeholder: %s", data)
	}
	// The value must be a real UTC timestamp, not any replacement.
	if !strings.Contains(string(data), `"initialized_at": "20`) {
		t.Errorf("state.json timestamp looks wrong: %s", data)
	}
}

// TestInitRendersTheRealStateTemplate runs Init against the template the
// repository actually ships, not the fakeHome fixture.
//
// TestInitReplacesTheLoopStatePlaceholder above could not catch the bug it was
// written to prevent: fakeHome writes `{"initialized_at": "[TIMESTAMP]"}`,
// while templates/state.template.json carries `{{TIMESTAMP}}`. The substitution
// matched the fixture and silently did nothing to the real file, so `plaesy
// init` handed every new project a literal {{TIMESTAMP}} in the very file /loop
// reads for its state. A test whose fixture disagrees with production is worse
// than no test, because it reports green.
func TestInitRendersTheRealStateTemplate(t *testing.T) {
	real := filepath.Join("..", "..", "..", "templates", "state.template.json")
	if _, err := os.Stat(real); err != nil {
		t.Fatalf("the real state template must be reachable from this package "+
			"(expected %s); a test that skips is a test that protects nothing: %v", real, err)
	}

	home := fakeHome(t)
	tmpl, err := os.ReadFile(real)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "templates", "state.template.json"), tmpl, 0o644); err != nil {
		t.Fatal(err)
	}
	target := newTarget(t)
	if err := Init(Options{TargetDir: target, PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(target, ".plaesy", "state.json"))
	if err != nil {
		t.Fatalf("read state.json: %v", err)
	}
	for _, leftover := range []string{"{{TIMESTAMP}}", "[TIMESTAMP]"} {
		if strings.Contains(string(data), leftover) {
			t.Errorf("state.json still holds %s after init, so the substitution is a no-op "+
				"against the real template: %s", leftover, data)
		}
	}
	// And it must still be the shape /loop parses. The timestamps live under a
	// "plaesy" object, not at the top level.
	var parsed struct {
		Plaesy struct {
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		} `json:"plaesy"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("state.json is not valid JSON after rendering: %v", err)
	}
	for key, s := range map[string]string{
		"created_at": parsed.Plaesy.CreatedAt,
		"updated_at": parsed.Plaesy.UpdatedAt,
	} {
		if s == "" {
			t.Errorf("state.json plaesy.%s is empty, so the timestamp never rendered", key)
			continue
		}
		if _, err := time.Parse("2006-01-02T15:04:05Z", s); err != nil {
			t.Errorf("state.json plaesy.%s is not a UTC timestamp: %q", key, s)
		}
	}
}

func TestInitWithoutAPlatformSkipsPlatformFiles(t *testing.T) {
	for _, ai := range []string{"", "none", "manual", "generic_ai"} {
		t.Run("ai="+ai, func(t *testing.T) {
			home := fakeHome(t)
			target := newTarget(t)
			if err := Init(Options{TargetDir: target, AIPlatform: ai, PlaesyHome: home}); err != nil {
				t.Fatalf("Init: %v", err)
			}
			got := listing(t, target)
			if hasEntry(got, "CLAUDE.md") || hasEntry(got, ".claude/") {
				t.Errorf("platform files created for ai=%q:\n  %s", ai, strings.Join(got, "\n  "))
			}
			// Roles are copied by SetupPlatformConfig, which "none" skips: with no
			// platform there is no agent runtime to bind them to, and the project
			// keeps .plaesy/roles/ out of the tree until a platform is chosen.
			if hasEntry(got, ".plaesy/roles/") {
				t.Errorf("roles created for ai=%q:\n  %s", ai, strings.Join(got, "\n  "))
			}
			// The platform-agnostic structure still happens.
			if !hasEntry(got, ".plaesy/instructions/core.md") {
				t.Errorf("instructions not copied for ai=%q", ai)
			}
		})
	}
}

func TestInitRemapsThePromptExtensionForCopilot(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)
	if err := Init(Options{TargetDir: target, AIPlatform: "github_copilot", PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	got := listing(t, target)
	if !hasEntry(got, ".github/prompts/start.prompt.md") {
		t.Errorf("copilot prompt extension not applied:\n  %s", strings.Join(got, "\n  "))
	}
	if hasEntry(got, ".github/prompts/start.md") {
		t.Errorf("copilot kept the plain .md prompt")
	}
	if !hasEntry(got, ".github/copilot-instructions.md") {
		t.Errorf("copilot core file missing:\n  %s", strings.Join(got, "\n  "))
	}
}

func TestInitPreservesExistingFiles(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)

	// Pre-create the files the copy steps would otherwise write, each with
	// content that must survive.
	preserved := map[string]string{
		filepath.Join(".plaesy", "context.md"):                    "MY CONTEXT\n",
		filepath.Join(".plaesy", "state.json"):                    `{"mine": true}`,
		filepath.Join(".plaesy", "instructions", "core.md"):       "MY INSTRUCTIONS\n",
		filepath.Join(".plaesy", "roles", "architect.md"):         "MY AGENT\n",
		filepath.Join(".plaesy", "templates", "spec.template.md"): "MY TEMPLATE\n",
		filepath.Join(".plaesy", "checklists", "requirements.md"): "MY CHECKLIST\n",
		"CLAUDE.md": "MY CORE FILE\n",
		filepath.Join(".claude", "commands", "start.md"): "MY PROMPT\n",
	}
	for rel, content := range preserved {
		path := filepath.Join(target, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := Init(Options{TargetDir: target, AIPlatform: "claude", PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	for rel, want := range preserved {
		assertContent(t, filepath.Join(target, filepath.FromSlash(rel)), want)
	}
	// Everything not pre-created is still created.
	if !exists(t, filepath.Join(target, ".plaesy", "instructions", "scoped.md")) {
		t.Error("a file that did not pre-exist was not copied")
	}
}

func TestInitIsIdempotent(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)
	opts := Options{TargetDir: target, AIPlatform: "claude", PlaesyHome: home}
	if err := Init(opts); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	first := listing(t, target)
	if err := Init(opts); err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if second := listing(t, target); strings.Join(second, "\n") != strings.Join(first, "\n") {
		t.Errorf("a second Init changed the tree:\nfirst:\n  %s\nsecond:\n  %s",
			strings.Join(first, "\n  "), strings.Join(second, "\n  "))
	}
}

func TestInitWritesNothingOutsideTheTarget(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)
	// A sibling project in the same temp root, which must stay untouched.
	neighbour := longTempDir(t, "neighbour")
	if err := os.WriteFile(filepath.Join(neighbour, "keep.txt"), []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Init(Options{TargetDir: target, AIPlatform: "claude", PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if !hasEntry(listing(t, neighbour), "keep.txt") {
		t.Error("Init wrote into a neighbouring directory")
	}
	// The write-probe must not survive, and the home must be unchanged.
	if exists(t, filepath.Join(target, ".plaesy-init-write-check")) {
		t.Error("the write probe was left behind in the target")
	}
	if !hasEntry(listing(t, home), "templates/spec.template.md") {
		t.Error("the home tree lost a file")
	}
	if _, err := os.Stat(filepath.Join(home, ".plaesy")); err == nil {
		t.Error("Init created a .plaesy inside the home tree")
	}
}

func TestInitTargetValidation(t *testing.T) {
	home := fakeHome(t)

	t.Run("missing directory", func(t *testing.T) {
		target := filepath.Join(newTarget(t), "nope")
		err := Init(Options{TargetDir: target, PlaesyHome: home})
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("err = %v, want a missing-directory error", err)
		}
	})

	t.Run("target is a file", func(t *testing.T) {
		target := newTarget(t)
		file := filepath.Join(target, "a.txt")
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		err := Init(Options{TargetDir: file, PlaesyHome: home})
		if err == nil || !strings.Contains(err.Error(), "not a directory") {
			t.Errorf("err = %v, want a not-a-directory error", err)
		}
	})

	t.Run("empty target means the working directory", func(t *testing.T) {
		target := newTarget(t)
		prev, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(target); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chdir(prev) })

		if err := Init(Options{PlaesyHome: home}); err != nil {
			t.Fatalf("Init: %v", err)
		}
		if !exists(t, filepath.Join(target, ".plaesy", "context.md")) {
			t.Error("Init did not use the working directory")
		}
	})
}

func TestInitRejectsAnUnknownPlatform(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)
	err := Init(Options{TargetDir: target, AIPlatform: "not-a-platform", PlaesyHome: home})
	if err == nil || !strings.Contains(err.Error(), "invalid AI platform") {
		t.Fatalf("err = %v, want an invalid-platform error", err)
	}
	if exists(t, filepath.Join(target, ".plaesy")) {
		t.Error("a rejected platform still created a structure")
	}
	// The message must list what is available, or the user cannot recover.
	for _, want := range []string{"claude", "github_copilot"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestInitFailsWhenTheHomeHasNoScriptsDir(t *testing.T) {
	home := fakeHome(t)
	if err := os.RemoveAll(filepath.Join(home, "scripts")); err != nil {
		t.Fatal(err)
	}
	target := newTarget(t)

	// A templates/ directory with no scripts/configs/platform.json is no
	// longer accepted as a Plaesy home at all: isPlaesyHome requires both, so
	// a stale/partial install is rejected up front instead of silently
	// falling back to a config with zero platforms (see home.go).
	err := Init(Options{TargetDir: target, PlaesyHome: home})
	if err == nil || !strings.Contains(err.Error(), "does not look like a Plaesy repo root") {
		t.Fatalf("Init = %v, want a rejected PlaesyHome", err)
	}
}

func TestInitFailsWhenTheConfigIsCorrupt(t *testing.T) {
	home := fakeHome(t)
	path := filepath.Join(home, "scripts", "configs", "platform.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := newTarget(t)

	// A corrupt platform.json must fail loudly, not fall back to a config
	// with zero platforms (which used to surface as a confusing
	// "invalid AI platform: ... (available: [])" later on).
	err := Init(Options{TargetDir: target, PlaesyHome: home})
	if err == nil || !strings.Contains(err.Error(), "has no usable platform.json") {
		t.Fatalf("Init = %v, want a platform.json error", err)
	}
}

func TestInitWithoutInstructionSources(t *testing.T) {
	// A home with no instructions/ at all: the copy step must warn and carry
	// on rather than fail the whole init.
	home := longTempDir(t, "bare-home")
	if err := os.MkdirAll(filepath.Join(home, "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A config, so "instructions" is a real core directory: the directory must
	// be created even though it stays empty.
	if err := os.MkdirAll(filepath.Join(home, "scripts", "configs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "scripts", "configs", "platform.json"),
		[]byte(`{"plaesy": {"core_directories": ["memory", "instructions"]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	target := newTarget(t)

	if err := Init(Options{TargetDir: target, PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	got := listing(t, target)
	if !hasEntry(got, ".plaesy/instructions/") {
		t.Errorf("the core directory itself should still exist:\n  %s", strings.Join(got, "\n  "))
	}
	// Without mapping.json there is no list, so nothing is invented.
	if hasEntry(got, ".plaesy/instructions/core.md") {
		t.Error("an instruction was copied without a list to copy it from")
	}
	// A missing context.template.md still leaves a usable context.md.
	if !exists(t, filepath.Join(target, ".plaesy", "context.md")) {
		t.Error("context.md was not created")
	}
}

func TestInitEmptyMappingListsAreAFallbackNotAFailure(t *testing.T) {
	home := fakeHome(t)
	path := filepath.Join(home, "instructions", "mapping.json")
	if err := os.WriteFile(path, []byte(`{"mappings": {"always_load": []}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	target := newTarget(t)
	if err := Init(Options{TargetDir: target, PlaesyHome: home}); err != nil {
		t.Fatalf("an empty always_load should skip the copy, not fail init: %v", err)
	}
	if hasEntry(listing(t, target), ".plaesy/instructions/core.md") {
		t.Error("instructions copied from an empty list")
	}
}

func TestNormalizePlatform(t *testing.T) {
	cfg, err := LoadPlatformConfig(fakeHome(t))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ in, want string }{
		{"", "none"},
		{"none", "none"},
		{"manual", "none"},
		{"generic", "none"},
		{"generic_ai", "none"},
		{"generic-ai", "none"},
		{"claude", "claude"},
		{"claude_code", "claude"},
		{"claude-code", "claude"},
		{"CLAUDE", "claude"},
		{"anthropic", "claude"},
		{"cursor", "cursor_ai"},
		{"cursor-ai", "cursor_ai"},
		{"github", "github_copilot"},
		{"gh-copilot", "github_copilot"},
		{"GitHub-Copilot", "github_copilot"},
		// A configured id and a configured display name both resolve.
		{"claude", "claude"},
		{"GitHub Copilot", "github_copilot"},
		// An unknown value is returned unchanged so Init can reject it.
		{"mystery", "mystery"},
	}
	for _, tc := range cases {
		if got := normalizePlatform(tc.in, cfg); got != tc.want {
			t.Errorf("normalizePlatform(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	// A nil config must not panic: the fallback path is reachable.
	if got := normalizePlatform("claude", nil); got != "claude" {
		t.Errorf("normalizePlatform with a nil config = %q", got)
	}
	if got := normalizePlatform("mystery", nil); got != "mystery" {
		t.Errorf("normalizePlatform(%q, nil) = %q", "mystery", got)
	}
}

func TestPromptExtension(t *testing.T) {
	if got := promptExtension("github_copilot"); got != ".prompt.md" {
		t.Errorf("promptExtension(github_copilot) = %q", got)
	}
	for _, p := range []string{"claude", "cursor_ai", "anything", ""} {
		if got := promptExtension(p); got != ".md" {
			t.Errorf("promptExtension(%q) = %q, want .md", p, got)
		}
	}
}

func TestPlatformConfigAccessors(t *testing.T) {
	cfg, err := LoadPlatformConfig(fakeHome(t))
	if err != nil {
		t.Fatal(err)
	}
	names := cfg.PlatformNames()
	if strings.Join(names, ",") != "claude,github_copilot" {
		t.Errorf("PlatformNames = %v, want sorted", names)
	}
	if got := cfg.DisplayName("claude"); got != "Claude Code" {
		t.Errorf("DisplayName = %q", got)
	}
	// An unknown platform falls back to its id rather than to an empty string,
	// which would print "Selected: " with nothing after it.
	if got := cfg.DisplayName("nope"); got != "nope" {
		t.Errorf("DisplayName(unknown) = %q", got)
	}
	if !cfg.HasPlatform("claude") || cfg.HasPlatform("nope") {
		t.Error("HasPlatform disagrees with the config")
	}
}

func TestSetupPlatformConfigWithAnUnknownPlatform(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)
	cfg, err := LoadPlatformConfig(home)
	if err != nil {
		t.Fatal(err)
	}
	// Roles are copied before the platform lookup, so an unknown platform still
	// produces roles and no platform files rather than an error.
	if err := SetupPlatformConfig(home, target, "nope", cfg); err != nil {
		t.Fatalf("SetupPlatformConfig: %v", err)
	}
	got := listing(t, target)
	if !hasEntry(got, ".plaesy/roles/architect.md") {
		t.Errorf("roles missing:\n  %s", strings.Join(got, "\n  "))
	}
	if hasEntry(got, "CLAUDE.md") {
		t.Error("platform files created for an unknown platform")
	}
}

func TestSetupPlatformConfigWithoutAMappingSource(t *testing.T) {
	home := fakeHome(t)
	target := newTarget(t)
	// A config whose plaesy.mapping has no core/prompts entries: there is
	// nothing to copy from, and that is not an error.
	cfg := &PlatformConfig{
		Plaesy: PlaesySection{BaseDirectory: ".plaesy"},
		Platforms: map[string]PlatformEntry{
			"x": {Name: "X", Mapping: map[string]string{"core": "X.md", "prompts": ".x/prompts"}},
		},
	}
	if err := SetupPlatformConfig(home, target, "x", cfg); err != nil {
		t.Fatalf("SetupPlatformConfig: %v", err)
	}
	if exists(t, filepath.Join(target, "X.md")) {
		t.Error("a core file was created with no mapping source")
	}
	if exists(t, filepath.Join(target, ".x", "prompts")) {
		t.Error("a prompt directory was created with no mapping source")
	}
}

func TestLoadMappingListErrors(t *testing.T) {
	home := fakeHome(t)

	if _, err := loadMappingList(home, "nonsense"); err == nil ||
		!strings.Contains(err.Error(), "unknown mapping list key") {
		t.Errorf("an unknown key should be an error, got %v", err)
	}
	if _, err := loadMappingList(filepath.Join(home, "nope"), "always_load"); err == nil ||
		!strings.Contains(err.Error(), "read instructions/mapping.json") {
		t.Errorf("a missing mapping.json should be an error, got %v", err)
	}
	// scope_load may legitimately be empty; that is not a misconfiguration.
	names, err := loadScopeLoad(home)
	if err != nil {
		t.Fatalf("loadScopeLoad: %v", err)
	}
	if len(names) != 1 || names[0] != "scoped.instructions.md" {
		t.Errorf("loadScopeLoad = %v", names)
	}

	bad := filepath.Join(home, "instructions", "mapping.json")
	if err := os.WriteFile(bad, []byte("{oops"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadAlwaysLoad(home); err == nil ||
		!strings.Contains(err.Error(), "parse instructions/mapping.json") {
		t.Errorf("a corrupt mapping.json should be an error, got %v", err)
	}
}

func TestWalkUpForHome(t *testing.T) {
	home := fakeHome(t)
	deep := filepath.Join(home, "scripts", "configs")
	got, ok := walkUpForHome(deep)
	if !ok || got != home {
		t.Errorf("walkUpForHome(%q) = %q, %v; want %q", deep, got, ok, home)
	}
	if _, ok := walkUpForHome(string(filepath.Separator) + "definitely-not-a-home"); ok {
		t.Error("walkUpForHome found a home where there is none")
	}
}

func TestFindHomeRejectsABadOverride(t *testing.T) {
	plain := longTempDir(t, "not-a-home")
	if _, err := FindHome(plain); err == nil ||
		!strings.Contains(err.Error(), "does not look like a Plaesy repo root") {
		t.Errorf("err = %v, want a rejected override", err)
	}

	// The env var is honored when it points at a real home, and ignored when it
	// does not (the search then continues).
	home := fakeHome(t)
	t.Setenv("PLAESY_HOME", home)
	got, err := FindHome("")
	if err != nil {
		t.Fatalf("FindHome with PLAESY_HOME: %v", err)
	}
	if got != home {
		t.Errorf("FindHome = %q, want %q", got, home)
	}

	t.Setenv("PLAESY_HOME", plain)
	if got, err := FindHome(""); err != nil {
		t.Errorf("a bad PLAESY_HOME should fall through to the search, got %v", err)
	} else if got == plain {
		t.Error("FindHome returned a directory that is not a Plaesy home")
	}
}

func TestFindHomeOverrideWins(t *testing.T) {
	fakeHome(t) // ensure the real repo, if reachable, is not the answer below
	home := fakeHome(t)
	t.Setenv("PLAESY_HOME", longTempDir(t, "other"))
	got, err := FindHome(home)
	if err != nil {
		t.Fatalf("FindHome: %v", err)
	}
	if got != home {
		t.Errorf("FindHome = %q, want the explicit override %q", got, home)
	}
}

func TestTemplateCopyIsNotRecursive(t *testing.T) {
	home := fakeHome(t)
	// The bash original globbed templates/*.md, so a nested template was never
	// installed. Pin that: silently installing nested files would change which
	// instructions a project ends up with.
	nested := filepath.Join(home, "templates", "advanced", "deep.template.md")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, []byte("deep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	target := newTarget(t)
	if err := Init(Options{TargetDir: target, PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if exists(t, filepath.Join(target, ".plaesy", "templates", "advanced")) {
		t.Error("a nested template was copied; the copy must stay flat")
	}
}

func TestInitToleratesMissingCopySources(t *testing.T) {
	home := fakeHome(t)
	// Every optional source is gone, one category at a time.
	if err := os.RemoveAll(filepath.Join(home, "prompts")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(home, "agents")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(home, "checklists")); err != nil {
		t.Fatal(err)
	}
	// scripts/configs still holds platform.json, so drop only the script
	// subdirectories that copyScripts iterates.
	for _, sub := range []string{"bash", "powershell"} {
		if err := os.RemoveAll(filepath.Join(home, "scripts", sub)); err != nil {
			t.Fatal(err)
		}
	}
	// mapping.json lists a file that is not in instructions/.
	if err := os.WriteFile(filepath.Join(home, "instructions", "mapping.json"),
		[]byte(`{"mappings": {"always_load": ["core.instructions.md", "ghost.instructions.md"],
			"scope_load": ["scoped.instructions.md"]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// platform.json points core at a file that does not exist.
	cfg := strings.Replace(platformJSON, "instructions/agents.instructions.md", "instructions/missing.instructions.md", 1)
	if err := os.WriteFile(filepath.Join(home, "scripts", "configs", "platform.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	target := newTarget(t)
	if err := Init(Options{TargetDir: target, AIPlatform: "claude", PlaesyHome: home}); err != nil {
		t.Fatalf("missing sources must warn, not fail init: %v", err)
	}
	got := listing(t, target)
	// The listed-but-missing instruction is skipped; the present ones copy.
	if !hasEntry(got, ".plaesy/instructions/core.md") || !hasEntry(got, ".plaesy/instructions/scoped.md") {
		t.Errorf("present instructions not copied:\n  %s", strings.Join(got, "\n  "))
	}
	if hasEntry(got, ".plaesy/instructions/ghost.md") {
		t.Error("an instruction was invented for a missing source")
	}
	// copyScripts pre-creates its three destinations, and the prompt copy
	// creates its directory before walking, so those exist even when empty.
	for _, dir := range []string{".plaesy/scripts/bash/", ".plaesy/scripts/powershell/", ".claude/commands/"} {
		if !hasEntry(got, dir) {
			t.Errorf("missing destination directory %s:\n  %s", dir, strings.Join(got, "\n  "))
		}
	}
	// copyFlatDir skips the whole step when the source directory is gone, so
	// checklists/ is never created rather than being created empty.
	if hasEntry(got, ".plaesy/checklists/") {
		t.Error("checklists/ was created for a source directory that does not exist")
	}
	if hasEntry(got, "CLAUDE.md") {
		t.Error("a platform core file was created with no source")
	}
}

func TestInitCopiesScriptsThatDoExist(t *testing.T) {
	home := fakeHome(t)
	// Only one PowerShell script, and a non-.ps1 file that must be filtered out.
	if err := os.RemoveAll(filepath.Join(home, "scripts", "bash")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "scripts", "powershell", "README.md"), []byte("notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(home, "scripts", "powershell", "tool.ps1")); err != nil {
		t.Fatal(err)
	}

	target := newTarget(t)
	if err := Init(Options{TargetDir: target, PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	got := listing(t, target)
	if hasEntry(got, ".plaesy/scripts/bash/tool.sh") {
		t.Error("a script was copied from a directory that was removed")
	}
	if hasEntry(got, ".plaesy/scripts/powershell/tool.ps1") {
		t.Error("a script was copied after it was removed from the source")
	}
	if hasEntry(got, ".plaesy/scripts/powershell/README.md") {
		t.Error("a .md file leaked into the PowerShell script directory")
	}
	// configs is copied with no extension filter, so platform.json must be there.
	if !hasEntry(got, ".plaesy/scripts/configs/platform.json") {
		t.Errorf("configs not copied:\n  %s", strings.Join(got, "\n  "))
	}
}

func TestCreateStructureSkipsEmptyDirectoryEntries(t *testing.T) {
	target := newTarget(t)
	// A config with empty strings in the directory lists: they must be skipped,
	// not turned into a directory named "".
	cfg := &PlatformConfig{Plaesy: PlaesySection{
		BaseDirectory:      defaultBaseDirectory,
		CoreDirectories:    []string{"memory", ""},
		ProjectDirectories: []string{"", "docs"},
	}}
	if err := CreateStructure(fakeHome(t), target, cfg); err != nil {
		t.Fatalf("CreateStructure: %v", err)
	}
	got := listing(t, target)
	if !hasEntry(got, ".plaesy/memory/") || !hasEntry(got, "docs/") {
		t.Errorf("configured directories missing:\n  %s", strings.Join(got, "\n  "))
	}
	if hasEntry(got, "specs/") {
		t.Error("a directory appeared that the config did not ask for")
	}
}

func TestDirHasFiles(t *testing.T) {
	base := t.TempDir()

	if dirHasFiles(filepath.Join(base, "nope")) {
		t.Error("a missing directory reported as having files")
	}
	empty := filepath.Join(base, "empty")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	if dirHasFiles(empty) {
		t.Error("an empty directory reported as having files")
	}
	// A directory entry alone is not a file.
	if err := os.MkdirAll(filepath.Join(empty, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if dirHasFiles(empty) {
		t.Error("a directory containing only subdirectories reported as having files")
	}
	if err := os.WriteFile(filepath.Join(empty, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !dirHasFiles(empty) {
		t.Error("a directory with a file reported as empty")
	}
}

func TestDirHasAnyFileRecursive(t *testing.T) {
	base := filepath.Join(t.TempDir(), "root")
	if dirHasAnyFileRecursive(filepath.Join(base, "nope")) {
		t.Error("a missing directory reported as having files")
	}
	if err := os.MkdirAll(filepath.Join(base, "a", "b", "c"), 0o755); err != nil {
		t.Fatal(err)
	}
	if dirHasAnyFileRecursive(base) {
		t.Error("a tree of empty directories reported as having files")
	}
	deep := filepath.Join(base, "a", "b", "c", "deep.md")
	if err := os.WriteFile(deep, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !dirHasAnyFileRecursive(base) {
		t.Error("a deeply nested file was not found")
	}
}

func TestHasAnyExt(t *testing.T) {
	if !hasAnyExt("a.MD", []string{".md", ".json"}) {
		t.Error("matching should be case-insensitive")
	}
	if hasAnyExt("a.txt", []string{".md", ".json"}) {
		t.Error("a non-matching extension matched")
	}
	// The filter is not "match nothing when empty": the callers guard with
	// len(exts) > 0, so an empty list only has to be false here, never a
	// silent accept-all.
	if hasAnyExt("platform.json", nil) {
		t.Error("an empty filter should match nothing; callers handle the no-filter case")
	}
	// A name with no extension is not "empty" and must not match ".md".
	if hasAnyExt("README", []string{".md"}) {
		t.Error("an extensionless file matched a filter")
	}
}

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 7: "7", 10: "10", 123: "123", -4: "-4"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Errorf("itoa(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestFilePermitsSurviveTheCopy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows only emulates the read-only bit, so mode bits cannot be asserted")
	}
	home := fakeHome(t)
	src := filepath.Join(home, "templates", "spec.template.md")
	if err := os.Chmod(src, 0o600); err != nil {
		t.Fatal(err)
	}
	target := newTarget(t)
	if err := Init(Options{TargetDir: target, PlaesyHome: home}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	info, err := os.Stat(filepath.Join(target, ".plaesy", "templates", "spec.template.md"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600", info.Mode().Perm())
	}
}
