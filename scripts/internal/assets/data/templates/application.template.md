# Application Development Specification: {{APPLICATION_NAME}}

**Created**: {{DATE}}
**Status**: {{DOC_STATUS|draft}}
**Type**: Software application
**Source request**: {{SOURCE_REQUEST|verbatim text that prompted this specification}}
**Spec owner**: {{SPEC_OWNER}}
**Constitution dimensions in force**: {{ACTIVE_DIMENSIONS|Technical, Design}}
**Related specs**: {{RELATED_SPEC_IDS|spec directory numbers this specification depends on}}

## Working constraints

Each constraint below binds the delivered application, not this document. Every
one is either satisfied by a named section further down or it is not — a
constraint with no home is a constraint nobody is checking.

- The application starts from a validated user need, not from a chosen stack.
- Interfaces are specified before they are built, and the written contract wins.
- Tests ship with the increment that introduces the behaviour, not after it.
- Security and privacy controls belong to the design, not to a later phase.
- Accessibility targets WCAG 2.1 AA and is verified rather than asserted.
- Every capacity claim carries a number; an unquantified "scales fine" is a defect.

## Application Overview

### Core purpose

{{CORE_PURPOSE|one paragraph: what this application does, for whom, and why it exists}}

### Target users

| Role | Who they are | Primary need | Frequency |
|------|--------------|--------------|-----------|
| Primary | {{PRIMARY_USER|role the application is built around}} | {{PRIMARY_USER_NEED}} | {{PRIMARY_USER_FREQUENCY|daily}} |
| Secondary | {{SECONDARY_USER|role that uses the application occasionally}} | {{SECONDARY_USER_NEED}} | {{SECONDARY_USER_FREQUENCY|weekly}} |
| Administrator | {{ADMIN_USER|role that configures and operates the application}} | {{ADMIN_USER_NEED}} | {{ADMIN_USER_FREQUENCY|as required}} |

### Value propositions

| # | Value delivered | How it is measured | Evidence |
|---|-----------------|--------------------|----------|
| 1 | {{VALUE_PROP_1|primary benefit this application provides}} | {{VALUE_PROP_1_METRIC}} | {{VALUE_PROP_1_EVIDENCE|interview, competitor analysis, or data}} |
| 2 | {{VALUE_PROP_2|secondary benefit this application provides}} | {{VALUE_PROP_2_METRIC}} | {{VALUE_PROP_2_EVIDENCE}} |
| 3 | {{VALUE_PROP_3|advantage over the alternatives}} | {{VALUE_PROP_3_METRIC}} | {{VALUE_PROP_3_EVIDENCE}} |

## Scope

| Area | In scope | Out of scope | Reason for the exclusion |
|------|----------|--------------|---------------------------|
| User interface | {{UI_IN_SCOPE}} | {{UI_OUT_OF_SCOPE}} | {{UI_EXCLUSION_REASON}} |
| Server behaviour | {{API_IN_SCOPE}} | {{API_OUT_OF_SCOPE}} | {{API_EXCLUSION_REASON}} |
| Data | {{DATA_IN_SCOPE}} | {{DATA_OUT_OF_SCOPE}} | {{DATA_EXCLUSION_REASON}} |
| Integrations | {{INTEGRATION_IN_SCOPE}} | {{INTEGRATION_OUT_OF_SCOPE}} | {{INTEGRATION_EXCLUSION_REASON}} |
| Reporting | {{REPORTING_IN_SCOPE}} | {{REPORTING_OUT_OF_SCOPE}} | {{REPORTING_EXCLUSION_REASON}} |
| Operations | {{OPS_IN_SCOPE}} | {{OPS_OUT_OF_SCOPE}} | {{OPS_EXCLUSION_REASON}} |

## Functional Requirements

### Must have

| ID | Capability | Requirement | Acceptance criteria |
|----|------------|-------------|---------------------|
| FR-01 | Authentication | {{FR01_AUTH_REQUIREMENT}} | {{FR01_AUTH_CRITERIA}} |
| FR-02 | Profile and preferences | {{FR02_PROFILE_REQUIREMENT}} | {{FR02_PROFILE_CRITERIA}} |
| FR-03 | Core business logic | {{FR03_CORE_REQUIREMENT}} | {{FR03_CORE_CRITERIA}} |
| FR-04 | Data management | {{FR04_DATA_REQUIREMENT}} | {{FR04_DATA_CRITERIA}} |
| FR-05 | Search and filtering | {{FR05_SEARCH_REQUIREMENT}} | {{FR05_SEARCH_CRITERIA}} |
| FR-06 | Notifications | {{FR06_NOTIFY_REQUIREMENT}} | {{FR06_NOTIFY_CRITERIA}} |

