# Requirements Traceability Matrix (RTM)

**Project**: {{PROJECT_NAME}}
**Feature**: {{FEATURE_NAME}}
**Version**: {{VERSION}}
**Date**: {{DATE}}
**Prepared by**: {{AUTHOR}}

## Executive Summary

### Purpose

This Requirements Traceability Matrix (RTM) establishes bidirectional traceability between business requirements, functional specifications, design elements, implementation components, and test cases
to ensure complete coverage and validation.

### Scope

{{RTM_SCOPE|requirements covered by this RTM}}

### Traceability Coverage

- **Forward Traceability**: Business Requirements → Functional Requirements → Design → Implementation → Tests
- **Backward Traceability**: Tests → Implementation → Design → Functional Requirements → Business Requirements
- **Bidirectional Validation**: Ensure all requirements are implemented and all implementations trace to requirements

## Requirements Traceability Matrix

### Format Legend

- **BR**: Business Requirement
- **FR**: Functional Requirement
- **NFR**: Non-Functional Requirement
- **UC**: Use Case
- **DS**: Design Specification
- **TC**: Test Case
- **IM**: Implementation Module

### Traceability Table

| Req ID | Requirement Description | Type | Priority | Source Document | Design Ref | Implementation Ref | Test Case Ref | Status | Verification Method |
|--------|------------------------|------|----------|-----------------|------------|-------------------|---------------|--------|--------------------|
| BR-001 | {{BR_001_DESCRIPTION}} | BR | High | {{BR_001_SOURCE_DOC}} | DS-001 | IM-001 | TC-001, TC-002 | ✅ Verified | Manual + Automated |
| FR-001 | {{FR_001_DESCRIPTION}} | FR | High | BR-001 | DS-001, DS-002 | IM-001, IM-003 | TC-003, TC-004 | ✅ Verified | Automated |
| NFR-001 | {{NFR_001_DESCRIPTION}} | NFR | Medium | BR-001 | DS-003 | IM-002 | TC-005 | ⚠️ Partial | Performance Test |
| UC-001 | {{UC_001_DESCRIPTION}} | UC | High | FR-001, FR-002 | DS-001 | IM-001 | TC-006 | ❌ Not Implemented | E2E Test |

### Status Legend

- ✅ **Verified**: Requirement fully implemented and tested
- ⚠️ **Partial**: Requirement partially implemented or under development
- ❌ **Not Implemented**: Requirement not yet implemented
- 🔄 **In Progress**: Currently being implemented
- 🚫 **Blocked**: Implementation blocked by dependencies

## Detailed Traceability Analysis

### Business Requirements Coverage

#### BR-001: {{BR_001_TITLE}}

- **Description**: {{BR_001_DETAILED_DESCRIPTION}}
- **Business Value**: {{BR_001_BUSINESS_VALUE|why this requirement is important}}
- **Acceptance Criteria**: {{BR_001_ACCEPTANCE_CRITERIA|how success is measured}}
- **Traced To**:
  - Functional Requirements: FR-001, FR-002
  - Design Elements: DS-001
  - Implementation: IM-001
  - Test Cases: TC-001, TC-002, TC-006
- **Coverage Status**: ✅ Fully Covered
- **Verification**: {{BR_001_VERIFICATION|how this requirement is verified}}

### Functional Requirements Coverage

#### FR-001: {{FR_001_TITLE}}

- **Description**: {{FR_001_DETAILED_DESCRIPTION}}
- **Input/Output**: {{FR_001_INPUTS_OUTPUTS|expected inputs and outputs}}
- **Business Rules**: {{FR_001_BUSINESS_RULES|business rules that apply}}
- **Traced From**: BR-001
- **Traced To**:
  - Design Elements: DS-001, DS-002
  - Implementation: IM-001, IM-003
  - Test Cases: TC-003, TC-004
