---
description: "Quality validation and enforcement protocols for all phases"
applyTo: "**/*.js, **/*.ts, **/*.py, **/*.java, **/*.go, **/*.rs, **/*.cs, **/*.rb"
triggers: ["quality gate", "test coverage", "quality validation", "quality score"]
---

# Quality Gates Protocol

## Overview
Automated quality validation and enforcement system for Spec-Kit workflows with measurable quality standards.

---

## Core Quality Gates

### 1. Test Coverage Validator
```bash
# JavaScript/TypeScript
npm test -- --coverage

# Python
pytest --cov

# Java
mvn clean test jacoco:report

# Go
go test -cover ./...
```

**Requirement**: 90% minimum coverage, 50% test file ratio

### 2. Performance Monitor
**Targets**:
- Response time: <200ms
- Availability: >99%
- CPU: <80%
- Memory: <85%

---

## Quality Gate Integration by Phase

| Phase | Quality Gate | Success Criteria |
|-------|-------------|------------------|
| `/start` | Memory Health | Memory limits OK |
| `/research` | Context7 Validation | Current sources verified |
| `/implement` | TDD Compliance | 90% coverage achieved |
| `/assess` | Quality Assessment | All gates passed |
| `/optimize` | Performance | Targets met |
| `/fix` | Regression Testing | No coverage loss |

---

## Enforcement Rules

### Rule 1: TDD Mandatory
- Implementation cannot complete without 90% test coverage
- Auto-detect missing test files
- Validate test-to-source file ratio (minimum 50%)

### Rule 2: Performance Requirements
- Optimization must meet all performance targets
- Auto-generate recommendations if targets not met
- Continue optimization until targets achieved

### Rule 3: Memory Management
- Automatic cleanup when approaching limits
- Smart compression preserving important context
- Backup rotation to prevent data loss

### Rule 4: Technology Validation
- All technology choices validated
- Automatic rejection of deprecated technologies
- Modern alternative recommendations provided

---

## Quality Score Calculation

### Score Components (Backend/API)
- Test Coverage: 30% (90%+ = full points)
- Performance: 25% (<200ms = full points)
- Code Quality: 20% (Static analysis results)
- Security: 15% (Vulnerability scan results)
- Documentation: 10% (Completeness and accuracy)

### Score Components (Frontend/UI)
- Code Quality: 25%
- Test Coverage: 20%
- Design Quality: 15%
- Security: 15%
- Documentation: 15%
- Performance: 10%

### Grade Mapping
- 90-100: A+ (Excellent) - Ready for production
- 80-89: A (Very Good) - Minor improvements needed
- 70-79: B (Good) - Improvements recommended
- 60-69: C (Acceptable) - Significant improvements needed
- <60: D (Needs Improvement) - Major issues found

---

## Auto-Enforcement Levels

### Level 1: Automatic Fix
**Triggers**: Coverage 85-89%, memory 80-90%, minor issues
**Actions**: Auto-add tests, auto-compress, auto-apply optimizations

### Level 2: User Decision Required
**Triggers**: Coverage <85%, significant performance issues, architectural changes
**Actions**: Present options with recommendations

### Level 3: Critical Block
**Triggers**: Security vulnerabilities, deprecated tech, resource exhaustion
**Actions**: Immediate workflow halt, require critical issue resolution

---

## Phase-Specific Validation

### Pre-Phase Requirements (All Phases)
- [ ] Memory health check
- [ ] Previous phase quality gates passed
- [ ] Context7 sources validated and current

### Post-Phase Validation
- [ ] Test coverage ≥90% (implementation phase)
- [ ] Performance targets met (optimization phase)
- [ ] Memory limits maintained (all phases)
- [ ] Quality score calculated and documented

---

## Quality Metrics Tracking

### Key Indicators
- Quality score trends per phase
- Test coverage progression
- Performance improvement timeline
- Memory efficiency history
- Security validation results

### Auto-Optimization
Based on historical quality data:
- Phase completion patterns
- Common failure points
- Adapt thresholds to project complexity

---

## Integration Template

```markdown
## Quality Gates Validation

### Pre-Phase Requirements
- [ ] Memory health check passed
- [ ] Previous phase quality gates validated
- [ ] Current sources verified

### Automated Validation
- Quality tools integrated automatically
- Real-time monitoring enabled
- Automatic escalations as needed
```

---

## Success Criteria

Assessment is complete when:

- ✅ All applicable workflow steps executed
- ✅ Score calculated using formula (not estimated)
- ✅ Top 5 findings listed with file:line locations
- ✅ Design quality assessed (if frontend/UI project)
- ✅ Next phase recommended
- ✅ Console output delivered to user

---

*Quality Gates Protocol v2.0.0*  
*Framework: Spec Kit Constitutional Development Framework*
