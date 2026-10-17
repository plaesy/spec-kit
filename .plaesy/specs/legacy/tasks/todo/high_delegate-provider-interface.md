---
title: "plaesy delegate: provider interface with cli and anthropic providers"
phase: implement
status: backlog
createdAt: "2026-10-13T00:00:00.000Z"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Task: `plaesy delegate` provider interface

## Description

Let Plaesy send a bounded sub-task to another model, either through another CLI (`claude`, `kilo`) or an API. Decision: `.plaesy/decisions/delegation-provider-and-token-savers.md` (Option A).

Follow the provider pattern in `scripts/internal/imagegen/imagegen.go`: a `Provider` field selects the implementation, and unknown providers return an error that lists the supported ones.

## Objective

A single entry point, `plaesy delegate`, that runs a prompt against a named provider and returns structured output.

## Scope

- Interface: `Complete(ctx, request) (response, error)` with a request that carries prompt, max output tokens, and a timeout.
- Providers, in order:
  1. `cli`: exec a configured command (`claude`, `kilo`), with the prompt on stdin. Reuse the exec pattern from `scripts/internal/reach/base.go`.
  2. `anthropic`: HTTP call, API key from an environment variable.
- Each call returns the provider name, token counts when available, and wall time.
- Output is written to stdout as JSON, or to a file with `--out`.

## Acceptance Criteria

- [x] Interface has a fake provider used by the tests. No test makes a network call.
- [x] `cli` provider: a timeout kills the child process. Covered by a test.
- [x] `anthropic` provider: the API key is read from the environment only. A test confirms the key is never written to output, logs, or files.
- [x] Calling a paid provider without `--confirm` fails with a clear message. This implements rule 8 (money/billing hard stop).
- [x] No new entries in `go.mod` for the first two providers. Standard library HTTP only.
- [ ] `plaesy doctor` reports which providers are configured, without printing keys.

## Blocked By

- `high_delegate-fallback-chain` is needed for multi-provider use, but single-provider calls can ship first.

## Out of Scope

- `openai` and `gemini` providers (follow-up task after this one passes).
- Host hooks.
- Account rotation or subscription pooling (rejected in the decision).

## References

- `.plaesy/decisions/delegation-provider-and-token-savers.md`
- `scripts/internal/imagegen/imagegen.go`, `scripts/internal/reach/base.go`

## Status (implementation pass)

**Built**
- `scripts/internal/delegate/`: `Provider` interface (`Complete`), `Request` and `Response` (provider name, token counts as nullable, wall time), `Select` (enabled and billable checks), `Build`, `Run` (applies the timeout). Providers: `cli` (stdin prompt, `WaitDelay` so a stuck grandchild cannot hang Wait), `anthropic` (stdlib HTTP, key read from an env var per call, never placed in a Response or error), `ollama` (local HTTP, not billable).
- `scripts/internal/config/delegate.go`: `delegate.providers` list. Each entry has `name`, `type`, `enabled`, `billable`, plus per-type fields. Added to `UnifiedConfig` as `Delegate`.
- `scripts/cmd/plaesy/delegate.go`: `plaesy delegate --provider <name> --prompt ... [--confirm] [--timeout] [--max-tokens] [--out]` and `--list`.

**Decisions the user made**
- Providers are chosen in config: `enabled` must be true, and a `billable` provider also needs `--confirm`.
- Providers in scope: CLI claude, CLI kilo, Anthropic API, Ollama.

**ASSUMED (defaults chosen here; check them)**
- Default providers: `claude-cli` enabled, not billable (runs on the user's own login). `kilo-cli` disabled by default. `anthropic` disabled and billable. `ollama` disabled, not billable.
- CLI arguments: `claude -p` and `kilo run`. The `kilo run` argument is unverified against the installed kilo version.
- Default model `claude-sonnet-5-5` for `anthropic`, `llama3.2` for `ollama`. Both are placeholders to confirm.
- CLI providers ignore `--max-tokens` (the CLI does not expose it here) and report no token counts.

**Not done**
- The fallback chain (`high_delegate-fallback-chain`) is out of this pass. Only single-provider calls exist.
- `plaesy doctor` does not yet list providers (acceptance criterion 6). `doctor.go` is outside this pass.
- `openai` and `gemini` are out of scope, as the task states.

**Untested (test run pending)**
- `go test ./internal/delegate/` has not been run. The tests cover: Select (disabled, billable without and with confirm, unknown name), Build errors, the CLI stdin echo, the CLI timeout that kills the child (uses the test binary as the child, no external CLI), the Anthropic parse and headers, the key never appearing in errors or JSON output (401 path), the missing-key message, and Ollama parse. No test calls a real network host; the HTTP tests use `httptest` on localhost.

## Status at /save (2026-10-13)

Criteria 1-5 are checked against tests that pass: fake provider and no network
(`chain_test.go`), CLI timeout (`TestCLI_TimeoutKillsChild`), the key never in output
(`TestAnthropic_KeyNeverInOutput`), the billable gate (`TestSelect_BillableNeedsConfirm`),
and `go.mod` (only the onnxruntime removal, no new entries).

Criterion 6 (`plaesy doctor` reports providers without printing keys) is implemented
in `cmd/plaesy/doctor.go` but no test covers it. The task stays open until a test
exists, so it moves to `todo/`.
