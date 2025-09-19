# Task Breakdown and Execution Planning

Transform implementation plans into specific, executable development tasks with constitutional framework compliance and optimal parallel execution strategy.

**Input**: {{$ARGUMENTS}} - Feature implementation plan or task specification
**Output**: Detailed task breakdown with dependencies, parallel execution guidance, and TDD structure

This is the fourth step in the Spec-Driven Development lifecycle.

Given the task or feature for implementation as {{$ARGUMENTS}}, do this:

## Phase 1: Prerequisites Analysis

1. **Environment Setup and Validation**:
   - Run platform-appropriate script from repo root and parse FEATURE_DIR and AVAILABLE_DOCS list:
     - **Linux/macOS**: `.plaesy/scripts/bash/check-task-prerequisites.sh --json`
     - **Windows**: `.plaesy/scripts/powershell/check-task-prerequisites.ps1 -Json`
   - Validate all paths are absolute
   - Confirm script execution success and available documents

2. **Implementation Plan Analysis**:
   - **Core Planning**: Always read `plan.md` for tech stack, libraries, and architecture decisions
   - **Data Architecture**: IF EXISTS, read `data-model.md` for entities, relationships, and schema design
   - **API Contracts**: IF EXISTS, read `contracts/` directory for API endpoints and interface specifications
   - **Research Decisions**: IF EXISTS, read `research.md` for technical decisions and technology choices
   - **Test Scenarios**: IF EXISTS, read `quickstart.md` for test scenarios and validation criteria
   - **Constitutional Framework**: Load `.plaesy/memory/constitution.md` for TDD and compliance requirements

3. **Project Context Assessment**:
   - **Project Type Identification**: Determine if CLI tool, web application, library, service, etc.
   - **Architecture Complexity**: Assess monolithic vs microservice vs serverless patterns
   - **Integration Requirements**: Identify external dependencies and service interactions
   - **Technology Stack Validation**: Confirm tech choices align with constitutional principles

   **Context-Aware Planning**:
   - CLI tools typically don't require `contracts/` but need command parsing and output formatting
   - Libraries focus on API design and documentation rather than database models
   - Web applications require full-stack considerations including frontend and backend tasks
   - Microservices emphasize contract testing and service communication

## Phase 2: Constitutional Task Generation

4. **Load Task Framework**:
   - Use `.plaesy/templates/tasks.template.md` as the structural foundation
   - Apply constitutional TDD principles to task ordering and dependencies
   - Incorporate domain-specific patterns from `.plaesy/memory/*.md` files
   - Ensure real dependency testing approach (no mocks in integration tests)

5. **Task Category Generation** (following constitutional TDD order):
   - **Setup Tasks**: Project initialization, dependencies, linting, constitutional compliance setup
   - **Test Infrastructure [P]**: Test frameworks, CI/CD pipeline, quality gates setup
   - **Contract Tests [P]**: One per API contract, service interface, external integration
   - **Unit Test Tasks [P]**: One per entity, service, utility function (written BEFORE implementation)
   - **Core Implementation**: One per entity, service, CLI command, endpoint (after tests exist)
   - **Integration Tasks**: Database connections, middleware, logging, monitoring
   - **End-to-End Tests [P]**: User story validation with real dependencies
   - **Performance & Security [P]**: Load testing, security scanning, compliance validation
   - **Documentation & Polish [P]**: API docs, user guides, deployment guides

6. **Constitutional Task Ordering Rules**:
   - **TDD Compliance**: ALL tests must be written before implementation (Red-Green-Refactor)
   - **Dependency Chain**: Setup → Test Infrastructure → Contract Tests → Unit Tests → Implementation → Integration → E2E → Polish
   - **Parallel Execution Strategy**:
     - Contract tests for different services → [P] (parallel)
     - Unit tests for different modules → [P] (parallel)
     - Implementation tasks affecting same files → Sequential (no [P])
     - Documentation tasks → [P] (parallel with final implementation)
     - Different domain boundaries → [P] (parallel)
   - **Real Dependencies**: Integration tests use actual databases, APIs, services (constitutional requirement)
   - **Interface First**: API contracts and schemas before implementation

7. **Task Specification Standards**:
   - **Task Numbering**: T001, T002, etc. with category prefixes (T-SETUP-001, T-TEST-001, T-IMPL-001)
   - **File Path Clarity**: Absolute paths for all files created or modified
   - **Dependency Mapping**: Explicit task dependencies with blocking relationships
   - **Acceptance Criteria**: Each task includes specific completion criteria
   - **Estimation**: Time estimates based on complexity (S/M/L/XL)
   - **Constitutional Validation**: Each task includes compliance checkpoints

## Phase 3: Execution Planning

8. **Parallel Execution Optimization**:
   - **Concurrent Task Groups**: Bundle [P] tasks that can execute simultaneously
   - **Resource Allocation**: Consider team size and skill distribution
   - **Critical Path Analysis**: Identify bottleneck tasks that block others
   - **Risk Mitigation**: Plan for task failures and recovery strategies
   - **Quality Gates**: Define checkpoints for code review and testing

9. **Generate Comprehensive Tasks Document**:
   - **Create**: `FEATURE_DIR/tasks.md` with complete task breakdown
   - **Include**:
     - Feature name and version from implementation plan
     - Constitutional framework compliance checklist
     - Numbered tasks with clear ownership and timelines
     - Dependency graph and critical path analysis
     - Parallel execution guidance with specific task agent commands
     - Quality gates and acceptance criteria for each task
     - Risk assessment and mitigation strategies
     - Integration points and external dependency management

## Quality Assurance Guidelines

### Constitutional Framework Compliance
- **TDD Enforcement**: Ensure tests are written before implementation in every task
- **Real Dependencies**: Plan integration tests with actual external systems (no mocks)
- **Interface Contracts**: Design tasks to implement clear API contracts with error handling
- **Observability**: Include monitoring, logging, and alerting in every implementation task
- **Security First**: Integrate security considerations into every development task
- **Platform Agnostic**: Ensure tasks work across different deployment environments

### Task Quality Standards
- **Atomic Tasks**: Each task should be completable in 1-4 hours by a skilled developer
- **Clear Acceptance Criteria**: Every task includes specific, measurable completion criteria
- **Context Independence**: Tasks include all necessary context for independent execution
- **Error Recovery**: Each task includes error handling and rollback procedures
- **Documentation**: All tasks include documentation updates and knowledge transfer

### Execution Optimization
- **Parallel Processing**: Maximize concurrent development through smart task dependencies
- **Resource Efficiency**: Optimize task allocation based on team skills and availability
- **Risk Management**: Include contingency planning for high-risk or complex tasks
- **Quality Gates**: Define review and validation checkpoints throughout execution
- **Continuous Integration**: Structure tasks for seamless CI/CD pipeline integration

### Team Collaboration
- **Knowledge Sharing**: Plan for effective handoffs between team members
- **Code Review**: Structure tasks to facilitate meaningful peer review
- **Progress Tracking**: Enable clear progress visibility and bottleneck identification
- **Communication**: Define communication protocols for task updates and blockers

**Context Integration**: Incorporate user-provided details from {{$ARGUMENTS}} into technical implementation context.

**Execution Readiness**: The generated tasks.md should be immediately executable - each task must be specific enough that an AI agent or human developer can complete it without additional context, following constitutional framework principles throughout.
