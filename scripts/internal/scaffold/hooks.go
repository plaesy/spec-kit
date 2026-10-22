package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// hookCommandMarker identifies hook entries this package owns. Every plaesy hook
// command names the design-lint script, so an entry that contains it is replaced on
// each install, and an entry without it (a user's own hook) is never touched.
const hookCommandMarker = "design-lint."

// The design-lint hook is advisory: it runs `plaesy design validate` on the frontend
// files an edit names, prints FAIL and WARN lines, and always exits 0 so it never blocks
// an edit. The same script is used by every harness; only the manifest that calls it differs.
const designLintBash = `#!/bin/sh
# plaesy design-lint: advisory design check after a file edit. Never blocks; always exits 0.
# Reads the hook JSON on stdin. Lints the files the edit names. When the payload names
# none, it lints the frontend files git reports as changed. Needs the plaesy binary at the
# project root or on PATH; without it the hook does nothing.
root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
input=$(cat)
named=$(printf '%s' "$input" | grep -oE '"file_path"[[:space:]]*:[[:space:]]*"[^"]*"' | sed -E 's/.*:[[:space:]]*"(.*)"$/\1/')
patched=$(printf '%s' "$input" | grep -oE '\*\*\* (Add|Update) File: [^\\"]+' | sed -E 's/^\*\*\* (Add|Update) File: //')
files=$(printf '%s\n%s\n' "$named" "$patched" | sed 's/\\\\/\//g' | sed '/^$/d')
if [ -z "$files" ]; then
  files=$(git -C "$root" status --porcelain 2>/dev/null | sed -E 's/^...//')
fi
bin="$root/plaesy"
[ -x "$bin" ] || bin="$root/plaesy.exe"
[ -x "$bin" ] || bin=$(command -v plaesy 2>/dev/null)
[ -n "$bin" ] || exit 0
printf '%s\n' "$files" | sort -u | while IFS= read -r f; do
  case "$f" in *.html|*.htm|*.css|*.scss|*.jsx|*.tsx|*.vue|*.svelte) ;; *) continue ;; esac
  p="$f"
  case "$p" in /*|?:*) ;; *) p="$root/$f" ;; esac
  [ -f "$p" ] || continue
  "$bin" validate design "$p" 2>&1 | grep -E '^(FAIL|WARN) '
done
exit 0
`

// designLintPowerShell is the Windows twin of designLintBash, for harnesses whose hooks
// run PowerShell (Copilot's "powershell" key, and Codex's commandWindows on Windows).
const designLintPowerShell = `# plaesy design-lint: advisory design check after a file edit. Never blocks; always exits 0.
$ErrorActionPreference = 'SilentlyContinue'
$root = git rev-parse --show-toplevel 2>$null
if (-not $root) { $root = (Get-Location).Path }
$payload = [Console]::In.ReadToEnd()
$named = [regex]::Matches($payload, '"file_path"\s*:\s*"([^"]*)"') | ForEach-Object { $_.Groups[1].Value }
$patched = [regex]::Matches($payload, '\*\*\* (?:Add|Update) File: ([^\\"]+)') | ForEach-Object { $_.Groups[1].Value }
$files = @($named) + @($patched) | ForEach-Object { $_ -replace '\\\\', '/' } | Where-Object { $_ }
if (-not $files) { $files = git -C $root status --porcelain | ForEach-Object { $_.Substring(3) } }
$bin = Join-Path $root 'plaesy.exe'
if (-not (Test-Path $bin)) { $bin = Join-Path $root 'plaesy' }
if (-not (Test-Path $bin)) {
  $cmd = Get-Command plaesy -ErrorAction SilentlyContinue
  if (-not $cmd) { exit 0 }
  $bin = $cmd.Source
}
foreach ($f in ($files | Sort-Object -Unique)) {
  if ($f -notmatch '\.(html|htm|css|scss|jsx|tsx|vue|svelte)$') { continue }
  $p = if ([System.IO.Path]::IsPathRooted($f)) { $f } else { Join-Path $root $f }
  if (-not (Test-Path $p -PathType Leaf)) { continue }
  & $bin validate design $p 2>&1 | Select-String -Pattern '^(FAIL|WARN) '
}
exit 0
`

