---
title: "Retrieval Evaluation and Chunking — Gotchas"
description: "Forced index rebuild on Windows (open handles, leftover progress file), config fields never wired (MaxChunkTokens = 0), silent model truncation at 512 tokens, parent-child expansion starving the top-K list, measuring retrieval with small labelled sets"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Retrieval Evaluation and Chunking — Gotchas

Learned while building the labelled docs evaluation and tuning `plaesy search`.
Full trail: [delegation-provider-and-token-savers](../decisions/delegation-provider-and-token-savers.md),
[remove-onnx-and-http-embedders](../decisions/remove-onnx-and-http-embedders.md),
[chunk-cap-model-tokens](../decisions/chunk-cap-model-tokens.md),
[hybrid-matches-before-parent-child](../decisions/hybrid-matches-before-parent-child.md).

## Silent failures that looked like success

- **A pipe hides the exit code.** `plaesy ... 2>&1 | tail -15` reported `exit 0`
  while the command had failed. Read the log, or capture `$?` before the pipe.
  See also [instructions-maintenance](instructions-maintenance.md) (`go test | tail`).
- **`--force` could not delete the index on Windows.** The indexer opened the
  bolt database before `cleanIndex` ran. An open file cannot be deleted there.
  Clean first, then open.
- **`--force` could leave an empty index.** `indexer_progress.json` survived the
  clean, and resume skipped every file it listed. Clean must remove the progress
  file as well.
- **A config cap was never applied.** The pipeline built `chunker.Config` without
  `MaxChunkTokens`, so the value was 0 and every paragraph became its own chunk.
  Check each config field is read where it is written. A field nothing reads is dead.
- **The embedding model truncates without a warning.** E5 takes 512 input tokens
  (`max_position_embeddings` 514 minus 2). `rembed` cuts longer input silently, so a
  4,000-token chunk is embedded from its first 512 tokens only. Size chunks with the
  model's own tokenizer, not a word-count estimate.

## Retrieval behaviour

- **Parent-child expansion starved the top-K list.** Expanding before truncating let
  the sibling chunks of a few documents fill all K slots. Truncate matches first,
  then append siblings.
- **Measure at document level.** Chunk-level lists hide this; the eval adapter
  de-duplicates chunks into files and only then compares to the labels.
- **Small labelled sets are noisy.** With 39 queries, one swapped rank moves MRR by
  about 0.02–0.03. Do not decide a parameter on differences that small. Get about
  100 labels first.
- **BM25 and vector respond differently to chunk size.** A smaller cap helped vector
  search and hurt BM25 on the same corpus. Expect a trade-off, not a single winner.

## Labels

- Write each label by reading the document, then check it mechanically (a keyword
  that must appear in the labelled file). `TestGoldenDocsLabels` does this.
- A query built from a document's own headings favours lexical search. Vary the
  wording, and include paraphrases, before drawing conclusions about vector search.

## One verb, two search systems

`plaesy search` names two different things, and agents were told to use it for both.

- `plaesy search "<query>"` ranks files by their **symbol** embeddings. The index is
  built by `plaesy analyze` in `.plaesy/analysis` and is used for duplicate-avoidance.
- `plaesy search index` / `plaesy search query` search **document text** in
  `.plaesy/search`. Nothing in the normal workflow builds that index, so a project
  that is not this repo has no index and no way to get one from the instructions.

Agents fall back to `grep` for two reasons. `grep` needs no index, no model, and no
extra command. And the instructions put `grep` first, and `plaesy search` only as the
second step, after a literal miss (`.plaesy/instructions/plaesy.md`, Anti-Duplication
Protocol, step 1). Exact symbol names are better served by grep. Meaning-based lookup
is the case `plaesy search` is for.

Docs gap (corrected): `docs/scripts/plaesy-search.md` matches the code for the plain
`plaesy search` command (default `--outdir` `.plaesy/analysis`). The subcommands
`search index` and `search query` (default `.plaesy/search`) have no page of their own.
They appear only in `reference.md` and `plaesy-analyze.md`.
