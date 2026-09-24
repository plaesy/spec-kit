---
description: "Generate visual diagrams (architecture, flowchart, ERD, sequence, swimlane, mindmap) from natural language or code context — produces SVG and Mermaid markdown files, not just descriptions"
---

# `/create:diagram` command instructions

This is the **diagram**-scoped entry point into `/create`. It produces real
saved diagram files (SVG primary, Mermaid markdown secondary), not descriptions
or abstract prompts.

## Usage Format

```bash
/create:diagram "System architecture with microservices, API gateway, databases"
/create:diagram --type flowchart --title "User Signup Flow" --out docs/diagrams/signup-flow.mmd
/create:diagram --type erd --for .plaesy/memory/design-spine.md --component "data-model" --format svg
/create:diagram "Payment processing workflow across frontend, backend, payment processor" --type swimlane --roles "User,Backend,Provider" --out assets/diagrams/payment-swimlane.svg
/create:diagram --type sequence "Mobile app requests data → backend query → database response" --out docs/sequences/data-fetch.mmd
```

| Flag | Default | Meaning |
|---|---|---|
| `--type` | `flowchart` | Diagram type: `flowchart`, `architecture`, `erd` (entity-relationship), `sequence`, `swimlane`, `mindmap`, `activity`, `state`, `class`, `bpmn` |
| `--title` | Inferred from description | Diagram title/heading |
| `--format` | `mmd` (Mermaid markdown) | Output format: `mmd` (markdown), `svg`, `json` (for tooling); SVG requires tool support |
| `--for` | — | Path to design-spine/spec file to pull diagram context from instead of freeform description |
| `--component` | — | With `--for`, which component/system to diagram |
| `--style` | Inferred from `.plaesy/memory/design-spine.md` | Visual style: `default`, `dark`, `neutral` (for Mermaid themes) |
| `--roles` | — | For swimlanes: comma-separated participant/actor names |
| `--output-dir` | `docs/diagrams` | Base directory where diagram files are written |
| `--out` | `<output-dir>/<slugified-title>.<format>` | Explicit output file path |
| `--validate` | `true` | Validate syntax before writing; set to `false` to skip |

## Protocol

### Step 1: Resolve the Brief

- If `--for`/`--component` given: read that file, extract the component's system architecture or data model. If missing entirely, ask the user for a one-line diagram scope rather than guessing.
- Otherwise the description/argument *is* the brief.
- **Infer diagram type** (if not `--type` specified):
  - Contains "flow" / "process" / "steps" → `flowchart`
  - Contains "system" / "architecture" / "services" / "APIs" → `architecture`
  - Contains "table" / "entity" / "relationship" / "data model" → `erd`
  - Contains "message" / "request-response" / "sequence of calls" → `sequence`
  - Contains "across teams" / "handoff" / "roles" / "swimlane" → `swimlane`
  - Contains "idea" / "topic" / "breakdown" / "hierarchy" → `mindmap`
- Layer in project context when it exists:
  - `.plaesy/memory/design-spine.md` — system architecture patterns, data structures, component conventions
  - `.plaesy/roles/technical.md` / `.plaesy/roles/architect.md` — technical constraints, naming conventions, standards
- Compose the final diagram brief: `{description}, type={inferred or given}, title={title or auto}, roles={roles if swimlane}, audience={technical/executive/designer}`.

### Step 2: Resolve the Provider & Tool

Check tool availability in this order:

1. **Mermaid CLI** (preferred for `mmd` output, works offline):
   - Check: `mmdc --version` succeeds
   - Handles: flowchart, architecture (graph), ERD, sequence, activity, state, class, mindmap
   - Output: `.mmd` markdown (always), `.svg` (with `--output` flag)

2. **PlantUML** (for advanced UML, SVG export):
   - Check: `plantuml -version` succeeds and `PLANTUML_JAR` env var OR online render
   - Handles: sequence, activity, class, state, component, deployment diagrams
   - Output: `.puml` definition, `.svg` export (with `-tsvg` flag)

3. **Graphviz** (for network/hierarchical graphs):
   - Check: `dot -V` succeeds
   - Handles: directed graphs, hierarchies, network topologies
   - Output: `.dot` definition, `.svg` (with `-Tsvg` flag)

4. **JSON fallback** (structured definition, no rendering):
   - Always available
   - Returns JSON schema of diagram structure (for further tooling/automation)

**If no tool is available for the requested type and format, stop and report exactly what is missing.** Do not attempt manual SVG generation or ASCII art — be explicit about the fallback.

### Step 3: Generate

