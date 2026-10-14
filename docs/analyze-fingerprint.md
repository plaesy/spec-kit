# Analysis fingerprint

`plaesy analyze` writes `.plaesy/analysis/.analysis-fingerprint` after a run that
built every artifact it covers. A later run with the same fingerprint skips
regeneration and reports the artifacts as cached.

## Format

```text
<count>|<maxMtime>|<version>|components=structure,graph,symbols
```

- `count` and `maxMtime` cover the source and config files that the analysis
  counts (see `fingerprintExts` in `scripts/internal/analyze/fingerprint.go`).
- `version` is the framework version.
- `components` lists what the recording run built.

## When a run records

- A plain run or `--force` run that built the graph and symbols records
  `structure,graph,symbols`.
- `--no-graph` records nothing. It writes no graph, so a fingerprint from it would
  claim artifacts that do not exist.
- A run whose graph build failed records nothing.

## When a run skips

A run skips only if the walk fingerprint matches and the recorded components cover
what the run needs. `structure` is always needed. `graph` and `symbols` are needed
unless `--no-graph` is set.

A fingerprint from before components were tracked has no `components=` suffix. It
never skips, so the first run after upgrading rebuilds and writes the new form.

Decision record: `.plaesy/decisions/fingerprint-component-tracking.md`.
