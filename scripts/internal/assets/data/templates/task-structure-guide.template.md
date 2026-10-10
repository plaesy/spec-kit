# Task Breakdown Structure Guide

## Recommended Folder Structure

```text

.plaesy/specs/001-user-authentication/
├── status.md                 # Overall project status dashboard
├── idea/                     # Idea phase
│   ├── README.md             # Main idea documentation (was idea.md)
│   ├── research/             # Research supporting documents
│   ├── brainstorming/        # Creative exploration
│   ├── stakeholders/         # Stakeholder engagement
│   ├── validation/           # Idea validation
│   └── artifacts/            # Generated handoff data
│       └── idea.handoff.json
├── specify/                  # Specification phase
│   ├── README.md             # Main specification (was specify.md)
│   ├── requirements/         # Requirements documentation
│   ├── design/               # System design
│   ├── specifications/       # Technical specifications
│   ├── testing/              # Testing strategy
│   └── artifacts/            # Generated artifacts
│       └── specify.handoff.json
├── plan/                     # Planning phase
│   ├── README.md             # Main planning doc (was plan.md)
│   ├── architecture/         # Architecture planning
│   ├── implementation/       # Implementation planning
│   ├── project-management/   # PM planning
│   ├── operations/           # Operational planning
│   └── artifacts/            # Generated artifacts
│       └── plan.handoff.json
├── tasks/                    # Tasks phase
│   ├── README.md             # Task summary (was tasks.md)
│   ├── epic-001-auth-core/   # Core authentication epic
│   │   ├── epic.md           # Epic overview and goals
│   │   ├── story-001-login.md        # User login story
│   │   ├── story-002-register.md     # User registration story
│   │   ├── story-003-logout.md       # User logout story
│   │   └── checklist.md      # Epic completion checklist
│   ├── epic-002-security/    # Security features epic
│   │   ├── epic.md           # Security epic overview
│   │   ├── story-001-2fa.md          # Two-factor authentication
│   │   ├── story-002-password.md     # Password security
│   │   ├── story-003-session.md      # Session management
│   │   └── checklist.md      # Epic completion checklist
│   ├── epic-003-integration/ # Integration epic
│   │   ├── epic.md           # Integration overview
│   │   ├── story-001-database.md     # Database integration
│   │   ├── story-002-frontend.md     # Frontend integration
│   │   └── checklist.md      # Epic completion checklist
│   └── tasks-summary.md      # Cross-epic task summary
├── tasks/{{TASKS_ID}}/implements/  # Implementation phase (per-task under tasks/)
│   ├── README.md             # Implementation doc (was implements.md)
│   ├── development/          # Development artifacts
│   ├── testing/              # Test results
│   ├── deployment/           # Deployment guides
│   ├── documentation/        # User & tech docs
│   └── artifacts/            # Final deliverables
└── shared/                   # Cross-phase resources
    ├── glossary.md           # Project terminology
    ├── references.md         # External references
    ├── meeting-notes/        # Meeting records
    ├── decisions/            # Architecture decisions
    └── communications/       # Stakeholder communications

```

## Epic Template Structure

### epic.md Template

