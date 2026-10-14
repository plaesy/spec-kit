# Solution Architect Validation Checklist

This checklist serves as a comprehensive framework for the Architect to validate the technical design and architecture before development execution. The Architect should systematically work through
each item, ensuring the architecture is robust, scalable, secure, and aligned with the product requirements.

It is also the **technical** counterpart to `pm.checklist.md` (PRD and epic structure) and
`po.checklist.md` (backlog, stories, acceptance criteria, value delivery): project
scaffolding, dependency and environment setup, database and service configuration,
deployment pipeline, testing infrastructure, third-party and external API integration, and
design-system setup are owned **here**, not in the product checklists.

[[LLM: Required inputs: architecture.md (docs/architecture.md), prd.md (docs/prd.md), frontend-architecture.md if UI project, system diagrams, API docs, tech stack/versions — ask user for missing
ones. Detect project type (frontend present? frontend-architecture.md? PRD mentions UI?); if backend-only, skip [[FRONTEND ONLY]] sections and note skip in report, focus extra on
API/service/integration design. Also detect project type GREENFIELD (new) vs BROWNFIELD (change to an existing system — brownfield-architecture.md plus access to the existing codebase); skip
[[BROWNFIELD ONLY]] items that do not apply and note each skip. Validate each item with deep analysis, cite evidence from docs, question assumptions, assess risk — don't just check boxes. Ask user:
interactive (section-by-section with confirmation) or comprehensive (full report at end)?]]

## 1. REQUIREMENTS ALIGNMENT

