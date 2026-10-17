---
title: "Symbol index units: declarations with start and end lines"
phase: implement
status: backlog
createdAt: "2026-10-13T00:00:00.000Z"
updatedAt: "2026-10-13T00:00:00.000Z"
---

# Task: Declaration units with line ranges

## Description

The symbol index is file-level: one document per file, with symbol names joined into
it. Search results therefore point at files, not at the function that matches. Each
unit should be one declaration, with its file, its start and end line, and its kind.

Rules agreed with the user:
- A unit is a declaration (function, method, type, class) or a markdown section.
- A doc comment is part of its declaration: the range starts at the first line of the
  doc comment.
- A free-standing comment (license header, TODO, comment block not attached to a
  declaration) is not a unit.
- Lines are 1-based, end inclusive. Conversion to LSP or SARIF happens only at output
  boundaries.

## Scope

- Go first, via `go/ast`: exact start and end of each declaration.
- Other languages: a tree-sitter grammar (pure Go, `gotreesitter`), only after the
  Go version is verified. A regex that guesses block ends is not acceptable.
- Output: `score`, `path:start-end`, `extension`, `kind`, `name`.

## Acceptance Criteria

- [x] Go declarations get exact start and end lines; tests use fixtures with known ranges.
- [x] A doc comment directly above a declaration is included in its range.
- [x] A free-standing comment produces no unit.
- [x] The E5 symbol index is rebuilt with declaration units, and queries return a `path:start-end` result.
- [x] Line ranges are 1-based and end-inclusive, and this is documented in `reference.md`.