```markdown

# Epic: {{EPIC_NAME}}

## Epic Overview
- **Epic ID**: EPIC-{{NUMBER}}
- **Epic Name**: {{DESCRIPTIVE_NAME}}
- **Priority**: CRITICAL | HIGH | MEDIUM | LOW
- **Estimated Effort**: {story_points | weeks}
- **Owner**: {{TEAM_OR_PERSON}}
- **Status**: NOT_STARTED | IN_PROGRESS | BLOCKED | COMPLETED

## Epic Goals
- **Primary Goal**: {{MAIN_OBJECTIVE}}
- **Success Criteria**: {{MEASURABLE_OUTCOMES}}
- **Acceptance Criteria**: {{EPIC_LEVEL_ACCEPTANCE}}

## User Stories
1. **Story 001**: {{STORY_NAME}} - {{STATUS}} - {{EFFORT}}
2. **Story 002**: {{STORY_NAME}} - {{STATUS}} - {{EFFORT}}
3. **Story 003**: {{STORY_NAME}} - {{STATUS}} - {{EFFORT}}

## Dependencies
- **Depends On**: {{OTHER_EPICS_OR_STORIES}}
- **Blocks**: {{WHAT_THIS_EPIC_BLOCKS}}
- **External Dependencies**: {{EXTERNAL_SYSTEMS_OR_TEAMS}}

## Technical Notes
- **Key Technical Decisions**: {{LIST_DECISIONS}}
- **Architecture Impact**: {{HOW_THIS_AFFECTS_ARCHITECTURE}}
- **Performance Considerations**: {{PERFORMANCE_NOTES}}
- **Security Considerations**: {{SECURITY_NOTES}}

## Progress Tracking
- **Started**: {{DATE}}
- **Target Completion**: {{DATE}}
- **Actual Completion**: {{DATE}}
- **Completion Percentage**: {0-100}%

## Risks & Mitigation
- **Risk 1**: {{DESCRIPTION}} - Mitigation: {{STRATEGY}}
- **Risk 2**: {{DESCRIPTION}} - Mitigation: {{STRATEGY}}
```

### story-{{NUMBER}}-{{NAME}}.md Template

```markdown

# User Story: {{STORY_NAME}}

## Story Information
- **Story ID**: STORY-{{EPIC_NUMBER}}-{{STORY_NUMBER}}
- **Epic**: {{PARENT_EPIC_NAME}}
- **Title**: {{STORY_TITLE}}
- **Priority**: CRITICAL | HIGH | MEDIUM | LOW
- **Status**: NOT_STARTED | IN_PROGRESS | BLOCKED | COMPLETED
- **Assigned To**: {{DEVELOPER_NAME}}
- **Estimated Effort**: {hours | story_points}

## User Story
**As a** {{USER_TYPE}}
**I want** {{FUNCTIONALITY}}
**So that** {{BUSINESS_VALUE}}

## Acceptance Criteria
Given-When-Then format:

### Scenario 1: {{SCENARIO_NAME}}
- **Given** {{INITIAL_CONTEXT}}
- **When** {{ACTION_PERFORMED}}
- **Then** {{EXPECTED_OUTCOME}}
- **And** {{ADDITIONAL_EXPECTATIONS}}

### Scenario 2: {{SCENARIO_NAME}}
- **Given** {{INITIAL_CONTEXT}}
- **When** {{ACTION_PERFORMED}}
- **Then** {{EXPECTED_OUTCOME}}

## Technical Details
- **Affected Components**: {{LIST_COMPONENTS}}
- **API Endpoints**: {{LIST_ENDPOINTS}}
- **Database Changes**: {{DESCRIBE_CHANGES}}
- **Frontend Changes**: {{DESCRIBE_CHANGES}}
- **Test Requirements**: {{DESCRIBE_TESTS}}

## Definition of Done
- [ ] Code implemented according to specifications
- [ ] Unit tests written and passing (>{{TEST_COVERAGE_MIN|90}}% coverage)
- [ ] Integration tests written and passing
- [ ] API documentation updated
- [ ] Security review completed
- [ ] Performance requirements met
- [ ] Code review completed and approved
- [ ] Feature tested in staging environment
- [ ] Acceptance criteria validated
- [ ] User documentation updated (if applicable)

## Tasks Breakdown
### Development Tasks
- [ ] **DEV-001**: Setup database schema changes
- [ ] **DEV-002**: Implement backend API endpoints
- [ ] **DEV-003**: Create frontend components
- [ ] **DEV-004**: Implement validation logic
- [ ] **DEV-005**: Add error handling

### Testing Tasks
- [ ] **TEST-001**: Write unit tests for backend logic
- [ ] **TEST-002**: Write unit tests for frontend components
- [ ] **TEST-003**: Write integration tests
- [ ] **TEST-004**: Write end-to-end tests
- [ ] **TEST-005**: Perform security testing

### Documentation Tasks
- [ ] **DOC-001**: Update API documentation
- [ ] **DOC-002**: Update user documentation
- [ ] **DOC-003**: Update technical documentation

## Dependencies
- **Depends On**: {{OTHER_STORIES_OR_TASKS}}
- **Blocks**: {{WHAT_THIS_STORY_BLOCKS}}
- **External Dependencies**: {{EXTERNAL_DEPENDENCIES}}

## Notes & Decisions
- **Technical Decisions**: {{KEY_DECISIONS_MADE}}
- **Design Decisions**: {{UI_UX_DECISIONS}}
- **Known Issues**: {{ANY_KNOWN_ISSUES}}
- **Future Enhancements**: {{POTENTIAL_IMPROVEMENTS}}

## Progress Log
- **{{DATE}}**: Story created and refined
- **{{DATE}}**: Development started
- **{{DATE}}**: Backend implementation completed
- **{{DATE}}**: Frontend implementation completed
- **{{DATE}}**: Testing completed
- **{{DATE}}**: Story completed and delivered

---
*Last Updated: {{TIMESTAMP}} | Updated By: {{DEVELOPER_NAME}}*
```

