---
description: "Generate task backlog from requirements specifications — produces `.plaesy/tasks/` frontmatter-formatted task files conforming to instructions/tasks.instructions.md"
---

# `/create:tasks` command instructions

This is the **tasks**-scoped entry point into `/create`. It produces real
task files (`.md`) conforming to `.plaesy/tasks/` structure, not freeform markdown or descriptions.

**CRITICAL**: All generated tasks MUST conform to `instructions/tasks.instructions.md` format:

- Frontmatter: `title`, `phase`, `status`, `createdAt`, `updatedAt`
- Directory: `.plaesy/tasks/backlog/` (or `/todo/` if pre-prioritized)
- Naming: `{priority}_{title}.md` (critical_, high_, medium_, low_ prefixes)
- Sections: Description, Acceptance Criteria, Dependencies, References

## Usage Format

```bash

/create:tasks "requirements.md" --format markdown
/create:tasks "requirements.md" --epic "EPIC-001" --format json --out .plaesy/tasks/backlog/features.json
/create:tasks --for .plaesy/memory/design.md --requirements-section "Feature Scope" --stories-only
```

| Flag | Default | Meaning |
|---|---|---|
| `--spec` / description | — | Path to requirements document (Markdown, text) or freeform requirement text |
| `--format` | `markdown` | Output format: `markdown` (one task file per task — primary), `json` (single derived index), or `both` |
| `--for` | — | Path to a design-context/spec file (e.g. `.plaesy/memory/design.md`) to pull context/constraints from |
| `--epic-id` | auto-generated | Parent epic ID (if adding stories to existing epic); omit to create new epic |
| `--stories-only` | false | Generate only user stories/acceptance criteria; skip technical tasks |
| `--tasks-only` | false | Generate only technical tasks; assumes stories already exist |
| `--out` | `.plaesy/tasks/backlog/` (per-task files) / `.plaesy/tasks/backlog/backlog.json` (`--format json`) | Where the backlog is written; a directory for `markdown`, a file path for `json` |
| `--estimate` | true | Include story point estimates (Fibonacci: 1-13) |
| `--gwt-format` | true | Use Given-When-Then for acceptance criteria (vs. checklist-only) |
| `--max-stories` | unlimited | Limit output to top N stories by priority (for scoping large specs) |

## Protocol

### Step 0: Ensure Design Context Exists

- Check whether `.plaesy/memory/design.md` exists (project-wide design tokens +
  rationale, Google Labs DESIGN.md convention; field list in
  `.plaesy/instructions/how-to-create-designmd.md`).
- **Missing** → derive features/personas/naming from the actual spec and
  codebase where possible; seed `.plaesy/memory/design.md` from
  `.plaesy/templates/design.template.md` for what can't be derived, label
  that part `ASSUMED`, record it, and continue.
- **Exists** → use it as source of truth for features, personas, and
  constraints in Step 1.

### Step 1: Resolve the Specification

- If `--spec` / `--for` given: read requirements file, extract scope/features/user needs
- Extract hierarchical context: features → user needs → constraints
- Identify non-functional requirements (performance, security, compliance) → capture as constraint tags on stories
- **Reference project context** when available (without being asked):
  - `.plaesy/memory/design.md` — existing features, naming conventions, user personas
  - `.plaesy/instructions/brandkit.md` — brand/product positioning (shapes story
    language); only installed when the project declares a brand kit, so if the file
    is missing, derive positioning from existing product docs/README where
    possible, otherwise label it `ASSUMED — generic positioning` and record it
  - `.plaesy/roles/po.md` — product strategy (affects prioritization)
- **Compose hierarchical brief**: epics (2-3 sentences each) + story count estimate
- Flag ambiguities: incomplete acceptance criteria, vague scope boundaries → apply the best-practice default, label `ASSUMED`, record it — don't silently guess

### Step 2: Resolve LLM Provider

Resolve the provider, in this order, stop at first match:

