---
description: "Autonomous workflow loop orchestrator - assess, fix, verify, repeat without user input"
subagent: true
---

# `/loop` - Autonomous Continuous Improvement

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective

Execute autonomous workflow loop without user intervention:
```
ASSESS (find issues) 
  → IMPLEMENT (fixes) 
  → VERIFY (improvements) 
  → REPEAT 
  → STOP (only when user commands)
```

**Philosophy**: Framework knows what to do. No more "what's next?" — workflow is self-directed.

---

## When to Use `/loop`

- **During development**: Run in background, fixes issues autonomously
- **Continuous improvement**: Keep running until quality threshold reached
- **Unattended execution**: Walk away, framework keeps improving
- **Post-implementation**: Auto-optimize after `/implement` completes

**Do NOT use `/loop` when**: Manual decisions needed, user must review changes per phase

---

## Autonomous Loop Protocol

### Phase 0: Load Loop Config (`.plaesy/state.json`)

Read `.plaesy/state.json` (created by `plaesy init` from `.plaesy/templates/state.template.json`)
for this run's controls — do not re-derive these from narrative defaults if the file
has them:
- `autonomous_loop.max_iterations` (default 10), `.current_iteration`,
  `.consecutive_failures` / `.max_consecutive_failures` (default 2), `.status`
  (`paused`/`running`), `.checkpoint_interval` (default 3), `.last_human_checkpoint`
- `quality_gates.*` — which gates are mandatory for this project (code review, tests,
  docs, constitutional compliance)

Update `current_iteration`, `consecutive_failures`, and `status` in this file after
every iteration (see Phase 4) — this is the loop's actual state, not just a narrative
log. If the file is missing or empty (pre-existing project from before this file
existed), fall back to the defaults above and note that in the first iteration's output.

### Phase 1: Intelligent Assessment (No User Input)

**Execution**:
1. Load project state (`.plaesy/context.md`, `.plaesy/memory.md`, `.plaesy/memory/constitution.md`)
2. Read the active dimension(s) from the constitution and run `/assess:{dimension}`
   for each active one (`/loop:{dimension}` overrides this to a single dimension —
   see Loop Configuration below). Default to `technical` only when the constitution
   declares no other dimension.
3. Parse findings: categorize by severity & type
4. Filter findings:
   - ✅ **Auto-fixable** (software: code quality, simple bugs, tests, docs. Other
     dimensions: sourcing a missing citation, filling an ASSUMED-labeled default,
     fixing a broken cross-reference, updating a stale figure)
   - ⚠️ **Needs review** (software: architecture, security, UX. Other dimensions:
     unit-economics assumptions, legal clause interpretation, brand positioning calls)
   - ⏸️ **Blocked** (external dependencies, user decisions)

**Output**: List of fixable issues ranked by impact

### Phase 2: Intelligent Implementation (Parallel Execution)

**For each auto-fixable issue**:
1. Load appropriate instruction file (tech-specific/tool-specific for software;
   `assess-{dimension}.md` for other dimensions)
2. Execute fix — software: with TDD enforcement. Other dimensions: revise the
   deliverable section, re-run Mode 1 ambiguity resolution on just that section
3. Validate fix doesn't break existing tests (software) or introduce a new
   uncited claim/broken cross-reference (other dimensions)
4. Commit with clear message

**Parallel execution** (if multiple independent fixes):
- Fix #1: Code quality → Test improvements → Commit
- Fix #2: Performance bug → Refactor → Commit
- Fix #3: Documentation → Spell check → Commit
- (All run in parallel, merge results)

**Order of execution** (by impact):
1. Security issues (highest priority)
2. Test failures (blocks builds)
3. Performance issues (per constitution target, default >200ms)
4. Code quality (duplication, complexity)
5. Documentation gaps

### Phase 3: Verification (Automated Validation)

**After each fix**:
1. Run full test suite
2. Verify coverage still meets constitution minimum (default ≥90%)
3. Check performance metrics
4. Validate against `.plaesy/instructions/quality-gates.md`
5. Compare before/after metrics

**If verification fails**:
- Rollback fix
- Move (or create) the task file to `.plaesy/tasks/backlog/` with the failure noted in
  its frontmatter — this is the only backlog; there is no separate `backlog.md`
- Mark as "needs review"
- Continue to next issue

**If verification passes**:
- Keep fix
- Update metrics
- Log to progress tracker
- Continue loop

### Phase 4: Loop Control (When to Continue vs Stop)

