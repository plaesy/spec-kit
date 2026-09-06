# plaesy-graph.sh / plaesy-graph.ps1

**Lightweight knowledge-graph builder for the Plaesy Spec-Kit repo.** No external dependencies (no jq/python on bash; native `ConvertTo-Json` on PowerShell).

Full usage guide: [instructions/plaesy-graph.instructions.md](../../instructions/plaesy-graph.instructions.md).

## Purpose

Builds a graph of nodes (instructions, chatmodes, checklists, templates, scripts, prompts) and edges (references, mentions) for the repo, and lets you query it — which file references what, shortest path between two files, blast-radius of changing a file.

## Output

- `.plaesy/memory/analysis/project.graph.json` — nodes + edges (canonical artifact)
- `.plaesy/memory/analysis/project.html` — self-contained force-directed visualization
- `.plaesy/memory/analysis/reports.md` — plain-language summary
- `.plaesy/memory/analysis/.nodes.tsv`, `.edges.tsv` — internal cache used by query/explain/path/impact commands

## Quick Start

### Bash
```bash
./scripts/bash/plaesy-graph.sh                              # build graph for repo root
./scripts/bash/plaesy-graph.sh --path <dir>                  # build graph for a specific dir
./scripts/bash/plaesy-graph.sh --query "common.ps1"           # keyword query
./scripts/bash/plaesy-graph.sh --path-query "a.md" "b.md"     # shortest path
./scripts/bash/plaesy-graph.sh --explain "go.instructions.md" # explain one node
./scripts/bash/plaesy-graph.sh --impact-check "common.ps1" --impact-depth 1
./scripts/bash/plaesy-graph.sh --semantic-queue                # export INFERRED edges for LLM rewrite
./scripts/bash/plaesy-graph.sh --apply-semantic annotations.json  # merge rationale back
./scripts/bash/plaesy-graph.sh --watch                         # rebuild automatically on source changes
./scripts/bash/plaesy-graph.sh --watch --watch-interval 5       # poll every 5s instead of the 3s default
```

### Windows PowerShell
```powershell
.\scripts\powershell\plaesy-graph.ps1
.\scripts\powershell\plaesy-graph.ps1 -Path <dir>
.\scripts\powershell\plaesy-graph.ps1 -Query "common.ps1"
.\scripts\powershell\plaesy-graph.ps1 -PathQuery "a.md","b.md"
.\scripts\powershell\plaesy-graph.ps1 -Explain "go.instructions.md"
.\scripts\powershell\plaesy-graph.ps1 -ImpactCheck "common.ps1" -ImpactDepth 1
.\scripts\powershell\plaesy-graph.ps1 -SemanticQueue
.\scripts\powershell\plaesy-graph.ps1 -ApplySemantic annotations.json
.\scripts\powershell\plaesy-graph.ps1 -Watch
.\scripts\powershell\plaesy-graph.ps1 -Watch -WatchInterval 5
```

## Options

| Bash | PowerShell | Purpose |
|------|------------|---------|
| `--path <dir>` | `-Path <dir>` | Directory to scan (default: repo root) |
| `--outdir <dir>` | `-OutDir <dir>` | Output directory (default: `.plaesy/memory/analysis`) |
| `--query <text>` | `-Query <text>` | Keyword search across nodes |
| `--path-query <a> <b>` | `-PathQuery <a>,<b>` | Shortest path between two nodes |
| `--explain <node>` | `-Explain <node>` | Show a node's edges and context |
| `--impact-check <node>` | `-ImpactCheck <node>` | Files affected if `<node>` changes |
| `--impact-depth <n>` | `-ImpactDepth <n>` | Traversal depth for impact check (default 2) |
| `--semantic-queue` | `-SemanticQueue` | Export INFERRED-confidence edges for an LLM to annotate |
| `--apply-semantic <file>` | `-ApplySemantic <file>` | Merge LLM-provided rationale back into the graph |
| `--watch` | `-Watch` | Rebuild automatically on source file changes |
| `--watch-interval <n>` | `-WatchInterval <n>` | Poll interval in seconds for `--watch` (default 3) |

Included file types: `.md .ps1 .sh .js .jsx .ts .tsx .py .go`. Excluded: `.git, node_modules, .plaesy, dist, build, __pycache__, vendor, .venv`.
