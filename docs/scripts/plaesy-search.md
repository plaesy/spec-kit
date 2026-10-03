# plaesy search

**Semantic (meaning-based) search over a repo's extracted symbols.**

Source: `scripts/cmd/plaesy/search.go` + `scripts/internal/semantic/`.

## Purpose

Finds files whose extracted symbols (function/class/heading names — the same
data `plaesy graph`/`plaesy analyze` already produce in
`project.symbols.md`/`project.graph.json`) are similar *in meaning* to a
query, not just by literal keyword. This closes the gap literal
`grep -r "similar purpose"` can't: two functions can serve the same purpose
under completely different names (`ValidateUser` vs. `CheckUserIsValid`) and
grep will miss the match. See the Anti-Duplication Protocol in
`instructions/plaesy.instructions.md` — this is step 1b there, alongside the
existing literal `find`/`grep` step 1a.

Each node's symbol *names* are embedded, plus their doc comments. Every
language `plaesy graph` extracts symbols for now has an extractor, except
Markdown (whose "symbols" are headings and whose prose is the content itself):

| Language | Extensions | Doc-comment source |
|----------|-----------|--------------------|
| Go | `.go` | `go/parser` — not a regex, because Go's doc-comment grammar is precise enough that the standard library already handles it correctly (`godocs.go`) |
| JS/TS | `.js` `.jsx` `.ts` `.tsx` | JSDoc `/** ... */`, or a run of `//` lines, directly above a `function`/`class` (`jsdocs.go`) |
| Python | `.py` | the docstring as the first statement of a `def`/`class` body — Python's convention is *inside* the block, unlike Go's and JS's *above* it (`pydocs.go`) |
| Rust | `.rs` | `///` outer line-docs or `/** */` outer block-docs. `//!` inner docs are deliberately skipped: they document the enclosing crate, which has no name in `graph.Node.Symbols` to key against (`rustdocs.go`) |
| Ruby | `.rb` | `#` run above the declaration, or a `=begin`/`=end` block (`hashdocs.go`) |
| Shell | `.sh` | `#` run above the declaration; `#!` shebangs are never treated as docs (`hashdocs.go`) |
| PowerShell | `.ps1` | `#` run, or a `<# ... #>` block (`hashdocs.go`) |
| Dart | `.dart` | `///` or `/** */` (`javadoc4docs.go`) |
| Java | `.java` | `/** */` Javadoc (`javadoc4docs.go`) |
| Kotlin | `.kt` `.kts` | KDoc `/** */` (`javadoc4docs.go`) |
| C# | `.cs` | `///` XML doc comments (`javadoc4docs.go`) |
| C-family | `.c` `.h` `.cpp` `.hpp` `.cc` `.hh` `.cxx` | any contiguous `/* */` or `//` block above the declaration — C has no doc-comment standard, so this is permissive by necessity (`cstyle3docs.go`) |
| Swift | `.swift` | `///` or `/** */` (`cstyle3docs.go`) |
| PHP | `.php` | PHPDoc `/** */` (`cstyle3docs.go`) |

Only the **first line** of each doc comment is embedded — enough semantic
signal without dragging a whole paragraph (and its unrelated detail) into the
vector. Extraction is line-based and heuristic outside Go, mirroring the
per-language regexes `plaesy graph` already uses for symbol names; it is
best-effort, so a file that can't be read or parsed simply contributes bare
symbol names. Annotations and attributes sitting between a comment and its
declaration (`#[derive(...)]`, `@Override`, `[Serializable]`, `@available`)
are stepped over rather than treated as code, since otherwise the
documentation of most annotated members would be silently dropped.

See `.plaesy/decisions/embedding-based-semantic-duplicate-search.md` for the
full rationale.

Each non-Go extractor keeps private copies of `graph/symbols.go`'s symbol-name
regexes, because a doc comment is looked up by the name in
`graph.Node.Symbols` and a mismatch would silently disable enrichment.
`drift_test.go` and the per-language test files fail if the copies and
`symbols.go` disagree.

## How it works

- Embedding model: `sentence-transformers/all-MiniLM-L6-v2`, loaded locally
  via [rembed](https://github.com/rostamlabs/rembed) (pure Go, no cgo).
  Downloaded once from the Hugging Face Hub on first use and cached
  thereafter — no network access needed on later runs.
- Vector store: [chromem-go](https://github.com/philippgille/chromem-go),
  persisted under `<outdir>/embeddings/` (gob files, gzip-compressed).
- Requires a graph already built (`plaesy graph` or `plaesy analyze`) — it
  reads `project.graph.json`'s nodes.

## Quick Start

```bash

plaesy search --index                                       # (re)build the semantic index from the current graph
plaesy search "validates a user's credentials"              # find files similar in meaning, not just name
plaesy search "formats a date for display" --top 5           # cap results
```

Output is `similarity<TAB>file<TAB>matched symbols` per line, highest
similarity first — the matched symbols are included because
`project.graph.json` already carries them per node, so a result tells you
*which* function/class matched, not just which file.

```text

0.8214  internal/auth/check.go   CheckUserIsValid
0.7958  internal/auth/validate.go   ValidateUser, hashPassword
```

## Options

| Flag | Purpose |
|------|---------|
| `--path <dir>` | Project directory to scan (default: `.`) |
| `--outdir <dir>` | Output directory, relative to `--path` (default: `.plaesy/analysis`; shared with `plaesy graph`) |
| `--index` | (Re)build the semantic index from the current graph, and record the source fingerprint used for staleness detection. Run this first, and again whenever symbols change meaningfully. |
| `--top <n>` | Max results (default 10) |

Run `plaesy search --help` for the exact current flag set (source of truth:
`scripts/cmd/plaesy/search.go`).

## Staleness detection

`plaesy search` detects when the index is stale — i.e. source files changed
since the last `--index` run — by reusing the same `graph.SourceFingerprint`
mechanism as `plaesy graph --if-changed` and `plaesy analyze`. One staleness
scheme for the repo, not two that can drift apart. The fingerprint is stored at
`<outdir>/embeddings/.fingerprint` when `--index` completes.

Behavior:

- **Fresh index** — the query runs immediately, unchanged.
- **Stale index** (source files changed) — `plaesy search` exits non-zero
  rather than auto-rebuilding. A read-only query should not silently start
  paying model-load and embedding cost.
- **No fingerprint recorded** — exits non-zero with its own message. This is
  a different condition from "changed": it means the index was built before
  staleness detection existed, so its freshness cannot be proven. Rebuilding
  once enables detection from then on.

```text
semantic index is stale (source files changed since last --index). Run 'plaesy search --index' to rebuild
```

**Limitation:** the fingerprint is *mtime*-based (file count + newest
modification time), not content-hashed. It detects edits because writing a file
updates its mtime, so a content change that preserves mtime (`touch -d`, some
editor atomic-write patterns) is not detected. This is the accepted cost of
reusing `SourceFingerprint` rather than adding a second mechanism.

## Troubleshooting

```bash

# "semantic index is stale (source files changed since last --index)"
plaesy search --index

# "no source fingerprint recorded for the semantic index"
plaesy search --index   # rebuild once to enable staleness detection

# "no embedding index at ... — run with --index first"
plaesy search --index

# "no graph found in ... Run without flags first to build it"
plaesy graph   # or: plaesy analyze
```
