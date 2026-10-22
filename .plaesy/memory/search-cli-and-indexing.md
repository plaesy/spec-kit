---
title: "Search CLI and Indexing — Lessons"
description: "One search verb, indexes built by analyze, E5 for every index, declaration units with line ranges, comments attached to declarations, and a mistake to avoid: building a command tree before the user approved its shape"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Search CLI and Indexing — Lessons

Decision: [search-single-verb-and-e5-index](../decisions/search-single-verb-and-e5-index.md).

## What the user wants

- One search verb, `plaesy search "<text>"`. The data depends on the project: code,
  second-brain notes, business documents, or a mix. So the verb does not split by data type.
- `plaesy analyze` builds the indexes. Users do not run an index command.
- One model for every index: E5.

## A mistake to avoid

The session built `search symbols|docs ...` from the agent's own proposal and changed
many files before the user approved the shape. The user had already said "search
saja". Before renaming a public verb, state the proposed shape in one line and wait for
a yes, or work only on what the user has already asked for.

## Line ranges

- A search unit is a declaration (or a markdown section), not a file.
- A doc comment belongs to the declaration below it, so the range starts at the doc
  comment. A free-standing comment is not a unit.
- Store 1-based lines with an inclusive end. Convert at output boundaries only:
  LSP uses 0-based lines with an exclusive end; SARIF uses 1-based lines with an
  exclusive `endColumn`.
- Sources checked: [LSP 3.17 specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/),
  [SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html).

## Symbol index quality

With file-level units, a query like "compute cosine similarity" returns the file that
contains the function, plus a long list of other symbols in that file. The list is
hard to scan. Declaration units fix this, so their output format should be designed
together with the units.