[[LLM: Understand the PRD's core problem, users, and success factors first. Verify each item has a concrete technical solution, not just a mention.]]

### 1.1 Functional Requirements Coverage

- [ ] Architecture supports all functional requirements in the PRD
- [ ] Technical approaches for all epics and stories are addressed
- [ ] Edge cases and performance scenarios are considered
- [ ] All required integrations are accounted for
- [ ] User journeys are supported by the technical architecture

### 1.2 Non-Functional Requirements Alignment

- [ ] Performance requirements are addressed with specific solutions
- [ ] Scalability considerations are documented with approach
- [ ] Security requirements have corresponding technical controls
- [ ] Reliability and resilience approaches are defined
- [ ] Compliance requirements have technical implementations

### 1.3 Technical Constraints Adherence

- [ ] All technical constraints from PRD are satisfied
- [ ] Platform/language requirements are followed
- [ ] Infrastructure constraints are accommodated
- [ ] Third-party service constraints are addressed
- [ ] Organizational technical standards are followed

## 2. ARCHITECTURE FUNDAMENTALS

[[LLM: Check clarity as if explaining to a new developer — flag ambiguities that would confuse a human or AI implementer. Require actual diagrams, component definitions, and interaction patterns.]]

### 2.1 Architecture Clarity

- [ ] Architecture is documented with clear diagrams
- [ ] Major components and their responsibilities are defined
- [ ] Component interactions and dependencies are mapped
- [ ] Data flows are clearly illustrated
- [ ] Technology choices for each component are specified

### 2.2 Separation of Concerns

- [ ] Clear boundaries between UI, business logic, and data layers
- [ ] Responsibilities are cleanly divided between components
- [ ] Interfaces between components are well-defined
- [ ] Components adhere to single responsibility principle
- [ ] Cross-cutting concerns (logging, auth, etc.) are properly addressed

### 2.3 Design Patterns & Best Practices

- [ ] Appropriate design patterns are employed
- [ ] Industry best practices are followed
- [ ] Anti-patterns are avoided
- [ ] Consistent architectural style throughout
- [ ] Pattern usage is documented and explained

### 2.4 Modularity & Maintainability

- [ ] System is divided into cohesive, loosely-coupled modules
- [ ] Components can be developed and tested independently
- [ ] Changes can be localized to specific components
- [ ] Code organization promotes discoverability
- [ ] Architecture specifically designed for AI agent implementation

## 3. TECHNICAL STACK & DECISIONS

[[LLM: For each tech choice, judge simplicity vs over-engineering, scalability, maintenance cost, and version-specific security risk. Versions must be pinned, not ranges.]]

### 3.1 Technology Selection

- [ ] Selected technologies meet all requirements
- [ ] Technology versions are specifically defined (not ranges)
- [ ] Technology choices are justified with clear rationale
- [ ] Alternatives considered are documented with pros/cons
- [ ] Selected stack components work well together

### 3.2 Frontend Architecture [[FRONTEND ONLY]]

[[LLM: Skip if backend-only. Design-system and component setup is validated here, not in the product checklists.]]

- [ ] UI framework and libraries are specifically selected
- [ ] State management approach is defined
- [ ] Component structure and organization is specified
- [ ] Design system or component library is established before feature components are built
- [ ] Styling approach is defined (CSS modules, styled-components, utility classes, etc.) with tooling
- [ ] Responsive/adaptive design strategy is defined
- [ ] Build and bundling strategy is determined
- [ ] Frontend build pipeline is configured before feature development starts
- [ ] Asset optimization strategy is defined (images, fonts, code splitting)
- [ ] Component development workflow is defined (how a new component is added and reviewed)

### 3.3 Backend Architecture

- [ ] API design and standards are defined
- [ ] Service organization and boundaries are clear
- [ ] API framework and shared middleware are set up before endpoints are implemented
- [ ] Authentication framework is established before any protected route is added
- [ ] Error handling strategy is outlined
- [ ] Backend scaling approach is defined
- [ ] [[BROWNFIELD ONLY]] Existing API contracts remain compatible — no breaking change without a version bump and a migration plan
- [ ] [[BROWNFIELD ONLY]] Integration with the existing authentication mechanism is preserved

### 3.4 Data Architecture

- [ ] Data models are fully defined
- [ ] Database technologies are selected with justification
- [ ] Database selection and setup occur before any data operation depends on it
- [ ] Schema definitions are created before data operations are written
- [ ] Data access patterns are documented
- [ ] Data migration/seeding approach is specified
- [ ] Seed data or initial data setup is planned where required
- [ ] Data backup and recovery strategies are outlined
- [ ] [[BROWNFIELD ONLY]] Database migration risks are identified and mitigated
- [ ] [[BROWNFIELD ONLY]] Schema changes remain backward compatible, or a migration window is planned

## 4. FRONTEND DESIGN & IMPLEMENTATION [[FRONTEND ONLY]]

[[LLM: Skip if backend-only. Verify alignment between main and frontend-specific architecture docs.]]

### 4.1 Frontend Philosophy & Patterns

- [ ] Framework & Core Libraries align with main architecture document
- [ ] Component Architecture (e.g., Atomic Design) is clearly described
- [ ] State Management Strategy is appropriate for application complexity
- [ ] Data Flow patterns are consistent and clear
- [ ] Styling Approach is defined and tooling specified

### 4.2 Frontend Structure & Organization

- [ ] Directory structure is clearly documented with ASCII diagram
- [ ] Component organization follows stated patterns
- [ ] File naming conventions are explicit
- [ ] Structure supports chosen framework's best practices
- [ ] Clear guidance on where new components should be placed

### 4.3 Component Design

- [ ] Component template/specification format is defined
- [ ] Component props, state, and events are well-documented
- [ ] Shared/foundational components are identified
- [ ] Component reusability patterns are established
- [ ] Accessibility requirements are built into component design

### 4.4 Frontend-Backend Integration

- [ ] API interaction layer is clearly defined
- [ ] HTTP client setup and configuration documented
- [ ] Error handling for API calls is comprehensive
- [ ] Service definitions follow consistent patterns
- [ ] Authentication integration with backend is clear

### 4.5 Routing & Navigation

- [ ] Routing strategy and library are specified
- [ ] Route definitions table is comprehensive
- [ ] Route protection mechanisms are defined
- [ ] Deep linking considerations addressed
- [ ] Navigation patterns are consistent

### 4.6 Frontend Performance

- [ ] Image optimization strategies defined
- [ ] Code splitting approach documented
- [ ] Lazy loading patterns established
- [ ] Re-render optimization techniques specified
- [ ] Performance monitoring approach defined

## 5. RESILIENCE & OPERATIONAL READINESS

[[LLM: Apply Murphy's Law — peak load, critical service outage, 3am on-call diagnosability. Require specific resilience patterns, not just "error handling" mentions.]]

### 5.1 Error Handling & Resilience

- [ ] Error handling strategy is comprehensive
- [ ] Retry policies are defined where appropriate
- [ ] Circuit breakers or fallbacks are specified for critical services
- [ ] Graceful degradation approaches are defined
- [ ] System can recover from partial failures

### 5.2 Monitoring & Observability

- [ ] Logging strategy is defined
- [ ] Monitoring approach is specified
- [ ] Key metrics for system health are identified
- [ ] Alerting thresholds and strategies are outlined
- [ ] Debugging and troubleshooting capabilities are built in
- [ ] Usage or product analytics are specified where the product needs them
- [ ] [[BROWNFIELD ONLY]] Existing monitoring is preserved or deliberately extended, not replaced silently

### 5.3 Performance & Scaling

- [ ] Performance bottlenecks are identified and addressed
- [ ] Caching strategy is defined where appropriate
- [ ] Load balancing approach is specified
- [ ] Horizontal and vertical scaling strategies are outlined
- [ ] Resource sizing recommendations are provided

### 5.4 Deployment & DevOps

- [ ] Deployment strategy is defined
- [ ] CI/CD pipeline is established before any deployment action depends on it
- [ ] Environment strategy (dev, staging, prod) is specified, with per-environment configuration defined early
- [ ] Infrastructure as Code approach is defined and set up before it is used
- [ ] Rollback and recovery procedures are outlined, with triggers and thresholds
- [ ] Disaster recovery approach is defined — RTO/RPO targets, backup retention, and how often a restore is actually exercised
- [ ] Deployment minimizes downtime for the affected users
- [ ] Blue-green or canary deployment is used for high-risk releases, or its absence is justified
- [ ] Cloud resource provisioning is sequenced before the resources are needed
- [ ] DNS, domain, email/messaging, and CDN or static-asset hosting needs are identified and sequenced before first use
- [ ] [[BROWNFIELD ONLY]] Deployment preserves existing infrastructure services

## 6. SECURITY & COMPLIANCE

[[LLM: Review with a hacker's mindset — how could this be exploited? Check applicable regulations (GDPR/HIPAA/PCI). Require specific controls, not general statements.]]

### 6.1 Authentication & Authorization

- [ ] Authentication mechanism is clearly defined
- [ ] Authorization model is specified
- [ ] Role-based access control is outlined if required
- [ ] Session management approach is defined
- [ ] Credential management is addressed

### 6.2 Data Security

- [ ] Data encryption approach (at rest and in transit) is specified
- [ ] Sensitive data handling procedures are defined
- [ ] Data retention and purging policies are outlined
- [ ] Backup encryption is addressed if required
- [ ] Data access audit trails are specified if required

### 6.3 API & Service Security

- [ ] API security controls are defined
- [ ] Rate limiting and throttling approaches are specified
- [ ] Input validation strategy is outlined
- [ ] CSRF/XSS prevention measures are addressed
- [ ] Secure communication protocols are specified

### 6.4 Infrastructure Security

- [ ] Network security design is outlined
- [ ] Firewall and security group configurations are specified
- [ ] Service isolation approach is defined
- [ ] Least privilege principle is applied
- [ ] Security monitoring strategy is outlined

## 7. IMPLEMENTATION GUIDANCE

[[LLM: Imagine a developer's day one — do they have what they need to be productive, with concrete examples and consistent standards?]]

### 7.1 Coding Standards & Practices

- [ ] Coding standards are defined
- [ ] Documentation requirements are specified
- [ ] Testing expectations are outlined
- [ ] Code organization principles are defined
- [ ] Naming conventions are specified

### 7.2 Testing Strategy

- [ ] Testing frameworks are selected and installed before any test is written
- [ ] Unit testing approach is defined
- [ ] Integration testing strategy is outlined
- [ ] E2E testing approach is specified
- [ ] Test environment setup is defined before test implementation depends on it
- [ ] Test data requirements are defined — constitutional rule: real dependencies, no mocks in integration tests
- [ ] Performance testing requirements are outlined
- [ ] Security testing approach is defined
- [ ] [[BROWNFIELD ONLY]] Regression testing covers existing functionality
- [ ] [[BROWNFIELD ONLY]] Integration testing validates each new-to-existing connection

### 7.3 Frontend Testing [[FRONTEND ONLY]]

[[LLM: Skip if backend-only.]]

- [ ] Component testing scope and tools defined
- [ ] UI integration testing approach specified
- [ ] Visual regression testing considered
- [ ] Accessibility testing tools identified
- [ ] Frontend-specific test data management addressed

### 7.4 Development Environment

- [ ] Local development environment setup is documented
- [ ] Project scaffolding and initialization are defined — repository creation, initial commit, README, starter template or from-scratch steps
- [ ] Required tools, versions, and configurations are specified
- [ ] Development workflows are outlined, including the local development server
- [ ] Source control practices are defined (branching, review, merge)
- [ ] Dependency management approach is specified, with critical packages installed early
- [ ] Credential provisioning and secure storage approach is defined
- [ ] [[BROWNFIELD ONLY]] Local setup preserves the existing system and can run existing features for verification
- [ ] [[BROWNFIELD ONLY]] Version compatibility of new dependencies with the existing stack is verified

### 7.5 Technical Documentation

- [ ] API documentation standards are defined
- [ ] Architecture documentation requirements are specified
- [ ] Code documentation expectations are outlined
- [ ] System diagrams and visualizations are included
- [ ] Decision records for key choices are included
- [ ] [[BROWNFIELD ONLY]] Integration points with the existing system are documented in detail
- [ ] [[BROWNFIELD ONLY]] Existing-system knowledge and integration knowledge are captured for handoff

## 8. DEPENDENCY & INTEGRATION MANAGEMENT

[[LLM: For each dependency, weigh unavailability impact, patch currency, vendor lock-in, and contingency/fallback plans.]]

### 8.1 External Dependencies

- [ ] All external dependencies are identified, and critical ones are installed early
- [ ] Versions are pinned, not ranges, and conflicts between them are resolved or noted
- [ ] Versioning strategy for dependencies is defined
- [ ] Fallback approaches for critical dependencies are specified
- [ ] Licensing implications are addressed
- [ ] Update and patching strategy is outlined
- [ ] [[BROWNFIELD ONLY]] Compatibility with the existing dependency set is verified

### 8.2 Internal Dependencies

- [ ] Component dependencies are clearly mapped
- [ ] Build order dependencies are addressed — lower-level services are built before higher-level ones, shared components before their use
- [ ] Shared services and utilities are identified
- [ ] Circular dependencies are eliminated
- [ ] Versioning strategy for internal components is defined
- [ ] Cross-epic dependencies are ordered: no epic requires functionality from a later epic, and each epic leaves the system in a working state
- [ ] [[BROWNFIELD ONLY]] Integration points are tested at each step, and existing functionality is preserved throughout

### 8.3 Third-Party Integrations

- [ ] All third-party integrations are identified, including external APIs
- [ ] Integration approaches are defined
- [ ] Account creation and API key acquisition steps are identified, with a named owner for each
- [ ] Credentials are stored securely, never in source
- [ ] Authentication with third parties is addressed and sequenced before dependent features
- [ ] Error handling for integration failures is specified, with an offline or degraded-mode path where the project needs one
- [ ] Rate limits and quotas are considered
- [ ] [[BROWNFIELD ONLY]] Existing API dependencies and existing third-party integrations are maintained
- [ ] [[BROWNFIELD ONLY]] Impact on existing integrations is assessed

## 9. AI AGENT IMPLEMENTATION SUITABILITY

[[LLM: This may be implemented by AI agents — favor explicit over implicit. Check pattern consistency, minimized complexity, and risk of incorrect assumptions.]]

### 9.1 Modularity for AI Agents

- [ ] Components are sized appropriately for AI agent implementation
- [ ] Dependencies between components are minimized
- [ ] Clear interfaces between components are defined
- [ ] Components have singular, well-defined responsibilities
- [ ] File and code organization optimized for AI agent understanding

### 9.2 Clarity & Predictability

- [ ] Patterns are consistent and predictable
- [ ] Complex logic is broken down into simpler steps
- [ ] Architecture avoids overly clever or obscure approaches
- [ ] Examples are provided for unfamiliar patterns
- [ ] Component responsibilities are explicit and clear

### 9.3 Implementation Guidance

- [ ] Detailed implementation guidance is provided
- [ ] Code structure templates are defined
- [ ] Specific implementation patterns are documented
- [ ] Common pitfalls are identified with solutions
- [ ] References to similar implementations are provided when helpful
- [ ] Responsibilities are split between what only humans can do (account creation, purchasing, credential provisioning) and what agents automate, and each is assigned
- [ ] Configuration management ownership is explicit
- [ ] [[BROWNFIELD ONLY]] Existing user workflows are preserved or a migration is specified

### 9.4 Error Prevention & Handling

- [ ] Design reduces opportunities for implementation errors
- [ ] Validation and error checking approaches are defined
- [ ] Self-healing mechanisms are incorporated where possible
- [ ] Testing patterns are clearly defined
- [ ] Debugging guidance is provided

## 10. ACCESSIBILITY IMPLEMENTATION [[FRONTEND ONLY]]

[[LLM: Skip if backend-only.]]

### 10.1 Accessibility Standards

- [ ] Semantic HTML usage is emphasized
- [ ] ARIA implementation guidelines provided
- [ ] Keyboard navigation requirements defined
- [ ] Focus management approach specified
- [ ] Screen reader compatibility addressed

### 10.2 Accessibility Testing

- [ ] Accessibility testing tools identified
- [ ] Testing process integrated into workflow
- [ ] Compliance targets (WCAG level) specified
- [ ] Manual testing procedures defined
- [ ] Automated testing approach outlined

[[LLM: Generate a final validation report: (1) Executive Summary — readiness (H/M/L), critical risks, strengths, project type (greenfield/brownfield, frontend/backend) & sections evaluated; (2)
Section Analysis — pass rate per section, worst gaps, sections skipped; (3) Risk Assessment — top 5 risks, mitigations, timeline impact; (4) Recommendations — must-fix / should-fix / nice-to-have; (5)
AI Implementation Readiness — concerns, unclear areas, complexity hotspots; (6) Frontend-Specific Assessment if applicable — completeness, doc alignment, UI/UX coverage, component clarity; (7)
[BROWNFIELD ONLY] Integration Readiness — existing-system compatibility, data/API migration risk, rollback readiness, monitoring coverage. State explicitly that backlog, prioritization, and
value-metric concerns belong to `po.checklist.md` and are out of scope here rather than silently skipping them. Then offer deeper analysis of any flagged section.]]
