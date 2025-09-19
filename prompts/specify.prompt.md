# Feature Specification Creation Workflow

This creates comprehensive, implementation-ready technical specifications from researched ideas or natural language descriptions, following constitutional framework principles.

**Input**: {{$ARGUMENTS}} - Researched idea file path OR natural language feature description
**Output**: Complete technical specification with API contracts, database design, and implementation plan

Given the feature description provided as {{$ARGUMENTS}}, do this:

## Phase 1: Input Analysis & Setup

1. **Determine Input Type**:
   - If {{$ARGUMENTS}} references an existing idea file (contains `.md` or file path), treat as researched idea
   - Otherwise, treat as natural language feature description requiring new setup

2. **Setup Based on Input Type**:
   
   **For Researched Ideas (from idea.prompt.md)**:
   - Parse ARGUMENTS to extract idea file path 
   - Load the idea file to extract research findings, problem analysis, and solution approaches
   - Use existing branch and directory structure from idea phase
   - Extract BRANCH_NAME and SPECS_DIR from idea file metadata

   **For New Feature Descriptions**:
   - **Linux/macOS**: Run `.plaesy/scripts/bash/create-new-feature.sh --json "$ARGUMENTS"` from repo root
   - **Windows**: Run `.plaesy/scripts/powershell/create-new-feature.ps1 -Json "$ARGUMENTS"` from repo root
   - Parse JSON output for BRANCH_NAME and SPEC_FILE
   - All file paths must be absolute

## Phase 2: Research Integration & Constitutional Analysis

3. **Load Foundation Materials**:
   - **Templates**: Load from `.plaesy/templates/` (preferred) or fallback to `templates/`
   - **Spec Template**: Load `.plaesy/templates/spec.template.md` for required sections
   - **Constitutional Framework**: Load `.plaesy/memory/constitution.md` for compliance requirements
   - **Domain Knowledge**: Scan relevant `.plaesy/memory/*.md` files for domain-specific guidance
   - **SRS Template**: Load `.plaesy/templates/srs.template.md` for comprehensive requirements

4. **Research Integration** (if working from idea file):
   - Extract validated problem statement and user analysis with evidence
   - Integrate market research findings and competitive landscape analysis
   - Apply technical feasibility assessment and constraints
   - Incorporate stakeholder requirements and decision criteria
   - Validate success criteria and measurable KPIs
   - Review brainstorming insights and creative breakthrough solutions

5. **Constitutional Compliance Analysis**:
   - **Interface Contracts**: Plan for clear API contracts with validation
   - **Error Handling**: Design consistent error responses with meaningful context
   - **Versioning Strategy**: Plan semantic versioning with backward compatibility
   - **Testing Strategy**: Structure for TDD with real dependencies (no mocks in integration)
   - **Observability**: Plan structured logging, metrics, and health checks
   - **Security**: Embed OWASP compliance from specification phase

6. **Gap Analysis & Risk Assessment**:
   - Identify missing information critical for implementation
   - Flag areas requiring additional research or stakeholder validation
   - Assess technical risks and mitigation strategies
   - Prioritize missing elements by implementation impact
   - Estimate effort and complexity for each specification component

## Phase 3: Specification Development

