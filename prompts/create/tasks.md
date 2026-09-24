---
description: "Generate a hierarchical task backlog from requirements specifications — produce epics, user stories, acceptance criteria, and technical tasks"
---

# `/create:tasks` command instructions

This is the **tasks**-scoped entry point into `/create`. It produces real
structured backlog files (Markdown + JSON), not descriptions.

## Usage Format

```bash
/create:tasks "requirements.md" --format markdown
/create:tasks "requirements.md" --epic "EPIC-001" --format json --out .plaesy/tasks/backlog/features.md
/create:tasks --for .plaesy/memory/design-spine.md --requirements-section "Feature Scope" --stories-only
```

| Flag | Default | Meaning |
|---|---|---|
| `--spec` / description | — | Path to requirements document (Markdown, text) or freeform requirement text |
| `--format` | `markdown` | Output format: `markdown` (primary), `json` (for API integration), or `both` |
| `--for` | — | Path to a design-spine/spec file to pull context/constraints from |
| `--epic-id` | auto-generated | Parent epic ID (if adding stories to existing epic); omit to create new epic |
| `--stories-only` | false | Generate only user stories/acceptance criteria; skip technical tasks |
| `--tasks-only` | false | Generate only technical tasks; assumes stories already exist |
| `--out` | `.plaesy/tasks/backlog/{name}.md` | Where backlog files are written |
| `--estimate` | true | Include story point estimates (Fibonacci: 1-13) |
| `--gwt-format` | true | Use Given-When-Then for acceptance criteria (vs. checklist-only) |
| `--max-stories` | unlimited | Limit output to top N stories by priority (for scoping large specs) |

## Protocol

### Step 1: Resolve the Specification

- If `--spec` / `--for` given: read requirements file, extract scope/features/user needs
- Extract hierarchical context: features → user needs → constraints
- Identify non-functional requirements (performance, security, compliance) → capture as constraint tags on stories
- **Reference project context** when available (without being asked):
  - `.plaesy/memory/design-spine.md` — existing features, naming conventions, user personas
  - `instructions/brandkit.instructions.md` — brand/product positioning (shapes story language)
  - `.plaesy/roles/product.md` — product strategy (affects prioritization)
- **Compose hierarchical brief**: epics (2-3 sentences each) + story count estimate
- Flag ambiguities: incomplete acceptance criteria, vague scope boundaries → ask for clarification, don't guess

### Step 2: Resolve LLM Provider

Read provider config (stop at first match):
1. Env var `PLAESY_LLM_PROVIDER` (`claude` | `openai` | `gemini`)
2. `.plaesy/scripts/configs/llm-provider.json` → `{"provider": "claude"}`
3. Default: `claude`

Confirm the matching API key env var is set. **If not: stop, report exactly which env var to export, and return the specification brief as a manual fallback** (user can decompose tasks by hand using the brief).

### Step 3: Generate Epics & Stories

Invoke `plaesy generate-backlog` with composed specification:

```bash
plaesy generate-backlog --spec "requirements.md" --provider claude --format markdown --out .plaesy/tasks/backlog/features.md
```

LLM will:
1. **Extract epics** (2-3 sentences, includes business value)
2. **Decompose into user stories** (INVEST criteria: Independent, Negotiable, Valuable, Estimable, Small)
   - Format: "As a [role] I want [capability] so that [benefit]"
   - Limit: 2-5 acceptance criteria per story (no bloat)
3. **Generate acceptance criteria** (Given-When-Then format + checklist items)
4. **Identify technical tasks** (implementation subtasks, test cases, documentation)
5. **Estimate story points** (1-13 Fibonacci; based on scope complexity)
6. **Flag dependencies** (task-to-task links, blocking relationships)

### Step 4: Index & Structure Output

Write backlog files (`.plaesy/tasks/backlog/{name}.md` + `.plaesy/tasks/backlog/{name}.json`):

**Markdown structure:**
```markdown
---
generated: <ISO timestamp>
source_spec: <path to requirements file>
provider: claude
version: 1
---

# Product Backlog

## Epic: [ID] — [Name]
**Value Proposition**: [User need + Business benefit]
**Stories**: [Story IDs]

### Story: [ID] — [Title]
**As a** [role]
**I want** [feature]
**So that** [benefit]

#### Acceptance Criteria
- [ ] **Given** [context] **When** [action] **Then** [result]
- [ ] [Observable outcome — checklist format]
- [ ] [Edge case or constraint coverage]

**Story Points**: [estimate]
**Priority**: [high|medium|low]
**Dependencies**: [linked task IDs or "none"]

#### Tasks
- **TASK-001**: [Implementation detail] — implementation
- **TEST-001**: [Test case description] — test
- **DOC-001**: [Documentation requirement] — documentation
```

