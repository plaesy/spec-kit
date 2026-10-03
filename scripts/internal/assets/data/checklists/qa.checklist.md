# Quality Assurance Master Checklist

This checklist validates that a feature or release is actually tested to the standard
the project claims — coverage, test strategy, security, performance, and the evidence for
the sign-off decision. It is the **quality** checklist: test strategy and test types are
validated here, while system design and technology choices belong to `sa.checklist.md`,
testability of the spec belongs to `po.checklist.md`, and pipeline security gates belong to
`@devsecops`.

[[LLM: Required inputs: the spec/plan/tasks for the change under test, the test plan or testing strategy, the constitution's quality rules, the CI configuration (for what the pipeline actually
enforces), and the code under test — derive whatever is missing from the repo itself (read the code, the CI config, prior test runs) and label any gap `ASSUMED — <default>` rather than
leaving it unresolved. Detect project type: GREENFIELD (new feature, nothing in production yet) or BROWNFIELD (change to a live
system); for brownfield, regression against existing behaviour is a gate item, not an optional extra. Detect whether a UI exists (frontend-architecture.md, component tree) to decide if
the browser/device items apply; if backend-only, skip them and note the skip. Work every item top to bottom and cite the artifact or test-run output that satisfies it — an item with no
evidence is not checked. Default to comprehensive mode (full report at end) unless the user has already stated a preference for interactive (section-by-section).]]

## 1. REQUIREMENTS READINESS

[[LLM: You cannot test what was never made testable. If acceptance criteria are ambiguous, that is a finding against the spec — report it to the Product Owner rather than guessing the intent.]]

- [ ] Every requirement in scope has acceptance criteria that a tester can decide pass/fail from
- [ ] Acceptance criteria cover the error path and the boundary values, not only the happy path
- [ ] Business rules to be verified are enumerated
- [ ] Integration points in scope are identified
- [ ] Non-functional expectations in scope (limits, permissions, data volumes) are stated
- [ ] Out-of-scope behaviour is stated, so a failure there is not counted against this change
- [ ] Every in-scope requirement is traceable to at least one test
- [ ] Open questions have answers, not assumptions

## 2. CODE QUALITY

[[LLM: Code quality gates the tests: a suite that passes on unreviewable code is not evidence. Record the actual value for each metric — the number belongs in your report, not as a blank in
this file.]]

- [ ] Code follows the project's established coding standards
- [ ] Code is formatted by the project's formatter, with no unformatted files
- [ ] Naming conventions are followed
- [ ] Every change has been peer-reviewed, and review comments are resolved or explicitly deferred
- [ ] No new linter errors or warnings are introduced
- [ ] No hardcoded secrets, credentials, or environment-specific values
- [ ] Public interfaces expose clear contracts with validation on the input side
- [ ] Error responses are consistent and carry enough context to diagnose from the message alone
- [ ] Cyclomatic complexity per function is within the project's limit (default ≤10)
- [ ] Duplicated logic is within the project's limit (default ≤3%)

## 3. TEST COVERAGE

[[LLM: Coverage is a floor, not a goal — a suite that hits 95% of lines and misses the error path proves nothing. Check the constitution's stated target (default ≥90%) and report the measured
value.]]

- [ ] Unit test coverage meets the project's target (`{{TEST_COVERAGE_MIN|90}}` of changed code, unless the constitution sets a different minimum)
- [ ] Every acceptance criterion has at least one test that fails when the behaviour regresses
- [ ] Negative and error-path tests exist for each in-scope input
- [ ] Boundary values are tested, not just representative values
- [ ] Integration tests exist for each cross-component boundary in scope
- [ ] End-to-end tests cover the critical user paths in scope
- [ ] Tests use real dependencies — no mocks in integration tests (constitutional rule)
- [ ] Tests are deterministic: no order dependency, no time or network flakiness left unexplained
- [ ] Test data setup is automated, not a manual step a new developer must rediscover
- [ ] Flaky or quarantined tests are listed with the reason and an owner

## 4. SECURITY & OWASP

[[LLM: Do not restate the OWASP Top 10 here. Load `instructions/security-audit.instructions.md` for the audit protocol and `instructions/security-and-owasp.instructions.md` for the
per-language secure-coding rules, then verify that the change is covered against the categories that apply to it. Security findings are escalated, not waived inside this checklist.]]