### Should have

| ID | Capability | Requirement | Acceptance criteria |
|----|------------|-------------|---------------------|
| FR-07 | Real-time updates | {{FR07_REALTIME_REQUIREMENT}} | {{FR07_REALTIME_CRITERIA}} |
| FR-08 | File upload and management | {{FR08_FILE_REQUIREMENT}} | {{FR08_FILE_CRITERIA}} |
| FR-09 | Reporting and analytics | {{FR09_REPORTING_REQUIREMENT}} | {{FR09_REPORTING_CRITERIA}} |
| FR-10 | External integrations | {{FR10_INTEGRATION_REQUIREMENT}} | {{FR10_INTEGRATION_CRITERIA}} |
| FR-11 | Responsive layouts | {{FR11_RESPONSIVE_REQUIREMENT}} | {{FR11_RESPONSIVE_CRITERIA}} |

### Could have

| ID | Capability | Requirement | Acceptance criteria | Trigger to build it |
|----|------------|-------------|---------------------|----------------------|
| FR-12 | Offline behaviour | {{FR12_OFFLINE_REQUIREMENT}} | {{FR12_OFFLINE_CRITERIA}} | {{FR12_OFFLINE_TRIGGER}} |
| FR-13 | Localisation | {{FR13_I18N_REQUIREMENT}} | {{FR13_I18N_CRITERIA}} | {{FR13_I18N_TRIGGER}} |
| FR-14 | Advanced authentication | {{FR14_2FA_REQUIREMENT}} | {{FR14_2FA_CRITERIA}} | {{FR14_2FA_TRIGGER}} |
| FR-15 | Assisted features | {{FR15_ASSISTED_REQUIREMENT}} | {{FR15_ASSISTED_CRITERIA}} | {{FR15_ASSISTED_TRIGGER}} |

## Technical Architecture

### Technology stack

| Layer | Choice | Version | Why this and not the alternative |
|-------|--------|---------|---------------------------------|
| Frontend framework | {{FRONTEND_FRAMEWORK}} | {{FRONTEND_VERSION}} | {{FRONTEND_RATIONALE}} |
| Styling approach | {{STYLING_CHOICE}} | {{STYLING_VERSION}} | {{STYLING_RATIONALE}} |
| Client state | {{CLIENT_STATE_CHOICE}} | {{CLIENT_STATE_VERSION}} | {{CLIENT_STATE_RATIONALE}} |
| Backend runtime | {{BACKEND_RUNTIME}} | {{BACKEND_VERSION}} | {{BACKEND_RATIONALE}} |
| Backend framework | {{BACKEND_FRAMEWORK}} | {{BACKEND_FRAMEWORK_VERSION}} | {{BACKEND_FRAMEWORK_RATIONALE}} |
| Database | {{DATABASE_CHOICE}} | {{DATABASE_VERSION}} | {{DATABASE_RATIONALE}} |
| Hosting | {{HOSTING_CHOICE}} | {{HOSTING_REGION}} | {{HOSTING_RATIONALE}} |
| CI/CD | {{CICD_CHOICE}} | n/a | {{CICD_RATIONALE}} |
| Monitoring | {{MONITORING_CHOICE}} | n/a | {{MONITORING_RATIONALE}} |

### System components

| Component | Responsibility | Holds state | Scaling approach |
|-----------|----------------|--------------|------------------|
| Web client | {{CLIENT_RESPONSIBILITY}} | {{CLIENT_STATE|no}} | {{CLIENT_SCALING}} |
| API gateway | {{GATEWAY_RESPONSIBILITY}} | {{GATEWAY_STATE|no}} | {{GATEWAY_SCALING}} |
| Application services | {{SERVICE_RESPONSIBILITY}} | {{SERVICE_STATE|yes}} | {{SERVICE_SCALING}} |
| Database | {{DATABASE_RESPONSIBILITY}} | {{DATABASE_STATE|yes}} | {{DATABASE_SCALING}} |
| Object storage | {{STORAGE_RESPONSIBILITY}} | {{STORAGE_STATE|yes}} | {{STORAGE_SCALING}} |
| Notification delivery | {{NOTIFY_RESPONSIBILITY}} | {{NOTIFY_STATE|no}} | {{NOTIFY_SCALING}} |
| Analytics | {{ANALYTICS_RESPONSIBILITY}} | {{ANALYTICS_STATE|yes}} | {{ANALYTICS_SCALING}} |