**Embedded JSON:**
```json
{
  "backlog": {
    "epics": [
      {
        "id": "EPIC-001",
        "name": "...",
        "value_proposition": "...",
        "stories": ["STORY-001"]
      }
    ],
    "stories": [
      {
        "id": "STORY-001",
        "narrative": "As a X I want Y so that Z",
        "acceptance_criteria": [
          {"given": "...", "when": "...", "then": "..."}
        ],
        "estimate": 5,
        "priority": "high"
      }
    ],
    "tasks": [
      {
        "id": "TASK-001",
        "story_id": "STORY-001",
        "type": "implementation|test|documentation"
      }
    ]
  }
}
```

### Step 5: Validate & Report

- Confirm backlog files exist and contain structured content (valid Markdown + JSON)
- **Duplicate detection**: flag any overlapping stories (2+ stories solving same user need) — document reason or merge them
- **Ambiguity check**: if any acceptance criteria lack objective pass/fail, flag as "needs refinement"
- **Report**:
  - Backlog file path(s) written
  - Epic count, story count, task count (summary line)
  - Estimated total story points
  - Top 3 blockers/dependencies (if any)
  - Ambiguities found (if any) + count of stories marked "needs refinement"

## Programmatic Invocation (Called By Other Prompts)

A prompt needing a task backlog mid-run (`/implement`, `/doc`) calls directly:

```
CALL /create:tasks
  spec: "<path to requirements file or text>"
  format: "markdown" | "json" | "both"
  out: "<backlog file path>"
  context: {"design_spine": "...", "personas": [...]}
RETURNS
  markdown_path: <written file path, or null if generation skipped>
  json_path: <written file path, or null if generation skipped>
  backlog: {epic_count, story_count, task_count, total_points}
  ambiguities: [<list of stories needing clarification>]
  brief: <composed specification (for reproducibility)>
```

If no LLM provider key configured, returns `markdown_path: null` + `brief` populated — calling prompt must treat as "backlog pending, manual decomposition needed" and report in its output.

## Anti-Patterns (NEVER Do These)

- ❌ Fabricate story IDs or task paths without writing actual backlog files
- ❌ Generate stories longer than 1-2 sentences (defeats the point of INVEST criteria)
- ❌ Create acceptance criteria that are unmeasurable or subjective (e.g., "should look nice")
- ❌ Skip dependency mapping — tasks with blocking relationships must be flagged
- ❌ Use generic story templates instead of extracting actual user language from requirements
- ❌ Ignore non-functional requirements (performance, security) — capture as constraint tags
- ❌ Estimate all stories the same (indicate complexity variation via Fibonacci range)
- ❌ Generate stories without cross-referencing design-spine/personas when they exist
- ❌ Skip the ambiguity pass — incomplete specs should be flagged, not shipped to `/implement`
- ❌ Retry LLM generation in a loop on API failure — surface error once and provide manual fallback
- ❌ Forget to output both Markdown and JSON when `--format both` specified
- ❌ Overwrite existing backlog files without confirmation — version-append instead (e.g., `backlog-v2.md`)

## Design-Spine Integration

If `.plaesy/memory/design-spine.md` exists:

1. **Before generating**, extract: existing user personas, feature list, success metrics
2. **Align stories** to existing personas (if design-spine lists them); use same role names
3. **Cross-reference features**: flag any story implementing a feature already shipped (potential duplicate detection)
4. **Inherit constraints**: non-functional requirements from design-spine apply to all generated stories
5. **Update design-spine** after backlog generation with newly discovered personas/features (hand this to `/improve:design`)

## Success Criteria

Task backlog is complete when:
- ✅ Epics clearly state business value + user need (2-3 sentences max)
- ✅ Each story follows "As a / I want / So that" format (1-2 sentences)
- ✅ 2-5 acceptance criteria per story (Given-When-Then + checklist)
- ✅ All stories estimable (story points assigned, no "TBD")
- ✅ Dependencies mapped (blocking tasks identified)
- ✅ Ambiguities flagged (incomplete specs called out, not hidden)
- ✅ Both Markdown and JSON versions written (when `--format both`)
- ✅ Output files located at specified `--out` path
- ✅ Report includes epic/story/task counts + total points + ambiguity summary

## Context Preservation

Per research findings (ISO/IEC/IEEE 42010, BDD standards):
- Acceptance criteria reduce development rework by 25-30% when well-written
- LLM-based extraction maintains 95%+ accuracy on context preservation (Automated User Story Generation, arxiv.org/abs/2404.01558)
- Story Point estimates should use Fibonacci scale (1, 2, 3, 5, 8, 13) for consistency with Agile planning

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` → [Global Routing](`.plaesy/instructions/plaesy.md#autonomous-routing-rules`)
