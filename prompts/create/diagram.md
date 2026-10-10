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
/create:diagram --type erd --for .plaesy/memory/design.md --component "data-model" --format svg
/create:diagram "Payment processing workflow across frontend, backend, payment processor" --type swimlane --roles "User,Backend,Provider" --out assets/diagrams/payment-swimlane.svg
/create:diagram --type sequence "Mobile app requests data → backend query → database response" --out docs/sequences/data-fetch.mmd
```

| Flag | Default | Meaning |
|---|---|---|
| `--type` | `flowchart` | Diagram type: `flowchart`, `architecture`, `erd` (entity-relationship), `sequence`, `swimlane`, `mindmap`, `activity`, `state`, `class`, `bpmn` |
| `--title` | Inferred from description | Diagram title/heading |
| `--format` | `mmd` (Mermaid markdown) | Output format: `mmd` (markdown), `svg`, `json` (for tooling); SVG requires tool support |
| `--for` | — | Path to a design-context/spec file (e.g. `.plaesy/memory/design.md`) to pull diagram context from instead of freeform description |
| `--component` | — | With `--for`, which component/system to diagram |
| `--style` | Inferred from `.plaesy/memory/design.md` | Visual style: `default`, `dark`, `neutral` (for Mermaid themes) |
| `--roles` | — | For swimlanes: comma-separated participant/actor names |
| `--output-dir` | `docs/diagrams` | Base directory where diagram files are written |
| `--out` | `<output-dir>/<slugified-title>.<format>` | Explicit output file path |
| `--validate` | `true` | Validate syntax before writing; set to `false` to skip |

## Protocol

### Step 0: Ensure Design Context Exists

- Check whether `.plaesy/memory/design.md` exists (project-wide design tokens +
  rationale, Google Labs DESIGN.md convention; field list in
  `.plaesy/instructions/how-to-create-designmd.md`).
- **Missing** → derive architecture/data-model conventions from the actual
  codebase (read the code, don't invent), seed `.plaesy/memory/design.md`
  from `.plaesy/templates/design.template.md` where conventions genuinely
  can't be derived, label that part `ASSUMED`, record it, and continue.
- **Exists** → use it as source of truth for component names, data structures,
  and architectural patterns in Step 1.

### Step 1: Resolve the Brief

- If `--for`/`--component` given: read that file, extract the component's system
  architecture or data model. If missing entirely, derive a one-line diagram
  scope from the component/file name and surrounding code, label it
  `ASSUMED — <scope>`, and record it rather than guessing silently.
- Otherwise the description/argument *is* the brief.
- **Infer diagram type** (if not `--type` specified):
  - Contains "flow" / "process" / "steps" → `flowchart`
  - Contains "system" / "architecture" / "services" / "APIs" → `architecture`
  - Contains "table" / "entity" / "relationship" / "data model" → `erd`
  - Contains "message" / "request-response" / "sequence of calls" → `sequence`
  - Contains "across teams" / "handoff" / "roles" / "swimlane" → `swimlane`
  - Contains "idea" / "topic" / "breakdown" / "hierarchy" → `mindmap`
- Layer in project context when it exists:
  - `.plaesy/memory/design.md` — system architecture patterns, data structures, component conventions
  - `.plaesy/roles/dev.md` / `.plaesy/roles/sa.md` — technical constraints, naming conventions, standards
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
  - Source of truth (code repo, database, `.plaesy/memory/design.md`, manual)
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

```text

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

If no tool is available for the requested type, return `tool_used: null` and `files: {json}` (structured definition only). The calling prompt must treat this as "diagram pending, manual rendering
needed" and say so.

## Anatomy of Diagram Output

```text

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

## What a Good Diagram Does

- **Produce a saved artifact.** A text description of what the diagram should
  show is not a diagram. If the tool is unavailable, say so explicitly and hand
  over the DSL as a documented fallback — the two are not the same result
- **Confirm the file exists before reporting it** — a path that was invented is a
  claim, not an artifact
- **Match the diagram type to the content**: an ERD for data models, a flowchart
  for single-actor workflows. A flowchart of a data model loses the relationships
  it exists to show.
- **Save the diagram as a version-controlled project file** — `/create:diagram`
  produces files, not embedded visualizations
- **Surface a tool failure once and stop**, naming what is missing — tool,
  version, or environment variable
- **Establish the audience before generating** — an executive diagram carries
  different information and abstraction than an engineer diagram (per
  https://architecturediagram.ai/blog/architecture-diagram-best-practices)
- **Add a legend once more than two visual conventions encode meaning**
  (color, shape, line style)
- **Write to a new path** unless the user explicitly asked to replace the existing
  diagram file
- **Write the `.plaesy/memory/diagrams/<name>.md` index file** — other prompts and
  documentation reach the diagram through it
- **Layer the project's naming conventions and patterns from
  `.plaesy/memory/design.md` into an architecture diagram** — it should look like
  it belongs to the project, not generic stock art
- **Split unrelated systems into separate diagrams**, one per subsystem or flow
- **Document every diagram in its index file** — purpose, audience, and links to
  the files
- **Derive conventions from the code, or seed the template, at Step 0 when
  `.plaesy/memory/design.md` is absent**, labeled `ASSUMED` where seeded,
  rather than diagramming generic placeholders unlabeled

## Design-Spine Integration

If the project has a `.plaesy/memory/design.md` with defined system architecture or data structures:

1. **Before generating**, check for the diagram name/scope in the Design Spine
2. If found, extract component names, architectural patterns, data model constraints
3. Layer these into the diagram DSL (use project-standard names, color schemes, symbols)
4. If a diagram is **missing** from the Design Spine but the user requests it, that's a finding for `/assess:technical` — note it in the diagram index as a gap

Example: if `.plaesy/memory/design.md` defines microservices (auth, api, worker, database), an architecture diagram should use those exact names and show their relationships per that file, not generic
placeholders.

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