// installPlatformHooks writes the design-lint hook for a platform that declares a
// "hooks" mapping in platform.json. It writes the two scripts under <base>/hooks/, then
// the platform's manifest. A shared settings file keeps every entry that is not ours.
func installPlatformHooks(targetDir, platform string, cfg *PlatformConfig, pol copyPolicy) error {
	entry, ok := cfg.Platforms[platform]
	if !ok {
		return nil
	}
	manifest := entry.Mapping["hooks"]
	format := entry.Mapping["hooks_format"]
	if manifest == "" || format == "" {
		return nil
	}

	scriptsRel := filepath.ToSlash(filepath.Join(cfg.Plaesy.BaseDirectory, "hooks"))
	if err := writeHookScript(filepath.Join(targetDir, filepath.FromSlash(scriptsRel), "design-lint.sh"), designLintBash, pol); err != nil {
		return err
	}
	if err := writeHookScript(filepath.Join(targetDir, filepath.FromSlash(scriptsRel), "design-lint.ps1"), designLintPowerShell, pol); err != nil {
		return err
	}

	shCmd := func(prefix string) string { return "sh \"" + prefix + "/" + scriptsRel + "/design-lint.sh\"" }
	psCmd := "powershell -NoProfile -ExecutionPolicy Bypass -File " + scriptsRel + "/design-lint.ps1"
	dst := filepath.Join(targetDir, filepath.FromSlash(manifest))

	var err error
	switch format {
	case "claude":
		err = mergeHookManifest(dst, "claude", desiredClaudeHooks(shCmd(`$CLAUDE_PROJECT_DIR`)))
	case "codex":
		err = mergeHookManifest(dst, "codex", desiredCodexHooks(shCmd(`$(git rev-parse --show-toplevel)`), psCmd))
	case "gemini":
		err = mergeHookManifest(dst, "gemini", desiredGeminiHooks(shCmd(`$GEMINI_PROJECT_DIR`)))
	case "windsurf":
		err = mergeHookManifest(dst, "windsurf", desiredWindsurfHooks("sh "+scriptsRel+"/design-lint.sh", psCmd))
	case "codewhale":
		err = writeCodewhaleHooks(dst, "sh "+scriptsRel+"/design-lint.sh")
	case "qoder":
		err = mergeHookManifest(dst, "qoder", desiredQoderHooks(shCmd(`${QODER_PROJECT_DIR}`)))
	case "cursor":
		err = mergeHookManifest(dst, "cursor", desiredCursorHooks("sh "+scriptsRel+"/design-lint.sh"))
	case "copilot":
		err = writeCopilotHooks(dst, "sh ./"+scriptsRel+"/design-lint.sh", psCmd)
	default:
		return fmt.Errorf("unknown hooks_format %q for platform %s", format, platform)
	}
	if err != nil {
		return err
	}
	common.LogSuccess("%s", "  ✓ design-lint hook for "+cfg.DisplayName(platform)+" → "+manifest)
	return nil
}

// writeHookScript writes one script file, honouring the copy policy: an existing script
// is replaced only on reload (overwrite), and never in a dry run.
func writeHookScript(path, content string, pol copyPolicy) error {
	if _, err := os.Stat(path); err == nil && !pol.overwrite {
		return nil
	}
	if pol.dryRun {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o755)
}

func desiredClaudeHooks(cmd string) map[string][]any {
	return map[string][]any{
		"PostToolUse": {map[string]any{
			"matcher": "Edit|Write",
			"hooks": []any{map[string]any{
				"type": "command", "command": cmd, "timeout": 20, "statusMessage": "Design lint",
			}},
		}},
	}
}

func desiredCodexHooks(shCommand, psCommand string) map[string][]any {
	return map[string][]any{
		"PostToolUse": {map[string]any{
			"matcher": "apply_patch|Edit|Write",
			"hooks": []any{map[string]any{
				"type": "command", "command": shCommand, "commandWindows": psCommand,
				"timeout": 20, "statusMessage": "Design lint",
			}},
		}},
	}
}

func desiredGeminiHooks(cmd string) map[string][]any {
	return map[string][]any{
		"AfterTool": {map[string]any{
			"matcher": "write_file|replace",
			"hooks": []any{map[string]any{
				"name": "plaesy-design-lint", "type": "command", "command": cmd, "timeout": 20000,
			}},
		}},
	}
}