- [ ] The OWASP Top 10 categories relevant to this change have been reviewed, and the review is recorded
- [ ] Authentication and authorization are tested, including negative cases (unauthenticated, wrong role, expired session)
- [ ] Authorization is enforced server-side on every protected endpoint, not only in the UI
- [ ] Input validation and output encoding are tested at each trust boundary
- [ ] Injection paths (SQL, command, template, LDAP) are tested with malicious input
- [ ] Sensitive data is protected in transit and at rest, and is absent from logs, errors, and test fixtures
- [ ] Secrets are managed by the project's secret store, never committed
- [ ] Dependency vulnerabilities are scanned, and no known critical or high finding ships unaddressed or unexpired
- [ ] Security-relevant configuration is hardened (security headers, CORS, cookie flags, debug modes off)
- [ ] OWASP-relevant tooling runs in CI, and a failure blocks the merge
- [ ] Findings are reported to `@security` for severity confirmation and to `@devsecops` where a control must be enforced automatically in the pipeline
- [ ] Any accepted risk has a named owner, a written justification, and an expiry date

## 5. PERFORMANCE

[[LLM: Compare against the targets stated in the spec or constitution, not against whatever the current numbers happen to be. A regression against a prior measurement is a finding even if the
absolute number passes.]]

- [ ] Response time targets are met at the stated percentile (default: p95 within the stated budget)
- [ ] Throughput and concurrency requirements are met under expected load
- [ ] Load testing has been run at or above the expected production load
- [ ] Stress/soak testing has exposed the breaking point, and it is documented
- [ ] Resource utilization (CPU, memory, connections) is within limits under load
- [ ] Performance has not regressed against the previous measurement
- [ ] Identified bottlenecks have a fix or a documented, accepted mitigation
- [ ] Scalability assumptions hold — the measured behaviour matches what the architecture promised

## 6. TEST STRATEGY & ENVIRONMENT

[[LLM: Strategy and environment are preconditions for the results in §3 and §5. If the environments are not equivalent to production, say so in the report — results are only as valid as the
environment.]]

- [ ] Test strategy is documented, with scope and out-of-scope explicitly stated
- [ ] Test levels in use (unit, integration, contract, E2E) are defined and their purpose is clear
- [ ] Test environment is provisioned and its parity with production is documented
- [ ] Test data requirements are met, including fixtures for the acceptance criteria in scope
- [ ] Test environments are isolated from production data, or the handling of production data is justified
- [ ] Test execution is reproducible from a clean checkout with one documented command
- [ ] Tests are wired into CI and a failure blocks the pipeline
- [ ] Test reporting produces a result a non-tester can read

## 7. FUNCTIONAL & ACCEPTANCE TESTING

[[LLM: This is the section that proves the change does what the spec said. Walk each acceptance criterion and name the test that covers it; a criterion with no named test fails this section.]]

- [ ] Every feature in scope is tested against its acceptance criteria
- [ ] Each primary user workflow in scope is executed end to end
- [ ] Each business rule in scope has at least one positive and one negative test
- [ ] Each integration point in scope is exercised, including failure of the remote side
- [ ] Data is persisted and read back correctly, including the migration path
- [ ] User acceptance testing is planned with the named stakeholder, or explicitly waived with a reason
- [ ] Feedback from UAT is either incorporated or rejected with a recorded reason
- [ ] [[UI ONLY]] Usability findings from the walkthrough are recorded and routed to `@designer`
- [ ] [[UI ONLY]] Accessibility requirements are validated per `sa.checklist.md` §10 and `@accessibility`

## 8. NON-FUNCTIONAL TESTING

[[LLM: Test the support matrix the project actually claims, not the matrix you wish it had. Skipped items must be reported as skipped, never quietly checked.]]

- [ ] [[UI ONLY]] Browser support matrix is tested against the project's stated support policy
- [ ] [[UI ONLY]] Device and viewport sizes in the support matrix are tested
- [ ] [[UI ONLY]] Operating system and browser version combinations in the support matrix are tested
- [ ] [[BACKEND ONLY]] API contract tests are run against the versioned interface contract
- [ ] [[BACKEND ONLY]] Retry, timeout, and partial-failure behaviour is tested
- [ ] Accessibility is validated with real assistive technology, not only an automated scanner
- [ ] [[BROWNFIELD ONLY]] Regression tests cover the existing functionality this change could affect
- [ ] [[BROWNFIELD ONLY]] Each new-to-existing connection is integration tested
- [ ] [[BROWNFIELD ONLY]] Backward compatibility of the public interface is verified

