---
description: "Comprehensive testing strategy and validation framework for all project types"
applyTo: "**/*.js, **/*.ts, **/*.py, **/*.java, **/*.go, **/*.rs, **/*.cs, **/*.rb, **/*.kt, **/*.dart"
---

# Testing Strategy & Validation Framework

## Overview
Unified testing approach across all phases, enforcing quality gates and preventing regressions.

---

## Testing Pyramid (Standard Model)

```
        △
       /|\
      / | \  End-to-End Tests (5-10%)
     /  |  \
    /---|---\
   /  | | |  \ Integration Tests (15-20%)
  / ---|---|--\
 /   | | | |   \ Unit Tests (70-75%)
/___|_|_|_|_____\
```

### Ratio by Project Type

| Type | Unit | Integration | E2E | Total |
|------|------|-------------|-----|-------|
| **Backend/API** | 70% | 20% | 10% | ✅ 90%+ |
| **Frontend/Web** | 60% | 20% | 20% | ✅ 90%+ |
| **Full-Stack** | 65% | 20% | 15% | ✅ 90%+ |
| **CLI/Library** | 80% | 10% | 10% | ✅ 90%+ |
| **Mobile** | 65% | 20% | 15% | ✅ 90%+ |

---

## Phase-by-Phase Testing Requirements

### /implement Phase
**Requirement**: 90% code coverage minimum

```
Test checklist:
- [ ] Unit tests: All functions/methods tested
- [ ] Happy path: Core functionality works
- [ ] Error paths: Exception handling tested
- [ ] Edge cases: Boundary conditions validated
- [ ] Integration: Component interactions verified
- [ ] Coverage report: ≥90% lines covered
- [ ] No regressions: All prior tests still pass
```

### /assess Phase
**Requirement**: Validate test quality

```
Test quality checklist:
- [ ] Tests pass locally
- [ ] Tests pass in CI/CD
- [ ] Coverage ≥90%
- [ ] No flaky tests (deterministic)
- [ ] Test naming clear (describes behavior)
- [ ] Arrange-Act-Assert pattern followed
- [ ] No test interdependencies
- [ ] Mocks used sparingly (real dependencies preferred)
```

### /optimize Phase
**Requirement**: Performance testing

```
Performance test checklist:
- [ ] Baseline metrics recorded (before optimization)
- [ ] Load test passes (1000+ req/s for APIs)
- [ ] Response time <200ms (for APIs)
- [ ] Memory usage <512MB (baseline)
- [ ] CPU <70% under load
- [ ] No memory leaks detected
- [ ] Regression test suite passes
- [ ] Performance improvements documented
```

---

## Integration Testing Between Phases

### Phase Interdependencies
- /start → /research: Technology validation
- /research → /design (if UI): Tech decisions → design specs
- /research → /implement: Tech validated → ready to build
- /design → /implement: Design specs → component API
- /implement → /assess: Code → quality measurement
- /assess → /research (loop): Assessment findings → validate solutions
- /assess → /optimize: Quality baseline → performance targets
- /optimize → /fix: Performance issues → bug fixes

### Validation Template
```
Validation for [Phase A] → [Phase B]:
- [ ] Previous phase outputs exist
- [ ] Current phase reads from previous
- [ ] No breaking changes between phases
- [ ] State preserved (no data loss)
- [ ] Decisions documented
```

---

## Test Isolation & Independence

### Rule 1: No Test Dependencies
```
❌ BAD:
test('Create user', () => { userId = createUser(...); });
test('Update user', () => { updateUser(userId); }); // Depends on previous

✅ GOOD:
test('Create and update user', () => {
  const user = createUser(...);
  updateUser(user.id);
});
```

### Rule 2: Deterministic Tests
```
❌ BAD:
test('Process within timeout', async () => {
  const result = await slowOperation(); // Depends on system speed
  expect(result).toBeDefined();
});

✅ GOOD:
test('Process within timeout', async () => {
  const result = await mockFastOperation();
  expect(result).toBeDefined();
});
```

---

## Coverage Tools by Language

| Language | Tool | Command |
|----------|------|---------|
| **JavaScript/TypeScript** | Jest/Vitest | `npm test -- --coverage` |
| **Python** | pytest-cov | `pytest --cov` |
| **Java** | JaCoCo | `mvn clean test jacoco:report` |
| **Go** | Built-in | `go test -cover ./...` |
| **Rust** | tarpaulin | `cargo tarpaulin --out Html` |
| **C#** | Built-in | `dotnet test --collect:"XPlat Code Coverage"` |
| **Ruby** | SimpleCov | `bundle exec rspec` |

---

## CI/CD Test Gates

```yaml
# Sample GitHub Actions
- name: Run Tests
  run: npm test -- --coverage
  
- name: Check Coverage
  run: |
    COVERAGE=$(jq '.total.lines.pct' coverage/coverage-final.json)
    if (( $(echo "$COVERAGE < 90" | bc -l) )); then
      exit 1
    fi
```

---

## Anti-Patterns (NEVER Do This)

| Pattern | Issue | Fix |
|---------|-------|-----|
| **Test-Last** | Bugs not caught early | Always TDD (RED first) |
| **Mocking Database** | Queries not validated | Use real test DB |
| **Test Interdependencies** | One failure breaks others | Test isolation |
| **Flaky Tests** | Random failures | Make deterministic |
| **Testing Implementation** | Tests break on refactoring | Test behavior only |
| **No Error Scenario Tests** | Happy path only | Test errors explicitly |

---

*Testing Strategy v1.0.0*  
*Framework: Spec Kit Constitutional Development Framework*