1. Env var `PLAESY_LLM_PROVIDER` (`claude` | `openai` | `gemini`)
   ⚠️ No shipped code reads this variable — it is not plaesy configuration, it is
   a convention the agent is asked to honour itself. Setting it in a shell does
   nothing on its own. Read it if set, otherwise fall through.
2. Optional user override file `.plaesy/config/llm-provider.json` → `{"provider": "claude"}`
   (nothing writes this file; the user creates it themselves to pin a project-wide
   default)
3. Default: `claude`
4. If the env var was absent, the default is unconfirmed — name the resolved
   provider in the report so a wrong pick is visible immediately

Confirm the matching API key env var is set. **If not: stop, report exactly which env var to export, and return the specification brief as a manual fallback** (user can decompose tasks by hand using
the brief).

### Step 3: Generate Tasks Conforming to Task System Format

Follow `instructions/tasks.instructions.md` template structure (NOT freeform markdown):

For each task, LLM will:

1. **Extract requirements** from spec (features, stories, acceptance criteria)
2. **Decompose into tasks** (following `templates/tasks.template.md` pattern)
3. **Determine task type**: Setup, Test (TDD RED phase first), Core, Integration, Polish
4. **Assign phase**: design|implement|assess|optimize|fix
5. **Assign priority**: critical|high|medium|low (for filename prefix)
6. **Identify dependencies**: blocking tasks, external blockers
7. **Generate frontmatter**: title, phase, status (always "backlog" for new), createdAt, updatedAt

### Step 4: Write Task Files to `.plaesy/tasks/` Structure

Write individual task files conforming to `instructions/tasks.instructions.md`:

**File location**: `.plaesy/tasks/backlog/{priority}_{title}.md`
**Naming examples**:

- `.plaesy/tasks/backlog/critical_setup-auth-middleware.md`
- `.plaesy/tasks/backlog/high_implement-user-model.md`
- `.plaesy/tasks/backlog/medium_add-tests-for-api.md`

**File format** (must follow `instructions/tasks.instructions.md`):

```markdown

---
title: [title without priority prefix]
phase: [design|implement|assess|optimize|fix]
status: backlog
createdAt: [ISO timestamp]
updatedAt: [ISO timestamp]
---

## Description
Clear description of what needs to be done.

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## References
- [task-instructions](../../instructions/tasks.md)
- [other-related-tasks](./{status}/*.md)

## Dependencies
- [blocking-task](./{status}/*.md) [if blocked]
- External dependency: [description]

## Notes
- Implementation notes
- Constraints
- Related decisions
```

**For TDD-enforced tasks** (software implementation), follow `templates/tasks.template.md` pattern:

- Mark RED phase tests first (before implementation)
- Include exact file paths: `src/`, `tests/` directories
- Add `[P]` marker for parallel-executable tasks (different files, no dependencies)

### Step 5: Validate & Report

Before reporting complete:

1. **Format validation**:
  - Each task has required frontmatter (title, phase, status, createdAt, updatedAt)
  - All tasks use correct directory: `.plaesy/tasks/backlog/`
  - All filenames follow pattern: `{priority}_{title}.md`

2. **Content validation**:
  - Each task has Description + Acceptance Criteria (minimum)
  - Dependencies are valid cross-references to other `.plaesy/tasks/` files
  - TDD tasks mark RED phase test generation first

3. **Integration validation**:
  - Tasks reference `instructions/tasks.instructions.md` for format guide
  - Reference existing `.plaesy/memory/` files if available
  - Tasks conform to project phase (from constitution.md)

4. **Backlog quality validation** (every format, on the same task set):
  - **Duplicate detection**: flag any overlapping stories (2+ stories solving the same user need) — document the reason or merge them
  - **Ambiguity check**: if any acceptance criterion lacks an objective pass/fail, flag it "needs refinement"
  - **Estimates**: every user-story task carries a Fibonacci estimate, none left "TBD"

#### Output Format Branch

Build the task set once, then serialize it per `--format`:

- **`--format markdown`** (default) — one file per task in
  `.plaesy/tasks/backlog/{priority}_{title}.md`, exactly as written in Step 4.
  This is the primary contract and the shape `.plaesy/tasks/` expects.
