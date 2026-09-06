---
description: "Essential error recovery patterns and resilience protocols"
applyTo: "**/*"
---


# Error Recovery Protocol

## Overview
Comprehensive error recovery patterns for Spec-Kit workflows with automatic healing capabilities.

---

## Error Classification

### Critical Errors (Workflow Halt)
- File system corruption
- Memory exhaustion
- Security compromise
- Disk space exhaustion

### Warning Errors (Auto-Recovery)
- Context7 API failures
- Missing test frameworks
- Application not running
- Git conflicts

### Informational Events (Log & Continue)
- Minor quality issues
- Memory limit approaching
- Performance degradation

---

## Auto-Recovery Patterns

### 1. Retry with Backoff
**Conditions**: Network timeouts, API rate limits, temporary locks
**Strategy**: 3 attempts, exponential backoff (1s → 2s → 6s)
**Failure**: Escalate to user decision

### 2. Graceful Degradation
**Conditions**: Context7 unavailable, performance monitor down
**Levels**: Cached data → General knowledge → User input

### 3. Auto-Repair and Continue
**Conditions**: Missing dependencies, simple git conflicts, memory limits
**Actions**: Auto-install, auto-resolve, auto-cleanup
**Verification**: Validate after repair

---

## Phase-Specific Recovery Decision Trees

### /start Phase Recovery

**Error: Context7 Unavailable**
```
Decision Flow:
1. Check ~/.context7/cache exists?
   ├─ YES → Load cached tech definitions
   └─ NO → Proceed with user input
2. Report: "Context7 cache unavailable; using standard patterns"
3. Store decision in context.md for /continue
```
**Recovery**: Use cached data + generate warning; continue workflow

**Error: Project Already Exists**
```
Decision Flow:
1. Scan existing project structure
2. Load existing .plaesy/memory/* files
3. Ask user: Continue from checkpoint? (Y/N)
   ├─ Y → Move to /continue workflow
   └─ N → Backup existing, start fresh
```
**Recovery**: Integrate with existing state; don't overwrite

**Error: Insufficient Permissions**
```
Decision Flow:
1. Identify specific permission (read/write/execute)
2. Attempt auto-fix (chmod/chown if applicable)
3. If fails → Request elevated privileges or manual intervention
```
**Recovery**: Halt gracefully with clear permission error message

---

### /research Phase Recovery

**Error: Context7 Rate Limited**
```
Decision Flow:
1. Pause current request
2. Wait exponential backoff (2s → 4s → 8s)
3. If retry limit exceeded (3x) → Switch to fallback sources
   ├─ GitHub raw docs
   ├─ Official website docs
   ├─ Stack Overflow patterns
   └─ Cached prior research
```
**Recovery**: Automatic source switching; document source origin

**Error: Conflicting Technology Information**
```
Decision Flow:
1. Collect all conflicting recommendations
2. Assign confidence scores (Context7 > Official Docs > Community)
3. Present to user: "Multiple recommendations found"
   - Recommendation A (confidence: 95%) from Context7
   - Recommendation B (confidence: 70%) from community
   - Choose A / Choose B / Review both
```
**Recovery**: Let user decide with confidence data

**Error: No Current Data for Deprecated Tech**
```
Decision Flow:
1. Detect deprecated technology (var, callback, request library, MD5, etc.)
2. Auto-suggest modern alternative (const, async-await, axios, bcrypt)
3. Store decision in research.md with rationale
```
**Recovery**: Flag deprecation + auto-suggest upgrade path

---

### /implement Phase Recovery

**Error: Test Coverage Below 90%**
```
Decision Flow:
1. Identify untested code paths (from coverage report)
2. Options:
   ├─ AUTO: Generate missing test stubs (user must fill)
   ├─ MANUAL: Report gaps + ask user to write tests
   └─ SKIP: Document exception + move to /assess (not recommended)
3. If AUTO chosen → Create test files, update coverage, re-check
```
**Recovery**: Either auto-generate stubs or escalate to user decision

**Error: Build/Compilation Fails**
```
Decision Flow:
1. Capture error message + file:line
2. Classify error type:
   ├─ Missing dependency → npm/pip/cargo install + retry
   ├─ Syntax error → Report file:line, halt gracefully
   ├─ Type error → Report type mismatch, ask for clarification
   └─ Unknown → Report full error, ask user
3. Never auto-fix syntax/type errors (user review required)
```
**Recovery**: Fix dependencies only; halt on code errors