- **Coverage Status**: ✅ Fully Covered
- **Implementation Notes**: {{FR_001_IMPLEMENTATION_NOTES|implementation considerations}}

### Non-Functional Requirements Coverage

#### NFR-001: {{NFR_001_TITLE}}

- **Description**: {{NFR_001_DESCRIPTION|performance, security, usability, or other requirement}}
- **Acceptance Criteria**: {{NFR_001_ACCEPTANCE_CRITERIA|quantifiable metrics and thresholds}}
- **Priority**: {{NFR_001_PRIORITY|High, Medium, or Low}}
- **Traced From**: BR-001
- **Traced To**:
  - Design Elements: DS-003
  - Implementation: IM-002
  - Test Cases: TC-005
- **Coverage Status**: ⚠️ Partial Coverage
- **Gap Analysis**: {{NFR_001_GAP_ANALYSIS|what is missing and why}}

## Implementation Traceability

### Implementation Modules

| Module ID | Module Name | Description | Requirements Covered | Test Coverage | Status |
|-----------|-------------|-------------|---------------------|---------------|--------|
| IM-001 | {{IM_001_NAME}} | {{IM_001_DESCRIPTION}} | BR-001, FR-001, UC-001 | TC-001, TC-002, TC-003, TC-004, TC-006 | ✅ Complete |
| IM-002 | {{IM_002_NAME}} | {{IM_002_DESCRIPTION}} | NFR-001 | TC-005 | ⚠️ In Progress |
| IM-003 | {{IM_003_NAME}} | {{IM_003_DESCRIPTION}} | FR-001 | TC-003, TC-004 | ✅ Complete |

### Code Coverage Mapping

```text

[Project Structure]
src/
├── core/
│   ├── [module].ts          # Implements: FR-001, FR-002
│   └── [service].ts         # Implements: NFR-001
├── api/
│   └── [controller].ts      # Implements: FR-001, UC-001
└── utils/
    └── [helper].ts          # Implements: FR-002

tests/
├── unit/
│   ├── [module].test.ts     # Tests: TC-003, TC-004
│   └── [service].test.ts    # Tests: TC-005
├── integration/
│   └── [api].test.ts        # Tests: TC-001, TC-002
└── e2e/
    └── [workflow].test.ts   # Tests: TC-006

```

## Test Coverage Analysis

### Test Case Mapping

| Test ID | Test Description | Test Type | Requirements Tested | Implementation Tested | Status | Last Run | Result |
|---------|------------------|-----------|-------------------|---------------------|--------|----------|--------|
| TC-001 | {{TC_001_DESCRIPTION}} | Integration | BR-001, FR-001 | IM-001 | ✅ Active | {{TC_001_LAST_RUN}} | Pass |
| TC-002 | {{TC_002_DESCRIPTION}} | Integration | BR-001 | IM-001 | ✅ Active | {{TC_002_LAST_RUN}} | Pass |
| TC-003 | {{TC_003_DESCRIPTION}} | Unit | FR-001 | IM-001, IM-003 | ✅ Active | {{TC_003_LAST_RUN}} | Pass |
| TC-004 | {{TC_004_DESCRIPTION}} | Unit | FR-001 | IM-001, IM-003 | ✅ Active | {{TC_004_LAST_RUN}} | Pass |
| TC-005 | {{TC_005_DESCRIPTION}} | Performance | NFR-001 | IM-002 | ⚠️ Partial | {{TC_005_LAST_RUN}} | Fail |
| TC-006 | {{TC_006_DESCRIPTION}} | E2E | UC-001 | IM-001 | ❌ Not Created | N/A | N/A |

### Coverage Metrics

- **Requirements Coverage**: 85% (6/7 requirements have associated tests)
- **Implementation Coverage**: 100% (All modules have associated tests)
- **Test Execution Coverage**: 71% (5/7 tests passing)
- **Traceability Completeness**: 90% (All relationships documented)

## Gap Analysis

### Missing Requirements

