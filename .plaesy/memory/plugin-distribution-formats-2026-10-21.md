---
title: "Plugin distribution formats: verified per tool, 2026-10-21"
updatedAt: "2026-10-21T00:00:00.000Z"
---

# Plugin distribution formats (verified 2026-10-21)

Reference for packaging plaesy as a plugin. Every "verified" line below was read from an official
page on 2026-10-21 (source in brackets). Items marked **not verified** must not be used as fact.

## Claude Code

- Manifest: `.claude-plugin/plugin.json`. Component folders (`skills/`, `commands/`, `agents/`,
  `hooks/`) sit at the plugin root, not inside `.claude-plugin/`.
  [plugins-reference](https://code.claude.com/docs/en/plugins-reference)
- Commands are namespaced `<plugin>:<command>`. `skills/` is the recommended location for new
  plugins; `commands/` is the older form. [plugins-reference]
- `${CLAUDE_PLUGIN_ROOT}` points at the installed version directory. Store durable state in
  `${CLAUDE_PLUGIN_DATA}` instead, because the root path changes with each version.
  [plugins-reference]
- `CLAUDE.md` inside a plugin is not loaded as context. Put instructions in a skill.
  [plugins-reference]
- Marketplace: `.claude-plugin/marketplace.json` with required `name`, `owner`, `plugins`.
  Install: `claude plugin install <plugin>@<marketplace>`. Add marketplace: `claude plugin
  marketplace add owner/repo`. [marketplace-reference](https://code.claude.com/docs/en/plugins/marketplace-reference),
  [loading](https://code.claude.com/docs/en/plugins/loading)
- Validation: `claude plugin validate <path>`, `--strict` turns warnings into failures.
  [plugins-reference]
- Reserved names: a plugin name may not start with `claude-`, `anthropic-`, `anthropics-`, or
  `cc-plugin-`. "plaesy" is not reserved. [plugins-reference]
- Hooks: 33 events, including `SessionStart` and `PostToolUse`. Schema is event → matcher group →
  handler, with `type` and `command`. Plain stdout from `SessionStart` is added to Claude's context.
  The JSON form uses `hookSpecificOutput.additionalContext`, capped at 10,000 characters.
  [hooks](https://code.claude.com/docs/en/hooks)
- **Not verified:** the byte limit for plain-text stdout from `SessionStart`. The 10,000 limit
  belongs to the JSON `additionalContext` field.
- **Not verified:** whether exit code 2 on `PostToolUse` blocks anything. The page is inconsistent
  about it.

## OpenAI Codex CLI

- Manifest: `plugin.json` at the plugin root (portable form). `.codex-plugin/plugin.json` is the
  legacy form and is still read. OpenAI-specific fields go under `extensions.com.openai`.
  [build plugins](https://developers.openai.com/codex/plugins/build)
- Components: `skills/<name>/SKILL.md`, `mcp.json`, `.app.json`, `assets/`, and `hooks/hooks.json`.
  [build plugins]
- Hooks inside plugins: the default is `hooks/hooks.json`. The `hooks` key in the manifest
  overrides it. Hooks must be trusted before they run: installing a plugin does not trust its
  hooks. [hooks](https://learn.chatgpt.com/docs/hooks)
- **Conflict, unresolved:** the build page says plugin hooks apply only to plugins installed
  manually in the Codex desktop app and are excluded from the public directory. The hooks page says
  enabled plugins can bundle hooks. The two pages do not agree.
- Repo marketplace: `.agents/plugins/marketplace.json`. Local add: `codex plugin marketplace add
  ./dir`. [build plugins]
- Publishing: a submission portal exists. Approved public plugins appear in a directory shared by
  ChatGPT and Codex. Workspace publishing is also available. The page does not state whether
  self-serve submission is open. [build plugins]
- Hook locations: `~/.codex/hooks.json`, `<repo>/.codex/hooks.json`, or inline `[hooks]` in
  `config.toml`. [hooks]

## GitHub Copilot

- Agent Plugins 1.0: `plugin.json` at the root with `"$schema":
  "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"`. Skills in `skills/<name>/SKILL.md`.
  `mcp.json` at the root and cannot be overridden. Copilot-only components go in `com.github.copilot/`:
  `agents/<name>.agent.md` and `hooks/hooks.json`. [creating a plugin](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/plugins-creating)
- Legacy form (no `$schema`): `agents/*.agent.md`, `skills/<name>/SKILL.md`, root `hooks.json`
  or `hooks/`, `.mcp.json`. Component paths can be set in `plugin.json`. [same page]
- Local test: `copilot plugin install ./my-plugin`, then `copilot plugin list`. Reinstall after each
  edit because components are cached. Uninstall with `copilot plugin uninstall <name>`. [same page]
- Marketplace: `marketplace.json` with a `plugins` array. Enable with `enabledPlugins` in
  `~/.copilot/settings.json` or `.github/copilot/settings.json`. [about-plugins](https://docs.github.com/en/copilot/concepts/agents/about-plugins)
- **Not verified:** that `.claude-plugin/plugin.json` is accepted. The legacy section says "additional
  manifest locations" without naming them. Earlier notes that said otherwise were wrong.
- **Not verified:** any claim about Claude plugin compatibility. The page makes no such claim.

## Cursor

- Manifest: `.cursor-plugin/plugin.json`. Rules in `rules/*.mdc` with frontmatter `description`,
  `alwaysApply`, optional `globs`. Skills in `skills/<name>/SKILL.md`. Agents and commands are
  supported. MCP config in `mcp.json` at the root. [plugin reference](https://cursor.com/docs/reference/plugins)
- Hooks: `hooks/hooks.json` with a top-level `"hooks"` object. Example events: `afterFileEdit`,
  `beforeShellExecution`, `sessionEnd`. [same page]
- Distribution: public Git repo, submit at cursor.com/marketplace/publish, review by the Cursor team.
  Multi-plugin repos use `.cursor-plugin/marketplace.json`. [same page]
- **Not verified:** a local test procedure. The page does not describe one and points to a template
  repo.

## Agent Skills (open standard)

- Spec: `SKILL.md` with YAML frontmatter. `name`: 1–64 characters, lowercase letters, digits and
  hyphens, no leading, trailing or doubled hyphen, equal to the parent folder name. `description`:
  1–1024 characters. Optional: `license`, `compatibility` (≤500), `metadata`, `allowed-tools`
  (experimental). Optional folders: `scripts/`, `references/`, `assets/`. Keep `SKILL.md` under
  500 lines. Validate with `skills-ref validate ./my-skill`. [specification](https://agentskills.io/specification)
- **Not verified:** which tools support the standard. The client list at
  [clients.md](https://agentskills.io/clients.md) loads dynamically and could not be read. Earlier
  counts (26 to 40+) came from third-party articles and are not confirmed.

## Gemini CLI

- Extension manifest: `gemini-extension.json` (`name`, `version`, `settings`, `mcpServers`,
  `contextFileName`). Commands are TOML files under `commands/<group>/*.toml` with a `prompt` key and
  `{{args}}`. [writing extensions](https://geminicli.com/docs/extensions/writing-extensions/)
- Hooks: defined in `settings.json` under a `hooks` object. Events include `BeforeTool`, `AfterTool`,
  `SessionStart`, `SessionEnd`. [hooks reference](https://geminicli.com/docs/hooks/reference/)
- **Corrected:** earlier notes said extensions bundle `hooks/hooks.json`. The official hooks page
  does not say that, and the extensions page does not list hooks as a bundled component.
- **Not verified:** whether an extension can declare hooks at all.
- The official page says Gemini CLI was replaced by Antigravity CLI on 18 June 2026. Confirm before
  targeting Gemini.

## Windsurf (now Devin Desktop)

- Skills: `.devin/skills/<name>/SKILL.md` is preferred. `.windsurf/skills/` is still read. Global:
  `~/.codeium/windsurf/skills/` or `~/.config/devin/skills/`. Also scans `.agents/skills/` and,
  with Claude config reading on, `.claude/skills/`. Required frontmatter: `name`, `description`.
  [skills](https://docs.devin.ai/desktop/cascade/skills)
- Hooks: `.devin/hooks.json` in the workspace (legacy `.windsurf/hooks.json`), plus user, system and
  cloud levels. Schema: `{"hooks": {"<event>": [{"command", "powershell", "show_output",
  "working_directory"}]}}`. Twelve events, including `pre_write_code`, `post_write_code`,
  `pre_run_command`. Only `pre_*` hooks can block, with exit code 2. [hooks](https://docs.devin.ai/desktop/cascade/hooks)
- Packaging: none. Distribution is by committing `.devin/` to the repo, MDM, or the team cloud dashboard.

## Kiro

- Hooks: `.kiro/hooks/<id>.json`, with top-level `version: "v1"` and a `hooks` array. Each hook has
  `name`, `trigger`, `action` (`command` or `agent`), and optional `matcher`, `timeout`, `enabled`.
  Triggers shown: `PostFileSave`, `PreToolUse`, `PostToolUse`, `Stop`. The full trigger list is on a
  page that was not read. [hooks](https://kiro.dev/docs/hooks/index)
- Steering: `.kiro/steering/*.md`.
- Packaging: none found on the pages read.

## What this means for plaesy

- `plaesy init` writes project files and does not write `SKILL.md`. Skill generation is a new piece.
- The repo's `.gitignore` rule `.claude*` ignores `.claude-plugin/`. Fix before any Claude plugin
  work, using an exception such as `!/.claude-plugin/`.
- The existing Claude design-lint hook uses `$CLAUDE_PROJECT_DIR`. A plugin hook should use
  `${CLAUDE_PLUGIN_ROOT}`.
- Reference: [platform-hooks-per-provider](../decisions/platform-hooks-per-provider.md). Hook
  schemas differ per tool, so one hook file cannot serve all of them.

## Access log

All sources were read on 2026-10-21. Pages marked as redirected were followed to their new URL:
Windsurf docs → docs.devin.ai; Codex hooks → learn.chatgpt.com.
