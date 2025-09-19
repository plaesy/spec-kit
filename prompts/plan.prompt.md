# Implementation Planning Workflow

Create detailed, execution-ready implementation plans from comprehensive feature specifications, ensuring constitutional framework compliance and optimal development workflow.

**Input**: {{$ARGUMENTS}} - Feature specification path or implementation requirements
**Output**: Complete implementation plan with architecture, tasks breakdown, and development roadmap

This is the third step in the Spec-Driven Development lifecycle.

Given the implementation details provided as {{$ARGUMENTS}}, do this:

## Phase 1: Setup and Analysis

1. **Initialize Planning Environment**:
   - Run platform-appropriate script from repo root and parse JSON for FEATURE_SPEC, IMPL_PLAN, SPECS_DIR, BRANCH:
     - **Linux/macOS**: `.plaesy/scripts/bash/create-new-plan.sh --json`
     - **Windows**: `.plaesy/scripts/powershell/create-new-plan.ps1 -Json`
   - Validate all file paths are absolute
   - Confirm script execution success before proceeding

2. **Load and Analyze Specification**:
   - **Feature Specification**: Read FEATURE_SPEC to extract:
     - Feature requirements and user stories with acceptance criteria
     - Functional and non-functional requirements with success metrics
     - Technical constraints, dependencies, and integration requirements
     - API contracts, database schemas, and external service integrations
     - Security requirements and compliance mandates
   - **Constitutional Framework**: Load `.plaesy/memory/constitution.md` for compliance requirements
   - **Domain Knowledge**: Scan relevant `.plaesy/memory/*.md` files for implementation patterns
   - **Template Foundation**: Load `.plaesy/templates/plan.template.md` for structure guidance

3. **Implementation Readiness Assessment**:
   - **Specification Completeness**: Verify all required specification components exist
   - **Technical Feasibility**: Validate architectural decisions and technology choices
   - **Resource Requirements**: Assess development effort, infrastructure needs, and timelines
   - **Risk Evaluation**: Identify implementation risks and mitigation strategies
   - **Dependency Mapping**: Chart critical path dependencies and integration points

## Phase 2: Architectural Planning

4. **Execute Constitutional Implementation Planning**:
   - **Load Plan Template**: Use `.plaesy/templates/plan.template.md` (copied to IMPL_PLAN path)
   - **Set Specification Input**: Configure template with FEATURE_SPEC path
   - **Execute Template Workflow**: Follow template's main workflow following constitutional principles
   - **Constitutional Compliance**: Ensure TDD structure, real dependencies, interface contracts
   - **Error Handling**: Follow template's error handling and quality gate checks
   - **Progress Tracking**: Update completion status for each template phase

5. **Generate Implementation Artifacts** (guided by template to $SPECS_DIR):
   - **Phase 0**: Generate `research.md` with technology research and architectural decisions
   - **Phase 1**: Generate `data-model.md`, `contracts/` directory, and `quickstart.md`
   - **Phase 2**: Generate detailed `tasks.md` with implementation breakdown
   - **Technical Context**: Incorporate user-provided details from {{$ARGUMENTS}}
   - **Quality Validation**: Verify each artifact meets constitutional standards

## Phase 3: Validation and Completion

6. **Implementation Plan Validation**:
   - **Progress Verification**: Confirm Progress Tracking shows all phases complete
   - **Artifact Completeness**: Ensure all required artifacts were generated successfully
   - **Error State Check**: Verify no ERROR states exist in execution
   - **Constitutional Compliance**: Validate plan follows TDD, real dependency testing, interface design
   - **Implementation Readiness**: Assess if development team can start immediately

7. **Handoff Package Preparation**:
   - **Deliverables Summary**: List all generated files with absolute paths
   - **Implementation Roadmap**: Provide clear development sequence and milestones
   - **Risk Mitigation Plan**: Detail identified risks with specific mitigation strategies
   - **Success Metrics**: Define measurable criteria for implementation success
   - **Next Phase Input**: Prepare structured input for tasks.prompt.md phase

## Quality Assurance Guidelines

### Constitutional Framework Compliance
- **TDD Structure**: Organize implementation plan to support test-first development
- **Real Dependencies**: Plan for integration testing with actual external systems (no mocks)
- **Interface Contracts**: Design clear API contracts with comprehensive error handling
- **Observability**: Embed monitoring, logging, and performance measurement throughout
- **Security First**: Integrate security considerations into every implementation phase
- **Platform Agnostic**: Ensure implementation works across different deployment environments

### Implementation Excellence
- **Incremental Delivery**: Structure plan for continuous integration and deployment
- **Risk Management**: Identify and mitigate technical, business, and operational risks
- **Performance Baseline**: Establish measurable performance targets and monitoring
- **Scalability Planning**: Design for growth in users, data, and feature complexity
- **Documentation Standards**: Plan for comprehensive user and developer documentation

### Development Workflow Optimization
- **Parallel Development**: Identify tasks that can be executed concurrently
- **Dependency Management**: Minimize blocking dependencies and critical path bottlenecks
- **Quality Gates**: Define checkpoints for code review, testing, and deployment
- **Automation Strategy**: Plan for CI/CD, testing, and deployment automation
- **Knowledge Transfer**: Structure for effective team collaboration and knowledge sharing

### Technical Standards
- **Code Quality**: Establish coding standards, review processes, and static analysis
- **Testing Strategy**: Plan unit, integration, contract, and end-to-end testing
- **Data Management**: Design for data integrity, backup, recovery, and compliance
- **Monitoring & Alerting**: Plan comprehensive observability and incident response
- **Security & Compliance**: Integrate OWASP principles and regulatory requirements

**File Path Management**: Use absolute paths with repository root for all file operations to avoid path issues and ensure consistent artifact generation.
