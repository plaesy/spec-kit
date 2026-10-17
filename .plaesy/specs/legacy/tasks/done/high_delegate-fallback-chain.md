---
title: "Named provider fallback chains with error classification and capability gate"
phase: implement
status: backlog
createdAt: "2026-10-13T00:00:00.000Z"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Task: Named fallback chains for `plaesy delegate`

## Description

When a delegate provider fails or runs out of quota, move to the next provider in a named chain. Pattern from 9router's combo service (`open-sse/services/combo.js`, `accountFallback.js`). Decision: `.plaesy/decisions/delegation-provider-and-token-savers.md` (Option G).

## Objective

A chain such as `primary → cheap → local` is declared once in config. Calls fall through it in order, and the caller sees which provider answered.

## Scope

- Chains are named entries in the Plaesy config, for example `chains.review = ["cli:claude", "anthropic"]`.
- Error classes:
  - **Retryable with wait** (rate limit, quota): honor retry-after if given, then try the next provider.
  - **Retryable now** (timeout, 5xx): try the next provider.
  - **Terminal** (auth failure, bad request): stop the chain and report. Do not fall through, because the same request will fail elsewhere too.
- Capability gate: each provider declares its capabilities. A provider lacking a required capability is skipped, not degraded. Start with one required capability, text input.
- Output records every provider attempted and why it was skipped or failed.

## Acceptance Criteria

- [ ] Table-driven test covers each error class and confirms the chosen next step.
- [ ] Terminal errors never fall through. Covered by a test.
- [ ] Retry-after is respected. Covered by a test with a fake clock.
- [ ] A chain with no answering provider returns an error that lists every attempt. It does not return an empty success.
- [ ] A paid provider in a chain still requires `--confirm` (see `high_delegate-provider-interface`).

## Blocked By

- `high_delegate-provider-interface`

## Out of Scope

- Account rotation within one provider.
- Subscription or OAuth credential reuse (rejected in the decision).

## References

- 9router `open-sse/services/combo.js` and `open-sse/services/accountFallback.js`, retrieved 2026-10-13.

## Implementation Status (2026-10-13)

Implemented in code, tests written, NOT yet run (tests deferred by request).

- Chains: `delegate.chains` in config (`internal/config/delegate.go`), run with `plaesy delegate --chain <name>`.
- Error classes: `internal/delegate/errors.go` (Transient, RateLimited, Terminal). Unclassified errors are Terminal.
- Capability gate: `Capabilities()` on each provider, `CapText` required. A provider without it is skipped.
- Output: every attempt is recorded (`answered`, `skipped`, `failed`) and returned with the answer, or in the error.
- Tests: `internal/delegate/chain_test.go`, `internal/delegate/classify_test.go`. Not run.

Acceptance criteria, as code state:
- [x] Table-driven test per error class (`TestRunChain_ErrorClasses`), untested.
- [x] Terminal errors never fall through (same test, `second.calls == 0`), untested.
- [x] Retry-after respected with a fake sleep (`TestRunChain_RetryAfterRespected`), untested.
- [x] No answer returns an error listing every attempt (`TestRunChain_NoAnswerListsEveryAttempt`), untested.
- [x] Billable provider in a chain still requires `--confirm` (`TestRunChain_BillableNeedsConfirm`), untested.

ASSUMED defaults:
- Retry-after maximum wait is 60s. A longer Retry-After moves to the next provider without waiting.
- An unclassified error is Terminal, because falling through could repeat a billed request.
- A disabled, unknown, or billable-without-confirm provider in a chain is skipped, not fatal.
- A config error while building a provider (for example a missing command) is Terminal.
