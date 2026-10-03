# Research Project Specification: {{RESEARCH_TITLE}}

**Created**: {{DATE}}
**Status**: {{DOC_STATUS|draft}}
**Type**: Research or analysis project
**Source request**: {{SOURCE_REQUEST|verbatim text that prompted this research}}
**Principal investigator**: {{PRINCIPAL_INVESTIGATOR}}
**Reviewers**: {{REVIEWERS}}
**Constitution dimensions in force**: {{ACTIVE_DIMENSIONS|Business, Product}}
**Related specs**: {{RELATED_SPEC_IDS|spec directory numbers this research supports}}

## Working constraints

Each constraint below binds the research, not this document. Every one is either
satisfied by a named section further down or it is not.

- The research question is stated before any data is collected, and a question
  changed after collection is reported as a change, not quietly adopted.
- The method is chosen for the question, not for familiarity; a method that
  cannot answer the question is not a method.
- Every source carries an identifier and a retrieval date, and every claim
  traces to one.
- Assumptions and limitations are written down while they are still cheap to
  change, not in a section added at the end.
- Findings that disconfirm the preferred answer are reported with the same
  prominence as findings that support it.
- A result that cannot be reproduced from what is written here is not a result.

## Research Question and Objectives

### Question

| Role | Question | Answerable by |
|------|----------|---------------|
| Primary | {{PRIMARY_QUESTION|one question the whole study exists to answer}} | {{PRIMARY_QUESTION_METHOD}} |
| Secondary 1 | {{SECONDARY_QUESTION_1|supporting question}} | {{SECONDARY_QUESTION_1_METHOD}} |
| Secondary 2 | {{SECONDARY_QUESTION_2|supporting question}} | {{SECONDARY_QUESTION_2_METHOD}} |

### Hypotheses

| ID | Hypothesis | Type | Falsified when | Predicted direction |
|----|------------|------|----------------|---------------------|
| H1 | {{HYPOTHESIS_1|statement that can be wrong}} | {{HYPOTHESIS_1_TYPE\|directional or null}} | {{HYPOTHESIS_1_FALSIFIER}} | {{HYPOTHESIS_1_PREDICTION}} |
| H2 | {{HYPOTHESIS_2}} | {{HYPOTHESIS_2_TYPE}} | {{HYPOTHESIS_2_FALSIFIER}} | {{HYPOTHESIS_2_PREDICTION}} |
| H3 | {{HYPOTHESIS_3}} | {{HYPOTHESIS_3_TYPE}} | {{HYPOTHESIS_3_FALSIFIER}} | {{HYPOTHESIS_3_PREDICTION}} |

### Context and motivation

{{CONTEXT_AND_MOTIVATION|why this question is worth answering now, and what gap it fills}}

### Expected outcomes

| # | Outcome | Form | How it will be used | Decision it informs |
|---|---------|------|--------------------|---------------------|
| O1 | {{OUTCOME_1|finding or deliverable}} | {{OUTCOME_1_FORM\|report section, dataset, recommendation}} | {{OUTCOME_1_USE}} | {{OUTCOME_1_DECISION}} |
| O2 | {{OUTCOME_2}} | {{OUTCOME_2_FORM}} | {{OUTCOME_2_USE}} | {{OUTCOME_2_DECISION}} |
| O3 | {{OUTCOME_3}} | {{OUTCOME_3_FORM}} | {{OUTCOME_3_USE}} | {{OUTCOME_3_DECISION}} |

## Scope and Boundaries