5. **Create Comprehensive Specification**:
   - Write specification to SPEC_FILE using template structure
   - Replace placeholders with concrete details from research/analysis
   - Preserve section order and headings from template
   
   **Key Sections to Complete**:
   - **Problem Statement**: Use researched and validated problem definition
   - **User Stories & Scenarios**: Convert research into specific user stories with acceptance criteria
   - **Functional Requirements**: List specific capabilities based on solution approach
   - **Non-Functional Requirements**: Add performance, security, scalability needs from technical assessment
   - **System Architecture**: High-level design based on technical complexity analysis
   - **API Specifications**: Define interfaces and data contracts
   - **Security Requirements**: Authentication, authorization, data protection based on risk assessment
   - **Testing Strategy**: Test plans aligned with constitutional framework requirements
   - **Implementation Plan**: Phased approach based on complexity and dependencies
    - **Database Design**: Provide a database design artefact when the feature requires persistent storage. Include:
       - **ER Diagram**: Entity-relationship diagram showing main entities and relationships
       - **Schemas / DDL**: SQL CREATE TABLE statements (or equivalent for chosen DB) including columns, types, constraints, indexes, and foreign keys
       - **Indexes & Performance Notes**: Suggested indexes and rationale for read/write patterns
       - **Migrations Strategy**: How schema changes will be applied (migration tool, backward-compatibility notes)
       - **Seed / Example Data**: Minimal seed data for local testing and examples of typical rows
       - **Storage Location**: Save database files under `specs/<feature-name>/design/` (e.g., `specs/<feature-name>/design/schema.sql`, `erd.png`)

    - **External API Integrations**: When the feature depends on third-party APIs, include an integrations section with:
       - **API Catalog**: External endpoints used with base URL, purpose, and owner/contact
       - **Authentication & Permissions**: Required credentials, token types, scopes, rotation strategy
       - **Rate Limits & Quotas**: Expected limits and handling strategy (backoff, caching, queueing)
       - **Failure Modes & Fallbacks**: Retry strategy, circuit breaker recommendations, offline behavior
       - **Data Contracts**: Expected request/response shapes, transformation rules, and examples
       - **Security & Compliance**: PII handling, consent, and any legal considerations
       - **Storage Location**: Save integration docs under `specs/<feature-name>/integrations/` (e.g., `specs/<feature-name>/integrations/payments.md`)

    - **API Contract (Backend/Full-stack apps)**: Generate machine-readable OpenAPI 3.0 contract (YAML/JSON) including:
       - **Metadata**: Title, version, description matching spec requirements
       - **Paths/Endpoints**: Complete HTTP methods with paths, summaries, descriptions, and tags
       - **Schemas**: Reusable components with types, validation rules, required fields, and examples
       - **Parameters**: Path, query, header, cookie parameters with validation and examples
       - **Security**: Authentication schemes (JWT, OAuth2), authorization scopes, security requirements
       - **Error Handling**: Standardized error schemas for all HTTP status codes with examples
       - **Pagination**: Consistent patterns for list endpoints (cursor-based preferred over offset)
       - **Versioning**: API versioning strategy and backward compatibility plan
       - **Examples**: Complete request/response examples for all endpoints
       - **Rate Limiting**: API usage limits and throttling documentation
       - **Webhooks**: Event notification specifications if applicable
       - **Storage**: Save as `specs/<feature-name>/contracts/openapi.yaml` with absolute path reporting

6. **Quality Assurance**:
   - Ensure all research findings are properly integrated
   - Validate that specification addresses all identified user needs
   - Check that technical requirements align with feasibility assessment
   - Verify security and compliance requirements are addressed

## Phase 4: Research Gaps & Next Steps

7. **Document Remaining Unknowns**:
   - Clearly mark items requiring additional research with [NEEDS RESEARCH: specific question]
   - Identify dependencies on external stakeholders or technical validation
   - Suggest specific research activities to resolve gaps

8. **Integration Recommendations**:
   - Define how this feature integrates with existing systems
   - Specify data migration requirements if applicable
   - Identify potential breaking changes and mitigation strategies

## Phase 5: Validation & Handoff

9. **Comprehensive Validation**:
   - **Completeness Check**: Verify all required specification sections are complete
   - **Constitutional Compliance**: Validate TDD structure, real dependency testing, interface contracts
   - **Implementation Readiness**: Ensure developers can start coding immediately
   - **Stakeholder Alignment**: Confirm specification addresses original problem and success criteria
   - **Risk Assessment**: Validate that identified risks have mitigation strategies
   - **Quality Gates**: Check that all artifacts meet constitutional framework standards

10. **Handoff Package Generation**:
   - **Artifacts Inventory**: List all generated files with absolute paths
     - Primary specification file
     - API contract (OpenAPI) if applicable
     - Database design files and schemas
     - Integration documentation
     - Software Requirements Specification (SRS)
     - Security threat model
     - Performance requirements
   - **Readiness Assessment**: Rate implementation readiness (1-10 scale with justification)
   - **Implementation Recommendations**: Suggest optimal development sequence and priorities
   - **Risk Summary**: Highlight top 3 implementation risks with mitigation plans
   - **Success Metrics**: Define measurable criteria for implementation success
   - **Next Phase Input**: Prepare structured input for plan.prompt.md phase

## Additional Engineering Artefacts (optional but recommended)

- **Acceptance Criteria**: Comprehensive test specifications for each user story with constitutional TDD compliance:
   - **Story-Level Criteria**: Given-When-Then format for each user story with edge cases
   - **Component-Level Tests**: Unit test specifications for individual components
   - **Integration Test Plans**: End-to-end scenarios with real dependencies (no mocks)
   - **Contract Tests**: API contract validation between services
   - **Performance Acceptance**: Quantified performance criteria with measurement methods
   - **Security Test Cases**: Authentication, authorization, and data protection validation
   - **Accessibility Requirements**: WCAG compliance and usability test criteria
   - **Storage**: Save as `specs/<feature-name>/acceptance.md` with test priority matrix

- **Test Plans & Fixtures**: Provide a testing matrix mapping unit, integration, end-to-end, and contract tests to stories and components. Add example fixtures and test data under `specs/<feature-name>/tests/`.

- **CI/CD Requirements**: Describe required pipeline checks (lint, unit, integration, contract tests, security scans) and provide example pipeline snippets in `specs/<feature-name>/ci/` (e.g., `.github/workflows/` or pipeline YAML).

