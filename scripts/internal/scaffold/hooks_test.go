package scaffold

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hookTestConfig is a minimal platform.json: the five platforms that declare a hook manifest.
func hookTestConfig() *PlatformConfig {
	return &PlatformConfig{
		Plaesy: PlaesySection{BaseDirectory: ".plaesy"},
		Platforms: map[string]PlatformEntry{
			"claude":         {Name: "Claude Code", Mapping: map[string]string{"hooks": ".claude/settings.json", "hooks_format": "claude"}},
			"codex":          {Name: "OpenAI Codex CLI", Mapping: map[string]string{"hooks": ".codex/hooks.json", "hooks_format": "codex"}},
			"cursor_ai":      {Name: "Cursor AI", Mapping: map[string]string{"hooks": ".cursor/hooks.json", "hooks_format": "cursor"}},
			"github_copilot": {Name: "GitHub Copilot", Mapping: map[string]string{"hooks": ".github/hooks/plaesy-design-lint.json", "hooks_format": "copilot"}},
			"gemini":         {Name: "Gemini CLI", Mapping: map[string]string{"hooks": ".gemini/settings.json", "hooks_format": "gemini"}},
		},
	}
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return m
}

// countPlaesyHooks counts hook entries that carry the design-lint marker, across all events.
func countPlaesyHooks(t *testing.T, doc map[string]any) int {
	t.Helper()
	hooks, _ := doc["hooks"].(map[string]any)
	n := 0
	for _, entries := range hooks {
		list, _ := entries.([]any)
		for _, e := range list {
			if isPlaesyHookEntry(e) {
				n++
			}
		}
	}
	return n
}

func TestHooksCodexWritesPostToolUseAndScripts(t *testing.T) {
	dir := t.TempDir()
	if err := installPlatformHooks(dir, "codex", hookTestConfig(), copyPolicy{}); err != nil {
		t.Fatal(err)
	}
	doc := readJSONMap(t, filepath.Join(dir, ".codex", "hooks.json"))
	post, _ := doc["hooks"].(map[string]any)["PostToolUse"].([]any)
	if len(post) != 1 {
		t.Fatalf("want one PostToolUse group, got %d", len(post))
	}
	group := post[0].(map[string]any)
	if group["matcher"] != "apply_patch|Edit|Write" {
		t.Fatalf("matcher = %v", group["matcher"])
	}
	handler := group["hooks"].([]any)[0].(map[string]any)
	if !strings.Contains(handler["command"].(string), "design-lint.sh") || handler["commandWindows"] == nil {
		t.Fatalf("handler must name the script and carry a Windows override: %v", handler)
	}
	for _, name := range []string{"design-lint.sh", "design-lint.ps1"} {
		if _, err := os.Stat(filepath.Join(dir, ".plaesy", "hooks", name)); err != nil {
			t.Fatalf("script %s not written: %v", name, err)
		}
	}
}

func TestHooksInstallIsIdempotentAndKeepsUserEntries(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	userDoc := `{"permissions":{"allow":["Bash(go test:*)"]},"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo mine"}]}]}}`
	if err := os.WriteFile(settings, []byte(userDoc), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := hookTestConfig()
	for i := 0; i < 2; i++ {
		if err := installPlatformHooks(dir, "claude", cfg, copyPolicy{overwrite: true}); err != nil {
			t.Fatal(err)
		}
	}
	doc := readJSONMap(t, settings)
	if doc["permissions"] == nil {
		t.Fatal("user key 'permissions' was dropped")
	}
	if got := countPlaesyHooks(t, doc); got != 1 {
		t.Fatalf("want exactly one plaesy entry after two installs, got %d", got)
	}
	pre, _ := doc["hooks"].(map[string]any)["PreToolUse"].([]any)
	if len(pre) != 1 {
		t.Fatalf("user PreToolUse entry was not kept: %v", pre)
	}
}

func TestHooksLeaveInvalidJSONAlone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gemini", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	broken := "{ not json, the user is editing this file"
	if err := os.WriteFile(path, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installPlatformHooks(dir, "gemini", hookTestConfig(), copyPolicy{overwrite: true}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != broken {
		t.Fatalf("an invalid settings file must not be rewritten, got %q", raw)
	}
}

func TestHooksCursorUsesAfterFileEditAndVersion(t *testing.T) {
	dir := t.TempDir()
	if err := installPlatformHooks(dir, "cursor_ai", hookTestConfig(), copyPolicy{}); err != nil {
		t.Fatal(err)
	}
	doc := readJSONMap(t, filepath.Join(dir, ".cursor", "hooks.json"))
	if doc["version"] != float64(1) {
		t.Fatalf("version = %v, want 1", doc["version"])
	}
	if _, ok := doc["hooks"].(map[string]any)["afterFileEdit"]; !ok {
		t.Fatalf("afterFileEdit missing: %v", doc)
	}
}

func TestHooksCopilotWritesBothOperatingSystemKeys(t *testing.T) {
	dir := t.TempDir()
	if err := installPlatformHooks(dir, "github_copilot", hookTestConfig(), copyPolicy{}); err != nil {
		t.Fatal(err)
	}
	doc := readJSONMap(t, filepath.Join(dir, ".github", "hooks", "plaesy-design-lint.json"))
	entry := doc["hooks"].(map[string]any)["postToolUse"].([]any)[0].(map[string]any)
	if entry["bash"] == nil || entry["powershell"] == nil {
		t.Fatalf("Copilot needs both bash and powershell keys: %v", entry)
	}
}

func TestHooksNoMappingWritesNothing(t *testing.T) {
	dir := t.TempDir()
	cfg := hookTestConfig()
	cfg.Platforms["bare"] = PlatformEntry{Name: "Bare", Mapping: map[string]string{"core": "AI.md"}}
	if err := installPlatformHooks(dir, "bare", cfg, copyPolicy{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".plaesy")); !os.IsNotExist(err) {
		t.Fatal("a platform without a hooks mapping must not write hook files")
	}
}

// TestDesignLintScriptReportsWarnFromNamedFile runs the bash hook against a fake plaesy
// binary, so the payload parsing and the path resolution are exercised without a build.
func TestDesignLintScriptReportsWarnFromNamedFile(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	fake := "#!/bin/sh\necho \"WARN $3 DL-FAKE fake finding\"\n"
	if err := os.WriteFile(filepath.Join(dir, "plaesy"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "design-lint.sh")
	if err := os.WriteFile(script, []byte(designLintBash), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", script)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(`{"tool_name":"Edit","tool_input":{"file_path":"index.html"}}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook must exit 0, got %v: %s", err, out)
	}
	if !strings.Contains(string(out), "WARN") || !strings.Contains(string(out), "DL-FAKE") {
		t.Fatalf("expected the WARN line from the named file, got %q", out)
	}
}