// desiredWindsurfHooks uses Cascade's documented hook shape: event keys map to arrays of
// { command, powershell, show_output }, with no matcher. post_write_code runs after a
// write and cannot block it, which suits an advisory check.
func desiredWindsurfHooks(cmd, psCmd string) map[string][]any {
	return map[string][]any{
		"post_write_code": {map[string]any{
			"command": cmd, "powershell": psCmd, "show_output": true,
		}},
	}
}

// desiredQoderHooks follows Qoder CLI's documented shape, which matches Claude Code's
// (event -> groups with matcher -> hooks). ${QODER_PROJECT_DIR} is substituted by the CLI.
func desiredQoderHooks(cmd string) map[string][]any {
	return map[string][]any{
		"PostToolUse": {map[string]any{
			"matcher": "Edit|Write",
			"hooks": []any{map[string]any{
				"type": "command", "command": cmd, "timeout": 20, "statusMessage": "Design lint",
			}},
		}},
	}
}

// codewhaleOwnedMarker is the first line of every hooks.toml that plaesy wrote. Codewhale's
// project hooks file holds only [[hooks]] tables, so the whole file is plaesy's when this
// marker is present. A file without it belongs to the user and is never overwritten.
const codewhaleOwnedMarker = "# managed by plaesy"

// writeCodewhaleHooks writes Codewhale's project hook file (<workspace>/.codewhale/hooks.toml).
// Codewhale runs project hooks only after the user approves the exact file with /hooks approve.
// tool_call_after gives no file path, so the script falls back to the changed files from git.
func writeCodewhaleHooks(path, cmd string) error {
	if raw, err := os.ReadFile(path); err == nil {
		if !strings.HasPrefix(string(raw), codewhaleOwnedMarker) {
			common.LogWarning("%s", "  ⊘ "+path+" was not written by plaesy; hook not installed. Add the hook by hand or remove the file and run reload")
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	body := codewhaleOwnedMarker + ": design-lint hook. Regenerated by `plaesy init` and `plaesy reload`.\n" +
		"[[hooks]]\n" +
		"event = \"tool_call_after\"\n" +
		"name = \"plaesy-design-lint\"\n" +
		"command = " + strconv.Quote(cmd) + "\n" +
		"timeout_secs = 20\n" +
		"continue_on_error = true\n"
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

func desiredCursorHooks(cmd string) map[string][]any {
	return map[string][]any{
		"afterFileEdit": {map[string]any{
			"command": cmd, "matcher": "Write", "timeout": 20,
		}},
	}
}

// mergeHookManifest updates a JSON settings file in place. It drops the entries this
// package owned on an earlier install, appends the current ones, and keeps every other
// key and entry. A file that is not valid JSON is left alone and reported, so a user's
// broken settings file is never overwritten.
func mergeHookManifest(path, format string, desired map[string][]any) error {
	root := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(raw, &root); err != nil {
			common.LogWarning("%s", "  ⊘ "+path+" is not valid JSON; hook not installed, fix the file and run reload")
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if format == "cursor" {
		if _, ok := root["version"]; !ok {
			root["version"] = 1
		}
	}
	hooks, _ := root["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	for event, entries := range desired {
		var kept []any
		if existing, ok := hooks[event].([]any); ok {
			for _, e := range existing {
				if !isPlaesyHookEntry(e) {
					kept = append(kept, e)
				}
			}
		}
		hooks[event] = append(kept, entries...)
	}
	root["hooks"] = hooks
	return writeJSONFile(path, root)
}

// writeCopilotHooks writes Copilot's hook file. The file is ours alone, so it is
// rewritten whole. Copilot selects the bash or powershell key by operating system.
func writeCopilotHooks(path, bashCmd, psCmd string) error {
	doc := map[string]any{
		"version": 1,
		"hooks": map[string]any{
			"postToolUse": []any{map[string]any{
				"type": "command", "bash": bashCmd, "powershell": psCmd,
				"cwd": ".", "timeoutSec": 20,
			}},
		},
	}
	return writeJSONFile(path, doc)
}

func isPlaesyHookEntry(entry any) bool {
	b, err := json.Marshal(entry)
	return err == nil && strings.Contains(string(b), hookCommandMarker)
}

func writeJSONFile(path string, doc any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}