### Data model

| Entity | Key fields | Relationships | Retention | Volume estimate |
|--------|------------|---------------|-----------|-----------------|
| {{ENTITY_1_NAME}} | {{ENTITY_1_FIELDS}} | {{ENTITY_1_RELATIONSHIPS}} | {{ENTITY_1_RETENTION}} | {{ENTITY_1_VOLUME}} |
| {{ENTITY_2_NAME}} | {{ENTITY_2_FIELDS}} | {{ENTITY_2_RELATIONSHIPS}} | {{ENTITY_2_RETENTION}} | {{ENTITY_2_VOLUME}} |
| {{ENTITY_3_NAME}} | {{ENTITY_3_FIELDS}} | {{ENTITY_3_RELATIONSHIPS}} | {{ENTITY_3_RETENTION}} | {{ENTITY_3_VOLUME}} |

System data is held separately from business data: {{SYSTEM_DATA_INVENTORY|configuration, feature flags, audit trail, and application metrics}}

### HTTP interface

| Method | Path | Purpose | Auth required | Owner |
|--------|------|---------|----------------|-------|
| {{METHOD_1}} | {{PATH_1}} | {{ENDPOINT_1_PURPOSE}} | {{ENDPOINT_1_AUTH}} | {{ENDPOINT_1_OWNER}} |
| {{METHOD_2}} | {{PATH_2}} | {{ENDPOINT_2_PURPOSE}} | {{ENDPOINT_2_AUTH}} | {{ENDPOINT_2_OWNER}} |
| {{METHOD_3}} | {{PATH_3}} | {{ENDPOINT_3_PURPOSE}} | {{ENDPOINT_3_AUTH}} | {{ENDPOINT_3_OWNER}} |

The full interface is published as an OpenAPI 3.1 document at {{OPENAPI_LOCATION|path or URL}}.

## User Experience and Accessibility

### Personas and journeys

| Persona | Goal | Entry point | Success signal | Failure seen today |
|---------|------|-------------|----------------|--------------------|
| {{PERSONA_1}} | {{PERSONA_1_GOAL}} | {{PERSONA_1_ENTRY}} | {{PERSONA_1_SUCCESS}} | {{PERSONA_1_CURRENT_PAIN}} |
| {{PERSONA_2}} | {{PERSONA_2_GOAL}} | {{PERSONA_2_ENTRY}} | {{PERSONA_2_SUCCESS}} | {{PERSONA_2_CURRENT_PAIN}} |

### Accessibility requirements

Verification column stays empty until a check has actually run; a check that
cannot run is reported as skipped, never as passed.

| Requirement | Target | Verification method | Verified on |
|-------------|--------|---------------------|-------------|
| Contrast | WCAG 2.1 AA | {{A11Y_CONTRAST_CHECK}} | {{A11Y_CONTRAST_DATE}} |
| Keyboard navigation | WCAG 2.1 AA | {{A11Y_KEYBOARD_CHECK}} | {{A11Y_KEYBOARD_DATE}} |
| Screen reader output | WCAG 2.1 AA | {{A11Y_SCREENREADER_CHECK}} | {{A11Y_SCREENREADER_DATE}} |
| Focus visibility | WCAG 2.1 AA | {{A11Y_FOCUS_CHECK}} | {{A11Y_FOCUS_DATE}} |
| Colour independence | WCAG 2.1 AA | {{A11Y_COLOUR_CHECK}} | {{A11Y_COLOUR_DATE}} |

### Supported devices and browsers

| Surface | Minimum supported | Latest tested | Owner of the test |
|---------|-------------------|----------------|--------------------|
| {{SURFACE_1}} | {{SURFACE_1_MIN}} | {{SURFACE_1_LATEST}} | {{SURFACE_1_OWNER}} |
| {{SURFACE_2}} | {{SURFACE_2_MIN}} | {{SURFACE_2_LATEST}} | {{SURFACE_2_OWNER}} |

## Security and Privacy

### Controls

