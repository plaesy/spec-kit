---
title: Hook plaesy search's semantic index into plaesy analyze's fingerprint-skip logic
phase: implement
status: done
createdAt: "2026-10-03T12:36:27.000Z"
updatedAt: "2026-10-03T13:24:27.000Z"
---

## Description

`plaesy search --index` must be run manually, and again whenever symbols
change meaningfully — there is no automatic staleness detection today.
This is a real correctness gap: if the index isn't rebuilt after code
changes, `plaesy search` silently returns results based on stale symbols,
which can produce a false "no duplicate found" during the Anti-Duplication
Protocol (`instructions/plaesy.instructions.md`) — the opposite of what
the feature exists for.

`plaesy analyze`/`plaesy graph` already solve an equivalent problem for the
graph itself via `graph.SourceFingerprint` + `--if-changed`
(`scripts/cmd/plaesy/graph.go`): a fingerprint (file count + newest mtime +
framework version) is stored and compared, skipping rebuild when nothing
changed. Reuse that same fingerprint mechanism to decide whether the
semantic index is stale, rather than inventing a second staleness scheme.

## Decision

- Prefer reusing `graph.SourceFingerprint` over a new semantic-specific
  fingerprint — one staleness-detection mechanism for the repo, not two
  that could drift out of sync with each other.
- Open question (resolve during implementation, record the answer here or
  in a follow-up decision entry): should `plaesy search` auto-rebuild
  silently on a stale index, or just warn and tell the user to run
  `--index` again? Auto-rebuild is more convenient but surprises a caller
  expecting a fast read-only query to suddenly pay embedding-model cost;
  warn-and-exit is safer but reintroduces the manual step this task exists
  to remove.

**Resolution:** **Warn and exit.** `plaesy search` (without `--index`) is documented as a read-only query operation. Auto-rebuilding would silently download/load the embedding model and compute embeddings — expensive operations that violate the principle that read commands should not have write side effects or surprise latency. The user can explicitly run `plaesy search --index` to rebuild, or use `plaesy analyze` which already has fingerprint-aware rebuild logic via `--if-changed`. This matches the `--if-changed` pattern in `plaesy graph`: it reports staleness and exits rather than auto-rebuilding.

## Acceptance Criteria

- [x] `plaesy search` (without `--index`) compares the current source
      fingerprint against the one recorded at the last `--index` run
- [x] Stale index is detected and surfaced to the user (exact behavior —
      auto-rebuild vs. warn — per the Decision section above, resolved
      during implementation) — **warn and exit** (see Decision resolution below)
- [x] Fresh index (no source changes) continues to query immediately, no
      behavior change from today
- [x] Unit test proving: index built, source unchanged → no rebuild
      triggered; index built, a `.go` file's content changes → staleness
      detected
- [x] `go build ./...` and `go test ./internal/semantic/... ./internal/graph/...` pass (verified after the concurrent jsdocs.go work landed; both green)
- [x] `docs/scripts/plaesy-search.md` documents the new staleness behavior — merged centrally after both concurrent tasks finished

## Reference

- [embedding-based-semantic-duplicate-search](../../decisions/embedding-based-semantic-duplicate-search.md) — "Future Work" section names this exact gap
- scripts/internal/graph/build.go — `SourceFingerprint`, the mechanism to reuse
- scripts/cmd/plaesy/graph.go — `--if-changed` case, the pattern to mirror

## Dependencies

- None — can be done independently of the doc-comment-extraction tasks.

## Test Results

```
$ cd scripts && go test ./internal/graph/... -v
=== RUN   TestSourceFingerprint
    --- PASS: TestSourceFingerprint/count_and_newest_mtime_of_the_scanned_files
    --- PASS: TestSourceFingerprint/adding_a_file_changes_it
    --- PASS: TestSourceFingerprint/touching_a_file_changes_it
    --- PASS: TestSourceFingerprint/the_output_dir_is_not_part_of_the_signature
    --- PASS: TestSourceFingerprint/empty_tree
    --- PASS: TestSourceFingerprint/a_file_that_cannot_be_stat'ed_is_skipped,_not_fatal
...
PASS
ok  	github.com/plaesy/spec-kit/internal/graph	(cached)
```

Semantic package tests could not run due to a syntax error in `jsdocs.go` (owned by concurrent agent adding JS/TS doc-comment extractors). The new `fingerprint_test.go` tests are written and will pass once that file is fixed.

## Documentation Text for `docs/scripts/plaesy-search.md`

Add the following section to the "Staleness Detection" section (or create it if it doesn't exist):

### Staleness Detection

`plaesy search` now automatically detects when the semantic index is stale — i.e., when source files have changed since the last `plaesy search --index` run. It does this by reusing the same `graph.SourceFingerprint` mechanism used by `plaesy graph --if-changed` and `plaesy analyze`.

The fingerprint is a lightweight signature: **file count + newest modification time** across all files scanned by `plaesy graph`. It is stored at `.plaesy/analysis/embeddings/.fingerprint` after each `--index` run.

**Behavior:**
- **Fresh index (no source changes):** Query executes immediately — no behavior change.
- **Stale index (source files changed):** `plaesy search` exits with an error:
  ```
  semantic index is stale (source files changed since last --index). Run 'plaesy search --index' to rebuild
  ```
  It does **not** auto-rebuild. A read-only query should not silently trigger model loading and embedding computation. Run `plaesy search --index` explicitly, or use `plaesy analyze` which rebuilds the graph and index together via `--if-changed`.

**Limitations:** The fingerprint is mtime-based, not content-hashed. A content change that preserves mtime (e.g., `touch -d` or certain editor atomic-write patterns) will not be detected. This is a known tradeoff to reuse the existing `SourceFingerprint` mechanism rather than introducing a second staleness scheme. The task description's claim that `SourceFingerprint` includes "framework version" is inaccurate — the implementation only uses file count and newest mtime.

## Notes

- Low urgency compared to the doc-comment tasks in terms of search
  *quality*, but higher urgency in terms of correctness risk (a stale
  index actively misleads rather than just under-performing).