**Error: Dependency Conflicts**
```
Decision Flow:
1. Analyze conflicting versions (e.g., React 17 vs 18)
2. Options:
   ├─ Compatible? Update to compatible version
   └─ Incompatible? Report conflict + ask user choice
3. If user chooses version → Update lock file + rebuild
```
**Recovery**: Ask user for version preference on conflicts

---

### /optimize Phase Recovery

**Error: Application Not Running**
```
Decision Flow:
1. Try start app (npm start, python app.py, etc.)
2. If fails → Fall back to static analysis (code profiling without runtime)
3. Report: "Dynamic profiling unavailable; using static analysis"
4. Continue optimization with available data
```
**Recovery**: Degrade gracefully to static analysis

**Error: Performance Targets Not Met After Optimization**
```
Decision Flow:
1. Analyze what was optimized (DB, cache, algorithm, etc.)
2. If minimal impact → Report findings + ask next steps
3. Options:
   ├─ Continue optimizing (different area)
   ├─ Accept current performance + document why
   └─ Escalate to /fix if performance is regression
```
**Recovery**: Transparent reporting + user decision

**Error: Resource Exhaustion (Memory/Disk)**
```
Decision Flow:
1. Detect low memory/disk
2. Immediate actions:
   ├─ Stop background processes
   ├─ Clear build caches
   ├─ Cleanup temp files
3. If still insufficient → Halt optimization, ask user to free space
```
**Recovery**: Auto-cleanup + halt if unresolvable

---

### /assess Phase Recovery

**Error: Tests Fail to Run**
```
Decision Flow:
1. Try multiple test runners (jest, pytest, go test, etc.)
2. If all fail → Report which tools are missing
3. Options:
   ├─ AUTO: Attempt to install missing tools
   ├─ MANUAL: Ask user to install manually
   └─ SKIP: Report "Cannot assess; tests unavailable"
```
**Recovery**: Try auto-install; escalate if needed

**Error: Code Doesn't Compile**
```
Decision Flow:
1. Attempt to parse/understand code anyway (static analysis)
2. Report: "Code does not compile; running static analysis only"
3. Skip dynamic metrics (coverage, performance)
4. Provide quality assessment from source code alone
```
**Recovery**: Degrade to static analysis; report limitations

---

### /fix Phase Recovery

**Error: Fix Causes Test Regression**
```
Decision Flow:
1. Detect: Tests passed before fix, fail after fix
2. Revert fix immediately
3. Analyze: Which tests broke?
4. Report regression + ask user:
   ├─ Try different fix approach
   ├─ Review test assumptions (might be flawed)
   └─ Mark as blocker, escalate
```
**Recovery**: Revert + ask for guidance

**Error: Root Cause Unidentifiable**
```
Decision Flow:
1. Collect all diagnostic data (logs, error traces, git history)
2. Report: "Root cause unclear" + data collected
3. Options:
   ├─ User investigation (ask user for context)
   ├─ Patch approach (apply temporary workaround)
   └─ Escalate to /start (might be design issue)
```
**Recovery**: Transparent escalation to user

---

## Recovery State Management

### Checkpoint System
Auto-save at:
- Before major phase transitions
- After quality gate validation
- During long operations
- Before risky operations

### Recovery Flow
1. Detect error type and severity
2. Create recovery checkpoint
3. Execute primary recovery action
4. Validate success
5. Continue or escalate

---

## Prevention & Self-Healing

### Predictive Prevention
- Pre-flight checks (dependencies, space, permissions)
- Runtime monitoring (memory, API limits, performance)
- Adaptive thresholds based on patterns

### Self-Healing Actions
- Configuration repair (restore defaults + customizations)
- Dependency repair (reinstall/update broken packages)
- Performance repair (apply known fixes)
- Memory repair (cleanup, restart components)

---

## Recovery Metrics

### Key Indicators
- Success rate: >90% target
- Recovery time: <5 minutes for auto-recovery
- Error frequency: Decreasing trend
- User intervention: <10% of errors

### Analytics
- Error pattern analysis
- Recovery effectiveness by type
- Workflow impact assessment
- Prevention strategy optimization

---

## Integration Template

```markdown
## Error Recovery Protocol

### Automatic Recovery
- Context7 failures → Retry with alternatives
- Missing dependencies → Auto-install tools
- Memory issues → Auto-cleanup
- Git conflicts → Auto-resolve simple cases

### Manual Recovery Required
- Critical system errors
- Complex dependency conflicts
- Security issues
- Major architectural decisions

### Recovery Status
- [ ] Recovery mechanisms active
- [ ] Error patterns analyzed
- [ ] Prevention measures applied
- [ ] Metrics monitoring enabled
```

---

*Error Recovery Protocol v1.0.0*
*Framework: Spec Kit Constitutional Development Framework*