| Dimension | In scope | Out of scope | Consequence of the exclusion |
|-----------|----------|--------------|-------------------------------|
| Subject areas | {{SUBJECT_IN_SCOPE}} | {{SUBJECT_OUT_OF_SCOPE}} | {{SUBJECT_EXCLUSION_EFFECT}} |
| Time period | {{TIME_PERIOD_IN_SCOPE}} | {{TIME_PERIOD_OUT_OF_SCOPE}} | {{TIME_PERIOD_EFFECT}} |
| Geography or market | {{GEOGRAPHY_IN_SCOPE}} | {{GEOGRAPHY_OUT_OF_SCOPE}} | {{GEOGRAPHY_EFFECT}} |
| Population | {{POPULATION_IN_SCOPE}} | {{POPULATION_OUT_OF_SCOPE}} | {{POPULATION_EFFECT}} |
| Data sources | {{DATA_SOURCES_IN_SCOPE}} | {{DATA_SOURCES_OUT_OF_SCOPE}} | {{DATA_SOURCES_EFFECT}} |
| Decisions supported | {{DECISIONS_SUPPORTED}} | {{DECISIONS_NOT_SUPPORTED}} | {{DECISIONS_EFFECT}} |

Key concepts and their operational definitions — the boundary of each term decides
what counts as data:

| Term | Operational definition | Excluded cases |
|------|------------------------|----------------|
| {{TERM_1}} | {{TERM_1_DEFINITION}} | {{TERM_1_EXCLUSIONS}} |
| {{TERM_2}} | {{TERM_2_DEFINITION}} | {{TERM_2_EXCLUSIONS}} |

## Literature and Prior Art

### Search strategy

| Element | Plan |
|---------|------|
| Databases and archives | {{SEARCH_SOURCES}} |
| Query strings | {{SEARCH_QUERIES|exact strings issued, per source}} |
| Date range | {{SEARCH_DATE_RANGE}} |
| Inclusion criteria | {{SEARCH_INCLUSION_CRITERIA}} |
| Exclusion criteria | {{SEARCH_EXCLUSION_CRITERIA}} |
| Screening process | {{SEARCH_SCREENING_PROCESS\|how many reviewers, and how disagreements resolve}} |
| Search executed on | {{SEARCH_DATE}} |

### Prior work

| # | Source | Identifier | What it establishes | Gap it leaves |
|---|--------|-----------|---------------------|---------------|
| R1 | {{REFERENCE_1_TITLE}} | {{REFERENCE_1_ID\|DOI, ISBN, or URL}} | {{REFERENCE_1_FINDING}} | {{REFERENCE_1_GAP}} |
| R2 | {{REFERENCE_2_TITLE}} | {{REFERENCE_2_ID}} | {{REFERENCE_2_FINDING}} | {{REFERENCE_2_GAP}} |
| R3 | {{REFERENCE_3_TITLE}} | {{REFERENCE_3_ID}} | {{REFERENCE_3_FINDING}} | {{REFERENCE_3_GAP}} |
| R4 | {{REFERENCE_4_TITLE}} | {{REFERENCE_4_ID}} | {{REFERENCE_4_FINDING}} | {{REFERENCE_4_GAP}} |

Synthesis: {{LITERATURE_SYNTHESIS|what the prior work collectively shows, and therefore what this study must add}}

## Methodology

### Design

| Element | Decision | Rationale |
|---------|----------|-----------|
| Research design | {{DESIGN_TYPE\|exploratory, descriptive, explanatory, experimental}} | {{DESIGN_TYPE_RATIONALE}} |
| Approach | {{APPROACH\|qualitative, quantitative, or mixed methods}} | {{APPROACH_RATIONALE}} |
| Strategy | {{STRATEGY\|case study, survey, experiment, ethnography, or other}} | {{STRATEGY_RATIONALE}} |
| Unit of analysis | {{UNIT_OF_ANALYSIS}} | {{UNIT_OF_ANALYSIS_RATIONALE}} |
| Timeline of the study | {{STUDY_TIMELINE}} | {{STUDY_TIMELINE_RATIONALE}} |

### Sampling

| Element | Decision | Rationale |
|---------|----------|-----------|
| Population | {{POPULATION}} | {{POPULATION_RATIONALE}} |
| Sampling frame | {{SAMPLE_FRAME}} | {{SAMPLE_FRAME_RATIONALE}} |
| Sampling method | {{SAMPLING_METHOD}} | {{SAMPLING_METHOD_RATIONALE}} |
| Sample size | {{SAMPLE_SIZE}} | {{SAMPLE_SIZE_JUSTIFICATION\|power analysis or precision argument}} |
| Recruitment | {{RECRUITMENT}} | {{RECRUITMENT_RATIONALE}} |
| Non-response handling | {{NON_RESPONSE_HANDLING}} | {{NON_RESPONSE_RATIONALE}} |

