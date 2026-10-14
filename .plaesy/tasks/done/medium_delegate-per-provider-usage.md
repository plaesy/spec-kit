---
title: "Per-provider usage and budget tracking for plaesy delegate"
phase: implement
status: backlog
createdAt: "2026-10-13T00:00:00.000Z"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Task: Per-provider usage and budget tracking

## Description

Track tokens and wall time per delegate provider, and enforce a budget before a call is made. Extends the existing `.plaesy/memory/token-stats.json` log, which `plaesy trim` already writes. The quota-tracking idea comes from 9router's usage analytics.

Decision: `.plaesy/decisions/delegation-provider-and-token-savers.md`.

## Objective

A user can see how much each provider consumed and can stop calls before a budget is exceeded.

## Scope

- Add a `provider` field to each record in `token-stats.json`. Keep old records valid.
- `plaesy delegate report` prints totals per provider and per chain.
- Budget per provider per day, in tokens. Checked before each call. Exceeding it is a terminal error, not a silent skip.
- Budget values come from config. Defaults are zero, which means every paid call needs `--confirm`.

## Acceptance Criteria

- [ ] Old `token-stats.json` records still parse. Covered by a test using a pre-change fixture.
- [ ] A call that would exceed the budget is refused before any network request. Covered by a test with a fake provider that fails the test if called.
- [ ] `report` output matches the records in a fixture file.
- [ ] No budget configured means the defaults above apply, not unlimited spend.

## Blocked By

- `high_delegate-provider-interface`

## Out of Scope

- Billing integration with provider dashboards.
- Currency conversion.

## References

- `.plaesy/instructions/context-engineering.md` §2 (token-stats log).
- 9router real-time quota tracking, retrieved 2026-10-13.

## Implementation Status (2026-10-13)

Implemented in code, tests written, NOT yet run (tests deferred by request).

- Log location: `.plaesy/analysis/delegate-usage.jsonl`, one JSON record per line, appended. This follows the directive to keep it under the analysis directory. The task's original plan to extend `.plaesy/memory/token-stats.json` is NOT done, so the trimmer's log is not touched. Old-record compatibility is covered by `TestReadUsage_OldRecordsStillParse` for this new log.
- Records carry `provider`, `chain`, `billable`, token counts (nil when unknown), and wall time.
- `plaesy delegate report` prints totals per provider and per chain. Calls without token counts are shown in their own column.
- Budget: `daily_token_budget` per provider in config, checked by `CheckBudget` before each call. A refusal is Terminal and no request is sent.
- Tests: `internal/delegate/usage_test.go`. Not run.

Acceptance criteria, as code state:
- [x] Old records parse (`TestReadUsage_OldRecordsStillParse`), untested.
- [x] A call over budget is refused before any request (`TestRunChain_BudgetRefusalSendsNothing` and `TestCheckBudget_RefusesBeforeCall`), untested.
- [x] `report` totals match a fixture (`TestSummarize_MatchesFixture`), untested. The report command's output format is not tested.
- [x] No budget configured means the per-call `--confirm` gate still applies (`TestCheckBudget_RefusesBeforeCall`, zero-budget case), untested.

ASSUMED defaults:
- Budget tokens are estimated as prompt length / 4 plus the output allowance. This is an estimate, not a count.
- Zero budget means no extra budget limit; a billable call still needs `--confirm`. The task text read as "every paid call needs --confirm", which matches.
- Budget is per UTC day.
