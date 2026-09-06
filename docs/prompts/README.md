# 🤖 Plaesy Spec-Kit Prompts Documentation

**Complete reference for all AI-optimized workflow prompts that power the Plaesy development automation framework.**

> **Before creating prompts**: Read [AUTHORING.md](./AUTHORING.md) for design guidelines, and reference `.plaesy/memory/plaesy.md` for constitutional rules.

---

## 🎯 Quick Navigation

| Phase | Prompt | Purpose | Execution |
|-------|--------|---------|-----------|
| **1** | [`/start`](#phase-1-start) | Project orchestration | ultracode |
| **2** | [`/research`](#phase-2-research) | Technology validation | ultracode |
| **3** | [`/clarify`](#phase-3-clarify) | Specification resolution | ultracode |
| **4** | [`/design`](#phase-4-design) | UI/UX design generation | ultracode |
| **5** | [`/flow`](#phase-5-flow) | Architecture diagrams | ultracode |
| **6** | [`/implement`](#phase-6-implement) | TDD implementation | ultracode |
| **7** | [`/assess`](#phase-7-assess) | Quality assessment | ultracode |
| **8** | [`/optimize`](#phase-8-optimize) | Performance optimization | ultracode |
| **9** | [`/fix`](#phase-9-fix) | Issue resolution | standard |
| **10** | [`/continue`](#phase-10-continue) | Resume & state detection | ultracode |
| **11** | [`/save`](#phase-11-save) | Persist progress | standard |
| **12** | [`/doc`](#phase-12-doc) | Documentation generation | ultracode |

---

## 📋 Complete Workflow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    PLAESY 11-PHASE WORKFLOW                             │
└─────────────────────────────────────────────────────────────────────────┘

START (/start)
   ↓
RESEARCH (/research) ← Technology validation & requirements analysis
   ↓
CLARIFY (/clarify) ← Resolve ambiguities, generate specs
   ↓
DESIGN (/design) ← [NEW] Generate UI/UX, components, design tokens
   ↓                (Optional: skip for backend-only projects)
FLOW (/flow) ← Architecture diagrams, C4 model, workflows
   ↓
IMPLEMENT (/implement) ← TDD: Red → Green → Refactor (90%+ coverage)
   ↓
QUALITY GATES (shared/quality-gates.md)
   ↓
ASSESS (/assess) ← Comprehensive quality scoring + design audit
   ↓
OPTIMIZE (/optimize) ← Performance & UX optimization
   ↓
FINAL VALIDATION (shared/quality-gates.md)
   ↓
FIX (/fix) ← [If needed] Issue resolution
   ↓
SAVE (/save) ← Persist progress & knowledge
   ↓
COMPLETION + RECOMMENDATIONS
```

---

## 🔄 Phase Details

### Phase 1: START
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

### Phase 2: RESEARCH
**File**: `prompts/research.md`  
**Execution**: `ultracode`  
**Purpose**: Validate technology choices and resolve uncertainties

**Triggers**:
```bash
/research Validate PostgreSQL + Redis + Next.js stack for this project
```

**What It Does**:
- Detects TBD and placeholder requirements
- Uses Context7 to validate modern technologies
- Compares options with supporting evidence
- Identifies deprecated libraries and security issues
- Generates implementation-ready recommendations

**Outputs**:
- Technology recommendations with evidence
- Security/performance/complexity analysis
- Migration paths for issues found
- Complete findings documented

**When to Use**: Before `/clarify`, when technology choices are unclear

---

### Phase 3: CLARIFY
**File**: `prompts/clarify.md`  
**Execution**: `ultracode`  
**Purpose**: Resolve specification ambiguities to 100% implementation readiness

**Triggers**:
```bash
/clarify Clarify API design: REST vs GraphQL, auth strategy, error handling
```

**What It Does**:
- Detects vague requirements ("improve performance", "be scalable")
- Asks user for ALL technical and business decisions
- Validates choices against Context7 best practices
- Generates measurable acceptance criteria
- Includes UI/UX requirements (design system, accessibility, dark mode)

**Outputs**:
- Complete specification with no ambiguities
- All decisions documented with rationale
- Measurable success criteria
- Ready for `/design` phase (if frontend)

**When to Use**: Before architecture/design, when specs have gaps

---

### Phase 4: DESIGN ⭐ **NEW**
**File**: `prompts/design.md`  
**Execution**: `ultracode`  
**Purpose**: Generate production-ready UI/UX designs and component specifications

**Triggers**:
```bash
/design Create payment checkout flow mockups with component library
/design Refactor existing dashboard for accessibility & design system adoption
/design Build design system for React + Vue component libraries
```

**What It Does**:
- Analyzes project type: prototype, refactor, or design system
- Generates high-fidelity mockups or design improvements
- Creates component inventory with variants & states
- Defines design tokens (colors, typography, spacing)
- Audits WCAG 2.1 AA accessibility compliance
- Sets up Figma integration with Code Connect

**Outputs**:
- Figma file (or SVG/HTML mockups)
- Design specification document
- Component catalog with all variants
- Accessibility checklist (WCAG 2.1 AA)
- Handoff guide for engineers

**Constitutional Rules**:
- ✅ ACCESSIBILITY-FIRST (WCAG 2.1 AA minimum)
- ✅ COMPONENT-DRIVEN (all UI is reusable)
- ✅ DESIGN-TOKENS (no hardcoded values)
- ✅ MOBILE-FIRST (mobile first, scale to desktop)
- ✅ DARK MODE READY (light + dark support)
- ✅ MEASURABLE UX (clear success metrics)

**When to Use**: After `/clarify`, before `/flow` (for frontend projects)

**Skip When**: Backend-only, API-only, or CLI projects with no UI

---

### Phase 5: FLOW
**File**: `prompts/flow.md`  
**Execution**: `ultracode`  
**Purpose**: Generate architecture documentation and workflow diagrams

**Triggers**:
```bash
/flow Generate C4 architecture and deployment diagrams
```

**What It Does**:
- Analyzes project structure and technology stack
- Generates C4 model (Levels 1-3: System → Container → Component)
- Creates workflow diagrams (data flow, API interactions, user journeys)
- Documents deployment process and CI/CD
- Includes user journey diagrams (references `/design` specs if available)

**Outputs**:
- C4 architecture diagrams (System Context, Containers, Components)
- Data flow and API interaction diagrams
- User journey diagrams (if frontend)
- Development and deployment workflows
- Master index linking all diagrams

**When to Use**: After `/design` (or `/clarify` if no UI), before `/implement`

---

### Phase 6: IMPLEMENT
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

**When to Use**: After `/flow`, when ready to write code

---

### Phase 7: ASSESS
**File**: `prompts/assess.md`  
**Execution**: `ultracode`  
**Purpose**: Comprehensive quality assessment with design audit

**Triggers**:
```bash
/assess
/assess @design ./app  # Focus on design quality
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

### Phase 8: OPTIMIZE
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

### Phase 9: FIX
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

### Phase 10: CONTINUE
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
2. Research incomplete → `/research`
3. Specs incomplete → `/clarify`
4. Design missing (if UI) → `/design`
5. Missing architecture → `/flow`
6. Implementation incomplete → `/implement`
7. Quality issues → `/assess` → `/optimize`
8. Design quality issues → `/design` refactor
9. All complete → recommendations

**When to Use**: Resuming work after interruption

---

### Phase 11: SAVE
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

### Phase 12: DOC
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
- `overview.md` - Project purpose and tech stack
- `architecture.md` - System design + diagrams
- `components.md` - Component specifications
- `reference.md` - API/function reference
- `openapi.json` - OpenAPI specification (if API project)

**When to Use**: After implementation, for external documentation

---

## 🎓 Workflow Examples

### Example 1: New Web App with UI
```bash
/start Build an e-commerce product catalog with search and filters

# Workflow executes:
1. /research (validate Next.js + PostgreSQL + React)
2. /clarify (define requirements, UI/UX success metrics)
3. /design (create mockups, component library, design tokens)
4. /flow (architecture, user flows, data flow)
5. /implement (build components with TDD, 90%+ coverage)
6. /assess (code quality 90+, design quality 85+)
7. /optimize (improve component performance, image loading)
8. /save (persist progress)
```

### Example 2: Existing App Refactoring
```bash
/assess ./legacy-dashboard

# If design quality low:
/design Refactor dashboard for modern design system and accessibility

# Then:
/implement Apply design improvements (TDD)
/optimize Component performance and design system adoption
/assess Verify improvements
```

### Example 3: Resume Interrupted Work
```bash
/continue

# Auto-detects state:
# - If at design phase: continues from design
# - If design missing but frontend: suggests /design
# - Executes remaining phases until complete
```

---

## 📚 See Also

- **[plaesy.md](../instructions/plaesy.instructions.md)** — Core guidelines for all prompts
- **[design.md](./design.md)** — Detailed design phase documentation
- **[shared protocols](../shared/)** — Quality gates and error recovery

---

## 🔗 Related Documentation

- [Scripts Documentation](../scripts/) — Automation scripts
- [Instructions Documentation](../instructions/) — Technology-specific guidance
- [Chat Modes Documentation](../chatmodes/) — AI role configurations
- [Templates Documentation](../templates/) — Project structure templates

---

**Last Updated**: 2026-09-06  
**Spec-Kit Version**: 0.0.1  
**Total Prompts**: 12 (including new `/design` phase)
