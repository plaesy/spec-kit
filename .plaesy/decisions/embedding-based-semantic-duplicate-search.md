---
title: "Decision: Embedding-based semantic search for duplicate-avoidance"
description: "chromem-go + rembed added to plaesy graph as --semantic-build/--semantic-query, file-level v1, bumped repo to Go 1.26.1"
updatedAt: "2026-10-03T00:00:00Z"
decided: "2026-10-03"
status: "Accepted"
---

# Decision: Embedding-based semantic search for duplicate-avoidance

**Status**: Accepted — implemented
**Decided**: 2026-10-03
**Deciders**: user + session agent
**Related**: decisions.md (Anti-Duplication Protocol in instructions/plaesy.instructions.md)

## Context and Problem Statement

`instructions/plaesy.instructions.md`'s Anti-Duplication Protocol, step #1,
relied on literal matching only (`find` + `grep -r "similar purpose"`), which
misses two pieces of code that serve the same purpose but use different
wording/naming. This decision was initially recorded as **Deferred** pending
a choice of embedding provider; the user asked to proceed and research
alternatives instead of waiting.

## Options Researched

| Option | CGO? | Maturity | Verdict |
|---|---|---|---|
| **rembed** (pure-Go BERT-family embedder) | No | 11★, new | **Chosen** |
| chromem-go (embeddable vector DB) | No | 1.1k★, beta | **Chosen** |
| fasttextembed (BGE-small, bundled model) | Yes | 35★ | Rejected — CGO breaks plaesy's no-cgo cross-compile story |
| Hugot (ONNX pipelines) | Yes (ONNX Runtime) | Early-stage, general-purpose | Rejected — wrong scope (general LLM pipelines, not embedding-specific) |
| Ollama (local daemon) | N/A | Mature | Rejected — adds an external running process; plaesy analyze/graph is a single offline binary today |
| onnxruntime-go (purego) | No | Unexplored maturity | Not pursued — would need manual ONNX model conversion/bundling, more work than rembed for the same no-cgo property |

**User explicitly decided**: don't fork/vendor rembed preventively despite
its low star count — depend on it normally via `go.mod`, pinned version;
fork only if a real problem surfaces later.

## Decision Outcome