### Instruments

| Instrument | Measures | Format | Piloted | Known limitations |
|------------|----------|--------|---------|-------------------|
| {{INSTRUMENT_1}} | {{INSTRUMENT_1_MEASURES}} | {{INSTRUMENT_1_FORMAT\|questionnaire, interview guide, rubric, sensor}} | {{INSTRUMENT_1_PILOT_STATUS}} | {{INSTRUMENT_1_LIMITATIONS}} |
| {{INSTRUMENT_2}} | {{INSTRUMENT_2_MEASURES}} | {{INSTRUMENT_2_FORMAT}} | {{INSTRUMENT_2_PILOT_STATUS}} | {{INSTRUMENT_2_LIMITATIONS}} |

### Data collection plan

| Source | What is collected | Volume | Method | When | Owner |
|--------|-------------------|--------|--------|------|-------|
| {{COLLECTION_1_SOURCE}} | {{COLLECTION_1_DATA}} | {{COLLECTION_1_VOLUME}} | {{COLLECTION_1_METHOD}} | {{COLLECTION_1_WHEN}} | {{COLLECTION_1_OWNER}} |
| {{COLLECTION_2_SOURCE}} | {{COLLECTION_2_DATA}} | {{COLLECTION_2_VOLUME}} | {{COLLECTION_2_METHOD}} | {{COLLECTION_2_WHEN}} | {{COLLECTION_2_OWNER}} |

## Ethics, Consent, and Data Handling

| Requirement | Decision | Evidence kept on file |
|-------------|----------|------------------------|
| Ethics review | {{ETHICS_REVIEW\|required and approved, exempt, or pending}} | {{ETHICS_REVIEW_EVIDENCE}} |
| Informed consent | {{CONSENT_PROCESS}} | {{CONSENT_EVIDENCE}} |
| Withdrawals | {{WITHDRAWAL_POLICY}} | {{WITHDRAWAL_EVIDENCE}} |
| Privacy protection | {{PRIVACY_PROTECTION\|anonymisation or pseudonymisation scheme}} | {{PRIVACY_EVIDENCE}} |
| Data minimisation | {{DATA_MINIMISATION}} | {{DATA_MINIMISATION_EVIDENCE}} |
| Storage location | {{STORAGE_LOCATION}} | {{STORAGE_EVIDENCE}} |
| Retention period | {{RETENTION_PERIOD}} | {{RETENTION_EVIDENCE}} |
| Secure deletion | {{DELETION_PROCEDURE}} | {{DELETION_EVIDENCE}} |
| Sharing with third parties | {{THIRD_PARTY_SHARING}} | {{THIRD_PARTY_SHARING_EVIDENCE}} |

## Analysis Plan

### Quantitative

| Analysis | Data it uses | Test or model | Assumptions checked | Decision rule |
|----------|--------------|---------------|---------------------|---------------|
| {{ANALYSIS_1}} | {{ANALYSIS_1_DATA}} | {{ANALYSIS_1_TEST}} | {{ANALYSIS_1_ASSUMPTIONS}} | {{ANALYSIS_1_DECISION_RULE\|alpha, effect size, or threshold}} |
| {{ANALYSIS_2}} | {{ANALYSIS_2_DATA}} | {{ANALYSIS_2_TEST}} | {{ANALYSIS_2_ASSUMPTIONS}} | {{ANALYSIS_2_DECISION_RULE}} |

### Qualitative

| Analysis | Data it uses | Procedure | Saturation or stopping rule |
|----------|--------------|-----------|------------------------------|
| {{QUAL_ANALYSIS_1}} | {{QUAL_ANALYSIS_1_DATA}} | {{QUAL_ANALYSIS_1_PROCEDURE}} | {{QUAL_ANALYSIS_1_STOPPING_RULE}} |
| {{QUAL_ANALYSIS_2}} | {{QUAL_ANALYSIS_2_DATA}} | {{QUAL_ANALYSIS_2_PROCEDURE}} | {{QUAL_ANALYSIS_2_STOPPING_RULE}} |

