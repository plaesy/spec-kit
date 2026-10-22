package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func codewhaleConfig() *PlatformConfig {
	cfg := hookTestConfig()
	cfg.Platforms["deepseek"] = PlatformEntry{Name: "DeepSeek", Mapping: map[string]string{"hooks": ".codewhale/hooks.toml", "hooks_format": "codewhale"}}
	return cfg
}

func TestHooksCodewhaleWritesToolCallAfterTable(t *testing.T) {
	dir := t.TempDir()
	if err := installPlatformHooks(dir, "deepseek", codewhaleConfig(), copyPolicy{}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".codewhale", "hooks.toml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	for _, want := range []string{
		"[[hooks]]",
		`event = "tool_call_after"`,
		`command = "sh .plaesy/hooks/design-lint.sh"`,
		"timeout_secs = 20",
		"continue_on_error = true",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("hooks.toml missing %q:\n%s", want, body)
		}
	}
}

func TestHooksCodewhaleNeverOverwritesUserFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".codewhale", "hooks.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	mine := "[[hooks]]\nevent = \"session_start\"\ncommand = \"echo mine\"\n"
	if err := os.WriteFile(path, []byte(mine), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installPlatformHooks(dir, "deepseek", codewhaleConfig(), copyPolicy{overwrite: true}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != mine {
		t.Fatalf("a hooks.toml without the plaesy marker must not be rewritten, got %q", raw)
	}
}

func TestHooksCodewhaleRefreshesOwnFile(t *testing.T) {
	dir := t.TempDir()
	cfg := codewhaleConfig()
	for i := 0; i < 2; i++ {
		if err := installPlatformHooks(dir, "deepseek", cfg, copyPolicy{overwrite: true}); err != nil {
			t.Fatal(err)
		}
	}
	raw, _ := os.ReadFile(filepath.Join(dir, ".codewhale", "hooks.toml"))
	if n := strings.Count(string(raw), "[[hooks]]"); n != 1 {
		t.Fatalf("want exactly one [[hooks]] table after two installs, got %d", n)
	}
}
