# Spec Directory Structure

**NON-NEGOTIABLE**:
- Use this **ONLY** when run one of idea.prompt.md, specify.prompt.md, plan.prompt.md, tasks.prompt.md, implements.prompt.md 

## Complete Folder Organization

```
specs/{number}-{branch-name}/
├── status.md                          # Project status dashboard
├── idea/                              # Idea phase
│   ├── README.md                      # Main idea documentation (was idea.md)
│   ├── research/                      # Market and competitive research
│   │   ├── market-analysis.md         # Market size, opportunities, trends
│   │   ├── competitor-analysis.md     # Competitive landscape analysis
│   │   ├── user-research.md           # User interviews, surveys, personas
│   │   └── technical-research.md      # Technology feasibility research
│   ├── brainstorming/                 # Creative exploration documents
│   │   ├── solution-approaches.md     # Multiple solution approaches
│   │   ├── creative-sessions.md       # Brainstorming session notes
│   │   ├── assumptions.md             # Key assumptions and validations
│   │   └── breakthrough-insights.md   # Key insights and innovations
│   ├── stakeholders/                  # Stakeholder engagement
│   │   ├── personas.md                # Detailed user personas
│   │   ├── stakeholder-map.md         # Stakeholder identification
│   │   ├── interviews.md              # Stakeholder interview notes
│   │   └── requirements-gathering.md  # Initial requirements from stakeholders
│   ├── validation/                    # Idea validation materials
│   │   ├── problem-validation.md      # Problem statement validation
│   │   ├── solution-validation.md     # Solution approach validation
│   │   ├── risk-assessment.md         # Initial risk identification
│   │   └── success-criteria.md        # Success metrics and KPIs
│   └── artifacts/                     # Generated artifacts from idea phase
│       ├── idea.handoff.json          # Structured handoff data
│       ├── research-summary.pdf       # Research findings summary
│       └── presentation.pptx          # Stakeholder presentation
├── specify/                           # Specification phase
│   ├── README.md                      # Main specification documentation (was specify.md)
│   ├── requirements/                  # Detailed requirements documentation
│   │   ├── functional-requirements.md # Detailed functional requirements
│   │   ├── non-functional-requirements.md # Performance, security, etc.
│   │   ├── business-rules.md          # Business logic and constraints
│   │   ├── compliance-requirements.md # Regulatory and compliance needs
│   │   └── user-requirements.md       # User-focused requirements
│   ├── design/                        # System and technical design
│   │   ├── system-architecture.md     # High-level system design
│   │   ├── database-design.md         # Database schema and relationships
│   │   ├── api-design.md              # API endpoint specifications
│   │   ├── security-design.md         # Security architecture and controls
│   │   ├── ui-ux-design.md            # User interface and experience design
│   │   └── integration-design.md      # External system integrations
│   ├── specifications/                # Detailed technical specifications
│   │   ├── data-models.md             # Data structure specifications
│   │   ├── interface-contracts.md     # API and service contracts
│   │   ├── validation-rules.md        # Input validation specifications
│   │   ├── error-handling.md          # Error handling specifications
│   │   └── performance-specs.md       # Performance requirements and targets
│   ├── testing/                       # Testing strategy and specifications
│   │   ├── test-strategy.md           # Overall testing approach
│   │   ├── test-scenarios.md          # Detailed test scenarios
│   │   ├── acceptance-criteria.md     # User story acceptance criteria
│   │   ├── performance-testing.md     # Performance testing specifications
│   │   └── security-testing.md        # Security testing requirements
│   └── artifacts/                     # Generated artifacts from specify phase
│       ├── specify.handoff.json       # Structured handoff data
│       ├── openapi.yaml               # API contract specification
│       ├── database-schema.sql        # Database schema definition
│       ├── wireframes.figma           # UI/UX wireframes
│       └── technical-specs.pdf        # Comprehensive technical specification
├── plan/                              # Planning phase
│   ├── README.md                      # Main planning documentation (was plan.md)
│   ├── architecture/                  # Detailed architecture planning
│   │   ├── system-design.md           # Detailed system architecture
│   │   ├── component-design.md        # Component breakdown and responsibilities
│   │   ├── deployment-architecture.md # Deployment and infrastructure design
│   │   ├── security-architecture.md   # Security implementation planning
│   │   └── integration-architecture.md # Integration patterns and approaches
│   ├── implementation/                # Implementation planning
│   │   ├── technology-selection.md    # Technology stack decisions
│   │   ├── development-approach.md    # Development methodology and approach
│   │   ├── coding-standards.md        # Code quality and standards
│   │   ├── testing-approach.md        # Testing implementation strategy
│   │   └── deployment-strategy.md     # Deployment and release planning
│   ├── project-management/            # Project management planning
│   │   ├── timeline-estimation.md     # Project timeline and milestones
│   │   ├── resource-allocation.md     # Team and resource planning
│   │   ├── risk-mitigation.md         # Risk management strategies
│   │   ├── communication-plan.md      # Stakeholder communication strategy
│   │   └── quality-assurance.md       # QA processes and standards
│   ├── operations/                    # Operational planning
│   │   ├── infrastructure-plan.md     # Infrastructure requirements
│   │   ├── monitoring-strategy.md     # Monitoring and alerting planning
│   │   ├── backup-recovery.md         # Backup and disaster recovery
│   │   ├── security-operations.md     # Security operations planning
│   │   └── maintenance-plan.md        # Ongoing maintenance strategy
│   └── artifacts/                     # Generated artifacts from plan phase
│       ├── plan.handoff.json          # Structured handoff data
│       ├── architecture-diagrams.drawio # System architecture diagrams
│       ├── project-timeline.mpp       # Detailed project timeline
│       └── technical-decisions.md     # Architecture decision records
├── tasks/                             # Tasks phase
│   ├── README.md                      # Task summary and overview (was tasks.md)
│   ├── epic-001-auth-core/
│   ├── epic-002-security/
│   └── tasks-summary.md
├── tasks/{tasks-id}/implements/       # Implementation phase (per-task under tasks/)
│   ├── README.md                      # Implementation documentation (was implements.md)
│   ├── development/                   # Development artifacts
│   │   ├── code-standards.md          # Applied coding standards
│   │   ├── development-log.md         # Development progress log
│   │   ├── technical-decisions.md     # Implementation decisions made
│   │   ├── performance-optimization.md # Performance optimization notes
│   │   └── refactoring-log.md         # Code refactoring documentation
│   ├── testing/                       # Testing implementation
│   │   ├── test-results.md            # Test execution results
│   │   ├── coverage-reports/          # Test coverage reports
│   │   ├── performance-results/       # Performance test results
│   │   ├── security-scan-results/     # Security scanning results
│   │   └── integration-test-logs/     # Integration test logs
│   ├── deployment/                    # Deployment documentation
│   │   ├── deployment-guide.md        # Step-by-step deployment guide
│   │   ├── configuration-management.md # Configuration settings
│   │   ├── infrastructure-setup.md    # Infrastructure configuration
│   │   ├── monitoring-setup.md        # Monitoring configuration
│   │   └── rollback-procedures.md     # Rollback and recovery procedures
│   ├── documentation/                 # User and technical documentation
│   │   ├── user-guide.md              # End-user documentation
│   │   ├── api-documentation.md       # API usage documentation
│   │   ├── admin-guide.md             # Administrator documentation
│   │   ├── troubleshooting-guide.md   # Common issues and solutions
│   │   └── maintenance-guide.md       # Ongoing maintenance procedures
│   └── artifacts/                     # Generated artifacts from implementation
│       ├── implements.handoff.json    # Final project handoff data
│       ├── final-test-reports/        # Comprehensive test results
│       ├── performance-benchmarks/    # Performance measurement results
│       ├── security-audit-report.pdf  # Security assessment results
│       └── deployment-artifacts/      # Production deployment files
└── shared/                            # Shared resources across all phases
    ├── glossary.md                    # Project terminology and definitions
    ├── references.md                  # External references and links
    ├── meeting-notes/                 # All project meeting notes
    ├── decisions/                     # Architecture and design decisions
    └── communications/                # Stakeholder communications
```

