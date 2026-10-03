---
description: "Save session state and update memory checkpoints"
subagent: true
---

# `/save` command instructions

⚡ **Run with**: any parallel-capable mode — a host that provides
subagents, or a mode that does. Not a requirement: a single-agent host
runs this identically, just serially. `/loop` documents when fan-out is
worth its cost.

## Objective

Preserve session state with efficient memory management

## Why This Phase Matters

**Purpose**: Memory persistence captures learnings (reusable patterns across projects)

**Why Now**: After implementation, assessment, optimization, and error recovery — before context disappears, insights need preserving for future reference and cross-project knowledge transfer.

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

Save session state and update memory checkpoints per the Plaesy memory system.
**Follow**: `.plaesy/instructions/plaesy.md` → Memory Management + Anti-Duplication
Protocol (authoritative) for the size/structure rules this applies;
`.plaesy/instructions/tasks.md` → Task Management Rules for step 7.

1. **Validate** — confirm `.plaesy/context.md`, `.plaesy/memory.md`, and
   `.plaesy/decisions.md` exist (no backup needed — this is a direct write)
2. **Size check** — enforce `context.md ≤ 100 lines` (archive overflow to `.plaesy/memory/`)
3. **Learning something** — check `.plaesy/memory.md`, create or update `.plaesy/memory/{topic}.md`
4. **Update context** — write current task, checklists, next steps to `context.md`
5. **Update memory index** — refresh `.plaesy/memory.md` to point to all topic files in `.plaesy/memory/`
6. **Record decisions** — if a decision was made, create or update `.plaesy/decisions/{topic}.md` with reference links and access timestamps, then refresh `.plaesy/decisions.md`
7. **Persist tasks** — move incomplete `doing/` tasks to `todo/` if acceptance criteria unmet; update "## In-Flight Tasks" in `context.md`
8. **Validate self-containment** — all content under `.plaesy/`, no external refs
9. **Report** — emit Completion Format below

## Resuming Later

Resuming from what `/save` just wrote is `/continue`'s job — it loads
`context.md`/`memory.md`/`decisions.md`, detects the current phase, and
executes the remaining workflow. Not restated here; see `/continue`.

## Success Criteria (VERIFY Before Completing)

✅ **Clarity**:

- [ ] All specifications clear and documented
- [ ] No ambiguities remain
- [ ] Session decisions captured in `.plaesy/decisions.md` and its topic files

✅ **Quality**:

- [ ] Implementation matches specifications
- [ ] Gates from the phase just saved are green, as reported by that phase —
      `/save` does not re-run them and does not re-verify them

✅ **Self-Audit**:

- [ ] Closing check applied (`.plaesy/instructions/quality-gates.md`
      → **Closing check**): ready for the next phase, and would this ship?
- [ ] Memory system synchronized (context.md ≤100 lines)
- [ ] Knowledge index updated (.plaesy/memory.md points to all topic files)
- [ ] Decisions index updated (.plaesy/decisions.md points to all decision topic files)

## Completion Format

```text

Next: [next steps]

Note: Use /assess to refresh project understanding after context changes
```

---

## Worked Example — Save After Full Workflow Completion

**Input**: All 9 phases complete. `.plaesy/context.md`.
tasks in `doing/` with unmet acceptance criteria. User runs `/save`.

```text

Next: [none — project complete; /improve if improvements remain]

Actions taken:
- Moved doing/ tasks → todo/ (acceptance criteria unmet)
- Updated "## In-Flight Tasks" in context.md (now 8 lines)
- Refreshed memory.md index
- Validated self-containment
```

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
