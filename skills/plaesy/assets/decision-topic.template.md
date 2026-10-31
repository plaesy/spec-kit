---
title: "Decision: {{TOPIC_TITLE}}"
description: "Recorded decision on {{TOPIC_TITLE}} with rationale, references, and access timestamps"
updatedAt: "{{UPDATED_AT|<ISO-8601 timestamp, e.g. 2026-01-31T09:00:00Z>}}"
decided: "{{DECISION_DATE}}"
status: "{{STATUS|Proposed, Accepted, Rejected, Deprecated, or Superseded}}"
---

# Decision: {{TOPIC_TITLE}}

**Status**: {{STATUS|Proposed, Accepted, Rejected, Deprecated, or Superseded}}
**Decided**: {{DECISION_DATE}}
**Deciders**: {{DECIDERS|people involved}}
**Related**: {{RELATED_DECISIONS|other decision topic links}}

## Context and Problem Statement

{{CONTEXT_STATEMENT|two to three sentences describing the problem, ideally framed as a question}}

- **Business context**: impact, stakeholders, timeline/budget constraints
- **Technical context**: current state, constraints, integration/performance/security requirements

## Decision Drivers

- **Driver 1**: {{DRIVER_1|description and rationale}}
- **Driver 2**: {{DRIVER_2|description and rationale}}
- **Constitutional compliance**: {{CONSTITUTIONAL_COMPLIANCE|alignment with framework principles}}

## Considered Options

### Option 1: {{OPTION_TITLE}}

**Pros**: {{OPTION_1_PRO_1|advantage 1}}, {{OPTION_1_PRO_2|advantage 2}}
**Cons**: {{OPTION_1_CON_1|disadvantage 1}}, {{OPTION_1_CON_2|disadvantage 2}}
**Effort / Risk / Maintenance**: {{OPTION_1_EFFORT|High, Medium, or Low}} / {{OPTION_1_RISK|High, Medium, or Low}} / {{OPTION_1_MAINTENANCE|High, Medium, or Low}}

### Option 2: {{OPTION_TITLE}}

**Pros**: {{OPTION_2_PRO_1|advantage 1}}, {{OPTION_2_PRO_2|advantage 2}}
**Cons**: {{OPTION_2_CON_1|disadvantage 1}}, {{OPTION_2_CON_2|disadvantage 2}}
**Effort / Risk / Maintenance**: {{OPTION_2_EFFORT|High, Medium, or Low}} / {{OPTION_2_RISK|High, Medium, or Low}} / {{OPTION_2_MAINTENANCE|High, Medium, or Low}}

### Option 3: {{OPTION_TITLE}}

**Pros**: {{OPTION_3_PRO_1|advantage 1}}, {{OPTION_3_PRO_2|advantage 2}}
**Cons**: {{OPTION_3_CON_1|disadvantage 1}}, {{OPTION_3_CON_2|disadvantage 2}}
**Effort / Risk / Maintenance**: {{OPTION_3_EFFORT|High, Medium, or Low}} / {{OPTION_3_RISK|High, Medium, or Low}} / {{OPTION_3_MAINTENANCE|High, Medium, or Low}}

## Decision Outcome

**Chosen option**: "Option {{CHOSEN_OPTION_NUMBER}}: {{OPTION_TITLE}}"

**Rationale**: {{RATIONALE|why this option was chosen and how it addresses the decision drivers}}

**Expected benefits**: {{EXPECTED_BENEFIT_1|benefit 1}}, {{EXPECTED_BENEFIT_2|benefit 2}}
**Accepted risks**: {{ACCEPTED_RISK_1|risk 1}} → **mitigation**: {{RISK_1_MITIGATION|strategy}}; {{ACCEPTED_RISK_2|risk 2}} → **mitigation**: {{RISK_2_MITIGATION|strategy}}

## Implementation Plan

### Phase 1: {{PHASE_NAME}} ({{PHASE_1_TIMELINE}})

- [ ] {{PHASE_1_TASK_1|task 1}}
- [ ] {{PHASE_1_TASK_2|task 2}}
- **Success/exit criteria**: {{PHASE_1_EXIT_CRITERIA|how to measure completion}}

### Phase 2: {{PHASE_NAME}} ({{PHASE_2_TIMELINE}})

- [ ] {{PHASE_2_TASK_1|task 1}}
- [ ] {{PHASE_2_TASK_2|task 2}}
- **Success/exit criteria**: {{PHASE_2_EXIT_CRITERIA|how to measure completion}}

## Constitutional Compliance & Quality Gates

- **Framework alignment**: TDD, interface contracts, real dependencies, observability, security, platform-agnostic — note how this decision supports each
- [ ] Performance requirements met
- [ ] Security review passed
- [ ] Maintainability / testability / scalability / reliability acceptable

## Impact Assessment

- **Technical**: systems/APIs/data/infrastructure affected
- **Team**: training, skill gaps, tooling/process changes
- **Stakeholders**: end users, operations, support, business teams

## Validation Strategy

- **Validation criteria**: functional, performance, security, integration
- **Testing approach**: unit, integration, E2E, performance, security
- **Success metrics**: {{SUCCESS_METRIC_1|metric and target}}, {{SUCCESS_METRIC_2|metric and target}}

## Rollback Plan

- **Triggers**: {{ROLLBACK_TRIGGERS|conditions that would trigger rollback}}
- **Procedure**: {{ROLLBACK_PROCEDURE|steps, data recovery, and stakeholder communication}}

## References

<!-- Each reference MUST include a URL and an access timestamp.
     Format: REFERENCE_TITLE → URL — accessed YYYY-MM-DDTHH:MM:SSZ
     An entry without a URL or access time is incomplete. -->

- {{REFERENCE_TITLE_1}} → {{URL_1}} — accessed {{ACCESS_TIMESTAMP_1|YYYY-MM-DDTHH:MM:SSZ}}
- {{REFERENCE_TITLE_2}} → {{URL_2}} — accessed {{ACCESS_TIMESTAMP_2|YYYY-MM-DDTHH:MM:SSZ}}
- {{REFERENCE_TITLE_3}} → {{URL_3}} — accessed {{ACCESS_TIMESTAMP_3|YYYY-MM-DDTHH:MM:SSZ}}

## Related Decisions

- **Prerequisites**: {{PREREQUISITE_DECISIONS|other decision topic links}}
- **Follow-ups**: {{FOLLOW_UP_DECISIONS|future decisions needed as a result}}
- **Related docs**: {{RELATED_DOC_LINKS|links}}

---

**Document Control**: {{TOPIC_TITLE}} · v1.0 · Created {{CREATED_DATE}} by {{AUTHOR}} · Last modified {{LAST_MODIFIED_DATE}} · Next review {{NEXT_REVIEW_DATE}} · Approvers: {{APPROVERS|list}}

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0 | {{CREATED_DATE}} | {{AUTHOR}} | Initial version |
