# Tasks: [FEATURE NAME]

**Input**: Design documents from `specs/[###-feature-name]/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)

```
1. Load plan.md from feature directory
   → If not found: ERROR "No implementation plan found"
   → Extract: tech stack, libraries, structure
2. Load optional design documents:
   → data-model.md: Extract entities → model tasks
   → contracts/: Each file → contract test task
   → research.md: Extract decisions → setup tasks
3. Generate tasks by category:
   → Setup: project init, dependencies, linting
   → Tests: contract tests, integration tests
   → Core: models, services, CLI commands
   → Integration: DB, middleware, logging
   → Polish: unit tests, performance, docs
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All contracts have tests?
   → All entities have models?
   → All endpoints implemented?
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `src/`, `tests/` at repository root
- **Web app**: `backend/src/`, `frontend/src/`
- **Mobile**: `api/src/`, `ios/src/` or `android/src/`
- Paths shown below assume single project - adjust based on plan.md structure

## Phase 3.1: Setup

- [ ] T001 Create project structure per implementation plan
- [ ] T002 Initialize [language] project with [framework] dependencies
- [ ] T003 [P] Configure linting and formatting tools

## Phase 3.2: Tests First (TDD) ⚠️ CRITICAL - CONSTITUTIONAL REQUIREMENT

**NON-NEGOTIABLE: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests (RED phase required)

- [ ] T004 [P] Contract test POST /api/users in tests/contract/test_users_post.py
  - VERIFY: Test fails with "endpoint not found" or equivalent
  - COMMIT: "RED: Add failing contract test for POST /api/users"
- [ ] T005 [P] Contract test GET /api/users/{id} in tests/contract/test_users_get.py
  - VERIFY: Test fails with "endpoint not found" or equivalent
  - COMMIT: "RED: Add failing contract test for GET /api/users/{id}"

### Integration Tests (RED phase required)

- [ ] T006 [P] Integration test user registration in tests/integration/test_registration.py
  - VERIFY: Test fails with missing implementation
  - COMMIT: "RED: Add failing integration test for user registration"
- [ ] T007 [P] Integration test auth flow in tests/integration/test_auth.py
  - VERIFY: Test fails with missing implementation
  - COMMIT: "RED: Add failing integration test for auth flow"

### TDD Verification Gate

- [ ] T007.1 MANDATORY: Run test suite and verify ALL tests in Phase 3.2 FAIL
- [ ] T007.2 MANDATORY: Git log shows "RED:" commits before any implementation
- [ ] T007.3 MANDATORY: No implementation code exists yet (only test files)

## Phase 3.3: Core Implementation (ONLY after tests are failing)

**FORBIDDEN until Phase 3.2 verification complete**

### Library Implementation (GREEN phase)

- [ ] T008 [P] User model in src/models/user.py
  - GOAL: Make T006, T007 pass
  - COMMIT: "GREEN: Implement user model to pass integration tests"
- [ ] T009 [P] UserService CRUD in src/services/user_service.py
  - GOAL: Make T006, T007 pass
  - COMMIT: "GREEN: Implement user service to pass integration tests"
- [ ] T010 [P] CLI --create-user in src/cli/user_commands.py
  - INCLUDES: --help, --version, --format=json|text flags
  - COMMIT: "GREEN: Add CLI interface for user operations"

### API Endpoints (GREEN phase)

- [ ] T011 POST /api/users endpoint
  - GOAL: Make T004 pass
  - INCLUDES: Input validation, error handling, structured logging
  - COMMIT: "GREEN: Implement POST /api/users endpoint"
- [ ] T012 GET /api/users/{id} endpoint
  - GOAL: Make T005 pass
  - INCLUDES: Error handling, structured logging
  - COMMIT: "GREEN: Implement GET /api/users/{id} endpoint"

### Security & Validation (GREEN phase)

- [ ] T013 Input validation and sanitization
  - INCLUDES: SQL injection prevention, XSS protection
  - COMMIT: "GREEN: Add comprehensive input validation"
- [ ] T014 Error handling and structured logging
  - INCLUDES: JSON logging format, request/response logging
  - COMMIT: "GREEN: Implement error handling and observability"

## Phase 3.4: Integration & Observability

- [ ] T015 Database connection and query optimization
  - INCLUDES: Connection pooling, parameterized queries
  - VERIFY: No N+1 queries, p95 < 200ms response time
- [ ] T016 Authentication middleware implementation
  - INCLUDES: JWT validation, CORS configuration
- [ ] T017 Monitoring and health checks
  - INCLUDES: /health endpoint, performance metrics collection
- [ ] T018 Security headers and HTTPS enforcement
  - INCLUDES: CSP, HSTS, X-Frame-Options headers

## Phase 3.5: Polish & Documentation

### Performance & Reliability

- [ ] T019 [P] Performance testing in tests/performance/
  - VERIFY: p95 < 200ms, memory < 100MB
  - INCLUDES: Load testing, stress testing
- [ ] T020 [P] Security testing in tests/security/
  - VERIFY: Input validation, authentication bypass prevention
- [ ] T021 [P] Unit tests for edge cases in tests/unit/
  - VERIFY: Error conditions, boundary values

### Documentation & Release

- [ ] T022 Library documentation (README.md + llms.txt format)
  - INCLUDES: CLI usage examples, API documentation
- [ ] T023 Version management and changelog
  - INCLUDES: Semantic versioning, migration notes
- [ ] T024 Deployment verification
  - INCLUDES: Blue-green deployment test, rollback procedures

## Dependencies

- Tests (T004-T007) before implementation (T008-T014)
- T008 blocks T009, T015
- T016 blocks T018
- Implementation before polish (T019-T023)

## Parallel Example

```
# Launch T004-T007 together:
Task: "Contract test POST /api/users in tests/contract/test_users_post.py"
Task: "Contract test GET /api/users/{id} in tests/contract/test_users_get.py"
Task: "Integration test registration in tests/integration/test_registration.py"
Task: "Integration test auth in tests/integration/test_auth.py"
```

## Notes

- [P] tasks = different files, no dependencies
- Verify tests fail before implementing
- Commit after each task
- Avoid: vague tasks, same file conflicts

## Task Generation Rules

_Applied during main() execution_

1. **From Contracts**:
   - Each contract file → contract test task [P]
   - Each endpoint → implementation task
2. **From Data Model**:
   - Each entity → model creation task [P]
   - Relationships → service layer tasks
3. **From User Stories**:

   - Each story → integration test [P]
   - Quickstart scenarios → validation tasks

4. **Ordering**:
   - Setup → Tests → Models → Services → Endpoints → Polish
   - Dependencies block parallel execution

## Validation Checklist

_GATE: Checked by main() before returning_

- [ ] All contracts have corresponding tests
- [ ] All entities have model tasks
- [ ] All tests come before implementation
- [ ] Parallel tasks truly independent
- [ ] Each task specifies exact file path
- [ ] No task modifies same file as another [P] task
