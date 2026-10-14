# Threat Model: {{SYSTEM_NAME}}

**System**: {{SYSTEM_NAME}} · **Version**: {{VERSION}} · **Date**: {{DATE}}
**Security Team**: {{TEAM}} · **Review Status**: {{REVIEW_STATUS}} — one of: Draft,
In Review, Approved, Needs Update

Approach: STRIDE (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege), plus LINDDUN for privacy where personal data is involved.

## 1. Executive Summary

{{SYSTEM_OVERVIEW|brief overview of the system and key security considerations}}

## 2. Scope and Architecture

- **Components**: {{COMPONENT1}}, {{COMPONENT2}}
- **Data flows**: {{FLOW1}}, {{FLOW2}}
- **Trust boundaries**: {{BOUNDARY1}}, {{BOUNDARY2}}
- **External dependencies**: {{DEPENDENCY1}}, {{DEPENDENCY2}}

## 3. Data Classification

| Type | Sensitivity | Examples |
|------|-------------|----------|
| {{PUBLIC_DATA_EXAMPLES|examples of data in this class}} | Public | {{EXAMPLES}} |
| {{INTERNAL_DATA_EXAMPLES|examples of data in this class}} | Internal | {{EXAMPLES}} |
| {{CONFIDENTIAL_DATA_EXAMPLES|examples of data in this class}} | Confidential | {{EXAMPLES}} |
| {{RESTRICTED_DATA_EXAMPLES|examples of data in this class}} | Restricted | {{EXAMPLES}} |

## 4. Threat Enumeration (STRIDE)

| ID | Category | Description | Component | Attack Vector | Likelihood | Impact | Current Controls | Residual Risk |
|----|----------|--------------|-----------|----------------|------------|--------|-------------------|----------------|
| S001 | Spoofing | | | | Low/Med/High | Low/Med/High/Critical | | |
| T001 | Tampering | | | | | | | |
| R001 | Repudiation | | | | | | | |
| I001 | Info Disclosure | | | | | | | |
| D001 | Denial of Service | | | | | | | |
| E001 | Elevation of Privilege | | | | | | | |

## 5. Privacy Threats (LINDDUN)

| ID | Category | Description | Data | Likelihood | Impact | Mitigation |
|----|----------|--------------|------|------------|--------|------------|

## 6. Abuse/Misuse Cases

- {{THREAT_SCENARIO}}: {{THREAT_DEFENSE_OR_MITIGATION}}

## 7. Security Controls

**Current**: {{CONTROL}} — type: preventive/detective/corrective, effectiveness: Low/Med/High, coverage: {{CONTROL_COVERAGE|what it protects}}

**Required**: {{CONTROL}} — addresses {{ADDRESSED_THREAT_IDS|threat IDs}}, priority {Low/Med/High/Critical}, effort {Small/Medium/Large}, owner {{TEAM}}

## 8. Risk Ranking and Treatment

- Critical: {{COUNT}} · High: {{COUNT}} · Medium: {{COUNT}} · Low: {{COUNT}} (unmitigated critical/high called out separately)
- **Overall posture**: {Poor/Fair/Good/Excellent} — **Recommendation**: {Block/Review/Approve with conditions/Approve}
- Top risk scenarios ranked by score, each with threat IDs, impact, likelihood, and required mitigation

## 9. Compliance Mapping

Applicable regimes (GDPR, ISO 27001, SOC 2, PCI-DSS, ...) with relevant requirements/controls and compliance status per regime.

## 10. Implementation Plan

- **Critical (0-30d)**: {{MITIGATIONS}}
- **High (30-90d)**: {{MITIGATIONS}}
- **Medium (90-180d)**: {{MITIGATIONS}}
- **Low (180d+)**: {{MITIGATIONS}}

## 11. Monitoring & Incident Response

- **Monitoring**: SIEM integration, log sources, correlation rules, alerting/escalation, threat intel feeds
- **Incident response**: playbooks, response team, communication plan, recovery procedures

## 12. Validation

- Security testing, code review, penetration testing performed against the threats above

## 13. Review & Approval

- **Review cadence**: security review / threat model update / risk assessment / control effectiveness
- **Approvals**: security architect, CISO, business owner — name, date, status

## Appendix

- Attack trees, data flow diagrams, network diagrams, reference materials (link or embed as needed)

---
*Template version: 2025.1 · Created {{TIMESTAMP}} · Last updated {{TIMESTAMP}}*