| Control | Mechanism | Verified by |
|---------|-----------|-------------|
| Authentication | {{AUTHN_MECHANISM}} | {{AUTHN_VERIFICATION}} |
| Authorisation | {{AUTHZ_MECHANISM|RBAC or scoped claims}} | {{AUTHZ_VERIFICATION}} |
| Encryption at rest | {{ENCRYPTION_AT_REST}} | {{ENCRYPTION_AT_REST_VERIFICATION}} |
| Encryption in transit | {{ENCRYPTION_IN_TRANSIT|TLS 1.3}} | {{ENCRYPTION_IN_TRANSIT_VERIFICATION}} |
| Input validation | {{INPUT_VALIDATION_APPROACH}} | {{INPUT_VALIDATION_VERIFICATION}} |
| Rate limiting | {{RATE_LIMIT_RULE}} | {{RATE_LIMIT_VERIFICATION}} |
| Security headers | {{SECURITY_HEADERS}} | {{SECURITY_HEADERS_VERIFICATION}} |
| Vulnerability management | {{VULN_MANAGEMENT_APPROACH}} | {{VULN_MANAGEMENT_VERIFICATION}} |

### Personal data

| Data category | Purpose | Lawful basis | Retention | Deletion path |
|---------------|---------|--------------|-----------|---------------|
| {{DATA_CATEGORY_1}} | {{DATA_1_PURPOSE}} | {{DATA_1_BASIS}} | {{DATA_1_RETENTION}} | {{DATA_1_DELETION}} |
| {{DATA_CATEGORY_2}} | {{DATA_2_PURPOSE}} | {{DATA_2_BASIS}} | {{DATA_2_RETENTION}} | {{DATA_2_DELETION}} |

- Consent mechanism: {{CONSENT_MECHANISM|how opt-in and withdrawal are captured}}
- Applicable regulation: {{APPLICABLE_REGULATION|GDPR, or the regime that applies}}
- Data protection impact assessment: {{DPIA_REFERENCE|path to the DPIA, or not required}}

## Testing Strategy

| Level | Scope | Target | Tools |
|-------|-------|--------|-------|
| Unit | Functions and components | {{TEST_COVERAGE_MIN\|90}}% | {{UNIT_TEST_TOOLS}} |
| Integration | API endpoints and database interaction | {{INTEGRATION_COVERAGE_TARGET}} | {{INTEGRATION_TEST_TOOLS}} |
| End to end | Complete user journeys | {{E2E_JOURNEY_COUNT}} journeys | {{E2E_TEST_TOOLS}} |
| Performance | Load and stress | {{PERFORMANCE_TEST_PROFILE}} | {{PERFORMANCE_TEST_TOOLS}} |
| Accessibility | Automated and manual audit | WCAG 2.1 AA | {{A11Y_TEST_TOOLS}} |

### Critical journeys

- [ ] {{JOURNEY_1|end-to-end path that must never break}}
- [ ] {{JOURNEY_2}}
- [ ] {{JOURNEY_3}}
- [ ] {{JOURNEY_4}}

## Delivery

### Environments

| Environment | Purpose | Data | External services |
|-------------|---------|------|--------------------|
| Development | {{DEV_PURPOSE}} | {{DEV_DATA}} | {{DEV_EXTERNALS|mocked}} |
| Staging | {{STAGING_PURPOSE}} | {{STAGING_DATA}} | {{STAGING_EXTERNALS|sandbox}} |
| Production | {{PROD_PURPOSE}} | {{PROD_DATA}} | {{PROD_EXTERNALS|live}} |

### Pipeline stages

| # | Stage | Gate that must pass before the next stage |
|---|-------|---------------------------------------------|
| 1 | {{PIPELINE_STAGE_1}} | {{PIPELINE_GATE_1}} |
| 2 | {{PIPELINE_STAGE_2}} | {{PIPELINE_GATE_2}} |
| 3 | {{PIPELINE_STAGE_3}} | {{PIPELINE_GATE_3}} |
| 4 | {{PIPELINE_STAGE_4}} | {{PIPELINE_GATE_4}} |
| 5 | {{PIPELINE_STAGE_5}} | {{PIPELINE_GATE_5}} |

Rollback: {{ROLLBACK_STRATEGY|how a failed release is reverted and how long it takes}}

### Infrastructure requirements

| Resource | Requirement | Basis for the number |
|----------|-------------|----------------------|
| Compute | {{COMPUTE_REQUIREMENT}} | {{COMPUTE_BASIS}} |
| Storage | {{STORAGE_REQUIREMENT}} | {{STORAGE_BASIS}} |
| Network | {{NETWORK_REQUIREMENT}} | {{NETWORK_BASIS}} |
| Scaling policy | {{SCALING_POLICY}} | {{SCALING_BASIS}} |
| Backup and restore | {{BACKUP_REQUIREMENT}} | {{BACKUP_BASIS}} |

### Milestones

