---
description: "Bug fixing and error resolution"
subagent: true
---

# `/fix` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Usage Format

```bash
/fix                              # Fix defects wherever found, dimension auto-detected
/fix:technical                    # Code bugs, runtime errors, integration failures
/fix:legal                        # Untraceable clauses, compliance gaps
/fix:marketing                    # Unsourced claims, brand inconsistencies
/fix:business                     # Broken unit-economics assumptions
/fix:design                       # WCAG violations, token inconsistencies

# Narrow to one sub-area within a dimension
/fix:technical --focus security       # Security vulnerabilities specifically
/fix:technical --focus performance    # Performance regressions specifically
/fix:technical --focus regression     # Test/behavior regressions specifically
```

Same 8-dimension scope set as `/assess:{scope}` (see `/assess`) — `--focus`
narrows to a specific sub-area within the chosen dimension; it never selects the
dimension itself.

## Objective
Resolve defects — with comprehensive root cause analysis and permanent fixes —
whatever dimension they're found in: a code bug (`/assess:technical`), an unsourced
claim or untraceable clause (`/assess:legal`/`/assess:marketing`), a broken unit-
economics assumption (`/assess:business`), or a WCAG violation (`/assess:design`).
The workflow below is written for code bugs; where a step is software-specific
(tests, Context7, deprecated libraries) it's marked so — swap in the matching
`assess-{dimension}.md` check for other dimensions.

## Protocol
**Resolve defects with comprehensive root cause analysis and permanent fixes**

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

## Anti-Patterns (NEVER Do These)
- ❌ **Never fix without reproducing** - Always verify bug is real
- ❌ **Never make speculative fixes** - Fix only what's broken, not "might be broken"
- ❌ **Never skip regression testing** - Must verify fix doesn't break other parts
- ❌ **Never document in code what should be tests** - Use tests, not comments, to prevent recurrence

## Need Help Diagnosing Root Cause?

When bug origin is unclear, call **`@nara`**:
- Is this a code bug, architecture issue, or design flaw?
- Patch (quick fix) vs refactor (permanent solution)?
- Should we fix symptom or root cause (risky vs safe)?
- Multiple bugs with same symptom (systemic issue)?
- Accessibility bug vs functionality bug priority?

**Example:**
```
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

**BEFORE marking fix complete, verify**:

- [ ] **Root cause identified** - Not just symptom fixed, but underlying cause addressed
- [ ] **Fix tested** - Reproduces bug, applies fix, verifies bug gone
- [ ] **Regression tested** - Full test suite passes, no new failures
- [ ] **Assumptions verified** - Not guessing at cause, actual evidence shows fix works
- [ ] **Related issues checked** - Could the same cause affect other areas?
- [ ] **Documentation updated** - Code comments explain non-obvious fixes
- [ ] **Prevention test added** - Test prevents bug from coming back
- [ ] **Ready for production** - Fix won't introduce new issues? Would you ship this?

If ANY checkbox fails, fix is not complete.

## Completion Format
```
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
**For all bug fixes**:
1. `mcp__context7__resolve-library-id` for affected technologies
2. `mcp__context7__get-library-docs` for error patterns and solutions
3. **Citation**: "Fix based on Context7 (/library/id) - Retrieved {{CURRENT_DATE}}"

**Research Areas**:
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
original bug (not a pure one-line patch), work through
`.plaesy/checklists/update.checklist.md` (impact assessment, rollback/re-scope options)
before committing to an approach below.

### User Decisions Required
- Fix approach selection
- Architectural changes needed
- Priority decisions for complex issues
- Rollback vs fix decisions

### User Decision Format
```
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
```
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

**See also**: [Universal Autonomous Routing](`.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory`)

After `/fix` completes, `/assess` (verification mode) is MANDATORY and cannot be
skipped — it confirms the fix works and catches new regressions. Regressions or
new issues found → loop back to `/fix` until clean; otherwise route onward per
the detail below (more bugs → `/fix` again, performance gap → `/optimize`,
otherwise ready for `/doc`).

### Dimension-Based Routing After Fix Verification

**Technical Bug Fixed**:
- Code bug fixed + tests passing → Verified, ready next steps
- Test failure regression → `/fix:technical --focus regression` (identify + fix test or code)
- Integration broken → `/fix` (debug integration issue)

**Security Vulnerability Fixed**:
- Patch applied + no regressions → Security verified ✓
- Alternative security risk found → `/fix:technical --focus security` (address systematically)
- Compliance gap uncovered → Escalate to `/assess:legal --security`

**Legal/Compliance Issue Fixed**:
- Regulatory gap closed → Compliance verified ✓
- Privacy requirement met → Document + proceed
- Risk assessment updated → Route to `/assess:legal` (formal validation)

**Other-Dimension Defect Fixed** (same pattern as Legal above — the template for
any non-software dimension):
- Claim sourced / assumption re-verified / clause traced → Route to
  `/assess:{dimension}` (formal validation, confirms the fix holds)
- Related defects found in the same section → `/fix` again for that dimension
- Fix reveals the whole deliverable needs rework → same Escalation Pattern below
  (`/start` new spec), not unique to code

**Performance Regression Detected**:
- Fix caused slowdown → Identify optimization path
- If fix is optimal → `/optimize` (improve surrounding area)
- If fix is slow → `/fix:technical --focus performance` (find faster solution)

### Escalation Pattern
**When fix requires major architectural changes:**

If fixing bug reveals fundamental architecture problem → Document findings → `/start` new specification → Full workflow → Reference original bug for context

Example: "Single-use password leak (Bug #123) revealed we need per-session tokens" → `/start` authentication system redesign

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` → [Global Routing](`.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory`)