### Integration and robustness

| Aspect | Decision |
|--------|----------|
| Integration design | {{INTEGRATION_DESIGN\|convergent, sequential, or transformative}} |
| How the strands are reconciled | {{INTEGRATION_RECONCILIATION}} |
| Sensitivity analysis | {{SENSITIVITY_ANALYSIS\|which results move under plausible alternative specifications}} |
| Missing data treatment | {{MISSING_DATA_TREATMENT}} |
| Confounders and controls | {{CONFOUNDER_CONTROL}} |
| Disconfirming evidence sought | {{DISCONFIRMING_EVIDENCE}} |

### Reproducibility

| Requirement | Decision |
|-------------|----------|
| Materials available to others | {{REPRODUCIBILITY_MATERIALS}} |
| Code or scripts | {{REPRODUCIBILITY_CODE}} |
| Analysis code coverage, where the analysis is implemented in code | {{TEST_COVERAGE_MIN\|90}}% |
| Independent re-analysis | {{REPRODUCIBILITY_INDEPENDENT_REVIEW\|who re-runs it, and when}} |

## Quality Assurance

| Threat to validity | Type | Mitigation | Check performed |
|--------------------|------|------------|-----------------|
| {{THREAT_1}} | {{THREAT_1_TYPE\|internal, external, construct, or reliability}} | {{THREAT_1_MITIGATION}} | {{THREAT_1_CHECK}} |
| {{THREAT_2}} | {{THREAT_2_TYPE}} | {{THREAT_2_MITIGATION}} | {{THREAT_2_CHECK}} |
| {{THREAT_3}} | {{THREAT_3_TYPE}} | {{THREAT_3_MITIGATION}} | {{THREAT_3_CHECK}} |

| Bias | Mitigation in place | Residual risk accepted |
|------|---------------------|-------------------------|
| {{BIAS_1\|selection, confirmation, observer, response, publication}} | {{BIAS_1_MITIGATION}} | {{BIAS_1_RESIDUAL}} |
| {{BIAS_2}} | {{BIAS_2_MITIGATION}} | {{BIAS_2_RESIDUAL}} |

- Peer review: {{PEER_REVIEW_PLAN|who reviews, against which criteria, and when}}
- Expert validation: {{EXPERT_VALIDATION_PLAN}}
- Interim analysis point, if any: {{INTERIM_ANALYSIS_POINT}}

## Timeline and Milestones

| Milestone | Target date | Exit criteria | Depends on |
|-----------|-------------|---------------|------------|
| {{MILESTONE_1}} | {{MILESTONE_1_DATE}} | {{MILESTONE_1_EXIT}} | {{MILESTONE_1_DEPENDS}} |
| {{MILESTONE_2}} | {{MILESTONE_2_DATE}} | {{MILESTONE_2_EXIT}} | {{MILESTONE_2_DEPENDS}} |
| {{MILESTONE_3}} | {{MILESTONE_3_DATE}} | {{MILESTONE_3_EXIT}} | {{MILESTONE_3_DEPENDS}} |
| {{MILESTONE_4}} | {{MILESTONE_4_DATE}} | {{MILESTONE_4_EXIT}} | {{MILESTONE_4_DEPENDS}} |
| {{MILESTONE_5}} | {{MILESTONE_5_DATE}} | {{MILESTONE_5_EXIT}} | {{MILESTONE_5_DEPENDS}} |

## Deliverables