**Chosen option**: `chromem-go` (vector store) + `rembed` (local embedding
model, `sentence-transformers/all-MiniLM-L6-v2`), wired into the existing
new top-level command, `plaesy search` — reuses the symbol extraction
`plaesy graph`/`plaesy analyze` already do, but as its own verb rather than
flags bolted onto `plaesy graph` (`--semantic-build`/`--semantic-query` were
the first cut; moved to a dedicated command after the user asked for
something more familiar/discoverable than nested graph flags — consistent
with common CLI convention, e.g. `gh search`, `rg <pattern>`: a distinct
"find things" verb gets its own top-level command rather than hiding behind
a builder command's flags).

### Scope: symbol names everywhere, doc comments for Go only (v1)

Each graph node's already-extracted symbol names (`symbols.go`'s output,
same data `project.symbols.md` is built from) are joined into one text blob
and embedded as a single document per node. Per-function doc-comment
enrichment was added for `.go` files specifically
(`scripts/internal/semantic/godocs.go`): uses `go/parser` with
`parser.ParseComments` to pull the doc comment directly above each
`FuncDecl`/`TypeSpec`, appended after the symbol name
(`"ValidateUser: checks the given credentials..."`). Go was chosen as the
first language because it's exact (the standard library already parses Go's
doc-comment grammar correctly — no heuristic needed) and plaesy's own
codebase is Go. The other 14 languages `symbols.go` supports (JS/TS, Python,
Rust, Dart, Java, Kotlin, C#, C-family, Swift, Ruby, PHP, Shell, PowerShell)
still get symbol names only — each has its own doc-comment grammar
(JSDoc, docstrings, rustdoc, ...) and adding all of them in one pass was
judged out of scope; a natural per-language follow-up.

### Known consequence: repo-wide Go version bump

`go get github.com/rostamlabs/rembed` forced `scripts/go.mod`'s `go`
directive from `1.20` to `1.26.1` — this is **not** scoped to the new
feature, it raises the minimum Go version for the entire module. User was
asked explicitly and accepted this. CI (`.github/workflows/ci.yml` and
`release.yml`) uses `go-version-file: scripts/go.mod`, so no separate CI
edit was needed — it picks up the new version automatically.

## Implementation

- `scripts/internal/semantic/semantic.go` (new package): `Embedder` interface
  (satisfied by `rembed.Embedder`), `LoadDefaultEmbedder()`, `BuildIndex(ctx,
  *graph.Graph, outDir, Embedder)`, `Query(ctx, outDir, text, topN, Embedder)
  ([]Match, error)`. Persists via `chromem.NewPersistentDB(outDir+"/embeddings",
  compress=true)`.
- `scripts/internal/semantic/semantic_test.go`: tests use a deterministic
  fake `Embedder` (no network/model download) — verifies nodes with no
  symbols are excluded, ranking favors semantically-tied nodes over
  unrelated ones, and querying before `BuildIndex` errors instead of
  panicking.
- `scripts/cmd/plaesy/search.go` (new top-level command): `--index` (embeds +
  persists the index from the current graph), positional `[query text]`
  (prints `similarity\tnodeID\tmatched symbols` lines, highest first; symbols
  come from `project.graph.json`'s per-node data, not a second lookup),
  `--top <n>` (default 10). `plaesy graph` itself is unchanged — the
  semantic flags were removed from it in favor of this dedicated command.
- Docs: `docs/scripts/plaesy-search.md` (new), `docs/scripts/plaesy-graph.md`
  (pointer to it), `docs/scripts/README.md` (index entry),
  `instructions/plaesy.instructions.md` step #1 (semantic match as 1b,
  alongside the existing literal `find`/`grep` as 1a — additive, not a
  replacement).

- `scripts/internal/semantic/godocs.go` (new): `goDocComments(content
  string) (map[string]string, error)` — AST-based, not regex; handles both
  `type Foo struct{}` and grouped `type ( Foo struct{} )` forms. `firstLine()`
  trims a multi-line doc comment to its first line before embedding (keeps
  the vector focused, avoids unrelated detail in a long comment).
- Per rule 2 ("default and record, don't park"), the Anti-Duplication
  Protocol's steps 2-3 were changed from "ask user" to auto-decide +
  record-in-decisions, matching the user's explicit "saya tidak mau manual
  ... saya mau serba otomatis" — pausing is now reserved for ambiguous
  matches or an actual rule 8 hard-stop, not merely "a match exists".

Verified: `go build ./...` and
`go test ./internal/graph/... ./internal/semantic/...` pass, including new
`godocs_test.go` coverage (func docs, single/grouped type decls, no-comment
fallback, invalid-syntax error path).

## Future Work (not done here)

- ~~Doc-comment extraction for the remaining 11 `symbols.go` languages~~
  **Done 2026-10-03.** All 15 languages `symbols.go` covers now have an
  extractor (everything except Markdown, whose "symbols" are headings and
  whose prose is the content itself), across `godocs.go`, `jsdocs.go`,
  `pydocs.go`, `rustdocs.go`, `hashdocs.go` (Ruby/Shell/PowerShell),
  `javadoc4docs.go` (Dart/Java/Kotlin/C#) and `cstyle3docs.go`
  (C-family/Swift/PHP). `docsByExt` in `semantic.go` routes 24 extensions.

- Each extractor keeps private copies of `symbols.go`'s symbol-name regexes.
  A cleaner fix is to export one helper from `internal/graph` and delete the
  copies — the copies are guarded by drift tests, but guarded duplication is
  still duplication.

## Update 2026-10-03 — doc-comment coverage complete, and what it took

Two properties turned out to matter more than the extraction logic itself, and
both fail *silently* — the index still builds, results still look plausible,
and coverage is just quietly worse. Both are now guarded by tests.

**1. A symbol-name regex that drifts from `symbols.go` kills enrichment with no
error.** A doc comment is looked up by the name in `graph.Node.Symbols`, so a
name the extractor recognises but the graph doesn't extract produces a doc
nothing ever reads. The per-language extractors carry private copies of
`symbols.go`'s regexes (they are unexported in another package), so the copies
can drift. `drift_test.go` plus each language's own drift test assert **both**
directions: every documented fixture symbol must get a doc (catches a regex
that matches too little) and every keyed doc name must be a real
`graph.Node.Symbols` entry (catches mis-keying).

Note the direction matters. A subset-only check passes even when a broken regex
makes the extractor produce *fewer* docs, because the survivors are still
validly keyed — which is the same silent-degradation failure wearing a green
test.

**2. An attribute or annotation between the doc comment and the declaration
drops the documentation.** `#[derive(Debug)]`, `@Override`, `[Serializable]`,
`@available`, `#[Route(...)]` are not edge cases — in Java and C# they sit
between the comment and most members. Every extractor now steps over them
(bracket-depth tracked for multi-line forms) and excludes them from the emitted
text so they never reach the embedding. Known limitation: single-line
annotations only.

**A wiring test is not an extractor test.** Every per-language test calls its
extractor directly, so an extractor that is correct, fully tested and never
routed to is indistinguishable from a working one. `docsFor` dispatch is a
separate concern with its own test: `TestDocsFor_WiredExtensions` writes a real
file per extension and asserts the doc text comes back, which is what caught
the `///` marker leaking a stray `/` in Dart/Kotlin/C#.

## Update 2026-10-03 — staleness policy resolved

The "staleness/rebuild policy" item above is now **done**. Reusing
`graph.SourceFingerprint` rather than inventing a second mechanism was kept, and
the open question from implementation — auto-rebuild vs. warn — is resolved as
**warn and refuse** (`plaesy search` exits non-zero, pointing at `--index`).

Rationale: the whole reason the feature exists is that a stale index can make
the Anti-Duplication Protocol return a *false* "no duplicate found". Silently
paying model-load plus embedding cost on a read-only query was judged the lesser
evil against returning a confident wrong answer. Note the policy is
fail-closed, so an index built before this change refuses to query once, with a
distinct message, until rebuilt.

Known limitation, accepted rather than fixed: `SourceFingerprint` is
`fileCount|maxMtime`, so it detects an edit only because writing updates mtime.
A content change that preserves mtime is invisible. Content hashing was
declined to avoid a second staleness mechanism diverging from the graph's.

Also corrected here: this record previously described `SourceFingerprint` as
including "framework version". It does not.

## References

- `instructions/plaesy.instructions.md` lines 84-120 (Anti-Duplication Protocol) — edited 2026-10-03
- `scripts/internal/graph/symbols.go` (symbol extraction reused as embedding input; confirmed no doc-comment capture) — read 2026-10-03
- [github.com/rostamlabs/rembed](https://github.com/rostamlabs/rembed) — read 2026-10-03
- [github.com/philippgille/chromem-go](https://github.com/philippgille/chromem-go) — read 2026-10-03
- [github.com/cemsina/fasttextembed](https://github.com/cemsina/fasttextembed) — read 2026-10-03 (rejected, CGO)
- [knightsanalytics.com — Hugot: Running LLMs in Go](https://www.knightsanalytics.com/post/hugot-llms-in-go) — read 2026-10-03 (rejected, wrong scope)