- **Observability**: Define metrics, SLIs/SLOs, logs, and tracing needs. Provide dashboard and alerting guidance in `specs/<feature-name>/ops/` along with suggested thresholds and runbook links.

- **Runbook & Incident Playbook**: Provide operational runbook for common failures, rollback steps, escalation contacts, and severity mapping. Save as `specs/<feature-name>/ops/runbook.md`.

- **Deployment & Infrastructure-as-Code**: Include example IaC snippets and recommended resources under `specs/<feature-name>/infra/` with notes on required cloud services and permissions.

- **Data Migration & Rollback Scripts**: For schema changes or data transformations, include migration scripts and rollback plans under `specs/<feature-name>/design/migrations/` with versioning notes.

- **Security Threat Model / Compliance**: Provide a threat model, required audits, PII handling, retention and access controls. Save under `specs/<feature-name>/security/`.

- **Mocking & API Consumers**: Provide Postman collections, mock server configs, or example client snippets for API consumers under `specs/<feature-name>/contracts/mocks/`.

- **Performance Requirements**: Define baseline performance targets, load testing approach, and expected capacity. Save as `specs/<feature-name>/performance.md` and include scripts or k6/jmeter examples if available.

- **Versioning & Deprecation Policy**: State API versioning strategy and deprecation timelines. Optionally include versioned filenames like `openapi.v1.yaml`.

- **Operational Cost Estimate**: Provide a rough cost estimate for infrastructure, storage, and third-party API usage under `specs/<feature-name>/ops/costs.md`.

- **Software Requirements Specification (SRS)**: For applications or large features, generate comprehensive SRS using `.plaesy/templates/srs.template.md`:
   - **Document Structure**: Follow IEEE 830 standard with constitutional framework enhancements
   - **Scope & Purpose**: Clear system boundaries, stakeholders, objectives, and success criteria
   - **Functional Requirements**: Numbered, traceable requirements with acceptance criteria (Given-When-Then format)
   - **Non-Functional Requirements**: Quantified performance, security, reliability, usability, and compliance requirements
   - **Quality Attributes**: Specific, measurable quality targets with verification methods
   - **Interface Specifications**: Detailed API contracts, database schemas, and external integration requirements
   - **Data Requirements**: Data models, flow diagrams, storage, backup, and retention policies
   - **Security Requirements**: Authentication, authorization, encryption, audit, and compliance specifications
   - **Traceability Matrix**: Requirements mapped to user stories, test cases, and implementation components
   - **Verification Methods**: Test strategies for each requirement type with constitutional TDD compliance
   - **Glossary & Standards**: Domain terminology and applicable technical standards
   - **Assumptions & Dependencies**: Critical assumptions and external dependencies with risk assessment
   - **Storage**: Save as `specs/<feature-name>/docs/srs.md` with metadata for tracking and approval

## Quality Assurance Guidelines

### Evidence-Based Specification
- **Research Integration**: Leverage all available research for evidence-based technical decisions
- **Validation Requirements**: Ensure all specifications are testable and measurable
- **Traceability**: Maintain clear links from user needs through requirements to implementation
- **Stakeholder Alignment**: Validate specifications against business objectives and user pain points

### Technical Excellence
- **Implementation Ready**: Create specifications detailed enough for immediate development start
- **Architecture Coherence**: Ensure all components work together as a unified system
- **Scalability Planning**: Design for growth in users, data, and feature complexity
- **Performance Baseline**: Establish measurable performance targets and monitoring

### Constitutional Framework Compliance
- **TDD Structure**: Organize specifications to support test-first development
- **Interface Design**: Define clear contracts with comprehensive error handling
- **Real Dependencies**: Plan integration testing with actual external systems
- **Observability**: Embed monitoring, logging, and alerting from the specification
- **Security by Design**: Integrate OWASP compliance throughout all specifications
- **Platform Agnostic**: Ensure specifications work across different deployment environments

### Completeness Validation
- **API Contracts**: Machine-readable OpenAPI specifications for all endpoints
- **Database Design**: Complete schema with relationships, indexes, and migration plans
- **Integration Mapping**: Detailed external service integration requirements
- **Security Model**: Comprehensive authentication, authorization, and data protection
- **Operational Requirements**: Deployment, monitoring, and maintenance specifications
- **Documentation Standards**: User-facing and developer documentation requirements

### Risk Management
- **Technical Risk Assessment**: Identify and mitigate implementation challenges
- **Dependency Analysis**: Map external dependencies and failure scenarios
- **Performance Constraints**: Identify potential bottlenecks and optimization needs
- **Security Vulnerabilities**: Proactive threat modeling and mitigation planning

Note: This phase transforms researched ideas into actionable, constitutionally-compliant technical specifications. When working from idea.prompt.md output, leverage all research findings and creative insights to create comprehensive, well-founded specifications ready for immediate implementation.
