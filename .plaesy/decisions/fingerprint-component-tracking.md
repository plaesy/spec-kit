---
title: "Analysis fingerprint records the components it built"
status: accepted
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Analysis fingerprint records the components it built

## Context

`plaesy analyze --no-graph` recorded `.analysis-fingerprint` even though it wrote
no graph. The next plain `plaesy analyze` matched the fingerprint, skipped the
graph, and reported `project.graph.json` and `project.symbols.md` as cached. Those
files had never been written. Only `--force` recovered.

## Decision

- A `--no-graph` run records no fingerprint.
- A fingerprint records the components its run built, as a suffix:
  `<count>|<maxMtime>|<version>|components=structure,graph,symbols`.
- A skip requires the walk fingerprint to match and every component the current run
  needs to be recorded. `structure` is always needed. `graph` and `symbols` are
  needed unless `--no-graph` is set.
- A fingerprint without the `components=` tag (written before this change) records
  no components, so it never skips. The next plain run rebuilds and writes the tagged form.

## Rejected

- A JSON fingerprint file. The walk-fingerprint string format is pinned by existing
  tests, and the suffix achieves the same guarantee with a smaller change.
- Recording a fingerprint for `--no-graph` that omits `graph`. Then the skip logic
  would need to know which runs were partial, which is what the component list does
  without a second rule.

## Consequences

- The first plain run after upgrading rebuilds the graph once, because old
  fingerprints do not carry components.
- `--no-graph` runs are never skipped on a stored graph fingerprint. They rerun the
  structure generators each time, which is cheap.

## References

- Task: `.plaesy/tasks/done/medium_nograph-run-records-fingerprint-without-graph.md`
- Accessed: 2026-10-13
