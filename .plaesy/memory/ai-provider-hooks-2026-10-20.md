---
name: ai-provider-hooks-2026-10-20
description: Verified hook manifests for Claude Code, Codex CLI, Cursor, Copilot CLI and Gemini CLI, with payload and exit-code facts and their gaps, retrieved 2026-10-20
metadata:
  type: reference
---

# AI provider hooks (verified 2026-10-20)

How each harness loads hooks, from its own documentation. The plaesy design-lint hook
(`.plaesy/hooks/design-lint.sh` and `.ps1`) is written by `plaesy init --ai <p>` and
`plaesy reload --ai <p>` through `platform.json` keys `hooks` and `hooks_format`.

## Codex CLI (verified: learn.chatgpt.com/docs/hooks)
- Files: `~/.codex/hooks.json`, `<repo>/.codex/hooks.json`, or inline `[hooks]` in `config.toml`.
  Project hooks load only when the project `.codex/` layer is trusted.
- Schema: `hooks.<Event>[] -> { matcher, hooks: [{ type, command, commandWindows, timeout, statusMessage }] }`.
  `timeout` defaults to 600 s.
- Events include `PreToolUse`, `PostToolUse`, `Stop`, `UserPromptSubmit`.
- stdin: `tool_name`, `tool_input` (for `apply_patch` the patch text is in `tool_input.command`), `tool_response` on PostToolUse.
  No standalone file-path field for edits, so the script parses `*** Update File:` lines.
- Exit 2 blocks. Non-managed hooks need trust: run `/hooks` in Codex after install.
- Gap: `PostToolUse` cannot undo a command; `decision: block` replaces the tool result.

## Cursor (verified: cursor.com/docs/agent/hooks)
- Files: `<project>/.cursor/hooks.json` (project hooks run from the project root), `~/.cursor/hooks.json`.
- Schema: `{ "version": 1, "hooks": { "<event>": [{ command, matcher, timeout }] } }`.
- Event used: `afterFileEdit`, matcher `Write`. stdin has `file_path` and `edits`.
  The event has no documented output, so it is advisory by construction.

## GitHub Copilot CLI (verified: docs.github.com … use-hooks)
- Files: `.github/hooks/NAME.json` in the repo (committed to the default branch), or `~/.copilot/hooks/`.
- Schema: `{ "version": 1, "hooks": { "postToolUse": [{ type: "command", bash, powershell, cwd, env, timeoutSec }] } }`.
  Both `bash` and `powershell` keys are needed. The default timeout is 30 s.
- Gap: the article does not document the file-edit stdin payload, a `matcher` key, or exit-code
  behaviour. The script therefore falls back to `git status` for changed frontend files.
- Gap: the impeccable harness table says Copilot has no documented hook surface. The docs above
  contradict that; the docs win, but recheck after any Copilot release.

## Gemini CLI (verified: geminicli.com/docs/hooks)
- Files: `.gemini/settings.json` (project), `~/.gemini/settings.json`, `/etc/gemini-cli/settings.json`.
- Schema: `hooks.<Event>[] -> { matcher, hooks: [{ type: "command", name, command, timeout (ms) }] }`.
  Event used: `AfterTool`, matcher `write_file|replace`.
- Exit 0 uses stdout JSON; exit 2 blocks. The stdin payload for writes is on the reference page,
  not the hooks overview; the script does not depend on it.
- **Warning from the same page:** Gemini CLI is said to be replaced by Antigravity CLI on
  2026-06-18. Only one source says so. Verify before relying on the `gemini` platform entry.

## Claude Code (already in use)
- `.claude/settings.json`, `hooks.PostToolUse`, matcher `Edit|Write`, stdin `tool_input.file_path`.
  `.claude*` is gitignored, so the hook stays local unless the ignore rule changes.

## Not verified
- Windsurf, Cline, Continue, Tabnine, Codeium, DeepSeek, Trae, Qoder and the others in
  `platform.json`: no hook mapping, and their instruction paths are unverified (see the
  provider research note in this session). They get no hook until a source is checked.