- **`--format json`** — the *same* tasks serialized as one JSON index at `--out`
  (default `.plaesy/tasks/backlog/backlog.json`). It is a derived index of those
  per-task files, never a differently-shaped second backlog.
- **`--format both`** — the per-task Markdown files *and* the JSON index.

The hierarchy the JSON encodes (epic → story → task) is the same one the
per-task files express in sections. This layout is the readable reference when a
caller wants one document instead of many:

```markdown
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

JSON index shape (used by `--format json` and `both`):

```json
{
  "epics": [
    {
      "id": "EPIC-001",
      "name": "...",
      "value_proposition": "...",
      "story_ids": ["STORY-001"]
    }
  ],
  "stories": [
    {
      "id": "STORY-001",
      "epic_id": "EPIC-001",
      "narrative": "As a X I want Y so that Z",
      "acceptance_criteria": [{"given": "...", "when": "...", "then": "..."}],
      "estimate": 5,
      "priority": "high",
      "task_ids": ["TASK-001", "TEST-001"],
      "needs_refinement": false
    }
  ],
  "tasks": [
    {
      "id": "TASK-001",
      "story_id": "STORY-001",
      "path": ".plaesy/tasks/backlog/high_implement-user-model.md",
      "type": "implementation|test|documentation",
      "depends_on": []
    }
  ]
}
```

Report: task file count, output directory, frontmatter sample, epic/story/task
counts, estimated total story points, top 3 blockers/dependencies (if any),
ambiguities found (if any) + count of tasks marked "needs refinement", and
validation status. When `--format json` or `both` was used, also report the JSON
path. When the provider is unconfigured, the brief is returned instead and
`file_count` is 0.

## What a Good Task File Does

- **Generate structured `.plaesy/tasks/` files** — the format is what makes a task
  addressable by the rest of the workflow
- **Follow the format requirements in `instructions/tasks.instructions.md`**,
  frontmatter included
- **Write task files into `.plaesy/tasks/backlog/`** unless the user explicitly asks
  for another location
- **Prefix task filenames with priority** (`{priority}_{title}.md`) so the queue
  sorts by it
- **Split TDD RED/GREEN/REFACTOR into separate tasks** with an explicit dependency
  — one task cannot express a cycle
- **Write acceptance criteria that state an observable outcome** — "implement
  feature" names work, not a criterion
- **Make every acceptance criterion measurable** — "should look nice" cannot fail,
  so it cannot pass either
- **Flag every blocking relationship in the dependency map** — an unflagged
  blocker surfaces later as a stalled task
- **Extract the requirement into the task text** — "do what requirements.md says"
  defers the work to whoever reads it later
- **Write stories in the user's actual language from the requirements** rather than
  filling a generic template
- **Keep a story to 1-2 sentences** — past that it has stopped being a story
- **Capture non-functional requirements (performance, security) as constraint tags**
  so they survive into implementation
- **Order dependencies acyclically** — A→B→C→A cannot be scheduled, so break the
  cycle before it is written
- **Reserve story points for user story and feature tasks** — infrastructure and
  setup are sized differently
- **Let the estimate track real complexity across the Fibonacci range** —
  identical estimates carry no information for planning
- **Cross-reference `.plaesy/memory/design.md` personas** in the stories you
  generate, when that file exists
- **Run the ambiguity pass and flag what is incomplete** before shipping to
  `/implement` — an unflagged gap surfaces mid-build
- **Report only story IDs and task paths the task files actually have**
- **Derive `--format json` from the same task files** — a second, differently
  shaped backlog is a second source of truth
- **Version-append when a backlog file already exists** (e.g. `backlog-v2.json`)
  unless the user asked to replace it
- **Surface an API failure once and hand over the manual fallback** — a retry loop
  cannot recover an API failure
- **Derive from spec/code, or seed the template, at Step 0 when
  `.plaesy/memory/design.md` is absent**, labeled `ASSUMED` where seeded,
  rather than producing a generic backlog unlabeled

## Design-Spine & Context Integration

**Before generating tasks**, check for project context:

- `.plaesy/memory/design.md` — existing features, user personas, constraints
- `.plaesy/instructions/brandkit.md` — product positioning (affects feature
  prioritization); only installed when the project declares a brand kit, so if the
  file is missing, derive positioning from existing product docs/README,
  otherwise label it `ASSUMED — generic positioning` and record it
- `.plaesy/memory/constitution.md` — tech stack, team size, timeline (affects task scope/estimate)
- Existing `.plaesy/tasks/{backlog|todo}/*.md` — avoid duplicate tasks

If `.plaesy/memory/design.md` exists, additionally:

1. **Before generating**, extract existing user personas, feature list, and success metrics
2. **Align stories** to the personas it lists; use the same role names
3. **Cross-reference features**: flag any story implementing a feature already shipped (duplicate detection)
4. **Inherit constraints**: its non-functional requirements apply to all generated stories
5. **Hand back new personas/features** discovered during generation to `/improve:design`, which owns updates to that file — do not edit it from here

If generating tasks for a **feature within a larger project**, organize by feature:

- Create `.plaesy/tasks/backlog/epic_[feature-name]_{task}.md` for feature-related tasks
- Link to feature spec in `.plaesy/specs/[###-feature-name]/` if it exists

## Programmatic Invocation (Called By Other Prompts)

A prompt that needs a task backlog mid-run (`/implement`, `/doc`, `/assess`) calls this directly:

```text

CALL /create:tasks
  spec: <requirements document path or inline text>
  format: markdown | json | both
  out: .plaesy/tasks/backlog/            # directory for markdown, file path for json
  prioritize: true
RETURNS
  directory: .plaesy/tasks/backlog/
  file_count: <number of task files created, 0 if generation was skipped>
  files: [critical_task1.md, high_task2.md, …]
  index: <list of all task filenames with frontmatter summary>
  json_path: <written JSON index path, or null unless --format json|both>
  backlog: {epic_count, story_count, task_count, total_points}
  ambiguities: [<list of tasks/stories needing clarification>]
  brief: <composed specification, for reproducibility>
```

If no LLM provider is configured, the call returns `file_count: 0` with `brief`
populated — the calling prompt must treat this as "backlog pending, manual
decomposition needed" and report it, never silently drop the backlog or
fabricate paths.

## Success Criteria

A task backlog is complete when:

- ✅ All tasks use `.plaesy/tasks/backlog/{priority}_{title}.md` format
- ✅ Each task has frontmatter (title, phase, status, createdAt, updatedAt)
- ✅ Description is clear and concise
- ✅ Acceptance criteria are testable (not just checklist items)
- ✅ Epics clearly state business value + user need (2-3 sentences max)
- ✅ Each story follows the "As a / I want / So that" format (1-2 sentences)
- ✅ 2-5 acceptance criteria per story (Given-When-Then + checklist)
- ✅ All stories estimable (story points assigned, no "TBD")
- ✅ Dependencies mapped (blocking tasks identified, no cycles)
- ✅ TDD tasks mark RED phase tests first (if software implementation)
- ✅ References include `instructions/tasks.md` + project memory files
- ✅ Ambiguities flagged (incomplete specs called out, not hidden)
- ✅ Both Markdown and JSON versions written (when `--format both`); the JSON is an index of those same task files
- ✅ Output location reported — the per-task directory, and the JSON path when requested
- ✅ Report includes task count (e.g. "10 tasks generated: 2 critical, 4 high, 3 medium, 1 low") + epic/story counts, total points, and the ambiguity summary

## Context Preservation

Per research findings (ISO/IEC/IEEE 42010, BDD standards):

- Acceptance criteria reduce development rework by 25-30% when well-written
- LLM-based extraction maintains 95%+ accuracy on context preservation (Automated User Story Generation, arxiv.org/abs/2404.01558)
- Story Point estimates should use Fibonacci scale (1, 2, 3, 5, 8, 13) for consistency with Agile planning

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` → `instructions/tasks.instructions.md`
