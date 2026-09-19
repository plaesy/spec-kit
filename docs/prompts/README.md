# 🤖 Plaesy Spec-Kit Prompts Documentation

**Complete reference for all AI-optimized workflow prompts that power the Plaesy development automation framework.**

> **Before creating prompts**: reference `instructions/plaesy.instructions.md` (copied to `.plaesy/instructions/plaesy.md` on init) for constitutional rules.

---

## 🎯 Quick Navigation

| Prompt | Purpose | Execution |
|--------|---------|-----------|
| [`/start`](#start) | Project orchestration (generates constitution) | ultracode |
| [`/implement`](#implement) | TDD implementation + verifier pass | ultracode |
| [`/assess`](#assess) | Quality assessment + upfront research mode | ultracode |
| [`/optimize`](#optimize) | Performance optimization | ultracode |
| [`/improve`](#improve) | Best-practice gate across every active dimension | ultracode |
| [`/loop`](#loop) | Autonomous assess → fix → verify → repeat | ultracode |
| [`/fix`](#fix) | Issue resolution | standard |
| [`/continue`](#continue) | Resume & state detection | ultracode |
| [`/save`](#save) | Persist progress | standard |
| [`/doc`](#doc) | Documentation generation | ultracode |
| [`/generate:images`](#generateimages) | Generate a real image asset (UI/design) from a text prompt | standard |

Typical order for a new feature: `/start` → `/assess` (Mode 1: research, clarify,
design Spine if UI) → `/implement` → `/assess` (Mode 2) → `/optimize`/`/fix` as needed
→ `/save`/`/doc`.
`/continue` resumes this sequence from wherever it left off; it isn't a fixed slot in
the order.

> **Note**: technology/market/legal research **and** spec-ambiguity resolution are both
> **Mode 1 of `/assess`** (see `prompts/assess.md`) — run `/assess:{scope}` before a
> spec exists and it researches the web/Context7 for evidence and resolves vague
> requirements in the same pass, instead of splitting research and clarification into
> separate commands. This keeps one command per dimension, and keeps both
> citation-backed (anti-hallucination) rather than easier-to-skip separate steps.

---

## 📋 Complete Workflow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       PLAESY WORKFLOW                                   │
└─────────────────────────────────────────────────────────────────────────┘

START (/start) ← Generates project constitution (Phase 0)
   ↓
ASSESS (Mode 1) ← Technology/market validation + ambiguity resolution + Design Spine
   ↓                (production, if UI/architecture/etc. needs a design), upfront
IMPLEMENT (/implement) ← TDD: Red → Green → Refactor (90%+ coverage)
   ↓                    + independent verifier pass before "complete"
   ↓
QUALITY GATES (.plaesy/instructions/quality-gates.md)
   ↓
ASSESS (/assess) ← Comprehensive quality scoring + design audit
   ↓
OPTIMIZE (/optimize) ← Performance & UX optimization
   ↓
FINAL VALIDATION (.plaesy/instructions/quality-gates.md)
   ↓
FIX (/fix) ← [If needed] Issue resolution
   ↓
SAVE (/save) ← Persist progress & knowledge
   ↓
COMPLETION + RECOMMENDATIONS
```

---

## 🔄 Phase Details

### `/start`
**File**: `prompts/start.md`  
**Execution**: `ultracode` (parallel multi-agent)  
**Purpose**: Orchestrate complete project workflow automation

**Triggers**:
```bash
/start Build a real-time notification system for mobile apps
```

**What It Does**:
- Coordinates all phases sequentially
- Loads memory state from `.plaesy/memory/`
- Executes specialized prompts based on project type
- Tracks progress and handles state transitions
- Calls `@nara` for complex technical decisions

**Outputs**:
- Phase execution plan
- Progress tracking
- Handoff to next phase

**When to Use**: Starting new project, resuming with `/continue`

---


> **No standalone design command.** Design production and audit are folded into
> `/assess`'s **Design Spine** (see `prompts/assess.md` Mode 1 and Mode 2); redesign or
> refactor is handled by `/optimize --design`. The spine covers five dimensions —
> UI/UX, architecture, business model, org structure, process — with one pattern:
> components → tokens (no hardcoding) → states/edge cases → mandatory audit (hard
> stop: WCAG for UI, failure modes for architecture, unit economics for business
> model, sustainability for org, completeness for process) → handoff spec. Skip it for
> backend-only/API-only/CLI projects with no UI.
>
> ```bash
> /assess:design       # Mode 1: produce mockups, component library, design tokens
> /optimize --design   # Redesign/refactor when Mode 2's audit fails
> ```

---

### `/implement`
**File**: `prompts/implement.md`  
**Execution**: `ultracode`  
**Purpose**: Build production-ready code with TDD enforcement and 90%+ coverage

**Triggers**:
```bash
/implement Build authentication service with JWT tokens
/implement Create React component library from design specs
```

**What It Does**:
- Validates all specs clear, architecture defined, design specs available
- Implements TDD cycle: RED → GREEN → REFACTOR
- Enforces 90%+ test coverage (mandatory)
- Uses modern, Context7-validated technologies
- Implements design specs (components, tokens, accessibility)
- Applies OWASP security best practices

**Outputs**:
- Complete source code
- Comprehensive tests (90%+ coverage)
- Configuration files and deployment scripts
- Documentation with examples

**Constitutional Rules**:
- ✅ TDD REQUIRED (Red-Green-Refactor)
- ✅ 90% COVERAGE (minimum test coverage)
- ✅ MODERN TECH ONLY (Context7 validated)
- ✅ SECURITY-FIRST (OWASP compliant)
- ✅ DESIGN COMPLIANT (use design specs, tokens, no hardcode)

**When to Use**: After `/assess` (Mode 1, including Design Spine if UI), when ready to write code

---

### `/assess`
**File**: `prompts/assess.md`  
**Scoped commands**: `prompts/assess/{technical,design,business,marketing,legal,financial,management,product}.md`
— one file per dimension, each a thin wrapper that pins the scope and defers to
`prompts/assess.md` for the full protocol. This is what makes `/assess:design` (and
the other seven) resolve as their own recognized command instead of an argument to
`/assess`; the platform install step mirrors this same subfolder into each target
(e.g. `.claude/commands/assess/design.md`) automatically, since the mirror copy
already recurses through subdirectories.  
**Execution**: `ultracode`  
**Purpose**: Comprehensive quality assessment with design audit

**Triggers**:
```bash
/assess
/assess @design ./app  # Focus on design quality
/assess:design          # Same dimension, resolved as its own command
```

**What It Does**:
- Analyzes code quality, tests, security, documentation
- Optional: Audits design quality (if frontend/UI detected)
- Generates overall score 0-100 with per-category breakdown
- Lists top 5 findings with file:line citations
- Recommends next phase (optimize, fix, design, etc.)

**Scoring** (Backend/API):
- Code Quality: 30%
- Test Coverage: 25%
- Security: 20%
- Documentation: 15%
- Performance: 10%

**Scoring** (Frontend/UI - if detected):
- Code Quality: 25%
- Test Coverage: 20%
- Security: 15%
- Documentation: 15%
- **Design Quality: 15%** ← NEW
- Performance: 10%

**Design Quality Breakdown**:
- Component Consistency: 25%
- Accessibility (WCAG AA): 35% ← Hard stop
- Design Tokens: 20%
- Design System Adoption: 20%

**When to Use**: After `/implement`, before shipping

---

### `/optimize`
**File**: `prompts/optimize.md`  
**Execution**: `ultracode`  
**Purpose**: Performance and UX optimization

**Triggers**:
```bash
/optimize Improve API response time and reduce memory usage
/optimize Optimize component rendering and image loading
```

**What It Does**:
- Collects baseline metrics and identifies bottlenecks
- Applies optimizations incrementally with validation
- Measures improvements and prevents regressions

**Backend Priorities**:
1. Database (queries, indexes, connection pooling)
2. Caching (application, Redis, CDN)
3. Algorithms (complexity reduction)
4. I/O operations (async, batching)
5. Memory management (pooling, GC tuning)

**Frontend Priorities**:
1. Component performance (re-renders, memoization, lazy load)
2. Image optimization (WebP, responsive, lazy load)
3. Design system (increase component reuse)
4. Animation performance (60fps target)
5. Accessibility performance (screen reader optimization)

**When to Use**: After `/assess`, before shipping

---

### `/improve`
**File**: `prompts/improve.md`
**Execution**: `ultracode`
**Purpose**: Best-practice gate across **every active dimension** — not "is it
broken" (`/fix`), not "known gap against a target" (`/optimize`), not "what's the
score" (`/assess`), but "has practice moved on?"

**Triggers**:
```bash
/improve                # All dimensions active in the constitution
/improve:technical       # Tech stack / architecture / tooling currency
/improve:design           # Design-system / accessibility-pattern currency
/improve:business,legal,marketing,financial,management,product   # Same idea, that dimension
/improve:adoption        # Only unused Plaesy Spec-Kit capabilities + version drift
/improve:spec              # Only spec/plan/tasks quality + traceability
```

**What It Does**:
- No quality baseline yet → routes to `/assess` first
- Known performance/code-quality/audit gap → routes to `/optimize`
- Bug/error/regression/compliance breach → routes to `/fix`
- Everything else, per active dimension — tech stack behind current stable,
  design pattern superseded, pricing/business model outdated for the market,
  compliance approach below current norms, GTM practice stale, FinOps/process
  practice missing a now-standard step, roadmap framework outdated — handled
  directly here, every finding cited (same evidence discipline as `/assess`
  Mode 1: Context7 for technical, WebSearch/WebFetch for everything else)
- Plus two dimension-agnostic checks every run: framework adoption gaps/version
  drift, and spec traceability/ambiguity
- Never edits framework files, specs, or deliverables automatically — every
  recommendation needs per-item confirmation

**Two modes**: **Project Mode** (above) needs a constitution/dimensions to scope
against. **Artifact Mode** needs neither — point `/improve` at one file or link
(`/improve ./MOM.docx`, `/improve https://figma.com/file/...`, a `.pptx` deck, an
`.xlsx` sheet) and it improves just that deliverable against current convention for
its type (`word`/`powerpoint`/`excel`.instructions.md, or Figma MCP + design roles
for a Figma file), no spec/tasks/constitution required. If the artifact's type has
no matching instructions/template in the framework yet, that itself is reported —
as a framework gap, not a defect in the artifact.

**When to Use**: Periodically on a mature project, before a major feature, or any
time on a single deliverable (a doc, a deck, a design file) — instead of guessing
whether you need `/assess`, `/optimize`, `/fix`, or a best-practice pass

**Skip When**: You already know exactly which command you need (call it directly)

---

### `/loop`
**File**: `prompts/loop.md`  
**Execution**: `ultracode`  
**Purpose**: Autonomous assess → fix → verify → repeat, no user input between iterations

**Triggers**:
```bash
/loop
/loop --focus technical --quality-target 90
```

**What It Does**:
- Runs `/assess:technical`, filters auto-fixable vs. needs-review vs. blocked findings
- Implements auto-fixable fixes (parallel where independent), verifies each against
  `.plaesy/instructions/quality-gates.md`
- Repeats until quality threshold reached, all auto-fixable issues resolved, or an
  infinite-loop pattern is detected (same issue recurring 3+ iterations)

**When to Use**: After `/implement`, when you want unattended cleanup rather than
reviewing each fix manually

**Skip When**: Manual decisions are needed per fix, or changes must be reviewed before commit

---

### `/fix`
**File**: `prompts/fix.md`  
**Execution**: standard  
**Purpose**: Resolve issues and blockers

**Triggers**:
```bash
/fix tests are failing in auth module
/fix Design accessibility issues blocking WCAG compliance
```

**What It Does**:
- Diagnoses root causes of failures
- Applies targeted fixes
- Re-runs tests to validate resolution
- Documents what failed and how it was fixed

**When to Use**: When issues found in earlier phases

---

### `/continue`
**File**: `prompts/continue.md`  
**Execution**: `ultracode`  
**Purpose**: Resume project work and auto-detect current state

**Triggers**:
```bash
/continue
```

**What It Does**:
- Loads memory state from `.plaesy/memory/`
- Detects current project phase (research incomplete? design missing? etc.)
- Executes remaining phases automatically
- Detects design gaps for frontend projects
- Re-evaluates state after each phase
- Continues until all phases complete

**State Detection**:
1. No project started → `/start`
2. Research/specs/design incomplete (if UI) → `/assess` (Mode 1, includes Design Spine)
3. Implementation incomplete → `/implement`
4. Quality issues → `/assess` (Mode 2) → `/optimize` (includes redesign if Design Spine audit failed)
5. All complete → recommendations

**When to Use**: Resuming work after interruption

---

### `/save`
**File**: `prompts/save.md`  
**Execution**: standard  
**Purpose**: Persist progress and knowledge to project memory

**Triggers**:
```bash
/save
```

**What It Does**:
- Updates `.plaesy/context.md` (current session state)
- Updates `.plaesy/memory.md`
- Persists design decisions and rationale
- Records technical insights and architecture decisions
- Documents design system choices and accessibility decisions

**Outputs**:
- Context file (current session progress)
- Knowledge file (project-wide learnings)
- Ready for `/continue` in next session

**When to Use**: After each major phase or end of session

---

### `/doc`
**File**: `prompts/doc.md`  
**Execution**: `ultracode`  
**Purpose**: Generate comprehensive project documentation from code

**Triggers**:
```bash
/doc
/doc --output-types overview,architecture,reference
```

**What It Does**:
- Reads source code + project specs
- Generates reference documentation (API, functions, types)
- Creates architecture documentation with Mermaid diagrams
- Produces component specifications
- Generates how-to guides from actual usage patterns

**Outputs**:
- `.plaesy/memory/overview.md` - Project purpose and tech stack
- `.plaesy/memory/architecture.md` - System design + diagrams
- `.plaesy/memory/design.md` - UI/UX design
- `.plaesy/memory/components.md` - Component specifications
- `.plaesy/memory/reference.md` - API/function reference
- `.plaesy/memory/openapi.json` - OpenAPI specification (if API project)

**When to Use**: After implementation, for external documentation

---

## 🎓 Workflow Examples

### Example 1: New Web App with UI
```bash
/start Build an e-commerce product catalog with search and filters

# Workflow executes:
1. /assess (Mode 1: validate Next.js + PostgreSQL + React, define requirements,
   UI/UX success metrics, Design Spine mockups/component library/design tokens)
2. /implement (build components with TDD, 90%+ coverage, verifier pass)
3. /assess (code quality 90+, design quality 85+)
4. /optimize (improve component performance, image loading)
5. /save (persist progress)
```

### Example 2: Existing App Refactoring
```bash
/assess ./legacy-dashboard

# If design quality low:
/optimize --design Refactor dashboard for modern design system and accessibility

# Then:
/implement Apply design improvements (TDD)
/optimize Component performance and design system adoption
/assess Verify improvements
```

### Example 3: Resume Interrupted Work
```bash
/continue

# Auto-detects state:
# - If design/spec incomplete: continues with /assess (Mode 1)
# - Executes remaining phases until complete
```

---

## 📚 See Also

- **[plaesy.md](../../instructions/plaesy.instructions.md)** — Core guidelines for all prompts
- **[quality-gates.instructions.md](../../instructions/quality-gates.instructions.md)** — Quality gates
- **[error-recovery.instructions.md](../../instructions/error-recovery.instructions.md)** — Error recovery

---

## 🔗 Related Documentation

- [Scripts Documentation](../scripts/) — Automation scripts
- [Instructions Documentation](../instructions/) — Technology-specific guidance
- [Chat Modes Documentation](../chatmodes/) — AI role configurations
- [Templates Documentation](../templates/) — Project structure templates

---

**Last Updated**: 2026-09-15  
**Spec-Kit Version**: 0.0.1  
**Total Prompts**: 9 implemented (research, ambiguity-resolution, and design production/audit all folded into `/assess` Mode 1/2)