## Folder Purpose and Content Guidelines

### idea/ Supporting Documents

#### research/
- **market-analysis.md**: Market size, growth trends, opportunities
- **competitor-analysis.md**: Direct and indirect competitors, feature comparison
- **user-research.md**: User interviews, surveys, behavior analysis
- **technical-research.md**: Technology feasibility, constraints, alternatives

#### brainstorming/
- **solution-approaches.md**: Multiple solution alternatives with pros/cons
- **creative-sessions.md**: Brainstorming session outputs using various techniques
- **assumptions.md**: Key assumptions that need validation
- **breakthrough-insights.md**: Innovative ideas and creative breakthroughs

#### stakeholders/
- **personas.md**: Detailed user personas with goals, frustrations, behaviors
- **stakeholder-map.md**: All project stakeholders and their interests
- **interviews.md**: Stakeholder interview transcripts and insights
- **requirements-gathering.md**: Initial requirements from stakeholder input

### specify/ Supporting Documents

#### requirements/
- **functional-requirements.md**: Detailed functional capabilities
- **non-functional-requirements.md**: Performance, security, usability requirements
- **business-rules.md**: Business logic, constraints, validation rules
- **compliance-requirements.md**: Legal, regulatory, industry compliance

#### design/
- **system-architecture.md**: High-level system design and components
- **database-design.md**: Database schema, relationships, constraints
- **api-design.md**: RESTful API design, endpoints, data formats
- **security-design.md**: Authentication, authorization, encryption design

