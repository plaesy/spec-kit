# plaesy graph

**Lightweight knowledge-graph builder for the Plaesy Spec-Kit repo.** Go standard library only — no external dependencies.

Full usage guide: [instructions/plaesy-graph.instructions.md](../../instructions/plaesy-graph.instructions.md).

## Purpose

Builds a graph of nodes (instructions, chatmodes, checklists, templates, scripts, prompts) and edges (references, mentions) for the repo, and lets you query it — which file references what, shortest path between two files, blast-radius of changing a file.

## Output

- `.plaesy/analysis/project.graph.json` — nodes + edges (canonical artifact)
- `.plaesy/analysis/project.html` — self-contained force-directed visualization
- `.plaesy/analysis/reports.md` — plain-language summary
- `.plaesy/analysis/.fingerprint` — internal cache used to skip rebuilds when nothing changed

## Quick Start

```bash
plaesy graph                                              # build graph for repo root
plaesy graph --path <dir>                                 # build graph for a specific dir
plaesy graph --query "common.md"                          # keyword query
plaesy graph --fuzzy-query "common"                        # fuzzy/substring search
plaesy graph --path-query-from "a.md" --path-query-to "b.md"  # shortest path
plaesy graph --explain "go.instructions.md"                # explain one node
plaesy graph --impact-check "common.md" --impact-depth 1
plaesy graph --semantic-queue                              # export INFERRED edges for LLM rewrite
plaesy graph --apply-semantic annotations.json              # merge rationale back
plaesy graph --watch                                       # rebuild automatically on source changes
plaesy graph --watch --watch-interval 5                     # poll every 5s instead of the 3s default
```

## Options

| Flag | Purpose |
|------|---------|
| `--path <dir>` | Directory to scan (default: `.`) |
| `--outdir <dir>` | Output directory, relative to `--path` (default: `.plaesy/analysis`) |
| `--if-changed` | Rebuild only if source files changed since the last build |
| `--query <text>` | Keyword search across node ids |
| `--fuzzy-query <text>` | Fuzzy/substring search against node ids |
| `--search-depth <n>` | Search depth for `--fuzzy-query` (display only, default 2) |
| `--path-query-from <a>` / `--path-query-to <b>` | Shortest path between two nodes |
| `--explain <node>` | Show a node's edges and context |
| `--impact-check <node>` | Files affected if `<node>` changes |
| `--impact-depth <n>` | Traversal depth for `--impact-check` (default 2) |
| `--impact-visualize` | Write `impact-visualization.html` instead of printing |
| `--semantic-queue` | Export INFERRED-confidence edges for an LLM to annotate |
| `--apply-semantic <file>` | Merge a `{source,target,rationale}[]` annotations file back into `reports.md` |
| `--business-logic` | Export INFERRED edges to `business-logic-queue.json` |
| `--generate-paths` | Write a `learning-paths.json` stub |
| `--watch` | Rebuild automatically on source file changes |
| `--watch-interval <n>` | Poll interval in seconds for `--watch` (default 3) |

Run `plaesy graph --help` for the exact current flag set (source of truth: `scripts/cmd/plaesy/graph.go`).
