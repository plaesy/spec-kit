# Project Status Dashboard

## Project Information

- **Feature Name**: {{FEATURE_NAME}}
-  **Project Type**: {{PROJECT_TYPE}} — one of: web_app, mobile_app, desktop_app, api_service, microservice, cli_tool, library, extension, database, analytics, integration, automation, infrastructure,
  documentation
- **Status**: NOT_STARTED | TODO | PROGRESS | BLOCKED | NEED_REVIEW | COMPLETE
- **Active Phase** (9-phase command ladder, see `instructions/plaesy.instructions.md`): {{CURRENT_PHASE}}
- **Current Artifact Stage** (the spec/plan/tasks stages listed under *Stage Progress*): {{CURRENT_STAGE}}
- **Overall Completion**: {{PERCENTAGE}}%
- **Timeline Status**: ON_TRACK | AT_RISK | DELAYED | AHEAD
- **Spec Directory**: .plaesy/specs/{{NUMBER}}-{{BRANCH_NAME}}/
- **Created**: {{TIMESTAMP}}
- **Last Updated**: {{TIMESTAMP}}

## Stage Progress

The five stages below are the **artifact pipeline** (`/create spec` → `plan` →
`tasks` → `/implement`). They are *not* the 9-phase command ladder: the ladder
numbers are defined once in `instructions/plaesy.instructions.md` and are not
restated here. Track a feature against both — a stage is complete when its files
exist and pass their checklist, and a ladder phase is complete when its command
has been run to a green gate.

### Stage 1: Spec & Idea (ladder Phase 1) {{STATUS}}

- **Status**: NOT_STARTED | TODO | PROGRESS | BLOCKED | NEED_REVIEW | COMPLETE
- **Completion**: {{PERCENTAGE}}%
- **File**: `.plaesy/specs/{{NUMBER}}-{{BRANCH}}/spec.md` (idea + problem framing)
- **Quality Gates**:
  - [ ] Problem statement validated
  - [ ] Research completed
  - [ ] Solution approaches evaluated
  - [ ] User personas defined
  - [ ] Success criteria established
  - [ ] Constitutional compliance verified
  - [ ] AI validation passed
  - [ ] Brainstorming techniques applied
  - [ ] Creative breakthroughs documented
  - [ ] ...
- **Blockers**: {{LIST_ANY_BLOCKERS}}
- **Next Action**: {{SPECIFIC_NEXT_STEP}}
- **AI Analysis**: {{AI_INSIGHTS_ON_IDEA_QUALITY}}

### Stage 2: Specify (ladder Phase 1) {{STATUS}}

- **Status**: NOT_STARTED | TODO | PROGRESS | BLOCKED | NEED_REVIEW | COMPLETE
- **Completion**: {{PERCENTAGE}}%
- **File**: `.plaesy/specs/{{NUMBER}}-{{BRANCH}}/spec.md`
- **Quality Gates**:
  - [ ] Functional requirements detailed
  - [ ] Non-functional requirements defined
  - [ ] API contracts specified (OpenAPI 3.1)
  - [ ] Database design complete
  - [ ] Security requirements documented
  - [ ] Test strategy defined
  - [ ] Integration specifications complete
  - [ ] Performance requirements quantified
  - [ ] Existing project compatibility verified
  - [ ] ...
- **Artifacts**:
  - [ ] OpenAPI specification generated
  - [ ] Database schema created
  - [ ] Security threat model documented
- **Blockers**: {{LIST_ANY_BLOCKERS}}
- **Next Action**: {{SPECIFIC_NEXT_STEP}}
- **AI Analysis**: {{AI_INSIGHTS_ON_SPECIFICATION_QUALITY}}

### Stage 3: Plan (ladder Phase 1) {{STATUS}}

- **Status**: NOT_STARTED | TODO | PROGRESS | BLOCKED | NEED_REVIEW | COMPLETE
- **Completion**: {{PERCENTAGE}}%
- **File**: `.plaesy/specs/{{NUMBER}}-{{BRANCH}}/plan.md`
- **Quality Gates**:
  - [ ] Architecture design complete
  - [ ] Technology stack finalized
  - [ ] Implementation approach defined
  - [ ] Resource allocation planned
  - [ ] Timeline estimated
  - [ ] Risk mitigation strategies
  - [ ] CI/CD pipeline designed
  - [ ] Deployment strategy defined
  - [ ] Migration plan created (if existing project)
  - [ ] Parallel execution strategy optimized
  - [ ] ...
- **Artifacts**:
  - [ ] Architecture diagrams created
  - [ ] Technology decision matrix documented
  - [ ] Implementation roadmap generated
