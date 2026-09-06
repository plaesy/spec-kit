---
description: "Usage guide for the plaesy-graph knowledge-graph builder — load when the user wants to understand repo structure, relationships between chatmodes/instructions/scripts, or asks a codebase-graph question."
---

# Plaesy Graph — Repo Knowledge Graph

## Purpose
Plaesy Graph (available at `.plaesy/scripts/powershell/plaesy-graph.ps1` for Windows/PowerShell and `.plaesy/scripts/bash/plaesy-graph.sh` for Linux/macOS/Git Bash/WSL) turns your project — or any folder passed via `-Path`/`--path` — into a
lightweight knowledge graph: nodes are `.md`/`.ps1`/`.sh`/`.js`/`.ts`/`.py`/`.go`/`.dart` files, edges are markdown
links, PowerShell/Bash source/call references, resolved imports, install-mirror pairs, and plain-text path
mentions. No external dependency (Python, jq, LLM, network) is required — each is a single self-contained
script in its own language. See "Which script to run" below for how the choice between them is made.

## When to use
- User asks "how is X connected to Y", "what references this instruction/script", "which files are orphaned/unused".
- User wants an overview/visualization of chatmodes, instructions, checklists, templates, scripts relationships.
- Before large refactors of `.plaesy/memory/`, `.plaesy/chatmodes/`, or `.plaesy/scripts/`, to check what would be affected.

## Which script to run — pick automatically, don't ask the user
Two equivalent implementations exist: `.plaesy/scripts/powershell/plaesy-graph.ps1` and `.plaesy/scripts/bash/plaesy-graph.sh`.
There is no single cross-platform launcher (a dispatcher would itself need an interpreter chosen first, which
is the thing being decided — so the selection happens at the instruction level instead):
1. If the current tool-execution shell is PowerShell (Windows default, `pwsh`/`powershell.exe`) → use the `.ps1`.
2. If the current tool-execution shell is bash/zsh/sh (Linux, macOS, Git Bash, WSL) → use the `.sh`.
3. If genuinely unsure, prefer `.ps1` on Windows and `.sh` everywhere else — do not ask the user to choose.
4. The two scripts' outputs are independent (each rescans and overwrites `.plaesy/memory/analysis/`) — don't mix
   invocations across both in the same workflow, and note the `.sh` build is much slower on Windows/MSYS (subprocess
   spawn overhead) even though its `--query`/`--explain`/`--path-query`/`--impact-check` reads are fast either way.

```powershell
# PowerShell - Core Graph Operations
.plaesy\scripts\powershell\plaesy-graph.ps1                                          # build (or rebuild) the whole repo
.plaesy\scripts\powershell\plaesy-graph.ps1 -Path instructions                        # build for a specific subfolder
.plaesy\scripts\powershell\plaesy-graph.ps1 -Watch                                   # rebuild automatically on source changes

# Feature #1: Architecture Layer Detection (auto-included in graph)
# Nodes will include 'layer' attribute: API, Service, Data, UI, Utility

# Feature #2: Business Logic Mapper
.plaesy\scripts\powershell\plaesy-graph.ps1 -BusinessLogic                           # export edges with business context

# Feature #3: Learning Path Generator
.plaesy\scripts\powershell\plaesy-graph.ps1 -GenerateLearningPaths                   # generate dependency-ordered learning paths

# Feature #4: Fuzzy/Semantic Search
.plaesy\scripts\powershell\plaesy-graph.ps1 -FuzzyQuery "pattern" -SearchDepth 2     # multi-hop fuzzy search

# Feature #5: Diff Impact Visualization
.plaesy\scripts\powershell\plaesy-graph.ps1 -ImpactCheck "file.js" -ImpactVisualize  # generate HTML visualization

# Existing features (maintained compatibility)
.plaesy\scripts\powershell\plaesy-graph.ps1 -Query "common.ps1"                       # keyword query
.plaesy\scripts\powershell\plaesy-graph.ps1 -PathFrom "a.md" -PathTo "b.md"          # shortest path
.plaesy\scripts\powershell\plaesy-graph.ps1 -Explain "plaesy-init.ps1"                # node details
.plaesy\scripts\powershell\plaesy-graph.ps1 -ImpactCheck "common.ps1" -ImpactDepth 2 # text impact check
.plaesy\scripts\powershell\plaesy-graph.ps1 -SemanticQueue                           # export edges for AI
.plaesy\scripts\powershell\plaesy-graph.ps1 -ApplySemantic annotations.json          # merge explanations
```

