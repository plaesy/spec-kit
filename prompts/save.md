---
description: "Save session state and update memory checkpoints"
subagent: true
---

# `/save` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective
Preserve session state with efficient memory management

## Why This Phase Matters

**Purpose**: Memory persistence captures learnings (reusable patterns across projects)

**Why Now**: After implementation, assessment, optimization, and error recovery—before context disappears, insights need preservation for future reference and cross-project knowledge transfer.

**What Breaks If Skipped**: 
- Context loss between sessions (must re-read files, rebuild mental model)
- No accumulated knowledge base (patterns learned once get re-discovered)
- Lost decisions and reasoning (audit trail disappears)
- Duplicate work in future projects (no reference to past solutions)

**Success Enables**:
- Faster project resumption (read context.md, continue from checkpoint)
- Reusable patterns across projects (searchable memory system)
- Institutional knowledge preservation (decisions documented)
- Improved estimation (historical patterns inform future planning)

## Protocol
**Preserve session state with efficient memory management using Plaesy memory system**

## Validation Checklist (Before Running)
- ✅ Context file exists (.plaesy/context.md)
- ✅ Memory file exists (.plaesy/memory.md)
- ✅ Size limits trackable (current line counts accessible)
- ✅ No external references in memory paths

## Memory Management Workflow

**Follow**: `.plaesy/instructions/plaesy.md` → Memory Management + Anti-Duplication Protocol (authoritative)

Core workflow:
1. **Size check** - Context ≤100 lines (see plaesy.md)
2. **Update files** - `.plaesy/context.md` + `.plaesy/memory.md`
3. **Validate** - Check limits + integrity

## Task Persistence

**Follow**: `.plaesy/instructions/tasks.md` → Task Management Rules

Move incomplete tasks from `doing/` to `todo/` if acceptance criteria unmet.
Update "## In-Flight Tasks" section in `.plaesy/context.md` (keep under 100 lines).

## Save Process
1. **Validate files exist** - context.md, memory.md
2. **Check sizes** - context ≤100 lines, memory.md as index
3. **Direct write** - Update `.plaesy/` files (no backup needed)
4. **Validate self-containment** - All content in `.plaesy/` (no external refs)
5. **Archive if needed** - Move details to `.plaesy/memory/` topic files

## Recovery Instructions
**To resume**:
1. Read `.plaesy/context.md` for current session
2. Read `.plaesy/memory.md` for project knowledge index, follow links to `.plaesy/memory/{{topics}}.md`
3. Continue from "Next Steps"

## Success Criteria (VERIFY Before Completing)

✅ **Clarity**:
- [ ] All specifications clear and documented
- [ ] No ambiguities remain
- [ ] Session decisions captured in context.md

✅ **Quality**:
- [ ] Tests passing/metrics validated
- [ ] No regressions detected
- [ ] Implementation matches specifications

✅ **Self-Audit**:
- [ ] Ready for next phase?
- [ ] Would ship this to production?
- [ ] Memory system synchronized (context.md ≤100 lines)
- [ ] Knowledge index updated (.plaesy/memory.md points to all topic files)

## Success Format
```
✅ Context saved (X/100 lines) - Quality: Optimal
✅ Knowledge updated - Quality: Comprehensive
✅ Memory system synchronized - Performance: Efficient
Next: [next steps]

Note: Use /assess to refresh project understanding after context changes
```

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