- **Blockers**: {{LIST_ANY_BLOCKERS}}
- **Next Action**: {{SPECIFIC_NEXT_STEP}}
- **AI Analysis**: {{AI_INSIGHTS_ON_PLAN_FEASIBILITY}}

### Stage 4: Tasks (ladder Phase 1) {{STATUS}}

- **Status**: NOT_STARTED | TODO | PROGRESS | BLOCKED | NEED_REVIEW | COMPLETE
- **Completion**: {{PERCENTAGE}}%
- **File**: `.plaesy/specs/{{NUMBER}}-{{BRANCH}}/tasks.md`
- **Quality Gates**:
  - [ ] Epic breakdown complete
  - [ ] Story breakdown complete
  - [ ] Task dependencies mapped
  - [ ] Acceptance criteria defined
  - [ ] Effort estimation complete
  - [ ] Sprint planning ready
  - [ ] TDD task ordering verified
  - [ ] Parallel execution optimization complete
  - [ ] Constitutional compliance integrated
  - [ ] ...
- **Task Summary**:
  - **Total Epics**: {{NUMBER}}
  - **Total Stories**: {{NUMBER}}
  - **Total Tasks**: {{NUMBER}}
  - **Estimated Effort**: {{WEEKS_OR_DAYS}}
  - **Critical Path Length**: {{DAYS}}
  - **Parallel Execution Potential**: {{PERCENTAGE}}%
- **AI Optimization**:
  - [ ] Task complexity analysis complete
  - [ ] Resource allocation optimized
  - [ ] Risk-based prioritization applied
- **Blockers**: {{LIST_ANY_BLOCKERS}}
- **Next Action**: {{SPECIFIC_NEXT_STEP}}
- **AI Analysis**: {{AI_INSIGHTS_ON_TASK_BREAKDOWN_QUALITY}}

### Stage 5: Implementation (ladder Phase 2) {{STATUS}}

- **Status**: NOT_STARTED | TODO | PROGRESS | BLOCKED | NEED_REVIEW | COMPLETE
- **Completion**: "{{PERCENTAGE_FROM_CHECKLIST}}"
- **File**: `.plaesy/specs/{{NUMBER}}-{{BRANCH}}/tasks/{tasks-id}/implements/README.md`
- **Quality Gates**:
  - [ ] Development environment setup
  - [ ] Core functionality implemented
  - [ ] Tests written and passing
  - [ ] Security requirements met
  - [ ] Performance targets achieved
  - [ ] Documentation complete (check `.plaesy/templates/*` for the template's own documentation requirements)
  - [ ] ...
- **Implementation Progress**:
  - **Completed Tasks**: {{COMPLETED}}/{{TOTAL}}
  - **Current Sprint**: {{SPRINT_NAME}}
  - **Code Coverage**: {{PERCENTAGE}}%
  - **Test Pass Rate**: {{PERCENTAGE}}%
- **Blockers**: {{LIST_ANY_BLOCKERS}}
- **Next Action**: {{SPECIFIC_NEXT_STEP}}

## Overall Project Status

### Key Metrics

- **Days Since Start**: {{NUMBER}}
- **Days to Target Completion**: {{NUMBER}}
- **Constitutional Compliance**: VERIFIED | NEEDS_REVIEW | NON_COMPLIANT
- **Test Coverage**: {{PERCENTAGE}}%
- **Technical Debt**: LOW | MEDIUM | HIGH | CRITICAL

### Next Steps

1. **Immediate Action**: {{SPECIFIC_ACTION}}
2. **This Week**: {{WEEKLY_GOALS}}
3. **This Sprint**: {{SPRINT_GOALS}}
4. **Blockers to Resolve**: {{LIST_BLOCKERS}}

### Stakeholders & Communication

- **Product Owner**: {{NAME}}
- **Tech Lead**: {{NAME}}
- **Last Stakeholder Update**: {{DATE}}
- **Next Review**: {{DATE}}

## Risk Assessment

- **Technical Risks**: {{LIST_RISKS}}
- **Timeline Risks**: {{LIST_RISKS}}
- **Resource Risks**: {{LIST_RISKS}}
- **Mitigation Strategies**: {{LIST_STRATEGIES}}

## Notes & Decisions

- **Key Decisions Made**: {{LIST_DECISIONS}}
- **Lessons Learned**: {{LIST_LESSONS}}
- **Important Notes**: {{LIST_NOTES}}

---
*Last Updated: {{TIMESTAMP}} | Updated By: {{UPDATED_BY}}*
