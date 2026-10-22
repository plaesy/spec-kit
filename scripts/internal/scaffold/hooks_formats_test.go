package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHooksWindsurfUsesDevinWorkspaceFileAndWriteEvent(t *testing.T) {
	dir := t.TempDir()
	cfg := hookTestConfig()
	cfg.Platforms["windsurf_ai"] = PlatformEntry{Name: "Windsurf", Mapping: map[string]string{"hooks": ".devin/hooks.json", "hooks_format": "windsurf"}}
	if err := installPlatformHooks(dir, "windsurf_ai", cfg, copyPolicy{}); err != nil {
		t.Fatal(err)
	}
	doc := readJSONMap(t, filepath.Join(dir, ".devin", "hooks.json"))
	list, _ := doc["hooks"].(map[string]any)["post_write_code"].([]any)
	if len(list) != 1 {
		t.Fatalf("want one post_write_code hook, got %v", doc)
	}
	entry := list[0].(map[string]any)
	if entry["powershell"] == nil || !strings.Contains(entry["command"].(string), "design-lint.sh") {
		t.Fatalf("Windsurf hook needs a command and a powershell key: %v", entry)
	}
}

func TestHooksQoderUsesProjectDirAndEditMatcher(t *testing.T) {
	dir := t.TempDir()
	cfg := hookTestConfig()
	cfg.Platforms["qoder"] = PlatformEntry{Name: "Qoder", Mapping: map[string]string{"hooks": ".qoder/settings.json", "hooks_format": "qoder"}}
	if err := installPlatformHooks(dir, "qoder", cfg, copyPolicy{}); err != nil {
		t.Fatal(err)
	}
	doc := readJSONMap(t, filepath.Join(dir, ".qoder", "settings.json"))
	group := doc["hooks"].(map[string]any)["PostToolUse"].([]any)[0].(map[string]any)
	if group["matcher"] != "Edit|Write" {
		t.Fatalf("matcher = %v", group["matcher"])
	}
	cmd := group["hooks"].([]any)[0].(map[string]any)["command"].(string)
	if !strings.Contains(cmd, "${QODER_PROJECT_DIR}") {
		t.Fatalf("Qoder command must use the project dir variable: %s", cmd)
	}
}

// TestNoCoreMappingWritesOverAConfigFile guards a real hazard: `reload --ai` overwrites the
// core file, so a core mapping onto a provider's config.json would replace the user's model
// and provider settings. The repository's own platform.json must never do that.
func TestNoCoreMappingWritesOverAConfigFile(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "configs", "platform.json"))
	if err != nil {
		t.Fatalf("read platform.json: %v", err)
	}
	var doc struct {
		Platforms map[string]struct {
			Mapping map[string]string `json:"mapping"`
		} `json:"platforms"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse platform.json: %v", err)
	}
	for name, p := range doc.Platforms {
		core := p.Mapping["core"]
		if strings.HasSuffix(core, "config.json") {
			t.Errorf("platform %s maps core onto %s, which is a provider config file", name, core)
		}
	}
}