## How plaesy uses this
- `platform.json` `mapping.hooks` = manifest path, `mapping.hooks_format` = one of
  `claude`, `codex`, `cursor`, `copilot`, `gemini`.
- `internal/scaffold/hooks.go` writes both scripts under `<base>/hooks/`, then merges the
  manifest. Shared JSON files keep every entry that is not plaesy's (identified by the
  `design-lint.` marker); an invalid JSON file is left untouched and reported.
- Hooks follow `--ai`: `reload` without `--ai` refreshes no platform files, including hooks.

Related: [[ai-provider-research-2026-10-20]]

## Update: every provider in platform.json checked (2026-10-20)

Hook manifests now written for: Claude Code, Codex CLI, Cursor, Copilot CLI, Gemini CLI,
Windsurf (Devin Desktop), Qoder CLI.

- **Windsurf** (docs.devin.ai/desktop/cascade/hooks, the Windsurf Cascade docs now hosted
  under Devin): workspace file `.devin/hooks.json`; the legacy `.windsurf/hooks.json` is read
  only when `.devin/hooks.json` is absent. Events `pre_write_code` / `post_write_code`; keys
  `command`, `powershell`, `show_output`; file path in `tool_info.file_path`; exit 2 blocks
  pre-hooks only. Written as `post_write_code`.
- **Qoder CLI** (docs.qoder.com/cli/hooks): `.qoder/settings.json` (shared) and
  `.qoder/settings.local.json` (local, gitignored); Claude-shaped schema; file path in
  `tool_input.file_path`; `${QODER_PROJECT_DIR}` substituted by the CLI.
- **Tabnine CLI** (docs.tabnine.com … hooks/configuration): `.tabnine/agent/settings.json`,
  events `BeforeTool` / `AfterTool`. NOT written: the CLI is in maintenance mode (critical updates
  only through 2026-12-31) and the page gives no write-tool name or payload.
- **Cline**: the hooks page now redirects to SDK plugins (`beforeTool`, `afterTool`, TypeScript
  objects). The file-script hooks (`.clinerules/hooks/`) found in search snippets are not on any
  official page I could open. NOT written.
- **OpenCode**: plugins in `.opencode/plugins/` (JS/TS) with `tool.execute.before/after`. The
  after-handler payload is undocumented and this is not a JSON manifest. NOT written.
- **Kilo**: no hook mechanism on its docs. Rules come from `instructions` in `kilo.jsonc`
  (project) or the global config. NOT written.
- **Continue**: no hook mechanism on its docs. Rules in `.continue/rules/*.md`. NOT written.
- **Amazon Q Developer CLI**: custom agents have a `hooks` field (per secondary sources). The
  official page would not load. NOT written.
- **Trae**: no hooks found; rule path disputed between sources. NOT written.
- **DeepSeek**: no official hook or AGENTS.md documentation found. NOT written.

## Mapping bug fixed in the same pass

Nine platforms mapped `core` onto a provider configuration file (`.continue/config.json`,
`.tabnine/config.json`, `.codeium/config.json`, and six more). `reload --ai <p>` uses the
overwrite policy, so it would have replaced the user's model and provider settings.
- continue_dev -> `.continue/rules/plaesy.md` (verified)
- tabnine -> `.tabnine/guidelines/plaesy.md` (verified: project guidelines in `.tabnine/guidelines/`)
- the other seven: `core` removed (no verified rules path)
A regression test (`TestNoCoreMappingWritesOverAConfigFile`) fails if any `core` ends in `config.json`.

## Codewhale (DeepSeek) hook implemented (2026-10-21)
- Mapped to `.codewhale/hooks.toml`, format `codewhale`. Event `tool_call_after`,
  `timeout_secs = 20`, `continue_on_error = true`.
- The file starts with the marker `# managed by plaesy`. A file without the marker is never
  overwritten; a warning is printed instead.
- Codewhale runs project hooks only after the user approves the exact file: run `/hooks review`,
  then `/hooks approve <digest>`.
- `tool_call_after` has no file path, so the script lints the changed frontend files from git.
- On Windows, Codewhale runs commands via `cmd /C`. The `sh` command needs Git Bash on PATH.
