# Autonomous Loop: Continuous Idea → Implementation Cycle

Run the idea → implement → evaluate → next idea loop repeatedly, staying inside explicit governance limits.

---

## Quick Start

### 1. Start the Loop
```bash
claude /loop /evolve
```
This enters **self-paced mode**: `/evolve` runs one iteration, waits for the next automatic trigger.

Or, run manually:
```bash
claude /evolve
```

### 2. Set Backlog (if empty)
Edit `.plaesy/memory/backlog.md` and add ideas under `## Queued`:

```markdown
## Queued

- Add rate-limiting middleware to API
  - Rationale: Prevent abuse and DDoS attacks
  - Effort: Medium
  - Priority: 2

- Improve error messages for field validation
  - Rationale: Better UX, faster debugging for users
  - Effort: Low
  - Priority: 3
```

### 3. Loop Picks It Up
When you run `/evolve`, it will:
1. Read `.plaesy/memory/backlog.md`
2. Pop the first queued idea
3. Run full pipeline: `/research → /clarify → /flow → /implement → quality-gates → /assess → /optimize`
4. Check: does it pass quality gates?
5. If yes: move to `## Done`; if no: move to `## Blocked`

---

## Governance & Safety

Loop is designed to **not run away** — built-in stops every N iterations to ask you if it should continue.

### Configuration File: `.plaesy/state.json`

```yaml
max_iterations: 10              # Hard stop (default: 10 ideas)
max_consecutive_failures: 2     # Pause if 2+ ideas fail
checkpoint_interval: 3          # Ask you every 3 ideas
status: active                  # active | paused | stopped
```

**Edit this file anytime** to adjust limits. For example:
- Raise `max_iterations` to build more features
- Lower `checkpoint_interval` to review more often
- Set `status: paused` to pause between iterations

---

## How the Loop Works

### Step 1: Load Governance
Loop reads `state.json`:
- If `status != active` → **STOP** (waiting for you)
- If `iteration >= max_iterations` → **PAUSE** (ask you if you want to raise limit)

### Step 2: Pick Next Idea
Loop checks `.plaesy/memory/backlog.md`:
- **If queued ideas exist** → pop first one, start building
- **If empty** → generate idea from:
  1. Unresolved gaps in `.plaesy/memory/overview.md` ("Lessons Learned", "What Didn't Work")
  2. Open items from latest `/assess` or `/optimize` report
  3. Otherwise → **PAUSE** (ask you to provide backlog)

### Step 3: Build the Idea
Runs full pipeline:
```
/research → /clarify → /flow → /implement → 
quality-gates (run 1) → /assess → /optimize → 
quality-gates (final) → /fix (if needed) → /save
```

### Step 4: Check Quality Gates
Idea passes **only if all** of these hold:
- ✅ Level 1/2 issues resolved (no Level 3 critical blocks)
- ✅ Test coverage ≥90%
- ✅ No unresolved security findings
- ✅ Code saved & memory files within limits

**If it fails:** move to `## Blocked` in backlog.md, increment failure counter

**If it passes:** move to `## Done`, reset failure counter

### Step 5: Human Checkpoint
Loop always pauses to check in with you at these points:

| Condition | Action |
|-----------|--------|
| `consecutive_failures >= 2` | **STOP** — something is blocking ideas, review pattern |
| `iteration >= max_iterations` | **PAUSE** — ask if you want to raise limit |
| `(iteration - last_checkpoint) >= checkpoint_interval` | **PAUSE** — show summary, ask: continue / adjust backlog / stop |
| Otherwise | Continue to next iteration |

---

## Adjusting Loop Behavior

### I want the loop to run more ideas
Edit `.plaesy/state.json`:
```yaml
max_iterations: 20          # Was 10, now 20
```

### I want checkpoints more often
Edit:
```yaml
checkpoint_interval: 2      # Was 3, now every 2 ideas
```

### I want to pause and review
Edit:
```yaml
status: paused              # Loop will stop after current iteration
```

Then run `/evolve` again to resume.

### Ideas keep failing (stuck)
Loop will auto-pause if 2+ consecutive ideas fail. Check:
1. **Why are ideas failing?** Read quality-gates report for each blocked idea
2. **Fix root cause** (infrastructure issue, test setup, etc.)
3. **Edit blocked ideas** with a note of what needs to happen to unblock
4. **Reset counter**: edit `consecutive_failures: 0` in state.json
5. **Resume**: set `status: active` and run `/evolve`

---

## Backlog Format

### `.plaesy/memory/backlog.md`

**Queued** (next to pick up)
```markdown
## Queued

- One-line description of what to build
  - Rationale: Why (impact, priority, stakeholder need)
  - Effort: Low/Medium/High
  - Priority: 1-5 (1=highest)
```

**In Progress** (currently being built)
- Auto-filled by loop when it picks up a queued idea
- Don't edit while loop is running

**Blocked** (failed quality gates)
```markdown
## Blocked

- One-line description
  - Reason: Why it failed quality gates
  - Iteration: 7 (which loop iteration)
  - Next: What needs to happen to unblock
```

**Done** (passed all gates)
- Ideas that successfully shipped

---

## Driving the Loop

### Self-Paced (Recommended for Development)
```bash
claude /loop /evolve
```
- Runs one `/evolve` iteration
- Waits for next trigger (automatic)
- Continues until it hits a **STOP** condition

### Manual (One Idea at a Time)
```bash
claude /evolve
```
- Runs one iteration
- Reports result
- Returns — you decide if you want to run again

### Cron-Scheduled (Unattended)
```bash
claude /schedule "0 2 * * *" /evolve
```
- Runs `/evolve` every night at 2 AM
- Respects all governance limits (pauses on checkpoint/failure)
- Logs results to `.plaesy/state.json`

---

## Troubleshooting

### "Backlog is empty, loop paused"
**Why:** No ideas queued and loop cannot auto-generate ideas (missing overview.md or /assess reports)

**Fix:**
1. Add ideas to `.plaesy/memory/backlog.md` under `## Queued`
2. Run `/evolve` again

### "Consecutive failures = 2, loop paused"
**Why:** 2+ ideas failed quality gates in a row (likely a broken test, missing dependency, etc.)

**Fix:**
1. Check the blocked ideas — read why each failed
2. Fix the root cause (dependency, test environment, etc.)
3. Edit blocked ideas with "Next: [what needs to happen]"
4. Reset: `consecutive_failures: 0` in state.json
5. Set `status: active` and re-run `/evolve`

### "Iteration reached max_iterations"
**Why:** Loop hit the hard stop on iteration count

**Fix:**
Edit `max_iterations: 10` → `max_iterations: 20` in state.json, then re-run `/evolve`

### "Loop never ran (no status at all)"
**Why:** First time running `/evolve` — files don't exist yet

**Fix:**
Loop auto-creates `.plaesy/memory/backlog.md` and `.plaesy/state.json` on first run. Just run `/evolve`.

---

## Full Protocol Reference

For exhaustive details on loop iteration structure, exit criteria, and hard rules, see:
- **Orchestrator protocol**: `prompts/evolve.md`
- **Quality gate standards**: `instructions/quality-gates.instructions.md`
- **State files**: `.plaesy/state.json` and `.plaesy/memory/backlog.md`
