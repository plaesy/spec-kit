---
description: "Global mandatory instructions"
applyTo: "**/*"
---

# Global Mandatory Instruction

> **See also**: `.plaesy/memory/tasks.md` (tasks instructions)

## Core Principles

### 1. Assistant Behavior (Highest Priority)
1. **Be factual** — cite sources or say "I don't know"
2. **Ask when unsure** — one clarifying question, don't guess
3. **Don't invent** — label assumptions, confirm before acting
4. **Stay focused** — no tangents unless asked
5. **Preserve context** — cite files, summarize on resume
6. **Minimal changes** — test, report results
7. **Be concise** — short output, clear next steps
8. **Protect secrets** — flag credentials, ask before proceeding
9. **No duplication** — check existing code before creating new (see Anti-Duplication)

### 2. Context Handling
- Consult repo, `.plaesy/context.md`, and `.plaesy/memory.md` when available
- 1-2 sentence context summary for multi-message tasks

---

## Anti-Duplication Protocol (Mandatory)

**Default: always modify/extend existing code, never duplicate** — unless user explicitly asks for something new.

### Before Creating New File/Function

1. **Search existing**: `find . -name "*.{js,py,ts,java,go,rs,cs,cpp,c,dart,md,etc}"`, `grep -r "similar purpose"`
2. **Match found?** → Default to modifying it (add param, extend return, keep compatibility)
3. **Can't be extended?** → Ask user, present match(es): "Found [name] with similar purpose. Default: modify. Proceed, or create new (explain why)?"

### Consistency Rule
- 1 function = 1 purpose, no overlap
- Naming and structure folder follows best practice patterns (check `.plaesy/memory.md`, if not exist use skills, if skills not exist use find-skills, if not exists research via context7 or internet. after that save into `.plaesy/memory/[topics].md` and update index memory on `.plaesy/memory.md`)

### Decision Tree
```
Need: Get user data
├── Existing getUserData()?
│   ├── Yes → **Modify existing**
└── └── No → Create new
```

---

## Memory Management

### Structure
```
.plaesy/
├── context.md                # Current session only (≤100 lines)
├── memory.md                 # INDEX/TOC for memory & knowledge
├── memory/                   # Knowledge files (flat, no subfolders)
│   └── [topics].md           # Other memory topics
└── analysis/                 # Analysis outputs (separate)
    ├── project.graph.json    # From plaesy-graph
    ├── project.html
    ├── reports.md
    └── token-stats.json      # From plaesy-trim
```

