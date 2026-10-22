---
title: "Platform hooks are written by init and reload, per verified provider"
updatedAt: "2026-10-21"
---

# Platform hooks are written by init and reload, per verified provider

**Status:** accepted · 2026-10-21

## Context

plaesy built the design-lint check as a Claude-only hook. Other harnesses have hook systems too,
but each uses its own file and schema. Only a hook written from a verified schema is safe to ship.

## Decision

- `platform.json` declares `mapping.hooks` (manifest path) and `mapping.hooks_format` per platform.
- `plaesy init --ai <p>` and `plaesy reload --ai <p>` write the advisory design-lint hook for
  Claude, Codex, Cursor, Copilot, Gemini, Windsurf, Qoder, and Codewhale.
- Shared JSON files keep every entry that is not plaesy's, identified by the `design-lint.` marker.
  An unparsable file is left alone. Codewhale's `hooks.toml` is written only if it has the marker
  `# managed by plaesy`.
- The hook is advisory. It never blocks an edit: it always exits 0.
- Hooks follow `--ai`. `reload` without `--ai` refreshes no platform files, hooks included.
- Providers without a verified hook surface get no hook: Tabnine, Cline (its script hooks are not
  on the current official page), OpenCode (JS plugin, payload undocumented), Kilo, Continue.

## Consequences

- Copilot's `.github/hooks/plaesy-design-lint.json` is committed, so it reaches the team through git.
- `.claude*` is gitignored, so the Claude hook stays local unless the ignore rule changes.
- Codewhale and Codex need the user to approve the project hook once (`/hooks` or `/hooks approve`).
- On Windows, the `sh` command needs Git Bash on PATH for Claude, Cursor, Gemini, Qoder, and Codewhale.

## References

- Codex hooks: https://learn.chatgpt.com/docs/hooks (accessed 2026-10-20)
- Cursor hooks: https://cursor.com/docs/agent/hooks (accessed 2026-10-20)
- Copilot CLI hooks: https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/use-hooks (accessed 2026-10-20)
- Gemini CLI hooks: https://geminicli.com/docs/hooks/ (accessed 2026-10-20)
- Windsurf / Devin Cascade hooks: https://docs.devin.ai/desktop/cascade/hooks (accessed 2026-10-20)
- Qoder CLI hooks: https://docs.qoder.com/cli/hooks (accessed 2026-10-20)
- Codewhale hooks: https://github.com/Hmbown/DeepSeek-TUI/blob/main/docs/HOOKS.md (accessed 2026-10-21)

## Related

- Provider research: `.plaesy/memory/ai-provider-research-2026-10-20.md`
- Hook details per harness: `.plaesy/memory/ai-provider-hooks-2026-10-20.md`
