# plaesy search

**Semantic (meaning-based) search over a repo's extracted symbols.**

Source: `scripts/cmd/plaesy/search.go` + `scripts/internal/semantic/`.

## Purpose

Finds files whose extracted symbols (function/class/heading names — the same
data `plaesy analyze` already produces in
`project.symbols.md`/`project.graph.json`) are similar *in meaning* to a
query, not just by literal keyword. This closes the gap literal
`grep -r "similar purpose"` can't: two functions can serve the same purpose
under completely different names (`ValidateUser` vs. `CheckUserIsValid`) and
grep will miss the match. See the Anti-Duplication Protocol in
`.plaesy/instructions/plaesy.md` — this is step 1b there, alongside the
existing literal `find`/`grep` step 1a.

Each node's symbol *names* are embedded, plus their doc comments. Every
language `plaesy analyze` extracts symbols for now has an extractor, except
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
per-language regexes `plaesy analyze` already uses for symbol names; it is
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
- Requires `plaesy analyze` to have built the indexes first.

## Building the index

`plaesy analyze` builds both indexes (symbol and document) as part of its run, so
the normal workflow needs no separate step:

```bash
plaesy analyze                                   # graph + symbols + both indexes
plaesy search "validates a user's credentials"
```

Rerun `plaesy analyze` to rebuild after source changes. `plaesy analyze --no-index`
skips the index build; `plaesy search` then refuses to answer until an analyze run
builds the indexes. `plaesy analyze --watch` keeps them current during a long session.

## Quick Start

```bash
plaesy analyze                                                  # builds the indexes too
plaesy search "validates a user's credentials"                  # find files similar in meaning
plaesy search "formats a date for display" --top 5              # cap results
plaesy search "refund policy" --kind doc                        # document passages only
plaesy search "refund policy" --ext md,html                     # keep only these extensions
```

Output is tab-separated, highest score first: `score`, `kind`, `extension`,
`location`, `name`. `kind` is `symbol` (code declarations and Markdown headings) or
`doc` (passages from Markdown, HTML and JSON). For a Go function, method, or type the
location is `path:start-end` covering the declaration, doc comment included (1-based,
end inclusive); other symbol files show the path alone. Document results show the
path alone for now. See the `search` section of `docs/reference.md`.

```text
0.0164	symbol	go	internal/auth/check.go:12-19	CheckUserIsValid
0.0164	doc	md	guides/refunds.md	Refund policy
```

The score is Reciprocal Rank Fusion, `1/(60 + rank)` within each list, so the two
indexes' similarity scores never have to be compared directly.

## Options

| Flag | Default | Purpose |
|------|---------|---------|
| `--path <dir>` | `.` | Project directory |
| `--outdir <dir>` | `.plaesy/analysis` | Analysis directory, relative to `--path`; written by `plaesy analyze` |
| `--top <n>` | `10` | Maximum results |
| `--kind <symbol\|doc>` | both | Only symbol results (code declarations and Markdown headings) or only doc results |
| `--ext <list>` | all | Keep only these extensions, comma-separated, e.g. `go,ts` or `md,html`. Applies to both kinds |
| `--filter <spec>` | none | Narrow doc results, e.g. `source:markdown,path:docs,file_type:md,date:>2024-01-01`. Doc results only |
| `--parent-child` | `false` | Return whole sections with the matched passages as children. Doc results only |

Run `plaesy search --help` for the exact current flag set (source of truth:
`scripts/cmd/plaesy/search.go`).

## Staleness detection

`plaesy search` detects when an index is stale, meaning source files changed since
`plaesy analyze` built it, and refuses to answer rather than rebuilding. A
read-only query should not silently start paying model-load and embedding cost.
The symbol
index uses the same `graph.SourceFingerprint` mechanism as `plaesy analyze`'s
fingerprint fast path, stored at `<outdir>/embeddings/.fingerprint`. The document
index keeps its own fingerprint at `<outdir>/docs/fingerprint`.

Behavior:

- **Fresh index** — the query runs immediately.
- **Stale index** — the query exits non-zero and asks for `plaesy analyze`.
- **Missing index** — the query exits non-zero and asks for `plaesy analyze`.
  A symbol index with no fingerprint is treated the same way, since its freshness
  cannot be proven.

```text
index is stale (source files changed since the last analyze). Run 'plaesy analyze' to refresh it
```

**Limitation:** the fingerprint is *mtime*-based (file count + newest modification
time), not content-hashed. It detects edits because writing a file updates its mtime,
so a content change that preserves mtime (`touch -d`, some editor atomic-write
patterns) is not detected. This is the accepted cost of reusing `SourceFingerprint`
rather than adding a second mechanism.

## Troubleshooting

```text
# "no index found in <outdir>. Run 'plaesy analyze' first"
# "semantic: no embedding index at ... — run plaesy analyze first"
plaesy analyze

# "index is stale (source files changed since the last analyze)"
# "document index is missing or stale (a document changed since the last analyze)"
plaesy analyze

# "--kind must be symbol or doc, got ..."
plaesy search "<query>" --kind symbol   # or --kind doc, or omit the flag
```
