# Feature Specification: [FEATURE NAME]

**Feature Branch**: `[###-feature-name]`  
**Created**: [DATE]  
**Status**: Draft  
**Input**: User description: "$ARGUMENTS"

## Execution Flow (main)

```
1. Parse user description from Input
   → If empty: ERROR "No feature description provided"
2. Extract key concepts from description
   → Identify: actors, actions, data, constraints
3. For each unclear aspect:
   → Mark with [NEEDS CLARIFICATION: specific question]
4. Fill User Scenarios & Testing section
   → If no clear user flow: ERROR "Cannot determine user scenarios"
5. Generate Functional Requirements
   → Each requirement must be testable
   → Mark ambiguous requirements
6. Identify Key Entities (if data involved)
7. Run Review Checklist
   → If any [NEEDS CLARIFICATION]: WARN "Spec has uncertainties"
   → If implementation details found: ERROR "Remove tech details"
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines

- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

### Section Requirements

- **Mandatory sections**: Must be completed for every feature
- **Optional sections**: Include only when relevant to the feature
- When a section doesn't apply, remove it entirely (don't leave as "N/A")

### For AI Generation

When creating this spec from a user prompt:

1. **Mark all ambiguities**: Use [NEEDS CLARIFICATION: specific question] for any assumption you'd need to make
2. **Don't guess**: If the prompt doesn't specify something (e.g., "login system" without auth method), mark it
3. **Think like a tester**: Every vague requirement should fail the "testable and unambiguous" checklist item
4. **Security-first mindset**: Consider authentication, authorization, input validation for every feature
5. **Performance considerations**: Think about scale, response times, resource usage
6. **Common underspecified areas**:
   - User types and permissions (roles, access levels)
   - Data retention/deletion policies (GDPR compliance)
   - Performance targets and scale (concurrent users, data volume)
   - Error handling behaviors (user experience, system recovery)
   - Integration requirements (external APIs, services)
   - Security requirements (authentication methods, data protection)
   - Monitoring and observability needs (logging, metrics, alerts)
   - Security/compliance needs

---

## User Scenarios & Testing _(mandatory)_

### Primary User Story

[Describe the main user journey in plain language]

### Acceptance Scenarios

1. **Given** [initial state], **When** [action], **Then** [expected outcome]
2. **Given** [initial state], **When** [action], **Then** [expected outcome]

### Edge Cases

- What happens when [boundary condition]?
- How does system handle [error scenario]?

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST [specific capability, e.g., "allow users to create accounts"]
- **FR-002**: System MUST [specific capability, e.g., "validate email addresses"]
- **FR-003**: Users MUST be able to [key interaction, e.g., "reset their password"]
- **FR-004**: System MUST [data requirement, e.g., "persist user preferences"]
- **FR-005**: System MUST [behavior, e.g., "log all security events"]

_Example of marking unclear requirements:_

- **FR-006**: System MUST authenticate users via [NEEDS CLARIFICATION: auth method not specified - email/password, SSO, OAuth?]
- **FR-007**: System MUST retain user data for [NEEDS CLARIFICATION: retention period not specified]

### Non-Functional Requirements _(mandatory)_

- **NFR-001**: Performance - [specific target, e.g., "API responses < 200ms p95"]
- **NFR-002**: Security - [specific requirement, e.g., "HTTPS required, input validation"]
- **NFR-003**: Scalability - [specific target, e.g., "support 1000 concurrent users"]
- **NFR-004**: Availability - [specific target, e.g., "99.9% uptime SLA"]
- **NFR-005**: Monitoring - [specific requirement, e.g., "structured logging for all user actions"]

### Security Requirements _(mandatory for user-facing features)_

- **SEC-001**: Authentication - [method and requirements]
- **SEC-002**: Authorization - [access control model]
- **SEC-003**: Data Protection - [encryption, privacy requirements]
- **SEC-004**: Input Validation - [sanitization, limits]
- **SEC-005**: Audit Logging - [security event tracking]

### Key Entities _(include if feature involves data)_

- **[Entity 1]**: [What it represents, key attributes without implementation]
- **[Entity 2]**: [What it represents, relationships to other entities]

---

## Review & Acceptance Checklist

_GATE: Automated checks run during main() execution_

### Content Quality

- [ ] No implementation details (languages, frameworks, APIs)
- [ ] Focused on user value and business needs
- [ ] Written for non-technical stakeholders
- [ ] All mandatory sections completed
- [ ] Security considerations addressed for user-facing features

### Requirement Completeness

- [ ] No [NEEDS CLARIFICATION] markers remain
- [ ] Requirements are testable and unambiguous
- [ ] Success criteria are measurable
- [ ] Scope is clearly bounded
- [ ] Dependencies and assumptions identified
- [ ] Performance targets specified
- [ ] Security requirements defined
- [ ] Error handling scenarios covered
- [ ] Data retention/privacy policies specified (if applicable)

### Constitutional Alignment

- [ ] Feature can be implemented as standalone library
- [ ] CLI interface requirements identified
- [ ] Integration testing scenarios defined
- [ ] Observability requirements specified
- [ ] Simplicity principle maintained (avoid unnecessary complexity)

---

## Execution Status

_Updated by main() during processing_

- [ ] User description parsed
- [ ] Key concepts extracted
- [ ] Ambiguities marked
- [ ] User scenarios defined
- [ ] Requirements generated
- [ ] Entities identified
- [ ] Review checklist passed

---
