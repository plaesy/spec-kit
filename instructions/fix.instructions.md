---
description: "Bug fixing and error resolution"
---

# `/fix` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective
Resolve bugs with comprehensive root cause analysis and permanent fixes

## Protocol
**Resolve bugs with comprehensive root cause analysis and permanent fixes**

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

## Completion Format
✅ Fix Complete
- Root cause: [identified and documented]
- Fix: [applied with file:line]
- Tests: [pass/regression found]
- Prevention: [test added to prevent recurrence]

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

## Bug Classification

### Common Types
- **Syntax/Logic Errors** - Code implementation issues
- **Runtime Errors** - Execution failures
- **Configuration Issues** - Environment/setup problems
- **Integration Failures** - Component interaction issues
- **Performance Problems** - Efficiency bottlenecks
- **Security Vulnerabilities** - Security-related issues

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

## Completion Format
```
✅ Bug resolved
Root cause: [identified cause]
Fix: [implemented solution]
Tests: All passing
Prevention: [strategies implemented]

Modified Files: [list]
Tests Added: [list]
Documentation: Updated
```

## Critical Rules
- ✅ **Root cause analysis mandatory**
- ✅ **Minimal changes for fixes**
- ✅ **Comprehensive testing required**
- ✅ **Document prevention strategies**
- ✅ **No regressions allowed**

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

## Success Criteria
- Bug fully resolved
- No regressions introduced
- Root cause addressed
- Prevention measures in place
- Documentation updated

## Next Steps
- `/assess` for validation
- `/optimize` if performance can improve
- Complete if all resolved

/save

---

**Follow shared protocols**: `.plaesy/memory/quality-gates.md` → `.plaesy/memory/error-recovery.md`