### Content Rules
- **context.md**: Current task, recent decisions (≤5), next steps (session-only)
- **memory.md**: Index/TOC only — links to memory/*.md files
  - **Pattern**: `- [Title](file.md) — one-line description`
  - **Example**: `- [Project Overview](overview.md) — Complete framework summary, architecture, component status`
  - **Rules**: 
    - Each line = one memory file (filename matches `name:` in frontmatter)
    - Keep descriptions under 100 chars
    - Group related entries by category (e.g., ## 🎯 Core Reference, ## 📋 Guidance)
    - No nested bullets or sub-links (flat list only)
    - Link must point to actual file that exists
- **memory/*.md**: Detailed explanations, examples, guides by topic (flat structure)

### Size Enforcement (Mandatory)
- **context.md**: Max 100 lines (archive to memory/ if needed)
- **memory.md**: Keep as index only
- **Individual memory/*.md**: Can be substantial, organize by topic
- Check size before write: `wc -l .plaesy/context.md`
- Move completed sessions to memory/ archive as needed

---

## Token & Output Efficiency

### Decision Ladder (Before Any New Code)
Stop at first rung that applies:
1. Needed at all? (no → skip, YAGNI)
2. Already in codebase? (reuse, don't rewrite)
3. In stdlib? (use it)
4. Native platform feature? (use it)
5. Already-installed dependency? (use it)
6. One-liner? (write it)
7. Otherwise → smallest working solution, no speculation

### Command Output Compression
Use Plaesy Trim to compress noisy commands (git, npm, docker, tests). Compresses formatting; never modifies code/data. Reference: `.plaesy/memory/plaesy-trim.md`

**Default behavior** (no opt-in): Route noisy commands through trim without asking. Never fake/omit results — only reformats real output, errors still surface. Skip wrapping for short one-shot commands.

---

## Context Window Management

Context = conversation + files + tools; limits vary by model.

### Before Large Operations
- Estimate token usage
- Warn if near limit
- Suggest `/compact` or smaller tasks

### During Operations
- Be concise
- Use Agent tool for heavy lifting
- Read only needed sections

### When Context <20% Remaining
- Save state to `.plaesy/context.md`
- If learn something new, save to `.plaesy/memory/[topics].md`, if file is new, update `.plaesy/memory.md` for reference into that file.
- Run prompt `/compact` or start new conversation
- Never continue context-heavy work in low-context

### Resume Pattern
- Read `.plaesy/context.md` on resume
- Continue from checkpoint

---

## Workflow Phases

The Plaesy framework executes work through 8 sequential phases. Each phase has a dedicated command (`/start` orchestrates all; individual commands available for resume/override).

### Phase 1: Universal Research & Validation (`/assess` - Research Mode)
**Purpose**: Validate ANY project aspect upfront - technology, business model, market fit, product strategy, legal compliance, financial viability, operations readiness, etc.
**Output**: Research + validation across requested dimensions (ranked options, tradeoffs, recommendations)
**When to run**: Start of project OR when clarity needed on ANY dimension (tech, business, product, marketing, legal, financial, operations)
**Prerequisites**: None (first phase)
**Note**: `/assess` in RESEARCH mode - universal validator for any project aspect
**Flexibility**: Can assess one dimension or multiple simultaneously

### Phase 2: Implementation (`/implement`) - Universal Orchestrator
**Purpose**: Execute all implementation work (code, design, tools) by loading appropriate instruction files
**Smart Orchestration**:
- Detects project type, tech stack, tool requirements
- Loads appropriate technology instructions (Go, Java, React, Python, etc.)
- Loads appropriate tool instructions (Figma, Terraform, Docker, SQL, etc.)
- Executes per loaded instructions (TDD, security, design, code, etc.)

**Output**: Complete implementation (code + design specs + configs + 90%+ test coverage)
**When to run**: After research confirms tech stack, ready to build
**Prerequisites**: Tech stack clear, specs clear, TDD environment ready
**Approach**: Smart instruction-loading + Red-Green-Refactor cycle, commit per instruction phase

### Phase 3: Quality Assurance (quality-gates.md)
**Purpose**: Validate implementation meets standards (automated check)
**Output**: Pass/fail gate, blocking issues identified
**When to run**: Automatic after `/implement`
**Prerequisites**: Implementation complete

### Phase 4: Comprehensive Assessment (`/assess`) - MANDATORY
**Purpose**: Multi-dimensional quality & viability assessment AFTER implementation
**Can assess ANY dimension** (or combination):
- ✅ **Technical**: Code quality, tests, security, performance, architecture, documentation
- ✅ **Design**: UI/UX, accessibility (WCAG AA), design system, consistency
- ✅ **Business**: Market fit, business model viability, competitive positioning, revenue potential
- ✅ **Product**: Feature completeness, user need fit, competitive advantage
- ✅ **Marketing**: Positioning, messaging, audience fit, GTM readiness
- ✅ **Legal/Compliance**: Regulatory compliance, data privacy, risk assessment
- ✅ **Financial**: Cost structure, pricing, profitability, unit economics
- ✅ **Operations**: Team capacity, deployment readiness, scalability

**Output**: Multi-dimensional quality scores (0-100 each) + detailed findings + research-validated recommendations
**When to run**: MANDATORY after `/implement`, before any optimization
**Blocks**: Cannot proceed to `/optimize` without running assess
**Mode**: Assessment mode - validates implementation quality across requested dimensions
**Findings** feed into next loop: Assessment gaps → Task backlog → `/assess` (research mode) validation → `/implement` fixes → assess again

### Phase 5: Comprehensive Optimization (`/optimize`)
**Purpose**: Optimize performance, design quality, and code efficiency
**Optimizes ALL**:
- ✅ Performance (response time, throughput, memory, caching, algorithms)
- ✅ Design quality (if UI: component reuse, tokens, dark mode, accessibility, animations)
- ✅ Code quality (refactoring, duplication, type safety, error handling, testing, docs)

**Output**: Optimized code + performance metrics + design improvements + refactored code
**When to run**: After `/assess` finds improvement opportunities
**Prerequisites**: Baseline metrics from `/assess`
**MANDATORY NEXT**: Must run `/assess` after to verify improvements + no regressions

### Phase 6: Verification Assessment (`/assess`) - MANDATORY After Optimize
**Purpose**: Verify optimization improvements worked + detect any regressions
**MUST verify**: 
- ✅ Performance improvements (% gains measured)
- ✅ Design improvements applied (if UI)
- ✅ Code quality improvements (refactoring complete)
- ✅ No regressions (tests still passing)
- ✅ All quality dimensions still within acceptable range

**Output**: Verification report + confidence score (0-100)
**When to run**: MANDATORY immediately after `/optimize`
**Blocks**: Cannot proceed to `/fix` or `/doc` without verification

### Phase 7: Error Recovery (`/fix`)
**Purpose**: Resolve bugs with root cause analysis
**Output**: Fixed code + root causes documented
**When to run**: When issues found in assessment
**Triggers**: Failed quality gates OR `/assess` findings

### Phase 8: Documentation (`/doc`, OPTIONAL)
**Purpose**: Generate professional documentation from code + specs
**Output**: README + API docs + guides + architecture narratives
**When to run**: Before release or as team reference
**Skip when**: Backend-only projects with minimal public API
**Prerequisites**: Code complete + tests passing

### Phase 9: Session Management (`/save`)
**Purpose**: Persist session progress + update memory
**Output**: Archived decisions + memory updates + context snapshot
**When to run**: End of session OR checkpoint
**Prerequisites**: None (can run anytime)

---

### Execution Patterns

**New project** (`/start`): Execute all phases sequentially (1 → 9)

**Resume session** (`/continue`): Auto-detect current phase → execute remaining (e.g., 4 → 9)

**Assessment Loop Pattern** (recommended): Continuous improvement via assess validation:
```bash
/continue           # Execute implementation phases
# Outputs code → quality-gates → /assess (assessment mode) validates quality
# /assess findings:
#   If design issues found:
#     → /implement (refactor with design focus)
#     → /assess (verify improvements)
#   Else if performance issues:
#     → /optimize (improve performance + design + code)
#     → /assess (verify improvements)
#   Else if bugs:
#     → /fix (resolve)
#     → /assess (verify fixes)
```

**Agile continuous improvement** (loop):
```bash
/continue           # Execute remaining phases + /assess (assessment mode)
# /assess results → create new tasks in backlog
/continue           # Resume with new tasks
# /assess (research mode) validates → /implement → /assess (assessment mode) → [optimize/fix] → /assess loop
```

**Autonomous execution**: Use `/loop` to continuously process tasks:
```bash
/loop /continue     # Keep running /continue until no more tasks
```

**Selective phases**: Run assess in different modes:
```bash
/assess             # Research mode: Validate technology choices upfront
/assess             # Assessment mode: Measure quality after implementation
/optimize           # Performance/design/code optimization on-demand
```