## Checklist Template

### checklist.md Template

```markdown

# Epic Completion Checklist

## Epic: {{EPIC_NAME}}

### Pre-Development Checklist
- [ ] Epic properly defined and scoped
- [ ] All user stories documented and refined
- [ ] Acceptance criteria clearly defined
- [ ] Technical architecture designed
- [ ] Dependencies identified and resolved
- [ ] Resource allocation confirmed
- [ ] Timeline estimated and approved

### Development Checklist
- [ ] All user stories implemented
- [ ] Code follows constitutional framework principles
- [ ] TDD Red-Green-Refactor cycle followed
- [ ] Interface contracts implemented
- [ ] Security requirements implemented
- [ ] Performance requirements met
- [ ] Error handling implemented

### Testing Checklist
- [ ] Unit tests written and passing (>{{TEST_COVERAGE_MIN|90}}% coverage)
- [ ] Integration tests written and passing
- [ ] End-to-end tests written and passing
- [ ] Contract tests written and passing
- [ ] Security tests completed
- [ ] Performance tests completed
- [ ] Load testing completed (if applicable)
- [ ] Accessibility testing completed (if applicable)

### Documentation Checklist
- [ ] API documentation updated
- [ ] User documentation updated
- [ ] Technical documentation updated
- [ ] Deployment documentation updated
- [ ] Troubleshooting guide updated
- [ ] Change log updated

### Quality Assurance Checklist
- [ ] Code review completed and approved
- [ ] Security review completed
- [ ] Architecture review completed
- [ ] Performance review completed
- [ ] User experience review completed
- [ ] Constitutional compliance verified

### Deployment Checklist
- [ ] Staging deployment successful
- [ ] Staging testing completed
- [ ] Production deployment plan reviewed
- [ ] Rollback plan prepared
- [ ] Monitoring and alerting configured
- [ ] Production deployment successful

### Post-Deployment Checklist
- [ ] Production functionality verified
- [ ] Performance monitoring confirmed
- [ ] Error monitoring confirmed
- [ ] User acceptance testing completed
- [ ] Stakeholder sign-off received
- [ ] Epic marked as complete

### Retrospective
- [ ] What went well in this epic?
- [ ] What could be improved?
- [ ] Lessons learned documented
- [ ] Process improvements identified
- [ ] Knowledge sharing completed

---
**Epic Completion Status**: {{PERCENTAGE}}% Complete
**Next Epic**: {{NEXT_EPIC_NAME}}
**Completion Date**: {{COMPLETION_DATE}}
```

## Implementation Guidelines

### When to Create Task Structure

1. **After Plan Phase**: Create basic epic structure
2. **During Tasks Phase**: Detail all stories and tasks
3. **Before Implementation**: Finalize all checklists

### Naming Conventions

- **Epics**: `epic-{{NUMBER}}-{short-name}/`
- **Stories**: `story-{{NUMBER}}-{short-name}.md`
- **Tasks**: Use story format with task breakdown

### Status Management

- Update status.md after each story completion
- Update epic checklists as tasks complete
- Maintain cross-references between documents

This structure provides comprehensive task management while maintaining constitutional framework compliance.
