---
title: Extend semantic search doc-comment extraction to JS/TS and Python
phase: implement
status: done
createdAt: "2026-10-03T12:36:27.000Z"
updatedAt: "2026-10-03T13:42:23.000Z"
---

## Description

`plaesy search` (scripts/internal/semantic/) embeds each node's extracted
symbol names for semantic matching, and for `.go` files additionally
enriches each symbol with its doc comment via `go/parser`
(`scripts/internal/semantic/godocs.go`). JS/TS and Python are the next
highest-value languages to add the same enrichment to — `symbols.go`
already extracts their function/class names
(`jsSymbols`/`pySymbols`), so only the doc-comment lookup is missing.

Add two sibling extractors analogous to `godocs.go`:

- JS/TS: JSDoc blocks (`/** ... */`) immediately above a
  `function`/`class` declaration matched by `symJSFuncRE`/`symJSClassRE`.
- Python: docstrings (`"""..."""` or `'''...'''`) as the first statement
  inside a `def`/`class` body matched by `symPyDefRE`/`symPyClassRE`
  (Python's docstring convention is *inside* the block, unlike Go/JS's
  *above* the declaration — the extractor needs to look at the line(s)
  after the matched declaration line, not before).

Wire both into `nodeText()` in `semantic.go` the same way `goDocsFor` is
wired today, keyed off `filepath.Ext(n.ID)`.

## Decision

- Regex-based extraction is acceptable here (unlike Go, which has
  `go/parser` in the standard library) — there is no equivalent standard
  JS/Python AST parser in the Go standard library, and pulling in a
  third-party parser just for doc-comment text is disproportionate to the
  value. Mirror the existing heuristic, not-AST approach `symbols.go`
  itself already uses for every non-Go language.
- Keep the "first line only" truncation (`firstLine()`) for consistency
  with the Go extractor.

## Acceptance Criteria

- [x] `scripts/internal/semantic/jsdocs.go`: `jsDocComments(content string)
      map[string]string` extracts a JSDoc block immediately above a
      matched function/class declaration, keyed by symbol name
- [x] `scripts/internal/semantic/pydocs.go`: `pyDocComments(content string)
      map[string]string` extracts a docstring as the first statement of a
      matched `def`/`class` body, keyed by symbol name
- [x] `nodeText()` in `semantic.go` calls both for `.js`/`.jsx`/`.ts`/`.tsx`
      and `.py` files respectively, same pattern as `goDocsFor`
- [x] Unit tests for both extractors covering: a documented function, an
      undocumented function (no crash, no entry), a documented class, and
      (Python only) both `"""` and `'''` docstring delimiters
- [x] `go build ./...` and `go test ./internal/semantic/...` pass
- [x] `docs/scripts/plaesy-search.md` updated to say JS/TS/Python doc
      comments are now included (not just Go) — merged centrally after both
      concurrent tasks finished, per the ownership split

## Reference

- [embedding-based-semantic-duplicate-search](../../decisions/embedding-based-semantic-duplicate-search.md) — origin decision, "Future Work" section names this exact gap
- scripts/internal/semantic/godocs.go — the Go implementation to mirror
- scripts/internal/graph/symbols.go — existing per-language symbol regexes to reuse (`symJSFuncRE`, `symJSClassRE`, `symPyDefRE`, `symPyClassRE`)

## Dependencies

- None — additive to existing `internal/semantic` package.

## Notes

- Do not touch `scripts/internal/graph/symbols.go` itself; it is
  deliberately scoped to symbol *names* only. Doc-comment extraction lives
  entirely in `internal/semantic`, parallel to it.
- After this, 11 of 15 `symbols.go` languages still lack doc-comment
  enrichment (Rust, Dart, Java, Kotlin, C#, C-family, Swift, Ruby, PHP,
  Shell, PowerShell, Markdown) — a further backlog item, not in this task's
  scope.