- **For Mermaid** (`mmd` output):
  - Compose the Mermaid DSL syntax directly in the brief resolution step
  - Write to `<out>.mmd` file
  - If `--format svg` requested, invoke `mmdc --input <out>.mmd --output <out>.svg`
  
- **For PlantUML** (`puml` output):
  - Compose PlantUML DSL in brief
  - Write to `<out>.puml` file
  - If SVG output: `plantuml -tsvg <out>.puml`
  
- **For Graphviz** (`.dot` output):
  - Compose dot language
  - Write to `<out>.dot` file
  - If SVG output: `dot -Tsvg <out>.dot -o <out>.svg`

- **Run tool command once** (never retry on failure); surface stderr verbatim.

- **For multiple formats** (both `.mmd` and `.svg`): generate each separately; if either fails, surface the error and stop (don't proceed to the other format).

### Step 4: Index & Document

- Write a `.plaesy/memory/diagrams/<diagram-name>.md` index file containing:
  - Diagram title and type
  - Purpose/scope (what stakeholders see it / why it exists)
  - Link to generated files (`.mmd`, `.svg`, if multiple formats)
  - Source of truth (code repo, database, design-spine, manual)
  - Last updated date
  - Example:
    ```markdown
    # User Signup Flow
    
    **Type**: Flowchart  
    **Audience**: Product team, backend engineers  
    **Updated**: 2026-09-25  
    
    ## Files
    - [Mermaid source](../../docs/diagrams/signup-flow.mmd)
    - [SVG export](../../docs/diagrams/signup-flow.svg)
    
    ## Purpose
    Illustrates happy-path user signup from landing page through email verification.
    
    ## Steps
    1. User visits homepage
    2. Clicks "Sign Up" button
    3. Enters email/password
    4. Receives verification email
    5. Clicks link, completes onboarding
    ```

- Update `.plaesy/memory/diagrams.md` to index all diagrams:
  ```markdown
  # Diagrams Index
  
  ## Architecture
  - [System Architecture](diagrams/system-architecture.md) — Microservices, APIs, databases
  
  ## Flows
  - [User Signup](diagrams/signup-flow.md) — Happy path (5 steps) + error cases
  - [Payment Checkout](diagrams/payment-checkout.md) — Swimlane (3 actors)
  
  ## Data Models
  - [User & Org Schema](diagrams/user-org-erd.md) — Entity relationships
  ```

### Step 5: Validate & Report

- Confirm the output file(s) exist and are non-zero bytes
- If tool invocation failed: surface stderr once, stop (don't retry)
- Report:
  - File path(s) written (`.mmd`, `.svg`, etc.)
  - Diagram type and title
  - Index file path (`.plaesy/memory/diagrams/<name>.md`)
  - Audience/purpose (who should read this, why)
  - Tool used (Mermaid CLI / PlantUML / Graphviz)

## Programmatic Invocation (Called By Other Prompts)

A prompt that needs a diagram mid-run (`/implement:technical`, `/improve:spec`, `/doc`, `/assess:technical`) calls this protocol directly:

```
CALL /create:diagram
  brief: "<one-line or multi-sentence diagram scope>"
  type: "<flowchart|architecture|erd|sequence|swimlane|mindmap>"
  title: "<diagram title; omit to auto-compose>"
  format: "<mmd|svg|json; default: mmd>"
  roles: "<comma-separated names for swimlanes; omit for other types>"
  out: "<file path the calling prompt will reference>"
RETURNS
  files: {primary: <path>, secondary: <path if multi-format>, json: <path if requested>}
  type: <diagram_type>
  title: <diagram_title>
  index_file: <path to .plaesy/memory/diagrams/{name}.md>
  dsl_used: <raw Mermaid/PlantUML/Graphviz definition, for reproducibility>
  tool_used: <"Mermaid CLI"|"PlantUML"|"Graphviz"|null>
```

If a tool is not available for the requested type, return `tool_used: null` and `files: {json}` (structured definition only). The calling prompt must treat this as "diagram pending, manual rendering needed" and say so in its own output.

## Anatomy of Diagram Output

```
docs/diagrams/                       # Primary output directory
├── signup-flow.mmd                  # Mermaid markdown (text, version-control friendly)
├── signup-flow.svg                  # Mermaid SVG export (if --format svg requested)
├── system-architecture.mmd
├── system-architecture.svg
├── user-org-erd.mmd
└── user-org-erd.json                # JSON schema (if --format json requested)

.plaesy/memory/diagrams/             # Index & metadata
├── signup-flow.md                   # Diagram metadata & links
├── system-architecture.md
└── diagrams.md                      # Master index of all diagrams
```

Diagram types and their best-use cases (per research in https://clickhelp.com/clickhelp-technical-writing-blog/best-diagram-and-flowchart-tools/):

| Type | Use Case | Audience | Tools |
|---|---|---|---|
| **Flowchart** | Decision trees, process steps, user flows | Product, engineering | Mermaid, PlantUML, Graphviz |
| **Architecture** | System design, microservices, APIs, infrastructure | Engineering, architects | Mermaid (graph), Diagrams CLI, PlantUML |
| **ERD** | Database schema, entity relationships, data models | Data engineers, backend | Mermaid, PlantUML, Diagrams CLI |
| **Sequence** | Message flow, API calls, request-response patterns | Backend, QA | Mermaid, PlantUML |
| **Swimlane** | Cross-team processes, handoffs, BPMN workflows | Product, operations, legal | Mermaid, PlantUML, Creately |
| **Mindmap** | Brainstorm output, topic breakdown, hierarchies | Product, design, strategy | Mermaid, EdrawMax |
| **Activity** | Activity flow, parallel processes, decision points | Engineering | PlantUML, Mermaid |
| **State** | State machine, lifecycle, finite-state automaton | Backend, frontend | PlantUML, Mermaid |
| **Class** | OOP design, inheritance hierarchy, relationships | Architecture, code review | PlantUML |
| **BPMN** | Business process model, compliance workflows | Operations, legal, auditing | Creately, EdrawMax, Lucidchart |

## Anti-Patterns (NEVER Do These)

- ❌ Return only a text description ("here's what the diagram should show") without actually generating a file — the point is a saved artifact; if tool is unavailable, report that explicitly and provide the DSL as a documented fallback
- ❌ Invent a file path or say "diagram written to X" without actually creating it — confirm file existence before reporting
- ❌ Choose a diagram type that doesn't match the content — flowcharts for data models (use ERD), swimlanes for single-actor workflows (use flowchart)
- ❌ Embed the diagram in an artifact/page when it should be a saved file in the project — `/create:diagram` is for version-controlled files, not embedded visualizations
- ❌ Retry tool invocation in a loop on failure — surface the error once, stop, and tell the user what is missing (tool, version, environment variable)
- ❌ Generate a diagram without audience context — a diagram for executives needs different information/abstraction than one for engineers (per https://architecturediagram.ai/blog/architecture-diagram-best-practices)
- ❌ Ignore visual conventions and legends when using color/shape/line-style encoding — add a legend if more than two visual conventions are used
- ❌ Overwrite an existing diagram file without confirmation — write to a new path unless the user explicitly asked to replace it
- ❌ Skip the `.plaesy/memory/diagrams/<name>.md` index file — it must exist so other prompts and documentation can reference and explain the diagram
- ❌ Generate an architecture diagram without layering project naming conventions and patterns from design-spine — this should look like it belongs to the project, not generic stock art
- ❌ Mix multiple unrelated systems in one diagram — split into separate diagrams by scope (one per subsystem or flow)
- ❌ Leave a diagram undocumented — always write the index file with purpose, audience, and link to files

## Design-Spine Integration

If the project has a `.plaesy/memory/design-spine.md` with defined system architecture or data structures:

1. **Before generating**, check for the diagram name/scope in the Design Spine
2. If found, extract component names, architectural patterns, data model constraints
3. Layer these into the diagram DSL (use project-standard names, color schemes, symbols)
4. If a diagram is **missing** from the Design Spine but the user requests it, that's a finding for `/assess:technical` — note it in the diagram index as a gap

Example: if design-spine defines microservices (auth, api, worker, database), an architecture diagram should use those exact names and show their relationships per design-spine, not generic placeholders.

## Success Criteria

A diagram generation is complete when:
- ✅ File(s) written to disk (`.mmd` and/or `.svg` as requested)
- ✅ Files are non-zero bytes and valid syntax (tool did not error)
- ✅ Diagram matches the brief (scope, type, audience correct)
- ✅ Audience-targeted information is present (executives see context; engineers see implementation details)
- ✅ Visual conventions are consistent and documented (legend if using color/shape encoding)
- ✅ Index file written to `.plaesy/memory/diagrams/<diagram-name>.md`
- ✅ `.plaesy/memory/diagrams.md` updated with link to new diagram
- ✅ Output file path(s) reported (e.g., `docs/diagrams/signup-flow.mmd` + `.svg`)
- ✅ Tool used reported (Mermaid CLI / PlantUML / Graphviz)

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`

**Research sources**:
- https://clickhelp.com/clickhelp-technical-writing-blog/best-diagram-and-flowchart-tools/
- https://architecturediagram.ai/blog/architecture-diagram-best-practices
- https://mermaid.js.org/
- https://plantuml.com/
- https://graphviz.org/