- {{MISSING_REQUIREMENTS|requirements gaps identified}}

### Missing Implementation

- UC-001: User workflow not yet implemented
- NFR-001: Performance optimization incomplete

### Missing Tests

- TC-006: End-to-end test for UC-001 not created
- Performance baseline tests for NFR-001

### Orphaned Implementation

- {{ORPHANED_IMPLEMENTATION|implementation not traced to requirements}}

### Orphaned Tests

- {{ORPHANED_TESTS|tests not traced to requirements}}

## Change Impact Analysis

### Recent Changes

| Date | Change Description | Requirements Affected | Implementation Impact | Test Impact | Status |
|------|-------------------|---------------------|---------------------|-------------|--------|
| {{CHANGE_1_DATE}} | {{CHANGE_1_DESCRIPTION}} | FR-001, NFR-001 | IM-001, IM-002 | TC-003, TC-005 | ✅ Updated |

### Change Management Process

1. **Requirement Change**: Update RTM with new/modified requirements
2. **Impact Assessment**: Identify affected design, implementation, and tests
3. **Implementation Update**: Modify implementation to meet new requirements
4. **Test Update**: Update test cases to validate new requirements
5. **Verification**: Ensure traceability is maintained
6. **Approval**: Stakeholder approval of changes

## Quality Metrics

### Traceability Health

- **Forward Traceability**: 100% (All requirements trace forward)
- **Backward Traceability**: 95% (Most implementation traces back)
- **Bidirectional Completeness**: 97.5%

### Requirements Quality

- **Testable Requirements**: 85% (6/7 requirements have clear test criteria)
- **Ambiguous Requirements**: 1 (NFR-001 needs clarification)
- **Requirements Stability**: 90% (1 requirement changed this iteration)

### Implementation Quality

- **Code Coverage**: 92% (Line coverage of requirement-related code)
- **Test Coverage**: 71% (Passing tests / Total tests)
- **Documentation Coverage**: 100% (All modules documented)

## Compliance & Audit

### Regulatory Compliance

- [ ] All security requirements traced and implemented
- [ ] Privacy requirements validated through testing
- [ ] Accessibility requirements verified
- [ ] Performance requirements measured

### Audit Trail

- **Last Updated**: {{LAST_UPDATED_DATE}}
- **Updated By**: {{UPDATED_BY}}
- **Review Status**: {{REVIEW_STATUS|Pending/Approved}}
- **Next Review Date**: {{NEXT_REVIEW_DATE}}

### Sign-off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Business Analyst | {{BA_SIGNATORY_NAME}} | {{BA_SIGNATURE}} | {{BA_SIGNATURE_DATE}} |
| Technical Lead | {{TECH_LEAD_SIGNATORY_NAME}} | {{TECH_LEAD_SIGNATURE}} | {{TECH_LEAD_SIGNATURE_DATE}} |
| QA Lead | {{QA_LEAD_SIGNATORY_NAME}} | {{QA_LEAD_SIGNATURE}} | {{QA_LEAD_SIGNATURE_DATE}} |
| Product Owner | {{PO_SIGNATORY_NAME}} | {{PO_SIGNATURE}} | {{PO_SIGNATURE_DATE}} |

## Appendices

### A. Requirements Documents

- {{REQUIREMENTS_DOCUMENTS|all requirements documents referenced}}

### B. Design Documents

- {{DESIGN_DOCUMENTS|all design documents referenced}}

### C. Test Documentation

- {{TEST_DOCUMENTATION|all test documentation referenced}}

### D. Tools and Automation

{{TRACEABILITY_TOOLS|tools used for traceability management}}

---

## Document Control

- **Template Version**: 1.0
- **Created**: {{CREATED_DATE}}
- **Last Modified**: {{LAST_MODIFIED_DATE}}
- **Next Review**: {{NEXT_REVIEW_DATE}}
- **Owner**: {{OWNER|team or role}}