- [ ] {{MILESTONE_1|foundation: environment, schema, first tested path}} — target {{MILESTONE_1_DATE}}
- [ ] {{MILESTONE_2|core features}} — target {{MILESTONE_2_DATE}}
- [ ] {{MILESTONE_3|enhanced features and hardening}} — target {{MILESTONE_3_DATE}}
- [ ] {{MILESTONE_4|launch readiness}} — target {{MILESTONE_4_DATE}}

## Success Metrics

### Technical

| Metric | Target | Measured by | Current value |
|--------|--------|-------------|---------------|
| Response time | {{PERFORMANCE_TARGET}} | {{PERFORMANCE_MEASUREMENT}} | {{PERFORMANCE_CURRENT}} |
| Availability | {{AVAILABILITY_TARGET}} | {{AVAILABILITY_MEASUREMENT}} | {{AVAILABILITY_CURRENT}} |
| Capacity | {{CAPACITY_TARGET}} | {{CAPACITY_MEASUREMENT}} | {{CAPACITY_CURRENT}} |
| Known severe vulnerabilities | {{SECURITY_BAR}} | {{SECURITY_MEASUREMENT}} | {{SECURITY_CURRENT}} |
| Test coverage | {{TEST_COVERAGE_MIN\|90}}% | {{COVERAGE_MEASUREMENT}} | {{COVERAGE_CURRENT}} |

### Business

| Metric | Target | Measured by | Current value |
|--------|--------|-------------|---------------|
| Adoption | {{ADOPTION_TARGET}} | {{ADOPTION_MEASUREMENT}} | {{ADOPTION_CURRENT}} |
| Engagement | {{ENGAGEMENT_TARGET}} | {{ENGAGEMENT_MEASUREMENT}} | {{ENGAGEMENT_CURRENT}} |
| Feature usage | {{FEATURE_USAGE_TARGET}} | {{FEATURE_USAGE_MEASUREMENT}} | {{FEATURE_USAGE_CURRENT}} |
| Satisfaction | {{SATISFACTION_TARGET}} | {{SATISFACTION_MEASUREMENT}} | {{SATISFACTION_CURRENT}} |

### Operational

| Metric | Target | Measured by |
|--------|--------|-------------|
| Deployment frequency | {{DEPLOY_FREQUENCY_TARGET}} | {{DEPLOY_FREQUENCY_MEASUREMENT}} |
| Lead time | {{LEAD_TIME_TARGET}} | {{LEAD_TIME_MEASUREMENT}} |
| Mean time to recovery | {{MTTR_TARGET}} | {{MTTR_MEASUREMENT}} |
| Change failure rate | {{CFR_TARGET}} | {{CFR_MEASUREMENT}} |

## Open Questions

| # | Question | Why it matters | Owner | Blocking what | Resolution date |
|---|----------|----------------|-------|----------------|-----------------|
| Q1 | {{QUESTION_1}} | {{QUESTION_1_IMPACT}} | {{QUESTION_1_OWNER}} | {{QUESTION_1_BLOCKS}} | {{QUESTION_1_DUE}} |
| Q2 | {{QUESTION_2}} | {{QUESTION_2_IMPACT}} | {{QUESTION_2_OWNER}} | {{QUESTION_2_BLOCKS}} | {{QUESTION_2_DUE}} |
| Q3 | {{QUESTION_3}} | {{QUESTION_3_IMPACT}} | {{QUESTION_3_OWNER}} | {{QUESTION_3_BLOCKS}} | {{QUESTION_3_DUE}} |
| Q4 | {{QUESTION_4}} | {{QUESTION_4_IMPACT}} | {{QUESTION_4_OWNER}} | {{QUESTION_4_BLOCKS}} | {{QUESTION_4_DUE}} |

## Completion Checklist

Nothing below is true when this file is created. Each box is ticked only when the
thing it names has actually happened; an unticked box is the honest state.

- [ ] Every section above is filled in, not merely headed.
- [ ] Every `{{PLACEHOLDER}}` token in this document has been replaced.
- [ ] Every requirement carries acceptance criteria a tester could execute.
- [ ] Every metric carries a target and the name of the thing that measures it.
- [ ] Every assumption is labelled as fact, inferred, speculative, or unknown.
- [ ] Every threshold used here matches the active-dimension values in the constitution.
- [ ] Every verification column above records a check that ran, or is marked skipped with a reason.

## Handover

| Item | Value |
|------|-------|
| Next artifact | design.template.md |
| Reviewers | {{REVIEWERS}} |
| Approval required from | {{APPROVERS}} |
| Estimated timeline | {{ESTIMATED_TIMELINE}} |
| Team requirements | {{TEAM_REQUIREMENTS}} |
