---
description: "Bug fixing and error resolution"
subagent: true
---

# `/fix` command instructions

⚡ **Run with**: any parallel-capable mode — a host that provides
subagents, or a mode that does. Not a requirement: a single-agent host
runs this identically, just serially. `/loop` documents when fan-out is
worth its cost.

🔄 **Invocation mode**: dispatched as a background task per
`.plaesy/instructions/plaesy.md` → Long-Running Commands. Control returns to
the user immediately; completion is reported, not polled. Run synchronously
only if the user explicitly asks to wait or this conversation's next step
needs the result right away.

## Usage Format

```bash

/fix                              # Fix defects wherever found, dimension auto-detected
/fix:technical                    # Code bugs, runtime errors, integration failures
/fix:design                       # WCAG violations, token inconsistencies
/fix:business                     # Broken unit-economics assumptions
/fix:marketing                    # Unsourced claims, brand inconsistencies
/fix:legal                        # Untraceable clauses, compliance gaps
/fix:financial                    # Broken cost, pricing, or profitability figures
/fix:product                      # Feature/roadmap inconsistencies
/fix:management                   # Broken process, role, or org-health claims
/fix:operations                   # Deployment, release, observability, runbook defects

# Narrow to one sub-area within a dimension
/fix:technical --focus security       # Security vulnerabilities specifically
/fix:technical --focus performance    # Performance regressions specifically
/fix:technical --focus regression     # Test/behavior regressions specifically
```

Same 9-dimension scope set as `/assess:{scope}` — `technical`, `design`,
`business`, `marketing`, `legal`, `financial`, `product`, `management`,
`operations`. `--focus` narrows to a sub-area within the chosen dimension; it
never selects the dimension itself.

## Objective

Resolve defects — with root cause analysis and permanent fixes — in whatever
dimension they're found: a code bug (`/assess:technical`), an unsourced claim
or untraceable clause (`/assess:legal`/`/assess:marketing`), a broken unit-
economics assumption (`/assess:business`), a WCAG violation (`/assess:design`).
The workflow below is written for code bugs; software-specific steps (tests,
Context7, deprecated libraries) are marked — swap in the matching
`assess-{dimension}.md` check for other dimensions.

## Protocol

Resolve defects with root cause analysis and permanent fixes

## Validation Checklist (Before Running)

- ✅ Bug reproduced (steps to reproduce documented)
- ✅ Affected files identified (not speculative)
- ✅ Tests exist to prevent regression
- ✅ Impact assessed (critical/high/medium)

## Bug Fix Workflow

1. **Analysis** - Parse errors, identify affected components, assess impact
2. **Root Cause Investigation** - Trace execution flow, analyze recent changes
3. **Context7 Research** - Research error patterns and current solutions
4. **Fix Implementation** - Apply permanent fixes with validation
5. **Testing** - Verify fixes work and no regressions
6. **Documentation** - Update documentation and prevent future issues

## Fix Standards

- **Reproduce before fixing** — an unreproduced bug may not be the bug
- **Fix what is broken, not what might be** — a speculative fix widens blast radius
- **Regression-test every fix** — the only way to learn it didn't break something else
- **Encode prevention as a test, not a comment** — a comment is never run; a test is.

## Need Help Diagnosing Root Cause?

When bug origin is unclear, call **`@nara`**:

- Is this a code bug, architecture issue, or design flaw?
- Patch (quick fix) vs refactor (permanent solution)?
- Should we fix symptom or root cause (risky vs safe)?
- Multiple bugs with same symptom (systemic issue)?
- Accessibility bug vs functionality bug priority?

**Example:**

```text

@nara Bug: Modal dialog sometimes doesn't close. Happens in Chrome/Safari
but not Firefox. Could be timing issue, CSS bug, or React state problem.
Root cause unclear - which direction to investigate?
```

---

## Error Recovery

**If fix causes regression:**

- Revert fix immediately
- Diagnose root cause (why did it break?)
- Try different fix approach

## Self-Audit Before Finalizing (Anti-Hallucination)

**Before marking fix complete, verify**:

- [ ] **Root cause identified** - underlying cause addressed, not just the symptom
- [ ] **Fix tested** - reproduces bug, applies fix, verifies bug gone
- [ ] **Regression tested** - full test suite passes, no new failures
- [ ] **Assumptions verified** - evidence shows the fix works, not a guess
- [ ] **Related issues checked** - could the same cause affect other areas?
- [ ] **Documentation updated** - non-obvious fixes explained in comments
- [ ] **Prevention test added** - stops the bug from recurring
- [ ] **Ready for production** - would you ship this?

Any unchecked box means the fix is not complete.

