# [PROJECT NAME] Development Guidelines

Auto-generated from all feature plans. Last updated: [DATE]

## Constitutional Principles

### Library-First Architecture

- Every feature MUST be implemented as a standalone library
- Libraries MUST be self-contained, testable, and documented
- CLI interface required for all libraries (--help, --version, --format)
- No wrapper abstractions - use frameworks directly

### Test-First Development (NON-NEGOTIABLE)

- TDD strictly enforced: RED → GREEN → REFACTOR
- Tests must FAIL before implementation begins
- Test order: Contract → Integration → E2E → Unit
- Real dependencies required (no mocks for external services)
- Git commits must show tests before implementation

### Security & Performance Standards

- p95 response time < 200ms for CRUD operations
- Memory usage < 100MB base per service
- HTTPS mandatory, input validation at all boundaries
- Structured logging (JSON format) for all operations
- Security headers: CSP, HSTS, X-Frame-Options

### Radical Simplicity

- Maximum 3 projects per repository
- Avoid enterprise patterns without proven scaling need
- Single data model per domain
- YAGNI principle strictly enforced

## Active Technologies

[EXTRACTED FROM ALL PLAN.MD FILES]

## Project Structure

```
[ACTUAL STRUCTURE FROM PLANS]
```

## Development Commands

[ONLY COMMANDS FOR ACTIVE TECHNOLOGIES]

## Code Standards

### Universal Requirements

- Constitution compliance verification in all PRs
- Semantic versioning (MAJOR.MINOR.PATCH with BUILD increments)
- Breaking changes require migration plans
- Code coverage: 85% libraries, 70% applications

### Language-Specific Guidelines

[LANGUAGE-SPECIFIC, ONLY FOR LANGUAGES IN USE]

## Quality Gates

### Pre-Implementation

- [ ] Feature specification complete (no [NEEDS CLARIFICATION])
- [ ] Constitutional compliance verified
- [ ] Test plan defined (Contract → Integration → E2E → Unit)
- [ ] Performance targets specified
- [ ] Security requirements addressed

### Implementation Phase

- [ ] TDD cycle enforced (RED commits before GREEN)
- [ ] Real dependencies used in integration tests
- [ ] CLI interface implemented with standard flags
- [ ] Structured logging included
- [ ] Input validation and error handling complete

### Pre-Merge

- [ ] All tests passing (including performance benchmarks)
- [ ] Code coverage targets met
- [ ] Security scanning passed
- [ ] Documentation updated (README + llms.txt)
- [ ] Version incremented appropriately

## Recent Changes

[LAST 3 FEATURES AND WHAT THEY ADDED]

## Constitutional Enforcement

### Automated Checks

- TDD verification in CI/CD pipeline
- Performance regression testing
- Security vulnerability scanning
- Constitutional compliance validation

### Manual Reviews

- Architecture reviews for complexity justification
- Security reviews for auth/authorization changes
- Performance reviews for critical path modifications

<!-- MANUAL ADDITIONS START -->
<!-- Add project-specific guidelines, exceptions, or additional requirements here -->
<!-- These additions will be preserved during automatic updates -->
<!-- MANUAL ADDITIONS END -->