| Deliverable | Contents | Audience | Format and length | Due |
|-------------|----------|----------|-------------------|-----|
| {{DELIVERABLE_1}} | {{DELIVERABLE_1_CONTENTS}} | {{DELIVERABLE_1_AUDIENCE}} | {{DELIVERABLE_1_FORMAT}} | {{DELIVERABLE_1_DUE}} |
| {{DELIVERABLE_2}} | {{DELIVERABLE_2_CONTENTS}} | {{DELIVERABLE_2_AUDIENCE}} | {{DELIVERABLE_2_FORMAT}} | {{DELIVERABLE_2_DUE}} |
| {{DELIVERABLE_3}} | {{DELIVERABLE_3_CONTENTS}} | {{DELIVERABLE_3_AUDIENCE}} | {{DELIVERABLE_3_FORMAT}} | {{DELIVERABLE_3_DUE}} |
| {{DELIVERABLE_4}} | {{DELIVERABLE_4_CONTENTS}} | {{DELIVERABLE_4_AUDIENCE}} | {{DELIVERABLE_4_FORMAT}} | {{DELIVERABLE_4_DUE}} |

### Dissemination

| Channel | Audience | Timing | Materials required |
|---------|----------|--------|---------------------|
| {{CHANNEL_1}} | {{CHANNEL_1_AUDIENCE}} | {{CHANNEL_1_TIMING}} | {{CHANNEL_1_MATERIALS}} |
| {{CHANNEL_2}} | {{CHANNEL_2_AUDIENCE}} | {{CHANNEL_2_TIMING}} | {{CHANNEL_2_MATERIALS}} |

## Impact and Follow-on

| Aspect | Statement |
|--------|-----------|
| Who benefits | {{IMPACT_BENEFICIARIES}} |
| What changes as a result | {{IMPACT_CHANGE}} |
| What the work enables next | {{IMPACT_ENABLES}} |
| Questions left open | {{IMPACT_OPEN_QUESTIONS}} |

## Risks and Limitations

| Risk | Probability | Impact | Mitigation | Owner |
|------|-------------|--------|------------|-------|
| {{RISK_1\|e.g. low response rate}} | {{RISK_1_PROBABILITY\|low, medium, or high}} | {{RISK_1_IMPACT}} | {{RISK_1_MITIGATION}} | {{RISK_1_OWNER}} |
| {{RISK_2}} | {{RISK_2_PROBABILITY}} | {{RISK_2_IMPACT}} | {{RISK_2_MITIGATION}} | {{RISK_2_OWNER}} |
| {{RISK_3}} | {{RISK_3_PROBABILITY}} | {{RISK_3_IMPACT}} | {{RISK_3_MITIGATION}} | {{RISK_3_OWNER}} |
| {{RISK_4}} | {{RISK_4_PROBABILITY}} | {{RISK_4_IMPACT}} | {{RISK_4_MITIGATION}} | {{RISK_4_OWNER}} |

| Known limitation | Consequence for the conclusions |
|-------------------|---------------------------------|
| {{LIMITATION_1\|sample size or representativeness}} | {{LIMITATION_1_CONSEQUENCE}} |
| {{LIMITATION_2\|methodological constraint}} | {{LIMITATION_2_CONSEQUENCE}} |
| {{LIMITATION_3}} | {{LIMITATION_3_CONSEQUENCE}} |

| Contingency | Trigger | Reduced design | Decision owner |
|-------------|---------|----------------|----------------|
| Plan A | {{PLAN_A_TRIGGER}} | {{PLAN_A_DESIGN}} | {{PLAN_A_OWNER}} |
| Plan B | {{PLAN_B_TRIGGER}} | {{PLAN_B_DESIGN}} | {{PLAN_B_OWNER}} |
| Plan C | {{PLAN_C_TRIGGER}} | {{PLAN_C_DESIGN}} | {{PLAN_C_OWNER}} |

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
- [ ] Every hypothesis has a stated falsification condition.
- [ ] Every external claim carries a source and a retrieval date.
- [ ] Every instrument has been piloted, or its lack of piloting is recorded.
- [ ] Every validity threat has a mitigation and a check.
- [ ] Every assumption and limitation is labelled as fact, inferred, speculative, or unknown.
- [ ] Data handling meets the consent and retention decisions recorded above.

## Handover

| Item | Value |
|------|-------|
| Next artifact | data collection instruments and schedule |
| Peer reviewers | {{PEER_REVIEWERS}} |
| Approval required from | {{APPROVERS}} |
| Estimated duration | {{ESTIMATED_DURATION}} |
| Required resources | {{REQUIRED_RESOURCES}} |
