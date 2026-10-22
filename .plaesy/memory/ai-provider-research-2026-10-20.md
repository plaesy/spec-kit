---
name: ai-provider-research-2026-10-20
description: Which AI providers plaesy supports, which instruction paths are verified against official docs, and what was added on 2026-10-20 (Codex, Gemini CLI, Aider, hook manifests)
metadata:
  type: reference
---

# AI provider research (2026-10-20)

Question asked: why does plaesy target Claude first, and what about other providers.

## Verified instruction paths
| Provider | Project instruction file | Source (retrieved 2026-10-20) |
|---|---|---|
| Claude Code | `CLAUDE.md`, `.claude/commands`, `.claude/agents` | in use |
| OpenAI Codex CLI | `AGENTS.md` (global `~/.codex`, cascades root to cwd, 32 KiB cap via `project_doc_max_bytes`); custom prompts in `.codex/prompts/` | learn.chatgpt.com/docs/agent-configuration/agents-md |
| Gemini CLI | `GEMINI.md` (global `~/.gemini/GEMINI.md`; `context.fileName` may list `AGENTS.md`; `@file` imports) | geminicli.com/docs/cli/gemini-md |
| Aider | `CONVENTIONS.md`, loaded by `--read` or `read:` in `.aider.conf.yml` | aider.chat/docs/usage/conventions.html |
| GitHub Copilot CLI | `.github/copilot-instructions.md`; `.github/instructions/*.instructions.md` with `applyTo`; reads `AGENTS.md`, `CLAUDE.md`, `GEMINI.md` | docs.github.com … add-custom-instructions |

## Not verified
About 13 entries in `platform.json` (Cline, Continue, Tabnine, Codeium, CodeWhisperer, Studio Bot,
Replit, LlamaIndex, Ollama, LM Studio, DeepSeek, Trae, Windsurf) use paths I could not match to any
official source, for example `.cline/instructions.md` and `.ollama/config.json`. They are kept,
not removed, because removing a platform changes `plaesy init` output. Treat them as unverified.

## Why Claude was first
The `.claude/` wiring (commands, agents, hooks) and the framework's own `CLAUDE.md` were built first.
Other platforms were stubs in `platform.json`: paths only, no verified content, no hooks.

## Added on 2026-10-20
- `codex`, `gemini`, `aider` in `platform.json`. Gemini and Aider map only `core`. Codex also maps
  `prompts` (its custom prompts are Markdown). Agents are not mapped for Codex or Gemini, because
  `plaesy` copies agents as `.md` and their native formats differ (TOML).
- Hook manifests for Claude, Codex, Cursor, Copilot CLI and Gemini CLI. See
  [[ai-provider-hooks-2026-10-20]].

Related: [[ai-provider-hooks-2026-10-20]]

## Update: all providers checked (2026-10-20)

- Verified rules paths: Continue `.continue/rules/*.md` (frontmatter optional; undefined
  `alwaysApply` with no globs means always included); Tabnine `.tabnine/guidelines/*.md`;
  Kilo `instructions` entries in `kilo.jsonc`, examples under `.kilo/rules/`; Qoder `.qoder/rules/`
  (100,000-character limit across active rules; reads AGENTS.md, rules win on conflict).
- Still unverified: Windsurf rules file (the hooks file is verified, the rules file is not),
  Cline rules, Trae rules (sources disagree between `.trae/project_rules.md` and `.trae/rules/`),
  DeepSeek, Amazon Q rules, Ollama, LM Studio, LlamaIndex, Replit, Studio Bot, Codewhisperer, Codeium.
- Gemini CLI may be replaced by Antigravity CLI on 2026-06-18 (one source). Not yet confirmed.

## Final verdict per platform (checked 2026-10-21)

Verified rules path (mapping set or confirmed):
- Cline `.clinerules/` (docs.cline.bot/customization/cline-rules); mapped `core` to `.clinerules/plaesy.md`.
- Windsurf/Devin: root `AGENTS.md` is always-on (docs.devin.ai/desktop/cascade/agents-md); mapped `core` to `AGENTS.md`. Rules dir is `.devin/rules/` (legacy `.windsurf/rules/`). Hooks in `.devin/hooks.json`.
- DeepSeek / Codewhale: reads `AGENTS.md` by default (github.com/Hmbown/DeepSeek-TUI); mapped `core` to `AGENTS.md`.
- Replit: `replit.md` in the project root (docs.replit.com); mapped.
- Android Studio (Gemini, the `studio_bot` entry): `AGENTS.md`, needs Narwhal 4 canary or newer (developer.android.com); mapped.
- Kiro (successor to Amazon Q Developer, end of life): `.kiro/steering/*.md` (docs.aws.amazon.com agent-toolkit rules-files); new platform `kiro`.
- Gemini CLI: consumer access ended 2026-06-18; Antigravity CLI reads `GEMINI.md` and `AGENTS.md` (antigravity.google/docs/cli/gcli-migration). Mapping kept; name marks it as replaced.
- Qoder `.qoder/rules/` (already mapped).

Still unverified; left as they are:
- Trae (sources disagree: `.trae/project_rules.md` vs `.trae/rules/`). No official page found.
- Ollama, LM Studio, LlamaIndex: no project instruction file in official docs (checked docs.ollama.com and lmstudio.ai/docs). These entries write only prompts/agents folders, so they are safe but pointless. Recommend removal.
- Cline `.cline/prompts`, Windsurf prompts, and similar invented prompt folders: no verified path.
- Codeium `.codeiumrc`: not found in any source.
- Amazon Q project path: the official page gives `~/.aws/amazonq/cli-agents` only.

Hooks not written:
- Codewhale (DeepSeek-TUI): `.codewhale/hooks.toml` (TOML, project hooks need trust approval). Its `tool_call_after` payload has no file path; the script's git fallback would work. Not implemented yet.
- Cline script hooks: the current docs describe SDK plugins instead.

## Removed (2026-10-21)
Ollama, LM Studio, LlamaIndex, Codeium, Amazon CodeWhisperer, and Trae were removed from
`platform.json`, together with the `trae` alias. None had a verified project instruction file.
Kept: `ollama` as a delegate provider in the `delegate` package (a local model endpoint,
unrelated to platform files).
Amazon Q Developer is replaced by the `kiro` platform.
