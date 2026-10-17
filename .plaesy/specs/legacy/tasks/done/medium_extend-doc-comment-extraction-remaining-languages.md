---
title: Extend semantic search doc-comment extraction to the remaining 11 languages
phase: implement
status: done
createdAt: "2026-10-03T12:36:27.000Z"
updatedAt: "2026-10-03T12:36:27.000Z"
---

## Description

After Go (done) and JS/TS + Python (see
`high_extend-doc-comment-extraction-js-ts-python.md`), 11 of the 15
languages `scripts/internal/graph/symbols.go` extracts symbols for still
embed bare symbol names only, with no doc-comment enrichment in
`plaesy search`: Rust, Dart, Java, Kotlin, C#, C-family (C/C++/Obj-C),
Swift, Ruby, PHP, Shell, PowerShell.

Lower priority than the JS/TS/Python task: these languages are a smaller
share of typical plaesy-managed projects, and the marginal semantic-search
quality gain per language is smaller than the first two. Scope this task
once the JS/TS/Python extractors land and their test pattern can be
copied, rather than designing from scratch.

Doc-comment conventions to handle:

- Rust: `///` or `//!` lines immediately above the item (rustdoc)
- Dart/Java/Kotlin/C#/C-family/Swift: `/** ... */` or `///`-style blocks
  above the declaration (same shape as JSDoc — can likely share logic with
  the JS/TS extractor from the prerequisite task, parameterized by comment
  delimiter)
- Ruby: `#` comment lines immediately above the `def`/`class`
- PHP: PHPDoc `/** ... */` above the declaration (same shape as JSDoc)
- Shell/PowerShell: `#`/`<# ... #>` comment block above the function

## Acceptance Criteria

- [x] One extractor per language. Resolved as four new files grouped by comment
      convention rather than one per language: `rustdocs.go`,
      `hashdocs.go` (Ruby + Shell + PowerShell, sharing one `#`-comment scan),
      `javadoc4docs.go` (Dart, Java, Kotlin, C#), `cstyle3docs.go`
      (C-family, Swift, PHP — reusing `jsCommentBlock`/`jsDocText`)
- [x] `nodeText()` wires each new extractor in by file extension — via the
      `docsByExt` table in `docsFor()`, 24 extensions, replacing a switch
      plus eight near-identical wrappers
- [x] Unit tests per language: a documented symbol, an undocumented one, plus
      a drift test that builds a real `graph.Build` and asserts names match
- [x] `go build ./...` and `go test ./internal/semantic/...` pass
- [x] `docs/scripts/plaesy-search.md` and
      `.plaesy/decisions/embedding-based-semantic-duplicate-search.md`
      updated to reflect full language coverage

## Notes

- **Splitting further per-language was the right call** — see Acceptance
  Criteria for how it was resolved.

- **Attribute/annotation lines are the defect class that mattered.** A line
  between a doc comment and its declaration (`#[derive(...)]`, `@Override`,
  `[Serializable]`, `@available`, `#[Route(...)]`) terminates the contiguity
  scan, so the documentation is silently dropped. Not an edge case — in Java
  and C# it affects most members. Every extractor now steps over them, and
  every drift fixture contains an annotated declaration so the gap stays
  visible. Known limitation: single-line annotations only.

- **`///` leaked a stray `/`.** `jsDocText` stripped `//` but not `///`, so
  Dart/Kotlin/C# embedded `/ Loads the user.` Found by
  `TestDocsFor_WiredExtensions`, which asserts exact doc text per extension.
  Fixed in `jsDocText` so all languages benefit.

- **A wiring test is not an extractor test.** Every per-language test calls its
  extractor directly, so a correct, fully-tested extractor that was never wired
  looks identical to a working one from inside its own file. Only
  `TestDocsFor_WiredExtensions` (one row per extension, real file on disk)
  proves the extension is actually routed.

- **Follow-up raised, not fixed here:** `internal/graph` keeps three
  out-of-sync extension lists and never collects `.hh`/`.cxx`, so those two
  `docsByExt` entries cannot fire through the real pipeline. Tracked as
  `.plaesy/tasks/backlog/medium_graph-include-ext-missing-hh-cxx.md`; the fix
  belongs in `internal/graph`, which is outside this task's scope.

## Reference

- [embedding-based-semantic-duplicate-search](../../decisions/embedding-based-semantic-duplicate-search.md)
- [high_extend-doc-comment-extraction-js-ts-python](./high_extend-doc-comment-extraction-js-ts-python.md) — do this first; its extractor shapes are the template

## Dependencies

- Soft dependency: `high_extend-doc-comment-extraction-js-ts-python.md` —
  not a hard blocker, but doing it first establishes the test/extractor
  pattern this task should reuse rather than re-deriving independently.
- **Scheduling (2026-10-03): wave 2, not wave 1.** The `/continue` fan-out
  runs this alongside the fingerprint task, but must run *after* the
  JS/TS+Python task. Both edit `internal/semantic/semantic.go`'s
  `nodeText()`/`goDocsFor()` region — the single extension-dispatch point —
  so concurrent agents in one worktree are last-write-wins, not a merge.
  This task additionally *consumes* the extractor's file/test shape that
  task establishes, so it cannot even be scoped correctly before task 1
  lands.

## Notes

- Consider whether this is worth splitting further per-language once
  started — 11 languages in one task risks becoming an unreviewable diff.