#### specifications/
- **data-models.md**: Detailed data structure specifications
- **interface-contracts.md**: API contracts, service interfaces
- **validation-rules.md**: Input validation, data integrity rules
- **error-handling.md**: Error codes, messages, handling strategies

### plan/ Supporting Documents

#### architecture/
- **system-design.md**: Detailed technical architecture
- **component-design.md**: Component responsibilities and interactions
- **deployment-architecture.md**: Infrastructure and deployment design
- **security-architecture.md**: Security implementation architecture

#### implementation/
- **technology-selection.md**: Technology choices with justifications
- **development-approach.md**: Development methodology and practices
- **coding-standards.md**: Code quality standards and conventions
- **testing-approach.md**: Testing strategy and implementation

### tasks/{tasks-id}/implements/ Supporting Documents

#### development/
- **development-log.md**: Daily development progress and decisions
- **technical-decisions.md**: Implementation-specific technical decisions
- **performance-optimization.md**: Performance improvements made
- **refactoring-log.md**: Code refactoring activities and outcomes

#### testing/
- **test-results.md**: Comprehensive test execution results
- **coverage-reports/**: Test coverage analysis and reports
- **performance-results/**: Performance testing outcomes
- **security-scan-results/**: Security vulnerability assessments

## Implementation Guidelines

### When to Create Supporting Folders
1. **During Phase Execution**: Create folders as documents are needed
2. **Complex Projects**: Create all folders upfront for large projects
3. **Simple Projects**: Create only necessary folders to avoid over-engineering

### Document Naming Conventions
- Use kebab-case for all file names
- Include date stamps for versioned documents
- Use descriptive names that indicate content purpose
- Maintain consistent naming across phases

### Content Organization Principles
- **Single Responsibility**: Each document has one clear purpose
- **Cross-References**: Link related documents across phases
- **Version Control**: Track changes and decisions over time
- **Accessibility**: Organize for easy navigation and discovery

### Handoff Integration
- **Phase Completion**: Consolidate supporting documents into main phase file
- **Context Preservation**: Maintain links to supporting documents in handoff data
- **Artifact Management**: Organize generated artifacts for easy access
- **Knowledge Transfer**: Ensure supporting documents enhance understanding