## Completion Format

```text

✅ Fix Complete
Root cause: [identified and documented with evidence]
Fix: [applied with file:line]
Tests: [all pass/regression test added]
Prevention: [test added to prevent recurrence]
Confidence: [HIGH/MEDIUM - based on testing]

Modified Files: [list]
Tests Added: [list]
Documentation: Updated
```

## Context7 Protocol (Required)

Load `.plaesy/instructions/context7-protocol.md` and follow it as written — it
holds the mandatory-lookup table, the exact tool sequence, the required citation
form, and the fallback chain. It is the single source for this protocol; do not
restate or re-derive the sequence here.

For a bug fix: resolve the library IDs the fix touches, cite as
`Fix based on Context7 (/library/id) — Retrieved {{CURRENT_DATE}}`. A fix built
on a guessed API signature fixes a bug that doesn't exist.

**Research Areas** (what to look up, per that file's sequence):

- Error patterns in relevant frameworks
- Common bugs and their solutions
- Current best practices for bug fixing
- Security vulnerability fixes

## Defect Classification

### Common Types (Software)

- **Syntax/Logic Errors** - Code implementation issues
- **Runtime Errors** - Execution failures
- **Configuration Issues** - Environment/setup problems
- **Integration Failures** - Component interaction issues
- **Performance Problems** - Efficiency bottlenecks
- **Security Vulnerabilities** - Security-related issues

### Common Types (Other Dimensions)

- **Unsourced Claim** - Figure/statement with no citation (legal, marketing, business)
- **Untraceable Clause** - Contract/policy term with no owner or source (legal)
- **Broken Assumption** - Unit-economics or market-sizing figure that doesn't hold
  up under research (business, financial)
- **Brand/Positioning Inconsistency** - Copy contradicts established voice or prior
  claims (marketing)
- **Accessibility Violation** - WCAG 2.1 AA failure (design)

### Priority Levels

**Critical** (<15min):

- Hotfix deployment or rollback
- Immediate security patches
- Production downtime issues

**High** (<1hr):

- Significant functionality issues
- Performance degradation
- Integration failures

**Medium** (<4hr):

- Minor functionality issues
- UI/UX problems
- Documentation issues

## Fix Strategy

**Reference checklist**: for a fix that changes scope/behavior beyond the
original bug (not a one-line patch), work through
`.plaesy/checklists/update.checklist.md` (impact assessment, rollback/re-scope
options) before committing to an approach.

### User Decisions Required

- Fix approach selection
- Architectural changes needed
- Priority decisions for complex issues
- Rollback vs fix decisions

### User Decision Format

```text

DECISION NEEDED
📊 Bug Analysis: [issue description]
🎯 Options:
  1. [Option A] - [Quick fix approach]
  2. [Option B] - [Comprehensive fix]

💡 Recommendation: [Option X] - [Rationale]
```

## Implementation Process

1. **Apply Fix** - Minimal, targeted changes
2. **Test** - Comprehensive validation
3. **Validate** - No regressions introduced
4. **Document** - Prevention strategies

## Testing Protocol

**Required Tests**:

- Reproduce original bug
- Verify fix works
- Test related functionality
- Check for regressions
- Performance validation

```bash

# Test commands
npm test || yarn test || pytest || go test
npm run build || yarn build || make build
npm start || yarn start || python app.py
```

## Progress Format

```text

Analyzing: [bug type]
Investigating: Root cause
Fixing: [implementation]
Testing: Validation
```

## Fix Documentation

### Bug Record Template

```markdown

# Fix Record: [Bug Title]

## Bug Summary
- **Error**: [Description]
- **Severity**: Critical/High/Medium/Low
- **Impact**: [Affected users/features]

## Root Cause
[Detailed analysis]

## Fix Implementation
[What was changed and why]

## Validation
- Tests added: [List]
- Tests passed: [Results]

## Prevention Measures
[What was added to prevent recurrence]

## Related Files
- Modified: [List]
- Tests added/updated: [List]
```

### Prevention Strategy

- **Code**: Add validation, defensive programming, error handling, logging
- **Testing**: Automated tests, improved coverage, monitoring alerts
- **Process**: Update code review checklist, linting rules, CI/CD
- **Monitoring**: Error monitoring, performance monitoring, health checks

### Complexity Assessment

- **Simple Fix** (<1 hour): Single file, clear root cause, no architectural changes
- **Medium Fix** (1-4 hours): Multiple files, moderate investigation, refactoring
- **Complex Fix** (>4 hours): Architectural changes, deep investigation, extensive testing

### Escalation to New Spec

**Escalate when**: Major architectural changes, multiple features affected, breaking changes

**Process**: Document findings → `/start` new spec → full workflow → reference original bug

## Deprecated Technology Detection

**Critical Issues to Fix**:

- ❌ `request` library → ✅ `axios` or `node-fetch`
- ❌ `var` keyword → ✅ `const`/`let`
- ❌ Callback patterns → ✅ Promise/async-await
- ❌ MD5/SHA-1 for passwords → ✅ bcrypt/Argon2
- ❌ jQuery DOM manipulation → ✅ Native DOM API

**Bug Fix Protocol**:

1. **IDENTIFY**: If bug caused by deprecated library, flag as security issue
2. **REPLACE**: Fix by upgrading to modern alternatives, not patching deprecated code
3. **UPGRADE**: Include library upgrade as part of the fix
4. **DOCUMENT**: Explain why deprecated tech caused issue + benefits of replacement

---

## Autonomous Routing (After Bug Fix)

**See also**: [Universal Autonomous Routing](.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory)

After `/fix`, `/assess` (verification mode) is mandatory — it confirms the fix
works and catches new regressions. Regressions or new issues → loop back to
`/fix` until clean; otherwise route per the detail below (more bugs → `/fix`
again, performance gap → `/optimize`, otherwise ready for `/doc`).

### Dimension-Based Routing After Fix Verification

**Routing**: per-dimension routing and the severity→priority order are in `.plaesy/instructions/dimension-mapping.md`. Look the dimension up there; do not restate the table here.

What applies to **every** dimension, regardless of which row you land on:

- **Verified fix** → the fix holds; record it and move to the next finding. If the
  fix is in a non-software dimension (a sourced claim, a traced clause, a
  corrected unit-economics figure), the formal confirmation pass is
  `/assess:{dimension}` — the same route the canonical table gives that dimension
  under *Assess*.
- **New defect found while fixing** → `/fix` again for that dimension; it is the
  dimension's own error-recovery command, whatever dimension that is.
- **Regression from the fix** → same-dimension re-fix, narrowed with
  `--focus regression` (broken test/behavior) or `--focus performance` (the fix
  itself caused a slowdown) — and if the fix was already the optimal shape,
  hand the surrounding area to `/optimize` instead of patching again.
- **Fix reveals the deliverable as a whole needs rework** → the Escalation
  Pattern below (`/start` a new spec). This is not a code-only rule; it applies
  the same way to a business case or a clause set.
- **Security or compliance work uncovers a neighbouring gap** → a compliance gap
  surfaces during a security fix (or vice versa) → route to the compliance/legal
  assessment, not back into the same technical fix.

**Role handoff**: once the verified fix is recorded, look up
`.plaesy/instructions/role-mapping.md`'s Routing Table for the fixing role's
"Typical next role," write the handoff to `state.json.role_handoff_log`, and
invoke `/continue` rather than stopping. No hop limit; a hard-stop category
(`.plaesy/instructions/plaesy.md` rule 8) is the only thing that ends the
chain early.

### Escalation Pattern

**When fix requires major architectural changes:**

If fixing bug reveals fundamental architecture problem → Document findings → `/start` new specification → Full workflow → Reference original bug for context

Example: "Single-use password leak (Bug #123) revealed we need per-session tokens" → `/start` authentication system redesign

---

## Worked Example — Fix N+1 Query in Go Handler

**Input**: `/assess:technical` found `MEDIUM performance — p95 240ms exceeds 200ms target`
at `internal/handler/user.go:42` (N+1 query in `ListUsers`). User runs
`/fix:technical --focus performance`.

```text
Analyzing: performance regression (N+1 query)
Investigating: Root cause — ListUsers loads User then iterates to fetch Profile per user
Fixing: internal/handler/user.go:42 — replace loop with single JOIN query via repository
Testing: Validation — go test -run TestListUsers (covers N+1 scenario)
Testing: Regression — full suite passes (247 tests)

✅ Fix Complete
Root cause: N+1 query in ListUsers — each user triggered a separate Profile fetch
Fix: internal/handler/user.go:42 — replaced with repository.ListUsersWithProfiles (JOIN)
Tests: all pass + regression test added (TestListUsers_NoNPlusOne)
Prevention: TestListUsers_NoNPlusOne asserts query count = 1 via sqlmock
Confidence: HIGH — verified with EXPLAIN ANALYZE + test suite

Modified Files: [internal/handler/user.go, internal/repository/user.go, internal/repository/user_test.go]
Tests Added: [TestListUsers_NoNPlusOne]
Documentation: Updated
```

**Routing after**: `/assess:technical` (verification mode) confirms p95 180ms, score 87/100.

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` →
[Global Routing](.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory)
