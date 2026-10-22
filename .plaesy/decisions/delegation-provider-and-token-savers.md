---
title: "Decision: Delegation provider interface and token-saver patterns"
description: "Plaesy-level delegate providers (cli/API) with fallback chains, type-aware trim filters; no gateway dependency, no account pooling"
updatedAt: "2026-10-13T00:00:00.000Z"
decided: "2026-10-13"
status: "Accepted — not yet implemented"
---

# Decision: Delegation provider interface and token-saver patterns

**Status**: Accepted — not yet implemented
**Decided**: 2026-10-13
**Deciders**: user (requested the decision) + session agent
**Related**: decisions.md, `.plaesy/instructions/context-engineering.md` (§2 `plaesy trim`), `.plaesy/instructions/multi-agent-patterns.md`, `scripts/internal/imagegen` (provider precedent), `scripts/internal/reach` (external CLI precedent)

## Context and Problem Statement

Plaesy is a set of prompts plus a Go binary. The host (Claude Code, Kilo) runs the LLM loop, so Plaesy cannot intercept every model call. Three needs came up:

1. **Token compression.** `plaesy trim` exists but is heuristic only (filler stripping, whitespace collapse). It does not compress by type of output, and it has no model-based summarization.
2. **Delegation.** Plaesy agents run on the host model only. There is no way to send a bounded sub-task to another CLI (`claude`, `kilo`) or another API.
3. **Hooks.** The idea was hooks that compress or delegate automatically. The host hook APIs for Claude Code and Kilo are not verified for prompt rewriting or tool-call interception (Kilo's plugin hooks cover config and events only, as checked on 2026-10-13).

Two external projects were reviewed for patterns:

- **Kilo Code** (`Kilo-Org/kilocode`, MIT, CLI v7.8.9 on 2026-10-13). Source confirms a permission ruleset with default-deny for agents, `max-steps` forced summary, and an `--auto` flag where explicit deny still wins. Orchestrator mode is in source but documented as deprecated.
- **9router** (`decolua/9router`, MIT, pushed 2026-10-08). Source confirms RTK tool-result filters (`open-sse/rtk/filters/`: git diff/log/status, grep, ls, find, tree, smart truncate), applied fail-open (`safeApply` returns raw output on filter error). Its combo service orders fallback with hard capabilities (vision, pdf) first and classifies errors with retry-after. It runs as a separate gateway at `localhost:20128`.

## Options Considered

| Option | Verdict | Reason |
|---|---|---|
| **A. Plaesy-level `plaesy delegate` with provider interface** (cli, anthropic, openai, gemini), same pattern as `imagegen` | **Chosen** | Works on any host. No external process. Matches existing precedent. |
| B. Add the 9router gateway as a dependency | Rejected | Adds a long-running external process at `localhost:20128`. The Ollama rejection in `embedding-based-semantic-duplicate-search.md` applies the same way: Plaesy is a single offline binary. |
| C. Host-level hooks (Claude Code, Kilo plugins) for automatic compression | Deferred | Host APIs not verified. Not portable across hosts. Revisit after verification. |
| D. Multi-account round-robin and OAuth piggybacking on subscriptions (9router feature) | Rejected | Reusing subscription credentials across accounts risks breaking provider terms. Handling OAuth tokens for other tools conflicts with rule 8 (credentials are a hard stop). Needs `assess-legal` review before any reconsideration. |
| E. Model-based compression only | Rejected as sole option | Costs tokens, non-deterministic. Kept as one provider behind the heuristic default. |
| F. Type-aware filters in `plaesy trim run` (RTK pattern), fail-open | **Chosen** | Deterministic, no network, no new dependency. Largest share of savings comes from tool output such as diffs and listings. |
| G. Fallback chain with error classification and capability gate (combo pattern) | **Chosen** for the delegate provider | Needed once more than one provider exists. |

## Decision Outcome

1. **Hooks are Plaesy-level commands, not host hooks.** `plaesy trim run` and `plaesy delegate` are called explicitly by instructions or commands. Host hooks are deferred until their APIs are verified.
2. **Delegate provider interface** follows the `imagegen` pattern. Initial providers: `cli` (exec `claude` or `kilo`) and `anthropic`. Then `openai` and `gemini`.
3. **Token savings in two layers.** Type-aware filters with fail-open behavior come first. Model-based compression is an optional provider, off by default.
4. **Fallback chains are named config** (the combo idea). Providers that lack a required capability are excluded, not degraded. Retry-after is honored.
5. **No gateway dependency. No account pooling. No credential piggybacking.**
6. **Hard-stop rule 8 applies.** A paid provider call requires explicit user confirmation for each provider. The API key comes from an environment variable and is never stored in the repo. Sending project content to a third-party provider requires an `assess-legal` check of the data policy before first use.

## Dependencies

- Per-agent permission ruleset and budget enforcement (Kilo `permission.ts` and `max-steps.ts` patterns) are prerequisites for delegation. They are tracked separately and not part of this decision.

## Consequences

- Tasks in `.plaesy/tasks/backlog/` cover type-aware filters, the provider interface, the fallback chain, and per-provider usage tracking.
- No new external dependency in `go.mod`.
- Host hooks remain unresolved. The decision must be revisited if a host exposes a verified prompt-rewrite or tool-call hook.

## Sources

- Kilo Code source, retrieved 2026-10-13: `packages/core/src/permission.ts`, `packages/core/src/session/runner/max-steps.ts`, `packages/opencode/src/cli/cmd/run.ts`; docs `packages/kilo-docs/pages/code-with-ai/agents/orchestrator-mode.md`
- 9router source, retrieved 2026-10-13: `open-sse/rtk/index.js`, `open-sse/rtk/applyFilter.js`, `open-sse/services/combo.js`; README
- Plaesy: `.plaesy/instructions/context-engineering.md`, `scripts/internal/imagegen/imagegen.go`, `scripts/internal/reach/base.go`

## Measured Filter Savings (type-aware filters, 2026-10-13)

Measured with the built binary on real commands in this repo (`plaesy trim run -- <cmd>`, then `trim report`). Token counts use the existing chars/4 estimator. Savings include the shared head/tail truncation, not the filter alone.

| Command | Filter | Before (tokens) | After (tokens) | Saved |
|---|---|---|---|---|
| `git diff` | git-diff | 49693 | 943 | 98.1% |
| `git status` | git-status | 1569 | 293 | 81.3% |
| `grep -rn func scripts/internal/trimmer` | grep | 4299 | 1159 | 73.0% |
| `git log -n 15` | git-log | 39 | 17 | 56.4% (repo has one commit, so the sample is tiny) |
| `ls -la scripts/internal/trimmer` | listing | 207 | 205 | 1.0% |

The listing filter saves almost nothing. Its only effect is dropping the `total` line, so it stays because it is cheap, not because it pays off.

Sources: repo state at measurement time; `scripts/internal/trimmer/filters.go`.
