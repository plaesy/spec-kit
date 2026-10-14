---
title: "Document index built by plaesy analyze, searched by plaesy search"
phase: implement
status: done
createdAt: "2026-10-13T00:00:00.000Z"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Task: Document index in analyze

## Description

The document index (chunking, E5 embedding, chromem and Bleve stores) exists in
`scripts/internal/search`, and `cmd/plaesy/search_index.go` still has its code, but
no command runs it now. The user wants `plaesy analyze` to build it and
`plaesy search "<text>"` to return document passages alongside code symbols.

## Scope

- `analyze` builds the document index from the project's documents (markdown, HTML,
  JSON). Exclude `.plaesy/`, `.git/`, dependency folders, and anything the project
  ignores.
- The index is fingerprint-gated, the same way the symbol index is.
- `plaesy search` returns one ranked list. Each result says whether it is a symbol or
  a document, from the file extension it came from. Decided: one ranked list, each
  row labelled `symbol` or `doc`, merged with Reciprocal Rank Fusion because raw cosine
  scores from the two indexes are not on the same scale. `--kind symbol|doc` filters.
  Row format: `score  kind  ext  path:start-end  name`.
- Remove the TODO in `search_index.go` once the commands are wired or the code is removed.

## Acceptance Criteria

- [x] `plaesy analyze` builds both indexes; a query on a known document phrase returns it.
- [x] A document edited after analyze makes `plaesy search` refuse the stale index.
- [x] No dead code is left in `search_index.go`.

## Status

Done: 1, 2, 3, 4, 6 of the six steps, and all three acceptance criteria verified.

- Step 1, wiring: `analyze` builds the document index into `.plaesy/analysis/docs`
  (`internal/search/docindex`). It runs on both the fresh path and the cached path,
  and `--no-index` skips it. Verified on a scratch project.
- Step 2, exclusions: `parser.SkipDir` skips `.plaesy`, `.git`, `node_modules`,
  `vendor`, `dist`, `build`, `target`, `__pycache__`, `.venv`.
- Step 3, fingerprint: `docindex.Fingerprint` hashes each document's path and content,
  so edits, additions and deletions all change it. `search` refuses a stale or missing index.
- Step 4, fusion: `search` merges symbol and document results with Reciprocal Rank
  Fusion (k=60) and adds `--kind symbol|doc`.
- Step 6, dead code: `cmd/plaesy/search_index.go` and its test are removed. `cleanIndex`
  is replaced by `docindex.Build`, which deletes the whole index before a rebuild.

Open, not done:
- Step 5, line mapping (resolved for Markdown). The chunker now tracks each chunk's
  source lines (`Chunk.StartLine`/`EndLine`, 1-based, end-inclusive), and the
  Markdown parser keeps the source text in `Document.Content`, so the lines are
  the file's own lines. The indexer stores them as `start_line`/`end_line` metadata,
  and doc rows print `path:start-end`. Line-based offsets replace byte offsets:
  `StartOffset`/`EndOffset` are still the old placeholders and are unused.
  Not done: HTML, JSON, and plain text rows print `path` only. Their content is
  extracted text, so its lines are not lines of the file. Mapping them needs the
  parsers to keep source positions.
- `.gitignore` rules are not applied when collecting documents. Only the named
  directories are excluded.
- Document search uses the hybrid retriever (vector and BM25). Its rank order is what
  fusion uses. No golden-set evaluation of document results has been run.