**After every iteration**, update `.plaesy/state.json`:
- Increment `autonomous_loop.current_iteration`
- Reset `consecutive_failures` to 0 on a successful iteration, increment it on failure
- Set `status: "running"` while looping, `"paused"` on stop (see below)
- Set `last_human_checkpoint` when `current_iteration % checkpoint_interval == 0`

**Continue loop IF**:
- ✅ Issues found & fixed in last iteration
- ✅ Quality score improved
- ✅ `current_iteration < max_iterations` (state.json, default 10)
- ✅ `consecutive_failures < max_consecutive_failures` (state.json, default 2)
- ✅ User hasn't sent STOP command

**Stop loop IF**:
- ❌ User sends `/stop` command (immediate) → set `status: "paused"`
- ❌ All assessable issues auto-fixed
- ❌ Remaining issues require user decisions
- ❌ Quality threshold reached (score ≥85/100)
- ❌ `current_iteration ≥ max_iterations` or `consecutive_failures ≥ max_consecutive_failures`
  (infinite-loop guard, per state.json)

---

## Autonomous Decision Rules

**When `/loop` encounters choices**, apply these rules (NO USER INPUT NEEDED):

These rules are written for the `software` dimension by default. When the
constitution declares a different active dimension, apply the matching variant
listed under each rule instead — same priority order, different substance.

### 1. Tech Stack / Foundational Choices
**Default (software)**: Use tech stack already in project — extend existing, don't replace
**Business/marketing/legal**: Use the positioning, pricing model, or legal framework
already committed in the spec — extend, don't relitigate a settled choice
**Example**: Project uses React → Extend React, don't suggest Vue. Plan already commits
to a subscription model → extend it, don't re-open freemium-vs-subscription

### 2. Architecture / Structural Decisions
**Default (software)**: Preserve existing architecture — minimal refactor, maximum compatibility
**Business/legal/marketing**: Preserve existing document/campaign structure — minimal
reorganization, maximum compatibility with what stakeholders already reviewed
**Example**: Project has monolith → add features to monolith. Contract already
structured by clause type → add new terms to matching clause, don't restructure

### 3. Performance/Viability vs Features
**Default (software)**: Fix performance issues first — constitution performance
target is non-negotiable (default <200ms)
**Business**: Fix unit-economics viability gaps first — constitution viability
criteria are non-negotiable (e.g. CAC payback period)
**Legal**: Fix compliance gaps first — non-negotiable regardless of other priorities
**Example**: Feature request vs latency bug → fix latency first. New copy vs an
unsourced pricing claim → fix the claim first

### 4. Quality vs Speed
**Default (software)**: Code quality wins if <1 week refactor — technical debt compounds
**Other dimensions**: Accuracy/sourcing wins if a quick fix — an unverified claim or
untraceable clause compounds risk the same way technical debt does
**Example**: Duplicate code → extract even if it slows the current phase. Duplicated
uncited claim across two sections → fix both, even if it slows the current draft

### 5. Verification Strategy
**Default (software)**: constitution coverage minimum is non-negotiable (default
90%) — tests first (TDD), then code
**Other dimensions**: constitution's mandatory-audit row (see `/assess`
Design Spine) is non-negotiable — verify sourcing/traceability before drafting more
**Example**: Gap in tests → write tests before fixing bugs. Gap in citations → source
the claim before adding more content that depends on it

### 6. Documentation
**Default**: Auto-generate where possible, manual review where needed
**Pattern**: README + API docs + architecture narrative (software) — or handoff spec
per `/assess` Design Spine table (business case, ADRs, runbook, etc.) for
other dimensions
**Example**: Code changes → auto-update related docs. Business model changes →
auto-update the business case's assumption register

---

## Loop Iteration Example

### Iteration 1: Assessment
```
ASSESS Output:
- Security: 1 SQL injection risk (HIGH) ✅ Auto-fixable
- Tests: 15% coverage gap (HIGH) ✅ Auto-fixable
- Performance: 800ms query (HIGH) ✅ Auto-fixable
- Design: Missing dark mode (LOW) ⚠️ Needs review
- Docs: Missing API examples (LOW) ✅ Auto-fixable

Auto-fixable count: 4/5
Proceeding to fixes...
```

### Iteration 2: Implementation (Parallel)
```
Fix 1: SQL injection (use parameterized queries)
Fix 2: Coverage gap (write missing tests)
Fix 3: Query performance (add index + caching)
Fix 4: API docs (generate from code)

Status: All 4 fixes in progress...
```

