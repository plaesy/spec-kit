---
description: "Context engineering: token/context compression (plaesy trim) + knowledge-graph analysis (plaesy graph). Load when compressing output/memory files, wrapping noisy commands, analyzing repo structure/relationships, or doing a pre-refactor impact check."
applyTo: "**/*"
---

# Context Engineering — `plaesy trim` + `plaesy graph`

⚡ **Both are subcommands of the single cross-platform `plaesy` Go binary**
(`plaesy trim ...`, `plaesy graph ...`) — same behavior on Windows/macOS/Linux.
There is **no** standalone `plaesy-trim`/`plaesy-graph` script and no separate
PowerShell/Bash split (the old `scripts/bash/plaesy-trim.sh`,
`plaesy-trim.ps1`, `plaesy-graph.sh`, `plaesy-graph.awk`, `plaesy-graph.ps1`
ports were removed during the Go-CLI migration). This file is the single merged
guide.

The two commands solve the two places tokens actually accrue:

- **`trim`** = shrink what *already exists* (command output, memory/instruction files).
- **`graph`** = understand *what matters* (so you don't load/re-read the wrong files).

---

## 1. `plaesy graph` — Repo Knowledge Graph

### Purpose

Turns your project — or any folder passed via `--path` — into a lightweight
knowledge graph. Nodes are `.md`/`.go`/`.js`/`.ts`/`.py`/`.dart` files; edges are
markdown links, source/import/call references, resolved imports, install-mirror
pairs, and plain-text path mentions. **No external dependency** (Python, jq, LLM,
network) is required.

### When to use

- User asks "how is X connected to Y", "what references this file/instruction", "which files are orphaned/unused".
- User wants an overview/visualization of agents, instructions, checklists, templates, code relationships.
- Before large refactors of `.plaesy/memory/`, `.plaesy/instructions/`, `.plaesy/roles/`, or `scripts/`, to check the blast radius.

### Core Graph Operations

```bash

plaesy graph                                                     # build (or rebuild) the whole repo
plaesy graph --path instructions                                 # build for a specific subfolder
plaesy graph --watch                                              # rebuild automatically on source changes
```

#### Feature #1: Architecture Layer Detection (auto-included in graph)

Nodes include a `layer` attribute: API, Service, Data, UI, Utility.

- Folder naming patterns: `api/`, `controllers/`, `services/`, `repositories/`, `models/`, `components/`, `utils/`, etc.
- File naming patterns: `*Controller.js`, `*Service.js`, `*Repository.js`, `*Model.js`, etc.
- Inferred layers: **API** (handlers, routes), **Service** (business logic), **Data** (repositories, models), **UI** (components, views), **Utility** (helpers, config, logging)

Files without a detected layer omit the field.

#### Feature #2: Business Logic Mapper

```bash

plaesy graph --business-logic                                      # export edges with business context
```

Exports INFERRED edges with layer context to `business-logic-queue.json`. AI can explain the business purpose of each relationship without needing code inspection.

#### Feature #3: Learning Path Generator

```bash

plaesy graph --generate-paths                                    # dependency-ordered learning paths
```

Analyzes dependencies to suggest entry points for code exploration. Useful for onboarding, module understanding, refactor planning.

#### Feature #4: Fuzzy/Semantic Search

```bash

plaesy graph --fuzzy-query "pattern" --search-depth 2          # multi-hop fuzzy search
```

Multi-hop search with fuzzy matching on file names/types; searches up to N hops from first match.

#### Feature #5: Diff Impact Visualization

```bash

plaesy graph --impact-check "file.js" --impact-visualize       # interactive HTML impact dashboard
```

Pre-commit safety check: shows all files affected by changes to a given file, grouped by dependency depth.

### Existing features (compatibility)

```bash

plaesy graph --query "common.go"                               # keyword query
plaesy graph --path-query-from "a.md" --path-query-to "b.md"   # shortest path
plaesy graph --explain "plaesy-init.go"                          # node details
plaesy graph --impact-check "common.go" --impact-depth 2       # text impact check
plaesy graph --semantic-queue                                   # export edges for AI (see §3)
plaesy graph --apply-semantic annotations.json                 # merge explanations
```

### Output (`.plaesy/analysis/`)

- `project.graph.json` — full node/edge list. Nodes: `{id, label, type, group, community, degree, [layer]}`. Edges: `{source, target, type, confidence}`
- `project.html` — self-contained force-directed viz (vanilla canvas, no CDN)
- `reports.md` — counts, type/folder breakdown, **detected communities** (label-propagation clustering over edge structure), god nodes, orphan files

### Edge types & confidence

- `references` (EXTRACTED) — markdown `[text](path)` link
- `sources` (EXTRACTED) — PowerShell dot-sourcing / Bash `source`
- `calls` (EXTRACTED) — PowerShell `& "foo.ps1"` invocation
- `imports` (EXTRACTED) — JS/TS/Python/Go import resolved to a file in the scanned tree
- `mirrors` (EXTRACTED) — same filename in two+ top-level folders (e.g. `prompts/assess.md` → `.claude/commands/assess.md`); note `agents`/`instructions` are *not* filename-preserving mirrors (suffix
  stripped), so verify against `scripts/internal/scaffold/copy.go` before assuming a mirror exists
- `mentions` (INFERRED) — plain-text path mention (weaker signal)

EXTRACTED = a concrete syntactic construct was parsed; INFERRED = derived from a text match. This distinction is what lets `query`/`explain` output be trusted at a glance.

### Community detection

`group` is folder-based. `community` is computed via label-propagation clustering over the actual edge graph — surfaces subsystems spanning folders (script + the instruction that references it + the
README linking both) that folder grouping cannot see.

### Impact check

`--impact-check <node> --impact-depth N` (default 2) runs BFS from the matched node and lists everything reachable within N hops, grouped by depth — use before renaming/deleting/refactoring a shared
file.

### Notes (graph)

- `--query`/`--path-query-*`/`--explain`/`--impact-check` all require `project.graph.json` to exist first — run `plaesy graph` with no flags once to build it.
- Re-run without flags after adding/renaming files, or use `--watch`.
- Repo-local, regex-based tool — no tree-sitter AST parsing; `imports` edges are syntax-pattern based (ES `import`/`require`, Python `import`/`from…import`, Go string imports) resolved against files
  in the scanned tree. Strong for markdown/Go and import-level edges in JS/TS/Python/Go.

---

## 2. `plaesy trim` — Token & Context Compression

### Purpose

Cuts token spend at the three points it accrues:

1. **Layer 1 — command output**: `plaesy trim run <command...>` executes a shell command and compresses its output before it reaches context, per-tool aware where it matters (git, npm, cargo, pytest,
   dotnet, docker via `.plaesy/scripts/configs/plaesy-trim-rules.json`). `mode: "test-results"` (pytest, cargo test, npm test, dotnet test) filters to failures + summary only — a passing 500-line
   test run collapses to one line instead of a generic head/tail cut.
2. **Layer 2 — memory/instruction files**, two modes:
  - `plaesy trim compress --path <file|dir>` — fast, local, heuristic (filler-phrase stripping + whitespace collapsing). Compression level is chosen per file from its `plaesy-graph` in-degree
    (`degree` field in `.plaesy/analysis/project.graph.json`) when `--level` is not given: god nodes (high in-degree) get the lightest touch (`lite`); rarely-referenced files get compressed harder
    (`ultra`).
  - `plaesy trim llm-queue`/`plaesy trim apply-llm` — LLM-quality rewrite done by the calling assistant itself, no API key, same no-network pattern
    as `plaesy graph --semantic-queue`/`--apply-semantic`. Denser and more semantically faithful than heuristic mode at the cost of a manual round-trip.
3. **Layer 3 — generated output discipline**: governed by the **Token & Output Efficiency Protocol** in `.plaesy/instructions/plaesy.md` (decision ladder: skip → reuse → stdlib → native → one-liner →
   minimal solution). Applies to every code-writing task.

### What is never touched

Code blocks (fenced ``` regions), shell commands, error messages/stack traces, and structured data (JSON/YAML/JS objects) are always preserved byte-exact in Layers 1 and 2. Only narrative prose is
compressed.

### Usage (Layer 1 + Layer 2)

```bash

plaesy trim run git status                                                    # Layer 1: compress one command's output
plaesy trim compress --path .plaesy/instructions/plaesy.md                    # Layer 2: compress one file
plaesy trim compress --path .plaesy/memory --recurse                          # Layer 2: compress memory folder
plaesy trim compress --path .plaesy/memory --dry-run                          # preview only, no write
plaesy trim compress --path README.md --level ultra                          # force a level
plaesy trim llm-queue --path .plaesy/instructions/plaesy.md                   # export prose to rewrite
plaesy trim apply-llm --path .plaesy/instructions/plaesy.md --annotations ann.json  # merge back
plaesy trim report                                                            # cumulative savings across layers
```

### Compression levels (Layer 2)

- `lite` — strip filler phrases (`.plaesy/scripts/configs/plaesy-trim-rules.json` → `layer2_filler_patterns`), collapse repeated blank lines
- `full` (default when degree is mid-range) — `lite` + merge short redundant sentences, trim hedging clauses
- `ultra` — `full` + telegram-style fragments; use only for low-traffic files, readability drops

Every `compress` run backs up the original as `<file>.bak` and appends a record to `.plaesy/memory/token-stats.json`; `report` reads that log — no re-scanning needed.

### Layer 1 rule format

`.plaesy/scripts/configs/plaesy-trim-rules.json` → `layer1_commands` maps a command prefix (e.g. `"git status"`) to `{ max_lines, dedupe_consecutive, mode }`. Unmatched → `_default`.

- `mode` omitted or `"generic"` — consecutive identical lines collapse to `line (xN)`; output beyond `max_lines` keeps first/last few lines with a `... N lines omitted ...` marker.
- `mode: "test-results"` — keeps only lines matching fail/error markers + the trailing summary line (e.g. `X failed, Y passed`); all-passing collapses to one confirmation line.

Still per-tool-*aware* line-shape compression, not a full AST/output parser — extend `layer1_commands` with a new `mode` (and matching branch in `compress_command_output`/`Compress-CommandOutput`) if
a command's output needs finer handling than dedupe/truncate/filter.

---

## 3. Semantic pass (no API key needed — shared by `trim` and `graph`)

Non-obvious relationships are explained by **the assistant already running the tool**, instead of calling an external LLM backend. The same round-trip pattern powers `plaesy trim apply-llm`
and `plaesy graph --apply-semantic`:

1. `plaesy graph --semantic-queue` (or `plaesy trim llm-queue`) splits output on fenced code blocks (never touching code) and writes prose segments ≥40 chars to `.plaesy/analysis/semantic-queue.json`
   as `{source, target, index, text}`.
2. Hand that file to the calling assistant with: "explain why each pair is related, one sentence each, return JSON array of `{source,target,rationale}`".
3. Save the reply as `annotations.json`, then `plaesy graph --apply-semantic annotations.json` (or `plaesy trim apply-llm --path <file> --annotations annotations.json`) — segment indices must match
   the original file's fence-toggle split exactly.

No network calls happen inside the tools; the semantic reasoning happens in the calling AI session.

---

## 4. Watch mode & re-runs

- `plaesy graph --watch` (poll interval: `--watch-interval`, default 3s) builds once, then polls file count + newest mtime (no content hashing) and rebuilds automatically. Stop with Ctrl+C.
- Re-run `plaesy graph` without flags after adding/renaming files to refresh; `plaesy trim compress` re-reads every session after a save.

*Context Engineering v1.0.0 — merged from the former `plaesy-trim.instructions.md` and `plaesy-graph.instructions.md` (single source, no PowerShell/Bash split).*
