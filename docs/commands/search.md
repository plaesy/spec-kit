# `search`

Search the project by meaning. The indexes it reads are built by `plaesy analyze`;
a query refuses an index that is missing or older than the sources.

```bash
plaesy search "validates user credentials"            # top 10
plaesy search "formats a date for display" --top 5
```

| Flag | Default | Purpose |
|------|---------|---------|
| `--path` | `.` | Project directory |
| `--outdir` | `.plaesy/analysis` | Analysis directory, relative to `--path` |
| `--top` | `10` | Maximum results |
| `--kind` | both | `symbol` for code declarations and Markdown headings only, `doc` for document passages only |
| `--ext` | all | Keep only these extensions, comma-separated, e.g. `go,ts` or `md,html`. Matches any listed extension and applies to symbol and doc results |

Symbols and document passages are ranked in one list. Output is one result per
line, best first. The columns are `score`, `kind`, `extension`, `location`, and `name`:

```text
0.0164	symbol	go	internal/auth/check.go:12-19	CheckUserIsValid
0.0164	doc	md	guides/refunds.md	Refund policy
```

`kind` is `symbol` or `doc`, decided by the file's extension: Markdown, HTML and JSON
are `doc`, everything else is `symbol`. The `score` is the Reciprocal Rank Fusion score,
`1/(60 + rank)` within each list, because the two indexes' similarity scores are not on
the same scale. A result's position in its own list is what counts, not its raw score.

A **Go declaration** (function, method, or type) is one symbol unit, so `path:start-end`
points at the matching declaration. Other symbol files show `path` alone, with their symbol
names in the last column. A **document** result shows `path` alone for now: the chunker's
character offsets are placeholders, so a document's lines cannot be computed yet.

The document index covers `.md`, `.markdown`, `.mdx`, `.html`, `.htm` and `.json` files. It
skips the same directories as every other scan (see [What analyze leaves out](../graph-extensions.md#what-analyze-leaves-out)),
and the paths matched by the project's root `.gitignore`.

A query refuses the document index when it is missing, or when a document has changed since
`plaesy analyze` built it.

**Line ranges** are 1-based, and the end line is inclusive. A declaration's range
starts at its doc comment when one sits directly above it, with no blank line in
between. A free-standing comment is not a unit. Convert at output boundaries only:
LSP uses 0-based lines with an exclusive end, and SARIF uses 1-based lines with an
exclusive `endColumn`.

See [CLI Reference](../reference.md).