## 9. DEPLOYMENT VALIDATION

[[LLM: Deployment validation confirms the tested artifact is the shipped artifact. Pipeline and infrastructure design belong to `sa.checklist.md` §5.4; this section only checks the validation
performed.]]

- [ ] The artifact tested is the artifact that will be deployed (same commit, same build)
- [ ] Deployment and rollback procedures have been exercised at least once, not only documented
- [ ] Database migration scripts have been run against a copy of production-shaped data
- [ ] Configuration for the target environment is verified and differs from production only where intended
- [ ] Smoke tests run immediately after deploy and pass
- [ ] The critical path is validated in the deployed environment
- [ ] Monitoring and error reporting are active before traffic is allowed
- [ ] [[BROWNFIELD ONLY]] Rollback has been verified, and the rollback trigger and threshold are known to the person on call
- [ ] [[BROWNFIELD ONLY]] Existing users are not disrupted — affected workflows were checked post-deploy

## 10. DEFECTS & QUALITY RISK

[[LLM: A release decision made with an open critical defect is a decision, not an oversight. List what is open, who owns it, and what it blocks.]]

- [ ] Every defect found in this pass is logged with severity, area, and owner
- [ ] No open defect is Critical or High severity, or each has a written, expiring risk acceptance
- [ ] Each defect has a named owner and a due date, not just a severity label
- [ ] Defects deferred to a later release are recorded as visible backlog items
- [ ] The high-risk areas of this change are named, with why they are risky
- [ ] Each named risk has a mitigation and a trigger that would escalate it
- [ ] Test debt introduced by this change (skipped, quarantined, or missing tests) is recorded
- [ ] Residual risk after mitigation is stated explicitly for the sign-off decision

## QUALITY GATES

[[LLM: All four gates must pass for sign-off. State the measured value behind each gate, not a tick. If a gate cannot be evaluated because evidence is missing, the gate fails — an unmeasured
gate is not a passed gate.]]

- [ ] **Requirements Gate** — every in-scope requirement is traceable to a test, and no acceptance criterion is untestable
- [ ] **Test Gate** — coverage target met, integration tests use real dependencies, all tests pass and are deterministic, CI enforces both
- [ ] **Security Gate** — the applicable OWASP categories are reviewed and recorded, with no unaccepted Critical or High finding
- [ ] **Operational Gate** — the deployed artifact is validated with smoke tests, monitoring is active, and rollback is verified

### Sign-Off Decision

- [ ] All four gates above passed, and the measured evidence is attached
- [ ] Every open defect is either fixed or covered by a written, expiring risk acceptance with a named owner
- [ ] The person signing off is not the person who implemented the change
- [ ] Residual risk and known limitations are stated in the release note, not only in this report
- [ ] The decision is one of: **PROCEED**, **PROCEED WITH CONDITIONS** (conditions named and owned), or **HOLD** (blocking items named)

## REPORT

[[LLM: Produce: (1) Executive Summary — gate status per gate, measured coverage, open defect counts by severity, GO/NO-GO; (2) Section Analysis — pass rate per section, worst gaps, sections
skipped and why; (3) Security — OWASP categories reviewed, findings, accepted risks with expiry; (4) Performance — measured vs target, regressions; (5) Defect Summary — by severity and
area, with owners and due dates; (6) Recommendations — must-fix / should-fix / follow-up. Cross-reference `sa.checklist.md` for anything architectural that testing exposed, and route it
to `@sa`; route security findings to `@security`; do not fix architecture or security design inside this checklist. Then offer a deeper pass on any failed section.]]

### Category Statuses

| Category | Status | Critical Issues |
|----------|--------|-----------------|
| 1. Requirements Readiness | _TBD_ | |
| 2. Code Quality | _TBD_ | |
| 3. Test Coverage | _TBD_ | |
| 4. Security & OWASP | _TBD_ | |
| 5. Performance | _TBD_ | |
| 6. Test Strategy & Environment | _TBD_ | |
| 7. Functional & Acceptance Testing | _TBD_ | |
| 8. Non-Functional Testing | _TBD_ | |
| 9. Deployment Validation | _TBD_ | |
| 10. Defects & Quality Risk | _TBD_ | |
| Quality Gates | _TBD_ | |

### Critical Deficiencies

(To be populated during validation)

### Recommendations

(To be populated during validation)
