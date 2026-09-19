# Task Breakdown Structure Guide

## Recommended Folder Structure

```
specs/001-user-authentication/
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
├── tasks/{tasks-id}/implements/  # Implementation phase (per-task under tasks/)
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
# Epic: {Epic Name}

## Epic Overview
- **Epic ID**: EPIC-{number}
- **Epic Name**: {descriptive_name}
- **Priority**: CRITICAL | HIGH | MEDIUM | LOW
- **Estimated Effort**: {story_points | weeks}
- **Owner**: {team_or_person}
- **Status**: NOT_STARTED | IN_PROGRESS | BLOCKED | COMPLETED

## Epic Goals
- **Primary Goal**: {main_objective}
- **Success Criteria**: {measurable_outcomes}
- **Acceptance Criteria**: {epic_level_acceptance}

## User Stories
1. **Story 001**: {story_name} - {status} - {effort}
2. **Story 002**: {story_name} - {status} - {effort}
3. **Story 003**: {story_name} - {status} - {effort}

## Dependencies
- **Depends On**: {other_epics_or_stories}
- **Blocks**: {what_this_epic_blocks}
- **External Dependencies**: {external_systems_or_teams}

## Technical Notes
- **Key Technical Decisions**: {list_decisions}
- **Architecture Impact**: {how_this_affects_architecture}
- **Performance Considerations**: {performance_notes}
- **Security Considerations**: {security_notes}

## Progress Tracking
- **Started**: {date}
- **Target Completion**: {date}
- **Actual Completion**: {date}
- **Completion Percentage**: {0-100}%

## Risks & Mitigation
- **Risk 1**: {description} - Mitigation: {strategy}
- **Risk 2**: {description} - Mitigation: {strategy}
```

### story-{number}-{name}.md Template
```markdown
# User Story: {Story Name}

## Story Information
- **Story ID**: STORY-{epic_number}-{story_number}
- **Epic**: {parent_epic_name}
- **Title**: {story_title}
- **Priority**: CRITICAL | HIGH | MEDIUM | LOW
- **Status**: NOT_STARTED | IN_PROGRESS | BLOCKED | COMPLETED
- **Assigned To**: {developer_name}
- **Estimated Effort**: {hours | story_points}

## User Story
**As a** {user_type}
**I want** {functionality}
**So that** {business_value}

## Acceptance Criteria
Given-When-Then format:

### Scenario 1: {scenario_name}
- **Given** {initial_context}
- **When** {action_performed}
- **Then** {expected_outcome}
- **And** {additional_expectations}

### Scenario 2: {scenario_name}
- **Given** {initial_context}
- **When** {action_performed}
- **Then** {expected_outcome}

## Technical Details
- **Affected Components**: {list_components}
- **API Endpoints**: {list_endpoints}
- **Database Changes**: {describe_changes}
- **Frontend Changes**: {describe_changes}
- **Test Requirements**: {describe_tests}

## Definition of Done
- [ ] Code implemented according to specifications
- [ ] Unit tests written and passing (>80% coverage)
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
- **Depends On**: {other_stories_or_tasks}
- **Blocks**: {what_this_story_blocks}
- **External Dependencies**: {external_dependencies}

## Notes & Decisions
- **Technical Decisions**: {key_decisions_made}
- **Design Decisions**: {ui_ux_decisions}
- **Known Issues**: {any_known_issues}
- **Future Enhancements**: {potential_improvements}

## Progress Log
- **{date}**: Story created and refined
- **{date}**: Development started
- **{date}**: Backend implementation completed
- **{date}**: Frontend implementation completed
- **{date}**: Testing completed
- **{date}**: Story completed and delivered

---
*Last Updated: {timestamp} | Updated By: {developer_name}*
```

## Checklist Template

### checklist.md Template
```markdown
# Epic Completion Checklist

## Epic: {Epic Name}

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
- [ ] Unit tests written and passing (>80% coverage)
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
**Epic Completion Status**: {percentage}% Complete
**Next Epic**: {next_epic_name}
**Completion Date**: {completion_date}
```

## Implementation Guidelines

### When to Create Task Structure
1. **After Plan Phase**: Create basic epic structure
2. **During Tasks Phase**: Detail all stories and tasks
3. **Before Implementation**: Finalize all checklists

### Naming Conventions
- **Epics**: `epic-{number}-{short-name}/`
- **Stories**: `story-{number}-{short-name}.md`
- **Tasks**: Use story format with task breakdown

### Status Management
- Update status.md after each story completion
- Update epic checklists as tasks complete
- Maintain cross-references between documents

This structure provides comprehensive task management while maintaining constitutional framework compliance.