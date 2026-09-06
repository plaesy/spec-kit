---
description: "Orchestrate complete project workflow automation from idea to production"
---

# `/start` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective
Transform project description → production-ready code through modular orchestration

## Protocol
**Execute project automation by coordinating specialized prompts**
- Load memory state from `.plaesy/memory/`
- Execute phases using dedicated prompts
- Track progress and handle state transitions

## 🔄 Phase Orchestration

**For complete phase definitions, see**: `.plaesy/memory/plaesy.md` → Workflow Phases (single source of truth)

This command orchestrates all 9 phases sequentially:
1. `/assess` (Research Mode) - Technology validation & exploration
2. `/implement` - Universal orchestrator (code + design + tools via instruction loading)
3. quality-gates - Automated validation
4. `/assess` (Assessment Mode) - Quality measurement (MANDATORY)
5. `/optimize` - Performance/design/code improvements
6. `/assess` (Verification Mode) - Verification (MANDATORY)
7. quality-gates - Final validation
8. `/fix` - Bug resolution (if needed)
9. `/save` - Session persistence

**Note**: `/assess` is unified validator with 3 modes: research (upfront), assessment (post-impl), verification (post-optimize)

**Phases in detail**: See `.plaesy/memory/plaesy.md` for full phase descriptions, timing, prerequisites, and execution patterns.

### Initialize Task System

Create project task infrastructure:

**Create directory structure**:
```
.plaesy/tasks/
├── backlog/           # Ideas, features, bugs not yet scheduled
├── todo/              # Scheduled for current phase
├── doing/             # In active work
├── done/              # Completed and validated
└── blocked/           # Blocked on external factors or dependencies
```

**Seed initial tasks from project description** (optional):
1. Extract key work items from requirements / architecture specifications
2. Create task files in `.plaesy/tasks/backlog/`
3. Each task includes YAML frontmatter defining Status, Phase, Description, Acceptance Criteria, Priority

**Integration across phases**:
- `/continue` auto-detects tasks in `todo/` and `doing/`, can run in a loop for autonomous execution
- `/save` reports task counts per status and phase to `context.md`

## 🎯 Decision Advisor

### `@nara` - Decision Oracle (Available Anytime)
Call **`@nara`** from any phase when you need to make an important decision:
- Technology choices (tool X vs Y?)
- Architecture decisions (monolith vs services?)
- Refactoring vs shipping (now or later?)
- Priority conflicts (feature A vs B?)
- Approach disagreements (pattern X vs pattern Y?)

Nara gathers perspectives from relevant experts and **makes the decision** for you with clear rationale and implementation plan.

## 🎯 Execution Rules

### Quality Standards
- Test Coverage: Minimum 90%
- Performance: <200ms response time
- Security: Zero vulnerabilities
- Documentation: Complete for all deliverables

### Error Handling
- Follow `.plaesy/memory/error-recovery.md` protocol
- Retry with alternatives on failures
- Document all issues and solutions

## 📊 Progress Tracking

Format:
```
Phase: [current/11] | Quality: [score/100] | Memory: [usage%]
Status: [working/completed/failed] | Next: [next phase]
```

---

**Execute this orchestrator with `/start` command!**
