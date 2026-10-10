---
title: plaesy analyze rebuilds the semantic index so plaesy search works right after it
phase: implement
status: done
createdAt: "2026-10-04T10:45:00.000Z"
updatedAt: "2026-10-04T10:45:00.000Z"
---

## Description

`plaesy search` refused to query a stale index and told the user to run
`plaesy search --index`. Since the staleness fingerprint is mtime-based, every
code edit invalidated the index — so the normal loop was: edit → `plaesy
analyze` → `plaesy search` → *error naming another command* → `--index` →
`search`. The command whose entire job is refreshing project knowledge did not
refresh the index.

The user asked for `plaesy analyze` to do "what `plaesy search --index` does",
without stating a default.

## Decision

- **Default on, `--no-index` opts out.** ASSUMED default, recorded per rule 2
  ("default and record, don't park"). Two precedents agree: `project.symbols.md`
  is built by default in analyze with no flag, and so are `project.html` and
  `reports.md`. `--no-index` mirrors `--no-graph`, which exists for the same
  reason (the step costs a model load plus a full re-embed).
- **Build after the graph and the symbols index**, because the index embeds the
  graph's symbols. An index built from a stale graph is exactly the stale index
  `plaesy search` refuses.
- **Gate on `semantic.IsIndexStale`**, the same check `plaesy search` runs before
  every query — one definition of "fresh" for the repo, and an unchanged project
  pays no model load and no embedding. `--force` bypasses the gate, as it does
  for the graph.
- **The fingerprint fast path still re-checks the index.** "Skipping
  regeneration" is a claim about the analysis artifacts; the index is not one of
  them. A first run that could not load the model (offline) leaves it stale or
  missing, which is a state `plaesy search` refuses — so the cached path heals it.
- **A failed index build is non-fatal and does not suppress the analysis
  fingerprint.** The fingerprint claims `project.json`, the graph artifacts and
  `project.symbols.md` exist — all three were written. Folding the index failure
  into `graphFailed` would mean a machine that cannot load the embedding model
  re-runs the entire analysis on every invocation, forever, over a step that is
  not analysis.

## Acceptance Criteria

- [x] `plaesy analyze` builds `embeddings/` after a successful graph build
- [x] An unchanged project pays no model load and no embedding (gated on
      `semantic.IsIndexStale`)
- [x] The fingerprint fast path still checks the index, so a missing or stale one
      is healed on the cached path
- [x] `--no-index` skips the step; `--no-graph` skips it too (no symbols to embed)
- [x] A failed index build leaves `Run` successful, records the analysis
      fingerprint, and names `plaesy search --index` as the retry
- [x] "Generated files" lists the index honestly — written, or explicitly NOT
      rebuilt
- [x] Tests cover all of the above through a seam, so no unit test loads a model
- [x] `go build ./...`, `gofmt -l`, `go vet ./...`, `go test ./...` (19/19
      packages), `plaesy validate markdown` (527 files, 0 violations)
- [x] Verified on this repo for real: analyze built the index, the next
      `plaesy search` query ran without a staleness error, and the analyze run
      after it reported `embeddings/ - Semantic search index (fresh)`

## Reference

- [Embedding-based semantic duplicate search](../../decisions/embedding-based-semantic-duplicate-search.md) — "Update 2026-10-04" records the default-on choice and its reasoning
- [Hook semantic index into analyze fingerprint-skip logic](medium_hook-semantic-index-into-analyze-fingerprint.md) — the earlier half: the staleness check this reuses
- `scripts/internal/analyze/analyze.go` — `ensureSemanticIndex`
- `scripts/internal/semantic/fingerprint.go` — `IsIndexStale`, the shared gate

## Dependencies

- None. Builds on the staleness work in `done/medium_hook-semantic-index-into-analyze-fingerprint.md`.

## Notes

- The seam is a package-level `var ensureSemantic = ensureSemanticIndex`, the
  same pattern as the existing `Logf`/`Successf` vars. The e2e smoke test passes
  `--no-index` for the same reason: a test whose result depends on the machine's
  network is not a test.
- Docs updated together with the code: `docs/reference.md` flag table,
  `docs/scripts/plaesy-analyze.md` (§5), `docs/scripts/plaesy-search.md`
  ("Building the index"), Anti-Duplication step 1b in
  `instructions/plaesy.instructions.md` (plus its embedded mirror), CHANGELOG.
- Found while doing it, not fixed by it: `--no-graph` records the analysis
  fingerprint although it writes no graph — parked as a backlog task.
