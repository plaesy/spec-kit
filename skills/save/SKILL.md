---
name: save
description: "Save session state and update memory checkpoints Use when the user types /save."
disable-model-invocation: true
---

# `/save` command instructions

⚡ **Run with**: any parallel-capable mode. A single-agent host runs it serially, which is fine. See `/loop` for when fan-out is worth its cost.

## Objective

Persist session state and learnings, with memory kept small and well organized.

**Why it matters:** the work done in a session (implementation, assessment, optimization, recovery)
is lost once the context ends. Saving captures it. Without this step, a session must re-read
everything to resume, patterns get rediscovered, and decisions and their reasoning disappear.

With it, resuming is quick (read `context.md` and continue from the checkpoint), decisions stay on
record, and later projects can reuse past solutions.

## Success Contract

Write `contracts.json` in the task folder before work starts (one `run` and `expected` per item). Run `plaesy tasks validate <task>`. Report completion only when every item is `passed` (NN-08; `.plaesy/instructions/success-contract.md`).

## Protocol

Follow `.plaesy/instructions/plaesy.md` → Memory Management and Anti-Duplication (authoritative for size and structure). For step 7, follow `.plaesy/instructions/tasks.md` → Task Management Rules.

1. **Validate.** Confirm `.plaesy/context.md`, `.plaesy/memory.md`, and `.plaesy/decisions.md` exist. This is a direct write, so no backup is needed.
2. **Size check.** Keep `context.md` at 100 lines or fewer. Move overflow to `.plaesy/memory/`.
3. **New learning.** Check `.plaesy/memory.md`. Create or update `.plaesy/memory/{topic}.md`.
4. **Context.** Write the current task, checklists, and next steps to `context.md`.
5. **Memory index.** Refresh `.plaesy/memory.md` so it points to every topic file.
6. **Decisions.** If one was made, create or update `.plaesy/decisions/{topic}.md` with reference links and access timestamps. Then refresh `.plaesy/decisions.md`.
7. **Self-containment.** All content lives under `.plaesy/`. No external references.
8. **Report** (Completion Format).

Resuming from what was saved is `/continue`'s job. It loads `context.md`, `memory.md`, and `decisions.md`, detects the phase, and continues.

## Success Criteria

- [ ] Every Success Contract item is compared with its result at the gate. Blocking items are met.

**Clarity:** specifications are clear, with no open ambiguity. Session decisions are in `.plaesy/decisions.md` and its topic files.

**Quality:** the implementation matches the specifications. The gates of the phase just saved are green, as that phase reported them. `/save` does not re-run or re-verify them.

**Self-audit:**

- [ ] Closing check applied (`.plaesy/instructions/quality-gates.md` → **Closing check**): ready for the next phase, and would this ship?
- [ ] `context.md` is at most 100 lines.
- [ ] `.plaesy/memory.md` points to every topic file.
- [ ] `.plaesy/decisions.md` points to every decision topic file.

## Completion Format

```text
Next: [next steps]
Note: run /assess to refresh the project's understanding after context changes.
```

---

## Worked Example — Save After the Full Workflow

**Input:** all nine phases are complete. Three tasks in `doing/` have unmet acceptance criteria.

```text
Next: none. Project complete. Run /improve if improvements remain.

Actions:
- Moved doing/ tasks → todo/ (acceptance criteria unmet)
- Updated "## In-Flight Tasks" in context.md (now 8 lines)
- Refreshed memory.md index
- Validated self-containment
```

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
