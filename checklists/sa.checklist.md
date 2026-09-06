# Solution Architect Validation Checklist

This checklist serves as a comprehensive framework for the Architect to validate the technical design and architecture before development execution. The Architect should systematically work through each item, ensuring the architecture is robust, scalable, secure, and aligned with the product requirements.

[[LLM: Required inputs: architecture.md (docs/architecture.md), prd.md (docs/prd.md), frontend-architecture.md if UI project, system diagrams, API docs, tech stack/versions — ask user for missing ones. Detect project type (frontend present? frontend-architecture.md? PRD mentions UI?); if backend-only, skip [[FRONTEND ONLY]] sections and note skip in report, focus extra on API/service/integration design. Validate each item with deep analysis, cite evidence from docs, question assumptions, assess risk — don't just check boxes. Ask user: interactive (section-by-section with confirmation) or comprehensive (full report at end)?]]

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

[[LLM: Skip if backend-only.]]

- [ ] UI framework and libraries are specifically selected
- [ ] State management approach is defined
- [ ] Component structure and organization is specified
- [ ] Responsive/adaptive design approach is outlined
- [ ] Build and bundling strategy is determined

### 3.3 Backend Architecture

- [ ] API design and standards are defined
- [ ] Service organization and boundaries are clear
- [ ] Authentication and authorization approach is specified
- [ ] Error handling strategy is outlined
- [ ] Backend scaling approach is defined

### 3.4 Data Architecture

- [ ] Data models are fully defined
- [ ] Database technologies are selected with justification
- [ ] Data access patterns are documented
- [ ] Data migration/seeding approach is specified
- [ ] Data backup and recovery strategies are outlined

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

### 5.3 Performance & Scaling

- [ ] Performance bottlenecks are identified and addressed
- [ ] Caching strategy is defined where appropriate
- [ ] Load balancing approach is specified
- [ ] Horizontal and vertical scaling strategies are outlined
- [ ] Resource sizing recommendations are provided

### 5.4 Deployment & DevOps

- [ ] Deployment strategy is defined
- [ ] CI/CD pipeline approach is outlined
- [ ] Environment strategy (dev, staging, prod) is specified
- [ ] Infrastructure as Code approach is defined
- [ ] Rollback and recovery procedures are outlined

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

- [ ] Unit testing approach is defined
- [ ] Integration testing strategy is outlined
- [ ] E2E testing approach is specified
- [ ] Performance testing requirements are outlined
- [ ] Security testing approach is defined

### 7.3 Frontend Testing [[FRONTEND ONLY]]

[[LLM: Skip if backend-only.]]

- [ ] Component testing scope and tools defined
- [ ] UI integration testing approach specified
- [ ] Visual regression testing considered
- [ ] Accessibility testing tools identified
- [ ] Frontend-specific test data management addressed

### 7.4 Development Environment

- [ ] Local development environment setup is documented
- [ ] Required tools and configurations are specified
- [ ] Development workflows are outlined
- [ ] Source control practices are defined
- [ ] Dependency management approach is specified

### 7.5 Technical Documentation

- [ ] API documentation standards are defined
- [ ] Architecture documentation requirements are specified
- [ ] Code documentation expectations are outlined
- [ ] System diagrams and visualizations are included
- [ ] Decision records for key choices are included

## 8. DEPENDENCY & INTEGRATION MANAGEMENT

[[LLM: For each dependency, weigh unavailability impact, patch currency, vendor lock-in, and contingency/fallback plans.]]

### 8.1 External Dependencies

- [ ] All external dependencies are identified
- [ ] Versioning strategy for dependencies is defined
- [ ] Fallback approaches for critical dependencies are specified
- [ ] Licensing implications are addressed
- [ ] Update and patching strategy is outlined

### 8.2 Internal Dependencies

- [ ] Component dependencies are clearly mapped
- [ ] Build order dependencies are addressed
- [ ] Shared services and utilities are identified
- [ ] Circular dependencies are eliminated
- [ ] Versioning strategy for internal components is defined

### 8.3 Third-Party Integrations

- [ ] All third-party integrations are identified
- [ ] Integration approaches are defined
- [ ] Authentication with third parties is addressed
- [ ] Error handling for integration failures is specified
- [ ] Rate limits and quotas are considered

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

[[LLM: Generate a final validation report: (1) Executive Summary — readiness (H/M/L), critical risks, strengths, project type & sections evaluated; (2) Section Analysis — pass rate per section, worst gaps, sections skipped; (3) Risk Assessment — top 5 risks, mitigations, timeline impact; (4) Recommendations — must-fix / should-fix / nice-to-have; (5) AI Implementation Readiness — concerns, unclear areas, complexity hotspots; (6) Frontend-Specific Assessment if applicable — completeness, doc alignment, UI/UX coverage, component clarity. Then offer deeper analysis of any flagged section.]]
