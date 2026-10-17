---
title: "Type-aware output filters for plaesy trim run (fail-open)"
phase: implement
status: done
createdAt: "2026-10-13T00:00:00.000Z"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Task: Type-aware output filters for `plaesy trim run`

## Description

`plaesy trim run` compresses command output with heuristics only (filler stripping, whitespace collapse). Add command-specific filters that detect the output type and compress it structurally, following the RTK pattern in 9router (`open-sse/rtk/`).

Decision: `.plaesy/decisions/delegation-provider-and-token-savers.md` (Option F).

## Objective

Cut tokens from tool output (diffs, listings, search results) without losing the lines an agent needs.

## Scope

- Filters for: `git diff`, `git status`, `git log`, `grep`/`rg`, `ls`/`find`/`tree`.
- Auto-detect the filter from the command name. Unknown commands use the existing heuristic path.
- Fail-open: if a filter panics or errors, pass the raw output through and write a warning to stderr. Mirrors `safeApply` in `open-sse/rtk/applyFilter.js`.
- Smart truncation for very long output keeps head, tail, and a count of omitted lines.
- Record per-filter savings in `.plaesy/memory/token-stats.json` (the file `plaesy trim report` already reads).

## Acceptance Criteria

- [x] Each filter has table-driven tests with real captured output samples.
- [x] A filter that returns an error produces the raw output and a stderr warning, never an empty result. Covered by a test.
- [x] `plaesy trim report` shows savings per filter.
- [x] Measured savings on at least three sample outputs are recorded in the decision file, not estimated.
- [x] Unknown commands behave exactly as before (regression test).
- [x] Coverage of the new package meets Constitution §1 (90%), or the gap is recorded as a decision.

## Out of Scope

- Model-based summarization (see the provider task).
- Changes to Layer 2 file compression.

## References

- 9router `open-sse/rtk/filters/*.js`, `open-sse/rtk/autodetect.js`, retrieved 2026-10-13.
- `.plaesy/instructions/context-engineering.md` §2.
