---
title: "Plugin distribution strategy"
updatedAt: "2026-10-21T00:00:00.000Z"
---

# Plugin distribution strategy

**Status:** proposed · 2026-10-21. Not accepted by the owner. The defaults below are `ASSUMED`
and are recorded so work can continue. Each one can be changed.

## Context

The owner asked how plaesy can be distributed beyond the local `plaesy init --ai <platform>` flow.
`init` writes files into a project. It does not produce an installable plugin, and it does not
write `SKILL.md`. Each AI tool has its own plugin format, and some do not have one at all. Details
and sources are in [plugin-distribution-formats-2026-10-21](../memory/plugin-distribution-formats-2026-10-21.md).

## Options considered

1. **Keep `init` as the only channel.** Works for project use. Does not give a distributable package.
2. **Add a Claude Code plugin with a marketplace.** The mechanics are verified end to end: manifest,
   marketplace, install and validation commands.
3. **Add a portable skill layer (`SKILL.md`).** The format is verified. Tool support is not.
4. **Add plugin packaging for Copilot and Cursor.** Formats are verified. Cursor has no documented
   local test.
5. **Add Codex.** The format is verified. The hook behaviour conflicts between two official pages.
6. **Add Gemini, Windsurf or Kiro as plugins.** No plugin packaging was found for Windsurf or Kiro.
   Gemini's hook support in extensions is unverified, and its CLI was reportedly replaced.

## Decision (ASSUMED defaults)

- **ASSUMED:** keep `plaesy init` as the primary channel. Plugin channels are added on top of it.
- **ASSUMED:** Claude Code is the first plugin target, through a marketplace in this repo.
- **ASSUMED:** generate skills from `prompts/` and the on-demand `instructions/` before anything else,
  because `SKILL.md` is the most portable format.
- **ASSUMED:** Copilot and Cursor come after Claude, once their local test steps are confirmed.
- **ASSUMED:** Codex waits until the hook conflict is resolved against the official pages.
- **ASSUMED:** Gemini, Windsurf and Kiro stay on the file-based `init` path. No plugin packaging.
- **ASSUMED:** plugins ship content and skills only. The `plaesy` binary remains an optional
  dependency, and hooks degrade silently without it, as `design-lint.sh` already does.
- **Prerequisite:** fix the `.gitignore` rule `.claude*` so `.claude-plugin/` can be committed.

## Consequences

- A plugin's `commands/` duplicates the commands `init` already writes. Prefer `skills/` in the
  plugin to avoid two copies.
- Claude plugin hooks must use `${CLAUDE_PLUGIN_ROOT}`, not `$CLAUDE_PROJECT_DIR`.
- Versions in `plugin.json` must come from the repo's `VERSION` file, guarded by a drift test.

## Open questions for the owner

1. Is distribution internal to a team, or public?
2. Should a plugin include the `plaesy` binary, or content and skills only? Default: content only.
3. Should the plugin replace `plaesy init --ai claude`, or sit beside it? Default: beside it.

## References

- Claude plugins reference: https://code.claude.com/docs/en/plugins-reference (accessed 2026-10-21)
- Claude marketplace reference: https://code.claude.com/docs/en/plugins/marketplace-reference (accessed 2026-10-21)
- Claude hooks: https://code.claude.com/docs/en/hooks (accessed 2026-10-21)
- Codex build plugins: https://developers.openai.com/codex/plugins/build (accessed 2026-10-21)
- Codex hooks: https://learn.chatgpt.com/docs/hooks (accessed 2026-10-21)
- Cursor plugin reference: https://cursor.com/docs/reference/plugins (accessed 2026-10-21)
- GitHub Copilot plugins: https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/plugins-creating (accessed 2026-10-21)
- Agent Skills specification: https://agentskills.io/specification (accessed 2026-10-21)
- Gemini CLI hooks reference: https://geminicli.com/docs/hooks/reference/ (accessed 2026-10-21)
- Devin Desktop skills: https://docs.devin.ai/desktop/cascade/skills (accessed 2026-10-21)
- Devin Desktop hooks: https://docs.devin.ai/desktop/cascade/hooks (accessed 2026-10-21)
- Kiro hooks: https://kiro.dev/docs/hooks/index (accessed 2026-10-21)

## Related

- [Platform hooks are written by init and reload, per verified provider](platform-hooks-per-provider.md)
- [Keep only platforms with a verified instruction file](provider-verification-and-removals.md)
- [Plugin distribution formats (verified 2026-10-21)](../memory/plugin-distribution-formats-2026-10-21.md)