```bash
# Bash (identical behavior/output shape) - feature parity with PowerShell
.plaesy/scripts/bash/plaesy-graph.sh                                       # build (or rebuild) the whole repo
.plaesy/scripts/bash/plaesy-graph.sh --path .plaesy/memory                 # build for a specific subfolder
.plaesy/scripts/bash/plaesy-graph.sh --watch                               # rebuild automatically on source changes

# Feature #2: Business Logic Mapper
.plaesy/scripts/bash/plaesy-graph.sh --business-logic                                # export edges with business context

# Feature #3: Learning Path Generator
.plaesy/scripts/bash/plaesy-graph.sh --generate-paths                                # generate learning paths

# Feature #4: Fuzzy/Semantic Search
.plaesy/scripts/bash/plaesy-graph.sh --fuzzy-query "pattern" --search-depth 2        # multi-hop fuzzy search

# Feature #5: Diff Impact Visualization
.plaesy/scripts/bash/plaesy-graph.sh --impact-check "file.js" --impact-visualize     # generate HTML visualization

# Existing features (maintained compatibility)
.plaesy/scripts/bash/plaesy-graph.sh --query "common.ps1"                             # keyword query
.plaesy/scripts/bash/plaesy-graph.sh --path-query "a.md" "b.md"                       # shortest path
.plaesy/scripts/bash/plaesy-graph.sh --explain "plaesy-init.ps1"                      # node details
.plaesy/scripts/bash/plaesy-graph.sh --impact-check "file.js" --impact-depth 2        # text impact check
.plaesy/scripts/bash/plaesy-graph.sh --semantic-queue                                 # export edges for AI
.plaesy/scripts/bash/plaesy-graph.sh --apply-semantic annotations.json                # merge explanations
```

## Output (`.plaesy/memory/analysis/`)
- `project.graph.json` — full node/edge list. Nodes: `{id, label, type, group, community, degree, [layer]}`. Edges: `{source, target, type, confidence}`
- `project.html` — self-contained force-directed viz (vanilla canvas, no CDN) — open directly in a browser
- `reports.md` — node/edge counts, breakdown by type, folder-based grouping, **detected communities** (real label-propagation clustering over edge structure — surfaces cross-folder subsystems folder names hide), god nodes, orphan files

### Architecture Layer Detection
When analyzing arbitrary codebases (not the spec-kit repo itself), each node receives a semantic `layer` attribute based on:
- Folder naming patterns: `api/`, `controllers/`, `services/`, `repositories/`, `models/`, `components/`, `utils/`, etc.
- File naming patterns: `*Controller.js`, `*Service.js`, `*Repository.js`, `*Model.js`, etc.
- Inferred layer categories: **API** (handlers, routes), **Service** (business logic), **Data** (repositories, models), **UI** (components, views), **Utility** (helpers, config, logging)

The `layer` field helps visualize architectural boundaries and dependency violations at a glance. Files without a detected layer pattern are not assigned a layer (the field is omitted from JSON).

## Edge types & confidence
- `references` (EXTRACTED) — markdown `[text](path)` link
- `sources` (EXTRACTED) — PowerShell dot-sourcing / Bash `source`
- `calls` (EXTRACTED) — PowerShell `& "foo.ps1"` invocation
- `imports` (EXTRACTED) — JS/TS/Python/Go import resolved to a file in the scanned tree
- `mirrors` (EXTRACTED) — same filename present in two+ different top-level folders. This repo installs
  chatmodes/instructions/prompts into multiple per-editor target directories (see `.plaesy/scripts/configs/platform.json`,
  e.g. `chatmodes/*.chatmode.md` → `.claude/roles/*.chatmode.md`) — without this edge, every install copy looked
  like an orphan even though it's a 1:1 mirror of a source file. Cut this repo's orphan count from 87 to 23.
- `mentions` (INFERRED) — plain-text path mention, weaker signal

EXTRACTED means a concrete syntactic construct was parsed; INFERRED means it was derived from a text match — this distinction is what lets query/explain output be trusted at a glance instead of treating every edge as equally certain.

