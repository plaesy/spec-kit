---
title: "Keep only platforms with a verified instruction file; never map core onto a config file"
updatedAt: "2026-10-21"
---

# Keep only platforms with a verified instruction file; never map core onto a config file

**Status:** accepted · 2026-10-21

## Context

Many `platform.json` entries were placeholders: paths I could not match to any official
documentation. Nine of them mapped `core` onto a provider's configuration file, for example
`.continue/config.json` or `.tabnine/config.json`. `reload --ai <p>` uses the overwrite policy,
so it would have replaced the user's model and provider settings.

## Decision

- A platform stays only if an official source names a project instruction file for it.
  Verified mappings: Claude, Codex (`AGENTS.md`), Gemini (`GEMINI.md`), Aider (`CONVENTIONS.md`),
  Copilot, Cursor, Cline (`.clinerules/`), Windsurf (root `AGENTS.md`), Continue
  (`.continue/rules/`), Tabnine (`.tabnine/guidelines/`), Qoder (`.qoder/rules/`), Kilo, Replit
  (`replit.md`), Android Studio (`AGENTS.md`), Kiro (`.kiro/steering/`), Codewhale (`AGENTS.md`).
- Removed for lack of a verified file: Ollama, LM Studio, LlamaIndex, Codeium, Amazon CodeWhisperer,
  Trae. Ollama stays as a delegate provider, which is a different feature.
- `core` never ends in `config.json`. `TestNoCoreMappingWritesOverAConfigFile` enforces this.
- Kiro replaces Amazon Q Developer, which is end-of-life.
- Gemini CLI consumer access ended 2026-06-18. Antigravity CLI still reads `GEMINI.md` and
  `AGENTS.md`, so the Gemini mapping stays and the name says it was replaced.

## Consequences

- Platforms with no verified file write only their prompt and agent folders, which are unverified
  too. They should be checked before use.
- The Trae rule path is still disputed between sources. Re-check it when an official page is found.

## References

- Cline rules: https://docs.cline.bot/customization/cline-rules (accessed 2026-10-21)
- Windsurf / Devin AGENTS.md: https://docs.devin.ai/desktop/cascade/agents-md (accessed 2026-10-21)
- DeepSeek-TUI / Codewhale: https://github.com/Hmbown/DeepSeek-TUI (accessed 2026-10-21)
- Replit replit.md: https://docs.replit.com/references/configuration/replit-dot-md (accessed 2026-10-21)
- Android Studio agent files: https://developer.android.com/studio/gemini/agent-files (accessed 2026-10-21)
- AWS agent toolkit rules files: https://docs.aws.amazon.com/agent-toolkit/latest/userguide/rules-files.html (accessed 2026-10-21)
- Tabnine guidelines: https://docs.tabnine.com/main/getting-started/tabnine-agent/guidelines (accessed 2026-10-21)
- Continue rules: https://docs.continue.dev/customize/deep-dives/rules (accessed 2026-10-21)
- Kilo custom rules: https://kilo.ai/docs/customize/custom-rules (accessed 2026-10-21)
- Qoder rules and hooks: https://docs.qoder.com/cli/hooks (accessed 2026-10-21)
- Antigravity CLI migration: https://antigravity.google/docs/cli/gcli-migration (accessed 2026-10-21)

## Related

- `.plaesy/memory/ai-provider-research-2026-10-20.md`