### Iteration 3: Verification
```
Verification Results:
✅ Fix 1: Security verified (no SQL injection)
✅ Fix 2: Coverage now 92% (from 75%)
✅ Fix 3: Query now 45ms (from 800ms)
✅ Fix 4: API docs auto-generated

Quality score: 82/100 (was 68/100) ⬆️ +14 points
Proceeding to next assessment...
```

### Iteration 4: Re-assessment
```
ASSESS Output:
- No high-severity issues remaining
- Design improvements identified (LOW priority)
- Performance optimizations available (MEDIUM)
- Documentation complete

Quality score: 82/100
User decisions needed for remaining items.

Loop Status: PAUSED (waiting for user direction)
Recommendation: Run `/optimize` for performance polish, or `/continue` for next features
```

---

## Infinite Loop Prevention

**Detect infinite loop IF**:
- Same issue appears 3+ consecutive iterations
- Quality score doesn't improve for 2 iterations
- Same fix is applied twice in a row

**Action on infinite loop**:
1. Log issue with details
2. Mark as "needs review"
3. Pause loop
4. Recommend manual review OR `/fix` with user context

**Example**:
```
INFINITE LOOP DETECTED:
- Issue: "Test coverage below constitution minimum" recurring
- Cause: New code without tests keeps added
- Root: Auto-fix can't understand new business logic

Action: Paused loop
Recommendation: Run `/fix` with manual context about new features
```

---

## Progress Tracking

**Real-time output format**:
```
🔄 LOOP ITERATION: 3/10 max
├─ Status: ASSESSING
├─ Current Phase: Find issues
├─ Quality Score: 82/100 (⬆️ +14 from baseline)
├─ Issues Fixed: 4
├─ Tests Added: 23
├─ Performance: +2.3x improvement
└─ Next: Verify & assess again
```

**Completion format**:
```
✅ LOOP COMPLETE
├─ Iterations: 5
├─ Issues Fixed: 12
├─ Quality Improvement: 68 → 88/100 (+20)
├─ Coverage: 75% → 92%
├─ Performance: 800ms → 45ms queries
├─ Tests Added: 47
└─ Recommendation: Ready for `/optimize` phase
```

---

## Loop Configuration

**Customize loop behavior** (optional — overrides `.plaesy/state.json` for this run only,
does not persist the override back to the file):
```markdown
/loop:technical                # Only loop one dimension (technical/business/legal/
                                # marketing/design/product/financial/management;
                                # default: all dimensions active in the constitution)
/loop --max-iterations 15      # Stop after 15 iterations (default: state.json, else 10)
/loop --quality-target 90      # Stop when quality ≥90 (default: 85)
/loop --auto-commit            # Commit after each fix (default: batch at end)
/loop --parallel 4             # Run up to 4 independent fixes in parallel (default: sequential)
/loop --no-verify              # Skip verification (NOT recommended)
```

Same 8-dimension scope set as `/assess:{scope}` (see `/assess`) — the
other flags above are run parameters, not scope selectors, so they stay `--flag`.

---

## Integration with Other Commands

**`/loop` Workflow Pattern**:
```bash
/start Build X                 # Initial setup
/continue                      # Execute implementation
/loop                          # Auto-improve until user stops
/optimize                      # Manual optimization decisions
/loop                          # Auto-verify improvements
/save                          # Final checkpoint
```

**Stopping the loop**:
```bash
# User can stop anytime:
/stop                          # Stop current loop gracefully
Ctrl+C                         # Emergency stop

# Loop auto-stops when:
- Quality threshold reached
- All auto-fixable issues fixed
- Infinite loop detected
- User issues stop command
```

---

## Success Criteria

Loop is working when:
- ✅ Quality score improves each iteration
- ✅ No user prompts/questions
- ✅ Tests keep passing
- ✅ Performance improves or stays same
- ✅ Clear progress each iteration
- ✅ Stops when appropriate (not infinite loop)
- ✅ All fixes are documented

---

## Error Recovery

**If loop encounters error**:
1. Log full error with context
2. Rollback last change
3. Mark issue as "blocked"
4. Try next issue
5. Continue loop

**If critical error** (build broken):
1. Rollback all recent commits
2. Pause loop
3. Recommend manual `/fix`

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`

**Usage**: Run `/loop` after `/implement` to auto-improve until completion or user stops!