## Community detection
`group` is folder-based (cheap, structural). `community` is computed via label-propagation clustering over
the actual edge graph — nodes converge to the label most common among their neighbors over up to 15 iterations.
This surfaces subsystems that span folders (e.g. a script + the instructions doc that references it + the
README that links both), which a folder-name grouping cannot see.

## Impact check
`-ImpactCheck <node> -ImpactDepth N` (default depth 2) runs a breadth-first traversal from the matched node
and lists everything reachable within N hops, grouped by depth — use before renaming/deleting/refactoring a
shared file (e.g. `common.ps1`, a chatmode, an instruction) to see the blast radius before touching it.

## Semantic pass (no API key needed)
Non-obvious relationships get explained by using **the assistant already running this tool**, instead of
calling out to an external LLM backend:
1. `-SemanticQueue` dumps every INFERRED edge to `.plaesy/memory/analysis/semantic-queue.json`.
2. Give that file to Claude (this session or a fresh one) with: "explain why each pair is related, one sentence
   each, return JSON array of `{source,target,rationale}`".
3. Save the reply as `annotations.json`, then run `-ApplySemantic annotations.json` — it merges `rationale`
   into the matching edges in `project.graph.json` and adds an "Explained relationships" section to `reports.md`.
No network calls happen inside the script itself; the semantic reasoning happens in the calling AI session.

## Watch mode
`-Watch`/`--watch` (poll interval: `-WatchInterval`/`--watch-interval`, default 3s) builds the graph once, then
polls the same file set for changes (file count + newest mtime — no content hashing) and rebuilds automatically
whenever something changes. Useful while actively editing chatmodes/instructions/scripts and wanting
`project.graph.json`/`reports.md` to stay current without re-running the build by hand. Stop with Ctrl+C.

## Advanced Features (Phase 2)

### Feature #1: Architecture Layer Detection
Automatically detected during graph build. Each node gets a semantic `layer` attribute:
- **API** — handlers, controllers, middleware, routes, app setup
- **Service** — business logic, use cases, orchestration
- **Data** — repositories, models, entities, schemas
- **UI** — components, views, pages, screens
- **Utility** — helpers, logging, validation, configuration

Useful for: detecting layer violations, visualizing architecture, understanding code organization at a glance.

### Feature #2: Business Logic Mapper
`-BusinessLogic` / `--business-logic` exports INFERRED edges with layer context to `business-logic-queue.json`. 
AI can then explain the business purpose of each relationship without needing code inspection. Output includes source/target layers for context.

### Feature #3: Learning Path Generator
`-GenerateLearningPaths` / `--generate-paths` analyzes dependencies to suggest entry points for code exploration.
Useful for: onboarding new developers, understanding module dependencies, planning refactors.

### Feature #4: Fuzzy/Semantic Search
`-FuzzyQuery "pattern" -SearchDepth N` / `--fuzzy-query` performs multi-hop search with fuzzy matching on file names and types.
Searches up to N hops from first match, surfacing connected ecosystem around a concept.

### Feature #5: Diff Impact Visualization
`-ImpactCheck "file.js" -ImpactVisualize` / `--impact-check --impact-visualize` generates an interactive HTML dashboard showing all files affected by changes to a given file, grouped by dependency depth.
Useful for: pre-commit safety checks, understanding change scope, planning migration strategies.

## Language coverage (Phase 2)
When `-Path` targets an arbitrary codebase (not this repo), `.js/.jsx/.ts/.tsx/.py/.go` files are indexed too,
with per-language `imports` edges (ES `import`/`require`, Python `import`/`from…import`, Go string imports)
resolved against files actually in the scanned tree — bare package imports (npm packages, Python stdlib) are
intentionally skipped rather than guessed at, so every `imports` edge is EXTRACTED-grade, not a guess.

## Notes
- `-Query`/`-PathFrom`/`-Explain`/`-ImpactCheck` all require `project.graph.json` to exist first — run the script with no flags once to build it.
- Re-run without flags after adding/renaming files to refresh the graph; there is no `--watch` mode (keep it simple).
- This is a repo-local, regex-based tool — no tree-sitter AST parsing yet, so `calls`/`sources`/`references`/`imports`
  are syntax-pattern based rather than true compiler-level call-graph analysis. Strong for this repo's
  markdown/PowerShell/Bash files and for import-level edges in JS/TS/Python/Go; a deeper AST pass across more
  languages is tracked as future work if a target codebase needs it.
